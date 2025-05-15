package mlb

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	baseURL = "https://statsapi.mlb.com/api/v1"
)

// Client handles MLB API requests
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new MLB API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SearchPlayerResponse represents the API response for player search
type SearchPlayerResponse struct {
	SearchPlayerAll struct {
		CopyRight    string `json:"copyRight"`
		QueryResults struct {
			Created   string `json:"created"`
			TotalSize string `json:"totalSize"`
			Row       []struct {
				PlayerID             string `json:"player_id"`
				NameDisplayFirst     string `json:"name_first"`
				NameDisplayLast      string `json:"name_last"`
				NameDisplayFirstLast string `json:"name_display_first_last"`
				Position             string `json:"position"`
				TeamFull             string `json:"team_full"`
				TeamID               string `json:"team_id"`
				BirthCountry         string `json:"birth_country"`
				BirthDate            string `json:"birth_date"`
				Height               struct {
					Feet   string `json:"height_feet"`
					Inches string `json:"height_inches"`
				}
				Weight string `json:"weight"`
				Bats   string `json:"bats"`
				Throws string `json:"throws"`
			} `json:"row"`
		} `json:"queryResults"`
	} `json:"search_player_all"`
}

// PlayerInfoResponse represents the API response for player info
type PlayerInfoResponse struct {
	PlayerInfo struct {
		CopyRight    string `json:"copyRight"`
		QueryResults struct {
			Created   string `json:"created"`
			TotalSize string `json:"totalSize"`
			Row       struct {
				PlayerID             string `json:"player_id"`
				NameDisplayFirstLast string `json:"name_display_first_last"`
				Position             string `json:"primary_position_txt"`
				TeamName             string `json:"team_name"`
				BirthCountry         string `json:"birth_country"`
				BirthDate            string `json:"birth_date"`
				HeightFeet           string `json:"height_feet"`
				HeightInches         string `json:"height_inches"`
				Weight               string `json:"weight"`
				Bats                 string `json:"bats"`
				Throws               string `json:"throws"`
				Status               string `json:"status"`
				NickName             string `json:"name_nick"`
			} `json:"row"`
		} `json:"queryResults"`
	} `json:"player_info"`
}

// HittingStatsResponse represents the API response for hitting stats
type HittingStatsResponse struct {
	SportHittingTm struct {
		CopyRight    string `json:"copyRight"`
		QueryResults struct {
			Created   string `json:"created"`
			TotalSize string `json:"totalSize"`
			Row       struct {
				PlayerID string `json:"player_id"`
				Avg      string `json:"avg"`
				HR       string `json:"hr"`
				RBI      string `json:"rbi"`
				H        string `json:"h"`
				G        string `json:"g"`
			} `json:"row"`
		} `json:"queryResults"`
	} `json:"sport_hitting_tm"`
}

// PitchingStatsResponse represents the API response for pitching stats
type PitchingStatsResponse struct {
	SportPitchingTm struct {
		CopyRight    string `json:"copyRight"`
		QueryResults struct {
			Created   string `json:"created"`
			TotalSize string `json:"totalSize"`
			Row       struct {
				PlayerID string `json:"player_id"`
				ERA      string `json:"era"`
				W        string `json:"w"`
				L        string `json:"l"`
				SO       string `json:"so"`
				G        string `json:"g"`
			} `json:"row"`
		} `json:"queryResults"`
	} `json:"sport_pitching_tm"`
}

