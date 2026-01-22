package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// configCmd - parent command for config management
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage osmedeus configuration",
	Long:  UsageConfig(),
}

// configCleanCmd - reset config to defaults
var configCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Reset configuration to defaults",
	Long:  UsageConfigClean(),
	RunE:  runConfigClean,
}

// configSetCmd - set a config value
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long:  UsageConfigSet(),
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

var (
	configViewRedact      bool
	configViewForce       bool
	configListShowSecrets bool
)

var configViewCmd = &cobra.Command{
	Use:   "view <key>",
	Short: "View a configuration value",
	Long:  UsageConfigView(),
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigView,
}

var configListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List configuration values",
	Long:    UsageConfigList(),
	Args:    cobra.NoArgs,
	RunE:    runConfigList,
}

func init() {
	configCmd.AddCommand(configCleanCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configListCmd)

	configViewCmd.Flags().BoolVar(&configViewRedact, "redact", false, "redact sensitive values")
	configViewCmd.Flags().BoolVar(&configViewForce, "force", false, "required for wildcard pattern searches")
	configListCmd.Flags().BoolVar(&configListShowSecrets, "show-secrets", false, "show sensitive values")
}

// runConfigClean resets the configuration to defaults
func runConfigClean(cmd *cobra.Command, args []string) error {
	printer := terminal.NewPrinter()
	cfg := config.Get()

	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	settingsPath := filepath.Join(cfg.BaseFolder, "osm-settings.yaml")

	// Backup existing config if present
	if _, err := os.Stat(settingsPath); err == nil {
		backupPath := settingsPath + ".backup"
		printer.Info("Backing up existing config to %s", backupPath)
		if err := os.Rename(settingsPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup config: %w", err)
		}
	}

	// Write fresh default config
	if err := config.EnsureConfigExists(cfg.BaseFolder); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}

	printer.Success("Configuration reset to defaults at %s", settingsPath)
	return nil
}

