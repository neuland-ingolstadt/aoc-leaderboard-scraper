package filesystem

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/Koenigseder/aoc-leaderboard-scraper/internal/types"
)

const (
	envVarDataDumpDirectory = "DATA_DUMP_DIRECTORY"
)

// ReadHistoricScoreFile reads the dump file and returns it as a struct
func ReadHistoricScoreFile(eventYear, ownerId string) (*types.HistoricLocalScores, error) {
	dataDumpDirectory, ok := os.LookupEnv(envVarDataDumpDirectory)
	if !ok {
		slog.Error(fmt.Sprintf("failed looking up env var '%s'", envVarDataDumpDirectory))
		os.Exit(1)
	}

	filePath := fmt.Sprintf("%s/%s/%s.json", dataDumpDirectory, eventYear, ownerId)

	if _, err := os.Stat(filePath); err != nil {
		return nil, err
	}

	// File exists - Read it
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	fileContent := new(types.HistoricLocalScores)

	err = json.Unmarshal(fileBytes, fileContent)
	if err != nil {
		return nil, err
	}

	return fileContent, nil
}

// UpdateHistoricScoreFile updates the historic file used to store the leaderboards member's local scores
func UpdateHistoricScoreFile(information *types.LeaderboardInformation) error {
	dataDumpDirectory, ok := os.LookupEnv(envVarDataDumpDirectory)
	if !ok {
		slog.Error(fmt.Sprintf("failed looking up env var '%s'", envVarDataDumpDirectory))
		os.Exit(1)
	}

	dirPath := fmt.Sprintf("%s/%s", dataDumpDirectory, information.EventYear)
	filePath := dirPath + "/" + strconv.Itoa(information.OwnerID) + ".json"

	currentDate := time.Now().Format("2006-01-02")
	fileContent := new(types.HistoricLocalScores)

	if _, err := os.Stat(filePath); err == nil {
		// File exists - Read it
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		err = json.Unmarshal(fileBytes, fileContent)
		if err != nil {
			return err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// File does not exist - Create it
		err := os.MkdirAll(dirPath, 0777)
		if err != nil {
			return err
		}

		err = os.WriteFile(filePath, nil, 0777)
		if err != nil {
			return err
		}

		fileContent = &types.HistoricLocalScores{
			HistoricScores: make(map[string]map[string]int),
			OwnerId:        information.OwnerID,
			EventYear:      information.EventYear,
			NumDays:        information.NumDays,
		}
	} else if err != nil {
		return err
	}

	for userId, userInfo := range information.Members {
		if _, ok := fileContent.HistoricScores[currentDate]; !ok {
			fileContent.HistoricScores[currentDate] = make(map[string]int)
		}

		fileContent.HistoricScores[currentDate][userId] = userInfo.LocalScore
	}

	// Write file
	fileBytes, err := json.Marshal(fileContent)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, fileBytes, 0600)
	if err != nil {
		return err
	}

	return nil
}
