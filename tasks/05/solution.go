// Package main defines a solution for the task defined in README.md
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// Cell represents a square in the game board.
type Cell struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

// Neighbours returns all neighbours on the board for this particular cell.
func (c *Cell) Neighbours() (ret [8]Cell) {

	nbrs := [][2]int64{
		{-1, -1}, {0, -1}, {1, -1},
		{-1, 0}, {1, 0},
		{-1, 1}, {0, 1}, {1, 1},
	}

	for ind, nbr := range nbrs {
		ret[ind].X = c.X + nbr[0]
		ret[ind].Y = c.Y + nbr[1]
	}

	return
}

// GameOfLifeHandler represents a http.Handler which defines the behaviour of a
// Conaway's Game of Life board via various HTTP endpoints.
type GameOfLifeHandler struct {

	// GameOfLifeHandler is also a mutex. This is so in order to protect the sensitive
	// map of living cells against write operations from different threads.
	sync.RWMutex

	// writeMutex is used for synchronizing operations which change the current board during
	// the change.
	writeMutex sync.Mutex

	// This handler multiplexer is used to decompose the task in different handlers for
	// each path defined in the task.
	mux *http.ServeMux

	// cells is a collection of all living cells in the current generation.
	// A map of cells is used because most of the operations are about searching for
	// living cells.
	cells map[Cell]struct{}

	generation int64
}

// NewGameOfLifeHandler returns a new GameOfLifeHandler which is prepopulated with cells passed
// in `cells`. It initializes the handler by creating a http.ServerMux and populating with the
// appropriate callbacks.
func NewGameOfLifeHandler(cells [][2]int64) *GameOfLifeHandler {
	gofh := &GameOfLifeHandler{}
	gofh.cells = make(map[Cell]struct{})

	for _, coords := range cells {
		cell := Cell{X: coords[0], Y: coords[1]}
		gofh.cells[cell] = struct{}{}
	}

	gofh.mux = http.NewServeMux()
	gofh.mux.HandleFunc("/cell/status/", RestrictMethodHandler("GET", gofh.handleCellStatus))
	gofh.mux.HandleFunc("/generation/", RestrictMethodHandler("GET", gofh.handleGeneration))
	gofh.mux.HandleFunc("/generation/evolve/", RestrictMethodHandler("POST", gofh.handleEvolve))
	gofh.mux.HandleFunc("/reset/", RestrictMethodHandler("POST", gofh.handleReset))
	gofh.mux.HandleFunc("/cells/", RestrictMethodHandler("POST", gofh.handleAddCells))

	return gofh
}

// ServeHTTP implements the http.Handler interface. It just calls the request multiplexer
// which will find the correct callback for this request.
func (gofh *GameOfLifeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gofh.mux.ServeHTTP(w, r)
}

func (gofh *GameOfLifeHandler) handleGeneration(w http.ResponseWriter, r *http.Request) {
	gofh.RLock()
	defer gofh.RUnlock()

	living := make([][2]int64, 0, len(gofh.cells))

	for cell, _ := range gofh.cells {
		living = append(living, [2]int64{cell.X, cell.Y})
	}

	// This annonymous is used only for its json attributes' tags. This way we can leave
	// the JSON creation to the standard library. It is way safer than trying to generate
	// correct json ourselves.
	resp := &struct {
		Gen    int64      `json:"generation"`
		Living [][2]int64 `json:"living"`
	}{
		Gen:    gofh.generation,
		Living: living,
	}

	respBytes, err := json.Marshal(resp)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error marshalling results: %s", err)
		return
	}

	if _, err := w.Write(respBytes); err != nil {
		// At point we cannot send a different Status Code and we cannot continue to write
		// in the response. The only sante thing is to log the error.
		log.Printf("Error writing response: %s", err)
	}
}

