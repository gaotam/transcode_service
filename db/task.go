package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func NewLog(typeLog string, id_source string, metadata string) (id string, err error) {
	updatedAt := time.Now()
	err = Connect.QueryRow(context.Background(), "SELECT id FROM tasks WHERE \"type\" = $1 AND id_source = $2", typeLog, id_source).Scan(&id)
	if err == nil {
		return id, nil
	}

	id = uuid.New().String()
	_, err = Connect.Exec(context.Background(), "INSERT INTO tasks(id, type, id_source, metadata, \"updatedAt\") VALUES($1, $2, $3, $4, $5) RETURNING id;", id, typeLog, id_source, metadata, updatedAt)
	if err != nil {
		return "", err
	}

	return
}

func UpdateLogById(id string, status string, state string) (err error) {
	if state == "" {
		_, err = Connect.Exec(context.Background(), "UPDATE tasks SET status = $1 WHERE id = $2", status, id)
		return
	}
	_, err = Connect.Exec(context.Background(), "UPDATE tasks SET status = $1, state = $2 WHERE id = $3", status, state, id)
	return nil
}
