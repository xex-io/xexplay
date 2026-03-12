package oddsapi

import "time"

// SportResponse represents a sport from The Odds API.
type SportResponse struct {
	Key          string `json:"key"`
	Group        string `json:"group"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Active       bool   `json:"active"`
	HasOutrights bool   `json:"has_outrights"`
}

// OddsMatch represents an upcoming match with odds.
type OddsMatch struct {
	ID           string      `json:"id"`
	SportKey     string      `json:"sport_key"`
	SportTitle   string      `json:"sport_title"`
	CommenceTime time.Time   `json:"commence_time"`
	HomeTeam     string      `json:"home_team"`
	AwayTeam     string      `json:"away_team"`
	Bookmakers   []Bookmaker `json:"bookmakers"`
}

// Bookmaker represents a bookmaker's odds.
type Bookmaker struct {
	Key        string   `json:"key"`
	Title      string   `json:"title"`
	LastUpdate string   `json:"last_update"`
	Markets    []Market `json:"markets"`
}

// Market represents an odds market (h2h, spreads, totals).
type Market struct {
	Key        string    `json:"key"`
	LastUpdate string    `json:"last_update"`
	Outcomes   []Outcome `json:"outcomes"`
}

// Outcome represents a single outcome in an odds market.
type Outcome struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Point float64 `json:"point,omitempty"`
}

// ScoreResult represents a match score from The Odds API.
type ScoreResult struct {
	ID           string      `json:"id"`
	SportKey     string      `json:"sport_key"`
	SportTitle   string      `json:"sport_title"`
	CommenceTime time.Time   `json:"commence_time"`
	Completed    bool        `json:"completed"`
	HomeTeam     string      `json:"home_team"`
	AwayTeam     string      `json:"away_team"`
	Scores       []TeamScore `json:"scores"`
	LastUpdate   *time.Time  `json:"last_update"`
}

// TeamScore represents the score for a team.
type TeamScore struct {
	Name  string `json:"name"`
	Score string `json:"score"`
}

// H2HOdds extracts head-to-head odds from the first bookmaker.
func (m *OddsMatch) H2HOdds() (homeOdds, awayOdds, drawOdds float64) {
	for _, bm := range m.Bookmakers {
		for _, market := range bm.Markets {
			if market.Key == "h2h" {
				for _, o := range market.Outcomes {
					switch o.Name {
					case m.HomeTeam:
						homeOdds = o.Price
					case m.AwayTeam:
						awayOdds = o.Price
					case "Draw":
						drawOdds = o.Price
					}
				}
				return
			}
		}
	}
	return
}
