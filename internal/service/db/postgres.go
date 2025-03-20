package db

import (
	"context"
	"database/sql"
	"gitlab.smartbet.am/golang/smart-image/ent"
	"gitlab.smartbet.am/golang/smart-image/ent/migrate"

	"fmt"
	"os"
	"strconv"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var client *ent.Client

type configData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Db       string `json:"db"`
	Port     int    `json:"port"`
	SSL      string `json:"ssl"`
}

// Open new connection

func Open() (*ent.Client, error) {
	portStr := os.Getenv("POSTGRES_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	cc := configData{
		Username: os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Host:     os.Getenv("POSTGRES_HOST"),
		Db:       os.Getenv("POSTGRES_DB"),
		Port:     port,
		SSL:      os.Getenv("POSTGRES_SSL"),
	}

	databaseUrl := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
		cc.Username,
		cc.Password,
		cc.Host,
		cc.Port,
		cc.Db,
		cc.SSL,
	)

	db, err := sql.Open("pgx", databaseUrl)
	if err != nil {
		return nil, err
	}

	// Create an ent.Driver from `db`.
	drv := entsql.OpenDB(dialect.Postgres, db)
	client = ent.NewClient(ent.Driver(drv)).Debug()
	err = client.Schema.Create(context.Background(), migrate.WithDropColumn(true))
	if err != nil {
		return nil, err
	}
	return client, nil
}
func Client() *ent.Client {
	return client
}
