package mlb

// CareerStats combines both hitting and pitching career stats
type CareerStats struct {
	Hitting  *HittingStats
	Pitching *PitchingStats
}

// HittingStats represents career hitting statistics
type HittingStats struct {
	AVG     string
	HR      string
	RBI     string
	Runs    string
	SB      string
	OBP     string
	SLG     string
	OPS     string
	Hits    string
	Doubles string
	Triples string
	BB      string
	SO      string
	Games   string
	AB      string
}

// PitchingStats represents career pitching statistics
type PitchingStats struct {
	ERA    string
	Wins   string
	Losses string
	WHIP   string
	SO     string
	BB     string
	SV     string
	IP     string
	Games  string
	GS     string
	CG     string
	SHO    string
}

// CareerHittingResponse represents the response from the career hitting stats endpoint
type CareerHittingResponse struct {
	SportCareerHitting struct {
		CopyRight    string `json:"copyRight"`
		QueryResults struct {
			Created   string       `json:"created"`
			TotalSize string       `json:"totalSize"`
			Row       HittingStats `json:"row"`
		} `json:"queryResults"`
	} `json:"sport_career_hitting"`
}

// CareerPitchingResponse represents the response from the career pitching stats endpoint
type CareerPitchingResponse struct {
	SportCareerPitching struct {
		CopyRight    string `json:"copyRight"`
		QueryResults struct {
			Created   string        `json:"created"`
			TotalSize string        `json:"totalSize"`
			Row       PitchingStats `json:"row"`
		} `json:"queryResults"`
	} `json:"sport_career_pitching"`
}
