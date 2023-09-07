package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/j3ssie/osmedeus/libs"
	"io"
	"net"
	"net/http"
	"time"
)

var DefaultClient *http.Client
var (
	UA string
)

type Response struct {
	StatusCode int
	Status     string
	Body       string
}

func InitHTTPClient() {
	UA = fmt.Sprintf("Osmedeus/%s by %s", libs.VERSION, libs.AUTHOR)
	var transport = &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: time.Second,
		}).DialContext,
	}

	//if options.Proxy != "" {
	//	proxyUrl, err := url.Parse(options.Proxy)
	//	if err == nil {
	//		transport.Proxy = http.ProxyURL(proxyUrl)
	//	}
	//}

	DefaultClient = &http.Client{
		Transport: transport,
	}
}

// SendGET sending GET request
func SendGET(cred string, url string) (res Response) {
	DebugF("Sending GET request to: %v", url)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", UA)
	if cred != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", cred))
	}
	resp, err := DefaultClient.Do(req)

	if err != nil {
		ErrorF("Error sending to %v - %v", url, err)
		return res
	}

	defer resp.Body.Close()
	resbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return res
	}
	res.StatusCode = resp.StatusCode
	res.Status = resp.Status
	res.Body = string(resbody)
	return res
}

// SendPOST sending POST request
func SendPOST(cred string, url string, body string) (res Response) {
	DebugF("Sending POST request to: %v", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(body)))
	req.Header.Set("User-Agent", UA)
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", cred))
	req.Header.Set("Content-Type", "application/json")

	resp, err := DefaultClient.Do(req)
	if err != nil {
		ErrorF("Error sending to %v - %v", url, err)
		return res
	}
	defer resp.Body.Close()
	resbody, err := io.ReadAll(resp.Body)

	if err != nil {
		return res
	}
	res.StatusCode = resp.StatusCode
	res.Status = resp.Status
	res.Body = string(resbody)
	return res
}

// SendPUT sending POST request
func SendPUT(cred string, url string, body string) (res Response) {
	DebugF("Sending PUT request to: %v", url)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(body)))
	req.Header.Set("User-Agent", UA)
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", cred))
	req.Header.Set("Content-Type", "application/json")

	resp, err := DefaultClient.Do(req)
	if err != nil {
		ErrorF("Error sending to %v - %v", url, err)
		return res
	}
	defer resp.Body.Close()
	resbody, err := io.ReadAll(resp.Body)

	if err != nil {
		return res
	}
	res.StatusCode = resp.StatusCode
	res.Status = resp.Status
	res.Body = string(resbody)
	return res
}
