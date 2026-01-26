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
	"strconv"
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

// getParentURL strips the last path component from a URL and returns the parent directory.
// Usage: get_parent_url(url) -> string
//   - url: the URL to process
//
// Examples:
//
//	get_parent_url("https://example.com/j3ssie/sample.php?query=123") -> "https://example.com/j3ssie/"
//	get_parent_url("https://example.com/a/b/c/") -> "https://example.com/a/b/"
//	get_parent_url("https://example.com/file.txt") -> "https://example.com/"
//	get_parent_url("https://example.com/") -> "https://example.com/"
func (vf *vmFunc) getParentURL(call goja.FunctionCall) goja.Value {
	input := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("get_parent_url"),
		zap.String("input", input))

	if input == "undefined" || input == "" {
		logger.Get().Warn("get_parent_url: empty input provided")
		return vf.vm.ToValue("")
	}

	result := getParentURLImpl(input)

	logger.Get().Debug(terminal.HiGreen("get_parent_url")+" result",
		zap.String("input", input),
		zap.String("result", result))

	return vf.vm.ToValue(result)
}

// getParentURLImpl extracts the parent directory URL.
// This is the implementation that can be tested independently.
func getParentURLImpl(input string) string {
	// Parse the URL
	u, err := url.Parse(input)
	if err != nil {
		return input
	}

	// Get the path and strip query/fragment
	path := u.Path

	// Handle empty path
	if path == "" || path == "/" {
		// Ensure trailing slash
		u.Path = "/"
		u.RawQuery = ""
		u.Fragment = ""
		return u.String()
	}

	// Remove trailing slash for uniform processing
	path = strings.TrimSuffix(path, "/")

	// Find the last slash
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		// No slash found, return root
		u.Path = "/"
	} else {
		// Keep everything up to and including the last slash
		u.Path = path[:lastSlash+1]
	}

	// Clear query and fragment
	u.RawQuery = ""
	u.Fragment = ""

	return u.String()
}

// parseURL formats a URL using format directives similar to unfurl.
// Usage: parse_url(url, format) -> string
//   - url: the URL to parse
//   - format: format string with directives
//
// Format directives:
//
//	%% - Literal percent character
//	%s - Request scheme (http, https)
//	%u - User info (username:password)
//	%d - Full domain (sub.example.com)
//	%S - Subdomain (sub)
//	%r - Root domain (example)
//	%t - TLD (com)
//	%P - Port (8080)
//	%p - Path (/users/list)
//	%e - File extension (jpg)
//	%q - Raw query string (a=1&b=2)
//	%f - Fragment (section)
//	%@ - @ if user info exists, empty otherwise
//	%: - : if port exists, empty otherwise
//	%? - ? if query exists, empty otherwise
//	%# - # if fragment exists, empty otherwise
//	%a - Authority (user:pass@domain:port)
func (vf *vmFunc) parseURL(call goja.FunctionCall) goja.Value {
	urlStr := call.Argument(0).String()
	format := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("parse_url"),
		zap.String("url", urlStr), zap.String("format", format))

	if urlStr == "undefined" || urlStr == "" {
		logger.Get().Warn("parse_url: empty URL provided")
		return vf.vm.ToValue("")
	}

	if format == "undefined" || format == "" {
		logger.Get().Warn("parse_url: empty format provided")
		return vf.vm.ToValue("")
	}

	result := parseURLImpl(urlStr, format)

	logger.Get().Debug(terminal.HiGreen("parse_url")+" result",
		zap.String("url", urlStr),
		zap.String("format", format),
		zap.String("result", result))

	return vf.vm.ToValue(result)
}

