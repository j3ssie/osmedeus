package heuristics

import (
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"
)

// ParseDomain parses a domain and extracts relevant information
func ParseDomain(domain string) (*TargetInfo, error) {
	domain = strings.TrimSpace(domain)
	domain = strings.ToLower(domain)

	info := &TargetInfo{
		Type:     TargetTypeDomain,
		Original: domain,
		Host:     domain,
	}

	// Extract root domain using publicsuffix
	rootDomain, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		// Fallback: use the domain as-is or extract last two parts
		rootDomain = extractRootDomainFallback(domain)
	}
	info.RootDomain = rootDomain
	info.TLD, _ = publicsuffix.PublicSuffix(domain)
	// Extract SLD (second-level domain) by removing TLD from root domain
	info.SLD = extractSLD(rootDomain, info.TLD)

	return info, nil
}

// extractSLD extracts the second-level domain from a root domain
// e.g., "example.com" with TLD "com" -> "example"
// e.g., "example.co.uk" with TLD "co.uk" -> "example"
func extractSLD(rootDomain, tld string) string {
	if rootDomain == "" || tld == "" {
		return ""
	}
	suffix := "." + tld
	if strings.HasSuffix(rootDomain, suffix) {
		return strings.TrimSuffix(rootDomain, suffix)
	}
	return rootDomain
}

// CheckWildcard checks if a domain has wildcard DNS configuration
// It generates 20 random subdomains and 5 common ones, then checks if 90%+ resolve to the same IP
func CheckWildcard(domain string) bool {
	// Generate test subdomains
	randomSubs := generateRandomSubdomains(20)
	commonSubs := []string{"app", "admin", "user", "api", "test"}

	allSubs := append(randomSubs, commonSubs...)

	// Resolve all subdomains concurrently
	var wg sync.WaitGroup
	ipChan := make(chan string, len(allSubs))

	for _, sub := range allSubs {
		wg.Add(1)
		go func(subdomain string) {
			defer wg.Done()
			fqdn := subdomain + "." + domain
			ips, err := net.LookupHost(fqdn)
			if err == nil && len(ips) > 0 {
				ipChan <- ips[0]
			}
		}(sub)
	}

	// Wait for all lookups to complete
	go func() {
		wg.Wait()
		close(ipChan)
	}()

	// Collect results
	var ips []string
	for ip := range ipChan {
		ips = append(ips, ip)
	}

	// If no IPs resolved, not a wildcard
	if len(ips) == 0 {
		return false
	}

	// Count IP occurrences
	ipCount := make(map[string]int)
	for _, ip := range ips {
		ipCount[ip]++
	}

	// Find the most common IP
	maxCount := 0
	for _, count := range ipCount {
		if count > maxCount {
			maxCount = count
		}
	}

	// Check if 90%+ resolve to the same IP
	return float64(maxCount)/float64(len(ips)) >= 0.9
}

// generateRandomSubdomains generates n random subdomain prefixes
func generateRandomSubdomains(n int) []string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	subs := make([]string, n)
	for i := 0; i < n; i++ {
		length := 6 + r.Intn(5) // 6-10 characters
		sub := make([]byte, length)
		for j := range sub {
			sub[j] = charset[r.Intn(len(charset))]
		}
		subs[i] = string(sub)
	}
	return subs
}

// ResolveIP resolves a domain to its IP address
func ResolveIP(domain string) (string, error) {
	ips, err := net.LookupHost(domain)
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "", nil
	}
	return ips[0], nil
}
