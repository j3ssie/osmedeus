package functions

import "github.com/j3ssie/osmedeus/v5/internal/database"

// mergeString returns incoming if non-empty, otherwise existing.
func mergeString(existing, incoming string) string {
	if incoming != "" {
		return incoming
	}
	return existing
}

// mergeInt returns incoming if non-zero, otherwise existing.
func mergeInt(existing, incoming int) int {
	if incoming != 0 {
		return incoming
	}
	return existing
}

// mergeInt64 returns incoming if non-zero, otherwise existing.
func mergeInt64(existing, incoming int64) int64 {
	if incoming != 0 {
		return incoming
	}
	return existing
}

// mergeBool returns true if either value is true (once true, stays true).
func mergeBool(existing, incoming bool) bool {
	return existing || incoming
}

// mergeStringSlice returns incoming if non-nil and non-empty, otherwise existing.
func mergeStringSlice(existing, incoming []string) []string {
	if len(incoming) > 0 {
		return incoming
	}
	return existing
}

// mergeAssetFields merges incoming asset fields into existing.
// Non-zero incoming fields win; zero incoming fields preserve existing values.
// ID, CreatedAt, Workspace, and AssetValue are always taken from existing.
func mergeAssetFields(existing, incoming *database.Asset) {
	// Preserve identity fields from existing
	incoming.ID = existing.ID
	incoming.CreatedAt = existing.CreatedAt
	incoming.Workspace = existing.Workspace
	incoming.AssetValue = existing.AssetValue

	// HTTP data
	incoming.URL = mergeString(existing.URL, incoming.URL)
	incoming.Input = mergeString(existing.Input, incoming.Input)
	incoming.Scheme = mergeString(existing.Scheme, incoming.Scheme)
	incoming.Method = mergeString(existing.Method, incoming.Method)
	incoming.Path = mergeString(existing.Path, incoming.Path)

	// Response data
	incoming.StatusCode = mergeInt(existing.StatusCode, incoming.StatusCode)
	incoming.ContentType = mergeString(existing.ContentType, incoming.ContentType)
	incoming.ContentLength = mergeInt64(existing.ContentLength, incoming.ContentLength)
	incoming.Title = mergeString(existing.Title, incoming.Title)
	incoming.Words = mergeInt(existing.Words, incoming.Words)
	incoming.Lines = mergeInt(existing.Lines, incoming.Lines)

	// Network data
	incoming.HostIP = mergeString(existing.HostIP, incoming.HostIP)
	incoming.DnsRecords = mergeStringSlice(existing.DnsRecords, incoming.DnsRecords)
	incoming.TLS = mergeString(existing.TLS, incoming.TLS)

	// Metadata
	incoming.AssetType = mergeString(existing.AssetType, incoming.AssetType)
	incoming.Technologies = mergeStringSlice(existing.Technologies, incoming.Technologies)
	incoming.ResponseTime = mergeString(existing.ResponseTime, incoming.ResponseTime)
	incoming.Remarks = mergeStringSlice(existing.Remarks, incoming.Remarks)
	incoming.Source = mergeString(existing.Source, incoming.Source)
	incoming.RawJsonData = mergeString(existing.RawJsonData, incoming.RawJsonData)
	incoming.RawResponse = mergeString(existing.RawResponse, incoming.RawResponse)
	incoming.ScreenshotBase64Data = mergeString(existing.ScreenshotBase64Data, incoming.ScreenshotBase64Data)

	// CDN/WAF classification
	incoming.IsCDN = mergeBool(existing.IsCDN, incoming.IsCDN)
	incoming.IsCloud = mergeBool(existing.IsCloud, incoming.IsCloud)
	incoming.IsWAF = mergeBool(existing.IsWAF, incoming.IsWAF)

	// Repository/file assets
	incoming.Language = mergeString(existing.Language, incoming.Language)
	incoming.Size = mergeInt64(existing.Size, incoming.Size)
	incoming.LOC = mergeInt64(existing.LOC, incoming.LOC)
	incoming.ExternalURL = mergeString(existing.ExternalURL, incoming.ExternalURL)
	incoming.BlobContent = mergeString(existing.BlobContent, incoming.BlobContent)
}

// mergeVulnFields merges incoming vulnerability fields into existing.
// Non-zero incoming fields win; zero incoming fields preserve existing values.
// ID, CreatedAt, and Workspace are always taken from existing.
func mergeVulnFields(existing, incoming *database.Vulnerability) {
	// Preserve identity fields from existing
	incoming.ID = existing.ID
	incoming.CreatedAt = existing.CreatedAt
	incoming.Workspace = existing.Workspace

	// Core vulnerability fields
	incoming.VulnInfo = mergeString(existing.VulnInfo, incoming.VulnInfo)
	incoming.VulnTitle = mergeString(existing.VulnTitle, incoming.VulnTitle)
	incoming.VulnDesc = mergeString(existing.VulnDesc, incoming.VulnDesc)
	incoming.VulnPOC = mergeString(existing.VulnPOC, incoming.VulnPOC)
	incoming.Severity = mergeString(existing.Severity, incoming.Severity)
	incoming.Confidence = mergeString(existing.Confidence, incoming.Confidence)
	incoming.AssetType = mergeString(existing.AssetType, incoming.AssetType)
	incoming.AssetValue = mergeString(existing.AssetValue, incoming.AssetValue)
	incoming.Tags = mergeStringSlice(existing.Tags, incoming.Tags)

	// Detail fields
	incoming.DetailHTTPRequest = mergeString(existing.DetailHTTPRequest, incoming.DetailHTTPRequest)
	incoming.DetailHTTPResponse = mergeString(existing.DetailHTTPResponse, incoming.DetailHTTPResponse)
	incoming.RawVulnJSON = mergeString(existing.RawVulnJSON, incoming.RawVulnJSON)
}