// parseURLImpl formats a URL using format directives.
// This is the implementation that can be tested independently.
func parseURLImpl(urlStr, format string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	// Extract domain parts
	domain := u.Hostname()
	subdomain, root, tld := extractDomainParts(domain)

	// Extract file extension from path
	ext := ""
	if u.Path != "" {
		base := filepath.Base(u.Path)
		if dotIdx := strings.LastIndex(base, "."); dotIdx != -1 && dotIdx < len(base)-1 {
			ext = base[dotIdx+1:]
		}
	}

	// Extract user info
	userInfo := ""
	if u.User != nil {
		userInfo = u.User.String()
	}

	// Extract port
	port := u.Port()

	// Build result by parsing format string
	var result strings.Builder
	i := 0
	for i < len(format) {
		if format[i] == '%' && i+1 < len(format) {
			switch format[i+1] {
			case '%':
				result.WriteByte('%')
			case 's':
				result.WriteString(u.Scheme)
			case 'u':
				result.WriteString(userInfo)
			case 'd':
				result.WriteString(domain)
			case 'S':
				result.WriteString(subdomain)
			case 'r':
				result.WriteString(root)
			case 't':
				result.WriteString(tld)
			case 'P':
				result.WriteString(port)
			case 'p':
				result.WriteString(u.Path)
			case 'e':
				result.WriteString(ext)
			case 'q':
				result.WriteString(u.RawQuery)
			case 'f':
				result.WriteString(u.Fragment)
			case '@':
				if userInfo != "" {
					result.WriteByte('@')
				}
			case ':':
				if port != "" {
					result.WriteByte(':')
				}
			case '?':
				if u.RawQuery != "" {
					result.WriteByte('?')
				}
			case '#':
				if u.Fragment != "" {
					result.WriteByte('#')
				}
			case 'a':
				// Authority: user:pass@domain:port
				if userInfo != "" {
					result.WriteString(userInfo)
					result.WriteByte('@')
				}
				result.WriteString(domain)
				if port != "" {
					result.WriteByte(':')
					result.WriteString(port)
				}
			default:
				// Unknown directive, output as-is
				result.WriteByte('%')
				result.WriteByte(format[i+1])
			}
			i += 2
		} else {
			result.WriteByte(format[i])
			i++
		}
	}

	return result.String()
}

// knownMultiPartTLDs contains common multi-part TLDs
var knownMultiPartTLDs = map[string]bool{
	"co.uk":  true,
	"com.au": true,
	"co.jp":  true,
	"co.nz":  true,
	"co.za":  true,
	"com.br": true,
	"com.cn": true,
	"com.mx": true,
	"com.tw": true,
	"com.hk": true,
	"com.sg": true,
	"org.uk": true,
	"net.au": true,
	"gov.uk": true,
	"ac.uk":  true,
	"edu.au": true,
	"co.in":  true,
	"com.ar": true,
	"com.co": true,
	"co.kr":  true,
	"or.jp":  true,
	"ne.jp":  true,
	"ac.jp":  true,
	"go.jp":  true,
}

// extractDomainParts splits a domain into subdomain, root, and TLD.
// Examples:
//
//	sub.example.com -> (sub, example, com)
//	example.com -> ("", example, com)
//	api.sub.example.co.uk -> (api.sub, example, co.uk)
func extractDomainParts(domain string) (subdomain, root, tld string) {
	if domain == "" {
		return "", "", ""
	}

	parts := strings.Split(domain, ".")
	if len(parts) == 1 {
		// No dots, treat as root
		return "", parts[0], ""
	}

	// Check for multi-part TLDs
	if len(parts) >= 2 {
		potentialMultiTLD := parts[len(parts)-2] + "." + parts[len(parts)-1]
		if knownMultiPartTLDs[potentialMultiTLD] {
			tld = potentialMultiTLD
			if len(parts) == 2 {
				// e.g., "co.uk" - no root domain
				return "", "", tld
			}
			root = parts[len(parts)-3]
			if len(parts) > 3 {
				subdomain = strings.Join(parts[:len(parts)-3], ".")
			}
			return subdomain, root, tld
		}
	}

	// Standard single-part TLD
	tld = parts[len(parts)-1]
	if len(parts) == 2 {
		// e.g., "example.com"
		return "", parts[0], tld
	}

	// e.g., "sub.example.com" or "api.sub.example.com"
	root = parts[len(parts)-2]
	subdomain = strings.Join(parts[:len(parts)-2], ".")
	return subdomain, root, tld
}

