package core

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// Dependencies defines workflow requirements
type Dependencies struct {
	Commands            []string      `yaml:"commands"`
	Files               []string      `yaml:"files"`
	Variables           []VariableDep `yaml:"variables"`
	TargetTypes         []TargetType  `yaml:"target_types"`
	FunctionsConditions []string      `yaml:"functions_conditions"`
}

// VariableDep defines variable requirements
type VariableDep struct {
	Name     string       `yaml:"name"`
	Type     VariableType `yaml:"type"`
	Required bool         `yaml:"required"`
}

// HasCommandDeps returns true if there are command dependencies
func (d *Dependencies) HasCommandDeps() bool {
	return d != nil && len(d.Commands) > 0
}

// HasFileDeps returns true if there are file dependencies
func (d *Dependencies) HasFileDeps() bool {
	return d != nil && len(d.Files) > 0
}

// HasVariableDeps returns true if there are variable dependencies
func (d *Dependencies) HasVariableDeps() bool {
	return d != nil && len(d.Variables) > 0
}

// HasFunctionConditions returns true if there are function-based condition dependencies
func (d *Dependencies) HasFunctionConditions() bool {
	return d != nil && len(d.FunctionsConditions) > 0
}

// GetRequiredVariables returns all required variable dependencies
func (d *Dependencies) GetRequiredVariables() []VariableDep {
	if d == nil {
		return nil
	}
	var required []VariableDep
	for _, v := range d.Variables {
		if v.Required {
			required = append(required, v)
		}
	}
	return required
}

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
var numericRegex = regexp.MustCompile(`^-?\d+(\.\d+)?$`)
var simpleRepoRegex = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
var hostedRepoRegex = regexp.MustCompile(`^(?:(?:https?://)?(?:www\.)?(github\.com|gitlab\.com)/)([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+?)(?:\.git)?/?$`)
var sshRepoRegex = regexp.MustCompile(`^git@[^:]+:([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+?)(\.git)?$`)

func MatchesVariableType(value string, varType VariableType) (bool, error) {
	switch varType {
	case VarTypeDomain:
		return domainRegex.MatchString(value), nil
	case VarTypeSubdomain:
		if !domainRegex.MatchString(value) {
			return false, nil
		}
		return strings.Count(value, ".") >= 2, nil
	case VarTypeURL:
		u, err := url.Parse(value)
		if err != nil {
			return false, nil
		}
		if u.Scheme == "http" || u.Scheme == "https" {
			return u.Host != "", nil
		}
		return false, nil
	case VarTypeCIDR:
		_, _, err := net.ParseCIDR(value)
		return err == nil, nil
	case VarTypeIP:
		return net.ParseIP(value) != nil, nil
	case VarTypeRepo:
		return isRepo(value), nil
	case VarTypePath, VarTypeFile, VarTypeFolder:
		return value != "", nil
	case VarTypeNumber:
		return numericRegex.MatchString(value), nil
	case VarTypeString:
		return true, nil
	default:
		return false, fmt.Errorf("unknown variable type: %s", varType)
	}
}

// MatchesAnyVariableType checks if value matches ANY of the comma-separated types.
// Returns true if value matches at least one type.
func MatchesAnyVariableType(value string, typeSpec VariableType) (bool, error) {
	types := strings.Split(string(typeSpec), ",")

	for _, t := range types {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		matches, err := MatchesVariableType(value, VariableType(t))
		if err != nil {
			return false, err
		}
		if matches {
			return true, nil
		}
	}
	return false, nil
}

// MatchesAnyTargetType checks if target matches ANY of the comma-separated types.
// Returns true if target matches at least one type.
func MatchesAnyTargetType(target string, typeSpec TargetType) (bool, error) {
	types := strings.Split(string(typeSpec), ",")

	for _, t := range types {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		matches, err := MatchesTargetType(target, TargetType(t))
		if err != nil {
			return false, err
		}
		if matches {
			return true, nil
		}
	}
	return false, nil
}

func MatchesTargetType(target string, targetType TargetType) (bool, error) {
	switch targetType {
	case TargetTypeDomain:
		return MatchesVariableType(target, VarTypeDomain)
	case TargetTypeSubdomain:
		return MatchesVariableType(target, VarTypeSubdomain)
	case TargetTypeURL:
		return MatchesVariableType(target, VarTypeURL)
	case TargetTypeCIDR:
		return MatchesVariableType(target, VarTypeCIDR)
	case TargetTypeRepo:
		return MatchesVariableType(target, VarTypeRepo)
	case TargetTypePath:
		return MatchesVariableType(target, VarTypePath)
	case TargetTypeNumber:
		return MatchesVariableType(target, VarTypeNumber)
	case TargetTypeString:
		return true, nil
	case TargetTypeIP:
		return MatchesVariableType(target, VarTypeIP)
	case TargetTypeFile:
		info, err := os.Stat(target)
		if err != nil {
			return false, nil
		}
		return !info.IsDir(), nil
	case TargetTypeFolder:
		info, err := os.Stat(target)
		if err != nil {
			return false, nil
		}
		return info.IsDir(), nil
	default:
		return false, fmt.Errorf("unknown target type: %s", targetType)
	}
}

func isRepo(value string) bool {
	if simpleRepoRegex.MatchString(value) {
		return true
	}
	if hostedRepoRegex.MatchString(value) {
		return true
	}
	if sshRepoRegex.MatchString(value) {
		return true
	}
	u, err := url.Parse(value)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "ssh" || u.Scheme == "git") {
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) >= 2 {
			owner := parts[0]
			repo := strings.TrimSuffix(parts[1], ".git")
			if owner != "" && repo != "" {
				return true
			}
		}
	}
	return false
}
