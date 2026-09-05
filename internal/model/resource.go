// Package model defines the provider-agnostic types every cloud provider
// implementation maps its resources into, so the TUI never needs to know
// about AWS/GCP/Azure specifics.
package model

import "time"

// Category groups resources for display (tabs in the TUI, sidebar counts).
type Category string

const (
	CategoryCompute   Category = "Compute"
	CategoryStorage   Category = "Storage"
	CategoryDatabase  Category = "Databases"
	CategoryNetwork   Category = "Networking"
	CategoryUnknown   Category = "Other"
)

// AllCategories defines the fixed tab order shown in the TUI.
var AllCategories = []Category{CategoryCompute, CategoryStorage, CategoryDatabase, CategoryNetwork}

// Health summarizes a resource's state into one of three severities so the
// UI can color it consistently regardless of the underlying provider's own
// vocabulary (e.g. EC2 "running" vs RDS "available" both map to Healthy).
type Health string

const (
	HealthOK      Health = "ok"
	HealthWarn    Health = "warn"
	HealthDown    Health = "down"
	HealthUnknown Health = "unknown"
)

// Resource is one discovered cloud resource, normalized across providers.
type Resource struct {
	Provider  string // "aws", "gcp", "azure", ...
	Account   string // profile name / account alias, human-facing
	AccountID string // provider account/subscription/project id, if known
	Region    string
	Category  Category
	Type      string // "ec2-instance", "s3-bucket", "rds-instance", ...
	ID        string
	Name      string
	Status    string // raw provider status string, shown verbatim
	Health    Health
	Details   []Detail // ordered key/value pairs for the detail view
}

// Detail is a single labeled field shown in a resource's detail card.
type Detail struct {
	Label string
	Value string
}

// Account is one credential/identity a provider discovered locally.
type Account struct {
	Provider string
	Name     string // profile name
	ID       string // account id, filled in lazily once resolved
	Region   string // default region for this account
}

// AccountStatus tracks the live health of one account's polling loop.
type AccountStatus struct {
	Account       Account
	Health        Health
	Error         string
	ResourceCount int
	LastRefresh   time.Time
}
