package shared

import (
	"fmt"
)

var Azureus map[string]int
var Shadows map[string]int

func processMetrics() {
	for _, peermap := range PeerDB {
		for peerid := range peermap {
			if peerid[0] == '-' && peerid[7] == '-' {
				Azureus[string(peerid[1:6])]++
			}
		}
	}
}

// get the full name of the client w/ azureus method
func getAzureus(client string) {

}

// get the full name of the client w/ shadow method
func getShadow(shadow string) {
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
		client += "Shadow's client"
	case "T":
		client += "BitTornado"
	case "U":
		client += "UPnP NAT Bit Torrent"
	}

	client += " v"

	for _, c := range shadow[1:] {
		if string(c) == "-" {
			break
		}
		if int(c) > 47 && int(c) < 58 { // num
			client += fmt.Sprintf("%d.", int(c) - 48)
		} else if int(c) > 64 && int(c) < 91 { // letter
			client += fmt.Sprintf("%d.", int(c) - 55)
		} else {
			panic("invalid version char")
		}
	}

	client = client[0:len(client)-1]

	fmt.Println(client)
}
