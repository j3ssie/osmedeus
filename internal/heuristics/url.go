package heuristics

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"golang.org/x/net/publicsuffix"
)

// ParseURL parses a URL and extracts all relevant fields
func ParseURL(rawURL string) (*TargetInfo, error) {
	// Add scheme if missing
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	info := &TargetInfo{
		Type:     TargetTypeURL,
		Original: rawURL,
		Scheme:   u.Scheme,
	}

	// Extract host and port
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	info.Host = host
	info.Port = port
	info.Hostname = u.Host // includes port if specified

	// Extract root domain using publicsuffix
	rootDomain, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		// Fallback: use the last two parts of the domain
		rootDomain = extractRootDomainFallback(host)
	}
	info.RootDomain = rootDomain
	info.TLD, _ = publicsuffix.PublicSuffix(host)
	info.SLD = extractSLD(info.RootDomain, info.TLD)

	// Extract path and file
	urlPath := u.Path
	if urlPath == "" {
		urlPath = "/"
	}

	dir, file := path.Split(urlPath)
	info.Path = strings.TrimSuffix(dir, "/")
	if info.Path == "" {
		info.Path = "/"
	}
	info.File = file

	// Build RootURL (scheme + host + port)
	if (u.Scheme == "https" && port == "443") || (u.Scheme == "http" && port == "80") {
		info.RootURL = u.Scheme + "://" + host
	} else {
		info.RootURL = u.Scheme + "://" + host + ":" + port
	}

	// BaseURL is the original URL
	info.BaseURL = rawURL

	return info, nil
}

// extractRootDomainFallback extracts root domain when publicsuffix fails
func extractRootDomainFallback(host string) string {
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "." + parts[len(parts)-1]
	}
	return host
}

// FetchURLInfo fetches the URL and returns status code and content length
func FetchURLInfo(rawURL string) (int, int64, error) {
	// Add scheme if missing
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	// Create client with timeout and skip TLS verification
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	// Create HEAD request with User-Agent header
	req, err := http.NewRequest("HEAD", rawURL, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("User-Agent", core.DefaultUA)

	resp, err := client.Do(req)
	if err != nil {
		// Try GET if HEAD fails
		req, err = http.NewRequest("GET", rawURL, nil)
		if err != nil {
			return 0, 0, err
		}
		req.Header.Set("User-Agent", core.DefaultUA)

		resp, err = client.Do(req)
		if err != nil {
			return 0, 0, err
		}
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode, resp.ContentLength, nil
}
