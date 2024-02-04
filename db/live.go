package db

import (
	"context"
)

type VideoByLiveKeyResult struct {
	Id     string
	Status string
	App    string
}

func GetLiveByLiveKey(liveKey string) (result VideoByLiveKeyResult, err error) {
	var id, status string
	var isRecord bool
	var app string
	err = Connect.QueryRow(context.Background(), "SELECT id, status, \"isRecord\" FROM livestreams WHERE \"liveKey\" = $1", liveKey).Scan(&id, &status, &isRecord)
	if err != nil {
		return VideoByLiveKeyResult{}, err
	}

	if !isRecord {
		app = "nr_live"
	} else {
		app = "live"
	}

	return VideoByLiveKeyResult{Id: id, Status: status, App: app}, nil
}
