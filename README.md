# lightdns

This is a light weight DNS Server which serves A type record for now.

## Installation
`go get github.com/openmohan/lightdns`

### Requirements
* Go 1.90

## Usage

```
import (
	dns "github.com/openmohan/lightdns"
)

func main() {
	var records = map[string]string{
		"mail.google.com":  "192.168.0.2",
		"paste.google.com": "192.168.0.3",
	}
	var microrecords = map[string]string{
		"mail.microsoft.com":  "192.168.0.78",
		"paste.microsoft.com": "192.168.0.25",
	}
	dns := dns.NewDNSServer(1234)
	dns.AddZoneData("google.com", records)
	dns.AddZoneData("microsoft.com", microrecords)
	dns.StartToServe()
}

```

## Work in progress
### Support for
```
MX, AAAA, SOA and other records
Reverse lookup zone
```
### Code
 1. Use goroutines for optimizing the speed
 2. Unit Test Cases



## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.


## License
[MIT](https://choosealicense.com/licenses/mit/)