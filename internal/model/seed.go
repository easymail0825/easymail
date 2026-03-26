package model

import "errors"

// SeedDefaults inserts a minimal configuration tree for first-run databases.
// It is safe to call multiple times (best-effort idempotent).
func SeedDefaults(storageRoot, storageData string) error {
	if storageRoot == "" || storageData == "" {
		return errors.New("storageRoot/storageData must not be empty")
	}

	// Ensure root.
	if _, err := GetConfigureByName("easymail", 0); err != nil {
		_ = CreateRoot("easymail")
	}

	// Ensure storage.data
	if _, err := GetConfigureByNames("easymail", "storage", "data"); err != nil {
		_, _ = CreateConfigure(storageData, "storage data directory", DataTypeString, "easymail", "storage", "data")
	}

	// Ensure storage.root
	if _, err := GetConfigureByNames("easymail", "storage", "root"); err != nil {
		_, _ = CreateConfigure(storageRoot, "storage root directory", DataTypeString, "easymail", "storage", "root")
	}
	return nil
}
