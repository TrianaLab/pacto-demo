package service

// Config defines the runtime service configuration.
type Config struct {
	Host         string `json:"host" doc:"Host address to bind the server" default:"0.0.0.0"`
	Port         int    `json:"port" doc:"Port to listen on" default:"8081"`
	ModelPath    string `json:"model_path" doc:"Path to the ML model files"`
	MaxBatchSize int    `json:"max_batch_size" doc:"Maximum batch size for inference requests" default:"32"`
	GPUEnabled   bool   `json:"gpu_enabled" doc:"Enable GPU acceleration" default:"false"`
	LogLevel     string `json:"log_level" doc:"Log level (debug, info, warn, error)" default:"info" enum:"debug,info,warn,error"`
}