// StatsAPISearchResponse represents the MLB Stats API search response
type StatsAPISearchResponse struct {
	People []struct {
		ID              int    `json:"id"`
		FullName        string `json:"fullName"`
		FirstName       string `json:"firstName"`
		LastName        string `json:"lastName"`
		PrimaryNumber   string `json:"primaryNumber"`
		BirthDate       string `json:"birthDate"`
		CurrentAge      int    `json:"currentAge"`
		BirthCity       string `json:"birthCity"`
		BirthCountry    string `json:"birthCountry"`
		Height          string `json:"height"`
		Weight          int    `json:"weight"`
		Active          bool   `json:"active"`
		PrimaryPosition struct {
			Code         string `json:"code"`
			Name         string `json:"name"`
			Type         string `json:"type"`
			Abbreviation string `json:"abbreviation"`
		} `json:"primaryPosition"`
		BatSide struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"batSide"`
		PitchHand struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"pitchHand"`
		CurrentTeam struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"currentTeam,omitempty"`
	} `json:"people"`
}

// SearchPlayer searches for players by name
func (c *Client) SearchPlayer(name string) ([]Player, error) {
	// Build the URL
	searchURL := fmt.Sprintf("%s/people?sportId=1&names=%s", baseURL, url.QueryEscape(name))

	// Make the request
	resp, err := c.httpClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("error searching for player: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Debug: Print the raw response
	fmt.Println("Search URL:", searchURL)
	fmt.Println("Raw response:", string(body))

	// Parse the response
	var searchResp StatsAPISearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	// Check if any players were found
	if len(searchResp.People) == 0 {
		return []Player{}, nil
	}

	// Convert to Player objects
	var players []Player
	for _, person := range searchResp.People {
		// Parse birth date
		birthDate, _ := time.Parse("2006-01-02", person.BirthDate)

		// Create player object
		player := Player{
			PlayerID:     strconv.Itoa(person.ID),
			Name:         person.FullName,
			Position:     person.PrimaryPosition.Name,
			Team:         getTeamName(person),
			BirthDate:    birthDate,
			Height:       person.Height,
			Weight:       strconv.Itoa(person.Weight),
			BattingHand:  person.BatSide.Description,
			ThrowingHand: person.PitchHand.Description,
			Nationality:  person.BirthCountry,
		}

		players = append(players, player)
	}

	return players, nil
}

// getTeamName safely extracts the team name from a player
func getTeamName(person struct {
	ID              int    `json:"id"`
	FullName        string `json:"fullName"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	PrimaryNumber   string `json:"primaryNumber"`
	BirthDate       string `json:"birthDate"`
	CurrentAge      int    `json:"currentAge"`
	BirthCity       string `json:"birthCity"`
	BirthCountry    string `json:"birthCountry"`
	Height          string `json:"height"`
	Weight          int    `json:"weight"`
	Active          bool   `json:"active"`
	PrimaryPosition struct {
		Code         string `json:"code"`
		Name         string `json:"name"`
		Type         string `json:"type"`
		Abbreviation string `json:"abbreviation"`
	} `json:"primaryPosition"`
	BatSide struct {
		Code        string `json:"code"`
		Description string `json:"description"`
	} `json:"batSide"`
	PitchHand struct {
		Code        string `json:"code"`
		Description string `json:"description"`
	} `json:"pitchHand"`
	CurrentTeam struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"currentTeam,omitempty"`
}) string {
	if person.CurrentTeam.Name != "" {
		return person.CurrentTeam.Name
	}
	return "Free Agent"
}

// GetRandomPlayer fetches a random active MLB player
func (c *Client) GetRandomPlayer() (*Player, error) {
	// Get a list of common last names to search
	lastNames := []string{"Smith", "Johnson", "Williams", "Jones", "Brown", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez"}

	// Pick a random last name
	rand.Seed(time.Now().UnixNano())
	randomName := lastNames[rand.Intn(len(lastNames))]

	// Search for players with that last name
	players, err := c.SearchPlayer(randomName)
	if err != nil {
		return nil, err
	}

	if len(players) == 0 {
		return nil, fmt.Errorf("no players found with last name %s", randomName)
	}

	// Pick a random player from the results
	randomPlayer := players[rand.Intn(len(players))]

	// Get detailed player info
	return c.GetPlayerDetails(randomPlayer.PlayerID)
}

// StatsAPIPlayerResponse represents the MLB Stats API player details response
type StatsAPIPlayerResponse struct {
	People []struct {
		ID              int    `json:"id"`
		FullName        string `json:"fullName"`
		FirstName       string `json:"firstName"`
		LastName        string `json:"lastName"`
		PrimaryNumber   string `json:"primaryNumber"`
		BirthDate       string `json:"birthDate"`
		CurrentAge      int    `json:"currentAge"`
		BirthCity       string `json:"birthCity"`
		BirthCountry    string `json:"birthCountry"`
		Height          string `json:"height"`
		Weight          int    `json:"weight"`
		Active          bool   `json:"active"`
		PrimaryPosition struct {
			Code         string `json:"code"`
			Name         string `json:"name"`
			Type         string `json:"type"`
			Abbreviation string `json:"abbreviation"`
		} `json:"primaryPosition"`
		BatSide struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"batSide"`
		PitchHand struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"pitchHand"`
		CurrentTeam struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"currentTeam,omitempty"`
		Stats []struct {
			Type struct {
				DisplayName string `json:"displayName"`
			} `json:"type"`
			Group struct {
				DisplayName string `json:"displayName"`
			} `json:"group"`
			Stats map[string]interface{} `json:"stats"`
		} `json:"stats,omitempty"`
	} `json:"people"`
}

