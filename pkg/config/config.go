package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// GRPCServerURI is GRPC api used by CLI bind port
	GRPCServerURI string `yaml:"controllerGRPCServerURI,omitempty" envconfig:"GRPC_SERVER_URI"  default:":7000"`
	// LogLevel defines log level. Options: info, debug, trace
	LogLevel string `yaml:"logLevel,omitempty" envconfig:"LOG_LEVEL"  default:"debug"`
	// ProxyServerURI handles all requests and sends to either our webserver (8081) or to user provided URL.
	// Proxy job is to inject headers, like authentication if user provided one
	// Once proxy destination changes, webview will need to be triggered reload
	ProxyServerURI string `yaml:"proxyServerURI,omitempty" envconfig:"PROXY_SERVER_URI"  default:":8080"`
	// DefaultURI - always points to proxy. We load everything using our proxy to simplify injection of headers, credentials.
	// Webview points to DefaultURI
	DefaultURI string `yaml:"defaultURI,omitempty" envconfig:"DEFAULT_URI"  default:"http://127.0.0.1:8081"` // default url to open when starting
	// WebServerURI - URL of internal web server
	WebServerURI string `yaml:"webServerURI,omitempty" envconfig:"WEB_SERVER_URI"  default:":8081"` // web server bind port
	// StateDir defines where services keeps state
	StateDir string `yaml:"stateDir,omitempty" envconfig:"STATE_DIR"  default:"/data"` // where state is stored
	// Default webserver directory in the container to server content from
	WebServerDir string `yaml:"webServerDir,omitempty" envconfig:"WEB_SERVER_DIR"  default:"/www"` // Where web server expects page to be present
	// ProxyHeaders is key:value pairs of headers proxy will inject into requests. Example: "red:1,green:2,blue:3"
	ProxyHeaders map[string]string `yaml:"proxyHeaders,omitempty" envconfig:"PROXY_HEADERS"  default:""`
}

// Load loads the configuration from the environment.
func Load() (*Config, error) {
	// load .env files into env
	godotenv.Load()

	// load env into config
	c := &Config{}
	err := envconfig.Process("", c)
	if err != nil {
		return c, err
	}

	return c, err
}
