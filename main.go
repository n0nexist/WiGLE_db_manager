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
	"encoding/csv"

	"github.com/akamensky/argparse"
)

var (
	version = "1.1.1"
	author = "n0nexist"
	mapEntries = []string{}
	mu sync.Mutex
	map_path = "ciao.html"
)

func makeTimestamp() int64 {
    return time.Now().UnixNano() / 1e6
}

func addString(strToAdd string) {
    mapEntries = append(mapEntries, strToAdd)
}

func isNetworkOpen(authMode string) bool {
    return !(strings.Contains(authMode, "WPA2") || strings.Contains(authMode, "WPA") || strings.Contains(authMode, "WEP"))
}

func addPointInMap(info, latitude, longitude, icon string){
	addString(fmt.Sprintf("addMarker(%s,%s,\"%s\", \"%s\")", latitude, longitude, icon, info))
}

func removeDuplicateBSSID(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	newFileName := "new_" + fileName // Nome del nuovo file con contenuto non duplicato
	newFile, err := os.Create(newFileName)
	if err != nil {
		return err
	}
	defer newFile.Close()

	scanner := bufio.NewScanner(file)
	bssidSet := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "<b>BSSID:</b>") && strings.Contains(line, "<br>") {
			parts := strings.Split(line, "<b>BSSID:</b>")
			if len(parts) > 1 {
				bssid := strings.TrimSpace(strings.Split(parts[1], "<br>")[0])
				if bssidSet[bssid] {
					continue
				}
				bssidSet[bssid] = true
			}
		}

		// Scrivi la riga nel nuovo file se non è un duplicato
		_, err := fmt.Fprintln(newFile, line)
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return os.Rename(newFileName, fileName) // Rinomina il nuovo file con il nome del file originale
}

func removeExistingMap() {
    if _, err := os.Stat(map_path); os.IsNotExist(err) {
        throw(err)
    }
    err := os.Remove(map_path)
    throw(err)
}

func shouldFilter(filterssid, filtermac, ssid, mac string) bool {
	return (filterssid == ssid && filterssid != "nofilter" ) || (filtermac == mac && filtermac != "nofilter")
}

func getIconString(device_type, auth_mode string) string{
	if device_type == "BT" {
		return "blue"
	} else if device_type == "BLE" {
		return "violet"
	} else if device_type == "GSM" {
		return "grey"
	} else if device_type == "WIFI" {
		if isNetworkOpen(auth_mode) {
			return "green"
		} else if strings.Contains(auth_mode, "WEP") {
			return "yellow"
		} else if strings.Contains(auth_mode, "WPS") {
			return "red"
		} else if strings.Contains(auth_mode, "WPA") {
			return "red"
		} else{
			return "white"
		}
	} else {
		return "shadow"
	}
}

func processLine(line, action, filterssid, filtermac string) {
	r := csv.NewReader(strings.NewReader(line))
	r.Comma = ','

	parts, err := r.Read()
	if err != nil {
		fmt.Println("Error reading line", err)
		return
	}

	if len(parts) < 9 {
		fmt.Println("Line does not contain enough fields")
		return
	}

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

func throw(err error){
	if err != nil {
        log.Println(err)
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

	file, err := os.Create(map_path)
	throw(err)
	defer file.Close()

	_, err = file.WriteString(`<!DOCTYPE html><html><head><title>WiGLE db manager</title><link rel="stylesheet" href="https://unpkg.com/leaflet@1.7.1/dist/leaflet.css" /><script src="https://unpkg.com/leaflet@1.7.1/dist/leaflet.js"></script></head><body><div id="map" style="height: 100vh; width: 100%;"></div><script>var map = L.map('map').setView([53.0000, 9.0000], 3);L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'}).addTo(map);function addMarker(lat, lon, color, tooltip) {var marker = L.marker([lat, lon], {icon: L.icon({iconUrl: 'https://github.com/pointhi/leaflet-color-markers/blob/master/img/marker-icon-' + color + '.png?raw=true',shadowUrl: 'https://github.com/pointhi/leaflet-color-markers/blob/master/img/marker-icon-shadow.png?raw=true',iconSize: [18, 25]})}).bindTooltip(tooltip).addTo(map);marker.on('click', function (e) {var textArea = document.createElement("textarea");var textToCopy = lat + " " + lon;textArea.value = textToCopy;document.body.appendChild(textArea);textArea.select();document.execCommand('copy');document.body.removeChild(textArea);alert("Coordinates copied to clipboard: " + textToCopy);});}
`)
	throw(err)

	for _, s := range mapEntries {
		_, err = file.WriteString(s + "\n")
		throw(err)
	}
	_, err = file.WriteString("</script></body></html>")
	throw(err)

}

func checkDuplicates(dupl string){
	if (dupl=="yes") {
		log.Printf("Removing duplicates from %s\n", map_path)
		err := removeDuplicateBSSID(map_path)
		throw(err)
	}
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
	mappath := parser.String("p", "path", &argparse.Options{Required: false, Help: "The generated map path", Default: "index.html"})
	duplicates := parser.String("r", "remove-duplicates", &argparse.Options{Required: false, Help: "Remove duplicate bssids [no/yes]", Default: "no"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	startTime := makeTimestamp()

	map_path = *mappath

	loadDB(*db,*act,*filterssid,*filtermac,*bsize)

	log.Printf("Writing to %s\n", map_path)
	
	removeExistingMap()
	genHtmlMap()

	checkDuplicates(*duplicates)
	
	log.Printf("Whole operation took %dms", makeTimestamp()-startTime)
	log.Println("GoodBye")

}