// runConfigSet sets a configuration value using dot notation
func runConfigSet(cmd *cobra.Command, args []string) error {
	printer := terminal.NewPrinter()
	key := args[0]
	value := args[1]

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	settingsPath := filepath.Join(cfg.BaseFolder, "osm-settings.yaml")

	// Load fresh from file to avoid runtime-only fields
	freshCfg, err := config.LoadFromFile(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	currentAuthUsername, _ := primaryServerAuthUser(freshCfg)
	if currentAuthUsername == "" {
		currentAuthUsername = "osmedeus"
	}

	// Set the value using dot notation
	if err := setConfigValue(freshCfg, key, value); err != nil {
		return fmt.Errorf("failed to set %s: %w", key, err)
	}

	effectiveKey := key
	var writeErr error
	switch key {
	case "server.password":
		effectiveKey = fmt.Sprintf("server.simple_user_map_key.%s", currentAuthUsername)
		writeErr = updateSettingsYAMLScalarValue(settingsPath, effectiveKey, value)
	case "server.username":
		writeErr = updateSettingsYAMLMappingKey(settingsPath, []string{"server", "simple_user_map_key"}, currentAuthUsername, value)
	default:
		writeErr = updateSettingsYAMLScalarValue(settingsPath, effectiveKey, value)
	}

	if writeErr != nil {
		if err := writeSettingsYAMLFromConfig(settingsPath, freshCfg); err != nil {
			return fmt.Errorf("failed to write config: %w", writeErr)
		}
	}

	printer.Success("Set %s = %s", key, redactValueForDisplay(key, value, false))
	return nil
}

func runConfigView(cmd *cobra.Command, args []string) error {
	key := args[0]

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	settingsPath := filepath.Join(cfg.BaseFolder, "osm-settings.yaml")
	fileCfg, err := config.LoadFromFile(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check for wildcard pattern
	if strings.Contains(key, "*") {
		if !configViewForce {
			return fmt.Errorf("wildcard patterns require --force flag\n\nExample: osmedeus config view '%s' --force", key)
		}
		return runConfigViewPattern(key, settingsPath, fileCfg)
	}

	if key == "server.username" {
		username, _ := primaryServerAuthUser(fileCfg)
		fmt.Println(username)
		return nil
	}
	if key == "server.password" {
		_, password := primaryServerAuthUser(fileCfg)
		fmt.Println(redactValueForDisplay(key, password, !configViewRedact))
		return nil
	}

	content, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}

	var doc yaml.Node
	dec := yaml.NewDecoder(bytes.NewReader(content))
	dec.KnownFields(false)
	if err := dec.Decode(&doc); err != nil {
		return err
	}
	if len(doc.Content) == 0 {
		return fmt.Errorf("empty yaml document")
	}

	targetNode, err := findYAMLNodeByPath(doc.Content[0], strings.Split(key, "."))
	if err != nil {
		return err
	}

	if targetNode.Kind == yaml.ScalarNode {
		fmt.Println(redactValueForDisplay(key, targetNode.Value, !configViewRedact))
		return nil
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	_ = enc.Encode(targetNode)
	_ = enc.Close()

	output := buf.String()
	if configViewRedact {
		output = redactSensitiveFieldsYAML(output)
	}
	fmt.Print(output)
	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	settingsPath := filepath.Join(cfg.BaseFolder, "osm-settings.yaml")
	fileCfg, err := config.LoadFromFile(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	content, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}

	var doc yaml.Node
	dec := yaml.NewDecoder(bytes.NewReader(content))
	dec.KnownFields(false)
	if err := dec.Decode(&doc); err != nil {
		return err
	}
	if len(doc.Content) == 0 {
		return fmt.Errorf("empty yaml document")
	}

	out := map[string]string{}
	flattenYAMLScalars(doc.Content[0], "", out)

	username, password := primaryServerAuthUser(fileCfg)
	if username != "" {
		out["server.username"] = redactValueForDisplay("server.username", username, configListShowSecrets)
	}
	if password != "" {
		out["server.password"] = redactValueForDisplay("server.password", password, configListShowSecrets)
	}

	keys := make([]string, 0, len(out))
	for k := range out {
		keys = append(keys, k)
	}
	sortStrings(keys)

	for _, k := range keys {
		v := out[k]
		if !configListShowSecrets {
			v = redactValueForDisplay(k, v, false)
		}
		fmt.Printf("%s = %s\n", k, v)
	}
	return nil
}

func updateSettingsYAMLScalarValue(settingsPath, dotKey, newValue string) error {
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}

	updated, err := setYAMLScalarValuePreserveComments(content, strings.Split(dotKey, "."), newValue)
	if err != nil {
		return err
	}

	if _, err := config.ParseConfigStrict(updated); err != nil {
		return err
	}

	return os.WriteFile(settingsPath, updated, 0644)
}

func updateSettingsYAMLMappingKey(settingsPath string, mappingPath []string, oldKey, newKey string) error {
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}

	updated, err := renameYAMLMappingKeyPreserveComments(content, mappingPath, oldKey, newKey)
	if err != nil {
		return err
	}

	if _, err := config.ParseConfigStrict(updated); err != nil {
		return err
	}

	return os.WriteFile(settingsPath, updated, 0644)
}

func writeSettingsYAMLFromConfig(settingsPath string, cfg *config.Config) error {
	data, err := cfg.ToYAML()
	if err != nil {
		return err
	}
	if _, err := config.ParseConfigStrict(data); err != nil {
		return err
	}
	return os.WriteFile(settingsPath, data, 0644)
}

