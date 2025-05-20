package mlb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	// MLB Stats API base URL
	baseURL = "https://statsapi.mlb.com/api/v1"
)

// Debug flag to enable/disable debug output
var Debug = true

// Client represents an MLB API client
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new MLB API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

// SearchPlayer searches for players by name using the MLB Stats API
func (c *Client) SearchPlayer(name string) ([]Player, error) {
	// The direct search endpoint doesn't work as expected, so we'll use the teams endpoint
	// and then filter players from all teams
	requestURL := fmt.Sprintf("%s/teams?sportId=1&hydrate=roster(person(firstName,lastName))", baseURL)

	// Make the request
	if Debug {
		fmt.Printf("Making request to: %s\n", requestURL)
	}

	resp, err := c.httpClient.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Debug: Print a portion of the response
	if Debug && len(body) > 0 {
		previewLen := 200
		if len(body) < previewLen {
			previewLen = len(body)
		}
		fmt.Printf("Response preview: %s\n", body[:previewLen])
	}

	// Check if the response is an error message
	if strings.Contains(string(body), "could not be found") || strings.Contains(string(body), "error") {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	// Parse the response
	var teamsResp struct {
		Teams []struct {
			Name         string `json:"name"`
			Abbreviation string `json:"abbreviation"`
			Roster       struct {
				Roster []struct {
					Person struct {
						ID        int    `json:"id"`
						FullName  string `json:"fullName"`
						FirstName string `json:"firstName"`
						LastName  string `json:"lastName"`
						Link      string `json:"link"`
					} `json:"person"`
					Position struct {
						Code         string `json:"code"`
						Name         string `json:"name"`
						Type         string `json:"type"`
						Abbreviation string `json:"abbreviation"`
					} `json:"position"`
					JerseyNumber string `json:"jerseyNumber"`
				} `json:"roster"`
			} `json:"roster"`
		} `json:"teams"`
	}

	if err := json.Unmarshal(body, &teamsResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	// Filter players by name (case-insensitive)
	nameLower := strings.ToLower(name)
	var players []Player

	for _, team := range teamsResp.Teams {
		for _, rosterPlayer := range team.Roster.Roster {
			// Check if player name contains the search term
			playerFullName := strings.ToLower(rosterPlayer.Person.FullName)
			playerFirstName := strings.ToLower(rosterPlayer.Person.FirstName)
			playerLastName := strings.ToLower(rosterPlayer.Person.LastName)

			if strings.Contains(playerFullName, nameLower) ||
				strings.Contains(playerFirstName, nameLower) ||
				strings.Contains(playerLastName, nameLower) {

				player := Player{
					PlayerID:     fmt.Sprintf("%d", rosterPlayer.Person.ID),
					Name:         rosterPlayer.Person.FullName,
					Position:     rosterPlayer.Position.Name,
					Team:         team.Name,
					TeamAbbrev:   team.Abbreviation,
					JerseyNumber: rosterPlayer.JerseyNumber,
					Active:       "Y", // All players in the roster are active
				}
				players = append(players, player)
			}
		}
	}

	return players, nil
}

// GetPlayerInfo gets detailed information for a player by ID
func (c *Client) GetPlayerInfo(playerID string) (*PlayerInfo, error) {
	// Build the URL for the MLB Stats API
	requestURL := fmt.Sprintf("%s/people/%s?hydrate=currentTeam", baseURL, strings.TrimSpace(playerID))

	// Make the request
	if Debug {
		fmt.Printf("Making request to: %s\n", requestURL)
	}

	resp, err := c.httpClient.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Debug: Print a portion of the response
	if Debug && len(body) > 0 {
		previewLen := 200
		if len(body) < previewLen {
			previewLen = len(body)
		}
		fmt.Printf("Response preview: %s\n", body[:previewLen])
	}

	// Check if the response is an error message
	if strings.Contains(string(body), "could not be found") || strings.Contains(string(body), "error") {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	// Parse the response
	var infoResp struct {
		People []struct {
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
			CurrentTeam        struct {
				ID           int    `json:"id"`
				Name         string `json:"name"`
				Abbreviation string `json:"abbreviation"`
			} `json:"currentTeam"`
			Position struct {
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
			MlbDebutDate string `json:"mlbDebutDate"`
		} `json:"people"`
	}

	if err := json.Unmarshal(body, &infoResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	if len(infoResp.People) == 0 {
		return nil, fmt.Errorf("no player information found")
	}

	// Convert to our PlayerInfo struct
	p := infoResp.People[0]
	playerInfo := &PlayerInfo{
		PlayerID:     fmt.Sprintf("%d", p.ID),
		Name:         p.FullName,
		Position:     p.Position.Name,
		Team:         p.CurrentTeam.Name,
		TeamAbbrev:   p.CurrentTeam.Abbreviation,
		BatSide:      p.BatSide.Description,
		ThrowSide:    p.PitchHand.Description,
		Weight:       fmt.Sprintf("%d", p.Weight),
		Height:       p.Height,
		BirthDate:    p.BirthDate,
		BirthCity:    p.BirthCity,
		BirthState:   p.BirthStateProvince,
		BirthCountry: p.BirthCountry,
		Age:          fmt.Sprintf("%d", p.CurrentAge),
		Status:       fmt.Sprintf("%t", p.Active),
		Active:       fmt.Sprintf("%t", p.Active),
		JerseyNumber: p.PrimaryNumber,
		ProDebutDate: p.MlbDebutDate,
	}

	return playerInfo, nil
}

// GetCareerHittingStats gets career hitting stats for a player by ID
func (c *Client) GetCareerHittingStats(playerID string) (*HittingStats, error) {
	// Build the URL for the MLB Stats API
	requestURL := fmt.Sprintf("%s/people/%s/stats?stats=career&group=hitting", baseURL, strings.TrimSpace(playerID))

	// Make the request
	if Debug {
		fmt.Printf("Making request to: %s\n", requestURL)
	}

	resp, err := c.httpClient.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Debug: Print a portion of the response
	if Debug && len(body) > 0 {
		previewLen := 200
		if len(body) < previewLen {
			previewLen = len(body)
		}
		fmt.Printf("Response preview: %s\n", body[:previewLen])
	}

	// Check if the response is an error message
	if strings.Contains(string(body), "could not be found") || strings.Contains(string(body), "error") {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	// Parse the response
	var statsResp struct {
		Stats []struct {
			Group struct {
				DisplayName string `json:"displayName"`
			} `json:"group"`
			Type struct {
				DisplayName string `json:"displayName"`
			} `json:"type"`
			Splits []struct {
				Stat struct {
					Avg     interface{} `json:"avg"`
					HR      interface{} `json:"homeRuns"`
					RBI     interface{} `json:"rbi"`
					Runs    interface{} `json:"runs"`
					SB      interface{} `json:"stolenBases"`
					OBP     interface{} `json:"obp"`
					SLG     interface{} `json:"slg"`
					OPS     interface{} `json:"ops"`
					Hits    interface{} `json:"hits"`
					Doubles interface{} `json:"doubles"`
					Triples interface{} `json:"triples"`
					BB      interface{} `json:"baseOnBalls"`
					SO      interface{} `json:"strikeOuts"`
					Games   interface{} `json:"gamesPlayed"`
					AB      interface{} `json:"atBats"`
				} `json:"stat"`
			} `json:"splits"`
		} `json:"stats"`
	}

	if err := json.Unmarshal(body, &statsResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	// Check if any stats were found
	if len(statsResp.Stats) == 0 || len(statsResp.Stats[0].Splits) == 0 {
		return nil, nil
	}

	// Convert to our HittingStats struct
	stats := statsResp.Stats[0].Splits[0].Stat
	hittingStats := &HittingStats{
		AVG:     fmt.Sprintf("%v", stats.Avg),
		HR:      fmt.Sprintf("%v", stats.HR),
		RBI:     fmt.Sprintf("%v", stats.RBI),
		Runs:    fmt.Sprintf("%v", stats.Runs),
		SB:      fmt.Sprintf("%v", stats.SB),
		OBP:     fmt.Sprintf("%v", stats.OBP),
		SLG:     fmt.Sprintf("%v", stats.SLG),
		OPS:     fmt.Sprintf("%v", stats.OPS),
		Hits:    fmt.Sprintf("%v", stats.Hits),
		Doubles: fmt.Sprintf("%v", stats.Doubles),
		Triples: fmt.Sprintf("%v", stats.Triples),
		BB:      fmt.Sprintf("%v", stats.BB),
		SO:      fmt.Sprintf("%v", stats.SO),
		Games:   fmt.Sprintf("%v", stats.Games),
		AB:      fmt.Sprintf("%v", stats.AB),
	}

	return hittingStats, nil
}

// GetCareerPitchingStats gets career pitching stats for a player by ID
func (c *Client) GetCareerPitchingStats(playerID string) (*PitchingStats, error) {
	// Build the URL for the MLB Stats API
	requestURL := fmt.Sprintf("%s/people/%s/stats?stats=career&group=pitching", baseURL, strings.TrimSpace(playerID))

	// Make the request
	if Debug {
		fmt.Printf("Making request to: %s\n", requestURL)
	}

	resp, err := c.httpClient.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Debug: Print a portion of the response
	if Debug && len(body) > 0 {
		previewLen := 200
		if len(body) < previewLen {
			previewLen = len(body)
		}
		fmt.Printf("Response preview: %s\n", body[:previewLen])
	}

	// Check if the response is an error message
	if strings.Contains(string(body), "could not be found") || strings.Contains(string(body), "error") {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	// Parse the response
	var statsResp struct {
		Stats []struct {
			Group struct {
				DisplayName string `json:"displayName"`
			} `json:"group"`
			Type struct {
				DisplayName string `json:"displayName"`
			} `json:"type"`
			Splits []struct {
				Stat struct {
					ERA    interface{} `json:"era"`
					Wins   interface{} `json:"wins"`
					Losses interface{} `json:"losses"`
					WHIP   interface{} `json:"whip"`
					SO     interface{} `json:"strikeOuts"`
					BB     interface{} `json:"baseOnBalls"`
					SV     interface{} `json:"saves"`
					IP     interface{} `json:"inningsPitched"`
					Games  interface{} `json:"gamesPlayed"`
					GS     interface{} `json:"gamesStarted"`
					CG     interface{} `json:"completeGames"`
					SHO    interface{} `json:"shutouts"`
				} `json:"stat"`
			} `json:"splits"`
		} `json:"stats"`
	}

	if err := json.Unmarshal(body, &statsResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	// Check if any stats were found
	if len(statsResp.Stats) == 0 || len(statsResp.Stats[0].Splits) == 0 {
		return nil, nil
	}

	// Convert to our PitchingStats struct
	stats := statsResp.Stats[0].Splits[0].Stat
	pitchingStats := &PitchingStats{
		ERA:    fmt.Sprintf("%v", stats.ERA),
		Wins:   fmt.Sprintf("%v", stats.Wins),
		Losses: fmt.Sprintf("%v", stats.Losses),
		WHIP:   fmt.Sprintf("%v", stats.WHIP),
		SO:     fmt.Sprintf("%v", stats.SO),
		BB:     fmt.Sprintf("%v", stats.BB),
		SV:     fmt.Sprintf("%v", stats.SV),
		IP:     fmt.Sprintf("%v", stats.IP),
		Games:  fmt.Sprintf("%v", stats.Games),
		GS:     fmt.Sprintf("%v", stats.GS),
		CG:     fmt.Sprintf("%v", stats.CG),
		SHO:    fmt.Sprintf("%v", stats.SHO),
	}

	return pitchingStats, nil
}

// GetCareerStats gets both hitting and pitching career stats for a player
func (c *Client) GetCareerStats(playerID string) (*CareerStats, error) {
	stats := &CareerStats{}

	// Get hitting stats
	hittingStats, err := c.GetCareerHittingStats(playerID)
	if err != nil {
		return nil, fmt.Errorf("error getting hitting stats: %v", err)
	}
	stats.Hitting = hittingStats

	// Get pitching stats
	pitchingStats, err := c.GetCareerPitchingStats(playerID)
	if err != nil {
		return nil, fmt.Errorf("error getting pitching stats: %v", err)
	}
	stats.Pitching = pitchingStats

	return stats, nil
}

// FormatPlayerInfo formats player info for display
func FormatPlayerInfo(info *PlayerInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Player: %s (#%s)\n", info.Name, info.JerseyNumber))
	sb.WriteString(fmt.Sprintf("Team: %s (%s)\n", info.Team, info.TeamAbbrev))
	sb.WriteString(fmt.Sprintf("Position: %s\n", info.Position))
	sb.WriteString(fmt.Sprintf("Bats: %s, Throws: %s\n", info.BatSide, info.ThrowSide))
	sb.WriteString(fmt.Sprintf("Height: %s, Weight: %s lbs\n", info.Height, info.Weight))
	sb.WriteString(fmt.Sprintf("Born: %s (%s years old)\n", formatDate(info.BirthDate), info.Age))
	sb.WriteString(fmt.Sprintf("Birthplace: %s, %s, %s\n", info.BirthCity, info.BirthState, info.BirthCountry))

	if info.ProDebutDate != "" {
		sb.WriteString(fmt.Sprintf("MLB Debut: %s\n", formatDate(info.ProDebutDate)))
	}

	if info.College != "" {
		sb.WriteString(fmt.Sprintf("College: %s\n", info.College))
	}

	if info.HighSchool != "" {
		sb.WriteString(fmt.Sprintf("High School: %s\n", info.HighSchool))
	}

	if info.TwitterID != "" {
		sb.WriteString(fmt.Sprintf("Twitter: %s\n", info.TwitterID))
	}

	return sb.String()
}

// FormatCareerStats formats career stats for display
func FormatCareerStats(stats *CareerStats) string {
	var sb strings.Builder

	sb.WriteString("Career Statistics:\n")

	// Format hitting stats if available
	if stats.Hitting != nil {
		sb.WriteString("\nBatting Statistics:\n")
		sb.WriteString(fmt.Sprintf("Games: %s, At Bats: %s\n", stats.Hitting.Games, stats.Hitting.AB))
		sb.WriteString(fmt.Sprintf("Slash Line: %s/%s/%s (%s OPS)\n",
			stats.Hitting.AVG, stats.Hitting.OBP, stats.Hitting.SLG, stats.Hitting.OPS))
		sb.WriteString(fmt.Sprintf("Hits: %s (2B: %s, 3B: %s, HR: %s)\n",
			stats.Hitting.Hits, stats.Hitting.Doubles, stats.Hitting.Triples, stats.Hitting.HR))
		sb.WriteString(fmt.Sprintf("Runs: %s, RBI: %s, SB: %s\n",
			stats.Hitting.Runs, stats.Hitting.RBI, stats.Hitting.SB))
		sb.WriteString(fmt.Sprintf("BB: %s, SO: %s\n", stats.Hitting.BB, stats.Hitting.SO))
	}

	// Format pitching stats if available
	if stats.Pitching != nil {
		sb.WriteString("\nPitching Statistics:\n")
		sb.WriteString(fmt.Sprintf("Record: %s-%s, ERA: %s\n",
			stats.Pitching.Wins, stats.Pitching.Losses, stats.Pitching.ERA))
		sb.WriteString(fmt.Sprintf("Games: %s (Started: %s, Complete: %s, Shutouts: %s)\n",
			stats.Pitching.Games, stats.Pitching.GS, stats.Pitching.CG, stats.Pitching.SHO))
		sb.WriteString(fmt.Sprintf("IP: %s, WHIP: %s\n", stats.Pitching.IP, stats.Pitching.WHIP))
		sb.WriteString(fmt.Sprintf("SO: %s, BB: %s\n", stats.Pitching.SO, stats.Pitching.BB))
		sb.WriteString(fmt.Sprintf("Saves: %s\n", stats.Pitching.SV))
	}

	return sb.String()
}

// formatDate formats a date string from the API (YYYY-MM-DDT00:00:00) to a more readable format (YYYY-MM-DD)
func formatDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	parts := strings.Split(dateStr, "T")
	if len(parts) > 0 {
		return parts[0]
	}

	return dateStr
}
