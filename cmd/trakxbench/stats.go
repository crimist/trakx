package main

import (
	"encoding/json"
	"net/http"
	"time"
)

func fetchStats(url string, timeout time.Duration) (map[string]interface{}, error) {
	client := http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}
