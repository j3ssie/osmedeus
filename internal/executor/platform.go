package executor

import (
	"os"
	"runtime"
	"strings"
)

// DetectDocker checks if running inside a Docker container
func DetectDocker() bool {
	// Method 1: Check for /.dockerenv file
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Method 2: Check /proc/1/cgroup for /docker/ (Linux only)
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/1/cgroup")
		if err == nil && strings.Contains(string(data), "/docker/") {
			return true
		}
	}

	return false
}

// DetectKubernetes checks if running inside a Kubernetes pod
func DetectKubernetes() bool {
	// Method 1: Check for Kubernetes service account directory
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount"); err == nil {
		return true
	}

	// Method 2: Check /proc/1/cgroup for kubepods (Linux only)
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/1/cgroup")
		if err == nil {
			content := string(data)
			if strings.Contains(content, "/kubepods/") || strings.Contains(content, "/kubelet/") {
				return true
			}
		}
	}

	return false
}

// DetectCloudProvider detects AWS, GCP, Azure, or returns "local"
func DetectCloudProvider() string {
	// Only works on Linux - check DMI information
	if runtime.GOOS != "linux" {
		return "local"
	}

	// Check sys_vendor
	vendorPaths := []string{
		"/sys/class/dmi/id/sys_vendor",
		"/sys/devices/virtual/dmi/id/bios_vendor",
	}

	for _, path := range vendorPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		vendor := strings.ToLower(strings.TrimSpace(string(data)))

		switch {
		case strings.Contains(vendor, "amazon"):
			return "aws"
		case strings.Contains(vendor, "google"):
			return "gcp"
		case strings.Contains(vendor, "microsoft"):
			return "azure"
		}
	}

	return "local"
}
