package db

import (
	"context"
	"database/sql"
	"fmt"
	"gitlab.smartbet.am/golang/smart-image/ent"
	"gitlab.smartbet.am/golang/smart-image/ent/migrate"
	"gitlab.smartbet.am/golang/smart-image/internal/config"
	"strconv"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	_ "github.com/go-sql-driver/mysql"
)

var client *ent.Client

type configData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Db       string `json:"db"`
	Port     int    `json:"port"`
}

// Open new connection

func Open() (*ent.Client, error) {
	cgf := config.Get()
	portStr := cgf.DBPort
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	cc := configData{
		Username: cgf.DBUser,
		Password: cgf.DBPassword,
		Host:     cgf.DBHost,
		Db:       cgf.DBName,
		Port:     port,
	}

	databaseUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		cc.Username,
		cc.Password,
		cc.Host,
		cc.Port,
		cc.Db,
	)

	db, err := sql.Open("mysql", databaseUrl)
	if err != nil {
		return nil, err
	}

	drv := entsql.OpenDB(dialect.MySQL, db)
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
