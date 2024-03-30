package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hugolgst/rich-go/client"
	"github.com/roberteinhaus/go-csgsi"
	"github.com/shirou/gopsutil/process"
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
var mapGameModeName string
var buttons []*client.Button
var workshopLink string
var isCS2 bool

func setState(state *csgsi.State) {
	client.Login("937726683442712657")

	c = Connection{
		state:          state,
		activity:       client.Activity{},
		lastConnection: time.Now(),
	}

	isCS2 = false
	processes, err := process.Processes()
	if err != nil {
		fmt.Println("Error:", err)
		isCS2 = false
	}

	for _, p := range processes {
		name, _ := p.Name()
		if name == "cs2.exe" {
			isCS2 = true
		}
	}

	workshopLink = ""
	c.setButtons()

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
		workshopLink = ""

		if isCS2 {
			err := client.SetActivity(client.Activity{
				Details:    "On Menu",
				LargeImage: "https://raw.githubusercontent.com/Byllfighter/csgo-discord-rpc/main/images/cs2.png",
				LargeText:  "Counter-Strike 2",
				Timestamps: &lastMatch.timestamp,
				Buttons:    buttons,
			})

			if err != nil {
				panic(err)
			}
		} else {
			err := client.SetActivity(client.Activity{
				Details:    "On Menu",
				LargeImage: "csgo",
				LargeText:  "Counter-Strike: Global Offensive",
				Timestamps: &lastMatch.timestamp,
				Buttons:    buttons,
			})

			if err != nil {
				panic(err)
			}
		}
	}
}

func (c *Connection) setGameState() {
	c.setWorkshopLink()
	c.setMapIcon()
	c.checkIfIsSameGame()
	c.setScoreboard()
	c.setMapMode()
	c.setMapName()
	c.setButtons()
	c.activity.Buttons = buttons

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
	if c.state.Map.Mode == "survival" || c.state.Map.Mode == "gungameprogressive" || c.state.Map.Mode == "training" || c.state.Map.Mode == "deathmatch" {
		matchStatsMVP = ""
	} else {
		matchStatsMVP = " | ☆: " + strconv.Itoa(c.state.Player.Match_stats.Mvps)
	}
	matchStatsScore := strconv.Itoa(c.state.Player.Match_stats.Score)

	c.activity.State = " K: " + matchStatsKills + " | A: " + matchStatsAssists + " | D: " + matchStatsDeaths + matchStatsMVP + " | Score: " + matchStatsScore

	mapIconLink := "https://raw.githubusercontent.com/Byllfighter/csgo-discord-rpc/main/images/maps/" + c.state.Map.Name + ".png"
	if isCS2 {
		mapIconLinkCS2 := "https://raw.githubusercontent.com/Byllfighter/csgo-discord-rpc/main/images/maps/cs2/" + c.state.Map.Name + "_png.png"
		mapIconLinkCS2Response, err := http.Get(mapIconLinkCS2)
		if err == nil && mapIconLinkCS2Response.StatusCode == http.StatusOK {
			mapIconLink = mapIconLinkCS2
		}
	}

	noneMapIconLink := "https://raw.githubusercontent.com/Byllfighter/csgo-discord-rpc/main/images/maps/none.png"

	// Default CSGO icon if map has no icon

	mapIconLinkResponse, err := http.Get(mapIconLink)
	if err == nil && mapIconLinkResponse.StatusCode == http.StatusOK {
		c.activity.LargeImage = mapIconLink
	} else if strings.HasPrefix(c.state.Map.Name, "workshop/") {
		c.setMapWorkshopImage()
	} else {
		c.activity.LargeImage = noneMapIconLink
	}
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

	if c.state.Map.Mode == "survival" || c.state.Map.Mode == "gungameprogressive" || c.state.Map.Mode == "training" || c.state.Map.Mode == "deathmatch" {
	} else {
		if c.state.Player.Team == "CT" {
			c.activity.Details += fmt.Sprintf("[%d : %d]", c.state.Map.Team_ct.Score, c.state.Map.Team_t.Score)
		} else if c.state.Player.Team == "T" {
			c.activity.Details += fmt.Sprintf("[%d : %d]", c.state.Map.Team_t.Score, c.state.Map.Team_ct.Score)
		} else {
			c.activity.Details += fmt.Sprintf("[%d : %d]", c.state.Map.Team_ct.Score, c.state.Map.Team_t.Score)
		}
	}

	if c.state.Map.Mode == "gungameprogressive" || c.state.Map.Mode == "deathmatch" || c.state.Map.Mode == "gungametrbomb" {
		for weapon := range c.state.Player.Weapons {
			weapon := c.state.Player.Weapons[weapon]
			if weapon.State == "active" {
				// Fetch the localization file from Steam Database
				resp, err := http.Get("https://raw.githubusercontent.com/SteamDatabase/GameTracking-CS2/master/game/csgo/pak01_dir/resource/csgo_english.txt")
				if err != nil {
					return
				}
				defer resp.Body.Close()

				// Read the contents of the file
				bytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return
				}

				// Find the localization key for the current map
				removeprefix := regexp.MustCompile("^weapon_")
				re := regexp.MustCompile(`(?i)"SFUI_WPNHUD_` + removeprefix.ReplaceAllString(weapon.Name, "") + `"\s+"([^"]+)"`)
				match := re.FindSubmatch(bytes)
				if match == nil {
					return
				}

				c.activity.Details += " (" + string(match[1]) + ")"
			}
		}
	}
}

