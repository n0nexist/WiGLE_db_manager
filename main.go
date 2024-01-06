/*
WiGLE-db-manager by n0nexist
github.com/n0nexist/WiGLE-db-manager
*/

package main

import (
    "bufio"
	"fmt"
	"os"
	"log"
	"sync"
	"strings"
	"time"

	"github.com/akamensky/argparse"
)

var (
	version = "1.0.1"
	author = "n0nexist"
	mapEntries = []string{}
	foundBssids = []string{}
	mu sync.Mutex
	removeDuplicates = false
)

func makeTimestamp() int64 {
    return time.Now().UnixNano() / 1e6
}

func addNetwork(strToAdd string) {
    foundBssids = append(foundBssids, strToAdd)
}

func containsNetwork(search string) bool {
    mu.Lock()
    defer mu.Unlock()

    for _, str := range foundBssids {
        if str == search {
            return true
        }
    }

    return false
}

func addString(strToAdd string) {
    mapEntries = append(mapEntries, strToAdd)
}

func isNetworkOpen(authMode string) bool {
    return !(strings.Contains(authMode, "WPA2") || strings.Contains(authMode, "WPA") || strings.Contains(authMode, "WEP"))
}

func addPointInMap(info, latitude, longitude, icon string){
	addString(fmt.Sprintf("L.marker([%s, %s], {icon: %s}).addTo(map).bindTooltip(\"%s\");", latitude, longitude, icon, info))
}

func removeExistingMap() {
	filePath := "map/index.html"
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        throw(err)
    }
    err := os.Remove(filePath)
    throw(err)
}

func shouldFilter(filterssid, filtermac, ssid, mac string) bool {
	return (filterssid == ssid && filterssid != "nofilter" ) || (filtermac == mac && filtermac != "nofilter")
}

func getIconString(device_type, auth_mode string) string{
	if device_type == "BT" {
		return "bluetoothIcon"
	} else if device_type == "BLE" {
		return "bluetoothLeIcon"
	} else if device_type == "GSM" {
		return "gsmIcon"
	} else if device_type == "WIFI" {
		if isNetworkOpen(auth_mode) {
			return "openwifiIcon"
		} else if strings.Contains(auth_mode, "WEP") {
			return "wepwifiIcon"
		} else if strings.Contains(auth_mode, "WPS") {
			return "wpswifiIcon"
		} else if strings.Contains(auth_mode, "WPA") {
			return "wpawifiIcon"
		} else{
			return "whiteIcon"
		}
	} else {
		return "whiteIcon"
	}
}

func processLine(line, action, filterssid, filtermac string) {
	parts := strings.Split(line, ",")
	mac := parts[0]
	ssid := parts[1]
	auth_mode := parts[2]
	first_seen := parts[3]
	channel := parts[4]
	rssi := parts[6]
	latitude := parts[7]
	longitude := parts[8]
	device_type := parts[len(parts)-1]

	device_icon := getIconString(device_type, auth_mode)
	info := fmt.Sprintf(`<b>SSID:</b> %s<br><b>BSSID:</b> %s<br><b>OPEN:</b> %t<br><b>AUTH:</b> %s<br><b>SEEN:</b> %s<br><b>CH:</b> %s<br><b>RSSI:</b> %s<br><b>COORDS:</b> %s, %s<br><b>TYPE:</b> %s`, ssid, mac, isNetworkOpen(auth_mode), auth_mode, first_seen, channel, rssi, latitude, longitude, device_type)


	if !(shouldFilter(filterssid, filtermac, ssid, mac)){

		if removeDuplicates {
			if containsNetwork(mac){
				return
			}else{
				addNetwork(mac)
			}
		}

		if action == "none" {
			addPointInMap(info, latitude, longitude, device_icon)

		} else if action == "open" {

			if isNetworkOpen(auth_mode) && device_type == "WIFI" {
				addPointInMap(info, latitude, longitude, device_icon)
			}
			
		} else if action == "wep" {

			if strings.Contains(auth_mode, "WEP") && device_type == "WIFI" {
				addPointInMap(info, latitude, longitude, device_icon)
			}
			
		} else if action == "wps" {

			if strings.Contains(auth_mode, "WPS") && device_type == "WIFI" {
				addPointInMap(info, latitude, longitude, device_icon)
			}
			
		} else if action == "bt" {

			if device_type == "BT" {
				addPointInMap(info, latitude, longitude, device_icon)
			}
			
		} else if action == "ble" {

			if device_type == "BLE" {
				addPointInMap(info, latitude, longitude, device_icon)
			}
			
		} else if action == "wifi" {

			if device_type == "WIFI" {
				addPointInMap(info, latitude, longitude, device_icon)
			}
			
		} else if action == "bluetooth" { 

			if device_type == "BT" || device_type == "BLE" {
				addPointInMap(info, latitude, longitude, device_icon)
			}
			
		}

	}
    
}

