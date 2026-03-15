package service

// Config defines the API gateway configuration.
type Config struct {
	Host           string `json:"host" doc:"Host address to bind the server" default:"0.0.0.0"`
	Port           int    `json:"port" doc:"Port to listen on" default:"8080"`
	InferenceURL   string `json:"inference_url" doc:"URL of the inference service"`
	APIKey         string `json:"api_key" doc:"API key for authentication"`
	RateLimitRPS   int    `json:"rate_limit_rps" doc:"Rate limit in requests per second" default:"100"`
	CORSOrigins    string `json:"cors_origins" doc:"Allowed CORS origins" default:"*"`
	LogLevel       string `json:"log_level" doc:"Log level (debug, info, warn, error)" default:"info" enum:"debug,info,warn,error"`
}
