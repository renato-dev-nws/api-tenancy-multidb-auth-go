package shared

// BillingCycle representa os ciclos de cobrança disponíveis
type BillingCycle string

const (
	BillingCycleMonthly    BillingCycle = "monthly"
	BillingCycleQuarterly  BillingCycle = "quarterly"
	BillingCycleSemiannual BillingCycle = "semiannual"
	BillingCycleAnnual     BillingCycle = "annual"
)

// TenantStatus representa os status possíveis de um tenant
type TenantStatus string

const (
	TenantStatusProvisioning TenantStatus = "provisioning"
	TenantStatusActive       TenantStatus = "active"
	TenantStatusSuspended    TenantStatus = "suspended"
)
