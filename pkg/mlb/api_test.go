package mlb

import (
	"testing"
)

func TestSearchPlayer(t *testing.T) {
	client := NewClient()

	// Test with a common name that should return results
	players, err := client.SearchPlayer("Smith")
	if err != nil {
		t.Fatalf("Error searching for player: %v", err)
	}
	if len(players) == 0 {
		t.Errorf("Expected to find players with last name Smith, but found none")
	}

	// Test with a specific player that should exist
	players, err = client.SearchPlayer("Trout")
	if err != nil {
		t.Fatalf("Error searching for player: %v", err)
	}
	if len(players) == 0 {
		t.Errorf("Expected to find Mike Trout, but found no players")
	}

	// Test with a name that shouldn't exist
	players, err = client.SearchPlayer("XYZNonExistentPlayer")
	if err != nil {
		t.Fatalf("Error searching for player: %v", err)
	}
	if len(players) != 0 {
		t.Errorf("Expected to find no players with fake name, but found %d", len(players))
	}
}

func TestGetPlayerDetails(t *testing.T) {
	client := NewClient()

	// Test with Mike Trout's player ID (known player)
	// Note: If API changes, this ID might need to be updated
	troutID := "545361"
	player, err := client.GetPlayerDetails(troutID)
	if err != nil {
		t.Fatalf("Error getting player details: %v", err)
	}

	// Verify basic player info
	if player.Name == "" {
		t.Errorf("Expected player name to be populated")
	}
	if player.Team == "" {
		t.Errorf("Expected player team to be populated")
	}
	if player.Position == "" {
		t.Errorf("Expected player position to be populated")
	}

	// Test with invalid player ID
	_, err = client.GetPlayerDetails("999999999")
	if err == nil {
		t.Errorf("Expected error for invalid player ID, but got none")
	}
}

func TestGetRandomPlayer(t *testing.T) {
	client := NewClient()

	// Test getting a random player
	player, err := client.GetRandomPlayer()
	if err != nil {
		t.Fatalf("Error getting random player: %v", err)
	}

	// Verify player data is populated
	if player.PlayerID == "" {
		t.Errorf("Expected player ID to be populated")
	}
	if player.Name == "" {
		t.Errorf("Expected player name to be populated")
	}
}
