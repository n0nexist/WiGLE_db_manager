# :rocket: WiGLE-db-manager
![](https://github.com/n0nexist/WiGLE-db-manager/blob/main/screenshot.png?raw=true)<br><br>
<b>Fast csv database to html map converter for wigle<b>

# :arrow_down: Download 
```
git clone https://github.com/n0nexist/WiGLE-db-manager
cd WiGLE-db-manager
go build main.go
./main
```

# :mag: Usage
<code>./main -d (database.csv) -a (action)</code>

<b>Actions:</b>
```
'none' - no filtering
'open' - shows open wifi networks
'wep' - shows wep wifi networks
'wps' - shows wifi networks with wps enabled
'bt' - shows bluetooth devices
'ble' - shows bluetooth LE devices
'wifi' - shows only wifi networks
'bluetooth' - shows only bt and bt LE devices
```

<b>Optional arguments:</b>
```
-b  --batch-size         The lines at a time to process (default 1000).
                         Default: 1000
-s  --filter-ssid        Filter out an ssid from the results. Default:
                         nofilter
-m  --filter-bssid       Filter out a bssid from the results. Default:
                         nofilter
-r  --remove-duplicates  Removes duplicate bssid's. Default: false
```
