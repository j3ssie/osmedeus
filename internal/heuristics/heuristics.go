package heuristics

import (
	"net"
	"net/url"
	"regexp"
	"strings"
)

// TargetType represents the detected type of a target
type TargetType string

const (
	TargetTypeURL     TargetType = "url"
	TargetTypeDomain  TargetType = "domain"
	TargetTypeIP      TargetType = "ip"
	TargetTypeUnknown TargetType = "unknown"
)

// TargetInfo holds parsed information about a target
type TargetInfo struct {
	Type     TargetType
	Original string

	// URL fields (populated when Type == TargetTypeURL)
	BaseURL    string // https://sub.example.com:443/foo/bar.php
	RootURL    string // https://sub.example.com:443
	Hostname   string // sub.example.com:443
	RootDomain string // example.com
	TLD        string // com (or co.uk for multi-part TLDs)
	SLD        string // Second-level domain (example from example.com or mail.corp.example.com)
	Host       string // sub.example.com
	Port       string // 443
	Path       string // /foo
	File       string // bar.php
	Scheme     string // https

	// Domain fields
	IsWildcard bool
	ResolvedIP string // advanced only

	// URL advanced fields
	StatusCode    int   // advanced only
	ContentLength int64 // advanced only
}

// Analyze performs heuristic analysis on a target
// level can be: "none", "basic", "advanced"
func Analyze(target string, level string) (*TargetInfo, error) {
	if level == "none" {
		return &TargetInfo{
			Type:     TargetTypeUnknown,
			Original: target,
		}, nil
	}

	targetType := DetectType(target)
	var info *TargetInfo
	var err error

	switch targetType {
	case TargetTypeURL:
		info, err = ParseURL(target)
		if err != nil {
			return nil, err
		}
		// Advanced: fetch URL info
		if level == "advanced" {
			statusCode, contentLength, fetchErr := FetchURLInfo(target)
			if fetchErr == nil {
				info.StatusCode = statusCode
				info.ContentLength = contentLength
			}
		}

	case TargetTypeDomain:
		info, err = ParseDomain(target)
		if err != nil {
			return nil, err
		}
		// Basic: check wildcard
		info.IsWildcard = CheckWildcard(target)
		// Advanced: resolve IP
		if level == "advanced" {
			info.ResolvedIP, _ = ResolveIP(target)
		}

	case TargetTypeIP:
		info = &TargetInfo{
			Type:       TargetTypeIP,
			Original:   target,
			Host:       target,
			RootDomain: target,
		}

	default:
		info = &TargetInfo{
			Type:     TargetTypeUnknown,
			Original: target,
		}
	}

	return info, nil
}

// DetectType determines the type of target
func DetectType(target string) TargetType {
	target = strings.TrimSpace(target)

	// Check if it's a URL (has scheme)
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return TargetTypeURL
	}

	// Check if URL without scheme but has path
	if strings.Contains(target, "/") {
		// Could be URL without scheme, try parsing
		if _, err := url.Parse("https://" + target); err == nil {
			return TargetTypeURL
		}
	}

	// Check if it's an IP address
	if isIPAddress(target) {
		return TargetTypeIP
	}

	// Check if it's a valid domain
	if isValidDomain(target) {
		return TargetTypeDomain
	}

	return TargetTypeUnknown
}

// isIPAddress checks if the string is an IP address
func isIPAddress(s string) bool {
	// Remove port if present
	host := s
	if idx := strings.LastIndex(s, ":"); idx != -1 {
		host = s[:idx]
	}
	return net.ParseIP(host) != nil
}

// isValidDomain checks if the string is a valid domain
func isValidDomain(s string) bool {
	// Basic domain validation regex
	domainRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
	return domainRegex.MatchString(s)
}
