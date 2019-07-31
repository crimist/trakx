package shared

import (
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"
)

// https://wiki.theory.org/index.php/BitTorrentSpecification#peer_id

var StatsHTML string

func (db *PeerDatabase) generateMetrics() {
	start := time.Now()
	Logger.Info("Generating metrics...")
	stats := make(map[string]int, 300)

	db.mu.RLock()
	for _, peermap := range db.db {
		for peerid := range peermap {
			if peerid[0] == '-' {
				stats[getAzureus(string(peerid[1:7]))]++
			} else {
				stats[getShadow(string(peerid[0:6]))]++
			}
		}
	}
	db.mu.RUnlock()

	// Get keys in alpha order
	keys := make([]string, len(stats))
	for key := range stats {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	StatsHTML = "<p>There will be errors in this list because some BT clients use *unique* (stupid) peerid encoding methods and I don't want to build every single edge case in</p><p>Read https://wiki.theory.org/index.php/BitTorrentSpecification#peer_id for more info</p><table>"
	for _, k := range keys {
		if count, ok := stats[k]; ok && count > 0 {
			StatsHTML += fmt.Sprintf(`<tr><td>%s</td><td>%d</td></tr>`, k, count)
		}
	}
	StatsHTML += "</table>"
	Logger.Info("Metric generated", zap.Duration("duration", time.Since(start)))
}

// get the full name of the client w/ azureus method
func getAzureus(azureus string) string {
	var client string

	switch string(azureus[0:2]) {
	case "7T":
		client += "aTorrent for Android"
	case "AB":
		client += "AnyEvent::BitTorrent"
	case "AG":
		client += "Ares"
	case "A~":
		client += "Ares"
	case "AR":
		client += "Arctic"
	case "AV":
		client += "Avicora"
	case "AT":
		client += "Artemis"
	case "AX":
		client += "BitPump"
	case "AZ":
		client += "Azureus"
	case "BB":
		client += "BitBuddy"
	case "BC":
		client += "BitComet"
	case "BE":
		client += "Baretorrent"
	case "BF":
		client += "Bitflu"
	case "BG":
		client += "BTG"
	case "BL":
		client += "BitCometLite (uses 6 digit version number) / BitBlinder"
	case "BP":
		client += "BitTorrent Pro (Azureus + spyware)"
	case "BR":
		client += "BitRocket"
	case "BS":
		client += "BTSlave"
	case "BT":
		client += "mainline BitTorrent (versions >= 7.9) / BBtor"
	case "Bt":
		client += "Bt"
	case "BW":
		client += "BitWombat"
	case "BX":
		client += "~Bittorrent X"
	case "CD":
		client += "Enhanced CTorrent"
	case "CT":
		client += "CTorrent"
	case "DE":
		client += "DelugeTorrent"
	case "DP":
		client += "Propagate Data Client"
	case "EB":
		client += "EBit"
	case "ES":
		client += "electric sheep"
	case "FC":
		client += "FileCroc"
	case "FD":
		client += "Free Download Manager (versions >= 5.1.12)"
	case "FT":
		client += "FoxTorrent"
	case "FX":
		client += "Freebox BitTorrent"
	case "GS":
		client += "GSTorrent"
	case "HK":
		client += "Hekate"
	case "HL":
		client += "Halite"
	case "HM":
		client += "hMule (uses Rasterbar libtorrent)"
	case "HN":
		client += "Hydranode"
	case "IL":
		client += "iLivid"
	case "JS":
		client += "Justseed.it client"
	case "JT":
		client += "JavaTorrent"
	case "KG":
		client += "KGet"
	case "KT":
		client += "KTorrent"
	case "LC":
		client += "LeechCraft"
	case "LH":
		client += "LH-ABC"
	case "LP":
		client += "Lphant"
	case "LT":
		client += "libtorrent"
	case "lt":
		client += "libTorrent"
	case "LW":
		client += "LimeWire"
	case "MK":
		client += "Meerkat"
	case "MO":
		client += "MonoTorrent"
	case "MP":
		client += "MooPolice"
	case "MR":
		client += "Miro"
	case "MT":
		client += "MoonlightTorrent"
	case "NB":
		client += "Net::BitTorrent"
	case "NX":
		client += "Net Transport"
	case "OS":
		client += "OneSwarm"
	case "OT":
		client += "OmegaTorrent"
	case "PB":
		client += "Protocol::BitTorrent"
	case "PD":
		client += "Pando"
	case "PI":
		client += "PicoTorrent"
	case "PT":
		client += "PHPTracker"
	case "qB":
		client += "qBittorrent"
	case "QD":
		client += "QQDownload"
	case "QT":
		client += "Qt 4 Torrent example"
	case "RT":
		client += "Retriever"
	case "RZ":
		client += "RezTorrent"
	case "S~":
		client += "Shareaza alpha/beta"
	case "SB":
		client += "~Swiftbit"
	case "SD":
		client += "Thunder (aka XùnLéi)"
	case "SM":
		client += "SoMud"
	case "SP":
		client += "BitSpirit"
	case "SS":
		client += "SwarmScope"
	case "ST":
		client += "SymTorrent"
	case "st":
		client += "sharktorrent"
	case "SZ":
		client += "Shareaza"
	case "TB":
		client += "Torch"
	case "TE":
		client += "terasaur Seed Bank"
	case "TL":
		client += "Tribler (versions >= 6.1.0)"
	case "TN":
		client += "TorrentDotNET"
	case "TR":
		client += "Transmission"
	case "TS":
		client += "Torrentstorm"
	case "TT":
		client += "TuoTu"
	case "UL":
		client += "uLeecher!"
	case "UM":
		client += "µTorrent for Mac"
	case "UT":
		client += "µTorrent"
	case "VG":
		client += "Vagaa"
	case "WD":
		client += "WebTorrent Desktop"
	case "WT":
		client += "BitLet"
	case "WW":
		client += "WebTorrent"
	case "WY":
		client += "FireTorrent"
	case "XF":
		client += "Xfplay"
	case "XL":
		client += "Xunlei"
	case "XS":
		client += "XSwifter"
	case "XT":
		client += "XanTorrent"
	case "XX":
		client += "Xtorrent"
	case "ZT":
		client += "ZipTorrent"
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
			if !Config.Trakx.Prod {
				Logger.Info("invalid version char", zap.Any("char", c))
			}
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
			if !Config.Trakx.Prod {
				Logger.Info("invalid version char", zap.Any("char", c))
			}
		}
	}

	return client[0 : len(client)-1]
}
