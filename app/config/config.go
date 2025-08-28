package config

const (
	// Server configuration
	DefaultPort = "6379"
	DefaultHost = "0.0.0.0"
	DefaultAddress = DefaultHost + ":" + DefaultPort
	
	// Redis protocol constants
	MaxBulkStringLength = 512 * 1024 * 1024 // 512MB
	MaxArrayLength = 1024 * 1024            // 1M elements
	
	// Server limits
	MaxConnections = 10000
	ReadTimeout = 30 // seconds
	WriteTimeout = 30 // seconds
)