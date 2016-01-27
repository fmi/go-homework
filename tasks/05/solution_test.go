package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestCreatingAHandler(t *testing.T) {

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic: %s", r)
		}
	}()

	gofh := NewGameOfLifeHandler(nil)

	if gofh == nil {
		t.Error("Creating a game of life handler with nil failed")
	}
}

func TestCreatingWithCells(t *testing.T) {

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic: %s", r)
		}
	}()

	gofh := NewGameOfLifeHandler([][2]int64{
		{-1, 2},
		{3, 42},
		{0, 15},
		{14, -10},
	})

	if gofh == nil {
		t.Error("Creating a game of life handler with nil failed")
	}
}

func TestCreatingEmptyBoard(t *testing.T) {
	testSrv := setUpServer(nil)
	defer testSrv.Close()

	livingUrl := buildUrl(testSrv.URL, "/generation/")

	resp, err := http.Get(livingUrl)

	if err != nil {
		t.Errorf("Error getting empty board: %s", err)
	}

	defer resp.Body.Close()

	gen := &struct {
		generation int64             `json:"generation"`
		livingRaw  []json.RawMessage `json:"living"`
	}{}

	respBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil && err != io.EOF {
		t.Errorf("Error reading from response: %s", err)
	}

	if err := json.Unmarshal(respBytes, gen); err != nil {
		t.Errorf("Error decoding json: %s", err)
	}

	living := len(gen.livingRaw)

	if living != 0 {
		t.Errorf("Expected empty board to have no living cells but it had %d", living)
	}

	if gen.generation != 0 {
		t.Errorf("Expected generation number to be 0 but it was %d", gen.generation)
	}
}

func TestCreatingBoardWithSeed(t *testing.T) {
	testSrv := setUpServer([][2]int64{
		{0, 0},
		{1, 1},
		{3, 4},
		{-2, 5},
		{-19023482123, 5},
	})
	defer testSrv.Close()

	testTable := []struct {
		alive bool
		x, y  int64
	}{
		{x: 0, y: 0, alive: true},
		{x: 1, y: 1, alive: true},
		{x: 3, y: 4, alive: true},
		{x: -2, y: 5, alive: true},
		{x: -19023482123, y: 5, alive: true},
		{x: 0, y: -1, alive: false},
		{x: -100, y: 100, alive: false},
		{x: 100, y: 100, alive: false},
		{x: 55, y: 93, alive: false},
	}

	for _, testCase := range testTable {
		cellAlive, err := isCellAlive(testSrv, testCase.x, testCase.y)

		if err != nil {
			t.Error(err)
			continue
		}

		if testCase.alive != cellAlive {
			t.Errorf("Expected alive %t for (%d, %d) but it was %t.",
				testCase.alive, testCase.x, testCase.y, cellAlive)
		}
	}

}

func TestGenerationCounting(t *testing.T) {
	srv := setUpGliderServer()
	defer srv.Close()

	evolveUrl := buildUrl(srv.URL, "/generation/evolve/")

	for i := 0; i < 4; i++ {

		resp, err := http.Post(evolveUrl, "", nil)

		if err != nil {
			t.Fatal(err)
		}

		resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected status code %d on interation %d but it was %d",
				http.StatusNoContent, i, resp.StatusCode)
		}
	}

	respStruct, err := getGenerationStatus(srv)

	if err != nil {
		t.Fatal(err)
	}

	if respStruct.Gen != 4 {
		t.Errorf("Expected generation numer to be 4. It was %d", respStruct.Gen)
	}
}

func TestBoardReset(t *testing.T) {
	srv := setUpGliderServer()
	defer srv.Close()

	resetUrl := buildUrl(srv.URL, "/reset/")
	resp, err := http.Post(resetUrl, "", nil)

	if err != nil {
		t.Fatalf("Error resetting generation: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected clearing the board to respond with %d but it was %d",
			http.StatusNoContent, resp.StatusCode)
	}

	respStruct, err := getGenerationStatus(srv)

	if err != nil {
		t.Fatal(err)
	}

	if respStruct.Gen != 0 {
		t.Errorf("Resetting did not reset generetion number. It was left at %d", respStruct.Gen)
	}

	if len(respStruct.Living) > 0 {
		t.Error("Resetting did not remove all living cells")
	}
}