func setYAMLScalarValuePreserveComments(content []byte, path []string, newValue string) ([]byte, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("empty key")
	}

	var doc yaml.Node
	dec := yaml.NewDecoder(bytes.NewReader(content))
	dec.KnownFields(false)
	if err := dec.Decode(&doc); err != nil {
		return nil, err
	}
	if len(doc.Content) == 0 {
		return nil, fmt.Errorf("empty yaml document")
	}

	targetNode, err := findYAMLNodeByPath(doc.Content[0], path)
	if err != nil {
		return nil, err
	}
	if targetNode.Kind != yaml.ScalarNode {
		return nil, fmt.Errorf("target %s is not a scalar", strings.Join(path, "."))
	}
	if targetNode.Line <= 0 || targetNode.Column <= 0 {
		return nil, fmt.Errorf("unable to locate scalar position for %s", strings.Join(path, "."))
	}

	lines := bytes.Split(content, []byte("\n"))
	lineIdx := targetNode.Line - 1
	if lineIdx < 0 || lineIdx >= len(lines) {
		return nil, fmt.Errorf("invalid yaml line for %s", strings.Join(path, "."))
	}

	line := lines[lineIdx]
	start := targetNode.Column - 1
	if start < 0 || start >= len(line) {
		return nil, fmt.Errorf("invalid yaml column for %s", strings.Join(path, "."))
	}

	originalQuote := byte(0)
	if start < len(line) {
		if line[start] == '\'' || line[start] == '"' {
			originalQuote = line[start]
		}
	}

	end := scalarTokenEnd(line, start)
	if end < start {
		return nil, fmt.Errorf("unable to find scalar token end for %s", strings.Join(path, "."))
	}

	replacement := formatScalarReplacement(originalQuote, newValue)

	lines[lineIdx] = append(append(line[:start], replacement...), line[end:]...)

	endsWithNewline := len(content) > 0 && content[len(content)-1] == '\n'
	updated := bytes.Join(lines, []byte("\n"))
	if endsWithNewline && (len(updated) == 0 || updated[len(updated)-1] != '\n') {
		updated = append(updated, '\n')
	}
	if !endsWithNewline && len(updated) > 0 && updated[len(updated)-1] == '\n' {
		updated = updated[:len(updated)-1]
	}

	return updated, nil
}

func renameYAMLMappingKeyPreserveComments(content []byte, mappingPath []string, oldKey, newKey string) ([]byte, error) {
	var doc yaml.Node
	dec := yaml.NewDecoder(bytes.NewReader(content))
	dec.KnownFields(false)
	if err := dec.Decode(&doc); err != nil {
		return nil, err
	}
	if len(doc.Content) == 0 {
		return nil, fmt.Errorf("empty yaml document")
	}

	mappingNode, err := findYAMLNodeByPath(doc.Content[0], mappingPath)
	if err != nil {
		return nil, err
	}
	if mappingNode.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("target %s is not a mapping", strings.Join(mappingPath, "."))
	}

	var keyNode *yaml.Node
	for i := 0; i+1 < len(mappingNode.Content); i += 2 {
		k := mappingNode.Content[i]
		if k.Kind == yaml.ScalarNode && k.Value == oldKey {
			keyNode = k
			break
		}
	}
	if keyNode == nil {
		return nil, fmt.Errorf("key not found: %s", strings.Join(append(mappingPath, oldKey), "."))
	}
	if keyNode.Line <= 0 || keyNode.Column <= 0 {
		return nil, fmt.Errorf("unable to locate scalar position for %s", strings.Join(append(mappingPath, oldKey), "."))
	}

	lines := bytes.Split(content, []byte("\n"))
	lineIdx := keyNode.Line - 1
	if lineIdx < 0 || lineIdx >= len(lines) {
		return nil, fmt.Errorf("invalid yaml line for %s", strings.Join(append(mappingPath, oldKey), "."))
	}

	line := lines[lineIdx]
	start := keyNode.Column - 1
	if start < 0 || start >= len(line) {
		return nil, fmt.Errorf("invalid yaml column for %s", strings.Join(append(mappingPath, oldKey), "."))
	}

	originalQuote := byte(0)
	if start < len(line) {
		if line[start] == '\'' || line[start] == '"' {
			originalQuote = line[start]
		}
	}
	end := scalarTokenEnd(line, start)
	if end < start {
		return nil, fmt.Errorf("unable to find scalar token end for %s", strings.Join(append(mappingPath, oldKey), "."))
	}

	replacement := formatScalarReplacement(originalQuote, newKey)
	lines[lineIdx] = append(append(line[:start], replacement...), line[end:]...)

	endsWithNewline := len(content) > 0 && content[len(content)-1] == '\n'
	updated := bytes.Join(lines, []byte("\n"))
	if endsWithNewline && (len(updated) == 0 || updated[len(updated)-1] != '\n') {
		updated = append(updated, '\n')
	}
	if !endsWithNewline && len(updated) > 0 && updated[len(updated)-1] == '\n' {
		updated = updated[:len(updated)-1]
	}

	return updated, nil
}

