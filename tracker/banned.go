package tracker

import (
	"bufio"
	"os"
	"strings"

	"go.uber.org/zap"
)

// Ban holds banned hashes
type Ban struct {
	Hash   string `gorm:"unique"`
	Reason string
}

func initBans() error {
	// Wipe table
	db.DropTableIfExists(&Ban{})
	db.CreateTable(&Ban{})

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
		if strings.HasPrefix(line, "//") || line == "" {
			continue
		}

		split := strings.SplitN(line, " ", 2)
		reason := ""
		if len(split) > 1 {
			reason = split[1] // If a reason was provided use it
		}

		entry := Ban{
			Hash:   split[0],
			Reason: reason,
		}

		// Check if it's a valid hash
		if len(entry.Hash) != 20 {
			logger.Error("Invalid hash in banlist",
				zap.String("hash", entry.Hash),
				zap.Int("len", len(entry.Hash)),
			)
			continue
		}

		db.Create(&entry)
		logger.Info("Banned",
			zap.String("hash", entry.Hash),
		)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// IsBanned checks if the given hash is banned
func IsBanned(hash string) TrackErr {
	ban := Ban{}
	db.Where("hash = ?", hash).First(&ban)
	if ban.Hash != "" {
		return Banned
	}
	return OK
}
