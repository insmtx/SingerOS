package engines

// GetCLIStatusSummary returns the current detection status for all built-in CLI engines.
func GetCLIStatusSummary() []CLIToolStatus {
	return DiscoverAvailableCLI()
}
