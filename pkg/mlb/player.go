package mlb

// Player represents a player from the search results
type Player struct {
	PlayerID     string `json:"player_id"`
	Name         string `json:"name_display_first_last"`
	Position     string `json:"position"`
	Team         string `json:"team_full"`
	TeamAbbrev   string `json:"team_abbrev"`
	BatSide      string `json:"bats"`
	ThrowSide    string `json:"throws"`
	Weight       string `json:"weight"`
	Height       string `json:"height_feet,height_inches"`
	BirthDate    string `json:"birth_date"`
	BirthCountry string `json:"birth_country"`
	Active       string `json:"active_sw"`
	PrimaryPos   string `json:"position_txt"`
	JerseyNumber string `json:"jersey_number"`
	Status       string `json:"status"`
	College      string `json:"college"`
}

// PlayerInfo represents detailed player information
type PlayerInfo struct {
	PlayerID     string `json:"player_id"`
	Name         string `json:"name_display_first_last_html"`
	Position     string `json:"primary_position_txt"`
	Team         string `json:"team_name"`
	TeamAbbrev   string `json:"team_abbrev"`
	BatSide      string `json:"bats"`
	ThrowSide    string `json:"throws"`
	Weight       string `json:"weight"`
	Height       string `json:"height_feet,height_inches"`
	BirthDate    string `json:"birth_date"`
	BirthCity    string `json:"birth_city"`
	BirthState   string `json:"birth_state"`
	BirthCountry string `json:"birth_country"`
	Age          string `json:"age"`
	Status       string `json:"status"`
	Active       string `json:"active_sw"`
	JerseyNumber string `json:"jersey_number"`
	ProDebutDate string `json:"pro_debut_date"`
	TwitterID    string `json:"twitter_id"`
	Nickname     string `json:"name_nick"`
	College      string `json:"college"`
	HighSchool   string `json:"high_school"`
}

// SearchResponse represents the response from the player search endpoint
type SearchResponse struct {
	SearchPlayerAll struct {
		CopyRight    string `json:"copyRight"`
		QueryResults struct {
			Created   string   `json:"created"`
			TotalSize string   `json:"totalSize"`
			Row       []Player `json:"row"`
		} `json:"queryResults"`
	} `json:"search_player_all"`
}

// PlayerInfoResponse represents the response from the player info endpoint
type PlayerInfoResponse struct {
	PlayerInfo struct {
		CopyRight    string `json:"copyRight"`
		QueryResults struct {
			Created   string     `json:"created"`
			TotalSize string     `json:"totalSize"`
			Row       PlayerInfo `json:"row"`
		} `json:"queryResults"`
	} `json:"player_info"`
}