func throw(err error){
	if err != nil {
        log.Fatal(err)
    }
}

func loadDB(filePath, action, filterssid, filtermac string, lines_at_a_time int){
	log.Printf("Opening '%s'...\n",filePath)
	readFile, err := os.Open(filePath)
	throw(err)

	fileScanner := bufio.NewScanner(readFile)
    fileScanner.Split(bufio.ScanLines)
    var fileLines []string

	log.Println("Loading file into memory...")
    for fileScanner.Scan() {
        fileLines = append(fileLines, fileScanner.Text())
    }

    readFile.Close()

	log.Println("File loaded into memory")


    var wg sync.WaitGroup

    for i := 0; i < len(fileLines); i += lines_at_a_time {
        end := i + lines_at_a_time
        if end > len(fileLines) {
            end = len(fileLines)
        }
        linesToPrint := fileLines[i:end]
        wg.Add(1)
        go func(lines []string) {
            defer wg.Done()
            for _, line := range lines {
                processLine(line, action, filterssid, filtermac)
            }
        }(linesToPrint)
    }

    wg.Wait()
}

func genHtmlMap(){

	file, err := os.Create("map/index.html")
	throw(err)
	defer file.Close()

	_, err = file.WriteString(`<!DOCTYPE html><html><head><title>WiGLE db manager</title><link rel="stylesheet" href="https://unpkg.com/leaflet@1.7.1/dist/leaflet.css" /><script src="https://unpkg.com/leaflet@1.7.1/dist/leaflet.js"></script></head><body><div id="map" style="height: 100vh; width: 100%;"></div><script>var map = L.map('map').setView([53.0000, 9.0000], 3);L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'}).addTo(map);var bluetoothIcon = L.icon({iconUrl: 'icons/blue.png',iconSize:[18, 25]});var bluetoothLeIcon = L.icon({iconUrl: 'icons/cyan.png',iconSize:[18, 25]});var gsmIcon = L.icon({iconUrl: 'icons/black.png',iconSize:[18, 25]});var openwifiIcon = L.icon({iconUrl: 'icons/green.png',iconSize:[18, 25]});var wpswifiIcon = L.icon({iconUrl: 'icons/yellow.png',iconSize:[18, 25]});var wepwifiIcon = L.icon({iconUrl: 'icons/orange.png',iconSize:[18, 25]});var whiteIcon = L.icon({iconUrl: 'icons/white.png',iconSize:[18, 25]});var wpawifiIcon = L.icon({iconUrl: 'icons/red.png',iconSize:[18, 25]});`)
	throw(err)

	for _, s := range mapEntries {
		_, err = file.WriteString(s + "\n")
		throw(err)
	}
	_, err = file.WriteString("</script></body></html>")
	throw(err)

}

func main() {

	fmt.Println("Welcome to WiGLE local database manager")
	fmt.Printf("Version %s by %s\n", version, author)

	parser := argparse.NewParser("WiGLE-db-manager", "fast local WiGLE database manager")

	db := parser.String("d", "database", &argparse.Options{Required: true, Help: "The database path"})
	act := parser.String("a", "action", &argparse.Options{Required: true, Help: "The action to execute\n\nActions: 'none' - no filtering\n 'open' - shows open wifi networks\n 'wep' - shows wep wifi networks\n 'wps' - shows wifi networks with wps enabled\n 'bt' - shows bluetooth devices\n 'ble' - shows bluetooth LE devices\n 'wifi' - shows only wifi networks\n 'bluetooth' - shows only bt and bt LE devices\n"})
	bsize := parser.Int("b", "batch-size", &argparse.Options{Required: false, Help: "The lines at a time to process (default 1000)", Default: 1000})
	filterssid := parser.String("s", "filter-ssid", &argparse.Options{Required: false, Help: "Filter out an ssid from the results", Default: "nofilter"})
	filtermac := parser.String("m", "filter-bssid", &argparse.Options{Required: false, Help: "Filter out a bssid from the results", Default: "nofilter"})
	var tmp *bool = parser.Flag("r", "remove-duplicates", &argparse.Options{Required: false, Help: "Removes duplicate bssid's", Default: false})
	if tmp != nil {
		removeDuplicates = *tmp 
	} else {
		removeDuplicates = false
	}

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	startTime := makeTimestamp()

	loadDB(*db,*act,*filterssid,*filtermac,*bsize)

	log.Println("Writing to map/index.html")
	
	removeExistingMap()
	genHtmlMap()

	log.Printf("Whole operation took %dms", makeTimestamp()-startTime)
	log.Println("GoodBye")

}