// GetPlayerDetails fetches detailed information for a player
func (c *Client) GetPlayerDetails(playerID string) (*Player, error) {
	// Fetch basic player info with stats
	infoURL := fmt.Sprintf("%s/people/%s?hydrate=stats(group=[hitting,pitching],type=[yearByYear])", baseURL, playerID)

	resp, err := c.httpClient.Get(infoURL)
	if err != nil {
		return nil, fmt.Errorf("error getting player details: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Debug: Print the raw response
	fmt.Println("Player Details URL:", infoURL)
	fmt.Println("Raw response:", string(body))

	// Parse the response
	var playerResp StatsAPIPlayerResponse
	if err := json.Unmarshal(body, &playerResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	// Check if player was found
	if len(playerResp.People) == 0 {
		return nil, fmt.Errorf("player not found with ID: %s", playerID)
	}

	// Extract player info
	person := playerResp.People[0]

	// Parse birth date
	birthDate, _ := time.Parse("2006-01-02", person.BirthDate)

	// Create player object
	player := &Player{
		PlayerID:     strconv.Itoa(person.ID),
		Name:         person.FullName,
		Position:     person.PrimaryPosition.Name,
		Team:         getPlayerTeamName(person),
		BirthDate:    birthDate,
		Height:       person.Height,
		Weight:       strconv.Itoa(person.Weight),
		BattingHand:  person.BatSide.Description,
		ThrowingHand: person.PitchHand.Description,
		Nationality:  person.BirthCountry,
	}

	// Extract stats if available
	isPitcher := isPitcher(person.PrimaryPosition.Name)
	stats, err := c.extractPlayerStats(person, isPitcher)
	if err != nil {
		// Don't fail if stats can't be extracted, just log the error
		fmt.Printf("Warning: Could not extract stats for player %s: %v\n", playerID, err)
	} else {
		player.Stats = stats
	}

	return player, nil
}

// extractPlayerStats extracts stats from the player response
func (c *Client) extractPlayerStats(person struct {
	ID              int    `json:"id"`
	FullName        string `json:"fullName"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	PrimaryNumber   string `json:"primaryNumber"`
	BirthDate       string `json:"birthDate"`
	CurrentAge      int    `json:"currentAge"`
	BirthCity       string `json:"birthCity"`
	BirthCountry    string `json:"birthCountry"`
	Height          string `json:"height"`
	Weight          int    `json:"weight"`
	Active          bool   `json:"active"`
	PrimaryPosition struct {
		Code         string `json:"code"`
		Name         string `json:"name"`
		Type         string `json:"type"`
		Abbreviation string `json:"abbreviation"`
	} `json:"primaryPosition"`
	BatSide struct {
		Code        string `json:"code"`
		Description string `json:"description"`
	} `json:"batSide"`
	PitchHand struct {
		Code        string `json:"code"`
		Description string `json:"description"`
	} `json:"pitchHand"`
	CurrentTeam struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"currentTeam,omitempty"`
	Stats []struct {
		Type struct {
			DisplayName string `json:"displayName"`
		} `json:"type"`
		Group struct {
			DisplayName string `json:"displayName"`
		} `json:"group"`
		Stats map[string]interface{} `json:"stats"`
	} `json:"stats,omitempty"`
}, isPitcher bool) (PlayerStats, error) {
	var stats PlayerStats

	// Check if stats are available
	if len(person.Stats) == 0 {
		return stats, fmt.Errorf("no stats available for player")
	}

	// Find the most recent year's stats
	var relevantStats map[string]interface{}
	for _, statGroup := range person.Stats {
		if isPitcher && statGroup.Group.DisplayName == "pitching" {
			relevantStats = statGroup.Stats
			break
		} else if !isPitcher && statGroup.Group.DisplayName == "hitting" {
			relevantStats = statGroup.Stats
			break
		}
	}

	if relevantStats == nil {
		return stats, fmt.Errorf("no relevant stats found for player")
	}

	// Extract stats based on player type
	if isPitcher {
		// Extract pitching stats
		if gamesPlayed, ok := relevantStats["gamesPlayed"]; ok {
			stats.GamesPlayed = int(gamesPlayed.(float64))
		}
		if era, ok := relevantStats["era"]; ok {
			stats.ERA = era.(float64)
		}
		if wins, ok := relevantStats["wins"]; ok {
			stats.Wins = int(wins.(float64))
		}
		if losses, ok := relevantStats["losses"]; ok {
			stats.Losses = int(losses.(float64))
		}
		if strikeouts, ok := relevantStats["strikeOuts"]; ok {
			stats.Strikeouts = int(strikeouts.(float64))
		}
	} else {
		// Extract hitting stats
		if gamesPlayed, ok := relevantStats["gamesPlayed"]; ok {
			stats.GamesPlayed = int(gamesPlayed.(float64))
		}
		if avg, ok := relevantStats["avg"]; ok {
			stats.BattingAvg = avg.(float64)
		}
		if homeRuns, ok := relevantStats["homeRuns"]; ok {
			stats.HomeRuns = int(homeRuns.(float64))
		}
		if rbi, ok := relevantStats["rbi"]; ok {
			stats.RBI = int(rbi.(float64))
		}
		if hits, ok := relevantStats["hits"]; ok {
			stats.Hits = int(hits.(float64))
		}
	}

	return stats, nil
}

// GetPlayerStats fetches stats for a player
func (c *Client) GetPlayerStats(playerID string, isPitcher bool) (PlayerStats, error) {
	// This is now handled by GetPlayerDetails with the hydration parameter
	// We'll keep this method for backward compatibility
	player, err := c.GetPlayerDetails(playerID)
	if err != nil {
		return PlayerStats{}, err
	}
	return player.Stats, nil
}
