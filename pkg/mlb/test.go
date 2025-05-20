package mlb

import (
	"fmt"
	"io"
	"net/http"
)

// TestAPIConnection tests if the MLB Stats API is accessible
func TestAPIConnection() error {
	// Try different MLB Stats API endpoints
	urls := []string{
		"https://statsapi.mlb.com/api/v1/people?search=trout",
		"https://statsapi.mlb.com/api/v1/people/545361", // Mike Trout's ID
		"https://statsapi.mlb.com/api/v1/people/545361/stats?stats=career&group=hitting",
		"https://statsapi.mlb.com/api/v1/teams",
	}

	client := &http.Client{}

	for _, url := range urls {
		fmt.Printf("Testing URL: %s\n", url)

		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			continue
		}

		fmt.Printf("Status: %s\n", resp.Status)
		if len(body) > 100 {
			fmt.Printf("Response preview: %s\n\n", body[:100])
		} else {
			fmt.Printf("Response preview: %s\n\n", body)
		}
	}

	return nil
}