func findYAMLNodeByPath(root *yaml.Node, path []string) (*yaml.Node, error) {
	node := root
	for _, segment := range path {
		if node.Kind == yaml.DocumentNode {
			if len(node.Content) == 0 {
				return nil, fmt.Errorf("empty yaml document")
			}
			node = node.Content[0]
		}
		if node.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("%s is not a mapping", segment)
		}

		found := false
		for i := 0; i+1 < len(node.Content); i += 2 {
			k := node.Content[i]
			v := node.Content[i+1]
			if k.Kind == yaml.ScalarNode && k.Value == segment {
				node = v
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("key not found: %s", strings.Join(path, "."))
		}
	}

	return node, nil
}

func scalarTokenEnd(line []byte, start int) int {
	if start < 0 || start >= len(line) {
		return start
	}

	if line[start] == '\'' {
		for i := start + 1; i < len(line); i++ {
			if line[i] != '\'' {
				continue
			}
			if i+1 < len(line) && line[i+1] == '\'' {
				i++
				continue
			}
			return i + 1
		}
		return len(line)
	}

	if line[start] == '"' {
		for i := start + 1; i < len(line); i++ {
			if line[i] == '\\' {
				if i+1 < len(line) {
					i++
				}
				continue
			}
			if line[i] == '"' {
				return i + 1
			}
		}
		return len(line)
	}

	for i := start; i < len(line); i++ {
		b := line[i]
		if b == ' ' || b == '\t' {
			return i
		}
		if b == '#' {
			if i == start {
				return i
			}
			prev := line[i-1]
			if prev == ' ' || prev == '\t' {
				return i
			}
		}
	}

	return len(line)
}

