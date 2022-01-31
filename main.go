package main

import (
	"encoding/json"
	"github.com/dank/go-csgsi"
	"github.com/hugolgst/rich-go/client"
	"net/http"
	"time"
)

func stateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var s csgsi.State

		err := json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			return
		}

		setState(&s)
	}
}

func main() {
	now := time.Now()
	lastMatch = MatchDetails{
		tScore:  0,
		ctScore: 0,
		timestamp: client.Timestamps{
			Start: &now,
		},
	}

	initializeRPC()

	http.HandleFunc("/", stateHandler)
	http.ListenAndServe(":730", nil)
}
