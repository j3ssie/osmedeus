package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchesAnyVariableType(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		typeSpec VariableType
		want     bool
		wantErr  bool
	}{
		{
			name:     "single type domain matches domain",
			value:    "example.com",
			typeSpec: VarTypeDomain,
			want:     true,
		},
		{
			name:     "single type domain rejects url",
			value:    "https://example.com",
			typeSpec: VarTypeDomain,
			want:     false,
		},
		{
			name:     "comma-separated domain,url accepts domain",
			value:    "example.com",
			typeSpec: "domain,url",
			want:     true,
		},
		{
			name:     "comma-separated domain,url accepts url",
			value:    "https://example.com",
			typeSpec: "domain,url",
			want:     true,
		},
		{
			name:     "comma-separated domain,url rejects ip",
			value:    "192.168.1.1",
			typeSpec: "domain,url",
			want:     false,
		},
		{
			name:     "comma-separated with spaces works",
			value:    "example.com",
			typeSpec: "domain, url",
			want:     true,
		},
		{
			name:     "comma-separated url,domain (reversed order) accepts domain",
			value:    "example.com",
			typeSpec: "url,domain",
			want:     true,
		},
		{
			name:     "comma-separated url,domain (reversed order) accepts url",
			value:    "https://example.com",
			typeSpec: "url,domain",
			want:     true,
		},
		{
			name:     "three types works",
			value:    "10.0.0.0/8",
			typeSpec: "domain,url,cidr",
			want:     true,
		},
		{
			name:     "string type accepts anything",
			value:    "anything",
			typeSpec: "string",
			want:     true,
		},
		{
			name:     "number type accepts number",
			value:    "123",
			typeSpec: "number",
			want:     true,
		},
		{
			name:     "number type rejects non-number",
			value:    "abc",
			typeSpec: "number",
			want:     false,
		},
		{
			name:     "ip type accepts IPv4",
			value:    "192.168.1.1",
			typeSpec: VarTypeIP,
			want:     true,
		},
		{
			name:     "ip type accepts IPv6",
			value:    "2001:db8::1",
			typeSpec: VarTypeIP,
			want:     true,
		},
		{
			name:     "ip type rejects domain",
			value:    "example.com",
			typeSpec: VarTypeIP,
			want:     false,
		},
		{
			name:     "ip type rejects invalid ip",
			value:    "999.999.999.999",
			typeSpec: VarTypeIP,
			want:     false,
		},
		{
			name:     "comma-separated ip,cidr accepts ip",
			value:    "162.13.44.21",
			typeSpec: "ip,cidr",
			want:     true,
		},
		{
			name:     "comma-separated ip,cidr accepts cidr",
			value:    "10.0.0.0/8",
			typeSpec: "ip,cidr",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchesAnyVariableType(tt.value, tt.typeSpec)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMatchesAnyTargetType(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		typeSpec TargetType
		want     bool
		wantErr  bool
	}{
		{
			name:     "single type domain matches domain",
			target:   "example.com",
			typeSpec: TargetTypeDomain,
			want:     true,
		},
		{
			name:     "single type domain rejects url",
			target:   "https://example.com",
			typeSpec: TargetTypeDomain,
			want:     false,
		},
		{
			name:     "comma-separated domain,url accepts domain",
			target:   "example.com",
			typeSpec: "domain,url",
			want:     true,
		},
		{
			name:     "comma-separated domain,url accepts url",
			target:   "https://example.com",
			typeSpec: "domain,url",
			want:     true,
		},
		{
			name:     "comma-separated domain,url rejects ip",
			target:   "192.168.1.1",
			typeSpec: "domain,url",
			want:     false,
		},
		{
			name:     "comma-separated with spaces works",
			target:   "example.com",
			typeSpec: "domain, url",
			want:     true,
		},
		{
			name:     "comma-separated url,domain (reversed order) accepts domain",
			target:   "example.com",
			typeSpec: "url,domain",
			want:     true,
		},
		{
			name:     "comma-separated url,domain (reversed order) accepts url",
			target:   "https://example.com",
			typeSpec: "url,domain",
			want:     true,
		},
		{
			name:     "string type accepts anything",
			target:   "anything",
			typeSpec: "string",
			want:     true,
		},
		{
			name:     "ip type accepts IPv4",
			target:   "162.13.44.21",
			typeSpec: TargetTypeIP,
			want:     true,
		},
		{
			name:     "ip type accepts IPv6",
			target:   "2001:db8::1",
			typeSpec: TargetTypeIP,
			want:     true,
		},
		{
			name:     "ip type rejects domain",
			target:   "example.com",
			typeSpec: TargetTypeIP,
			want:     false,
		},
		{
			name:     "comma-separated ip,cidr accepts ip",
			target:   "192.168.1.1",
			typeSpec: "ip,cidr",
			want:     true,
		},
		{
			name:     "comma-separated ip,cidr accepts cidr",
			target:   "10.0.0.0/8",
			typeSpec: "ip,cidr",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchesAnyTargetType(tt.target, tt.typeSpec)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