func (gofh *GameOfLifeHandler) handleEvolve(w http.ResponseWriter, r *http.Request) {

	// We will work without mutating the old board. This would have been very difficult. Instead,
	// a new board is created. Every cell which should be alive in the next generation will
	// be added to this new board.
	newBoard := make(map[Cell]struct{})
	oldBoard := make(map[Cell]struct{})

	// One of the rules says we have to spawn dead cells with appropriate number of neighbours.
	// We have to decide which dead cells to consider for spawning. Obviosuly, we cannot try
	// with all dead cells since they are infinate. We can make the observation that only
	// dead cells which have a living neighbour can spawn. So this map will keep exactly such cells.
	// Note that are using a map in order to remove any duplications.
	interestingDead := make(map[Cell]struct{})

	// A small helper function which counts how many of the cells in the argument are alive.
	livingCells := func(neighbours [8]Cell) (count uint8) {
		for _, nbr := range neighbours {
			if _, ok := oldBoard[nbr]; ok {
				count += 1
			}
		}
		return
	}

	gofh.writeMutex.Lock()
	defer gofh.writeMutex.Unlock()

	gofh.RLock()
	for aliveCell, _ := range gofh.cells {
		oldBoard[aliveCell] = struct{}{}
	}
	gofh.RUnlock()

	// First pass. We decide which alive cells should live to see the next generation and
	// simultaniously populate the interestingDead map.
	for aliveCell, _ := range oldBoard {
		neighbours := aliveCell.Neighbours()
		livingCount := livingCells(neighbours)

		if livingCount == 2 || livingCount == 3 {
			newBoard[aliveCell] = struct{}{} // stayes alive
		}

		for _, nbr := range neighbours {
			interestingDead[nbr] = struct{}{}
		}
	}

	// Second pass. We have all the dead cells which can possibly spawn.
	for deadCell, _ := range interestingDead {
		neighbours := deadCell.Neighbours()

		if livingCells(neighbours) == 3 {
			newBoard[deadCell] = struct{}{} // spawns
		}
	}

	gofh.Lock()
	defer gofh.Unlock()

	// And it is here when we make the switch.
	gofh.cells = newBoard
	gofh.generation += 1
	w.WriteHeader(http.StatusNoContent)
}

func (gofh *GameOfLifeHandler) handleReset(w http.ResponseWriter, r *http.Request) {
	gofh.writeMutex.Lock()
	defer gofh.writeMutex.Unlock()

	gofh.Lock()
	defer gofh.Unlock()

	gofh.generation = 0
	gofh.cells = make(map[Cell]struct{})
	w.WriteHeader(http.StatusNoContent)
}

func (gofh *GameOfLifeHandler) handleAddCells(w http.ResponseWriter, r *http.Request) {

	reqBytes, err := ioutil.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error reading the request body: %s", err)
		return
	}

	defer r.Body.Close()

	added := []Cell{}

	if err := json.Unmarshal(reqBytes, &added); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error unmarshalling the request body: %s", err)
		return
	}

	gofh.writeMutex.Lock()
	defer gofh.writeMutex.Unlock()

	gofh.Lock()
	defer gofh.Unlock()

	for _, cell := range added {
		gofh.cells[cell] = struct{}{}
	}

	w.WriteHeader(http.StatusCreated)
}

func (gofh *GameOfLifeHandler) handleCellStatus(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	x, y := r.Form.Get("x"), r.Form.Get("y")

	intX, err := strconv.ParseInt(x, 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	intY, err := strconv.ParseInt(y, 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cell := Cell{X: intX, Y: intY}

	gofh.RLock()
	defer gofh.RUnlock()

	if _, ok := gofh.cells[cell]; ok {
		fmt.Fprint(w, `{"alive": true}`)
	} else {
		fmt.Fprint(w, `{"alive": false}`)
	}
}

// RestrictMethodHandler is used as a decorator to restrict a http.HandlerFunction to only one
// HTTP method. Every other method will get Method Not Allowed as a response.
func RestrictMethodHandler(method string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	}
}
