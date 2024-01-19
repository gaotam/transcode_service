package db

import (
	"context"
)

type VideoByLiveKeyResult struct {
	Id     string
	Status string
}

func GetLiveByLiveKey(liveKey string) (result VideoByLiveKeyResult, err error) {
	var id, status string
	err = Connect.QueryRow(context.Background(), "SELECT id, status FROM livestreams WHERE \"liveKey\" = $1", liveKey).Scan(&id, &status)
	if err != nil {
		return VideoByLiveKeyResult{}, err
	}
	return VideoByLiveKeyResult{Id: id, Status: status}, nil
}
