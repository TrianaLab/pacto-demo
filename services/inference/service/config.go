package service

// Config defines the inference service configuration.
type Config struct {
	Host           string `json:"host" doc:"Host address to bind the server" default:"0.0.0.0"`
	Port           int    `json:"port" doc:"Port to listen on" default:"8082"`
	RuntimeURL     string `json:"runtime_url" doc:"URL of the runtime service"`
	TimeoutSeconds int    `json:"timeout_seconds" doc:"Request timeout in seconds" default:"30"`
	MaxRetries     int    `json:"max_retries" doc:"Maximum number of retries for runtime calls" default:"3"`
	LogLevel       string `json:"log_level" doc:"Log level (debug, info, warn, error)" default:"info" enum:"debug,info,warn,error"`
}