func TestAddingCells(t *testing.T) {
	srv := setUpGliderServer()
	defer srv.Close()

	addUrl := buildUrl(srv.URL, "/cells/")

	resp, err := http.Post(addUrl, "application/json", bytes.NewReader([]byte(`[
		{"y": -1024, "x": 1024},
		{"x": 900432123, "y": -900432123}
	]`)))

	if err != nil {
		t.Errorf("Error adding new cells: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var responseText string

		if respBody, err := ioutil.ReadAll(resp.Body); err != nil {
			responseText = fmt.Sprintf("Response was not read: %s", err)
		} else {
			responseText = string(respBody)
		}

		t.Errorf("Expected status Created but it was %d with response: %s", resp.StatusCode,
			responseText)
	}

	respStruct, err := getGenerationStatus(srv)

	if err != nil {
		t.Fatalf("Error getting generation status: %s", err)
	}

	if len(respStruct.Living) != 7 {
		t.Errorf("Expected 7 living cells but they were %d", len(respStruct.Living))
	}

	var found1024, foundOver9Thousand = false, false

	for _, pair := range respStruct.Living {
		if pair[0] == 1024 && pair[1] == -1024 {
			found1024 = true
		}

		if pair[0] == 900432123 && pair[1] == -900432123 {
			foundOver9Thousand = true
		}
	}

	if !found1024 || !foundOver9Thousand {
		t.Errorf("Added cells were not found in the living: %v", respStruct.Living)
	}
}

func TestWrongHTTPMethods(t *testing.T) {
	srv := setUpGliderServer()
	defer srv.Close()

	tests := []struct {
		method string
		url    string
	}{
		{"GET", "/generation/evolve/"},
		{"PUT", "/generation/evolve/"},
		{"DELETE", "/generation/evolve/"},
		{"POST", "/generation/"},
		{"PUT", "/generation/"},
		{"DELETE", "/generation/"},
		{"GET", "/cells/"},
		{"PUT", "/cells/"},
		{"DELETE", "/cells/"},
		{"GET", "/reset/"},
		{"PUT", "/reset/"},
		{"DELETE", "/reset/"},
		{"POST", "/cell/status/"},
		{"PUT", "/cell/status/"},
		{"DELETE", "/cell/status/"},
	}

	client := &http.Client{}

	for _, test := range tests {
		testUrl := buildUrl(srv.URL, test.url)

		req, err := http.NewRequest(test.method, testUrl, nil)

		if err != nil {
			t.Errorf("Error creating request: %s", err)
			continue
		}

		resp, err := client.Do(req)

		if err != nil {
			t.Errorf("Error making a HTTP request: %s", err)
		}

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected method not allowed for %s %s but it was %d",
				test.method, test.url, resp.StatusCode)
		}
	}
}

