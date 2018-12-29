package tracker

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"go.uber.org/zap"
)

// Enviroment holds env
type Enviroment int

const (
	// Dev env
	Dev Enviroment = 0
	// Prod env
	Prod Enviroment = 1
)

var (
	db     *sql.DB
	logger *zap.Logger
	env    Enviroment
)

// Init x
func Init(isProd bool) (*sql.DB, error) {
	var err error

	// open mysql conn
	db, err = sql.Open("mysql", "root@/bittorrent")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Set env
	if isProd == true {
		env = Prod
	} else {
		env = Dev
	}

	// init logger
	if env == Prod {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	defer logger.Sync()

	// init ban table
	initBanTable()

	return db, nil
}

// Clean cleans the db
func Clean() {
	// Every 5 min
	for c := time.Tick(5 * time.Minute); ; <-c {
		logger.Info("Cleaning tables...")

		// Get all tables
		rows, err := db.Query("SELECT DISTINCT TABLE_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='bittorrent' AND TABLE_NAME LIKE 'Hash_%'")
		if err != nil {
			logger.Error(err.Error())
		}
		tables := []string{}
		for rows.Next() {
			var table string
			err = rows.Scan(&table)
			if err != nil {
				logger.Error(err.Error())
			}
			tables = append(tables, table)
		}
		rows.Close()

		// Delete entries that havn't checked in in 10 min
		for _, table := range tables {
			result, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE lastSeen < ?", table), time.Now().Unix()-int64(60*10))
			if err != nil {
				logger.Error(err.Error())
			}

			affected, err := result.RowsAffected()
			if err != nil {
				logger.Warn(err.Error())
			}

			logger.Info("Cleaned table",
				zap.String("table", table),
				zap.Int64("affected", affected),
			)
		}

		// Auto delete empty tables
		result, err := db.Exec("SELECT CONCAT('DROP TABLE ', GROUP_CONCAT(table_name), ';') AS query FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_ROWS = '0' AND TABLE_SCHEMA = 'bittorrent'")
		if err != nil {
			logger.Error(err.Error())
		}

		affected, err := result.RowsAffected()
		if err != nil {
			logger.Warn(err.Error())
		}

		logger.Info("Deleted empty tables",
			zap.Int64("count", affected),
		)
	}
}
