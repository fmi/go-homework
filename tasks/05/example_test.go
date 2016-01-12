package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
		path := fmt.Sprintf("/cell/status/?x=%d&y=%d", testCase.x, testCase.y)
		cellURL := buildUrl(testSrv.URL, path)

		resp, err := http.Get(cellURL)

		if err != nil {
			t.Errorf("Error getting empty board: %s", err)
		}

		defer resp.Body.Close()

		cellStatus := &struct {
			Alive bool `json:"alive"`
		}{}

		respBytes, err := ioutil.ReadAll(resp.Body)

		if err != nil && err != io.EOF {
			t.Errorf("Error reading from response: %s", err)
		}

		if err := json.Unmarshal(respBytes, cellStatus); err != nil {
			t.Errorf("Error decoding json: %s", err)
		}

		if testCase.alive != cellStatus.Alive {
			t.Errorf("Expected alive %t for (%d, %d) but it was %t. "+
				"JSON: %s", testCase.alive, testCase.x, testCase.y, cellStatus.Alive,
				string(respBytes))
		}
	}

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
