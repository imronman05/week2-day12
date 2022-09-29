package connection

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

var Conn *pgx.Conn

func ConnectionProject() {

	// postgres://postgres:password@localhost:5432/database_name
	urlDataBase := "postgres://postgres:postgres@localhost:5432/day9"

	var err error
	Conn, err = pgx.Connect(context.Background(), urlDataBase)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Success connect to database")
}
