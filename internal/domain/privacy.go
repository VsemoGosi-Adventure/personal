package domain

// PrivacySettings is the service contract for a future persistent settings store.
type PrivacySettings struct {
	UserID         string
	IsDiscoverable bool
	ShowLocation   bool
}

type PrivacySettingsUpdate struct {
	IsDiscoverable bool
	ShowLocation   bool
}