func formatScalarReplacement(originalQuote byte, newValue string) []byte {
	if originalQuote == '\'' {
		return []byte("'" + strings.ReplaceAll(newValue, "'", "''") + "'")
	}
	if originalQuote == '"' {
		escaped := strings.ReplaceAll(newValue, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		escaped = strings.ReplaceAll(escaped, "\n", "\\n")
		escaped = strings.ReplaceAll(escaped, "\r", "\\r")
		escaped = strings.ReplaceAll(escaped, "\t", "\\t")
		return []byte("\"" + escaped + "\"")
	}
	if needsYAMLQuoting(newValue) {
		escaped := strings.ReplaceAll(newValue, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		escaped = strings.ReplaceAll(escaped, "\n", "\\n")
		escaped = strings.ReplaceAll(escaped, "\r", "\\r")
		escaped = strings.ReplaceAll(escaped, "\t", "\\t")
		return []byte("\"" + escaped + "\"")
	}
	return []byte(newValue)
}

func needsYAMLQuoting(v string) bool {
	if v == "" {
		return true
	}
	if strings.TrimSpace(v) != v {
		return true
	}
	if strings.ContainsAny(v, "\t\n\r") {
		return true
	}
	if strings.ContainsAny(v, ":#{}[],&*?|-<>=!%@") {
		return true
	}
	switch strings.ToLower(v) {
	case "null", "~", "true", "false", "yes", "no", "on", "off":
		return true
	}
	return strings.HasPrefix(v, "-")
}

// setConfigValue sets a config field based on dot-notation key
func setConfigValue(cfg *config.Config, key, value string) error {
	parts := strings.Split(key, ".")

	if len(parts) == 0 {
		return fmt.Errorf("empty key")
	}

	switch parts[0] {
	case "base_folder":
		cfg.BaseFolder = value
	case "server":
		return setServerValue(cfg, parts[1:], value)
	case "database":
		return setDatabaseValue(cfg, parts[1:], value)
	case "scan_tactic":
		return setScanTacticValue(cfg, parts[1:], value)
	case "redis":
		return setRedisValue(cfg, parts[1:], value)
	case "global_vars":
		return setGlobalVarValue(cfg, parts[1:], value)
	case "notification":
		return setNotificationValue(cfg, parts[1:], value)
	case "environments":
		return setEnvironmentsValue(cfg, parts[1:], value)
	case "storage":
		return setStorageValue(cfg, parts[1:], value)
	case "llm_config":
		return setLLMValue(cfg, parts[1:], value)
	default:
		return fmt.Errorf("unknown config section: %s", parts[0])
	}
	return nil
}

// setServerValue sets a server config field
func setServerValue(cfg *config.Config, parts []string, value string) error {
	if len(parts) == 0 {
		return fmt.Errorf("missing server field")
	}
	switch parts[0] {
	case "host":
		cfg.Server.Host = value
	case "port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("port must be a number")
		}
		cfg.Server.Port = port
	case "ui_path":
		cfg.Server.UIPath = value
	case "workspace_prefix_key":
		cfg.Server.WorkspacePrefixKey = value
	case "username":
		oldUsername, oldPassword := primaryServerAuthUser(cfg)
		if cfg.Server.SimpleUserMapKey == nil {
			cfg.Server.SimpleUserMapKey = map[string]string{}
		}
		if oldUsername != "" {
			delete(cfg.Server.SimpleUserMapKey, oldUsername)
		}
		cfg.Server.SimpleUserMapKey[value] = oldPassword
	case "password":
		username, _ := primaryServerAuthUser(cfg)
		if username == "" {
			username = "osmedeus"
		}
		if cfg.Server.SimpleUserMapKey == nil {
			cfg.Server.SimpleUserMapKey = map[string]string{}
		}
		cfg.Server.SimpleUserMapKey[username] = value
	case "simple_user_map_key":
		if len(parts) < 2 {
			return fmt.Errorf("missing username (use server.simple_user_map_key.<username>)")
		}
		if cfg.Server.SimpleUserMapKey == nil {
			cfg.Server.SimpleUserMapKey = map[string]string{}
		}
		cfg.Server.SimpleUserMapKey[parts[1]] = value
	case "jwt":
		if len(parts) < 2 {
			return fmt.Errorf("missing jwt field (use server.jwt.secret_signing_key or server.jwt.expiration_minutes)")
		}
		return setJWTValue(cfg, parts[1:], value)
	case "enabled_auth_api":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("enabled_auth_api must be true or false")
		}
		cfg.Server.EnabledAuthAPI = enabled
	case "auth_api_key":
		cfg.Server.AuthAPIKey = value
	case "license":
		cfg.Server.License = value
	case "enable_metrics":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("enable_metrics must be true or false")
		}
		cfg.Server.EnableMetrics = &enabled
	case "cors_allowed_origins":
		cfg.Server.CORSAllowedOrigins = value
	case "event_receiver_url":
		cfg.Server.EventReceiverURL = value
	default:
		return fmt.Errorf("unknown server field: %s", parts[0])
	}
	return nil
}

func primaryServerAuthUser(cfg *config.Config) (string, string) {
	if cfg == nil {
		return "", ""
	}
	if len(cfg.Server.SimpleUserMapKey) == 0 {
		return "", ""
	}
	if pass, ok := cfg.Server.SimpleUserMapKey["osmedeus"]; ok {
		return "osmedeus", pass
	}
	keys := make([]string, 0, len(cfg.Server.SimpleUserMapKey))
	for k := range cfg.Server.SimpleUserMapKey {
		keys = append(keys, k)
	}
	sortStrings(keys)
	return keys[0], cfg.Server.SimpleUserMapKey[keys[0]]
}

func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[j] < s[i] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

func flattenYAMLScalars(node *yaml.Node, prefix string, out map[string]string) {
	if node == nil {
		return
	}
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) > 0 {
			flattenYAMLScalars(node.Content[0], prefix, out)
		}
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			k := node.Content[i]
			v := node.Content[i+1]
			if k.Kind != yaml.ScalarNode {
				continue
			}
			next := k.Value
			if prefix != "" {
				next = prefix + "." + next
			}
			flattenYAMLScalars(v, next, out)
		}
	case yaml.SequenceNode:
		for i, v := range node.Content {
			next := fmt.Sprintf("%s.%d", prefix, i)
			flattenYAMLScalars(v, next, out)
		}
	case yaml.ScalarNode:
		if prefix != "" {
			out[prefix] = node.Value
		}
	}
}

func isSensitiveKeyForDisplay(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "password") || strings.Contains(k, "secret") || strings.Contains(k, "_token") || strings.Contains(k, "token") || strings.Contains(k, "_key")
}

