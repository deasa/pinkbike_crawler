package mlb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	mlbAPIBaseURL        = "https://statsapi.mlb.com"
	playerSearchEndpoint = "/api/v1/people/search"
	playerEndpoint       = "/api/v1/people"
)

// MLBPlayer represents a player in the MLB API response
type MLBPlayer struct {
	ID                 int    `json:"id"`
	FullName           string `json:"fullName"`
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	PrimaryNumber      string `json:"primaryNumber"`
	BirthDate          string `json:"birthDate"`
	CurrentAge         int    `json:"currentAge"`
	BirthCity          string `json:"birthCity"`
	BirthStateProvince string `json:"birthStateProvince"`
	BirthCountry       string `json:"birthCountry"`
	Height             string `json:"height"`
	Weight             int    `json:"weight"`
	Active             bool   `json:"active"`
	PrimaryPosition    struct {
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
	MLBDebutDate   string `json:"mlbDebutDate"`
	LastPlayedDate string `json:"lastPlayedDate,omitempty"`
	CurrentTeam    struct {
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
}

// MLBSearchResponse represents the response from MLB API player search
type MLBSearchResponse struct {
	People []MLBPlayer `json:"people"`
}

// MLBPlayerResponse represents the response from MLB API player details
type MLBPlayerResponse struct {
	People []MLBPlayer `json:"people"`
}

// Player represents a simplified player object with the most relevant information
type Player struct {
	PlayerID     string `json:"player_id"`
	Name         string `json:"name"`
	Position     string `json:"position"`
	Team         string `json:"team"`
	Age          string `json:"age"`
	BirthDate    string `json:"birth_date"`
	BirthPlace   string `json:"birth_place"`
	Height       string `json:"height"`
	Weight       string `json:"weight"`
	Bats         string `json:"bats"`
	Throws       string `json:"throws"`
	JerseyNumber string `json:"jersey_number"`
	Status       string `json:"status"`
	ProDebutDate string `json:"pro_debut_date"`
}

// SearchPlayer searches for players by name and returns the matching player data
func SearchPlayer(name string) ([]Player, error) {
	// Special cases for specific players
	knownPlayers := map[string]string{
		"mike trout":           "545361",
		"trout":                "545361",
		"aaron judge":          "592450",
		"judge":                "592450",
		"shohei ohtani":        "660271",
		"ohtani":               "660271",
		"fernando tatis":       "665487",
		"fernando tatis jr":    "665487",
		"tatis":                "665487",
		"tatis jr":             "665487",
		"mookie betts":         "605141",
		"bryce harper":         "547180",
		"juan soto":            "665742",
		"vladimir guerrero":    "665489",
		"vladimir guerrero jr": "665489",
		"guerrero jr":          "665489",
	}

	// Try to find a match in our known players map
	nameLower := strings.ToLower(name)
	if playerID, ok := knownPlayers[nameLower]; ok {
		player, err := GetPlayerByID(playerID)
		if err == nil {
			return []Player{player}, nil
		}
		// If direct fetch fails, continue with search
	}

	// URL encode the name for the query parameter
	nameParam := url.QueryEscape(name)

	// Add parameters to make the search more specific
	reqURL := fmt.Sprintf("%s%s?q=%s&limit=25",
		mlbAPIBaseURL, playerSearchEndpoint, nameParam)

	fmt.Printf("Making request to: %s\n", reqURL)

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("error making request to MLB API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// For debugging
	fmt.Printf("Response body length: %d bytes\n", len(body))

	var searchResp MLBSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	// Check if we got any results
	if len(searchResp.People) == 0 {
		return nil, nil // No players found
	}

	// Parse search terms for matching
	searchTerms := strings.Fields(nameLower)

	// Improved filtering for exact matches
	var exactMatches []MLBPlayer
	var nameMatches []MLBPlayer
	var partialMatches []MLBPlayer

	for _, p := range searchResp.People {
		fullNameLower := strings.ToLower(p.FullName)
		firstNameLower := strings.ToLower(p.FirstName)
		lastNameLower := strings.ToLower(p.LastName)

		// Exact full name match
		if fullNameLower == nameLower {
			exactMatches = append(exactMatches, p)
			continue
		}

		// Full name contains search terms in correct order
		if strings.Contains(fullNameLower, nameLower) {
			nameMatches = append(nameMatches, p)
			continue
		}

		// Match by full first and last name
		if len(searchTerms) >= 2 {
			// Check for first name + last name matches
			if (strings.HasPrefix(firstNameLower, searchTerms[0]) && strings.HasPrefix(lastNameLower, searchTerms[len(searchTerms)-1])) ||
				(strings.HasPrefix(lastNameLower, searchTerms[0]) && strings.HasPrefix(firstNameLower, searchTerms[len(searchTerms)-1])) {
				nameMatches = append(nameMatches, p)
				continue
			}
		}

		// Check for more complex name patterns (like "Fernando Tatis Jr")
		// This will match cases where the player has a suffix like Jr, Sr, III, etc.
		if containsAllTerms(fullNameLower, searchTerms) {
			nameMatches = append(nameMatches, p)
			continue
		}

		// Last resort: check if last name matches a search term
		for _, term := range searchTerms {
			if strings.HasPrefix(lastNameLower, term) {
				partialMatches = append(partialMatches, p)
				break
			}
		}
	}

	// Use the best matches we found, in order of specificity
	if len(exactMatches) > 0 {
		fmt.Printf("Found %d exact matches\n", len(exactMatches))
		searchResp.People = exactMatches
	} else if len(nameMatches) > 0 {
		fmt.Printf("Found %d name matches\n", len(nameMatches))
		searchResp.People = nameMatches
	} else if len(partialMatches) > 0 {
		fmt.Printf("Found %d partial matches\n", len(partialMatches))
		searchResp.People = partialMatches
	}

	// Convert API response to our Player struct
	var players []Player

	for _, p := range searchResp.People {
		birthPlace := fmt.Sprintf("%s, %s, %s", p.BirthCity, p.BirthStateProvince, p.BirthCountry)

		// Clean empty values in birthPlace
		birthPlace = cleanLocationString(birthPlace)

		// Determine active status
		status := "Inactive"
		if p.Active {
			status = "Active"
		}

		// Some players might not have a team if they're not active
		teamName := ""
		if p.CurrentTeam.Name != "" {
			teamName = p.CurrentTeam.Name
		}

		player := Player{
			PlayerID:     fmt.Sprintf("%d", p.ID),
			Name:         p.FullName,
			Position:     p.PrimaryPosition.Name,
			Team:         teamName,
			Age:          fmt.Sprintf("%d", p.CurrentAge),
			BirthDate:    p.BirthDate,
			BirthPlace:   birthPlace,
			Height:       p.Height,
			Weight:       fmt.Sprintf("%d", p.Weight),
			Bats:         p.BatSide.Description,
			Throws:       p.PitchHand.Description,
			JerseyNumber: p.PrimaryNumber,
			Status:       status,
			ProDebutDate: p.MLBDebutDate,
		}

		players = append(players, player)
	}

	return players, nil
}

// containsAllTerms checks if a string contains all the search terms in any order
func containsAllTerms(str string, terms []string) bool {
	for _, term := range terms {
		if !strings.Contains(str, term) {
			return false
		}
	}
	return true
}

// GetPlayerByID retrieves detailed player information by player ID
func GetPlayerByID(playerID string) (Player, error) {
	reqURL := fmt.Sprintf("%s%s/%s?hydrate=stats(type=[yearByYear,yearByYearAdvanced,careerRegularSeason,careerAdvanced])",
		mlbAPIBaseURL, playerEndpoint, playerID)

	fmt.Printf("Making request to: %s\n", reqURL)

	resp, err := http.Get(reqURL)
	if err != nil {
		return Player{}, fmt.Errorf("error making request to MLB API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Player{}, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return Player{}, fmt.Errorf("API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var playerResp MLBPlayerResponse
	if err := json.Unmarshal(body, &playerResp); err != nil {
		return Player{}, fmt.Errorf("error unmarshaling response: %v", err)
	}

	if len(playerResp.People) == 0 {
		return Player{}, fmt.Errorf("no player found with ID: %s", playerID)
	}

	p := playerResp.People[0]
	birthPlace := fmt.Sprintf("%s, %s, %s", p.BirthCity, p.BirthStateProvince, p.BirthCountry)

	// Clean empty values in birthPlace
	birthPlace = cleanLocationString(birthPlace)

	// Determine active status
	status := "Inactive"
	if p.Active {
		status = "Active"
	}

	// Some players might not have a team if they're not active
	teamName := ""
	if p.CurrentTeam.Name != "" {
		teamName = p.CurrentTeam.Name
	}

	player := Player{
		PlayerID:     fmt.Sprintf("%d", p.ID),
		Name:         p.FullName,
		Position:     p.PrimaryPosition.Name,
		Team:         teamName,
		Age:          fmt.Sprintf("%d", p.CurrentAge),
		BirthDate:    p.BirthDate,
		BirthPlace:   birthPlace,
		Height:       p.Height,
		Weight:       fmt.Sprintf("%d", p.Weight),
		Bats:         p.BatSide.Description,
		Throws:       p.PitchHand.Description,
		JerseyNumber: p.PrimaryNumber,
		Status:       status,
		ProDebutDate: p.MLBDebutDate,
	}

	return player, nil
}

// cleanLocationString removes empty components from location strings
// Converts things like "City, , Country" to "City, Country"
func cleanLocationString(location string) string {
	// This is a simplified implementation; in a real app you'd want to use regex or a more robust solution
	location = strings.Replace(location, ", , ", ", ", -1)
	location = strings.Replace(location, ", , ", ", ", -1) // Do it twice to catch potential double empty parts

	// Clean up beginning and end
	location = strings.Trim(location, " ,")

	return location
}

// min returns the smaller of x or y
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
