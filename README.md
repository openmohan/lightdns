# lightdns

This is a light weight DNS Server which serves A type record for now.

## Installation
`go get github.com/openmohan/lightdns`

### Requirements
* Go 1.90

## Usage
1. Import the package `	lightdns "github.com/openmohan/lightdns"`
2. Create a DNS Server `dns := lightdns.NewDNSServer(port_int)`
3. Add Zone data using the function `AddZoneData`

```dns.AddZoneData(zone, staticRecords, lookupFunction, ZoneType)```

## Example
If the records data available we can code it and Add the zone data
```	var googleRecords = map[string]string{
		"mail.google.com":  "192.168.0.2",
		"paste.google.com": "192.168.0.3",
	}
	dns.AddZoneData("google.com", googleRecords, nil, lightdns.DNSForwardLookupZone)
```

or
 
If the data is not static and has to be taken from a DB or from any other sources or calculated dynamically then use lookup function.
Define a lookup function in the format 

`func lookupFunc(domain string) (ip string, err error)` ie: given a domain name string as parameter, IP string and error value has to be returned.
```
	func lookupFunc(string) (string, error) {
		//Do some action
		//Get data from DB
		//Process it further more
		return "192.2.2.1", nil
	}
	dns.AddZoneData("amazon.com", nil, lookupFunc, lightdns.DNSForwardLookupZone)
```


# Complete Example
```
package main

import (
	lightdns "github.com/openmohan/lightdns"
)

var records = map[string]string{
	"mail.amazon.com":  "192.162.1.2",
	"paste.amazon.com": "191.165.0.3",
}

func lookupFunc(string) (string, error) {
	//Do some action
	//Get data from DB
	//Process it further more
	return "192.2.2.1", nil
}

func main() {
	var googleRecords = map[string]string{
		"mail.google.com":  "192.168.0.2",
		"paste.google.com": "192.168.0.3",
	}
	var microsoftRecords = map[string]string{
		"mail.microsoft.com":  "192.168.0.78",
		"paste.microsoft.com": "192.168.0.25",
	}
	dns := lightdns.NewDNSServer(1234)
	dns.AddZoneData("google.com", googleRecords, nil, lightdns.DNSForwardLookupZone)
	dns.AddZoneData("microsoft.com", microsoftRecords, nil, lightdns.DNSForwardLookupZone)

	/* Incase if the records are not static or to be taken from DB or from any other sources
	lookupFunc method can be used.append*/
	dns.AddZoneData("amazon.com", nil, lookupFunc, lightdns.DNSForwardLookupZone)
	dns.StartToServe()
}

```

## Work in progress
### Support for
1. MX, AAAA, SOA and other records
2. Reverse lookup zone

### Code
 1. Use goroutines for optimizing the speed
 2. Unit Test Cases
 3. Logging mechanism for monitoring and debugging


## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.


## License
[MIT](https://choosealicense.com/licenses/mit/)