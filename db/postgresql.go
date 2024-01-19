package db

import (
	"context"
	"fmt"
	"os"

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
