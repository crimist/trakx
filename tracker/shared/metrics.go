package shared

import (
	"fmt"
)

var StatsHTML string

func processMetrics() {
	stats := make(map[string]int, 300)

	for _, peermap := range PeerDB {
		for peerid := range peermap {
			if peerid[0] == '-' && peerid[7] == '-' {
				stats[getAzureus(string(peerid[1:6]))]++
			} else {
				stats[getShadow(string(peerid[0:6]))]++
			}
			/*
				else if string(peerid[6:9]) == "---" {
					Shadows[string(peerid[0:6])]++
				} else if peerid[6] == peerid[7] && peerid[7] == peerid[8] {
					Shadows[string(peerid[0:6])]++
				}
			*/
		}
	}

	StatsHTML = ""
	StatsHTML += "<table>"
	for client, count := range stats {
		StatsHTML += fmt.Sprintf(`<tr><td>%s</td><td>%d</td></tr>`, client, count)
	}
	StatsHTML += "</table>"
}

// get the full name of the client w/ azureus method
func getAzureus(azureus string) string {
	var client string

	switch string(azureus[0:2]) {
	case "AZ":
		client += "Azureus"
	case "DE":
		client += "DelugeTorrent"
	case "qB":
		client += "qBittorrent"
	case "SD":
		client += "Thunder"
	case "UM":
		client += "uTorrent Mac"
	case "UT":
		client += "uTorrent"
	default:
		client += azureus[0:2]
	}

	client += " v"

	for _, c := range azureus[2:6] {
		if string(c) == "-" {
			break
		}
		if int(c) > 47 && int(c) < 58 {
			client += string(c) + "."
		} else if int(c) > 64 && int(c) < 72 {
			client += fmt.Sprintf("%d.", int(c)-55)
		} else {
			panic("invalid version char")
		}
	}

	return client[0 : len(client)-1]
}

// get the full name of the client w/ shadow method
func getShadow(shadow string) string {
	var client string

	switch string(shadow[0]) {
	case "A":
		client += "ABC"
	case "O":
		client += "Osprey Permaseed"
	case "Q":
		client += "BTQueue"
	case "R":
		client += "Tribler"
	case "S":
		client += "Shadow's"
	case "T":
		client += "BitTornado"
	case "U":
		client += "UPnP NAT Bit Torrent"
	default:
		client += shadow[0:1]
	}

	client += " v"

	for _, c := range shadow[1:6] {
		if string(c) == "-" {
			break
		}
		if int(c) > 47 && int(c) < 58 { // num
			client += string(c) + "."
		} else if int(c) > 64 && int(c) < 91 { // letter
			client += fmt.Sprintf("%d.", int(c)-55)
		} else {
			panic("invalid version char")
		}
	}

	return client[0 : len(client)-1]
}
