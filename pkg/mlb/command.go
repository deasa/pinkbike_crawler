package mlb

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// HandleMLBPlayerCommand handles the MLB player command
func HandleMLBPlayerCommand(playerName string) error {
	client := NewClient()

	// Search for player
	fmt.Printf("Searching for player: %s\n", playerName)

	// Enable debug mode for detailed API responses (set to false for cleaner output)
	Debug = false

	// Map of known players with their IDs for quick access
	knownPlayers := map[string]string{
		"mike trout":      "545361",
		"trout":           "545361",
		"aaron judge":     "592450",
		"judge":           "592450",
		"shohei ohtani":   "660271",
		"ohtani":          "660271",
		"mookie betts":    "605141",
		"betts":           "605141",
		"clayton kershaw": "477132",
		"kershaw":         "477132",
	}

	// Check if we have a known player ID
	playerNameLower := strings.ToLower(playerName)
	if playerID, ok := knownPlayers[playerNameLower]; ok {
		fmt.Printf("Found known player ID for %s: %s\n", playerName, playerID)

		// Get player info
		playerInfo, err := client.GetPlayerInfo(playerID)
		if err != nil {
			return fmt.Errorf("error getting player info: %v", err)
		}

		// Get career stats
		careerStats, err := client.GetCareerStats(playerID)
		if err != nil {
			return fmt.Errorf("error getting career stats: %v", err)
		}

		// Display player info
		fmt.Println("Player Information:")
		fmt.Println("==================")
		fmt.Println(FormatPlayerInfo(playerInfo))

		// Display career stats
		fmt.Println(FormatCareerStats(careerStats))

		return nil
	}

	// Regular search flow
	players, err := client.SearchPlayer(playerName)
	if err != nil {
		return fmt.Errorf("error searching for player: %v", err)
	}

	// Check if any players were found
	if len(players) == 0 {
		return fmt.Errorf("no players found matching '%s'", playerName)
	}

	// If only one player is found, use that player
	var selectedPlayer Player
	if len(players) == 1 {
		selectedPlayer = players[0]
		fmt.Printf("Found player: %s (%s)\n\n", selectedPlayer.Name, selectedPlayer.Team)
	} else {
		// If multiple players are found, let the user select one
		fmt.Printf("Found %d players matching '%s':\n", len(players), playerName)
		for i, player := range players {
			fmt.Printf("%d. %s (%s) - %s\n", i+1, player.Name, player.Position, player.Team)
		}

		// Get user selection
		selectedIndex := promptForSelection(len(players))
		if selectedIndex < 0 {
			return fmt.Errorf("invalid selection")
		}

		selectedPlayer = players[selectedIndex]
		fmt.Printf("\nSelected player: %s (%s)\n\n", selectedPlayer.Name, selectedPlayer.Team)
	}

	// Get player info
	playerInfo, err := client.GetPlayerInfo(selectedPlayer.PlayerID)
	if err != nil {
		return fmt.Errorf("error getting player info: %v", err)
	}

	// Get career stats
	careerStats, err := client.GetCareerStats(selectedPlayer.PlayerID)
	if err != nil {
		return fmt.Errorf("error getting career stats: %v", err)
	}

	// Display player info
	fmt.Println("Player Information:")
	fmt.Println("==================")
	fmt.Println(FormatPlayerInfo(playerInfo))

	// Display career stats
	fmt.Println(FormatCareerStats(careerStats))

	return nil
}

// promptForSelection prompts the user to select a player from the list
func promptForSelection(maxIndex int) int {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\nEnter the number of the player you want to select (or 'q' to quit): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		// Trim whitespace and newlines
		input = strings.TrimSpace(input)

		// Check if user wants to quit
		if input == "q" || input == "Q" {
			return -1
		}

		// Parse the selection
		index, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Please enter a valid number")
			continue
		}

		// Validate the selection
		if index < 1 || index > maxIndex {
			fmt.Printf("Please enter a number between 1 and %d\n", maxIndex)
			continue
		}

		// Return the zero-based index
		return index - 1
	}
}
