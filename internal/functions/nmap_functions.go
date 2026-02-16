package functions

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/json"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
)

// NmapRun represents the root element of nmap XML output
type NmapRun struct {
	XMLName xml.Name   `xml:"nmaprun"`
	Hosts   []NmapHost `xml:"host"`
}

// NmapHost represents a scanned host
type NmapHost struct {
	Address NmapAddress `xml:"address"`
	Ports   []NmapPort  `xml:"ports>port"`
}

// NmapAddress represents host address
type NmapAddress struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}

// NmapPort represents a port entry
type NmapPort struct {
	Protocol string      `xml:"protocol,attr"`
	PortID   string      `xml:"portid,attr"`
	State    NmapState   `xml:"state"`
	Service  NmapService `xml:"service"`
}

// NmapState represents port state
type NmapState struct {
	State  string `xml:"state,attr"`
	Reason string `xml:"reason,attr"`
}

// NmapService represents service information
type NmapService struct {
	Name      string `xml:"name,attr"`
	Product   string `xml:"product,attr"`
	Version   string `xml:"version,attr"`
	ExtraInfo string `xml:"extrainfo,attr"`
	Method    string `xml:"method,attr"`
}

// PortDetail represents detailed port information
type PortDetail struct {
	Protocol    string `json:"protocol,omitempty"`
	State       string `json:"state,omitempty"`
	Service     string `json:"service,omitempty"`
	Product     string `json:"product,omitempty"`
	Version     string `json:"version,omitempty"`
	ServiceInfo string `json:"service_info,omitempty"`
}

// AssetOutput represents the JSONL output format
type AssetOutput struct {
	AssetValue string                `json:"asset_value"`
	HostIP     string                `json:"host_ip"`
	AssetType  string                `json:"asset_type"`
	OpenPorts  []string              `json:"open_ports"`
	Ports      map[string]PortDetail `json:"ports"`
}

// parseNmapXML parses nmap XML format
func parseNmapXML(filePath string) ([]AssetOutput, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var nmapRun NmapRun
	if err := xml.Unmarshal(data, &nmapRun); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	var results []AssetOutput
	for _, host := range nmapRun.Hosts {
		if host.Address.Addr == "" {
			continue
		}

		asset := AssetOutput{
			AssetValue: host.Address.Addr,
			HostIP:     host.Address.Addr,
			AssetType:  "ip",
			OpenPorts:  []string{},
			Ports:      make(map[string]PortDetail),
		}

		for _, port := range host.Ports {
			portKey := port.PortID
			portProto := fmt.Sprintf("%s/%s", port.PortID, port.Protocol)

			// Add to open_ports array (only open/filtered ports)
			if port.State.State == "open" || port.State.State == "filtered" {
				asset.OpenPorts = append(asset.OpenPorts, portProto)
			}

			// Build service info
			serviceInfo := port.Service.ExtraInfo
			if serviceInfo == "" && port.Service.Method != "" {
				serviceInfo = fmt.Sprintf("method=%s", port.Service.Method)
			}

			// Add detailed port information
			detail := PortDetail{
				Protocol:    port.Protocol,
				State:       port.State.State,
				Service:     port.Service.Name,
				Product:     port.Service.Product,
				Version:     port.Service.Version,
				ServiceInfo: serviceInfo,
			}

			asset.Ports[portKey] = detail
		}

		if len(asset.Ports) > 0 {
			results = append(results, asset)
		}
	}

	return results, nil
}

// parseNmapGrepable parses nmap grepable format
func parseNmapGrepable(filePath string) ([]AssetOutput, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	var results []AssetOutput
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "Host:") || !strings.Contains(line, "Ports:") {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		// Extract host IP
		hostPart := strings.TrimPrefix(parts[0], "Host: ")
		hostIP := strings.Fields(hostPart)[0]

		asset := AssetOutput{
			AssetValue: hostIP,
			HostIP:     hostIP,
			AssetType:  "ip",
			OpenPorts:  []string{},
			Ports:      make(map[string]PortDetail),
		}

		// Parse ports
		for _, part := range parts {
			if !strings.HasPrefix(part, "Ports:") {
				continue
			}

			portsPart := strings.TrimPrefix(part, "Ports: ")
			portEntries := strings.Split(portsPart, ", ")

			for _, entry := range portEntries {
				// Format: portid/state/protocol//service//version
				fields := strings.Split(entry, "/")
				if len(fields) < 3 {
					continue
				}

				portID := fields[0]
				state := fields[1]
				protocol := fields[2]
				service := ""
				version := ""

				if len(fields) >= 5 {
					service = fields[4]
				}
				if len(fields) >= 7 {
					version = fields[6]
				}

				portProto := fmt.Sprintf("%s/%s", portID, protocol)

				// Add to open_ports
				if state == "open" || state == "filtered" {
					asset.OpenPorts = append(asset.OpenPorts, portProto)
				}

				// Add detail
				detail := PortDetail{
					Protocol: protocol,
					State:    state,
					Service:  service,
					Version:  version,
				}

				asset.Ports[portID] = detail
			}
		}

		if len(asset.Ports) > 0 {
			results = append(results, asset)
		}
	}

	return results, scanner.Err()
}

