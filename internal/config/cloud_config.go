package config

// CloudConfigs represents the cloud configuration schema
type CloudConfigs struct {
	Providers Providers `yaml:"providers"`
	Defaults  Defaults  `yaml:"defaults"`
	Limits    Limits    `yaml:"limits"`
	State     State     `yaml:"state"`
	SSH       SSH       `yaml:"ssh"`
	Setup     Setup     `yaml:"setup"`
}

// Providers contains credentials for all cloud providers
type Providers struct {
	AWS          AWSConfig          `yaml:"aws"`
	GCP          GCPConfig          `yaml:"gcp"`
	DigitalOcean DigitalOceanConfig `yaml:"digitalocean"`
	Linode       LinodeConfig       `yaml:"linode"`
	Azure        AzureConfig        `yaml:"azure"`
}

// AWSConfig contains AWS credentials and configuration
type AWSConfig struct {
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Region          string `yaml:"region"`
	InstanceType    string `yaml:"instance_type"`
	AMI             string `yaml:"ami"`
	UseSpot         bool   `yaml:"use_spot"`
}

// GCPConfig contains GCP credentials and configuration
type GCPConfig struct {
	ProjectID       string `yaml:"project_id"`
	CredentialsFile string `yaml:"credentials_file"`
	Region          string `yaml:"region"`
	Zone            string `yaml:"zone"`
	MachineType     string `yaml:"machine_type"`
	ImageFamily     string `yaml:"image_family"`
	UsePreemptible  bool   `yaml:"use_preemptible"`
}

// DigitalOceanConfig contains DigitalOcean credentials and configuration
type DigitalOceanConfig struct {
	Token             string `yaml:"token"`
	Region            string `yaml:"region"`
	Size              string `yaml:"size"`
	Image             string `yaml:"image"`
	SnapshotID        string `yaml:"snapshot_id"`
	SSHKeyID          string `yaml:"ssh_key_id"`
	SSHKeyFingerprint string `yaml:"ssh_key_fingerprint"`
}

// LinodeConfig contains Linode credentials and configuration
type LinodeConfig struct {
	Token        string `yaml:"token"`
	Region       string `yaml:"region"`
	Type         string `yaml:"type"`
	Image        string `yaml:"image"`
	SSHPublicKey string `yaml:"ssh_public_key"`
}

// AzureConfig contains Azure credentials and configuration
type AzureConfig struct {
	SubscriptionID string `yaml:"subscription_id"`
	TenantID       string `yaml:"tenant_id"`
	ClientID       string `yaml:"client_id"`
	ClientSecret   string `yaml:"client_secret"`
	Location       string `yaml:"location"`
	VMSize         string `yaml:"vm_size"`
	ImageReference string `yaml:"image_reference"`
}

// Defaults contains default cloud configuration values
type Defaults struct {
	Provider         string `yaml:"provider"`
	Mode             string `yaml:"mode"`
	MaxInstances     int    `yaml:"max_instances"`
	UseSpot          bool   `yaml:"use_spot"`
	Timeout          string `yaml:"timeout"`
	CleanupOnFailure bool   `yaml:"cleanup_on_failure"`
}

// Limits contains cost and resource limits
type Limits struct {
	MaxHourlySpend float64 `yaml:"max_hourly_spend"`
	MaxTotalSpend  float64 `yaml:"max_total_spend"`
	MaxInstances   int     `yaml:"max_instances"`
}

// State contains state storage configuration
type State struct {
	Backend string `yaml:"backend"`
	Path    string `yaml:"path"`
}

// SSH contains SSH configuration for accessing workers
type SSH struct {
	PrivateKeyPath    string `yaml:"private_key_path"`
	PrivateKeyContent string `yaml:"private_key_content"`
	PublicKeyPath     string `yaml:"public_key_path"`
	PublicKeyContent  string `yaml:"public_key_content"`
	User              string `yaml:"user"`
}

// Setup contains worker setup configuration
type Setup struct {
	Commands []string `yaml:"commands"`
}

// DefaultCloudConfigs returns a default cloud configuration
func DefaultCloudConfigs() *CloudConfigs {
	return &CloudConfigs{
		Providers: Providers{
			AWS: AWSConfig{
				AccessKeyID:     "${AWS_ACCESS_KEY_ID}",
				SecretAccessKey: "${AWS_SECRET_ACCESS_KEY}",
				Region:          "us-east-1",
				InstanceType:    "t3.medium",
				UseSpot:         false,
			},
			GCP: GCPConfig{
				ProjectID:       "${GCP_PROJECT_ID}",
				CredentialsFile: "${GCP_CREDENTIALS_FILE}",
				Region:          "us-central1",
				Zone:            "us-central1-a",
				MachineType:     "n1-standard-2",
				UsePreemptible:  false,
			},
			DigitalOcean: DigitalOceanConfig{
				Token:  "${DIGITALOCEAN_TOKEN}",
				Region: "nyc1",
				Size:   "s-2vcpu-4gb",
				Image:  "ubuntu-22-04-x64",
			},
			Linode: LinodeConfig{
				Token:  "${LINODE_TOKEN}",
				Region: "us-east",
				Type:   "g6-standard-2",
				Image:  "linode/ubuntu22.04",
			},
			Azure: AzureConfig{
				SubscriptionID: "${AZURE_SUBSCRIPTION_ID}",
				TenantID:       "${AZURE_TENANT_ID}",
				ClientID:       "${AZURE_CLIENT_ID}",
				ClientSecret:   "${AZURE_CLIENT_SECRET}",
				Location:       "eastus",
				VMSize:         "Standard_B2s",
			},
		},
		Defaults: Defaults{
			Provider:         "digitalocean",
			Mode:             "vm",
			MaxInstances:     10,
			UseSpot:          false,
			Timeout:          "30m",
			CleanupOnFailure: true,
		},
		Limits: Limits{
			MaxHourlySpend: 10.0,
			MaxTotalSpend:  100.0,
			MaxInstances:   20,
		},
		State: State{
			Backend: "local",
			Path:    "{{base_folder}}/cloud-state",
		},
		SSH: SSH{
			PrivateKeyPath: "~/.ssh/id_rsa",
			User:           "root",
		},
		Setup: Setup{
			Commands: []string{
				"# Add custom setup commands here",
			},
		},
	}
}
