package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/yanc0/beeping/sslcheck"
	"github.com/yanc0/greedee/collectd"
	"log"
	"net/http"
	"os"
	"time"
)

// TODO: use a BeePing package
type Check struct {
	URL      string        `json:"url" binding:"required"`
	Pattern  string        `json:"pattern"`
	Header   string        `json:"header"`
	Insecure bool          `json:"insecure"`
	Timeout  time.Duration `json:"timeout"`
	Auth     string        `json:"auth"`
}

type Timeline struct {
	NameLookup    int64 `json:"name_lookup"`
	Connect       int64 `json:"connect"`
	Pretransfer   int64 `json:"pretransfer"`
	StartTransfer int64 `json:"starttransfer"`
}

type Geo struct {
	Country string `json:"country"`
	City    string `json:"city,omitempty"`
	IP      string `json:"ip"`
}

type Response struct {
	HTTPStatus      string `json:"http_status"`
	HTTPStatusCode  int    `json:"http_status_code"`
	HTTPBodyPattern bool   `json:"http_body_pattern"`
	HTTPHeader      bool   `json:"http_header"`
	HTTPRequestTime int64  `json:"http_request_time"`

	InstanceName string `json:"instance_name"`

	DNSLookup        int64 `json:"dns_lookup"`
	TCPConnection    int64 `json:"tcp_connection"`
	TLSHandshake     int64 `json:"tls_handshake,omitempty"`
	ServerProcessing int64 `json:"server_processing"`
	ContentTransfer  int64 `json:"content_transfer"`

	Timeline *Timeline          `json:"timeline"`
	Geo      *Geo               `json:"geo,omitempty"`
	SSL      *sslcheck.CheckSSL `json:"ssl,omitempty"`
}

var beepingURL *string
var checkURL *string
var timeout *int64
var pattern *string
var host *string
var greedeeURL *string
var greedeeUser *string
var greedeePass *string

func main() {

	beepingURL = flag.String("beeping", "http://localhost:8080", "URL of your BeePing instance")
	checkURL = flag.String("check", "", "URL we want to check")
	timeout = flag.Int64("timeout", 20, "BeePing check timeout")
	pattern = flag.String("pattern", "", "pattern that's need to be found in the body")
	host = flag.String("host", "test.test.prod.host.beeping", "Collectd metric's host entry, for example:'customer.app.env.servername'")
	greedeeURL = flag.String("greedee", "http://localhost:9223", "URL of your Greedee instance")
	greedeeUser = flag.String("greedeeUser", "", "Greedee user if configured with basic auth")
	greedeePass = flag.String("greedeePass", "", "Greedee password if configured with basic auth")
	flag.Parse()

	if *checkURL == "" {
		fmt.Printf("Usage:\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	bpResp, err := requestBeepingCheck()

	if err != nil {
		log.Fatal(fmt.Printf("[FATAL] Beeping check failed: %s\n\n", err))
	}

	sendCollectdMetricsToGreedee(bpResp)
}

func requestBeepingCheck() (*Response, error) {
	check := &Check{
		*checkURL,
		*pattern,
		"",
		false,
		time.Duration(*timeout) * time.Second,
		"",
	}

	jsonReq, _ := json.Marshal(check)
	resp, err := http.Post(*beepingURL+"/check", "application/json", bytes.NewBuffer(jsonReq))

	if err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()
		response := &Response{}
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&response)
		if err != nil {
			return nil, err
		}
		return response, nil
	}
}

func createCMetric(timeNow int64, cMType string, cMValues float64) *collectd.CollectDMetric {
	cMetric := collectd.CollectDMetric{
		Host:           *host,
		Plugin:         "beeping",
		PluginInstance: "",
		Type:           cMType,
		TypeInstance:   "",
		Time:           float64(timeNow),
		Interval:       30, // why ?
		DSTypes:        []string{"gauge"},
		DSNames:        []string{"value"},
		Values:         []float64{cMValues},
	}
	return &cMetric
}

func convertBoolToCMetricVal(value bool) float64 {
	if value {
		return float64(1)
	} else {
		return float64(0)
	}
}
// Transform Beeping Response to a map of collectd.Metrics and send it to Greedee
func sendCollectdMetricsToGreedee(bpResp *Response) {


	cMetrics := []*collectd.CollectDMetric{}
	timeNow := time.Now().Unix()


	cMetrics = append(cMetrics, createCMetric(timeNow, "http_status_code", float64(bpResp.HTTPStatusCode)))
	cMetrics = append(cMetrics, createCMetric(timeNow, "http_body_pattern", convertBoolToCMetricVal(bpResp.HTTPBodyPattern)))
	cMetrics = append(cMetrics, createCMetric(timeNow, "http_request_time", float64(bpResp.HTTPRequestTime)))
	cMetrics = append(cMetrics, createCMetric(timeNow, "dns_lookup", float64(bpResp.DNSLookup)))
	cMetrics = append(cMetrics, createCMetric(timeNow, "tcp_connection", float64(bpResp.TCPConnection)))
	cMetrics = append(cMetrics, createCMetric(timeNow, "tls_handshake", float64(bpResp.TLSHandshake)))
	cMetrics = append(cMetrics, createCMetric(timeNow, "content_transfer", float64(bpResp.ContentTransfer)))
	cMetrics = append(cMetrics, createCMetric(timeNow, "server_processing", float64(bpResp.ServerProcessing)))
	if bpResp.SSL != nil {
		cMetrics = append(cMetrics, createCMetric(timeNow, "cert_expiry_days_left", float64(bpResp.SSL.CertExpiryDaysLeft)))
	}

	cMetricsJson, _ := json.Marshal(cMetrics)
	httpClient := &http.Client{}
	gdReq, _ := http.NewRequest("POST", *greedeeURL+"/metrics", bytes.NewBuffer(cMetricsJson))
	gdReq.Header.Set("Content-Type", "application/json")

	if greedeeUser != nil && greedeePass != nil {
		gdReq.SetBasicAuth(*greedeeUser, *greedeePass)
	}

	_, err := httpClient.Do(gdReq)

	if err != nil {
		log.Fatal(fmt.Printf("[FATAL] Greedee Error: %s\n\n", err))
	}
}
