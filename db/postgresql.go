package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
)

var Connect *pgx.Conn

func ConnectPostgresql() (err error) {
	Connect, err = pgx.Connect(context.Background(), viper.GetString("postgresql.dsn"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return err
	}

	fmt.Println("Connect postgresql success")
	return nil
}

func NewLog(typeLog string, metadata string) (err error) {
	id := uuid.New()
	updatedAt := time.Now()
	_, err = Connect.Exec(context.Background(), "INSERT INTO tasks(id, type, metadata, \"updatedAt\") VALUES($1, $2, $3, $4)", id, typeLog, metadata, updatedAt)
	if err != nil {
		return err
	}
	return nil
}
