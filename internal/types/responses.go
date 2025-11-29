package types //nolint:revive

type HistoricLocalScores struct {
	HistoricScores map[string]map[string]int `json:"historic_scores"`
	OwnerId        int                       `json:"owner_id"`
	EventYear      string                    `json:"event_year"`
	NumDays        int                       `json:"num_days"`
}

type Problem struct {
	Title    string `json:"title,omitempty"`
	Status   uint32 `json:"status"`
	Type     string `json:"type,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}
