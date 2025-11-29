package types

type LeaderboardInformation struct {
	Members   map[string]User `json:"members"`
	OwnerID   int             `json:"owner_id"`
	EventYear string          `json:"event"`
	NumDays   int             `json:"num_days"`
}

type User struct {
	ID         int `json:"id"`
	LocalScore int `json:"local_score"`
}
