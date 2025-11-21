package mysql

import (
	"FairLAP/internal/config"
	"context"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"time"
)

func Connect(cfg *config.MySQLConfig) (*sqlx.DB, error) {
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dataSource := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&multiStatements=true", cfg.User, cfg.Password, cfg.Host, cfg.Schema)

	db, err := sqlx.ConnectContext(ctx, "mysql", dataSource)
	if err != nil {
		return nil, err
	}

	return db, nil
}
