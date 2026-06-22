package db

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"gitlab.smartbet.am/golang/smart-image/ent"
	"gitlab.smartbet.am/golang/smart-image/ent/migrate"
	"gitlab.smartbet.am/golang/smart-image/internal/service/config"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	_ "github.com/go-sql-driver/mysql"
)

type configData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Db       string `json:"db"`
	Port     int    `json:"port"`
}

// Open new connection

type DB struct {
	*ent.Client
	config *config.Config
	raw    *sql.DB
}

func NewDB(c *config.Config) *DB {
	return &DB{
		config: c,
	}
}
func (db *DB) TX(handler func(tx *ent.Tx) error) error {
	tx, err := db.Tx(context.Background())
	if err != nil {
		return err
	}

	err = handler(tx)

	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()

}

func (db *DB) Probe(ctx context.Context) string {
	var host, name string
	var readOnly, connID int
	if err := db.raw.QueryRowContext(ctx,
		"SELECT @@hostname, DATABASE(), @@read_only, CONNECTION_ID()").
		Scan(&host, &name, &readOnly, &connID); err != nil {
		return "probe_err=" + err.Error()
	}
	return fmt.Sprintf("host=%s db=%s read_only=%d conn_id=%d", host, name, readOnly, connID)
}

func (db *DB) connect(ctx context.Context) error {
	portStr := db.config.DBPort
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	cc := configData{
		Username: db.config.DBUser,
		Password: db.config.DBPassword,
		Host:     db.config.DBHost,
		Db:       db.config.DBName,
		Port:     port,
	}

	databaseUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		cc.Username,
		cc.Password,
		cc.Host,
		cc.Port,
		cc.Db,
	)

	con, err := sql.Open("mysql", databaseUrl)
	if err != nil {
		return err
	}
	db.raw = con

	drv := entsql.OpenDB(dialect.MySQL, con)
	db.Client = ent.NewClient(ent.Driver(drv)).Debug()
	err = db.Client.Schema.Create(context.Background(), migrate.WithDropColumn(true))
	if err != nil {
		return err
	}
	return nil
}

func Provider(conf *config.Config) (*DB, error) {
	d := NewDB(conf)
	if err := d.connect(context.Background()); err != nil {
		return nil, err
	}
	return d, nil
}
