package tracker

import (
	"bufio"
	"net/url"
	"os"
	"strings"

	"go.uber.org/zap"
)

// Ban holds banned hashes
type Ban struct {
	Hash   Hash   `db:"hash"`
	Reason string `db:"reason"`
}

func initBans() error {
	bantable := `
	CREATE TABLE IF NOT EXISTS bans (
		hash varbinary(255) UNIQUE,
		reason varchar(255)
	)
	`

	if _, err := db.Exec(bantable); err != nil {
		panic(err)
	}
	result, err := db.Exec("DELETE FROM bans")
	if err != nil {
		panic(err)
	} else {
		count, err := result.RowsAffected()
		if err != nil {
			panic(err)
		}
		logger.Info("Cleared bans", zap.Int64("Count", count))
	}

	file, err := os.Open("banlist")
	if err != nil {
		return err
	}
	defer file.Close()

	// loop thru line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Ignore comments
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") || len(line) < 20 {
			continue
		}

		split := strings.SplitN(line, " ", 2)
		reason := ""
		if len(split) > 1 {
			reason = split[1] // If a reason was provided use it
		}

		hash, err := url.QueryUnescape(split[0])
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		entry := Ban{
			Hash:   []byte(hash),
			Reason: reason,
		}

		db.Exec("INSERT INTO bans (hash, reason) VALUES (?, ?)", entry.Hash, entry.Reason)
		logger.Info("Banned",
			zap.ByteString("hash", entry.Hash[:]),
		)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
