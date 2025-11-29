package aoc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Koenigseder/aoc-leaderboard-scraper/internal/clients/filesystem"
	"github.com/Koenigseder/aoc-leaderboard-scraper/internal/types"
)

var errAocUnauthorized = errors.New("aoc unauthorized")

func GetHistoricScoreFile(writer http.ResponseWriter, request *http.Request) {
	eventYear := request.PathValue("eventYear")
	leaderboardId := request.PathValue("leaderboardId")

	fileContent, err := filesystem.ReadHistoricScoreFile(eventYear, leaderboardId)
	if errors.Is(err, os.ErrNotExist) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		json.NewEncoder(writer).Encode(&types.Problem{
			Status: http.StatusNotFound,
			Title:  "leaderboard not found",
		})

		return
	} else if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(writer).Encode(&types.Problem{
			Title:  "failed fetching historic scores",
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
		})

		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(fileContent)
}

// PersistCurrentLocalScores fetches the current local scores of all members of a leaderboard and persists those
// in a historic file
func PersistCurrentLocalScores(writer http.ResponseWriter, request *http.Request) {
	eventYear := request.PathValue("eventYear")
	leaderboardId := request.PathValue("leaderboardId")

	viewKey := request.URL.Query().Get("viewKey")

	leaderboardInformation, err := getLeaderboard(eventYear, leaderboardId, viewKey)
	if errors.Is(err, errAocUnauthorized) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(writer).Encode(&types.Problem{
			Status: http.StatusUnauthorized,
		})

		return
	} else if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(writer).Encode(&types.Problem{
			Title:  "failed fetching current leaderboard status",
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
		})

		return
	}

	err = filesystem.UpdateHistoricScoreFile(leaderboardInformation)
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(writer).Encode(&types.Problem{
			Title:  "failed persisting current leaderboard status",
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
		})

		return
	}

	writer.WriteHeader(http.StatusCreated)
}

func getLeaderboard(eventYear, leaderboardId, viewKey string) (*types.LeaderboardInformation, error) {
	boardURL := fmt.Sprintf("https://adventofcode.com/%s/leaderboard/private/view/%s.json?view_key=%s", eventYear, leaderboardId, viewKey)

	leaderboardInformation := new(types.LeaderboardInformation)

	res, err := http.Get(boardURL)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusBadRequest {
		return nil, errAocUnauthorized
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("not status 200: %d", res.StatusCode)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resBody, leaderboardInformation)
	if err != nil {
		return nil, err
	}

	return leaderboardInformation, nil
}
