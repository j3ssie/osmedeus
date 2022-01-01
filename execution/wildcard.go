package execution

import (
    "fmt"

    //"github.com/OWASP/Amass/v3/requests"
    //amassresolvers "github.com/OWASP/Amass/v3/resolvers"
    "github.com/j3ssie/osmedeus/utils"
)

var defaultResolvers = []string{
    "1.1.1.1:53",   // Cloudflare
    "8.8.8.8:53",   // Google
    "64.6.64.6:53", // Verisign
    "8.8.4.4:53",   // Google Secondary
}

// IsWildCard check if target is wildcard or not
func IsWildCard(domain string) bool {
    //var resolvers []string
    //resolvers = defaultResolvers
    //resolverPool := amassresolvers.SetupResolverPool(resolvers, 1000, false, nil)
    //if resolverPool == nil {
    //	utils.ErrorF("Failed to init DNS pool")
    //	return false
    //}
    //
    //ctx := context.Background()
    //defer ctx.Done()
    //subdomains := genSubs(domain, 5)
    //var totalWildCard int
    //for _, subdomain := range subdomains {
    //	req := &requests.DNSRequest{
    //		Name:   subdomain,
    //		Domain: domain,
    //	}
    //	if !resolverPool.MatchesWildcard(ctx, req) {
    //		utils.DebugF("[wild] %s\n", req.Name)
    //	} else {
    //		totalWildCard += 1
    //		utils.DebugF("[wild] %s\n", req.Name)
    //	}
    //}
    //utils.DebugF("Total number of wildcard: %v/%v\n", totalWildCard, len(subdomains))
    //if totalWildCard == len(subdomains) {
    //	utils.DebugF("Target %v is wildcard\n", domain)
    //	return true
    //}
    return false
}

func genSubs(domain string, size int) []string {
    var subdomains []string
    subdomains = append(subdomains, fmt.Sprintf("notj3ssei.%s", domain))
    subdomains = append(subdomains, fmt.Sprintf("verylong.%s", domain))

    for i := 0; i < (size - 2); i++ {
        subdomains = append(subdomains, fmt.Sprintf("%s.%s", utils.RandomString(5), domain))
    }
    return subdomains
}