func redactValueForDisplay(key, value string, showSecrets bool) string {
	if showSecrets {
		return value
	}
	if isSensitiveKeyForDisplay(key) {
		return "[REDACTED]"
	}
	return value
}

func redactSensitiveFieldsYAML(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		if !isSensitiveKeyForDisplay(key) {
			continue
		}
		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		lines[i] = fmt.Sprintf("%s%s: \"[REDACTED]\"", indent, key)
	}
	return strings.Join(lines, "\n")
}

// setJWTValue sets JWT config fields
func setJWTValue(cfg *config.Config, parts []string, value string) error {
	if len(parts) == 0 {
		return fmt.Errorf("missing jwt field")
	}
	switch parts[0] {
	case "secret_signing_key":
		cfg.Server.JWT.SecretSigningKey = value
	case "expiration_minutes":
		minutes, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("expiration_minutes must be a number")
		}
		cfg.Server.JWT.ExpirationMinutes = minutes
	default:
		return fmt.Errorf("unknown jwt field: %s", parts[0])
	}
	return nil
}

// setDatabaseValue sets a database config field
func setDatabaseValue(cfg *config.Config, parts []string, value string) error {
	if len(parts) == 0 {
		return fmt.Errorf("missing database field")
	}
	switch parts[0] {
	case "db_engine":
		cfg.Database.DBEngine = value
	case "db_path":
		cfg.Database.DBPath = value
	case "host":
		cfg.Database.Host = value
	case "port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("port must be a number")
		}
		cfg.Database.Port = port
	case "username":
		cfg.Database.Username = value
	case "password":
		cfg.Database.Password = value
	case "db_name":
		cfg.Database.DBName = value
	case "connection_timeout":
		timeout, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("connection_timeout must be a number")
		}
		cfg.Database.ConnectionTimeout = timeout
	case "ssl_mode":
		cfg.Database.SSLMode = value
	default:
		return fmt.Errorf("unknown database field: %s", parts[0])
	}
	return nil
}

// setScanTacticValue sets a scan_tactic config field
func setScanTacticValue(cfg *config.Config, parts []string, value string) error {
	if len(parts) == 0 {
		return fmt.Errorf("missing scan_tactic field")
	}
	threads, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("value must be a number")
	}
	switch parts[0] {
	case "aggressive":
		cfg.ScanTactic.Aggressive = threads
	case "default":
		cfg.ScanTactic.Default = threads
	case "gently":
		cfg.ScanTactic.Gently = threads
	default:
		return fmt.Errorf("unknown scan_tactic field: %s (use aggressive, default, or gently)", parts[0])
	}
	return nil
}

// setRedisValue sets a redis config field
func setRedisValue(cfg *config.Config, parts []string, value string) error {
	if len(parts) == 0 {
		return fmt.Errorf("missing redis field")
	}
	switch parts[0] {
	case "host":
		cfg.Redis.Host = value
	case "port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("port must be a number")
		}
		cfg.Redis.Port = port
	case "username":
		cfg.Redis.Username = value
	case "password":
		cfg.Redis.Password = value
	case "db":
		db, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("db must be a number")
		}
		cfg.Redis.DB = db
	case "connection_timeout":
		timeout, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("connection_timeout must be a number")
		}
		cfg.Redis.ConnectionTimeout = timeout
	default:
		return fmt.Errorf("unknown redis field: %s", parts[0])
	}
	return nil
}

// setGlobalVarValue sets a global variable
func setGlobalVarValue(cfg *config.Config, parts []string, value string) error {
	if len(parts) == 0 {
		return fmt.Errorf("missing variable name (use global_vars.<name>)")
	}
	varName := parts[0]

	if cfg.GlobalVars == nil {
		cfg.GlobalVars = make(config.GlobalVarsConfig)
	}

	// Preserve existing as_env setting if variable exists
	existing, exists := cfg.GlobalVars[varName]
	if exists {
		existing.Value = value
		cfg.GlobalVars[varName] = existing
	} else {
		cfg.GlobalVars[varName] = config.GlobalVar{Value: value}
	}

	return nil
}