// nmapToJSONL converts nmap output to JSONL format
func (vf *vmFunc) nmapToJSONL(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		logger.Get().Warn("nmap_to_jsonl requires 2 arguments: input_path, output_path")
		return vf.vm.ToValue(false)
	}

	inputPath := call.Argument(0).String()
	outputPath := call.Argument(1).String()

	if inputPath == "" || inputPath == "undefined" {
		logger.Get().Warn("input_path is required")
		return vf.vm.ToValue(false)
	}

	if outputPath == "" || outputPath == "undefined" {
		logger.Get().Warn("output_path is required")
		return vf.vm.ToValue(false)
	}

	// Auto-detect format based on extension
	ext := strings.ToLower(filepath.Ext(inputPath))
	var results []AssetOutput
	var err error

	switch ext {
	case ".gnmap":
		results, err = parseNmapGrepable(inputPath)
	case ".nmap", ".txt":
		// Text format not supported yet, default to XML
		logger.Get().Warn("Text format (.nmap) not fully supported, attempting XML parse")
		results, err = parseNmapXML(inputPath)
	case ".xml", "":
		// Default to XML
		results, err = parseNmapXML(inputPath)
	default:
		// Try XML as fallback
		results, err = parseNmapXML(inputPath)
	}

	if err != nil {
		logger.Get().Warn("Failed to parse nmap file", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Write JSONL output
	outFile, err := os.Create(outputPath)
	if err != nil {
		logger.Get().Warn("Failed to create output file", zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = outFile.Close() }()

	writer := bufio.NewWriter(outFile)
	for _, asset := range results {
		data, err := json.Marshal(asset)
		if err != nil {
			logger.Get().Warn("Failed to marshal asset", zap.Error(err))
			continue
		}

		if _, err := writer.Write(data); err != nil {
			logger.Get().Warn("Failed to write line", zap.Error(err))
			continue
		}
		if _, err := writer.WriteString("\n"); err != nil {
			logger.Get().Warn("Failed to write newline", zap.Error(err))
			continue
		}
	}

	if err := writer.Flush(); err != nil {
		logger.Get().Warn("Failed to flush writer", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	logger.Get().Info("Converted hosts from nmap to JSONL", zap.Int("count", len(results)), zap.String("output", outputPath))
	return vf.vm.ToValue(true)
}

// sanitizeTargetForPath converts target to filesystem-safe string
func sanitizeTargetForPath(target string) string {
	s := strings.ReplaceAll(target, "/", "-")
	s = strings.ReplaceAll(s, ":", "-")
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.TrimSuffix(s, "-txt")
	s = strings.TrimSuffix(s, "-xml")
	return s
}

// runNmap executes nmap scan and converts output to JSONL
func (vf *vmFunc) runNmap(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		logger.Get().Warn("run_nmap requires at least 1 argument: target")
		return vf.vm.ToValue(false)
	}

	target := call.Argument(0).String()
	if target == "" || target == "undefined" {
		logger.Get().Warn("target is required")
		return vf.vm.ToValue(false)
	}

	// Extract optional flags (default: -sV -T4)
	flags := "-sV -T4"
	if len(call.Arguments) >= 2 && call.Argument(1).String() != "undefined" && call.Argument(1).String() != "" {
		flags = call.Argument(1).String()
	}

	// Extract optional output path
	var outputPath string
	if len(call.Arguments) >= 3 && call.Argument(2).String() != "undefined" && call.Argument(2).String() != "" {
		outputPath = call.Argument(2).String()
	} else {
		// Auto-generate output path
		baseDir := vf.getContext().workspacePath
		if baseDir == "" {
			baseDir = "."
		}
		sanitizedTarget := sanitizeTargetForPath(target)
		outputPath = filepath.Join(baseDir, fmt.Sprintf("nmap-%s.jsonl", sanitizedTarget))
	}

	// Check if nmap command exists
	nmapPath, err := exec.LookPath("nmap")
	if err != nil {
		logger.Get().Warn("nmap command not found in PATH", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Create temp directory for XML output
	tmpDir, err := os.MkdirTemp("", "osm-nmap-*")
	if err != nil {
		logger.Get().Warn("Failed to create temp directory", zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	xmlPath := filepath.Join(tmpDir, "scan.xml")

	// Build nmap command with XML output
	flagsList := strings.Fields(flags)
	args := append(flagsList, "-oX", xmlPath, target)

	logger.Get().Info("Executing nmap scan", zap.String("target", target), zap.Strings("args", args))

	// Execute nmap
	cmd := exec.Command(nmapPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if XML file exists (partial results may still be available)
		if _, statErr := os.Stat(xmlPath); statErr == nil {
			logger.Get().Warn("nmap exited with error but XML output exists, using partial results",
				zap.Error(err),
				zap.String("output", string(output)))
		} else {
			logger.Get().Warn("nmap execution failed",
				zap.Error(err),
				zap.String("output", string(output)))
			return vf.vm.ToValue(false)
		}
	}

	// Parse XML
	results, err := parseNmapXML(xmlPath)
	if err != nil {
		logger.Get().Warn("Failed to parse nmap XML output", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	if len(results) == 0 {
		logger.Get().Warn("No hosts found in nmap output")
		return vf.vm.ToValue(outputPath) // Return path even if empty for consistency
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Get().Warn("Failed to create output directory", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Write JSONL output
	outFile, err := os.Create(outputPath)
	if err != nil {
		logger.Get().Warn("Failed to create output file", zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = outFile.Close() }()

	writer := bufio.NewWriter(outFile)
	for _, asset := range results {
		data, err := json.Marshal(asset)
		if err != nil {
			logger.Get().Warn("Failed to marshal asset", zap.Error(err))
			continue
		}

		if _, err := writer.Write(data); err != nil {
			logger.Get().Warn("Failed to write line", zap.Error(err))
			continue
		}
		if _, err := writer.WriteString("\n"); err != nil {
			logger.Get().Warn("Failed to write newline", zap.Error(err))
			continue
		}
	}

	if err := writer.Flush(); err != nil {
		logger.Get().Warn("Failed to flush writer", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	logger.Get().Info("Completed nmap scan and converted to JSONL",
		zap.Int("hosts", len(results)),
		zap.String("output", outputPath))

	return vf.vm.ToValue(outputPath)
}
