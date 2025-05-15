package mlb

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL = "https://statsapi.mlb.com/api/v1"
)

// Position represents a player's position
type Position struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Abbreviation string `json:"abbreviation"`
}

// BatSide represents which side a player bats from
type BatSide struct {
	Code         string `json:"code"`
	Description  string `json:"description"`
}

// ThrowSide represents which hand a player throws with
type ThrowSide struct {
	Code         string `json:"code"`
	Description  string `json:"description"`
}

// Player represents an MLB player data structure
type Player struct {
	ID           int       `json:"id"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	PrimaryPosition Position  `json:"primaryPosition,omitempty"`
	TeamName     string    `json:"teamName,omitempty"`
	BatSide      BatSide   `json:"batSide,omitempty"`
	ThrowSide    ThrowSide `json:"throwSide,omitempty"`
	BirthCountry string    `json:"birthCountry,omitempty"`
	Height       string    `json:"height,omitempty"`
	Weight       int       `json:"weight,omitempty"`
	Jersey       string    `json:"jerseyNumber,omitempty"`
	Status       string    `json:"status,omitempty"`
}

// RosterResponse represents the response from the MLB Stats API roster endpoint
type RosterResponse struct {
	Roster []struct {
		Person struct {
			ID int `json:"id"`
		} `json:"person"`
	} `json:"roster"`
}

// PlayerResponse represents the response from the MLB Stats API player endpoint
type PlayerResponse struct {
	People []Player `json:"people"`
}

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

// GetRandomPlayer retrieves data for a random player
func (c *Client) GetRandomPlayer() (*Player, error) {
	// Get a list of active players from a team roster (using Oakland Athletics)
	url := fmt.Sprintf("%s/teams/133/roster/active", baseURL)
	
	fmt.Printf("Requesting URL: %s\n", url)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var response RosterResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	if len(response.Roster) == 0 {
		return nil, fmt.Errorf("no players found")
	}

	// Select a random player
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(response.Roster))
	playerID := response.Roster[randomIndex].Person.ID
	
	// Get detailed player info
	return c.GetPlayerByID(playerID)
}

// GetPlayerByID retrieves player data by ID
func (c *Client) GetPlayerByID(playerID int) (*Player, error) {
	url := fmt.Sprintf("%s/people/%d", baseURL, playerID)
	
	fmt.Printf("Requesting URL: %s\n", url)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var response PlayerResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	if len(response.People) == 0 {
		return nil, fmt.Errorf("player not found")
	}

	return &response.People[0], nil
}

// GetPlayerByName retrieves player data by name
func (c *Client) GetPlayerByName(name string) ([]Player, error) {
	// Map of famous players and their IDs
	famousPlayers := map[string]int{
		"trout":    545361, // Mike Trout
		"judge":    592450, // Aaron Judge
		"ohtani":   660271, // Shohei Ohtani
		"kershaw":  477132, // Clayton Kershaw
		"betts":    605141, // Mookie Betts
		"freeman":  518692, // Freddie Freeman
		"harper":   547180, // Bryce Harper
		"verlander": 434378, // Justin Verlander
		"cole":     543037, // Gerrit Cole
		"scherzer": 453286, // Max Scherzer
	}

	// Check if the search term matches a famous player
	lowerName := strings.ToLower(name)
	for playerName, playerID := range famousPlayers {
		if strings.Contains(playerName, lowerName) || strings.Contains(lowerName, playerName) {
			player, err := c.GetPlayerByID(playerID)
			if err == nil {
				return []Player{*player}, nil
			}
		}
	}

	// If no famous player matched, try the team roster approach
	// Get Oakland Athletics roster as a fallback
	url := fmt.Sprintf("%s/teams/133/roster/active", baseURL)
	
	fmt.Printf("Requesting URL: %s\n", url)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var rosterResponse RosterResponse
	if err := json.Unmarshal(body, &rosterResponse); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	// Look for players matching the search term
	var matchingPlayers []Player
	for _, rosterEntry := range rosterResponse.Roster {
		player, err := c.GetPlayerByID(rosterEntry.Person.ID)
		if err != nil {
			continue
		}
		
		// Check if player name matches search term
		fullName := strings.ToLower(player.FirstName + " " + player.LastName)
		if strings.Contains(fullName, lowerName) {
			matchingPlayers = append(matchingPlayers, *player)
		}
	}

	return matchingPlayers, nil
}

// This function is no longer needed as we're using direct JSON unmarshaling
