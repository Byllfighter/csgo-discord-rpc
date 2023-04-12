package main

import (
	"fmt"
	"github.com/roberteinhaus/go-csgsi"
	"github.com/hugolgst/rich-go/client"
	"time"
	"net/http"
	"strconv"
)

type MatchDetails struct {
	tScore    int
	ctScore   int
	mapName   string
	timestamp client.Timestamps
}

type Connection struct {
	state          *csgsi.State
	activity       client.Activity
	lastConnection time.Time
}

var lastMatch MatchDetails
var c Connection

func setState(state *csgsi.State) {
	client.Login("937726683442712657")

	c = Connection{
		state:          state,
		activity:       client.Activity{},
		lastConnection: time.Now(),
	}

	if state.Map != nil {
		c.setGameState()
	} else {
		// Ended a game
		if lastMatch.mapName != "Menu" {
			now := time.Now()
			lastMatch = MatchDetails{
				tScore:  0,
				ctScore: 0,
				mapName: "Menu",
				timestamp: client.Timestamps{
					Start: &now,
				},
			}
		}

		err := client.SetActivity(client.Activity{
			Details:    "On Menu",
			LargeImage: "csgo",
			LargeText:  "Counter-Strike: Global Offensive",
			Timestamps: &lastMatch.timestamp,
		})

		if err != nil {
			panic(err)
		}
	}
}

func (c *Connection) setGameState() {
	c.setMapIcon()
	c.checkIfIsSameGame()
	c.setScoreboard()

	err := client.SetActivity(c.activity)

	if err != nil {
		panic(err)
	}
}

func (c *Connection) checkIfIsSameGame() {
	if lastMatch.mapName != c.state.Map.Name || lastMatch.tScore > c.state.Map.Team_t.Score || lastMatch.ctScore > c.state.Map.Team_ct.Score {
		now := time.Now()
		lastMatch.timestamp = client.Timestamps{
			Start: &now,
		}
	}

	c.activity.Timestamps = &lastMatch.timestamp
	lastMatch.mapName = c.state.Map.Name
	lastMatch.tScore = c.state.Map.Team_t.Score
	lastMatch.ctScore = c.state.Map.Team_ct.Score
}

func (c *Connection) setMapIcon() {

	matchStatsKills := strconv.Itoa(c.state.Player.Match_stats.Kills)
	matchStatsAssists := strconv.Itoa(c.state.Player.Match_stats.Assists)
	matchStatsDeaths := strconv.Itoa(c.state.Player.Match_stats.Deaths)
	matchStatsMVP := ""
	if c.state.Map.Mode == "survival" || c.state.Map.Mode == "gungameprogressive" || c.state.Map.Mode == "training" {
		matchStatsMVP = ""
	} else {
		matchStatsMVP = " | ☆: " + strconv.Itoa(c.state.Player.Match_stats.Mvps)
	}
	matchStatsScore := strconv.Itoa(c.state.Player.Match_stats.Score)

	c.activity.State = " K: " + matchStatsKills + " | A: " + matchStatsAssists + " | D: " + matchStatsDeaths + matchStatsMVP + " | Score: " + matchStatsScore


	mapIconLink := "https://raw.githubusercontent.com/Byllfighter/csgo-discord-rpc/main/images/maps/" + c.state.Map.Name + ".png"
	noneMapIconLink := "https://raw.githubusercontent.com/Byllfighter/csgo-discord-rpc/main/images/maps/none.png"

	// Default CSGO icon if map has no icon

    response, err := http.Get(mapIconLink)
    if err == nil && response.StatusCode == http.StatusOK {
        c.activity.LargeImage = mapIconLink
    } else {
	    c.activity.LargeImage = noneMapIconLink
	}
	
	
	mapGameModeName := ""
	if c.state.Map.Mode == "casual" {
		mapGameModeName = "Casual"
	} else if c.state.Map.Mode == "competitive" {
		mapGameModeName = "Competitive"
	} else if c.state.Map.Mode == "scrimcomp2v2" {
		mapGameModeName = "Wingman"
	} else if c.state.Map.Mode == "scrimcomp5v5" {
		mapGameModeName = "Weapons Expert"
	} else if c.state.Map.Mode == "gungameprogressive" {
		mapGameModeName = "Arms Race"
	} else if c.state.Map.Mode == "gungametrbomb" {
		mapGameModeName = "Demolition"
	} else if c.state.Map.Mode == "deathmatch" {
		mapGameModeName = "Deathmatch"
	} else if c.state.Map.Mode == "training" {
		mapGameModeName = "Training"
	} else if c.state.Map.Mode == "custom" {
		mapGameModeName = "Custom"
	} else if c.state.Map.Mode == "cooperative" {
		mapGameModeName = "Guardian"
	} else if c.state.Map.Mode == "coopmission" {
		mapGameModeName = "Co-op Strike"
	} else if c.state.Map.Mode == "skirmish" {
		mapGameModeName = "War Games"
	} else if c.state.Map.Mode == "survival" {
		mapGameModeName = "Danger Zone"
	} else {
		mapGameModeName = c.state.Map.Mode
	}

	c.activity.LargeText = mapGameModeName + " | " + c.state.Map.Name
}

func (c *Connection) setScoreboard() {
	switch c.state.Map.Phase {
	case "live":
		c.activity.Details += "Playing "
	case "warmup":
		c.activity.Details += "Warming up "
	case "intermission":
		c.activity.Details += "Switching sides "
	case "gameover":
		c.activity.Details += "Ending "
	}

	if c.state.Player.Team == "CT" {
		c.activity.SmallImage = "ct"
		c.activity.SmallText = "Counter-Terrorist"
	} else if c.state.Player.Team == "T" {
		c.activity.SmallImage = "t"
		c.activity.SmallText = "Terrorist"
	} else {
		c.activity.SmallImage = "spectator"
		c.activity.SmallText = "Spectator"
	}
	
	if c.state.Map.Mode == "survival" || c.state.Map.Mode == "gungameprogressive" || c.state.Map.Mode == "training" {
	} else {
		if c.state.Player.Team == "CT" {
			c.activity.Details += fmt.Sprintf("[%d : %d]", c.state.Map.Team_ct.Score, c.state.Map.Team_t.Score)
		} else if c.state.Player.Team == "T" {
			c.activity.Details += fmt.Sprintf("[%d : %d]", c.state.Map.Team_t.Score, c.state.Map.Team_ct.Score)
		} else {
			c.activity.Details += fmt.Sprintf("[%d : %d]", c.state.Map.Team_ct.Score, c.state.Map.Team_t.Score)
		}
	}
}