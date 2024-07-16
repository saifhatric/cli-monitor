package main

import (
	"flag"
	"strings"

	"github.com/saifhatric/monitor/internal/monitor"
)

func main() {
	website := flag.String("w", "https://example.com", "the website you want to monitor")
	timing := flag.Uint("t", 10, "interval to monitor default is 10s")
	flag.Parse()
	url := *website

	/*checking if the url start with https://
	if not add https:// to the url*/
	if !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	monitors := monitor.NewMonitor()
	monitors.AddWebsite(&monitor.Website{URL: url, Interval: *timing})

	go monitors.Start()

	monitors.Display()
}