// setNotificationValue sets a notification config field
func setNotificationValue(cfg *config.Config, parts []string, value string) error {
	if len(parts) == 0 {
		return fmt.Errorf("missing notification field")
	}
	switch parts[0] {
	case "provider":
		cfg.Notification.Provider = value
	case "enabled":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("enabled must be true or false")
		}
		cfg.Notification.Enabled = enabled
	case "telegram":
		if len(parts) < 2 {
			return fmt.Errorf("missing telegram field")
		}
		return setTelegramValue(cfg, parts[1:], value)
	default:
		return fmt.Errorf("unknown notification field: %s", parts[0])
	}
	return nil
}

// setTelegramValue sets telegram config fields
func setTelegramValue(cfg *config.Config, parts []string, value string) error {
	if len(parts) == 0 {
		return fmt.Errorf("missing telegram field")
	}
	switch parts[0] {
	case "bot_token":
		cfg.Notification.Telegram.BotToken = value
	case "chat_id":
		chatID, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("chat_id must be a number")
		}
		cfg.Notification.Telegram.ChatID = chatID
	case "enabled":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("enabled must be true or false")
		}
		cfg.Notification.Telegram.Enabled = enabled
	default:
		return fmt.Errorf("unknown telegram field: %s", parts[0])
	}
	return nil
}

// setEnvironmentsValue sets an environments config field
func setEnvironmentsValue(cfg *config.Config, parts []string, value string) error {
	if len(parts) == 0 {
		return fmt.Errorf("missing environments field")
	}
	switch parts[0] {
	case "external_binaries_path":
		cfg.Environments.ExternalBinariesPath = value
	case "external_data":
		cfg.Environments.ExternalData = value
	case "external_configs":
		cfg.Environments.ExternalConfigs = value
	case "workspaces":
		cfg.Environments.Workspaces = value
	case "workflows":
		cfg.Environments.Workflows = value
	case "snapshot":
		cfg.Environments.Snapshot = value
	case "external_agent_configs":
		cfg.Environments.ExternalAgentConfigs = value
	case "markdown_report_templates":
		cfg.Environments.MarkdownReportTemplates = value
	case "external_scripts":
		cfg.Environments.ExternalScripts = value
	default:
		return fmt.Errorf("unknown environments field: %s", parts[0])
	}
	return nil
}

// setStorageValue sets a storage config field
func setStorageValue(cfg *config.Config, parts []string, value string) error {
	if len(parts) == 0 {
		return fmt.Errorf("missing storage field")
	}
	switch parts[0] {
	case "provider":
		cfg.Storage.Provider = value
	case "endpoint":
		cfg.Storage.Endpoint = value
	case "access_key_id":
		cfg.Storage.AccessKeyID = value
	case "secret_access_key":
		cfg.Storage.SecretAccessKey = value
	case "bucket":
		cfg.Storage.Bucket = value
	case "region":
		cfg.Storage.Region = value
	case "use_ssl":
		useSSL, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("use_ssl must be true or false")
		}
		cfg.Storage.UseSSL = useSSL
	case "enabled":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("enabled must be true or false")
		}
		cfg.Storage.Enabled = enabled
	case "account_id":
		cfg.Storage.AccountID = value
	case "path_style":
		pathStyle, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("path_style must be true or false")
		}
		cfg.Storage.PathStyle = pathStyle
	case "presign_expiry":
		cfg.Storage.PresignExpiry = value
	default:
		return fmt.Errorf("unknown storage field: %s", parts[0])
	}
	return nil
}

