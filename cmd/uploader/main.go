package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jmoiron/sqlx"
	appcloser "github.com/s4bb4t/bcmon/pkg/app_closer"
	"github.com/s4bb4t/bcmon/pkg/pgsql/pgconnector"
	"os"
)

func main() {
	path := flag.String("f", "../../contracts.json", "define filename to save contracts into it")
	flag.Parse()

	cfg := CreateConfig()

	closer := appcloser.InitCloser(nil)
	fmt.Println("cfg and closer initialized")

	pgConnector, err := pgconnector.CreateConnection(context.Background(),
		cfg.Db.Postgres.GetDsn(),
		cfg.Db.Postgres.GetMaxOpenConns(),
		cfg.Db.Postgres.GetIdleConns(),
		cfg.Db.Postgres.GetIdleTime(),
		closer)
	if err != nil {
		panic(err)
	}

	db, err := pgConnector.GetConnection(context.Background(), pgconnector.DBReadWrite)
	if err != nil {
		panic(err)
	}
	fmt.Println("GetConnection successfully")

	err = Upload(db, path)
	if err != nil {
		panic(err)
	}
	fmt.Println("saved at", *path)
}

type Data struct {
	InputData []string `json:"input_data"`
}

func Upload(db *sqlx.DB, path *string) error {
	rows, err := db.QueryContext(context.Background(), `select * from public.contract`)
	if err != nil {
		return err
	}

	var arr Data
	var addr string

	for rows.Next() {
		err := rows.Scan(&addr)
		if err != nil {
			return err
		}

		arr.InputData = append(arr.InputData, addr)
	}

	bytes, err := json.Marshal(arr)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(*path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(bytes); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}