func (c *Connection) setMapMode() {
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
}

func (c *Connection) setMapName() {
	// Fetch the localization file from Steam Database
	resp, err := http.Get("https://raw.githubusercontent.com/SteamDatabase/GameTracking-CS2/master/game/csgo/pak01_dir/resource/csgo_english.txt")
	if err != nil {
		c.setMapWorkshopName()
		return
	}
	defer resp.Body.Close()

	// Read the contents of the file
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.setMapWorkshopName()
		return
	}

	// Find the localization key for the current map
	re := regexp.MustCompile(`"SFUI_Map_` + c.state.Map.Name + `"\s+"([^"]+)"`)
	match := re.FindSubmatch(bytes)
	if match == nil {
		c.setMapWorkshopName()
		return
	}

	mapLocalizedName := string(match[1])
	c.activity.LargeText = mapGameModeName + " | " + mapLocalizedName
}

func (c *Connection) setMapWorkshopName() {

	if strings.HasPrefix(c.state.Map.Name, "workshop/") {

		// URL to be converted to HTML
		// Make a GET request to the URL
		resp, err := http.Get(workshopLink)
		if err != nil {
			c.setMapNonLocalizedName()
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.setMapNonLocalizedName()
		}

		// Convert the response body to a string
		html := string(body)

		// Define the regular expression to find text between two words
		re := regexp.MustCompile(`<div class="workshopItemTitle">(.*?)</div>`)

		// Find all matches of the regular expression in the HTML
		matches := re.FindAllStringSubmatch(html, -1)

		// Print the text between the two words
		for _, match := range matches {
			c.activity.LargeText = mapGameModeName + " | Workshop | " + (match[1])
		}
	} else {
		c.setMapNonLocalizedName()
	}
}

func (c *Connection) setMapNonLocalizedName() {
	c.activity.LargeText = mapGameModeName + " | " + c.state.Map.Name
}

func (c *Connection) setMapWorkshopImage() {
	// URL to be converted to HTML
	// Make a GET request to the URL
	resp, err := http.Get(workshopLink)
	if err != nil {
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	}

	// Convert the response body to a string
	html := string(body)

	// Define the regular expression to find text between two words
	re := regexp.MustCompile(`onclick="ShowEnlargedImagePreview\( '(.*?)\?imw=5000&imh=5000&ima=fit&impolicy=Letterbox&imcolor=%23000000&letterbox=false`)

	// Find all matches of the regular expression in the HTML
	matches := re.FindAllStringSubmatch(html, -1)

	// Print the text between the two words
	for _, match := range matches {
		c.activity.LargeImage = (match[1])
	}
}

func (c *Connection) setWorkshopLink() {
	workshopID := regexp.MustCompile("").FindStringSubmatch("")
	if strings.HasPrefix(c.state.Map.Name, "workshop/") {
		workshopID = regexp.MustCompile(`workshop/(.+?)/`).FindStringSubmatch(c.state.Map.Name)
		//		fmt.Println("Workshop ID is" + workshopID[1])
		workshopLink = "https://steamcommunity.com/sharedfiles/filedetails/?id=" + workshopID[1]
	} else {
		workshopLink = ""
		//		fmt.Println("No Workshop Map")
	}
}

func (c *Connection) setButtons() {

	joinButtonEnable := 1
	// URL to be converted to HTML
	url := "https://steamcommunity.com/profiles/" + c.state.Provider.SteamId

	// Make a GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		joinButtonEnable = 0
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		joinButtonEnable = 0
	}

	// Convert the response body to a string
	html := string(body)

	// Define the regular expression to find text between two words
	re := regexp.MustCompile(`<a href="steam://joinlobby/730/(.*?)" class="btn_green_white_innerfade btn_small_thin">`)

	// Find all matches of the regular expression in the HTML
	matches := re.FindAllStringSubmatch(html, -1)

	if len(matches) == 0 {
		joinButtonEnable = 0
	}

	joinButton := &client.Button{}
	workshopButton := &client.Button{}
	html = ""
	re = regexp.MustCompile(``)

	// Print the text between the two words
	for _, match := range matches {
		joinButton = &client.Button{
			Label: "Join Game",
			Url:   "steam://joinlobby/730/" + match[1],
		}
	}
	workshopButton = &client.Button{
		Label: "Workshop",
		Url:   workshopLink,
	}

	c.activity.Buttons = []*client.Button{}
	buttons = []*client.Button{}

	if joinButtonEnable == 1 {
		buttons = append(buttons, joinButton)
	}
	if workshopLink != "" {
		buttons = append(buttons, workshopButton)
	}
}
