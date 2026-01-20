package functions

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/valyala/fastjson"
	"go.uber.org/zap"
)

// interestingUrls deduplicates URLs by hostname+path+param names and filters out static/noise patterns.
// Extracts unique, interesting URLs from a source file.
// Usage: interesting_urls(src, dest, json_field?) -> bool
//   - src: source file path (plain text URLs or JSONL)
//   - dest: destination file path for filtered URLs
//   - json_field: optional JSON field to extract URL from (for JSONL input)
//
// Filtering rules:
//  1. Static files are excluded: CSS, fonts, images (but .js is kept)
//  2. Noise patterns are excluded: blog, news, calendar dates, numeric-only paths
//  3. URLs are deduplicated by hostname + path + sorted parameter names (values ignored)
//  4. Path segments longer than 100 chars or with >3 dashes are filtered
func (vf *vmFunc) interestingUrls(call goja.FunctionCall) goja.Value {
	src := call.Argument(0).String()
	dest := call.Argument(1).String()
	jsonField := call.Argument(2).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("interesting_urls"),
		zap.String("src", src), zap.String("dest", dest), zap.String("json_field", jsonField))

	if src == "undefined" || src == "" || dest == "undefined" || dest == "" {
		logger.Get().Warn("interesting_urls: empty src or dest provided")
		return vf.vm.ToValue(false)
	}

	// Handle "undefined" json_field
	if jsonField == "undefined" {
		jsonField = ""
	}

	file, err := os.Open(src)
	if err != nil {
		logger.Get().Warn("interesting_urls: failed to open source file", zap.String("src", src), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = file.Close() }()

	var p fastjson.Parser
	data := make(map[string]string)        // hash -> original line/URL
	hostMapping := make(map[string]string) // hostname -> first URL seen (for noise fallback)

	var outputLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}

		var original string
		urlStr := raw

		// Handle JSON input - extract URL from specified field
		if jsonField != "" {
			v, err := p.Parse(raw)
			if err != nil {
				continue
			}
			original = raw
			urlStr = string(v.GetStringBytes(jsonField))
			if urlStr == "" {
				continue
			}
		}

		// Filter static files (but keep .js)
		if isStaticURL(urlStr) {
			continue
		}

		// Parse the URL
		u, err := url.Parse(urlStr)
		if err != nil || u.Hostname() == "" {
			continue
		}

		// Generate unique hash based on hostname + path + param names
		hash := hashURL(u)
		if hash == "" {
			continue
		}

		// Check if we've seen this hash before
		if _, exists := data[hash]; exists {
			continue
		}

		// Check noise patterns
		if isNoiseURL(urlStr) {
			// For noise URLs, only keep one per hostname
			if _, seen := hostMapping[u.Hostname()]; !seen {
				hostMapping[u.Hostname()] = urlStr
				if jsonField != "" {
					outputLines = append(outputLines, original)
				} else {
					outputLines = append(outputLines, urlStr)
				}
			}
			continue
		}

		// Store and output
		if jsonField != "" {
			data[hash] = original
			outputLines = append(outputLines, original)
		} else {
			data[hash] = urlStr
			outputLines = append(outputLines, urlStr)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Get().Warn("interesting_urls: error reading source", zap.String("src", src), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Write output
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		logger.Get().Warn("interesting_urls: failed to create dest directory", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	content := strings.Join(outputLines, "\n")
	if len(outputLines) > 0 {
		content += "\n"
	}
	if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
		logger.Get().Warn("interesting_urls: failed to write dest", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	logger.Get().Debug(terminal.HiGreen("interesting_urls")+" result",
		zap.String("src", src), zap.String("dest", dest),
		zap.Int("input_count", len(data)+len(hostMapping)),
		zap.Int("output_count", len(outputLines)))
	return vf.vm.ToValue(true)
}

// isStaticURL checks if URL points to a static file (CSS, fonts, images, etc.)
// Returns true if the URL should be excluded (static file)
// Note: .js files are NOT excluded
func isStaticURL(raw string) bool {
	// Static file extensions to exclude (case-insensitive)
	// Excludes: images, fonts, stylesheets, audio, video
	// Explicitly allows: .js files
	staticPattern := `(?i)\.(png|apng|bmp|gif|ico|cur|jpg|jpeg|jfif|pjp|pjpeg|svg|tif|tiff|webp|xbm|3gp|aac|flac|mpg|mpeg|mp3|mp4|m4a|m4v|m4p|oga|ogg|ogv|mov|wav|webm|eot|woff|woff2|ttf|otf|css)(?:\?|#|$)`
	return regexMatch(staticPattern, raw)
}

// isNoiseURL checks if URL matches noise patterns (blog, news, calendar, numeric content)
// Returns true if the URL is likely noise/uninteresting
func isNoiseURL(raw string) bool {
	// Calendar/date pattern: /2022/01/02/ or /2022-01-02/
	calendarPattern := `(\d{2,4})(-|/)(\d{1,2})(-|/)(\d{1,2})`
	if regexMatch(calendarPattern, raw) {
		return true
	}

	// Noise content paths (blog, news, articles, etc.)
	noiseContentPattern := `/(articles|about|blog|event|events|shop|post|posts|product|products|docs|support|pages|media|careers|jobs|video|videos|resource|resources|news)/.*`
	if regexMatch(noiseContentPattern, raw) {
		return true
	}

	// Numeric-only path segment without extension: /abc/1234
	idContentNoExtPattern := `.*\/[0-9]+$`
	if regexMatch(idContentNoExtPattern, raw) {
		return true
	}

	// Numeric-only path segment with extension: /abc/1234.html
	idContentPattern := `.*\/[0-9]+\.[a-z]+$`
	return regexMatch(idContentPattern, raw)
}

// hashURL generates a unique hash based on hostname + path + sorted parameter names
// Returns empty string if URL should be filtered out
func hashURL(u *url.URL) string {
	const maxPathItemLength = 100
	const maxDashCount = 3

	// Check path segments for length and dash count
	if strings.Count(u.Path, "/") >= 1 {
		paths := strings.Split(u.Path, "/")
		for _, item := range paths {
			if len(item) > maxPathItemLength || strings.Count(item, "-") > maxDashCount {
				return ""
			}
		}
	}

	// Extract and sort parameter names (ignore values)
	var paramNames []string
	for k := range u.Query() {
		paramNames = append(paramNames, k)
	}
	sort.Strings(paramNames)
	paramKey := strings.Join(paramNames, "-")

	// Create unique key: hostname-path-paramnames
	key := fmt.Sprintf("%v-%v-%v", u.Hostname(), u.Path, paramKey)
	return genSHA1(key)
}

// regexMatch checks if raw matches the given pattern
func regexMatch(pattern string, raw string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(raw)
}

// genSHA1 generates a SHA1 hash from text
func genSHA1(text string) string {
	h := sha1.New()
	h.Write([]byte(text))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// getIP resolves a domain or URL to its IP address.
// If input is a URL, extracts the hostname first.
// Usage: get_ip(domain_or_url) -> string (IP address or empty string on failure)
func (vf *vmFunc) getIP(call goja.FunctionCall) goja.Value {
	input := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("get_ip"),
		zap.String("input", input))

	if input == "undefined" || input == "" {
		logger.Get().Warn("get_ip: empty input provided")
		return vf.vm.ToValue("")
	}

	// Try to extract hostname from URL if input looks like a URL
	hostname := extractHostname(input)

	// Resolve the hostname to IP
	ip := resolveToIP(hostname)
	if ip == "" {
		logger.Get().Debug("get_ip: failed to resolve",
			zap.String("hostname", hostname))
	} else {
		logger.Get().Debug(terminal.HiGreen("get_ip")+" result",
			zap.String("input", input),
			zap.String("hostname", hostname),
			zap.String("ip", ip))
	}

	return vf.vm.ToValue(ip)
}

// extractHostname extracts the hostname from a URL or returns the input as-is if it's a domain.
// Handles: URLs (http://example.com/path), domains (example.com), domains with port (example.com:8080)
func extractHostname(input string) string {
	input = strings.TrimSpace(input)

	// If it looks like a URL (has scheme), parse it
	if strings.Contains(input, "://") {
		u, err := url.Parse(input)
		if err == nil && u.Hostname() != "" {
			return u.Hostname()
		}
	}

	// Check if it's a domain with port (example.com:8080)
	if strings.Contains(input, ":") && !strings.Contains(input, "/") {
		parts := strings.SplitN(input, ":", 2)
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}

	// Remove any trailing path if present (e.g., example.com/path)
	if idx := strings.Index(input, "/"); idx > 0 {
		input = input[:idx]
	}

	return input
}

// resolveToIP resolves a hostname to its first IPv4 address.
// Returns empty string on failure.
func resolveToIP(hostname string) string {
	if hostname == "" {
		return ""
	}

	// Use net.LookupIP for DNS resolution
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return ""
	}

	// Return the first IPv4 address found
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}

	// If no IPv4 found, return first IPv6
	for _, ip := range ips {
		return ip.String()
	}

	return ""
}
