package mlb

import (
	"fmt"
	"time"
)

// Player represents basic MLB player information
type Player struct {
	PlayerID      string
	Name          string
	Position      string
	Team          string
	BirthDate     time.Time
	Height        string
	Weight        string
	BattingHand   string
	ThrowingHand  string
	Nationality   string
	Stats         PlayerStats
	DetailedStats map[string]interface{} // For additional stats
}

// PlayerStats represents common player statistics
type PlayerStats struct {
	// Common stats for all players
	GamesPlayed int

	// Batting stats
	BattingAvg float64
	HomeRuns   int
	RBI        int
	Hits       int

	// Pitching stats
	ERA        float64
	Wins       int
	Losses     int
	Strikeouts int
}

// String returns a formatted string representation of a player
func (p Player) String() string {
	result := fmt.Sprintf("Player: %s\n", p.Name)
	result += fmt.Sprintf("ID: %s\n", p.PlayerID)
	result += fmt.Sprintf("Team: %s\n", p.Team)
	result += fmt.Sprintf("Position: %s\n", p.Position)
	result += fmt.Sprintf("Born: %s\n", p.BirthDate.Format("January 2, 2006"))
	result += fmt.Sprintf("Height/Weight: %s / %s\n", p.Height, p.Weight)
	result += fmt.Sprintf("Bats/Throws: %s/%s\n", p.BattingHand, p.ThrowingHand)
	result += fmt.Sprintf("Nationality: %s\n", p.Nationality)

	// Display relevant stats based on position
	if isPitcher(p.Position) {
		result += fmt.Sprintf("\nPitching Stats:\n")
		result += fmt.Sprintf("ERA: %.2f\n", p.Stats.ERA)
		result += fmt.Sprintf("W-L: %d-%d\n", p.Stats.Wins, p.Stats.Losses)
		result += fmt.Sprintf("Strikeouts: %d\n", p.Stats.Strikeouts)
		result += fmt.Sprintf("Games: %d\n", p.Stats.GamesPlayed)
	} else {
		result += fmt.Sprintf("\nBatting Stats:\n")
		result += fmt.Sprintf("AVG: %.3f\n", p.Stats.BattingAvg)
		result += fmt.Sprintf("HR: %d\n", p.Stats.HomeRuns)
		result += fmt.Sprintf("RBI: %d\n", p.Stats.RBI)
		result += fmt.Sprintf("Hits: %d\n", p.Stats.Hits)
		result += fmt.Sprintf("Games: %d\n", p.Stats.GamesPlayed)
	}

	return result
}

// isPitcher determines if a position is a pitching position
func isPitcher(position string) bool {
	return position == "P" || position == "SP" || position == "RP" || position == "CL"
}