func TestWrongURLs(t *testing.T) {
	srv := setUpGliderServer()
	defer srv.Close()

	wrongPaths := []string{
		"/some/url/",
		"/other/",
		"/does/not/exists/",
	}

	for _, wrongPath := range wrongPaths {
		wrongUrl := buildUrl(srv.URL, wrongPath)
		resp, err := http.Get(wrongUrl)
		if err != nil {
			t.Errorf("Error getting url %s: %s", wrongUrl, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Did not receive Not Found but %d for %s", resp.StatusCode, wrongPath)
		}
	}
}

func TestWrongJSONs(t *testing.T) {
	srv := setUpGliderServer()
	defer srv.Close()

	tests := []struct {
		path string
		body string
	}{
		{"/cells/", "asdasdasd"},
		{"/cells/", "{2: 3}"},
		{"/cells/", "[[[[[]"},
		{"/cells/", "{'x': 2, 'y': -2}"},    // wrong quotes
		{"/cells/", "[{'x': 2, 'y': -2}]"},  // wrong quotes
		{"/cells/", "[]]"},                  // additional closing bracket
		{"/cells/", `[{"x": 2, "y": -2}]]`}, // additional closing bracket
	}

	for _, test := range tests {
		postUrl := buildUrl(srv.URL, test.path)
		resp, err := http.Post(postUrl, "application/json", bytes.NewReader([]byte(test.body)))
		if err != nil {
			t.Fatalf("Error seding POST request: %s", err)
		}

		if resp.StatusCode == http.StatusCreated {
			t.Errorf("Expected non %d but got %d for adding cells with body: %s",
				http.StatusCreated, resp.StatusCode, test.body)
		}
	}
}

func TestBoardEvolutionBasicRules(t *testing.T) {

	tests := []testCase{
		{
			name: "Dies of underpopulation",
			board: [][2]int64{
				{1, 1},
			},

			dead: [][2]int64{
				{1, 1},
			},
		},
		{
			name: "Alive if has enough neighbours",
			board: [][2]int64{
				{0, 1}, {1, 1}, {2, 1}, {1, 2},
			},

			alive: [][2]int64{
				{0, 1}, {1, 1}, {2, 1}, {1, 2}, {1, 0},
			},
		},
		{
			name: "Kills overpopulated",
			board: [][2]int64{
				{0, 1}, {1, 1}, {2, 1}, {1, 2}, {1, 0},
			},

			dead: [][2]int64{
				{1, 1},
			},
		},
		{
			name: "New life is spawned",
			board: [][2]int64{
				{0, 1}, {1, 1}, {2, 1},
			},

			alive: [][2]int64{
				{1, 0}, {1, 2},
			},
		},
	}

	for _, test := range tests {
		srv := setUpServer(test.board)
		defer srv.Close()

		if err := nextGeneration(srv); err != nil {
			t.Errorf("In test %s: %s", test.name, err)
		}

		test.testAliveAndDead(t, srv)
	}

}

func TestBoardEvolutionStills(t *testing.T) {

	tests := []testCase{
		{
			name: "Block",
			board: [][2]int64{
				{0, 0}, {1, 0}, {0, 1}, {1, 1},
			},
			alive: [][2]int64{
				{0, 0}, {1, 0}, {0, 1}, {1, 1},
			},
		},
		{
			name: "Boat",
			board: [][2]int64{
				{0, 0}, {1, 0}, {0, 1}, {2, 1}, {1, 2},
			},

			alive: [][2]int64{
				{0, 0}, {1, 0}, {0, 1}, {2, 1}, {1, 2},
			},
		},
	}

	for _, test := range tests {
		srv := setUpServer(test.board)
		defer srv.Close()

		for i := 0; i < 5; i++ {
			if err := nextGeneration(srv); err != nil {
				t.Errorf("In test %s: %s", test.name, err)
			}
		}

		test.testAliveAndDead(t, srv)
	}

}

func TestBoardEvolutionOscillators(t *testing.T) {

	tests := []testCase{
		{
			name: "Blinker",
			board: [][2]int64{
				{-1, 0}, {0, 0}, {1, 0},
			},
			dead: [][2]int64{
				{-1, 0}, {1, 0},
			},
			alive: [][2]int64{
				{0, -1}, {0, 0}, {0, 1},
			},
		},
		{
			name: "Beacon",
			board: [][2]int64{
				{1, 1}, {1, 2}, {2, 1}, {2, 2}, {3, 3}, {3, 4}, {4, 3}, {4, 4},
			},

			alive: [][2]int64{
				{1, 1}, {1, 2}, {2, 1}, {3, 4}, {4, 3}, {4, 4},
			},

			dead: [][2]int64{
				{2, 2}, {3, 3},
			},
		},
	}

	for _, test := range tests {
		srv := setUpServer(test.board)
		defer srv.Close()

		if err := nextGeneration(srv); err != nil {
			t.Errorf("In test %s: %s", test.name, err)
		}

		test.testAliveAndDead(t, srv)

		if err := nextGeneration(srv); err != nil {
			t.Errorf("In test %s: %s", test.name, err)
		}

		test.alive = test.board
		test.dead = [][2]int64{}

		test.testAliveAndDead(t, srv)

	}
}

// We are mainly interested in panics because of race conditions here
func TestConcurrentOperation(t *testing.T) {
	srv := setUpGliderServer()
	defer srv.Close()

	start := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			addUrl := buildUrl(srv.URL, "/cells/")

			for i := 0; i < 100; i++ {
				x, y := rand.Int63(), rand.Int63()
				body := fmt.Sprintf(`[
					{"x": %d, "y": %d}
				]`, x, y)
				resp, err := http.Post(addUrl, "application/json", bytes.NewReader([]byte(body)))

				if err != nil {
					t.Errorf("Error adding new cells: %s", err)
				}

				resp.Body.Close()
			}
		}()
	}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start

			for i := 0; i < 100; i++ {
				if _, err := isCellAlive(srv, rand.Int63(), rand.Int63()); err != nil {
					t.Error(err)
				}
			}
		}()
	}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if err := nextGeneration(srv); err != nil {
				t.Error(err)
			}
		}()
	}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start

			generationUrl := buildUrl(srv.URL, "/generation/")

			resp, err := http.Get(generationUrl)

			if err != nil {
				t.Errorf("Got HTTP error on getting generation: %s", err)
			}

			resp.Body.Close()
		}()
	}

	close(start)
	wg.Wait()
}

