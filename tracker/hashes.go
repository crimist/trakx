package tracker

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
)

func initBanTable() {
	_, err := db.Exec("SELECT 1 FROM banned LIMIT 1")
	if err != nil { // If error the table doesn't exist
		_, err = db.Exec(fmt.Sprintf("CREATE TABLE banned (hash varchar(45) UNIQUE)"))
	}

	// read thru ban list
	file, err := os.Open("banlist")
	if err != nil {
		logger.Error(err.Error())
	}
	defer file.Close()

	// loop thru line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Ignore comments
		if strings.HasPrefix(line, "//") || line == "" {
			continue
		}

		if len(line) != 20 {
			logger.Warn("Invalid hash len in banlist")
			continue
		}

		hashEncodded := EncodeInfoHash(line)

		// Add it to table if it's not already there
		// IGNORE stops the duplicate entry from throwing an error
		_, err := db.Exec("INSERT IGNORE INTO banned VALUES (?)", hashEncodded)
		if err != nil {
			logger.Error(err.Error())
		} else {
			logger.Info("Banned",
				zap.String("hash", hashEncodded),
			)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Error(err.Error())
	}
}

// IsBannedHash chekcs if the hash is banned
// 0 = not banned
// 1 = banned
// -1 = failure
func IsBannedHash(hash string) BanErr {
	// sql lookup in hash db
	rows, err := db.Query("SELECT * FROM banned WHERE hash = ?", hash)
	if err != nil {
		logger.Error(err.Error())
		return Err
	}
	defer rows.Close()

	if rows.Next() {
		return ErrBanned
	}

	return ErrOK
}
