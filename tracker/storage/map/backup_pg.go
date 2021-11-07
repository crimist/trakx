package gomap

import (
	"database/sql"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/storage"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	// Maximum retention for entries. Rows older than this will be removed
	// empty to disable
	maxDate = "7 days"

	// Maximum number of rows. Rows exceeding this will be removed by timestamp
	// -1 for unlimited
	maxRows = 10

	// If backup is older than this it will wait for a new backup
	backupRecentWindow = 10 * time.Minute

	// Time to wait for if backup is older than backupRecentWindow
	backupRecentWait = 7 * time.Second
)

type PgBackup struct {
	pg *sql.DB
	db *Memory
}

func (bck *PgBackup) Init(db storage.Database) error {
	var err error

	bck.db = db.(*Memory)
	if bck.db == nil {
		return errors.New("nil database on backup Init()")
	}

	// resolve database address from env variable if needed
	dbAddress := config.Conf.Database.Address
	if strings.HasPrefix(dbAddress, "ENV:") {
		dbAddress = os.Getenv(strings.TrimPrefix(dbAddress, "ENV:"))
	}

	bck.pg, err = sql.Open("postgres", dbAddress)
	if err != nil {
		return errors.Wrap(err, "failed to open pg connection")
	}

	err = bck.pg.Ping()
	if err != nil {
		return errors.Wrap(err, "failed to ping postgres")
	}

	_, err = bck.pg.Exec("CREATE TABLE IF NOT EXISTS trakx (ts TIMESTAMP DEFAULT now(), bytes BYTEA)")
	if err != nil {
		return errors.Wrap(err, "failed to `CREATE TABLE`")
	}

	return nil
}

func (bck PgBackup) Save() error {
	config.Logger.Info("Saving database to pg")
	var data []byte
	var err error
	start := time.Now()

	if fast {
		data, err = bck.db.encodeBinaryUnsafe()
	} else {
		data, err = bck.db.encodeBinaryUnsafeAutoalloc()
	}

	if err != nil {
		return errors.Wrap(err, "failed to encode binary")
	}
	config.Logger.Info("Encoded binary", zap.Duration("duration", time.Since(start)))
	start = time.Now()

	_, err = bck.pg.Query("INSERT INTO trakx(bytes) VALUES($1)", data)
	if err != nil {
		return errors.Wrap(err, "`INSERT` statement failed")
	}

	config.Logger.Info("Saved database to pg", zap.Any("hash", data[:20]), zap.Duration("pg duration", time.Since(start)))

	return nil
}

func (bck PgBackup) Load() error {
	firstTry := true
	var bytes []byte
	var ts time.Time

	config.Logger.Info("Loading stored database from postgres")

attemptLoad:

	err := bck.pg.QueryRow("SELECT bytes, ts FROM trakx ORDER BY ts DESC LIMIT 1").Scan(&bytes, &ts)
	if err != nil {
		// if postgres is empty than create an empty database and return success
		if strings.Contains(err.Error(), "no rows in result set") {
			bck.db.make()
			config.Logger.Info("No rows found in postgres, created empty database")
			return nil
		}

		// otherwise return failure
		return errors.Wrap(err, "`SELECT` statement failed")
	}

	// If backup is older than 20 min wait a sec for a backup to arrive
	if time.Since(ts) > backupRecentWindow && firstTry {
		firstTry = false

		config.Logger.Info("Failed to detect a pg backup within window, waiting...", zap.Duration("window", backupRecentWindow), zap.Duration("wait", backupRecentWait))
		time.Sleep(backupRecentWait)
		config.Logger.Info("Reattempting load...")

		goto attemptLoad
	}

	peers, hashes, err := bck.db.decodeBinaryUnsafe(bytes)
	if err != nil {
		return errors.Wrap(err, "failed to decode saved data")
	}

	config.Logger.Info("Loaded stored database from pg", zap.Int("size", len(bytes)), zap.Any("hash", bytes[:20]), zap.Int("peers", peers), zap.Int("hashes", hashes))

	return nil
}

func (bck PgBackup) trim() (int64, error) {
	var trimmed int64

	if len(maxDate) != 0 {
		result, err := bck.pg.Exec("DELETE FROM trakx WHERE ts < NOW() - INTERVAL '" + maxDate + "'")
		if err != nil {
			return -1, err
		}

		trimmed, err = result.RowsAffected()
		if err != nil {
			return -1, err
		}
	}

	if maxRows != -1 {
		result, err := bck.pg.Exec("DELETE FROM trakx WHERE ctid IN (SELECT ctid FROM trakx ORDER BY ctid DESC OFFSET " + strconv.Itoa(maxRows) + ")")
		if err != nil {
			return -1, err
		}
		removedRows, err := result.RowsAffected()
		if err != nil {
			return -1, err
		}

		trimmed += removedRows
	}

	return trimmed, nil
}