/* Utility functions */

func buildUrl(baseUrl, path string) string {
	return fmt.Sprintf("%s%s", baseUrl, path)
}

// Users of this function are resposible for calling Close() on the returned server.
// Failure to do so will result in leaked resources.
func setUpServer(cells [][2]int64) *httptest.Server {
	gofh := NewGameOfLifeHandler(cells)
	return httptest.NewServer(gofh)
}

// Will setup a server with a slider
// thus making sure it will have cells for eternity
func setUpGliderServer() *httptest.Server {
	return setUpServer([][2]int64{
		{0, 0},
		{1, 1},
		{1, 2},
		{0, 2},
		{-1, 2},
	})
}

func getGenerationStatus(srv *httptest.Server) (*GenerationStatus, error) {

	generationUrl := buildUrl(srv.URL, "/generation/")

	resp, err := http.Get(generationUrl)

	if err != nil {
		return nil, fmt.Errorf("Got HTTP error on getting generation: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Wrong status code on GETing generation status: %d", resp.StatusCode)
	}

	respStruct := &GenerationStatus{}

	respBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("Error reading generation status: %s", err)
	}

	if err := json.Unmarshal(respBytes, respStruct); err != nil {
		return nil, fmt.Errorf("Error unmarshaling generation status: %s", err)
	}

	return respStruct, nil
}

func isCellAlive(srv *httptest.Server, x, y int64) (bool, error) {

	path := fmt.Sprintf("/cell/status/?x=%d&y=%d", x, y)
	cellURL := buildUrl(srv.URL, path)

	resp, err := http.Get(cellURL)

	if err != nil {
		return false, fmt.Errorf("Error GETing from board: %s", err)
	}

	defer resp.Body.Close()

	cellStatus := &struct {
		Alive bool `json:"alive"`
	}{}

	respBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil && err != io.EOF {
		return false, fmt.Errorf("Error reading from response: %s", err)
	}

	if err := json.Unmarshal(respBytes, cellStatus); err != nil {
		return false, fmt.Errorf("Error decoding json: %s", err)
	}

	return cellStatus.Alive, nil
}

func nextGeneration(srv *httptest.Server) error {
	evolveUrl := buildUrl(srv.URL, "/generation/evolve/")

	resp, err := http.Post(evolveUrl, "", nil)

	if err != nil {
		return err
	}

	resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Wrong status code on evolving generataion: %d", resp.StatusCode)
	}

	return nil
}

type GenerationStatus struct {
	Gen    int64      `json:"generation"`
	Living [][2]int64 `json:"living"`
}

type testCase struct {
	name  string
	board [][2]int64
	alive [][2]int64
	dead  [][2]int64
}

func (test *testCase) testAliveAndDead(t *testing.T, srv *httptest.Server) {

	for _, pair := range test.alive {
		alive, err := isCellAlive(srv, pair[0], pair[1])
		if err != nil {
			t.Errorf("Error getting cell status: %s", err)
		}

		if !alive {
			t.Errorf("Expected cell %v to be alive in test %s", pair, test.name)
		}
	}

	for _, pair := range test.dead {
		alive, err := isCellAlive(srv, pair[0], pair[1])
		if err != nil {
			t.Errorf("Error getting cell status: %s", err)
		}

		if alive {
			t.Errorf("Expected cell %v to be dead in test %s", pair, test.name)
		}
	}

}
