package core

// Project metadata constants
const (
	// VERSION of this project
	VERSION = "v5.0.0-beta"
	// DESC description of the tool
	DESC = "A Modern Orchestration Engine for Security"
	// BINARY name of osmedeus
	BINARY = "osmedeus"
	// SNAPSHOT binary name of osmedeus
	SNAPSHOT = "osm"
	// AUTHOR of this
	AUTHOR = "@j3ssie"
	// DOCS private document
	DOCS = "https://docs.osmedeus.org"
	// DOCS private document
	LICENSE = "open-source"
	// REPO_URL private document
	REPO_URL = "https://github.com/j3ssie/osmedeus"
	// DEFAULT_BASE_REPO default repository for base folder
	DEFAULT_BASE_REPO = "https://github.com/osmedeus/osmedeus-base.git"
	// DEFAULT_WORKFLOW_REPO default repository for workflows
	DEFAULT_WORKFLOW_REPO = "https://github.com/osmedeus/osmedeus-workflow.git"
	// METADATA domain for checking update
	METADATA = "https://metadata.osmedeus.org"
	// INSTALL default install script
	INSTALL = "https://raw.githubusercontent.com/osmedeus/osmedeus-base/master/install.sh"
	// DefaultUA is the default User-Agent for HTTP clients
	DefaultUA = "Mozilla/5.0 (compatible; Osmedeus/" + VERSION + "; +" + REPO_URL + ")"
)
