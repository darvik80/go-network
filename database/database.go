package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

func NewDb(cfg DataSourceConfig) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.WithFields(log.Fields{"module": "db"}).Errorf("create faled: %s:%s, %e", cfg.Host, cfg.Database, err)
		return nil, err
	}

	log.WithFields(log.Fields{"module": "db"}).Infof("connected db: %s:%s", cfg.Host, cfg.Database)
	return db, db.Ping()
}
