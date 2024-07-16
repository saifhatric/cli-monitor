package monitor

import (
	"net/http"
	"sync"
)

type Website struct {
	ID       uint
	URL      string `json:"url"`
	Interval uint   `json:"interval"` //time in seconds to recheck
}

type WebsiteResult struct {
	WebsiteID      uint
	URL            string  `json:"url"`
	ResponseTime   float64 `json:"response_time"`
	Status         int     `json:"status"`
	Message        string  `json:"message"`
	SSLCertificate bool    `json:"ssl_certificate"`
}

type Moniter struct {
	websites map[uint]*Website
	results  chan WebsiteResult
	Client   *http.Client
	sync.Mutex
}
type MoniterService interface {
	Start()
	Monitor(w *Website)
	Display()
	AddWebsite(w *Website)
}