// setLLMValue sets an llm_config field
func setLLMValue(cfg *config.Config, parts []string, value string) error {
	if len(parts) == 0 {
		return fmt.Errorf("missing llm_config field")
	}
	switch parts[0] {
	case "llm_providers":
		// Handle llm_providers.<index>.<field> format
		// e.g., llm_providers.0.auth_token
		if len(parts) < 3 {
			return fmt.Errorf("usage: llm_config.llm_providers.<index>.<field> (e.g., llm_providers.0.auth_token)")
		}
		index, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("provider index must be a number: %s", parts[1])
		}
		// Auto-expand the slice if needed
		for len(cfg.LLM.LLMProviders) <= index {
			cfg.LLM.LLMProviders = append(cfg.LLM.LLMProviders, config.LLMProvider{})
		}
		return setLLMProviderField(&cfg.LLM.LLMProviders[index], parts[2], value)
	case "enabled_tool_call":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("enabled_tool_call must be true or false")
		}
		cfg.LLM.EnabledToolCall = enabled
	case "max_tokens":
		maxTokens, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("max_tokens must be a number")
		}
		cfg.LLM.MaxTokens = maxTokens
	case "temperature":
		temp, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("temperature must be a decimal number")
		}
		cfg.LLM.Temperature = temp
	case "top_k":
		topK, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("top_k must be a number")
		}
		cfg.LLM.TopK = topK
	case "top_p":
		topP, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("top_p must be a decimal number")
		}
		cfg.LLM.TopP = topP
	case "n":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("n must be a number")
		}
		cfg.LLM.N = n
	case "max_retries":
		maxRetries, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("max_retries must be a number")
		}
		cfg.LLM.MaxRetries = maxRetries
	case "timeout":
		cfg.LLM.Timeout = value
	case "stream":
		stream, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("stream must be true or false")
		}
		cfg.LLM.Stream = stream
	case "structured_json_format":
		structured, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("structured_json_format must be true or false")
		}
		cfg.LLM.StructuredJSONFormat = structured
	case "system_prompt":
		cfg.LLM.SystemPrompt = value
	case "custom_headers":
		cfg.LLM.CustomHeaders = value
	default:
		return fmt.Errorf("unknown llm_config field: %s", parts[0])
	}
	return nil
}

// setLLMProviderField sets a field on an LLMProvider
func setLLMProviderField(provider *config.LLMProvider, field string, value string) error {
	switch field {
	case "provider":
		provider.Provider = value
	case "base_url":
		provider.BaseURL = value
	case "auth_token":
		provider.AuthToken = value
	case "model":
		provider.Model = value
	default:
		return fmt.Errorf("unknown llm_provider field: %s (valid: provider, base_url, auth_token, model)", field)
	}
	return nil
}

// runConfigViewPattern handles wildcard pattern searches for config view
func runConfigViewPattern(pattern, settingsPath string, fileCfg *config.Config) error {
	// Read and parse YAML
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}

	var doc yaml.Node
	dec := yaml.NewDecoder(bytes.NewReader(content))
	dec.KnownFields(false)
	if err := dec.Decode(&doc); err != nil {
		return err
	}
	if len(doc.Content) == 0 {
		return fmt.Errorf("empty yaml document")
	}

	// Flatten YAML to key=value map
	out := map[string]string{}
	flattenYAMLScalars(doc.Content[0], "", out)

	// Add synthetic server.username/password keys
	username, password := primaryServerAuthUser(fileCfg)
	if username != "" {
		out["server.username"] = username
	}
	if password != "" {
		out["server.password"] = password
	}

	// Convert glob pattern to regex
	re, err := globToRegex(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	// Filter matching keys
	var matchingKeys []string
	for k := range out {
		if re.MatchString(k) {
			matchingKeys = append(matchingKeys, k)
		}
	}

	if len(matchingKeys) == 0 {
		return fmt.Errorf("no keys matching pattern: %s", pattern)
	}

	// Sort keys
	sortStrings(matchingKeys)

	// Print matching key=value pairs
	for _, k := range matchingKeys {
		v := out[k]
		if configViewRedact {
			v = redactValueForDisplay(k, v, false)
		}
		fmt.Printf("%s = %s\n", k, v)
	}

	return nil
}

// globToRegex converts a glob pattern (with * wildcards) to a compiled regex
func globToRegex(pattern string) (*regexp.Regexp, error) {
	// Escape regex special characters except *
	var sb strings.Builder
	sb.WriteString("^")
	for _, c := range pattern {
		switch c {
		case '*':
			sb.WriteString(".*")
		case '.', '+', '?', '^', '$', '(', ')', '[', ']', '{', '}', '|', '\\':
			sb.WriteRune('\\')
			sb.WriteRune(c)
		default:
			sb.WriteRune(c)
		}
	}
	sb.WriteString("$")
	return regexp.Compile(sb.String())
}
