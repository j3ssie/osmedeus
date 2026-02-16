package cloud

import (
	"context"
	"time"
)

// ProviderType represents the cloud provider type
type ProviderType string

const (
	ProviderAWS          ProviderType = "aws"
	ProviderGCP          ProviderType = "gcp"
	ProviderDigitalOcean ProviderType = "digitalocean"
	ProviderLinode       ProviderType = "linode"
	ProviderAzure        ProviderType = "azure"
)

// ExecutionMode represents the execution mode (VM or serverless)
type ExecutionMode string

const (
	ModeVM         ExecutionMode = "vm"
	ModeServerless ExecutionMode = "serverless"
)

// Provider is the interface that all cloud providers must implement
type Provider interface {
	// Validate checks if the provider configuration is valid
	Validate(ctx context.Context) error

	// EstimateCost estimates the cost for the given configuration
	EstimateCost(mode ExecutionMode, instanceCount int) (*CostEstimate, error)

	// CreateInfrastructure provisions the cloud resources
	CreateInfrastructure(ctx context.Context, opts *CreateOptions) (*Infrastructure, error)

	// DestroyInfrastructure tears down the cloud resources
	DestroyInfrastructure(ctx context.Context, infra *Infrastructure) error

	// GetStatus retrieves the current status of infrastructure
	GetStatus(ctx context.Context, infra *Infrastructure) (*InfraStatus, error)

	// Type returns the provider type
	Type() ProviderType
}

// CreateOptions contains options for creating infrastructure
type CreateOptions struct {
	// Mode is the execution mode (vm or serverless)
	Mode ExecutionMode

	// InstanceCount is the number of instances to create
	InstanceCount int

	// InstanceType is the instance size/type (e.g., "s-2vcpu-4gb" for DigitalOcean)
	InstanceType string

	// UseSpot indicates whether to use spot/preemptible instances
	UseSpot bool

	// ImageID is the cloud image/snapshot to use
	ImageID string

	// RedisURL is the master Redis URL for worker registration
	RedisURL string

	// SSHPublicKey is the SSH public key for authentication
	SSHPublicKey string

	// SSHPrivateKey is the SSH private key for result collection
	SSHPrivateKey string

	// SSHUser is the SSH username
	SSHUser string

	// SetupCommands are additional commands to run on boot
	SetupCommands []string

	// Tags are resource tags/labels
	Tags map[string]string

	// Timeout is the maximum time to wait for infrastructure creation
	Timeout time.Duration
}

// Infrastructure represents created cloud infrastructure
type Infrastructure struct {
	// ID is a unique identifier for this infrastructure
	ID string

	// Provider is the cloud provider type
	Provider ProviderType

	// Mode is the execution mode
	Mode ExecutionMode

	// CreatedAt is when the infrastructure was created
	CreatedAt time.Time

	// PulumiStackID is the Pulumi stack identifier
	PulumiStackID string

	// Resources are the created resources (VMs, functions, etc.)
	Resources []Resource

	// Metadata contains additional provider-specific data
	Metadata map[string]interface{}
}

// Resource represents a single cloud resource
type Resource struct {
	// Type is the resource type (e.g., "vm", "function")
	Type string

	// ID is the cloud provider's resource ID
	ID string

	// Name is a human-readable name
	Name string

	// PublicIP is the public IP address (if applicable)
	PublicIP string

	// PrivateIP is the private IP address (if applicable)
	PrivateIP string

	// SSHEnabled indicates if SSH access is available
	SSHEnabled bool

	// WorkerID is the Osmedeus worker ID after registration
	WorkerID string

	// Status is the current resource status
	Status string

	// Metadata contains additional resource-specific data
	Metadata map[string]interface{}
}

// InfraStatus represents the status of infrastructure
type InfraStatus struct {
	// Status is the overall status (running, stopped, error, etc.)
	Status string

	// ReadyCount is the number of ready resources
	ReadyCount int

	// TotalCount is the total number of resources
	TotalCount int

	// WorkersRegistered is the number of workers that have registered
	WorkersRegistered int

	// Details contains status details for each resource
	Details []ResourceStatus
}

// ResourceStatus represents the status of a single resource
type ResourceStatus struct {
	// ResourceID is the resource identifier
	ResourceID string

	// Status is the resource status
	Status string

	// Message provides additional status information
	Message string

	// WorkerRegistered indicates if the worker has registered
	WorkerRegistered bool
}

// CostEstimate represents estimated costs
type CostEstimate struct {
	// HourlyCost is the estimated hourly cost
	HourlyCost float64

	// DailyCost is the estimated daily cost
	DailyCost float64

	// Currency is the currency code (default: USD)
	Currency string

	// Breakdown provides cost breakdown by resource type
	Breakdown map[string]float64

	// Notes contains additional cost information
	Notes []string
}
