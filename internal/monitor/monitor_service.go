package monitor

import (
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func NewMonitor() MoniterService {

	m := &Moniter{
		websites: make(map[uint]*Website),
		results:  make(chan WebsiteResult, 100),
		Client: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}

	return m
}

func (m *Moniter) AddWebsite(w *Website) {
	id := rand.Intn(10000)
	w.ID = uint(id)
	m.websites[w.ID] = w
}

/*
monitor() sends a GET request to the specified website,
and checks for( response time, ssl certificate) sends the result to the result (channel) to be displied
*/
func (m *Moniter) Monitor(w *Website) {

	start := time.Now()
	resp, err := m.Client.Get(w.URL)
	responseTime := time.Since(start).Seconds()
	result := WebsiteResult{
		WebsiteID:    w.ID,
		URL:          w.URL,
		Message:      "website is up!",
		ResponseTime: responseTime,
	}
	if err != nil {
		result.Status = http.StatusInternalServerError
		result.Message = err.Error()
		log.Fatalf("error: %s", err)
	} else {
		result.Status = resp.StatusCode
		resp.Body.Close()
	}
	if err := m.CheckSSl(w); err != nil {
		result.SSLCertificate = false
		result.Status = http.StatusInternalServerError
		result.Message = err.Error()
		log.Fatalf("error: %s", err)
	}
	result.SSLCertificate = true
	m.results <- result
}

func (m *Moniter) Start() {
	log.Println("running")
	for _, website := range m.websites {
		log.Printf("Monitoring:%s", website.URL)
		go func(w *Website) {
			ticker := time.NewTicker(time.Duration(w.Interval) * time.Second)
			for range ticker.C {
				m.Monitor(w)
			}
		}(website)
	}
}

// Checking for SSL/TLS certificate
func (m *Moniter) CheckSSl(w *Website) error {
	conn, err := tls.Dial("tcp", strings.TrimPrefix(w.URL, "https://")+":433", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return fmt.Errorf("no certificate found")

	}
	cert := certs[0]

	//check if cert expired
	if time.Now().After(cert.NotAfter) {
		return fmt.Errorf("certificate has expired")
	}
	//check if cert will expire in 30 days
	if time.Now().Add(30 * 24 * time.Hour).After(cert.NotAfter) {
		log.Printf("Warning: %s cert will expire in a month", w.URL)

	}
	if err := cert.VerifyHostname(w.URL); err != nil {
		return fmt.Errorf("certificate is not valid for the domain %v", err)
	}
	return nil
}

// display the results that was sent in the channel
func (m *Moniter) Display() {
	for result := range m.results {
		log.Printf("Website %s: Message %s Status %d, Response Time %.2fs\n",
			result.URL, result.Message, result.Status, result.ResponseTime)
		fmt.Println("--------------------------------------------")

		if result.Status != http.StatusOK {
			log.Printf("Alert: Website %s is down. Status: %d\n", result.URL, result.Status)
			fmt.Println("-----------------------------------------")

		}
	}
}

// func (m *Moniter) TriggerAlert(result WebsiteResult) {
// 	log.Printf("Alert: Website %s is down. Status: %d\n", result.URL, result.Status)
// }
