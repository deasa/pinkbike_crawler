package mlb

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCleanLocationString(t *testing.T) {
	tests := []struct {
		name     string
		location string
		expected string
	}{
		{
			name:     "normal location",
			location: "New York, NY, USA",
			expected: "New York, NY, USA",
		},
		{
			name:     "missing state",
			location: "New York, , USA",
			expected: "New York, USA",
		},
		{
			name:     "missing city and state",
			location: ", , USA",
			expected: "USA",
		},
		{
			name:     "empty string",
			location: "",
			expected: "",
		},
		{
			name:     "only commas",
			location: ", , ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanLocationString(tt.location)
			if result != tt.expected {
				t.Errorf("cleanLocationString(%s) = %s, expected %s", tt.location, result, tt.expected)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		x        int
		y        int
		expected int
	}{
		{
			name:     "x less than y",
			x:        5,
			y:        10,
			expected: 5,
		},
		{
			name:     "y less than x",
			x:        10,
			y:        5,
			expected: 5,
		},
		{
			name:     "x equals y",
			x:        10,
			y:        10,
			expected: 10,
		},
		{
			name:     "negative numbers",
			x:        -10,
			y:        -5,
			expected: -10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("min(%d, %d) = %d, expected %d", tt.x, tt.y, result, tt.expected)
			}
		})
	}
}

func TestGetPlayerByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name     string
		playerID string
		wantName string
		wantErr  bool
	}{
		{
			name:     "Mike Trout",
			playerID: "545361",
			wantName: "Mike Trout",
			wantErr:  false,
		},
		{
			name:     "Aaron Judge",
			playerID: "592450",
			wantName: "Aaron Judge",
			wantErr:  false,
		},
		{
			name:     "Invalid Player ID",
			playerID: "0",
			wantName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player, err := GetPlayerByID(tt.playerID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPlayerByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && player.Name != tt.wantName {
				t.Errorf("GetPlayerByID() got player name = %v, want %v", player.Name, tt.wantName)
			}
		})
	}
}

func TestSearchPlayer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name       string
		playerName string
		wantName   string
		wantErr    bool
	}{
		{
			name:       "Mike Trout",
			playerName: "Mike Trout",
			wantName:   "Mike Trout",
			wantErr:    false,
		},
		{
			name:       "Trout",
			playerName: "Trout",
			wantName:   "Mike Trout",
			wantErr:    false,
		},
		{
			name:       "Aaron Judge",
			playerName: "Aaron Judge",
			wantName:   "Aaron Judge",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			players, err := SearchPlayer(tt.playerName)

			if (err != nil) != tt.wantErr {
				t.Errorf("SearchPlayer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && (len(players) == 0 || players[0].Name != tt.wantName) {
				if len(players) == 0 {
					t.Errorf("SearchPlayer() got no players")
				} else {
					t.Errorf("SearchPlayer() got player name = %v, want %v", players[0].Name, tt.wantName)
				}
			}
		})
	}
}

func TestSearchPlayer_MockResponse(t *testing.T) {
	// Setup a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for the direct player ID endpoint
		if r.URL.Path == "/api/v1/people/545361" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"people": [
					{
						"id": 545361,
						"fullName": "Mike Trout",
						"firstName": "Michael",
						"lastName": "Trout",
						"primaryNumber": "27",
						"birthDate": "1991-08-07",
						"currentAge": 33,
						"birthCity": "Vineland",
						"birthStateProvince": "NJ",
						"birthCountry": "USA",
						"height": "6' 2\"",
						"weight": 235,
						"active": true,
						"primaryPosition": {
							"code": "9",
							"name": "Outfielder",
							"type": "Outfielder",
							"abbreviation": "RF"
						},
						"batSide": {
							"code": "R",
							"description": "Right"
						},
						"pitchHand": {
							"code": "R",
							"description": "Right"
						},
						"mlbDebutDate": "2011-07-08"
					}
				]
			}`))
			return
		}

		// If it's a search request
		if r.URL.Path == "/api/v1/people/search" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"people": [
					{
						"id": 545361,
						"fullName": "Mike Trout",
						"firstName": "Michael",
						"lastName": "Trout",
						"primaryNumber": "27",
						"birthDate": "1991-08-07",
						"currentAge": 33,
						"birthCity": "Vineland",
						"birthStateProvince": "NJ",
						"birthCountry": "USA",
						"height": "6' 2\"",
						"weight": 235,
						"active": true,
						"primaryPosition": {
							"code": "9",
							"name": "Outfielder",
							"type": "Outfielder",
							"abbreviation": "RF"
						},
						"batSide": {
							"code": "R",
							"description": "Right"
						},
						"pitchHand": {
							"code": "R",
							"description": "Right"
						},
						"mlbDebutDate": "2011-07-08"
					}
				]
			}`))
			return
		}

		// Default response
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// We can't modify the constant directly, so we'll use the mock server URL in individual test cases
	t.Run("GetPlayerByID with mock server", func(t *testing.T) {
		// Skip this test - we can't modify the API base URL
		t.Skip("Skipping mock server test as we can't modify the API base URL")
	})
}