// queryReplace replaces all query parameter values in a URL.
// Usage: query_replace(url, value, mode?) -> string
//   - url: the URL to modify
//   - value: the replacement value
//   - mode: "replace" (default) or "append"
//
// Examples:
//
//	query_replace("https://example.com?a=1&b=2", "new") -> "https://example.com?a=new&b=new"
//	query_replace("https://example.com?a=1&b=2", "FUZZ", "append") -> "https://example.com?a=1FUZZ&b=2FUZZ"
func (vf *vmFunc) queryReplace(call goja.FunctionCall) goja.Value {
	urlStr := call.Argument(0).String()
	value := call.Argument(1).String()
	mode := call.Argument(2).String()

	logger.Get().Debug("Calling "+terminal.HiGreen("query_replace"),
		zap.String("url", urlStr), zap.String("value", value), zap.String("mode", mode))

	if urlStr == "undefined" || urlStr == "" {
		logger.Get().Warn("query_replace: empty URL provided")
		return vf.vm.ToValue("")
	}
	if value == "undefined" {
		value = ""
	}
	if mode == "undefined" || mode == "" {
		mode = "replace"
	}

	result := queryReplaceImpl(urlStr, value, mode)

	logger.Get().Debug(terminal.HiGreen("query_replace")+" result",
		zap.String("url", urlStr), zap.String("value", value),
		zap.String("mode", mode), zap.String("result", result))

	return vf.vm.ToValue(result)
}

// queryReplaceImpl is the implementation for testing.
func queryReplaceImpl(urlStr, value, mode string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	q := u.Query()
	if len(q) == 0 {
		return urlStr
	}

	newQuery := url.Values{}
	for key, values := range q {
		for _, v := range values {
			switch mode {
			case "append":
				newQuery.Add(key, v+value)
			default: // "replace"
				newQuery.Add(key, value)
			}
		}
	}
	u.RawQuery = newQuery.Encode()
	return u.String()
}

// pathReplace replaces a path segment at a specific position.
// Usage: path_replace(url, value, position?) -> string
//   - url: the URL to modify
//   - value: the replacement value
//   - position: 1-indexed position (default 1), 0 or negative replaces all segments
//
// Examples:
//
//	path_replace("https://example.com/a/b/c", "new") -> "https://example.com/new/b/c"
//	path_replace("https://example.com/a/b/c", "new", 2) -> "https://example.com/a/new/c"
//	path_replace("https://example.com/a/b/c", "new", 0) -> "https://example.com/new/new/new"
func (vf *vmFunc) pathReplace(call goja.FunctionCall) goja.Value {
	urlStr := call.Argument(0).String()
	value := call.Argument(1).String()
	posArg := call.Argument(2)

	logger.Get().Debug("Calling "+terminal.HiGreen("path_replace"),
		zap.String("url", urlStr), zap.String("value", value))

	if urlStr == "undefined" || urlStr == "" {
		logger.Get().Warn("path_replace: empty URL provided")
		return vf.vm.ToValue("")
	}
	if value == "undefined" {
		value = ""
	}

	// Parse position (default 1)
	position := 1
	if !goja.IsUndefined(posArg) && !goja.IsNull(posArg) {
		if p, ok := posArg.Export().(int64); ok {
			position = int(p)
		} else if p, ok := posArg.Export().(float64); ok {
			position = int(p)
		} else if s := posArg.String(); s != "undefined" && s != "" {
			if p, err := strconv.Atoi(s); err == nil {
				position = p
			}
		}
	}

	result := pathReplaceImpl(urlStr, value, position)

	logger.Get().Debug(terminal.HiGreen("path_replace")+" result",
		zap.String("url", urlStr), zap.String("value", value),
		zap.Int("position", position), zap.String("result", result))

	return vf.vm.ToValue(result)
}

// pathReplaceImpl is the implementation for testing.
func pathReplaceImpl(urlStr, value string, position int) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	// Split path into segments (skip empty segments from leading slash)
	path := strings.TrimPrefix(u.Path, "/")
	if path == "" {
		return urlStr
	}

	segments := strings.Split(path, "/")
	if len(segments) == 0 {
		return urlStr
	}

	// Replace based on position
	if position <= 0 {
		// Replace all segments
		for i := range segments {
			segments[i] = value
		}
	} else if position <= len(segments) {
		// Replace specific segment (1-indexed)
		segments[position-1] = value
	}
	// If position > len(segments), return unchanged

	u.Path = "/" + strings.Join(segments, "/")
	return u.String()
}
