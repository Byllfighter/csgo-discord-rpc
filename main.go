package main

import (
	"encoding/json"
	"fmt"
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

	c.lastConnection = time.Now()

	// Function to check if user is still playing
	go func() {
		waitTime := 180
		for {
			// User exited game
			if 1.5 < time.Now().Sub(c.lastConnection).Minutes() {
				client.Logout()
				waitTime = 180
			} else {
				waitTime = 90
			}
			time.Sleep(time.Duration(waitTime) * time.Second)
		}
	}()

	http.HandleFunc("/", stateHandler)
	err := http.ListenAndServe(":730", nil)
	fmt.Println(err)
}
