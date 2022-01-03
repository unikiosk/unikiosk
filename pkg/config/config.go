package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/phayes/freeport"
)

type Config struct {
	// Address sections
	// GRPCServerAddr is GRPC API address bind used by CLI
	GRPCServerAddr string `yaml:"controllerGRPCServerAddr,omitempty" envconfig:"GRPC_SERVER_ADDR"  default:":7000"`

	// ProxyServerAddr defines address to which Proxy should bind. It handles all requests and sends them to either our webserver (8081) or to user provided URL.
	// Proxy purpose is to inject headers, like authentication.
	// Once proxy destination changes, webview will need to be triggered reload
	ProxyHTTPServerAddr  string `yaml:"proxyHTTPServerAddr,omitempty" envconfig:"PROXY_HTTP_SERVER_ADDR"  default:""`   // Default -random
	ProxyHTTPSServerAddr string `yaml:"proxyHTTPSServerAddr,omitempty" envconfig:"PROXY_HTTPS_SERVER_ADDR"  default:""` // Default -random

	// MITM proxy cert. Must be trusted by executor OS
	ProxyHTTPSCertLocation    string `yaml:"proxyHTTPSCertLocation,omitempty" envconfig:"PROXY_HTTPS_CERT"  default:"rootCA.pem"`
	ProxyHTTPSCertKeyLocation string `yaml:"proxyHTTPSCertKeyLocation,omitempty" envconfig:"PROXY_HTTPS_CERT_KEY"  default:"rootCA-key.pem"`

	// WebServerAddr - address of where internal web server binds
	WebServerAddr string `yaml:"webServerAddr,omitempty" envconfig:"WEB_SERVER_ADDR"  default:":8081"` // web server bind port

	// default routing section
	// DefaultHTTPProxyURL - is proxy URL. It used in webview to route requests via proxy
	// Populated automatically
	DefaultHTTPProxyURL string

	// default routing section
	// DefaultHTTPProxyURL - is proxy URL. It used in webview to route requests via proxy
	// Populated automatically
	DefaultHTTPSProxyURL string

	// DefaultWebServerURL is default webserver url. Used to serve default content
	// Populated automatically
	DefaultWebServerURL string
	// ProxyHeaders is key:value pairs of headers proxy will inject into requests. Example: "red:1,green:2,blue:3"
	ProxyHeaders map[string]string `yaml:"proxyHeaders,omitempty" envconfig:"PROXY_HEADERS"  default:""`

	// LogLevel defines log level. Options: info, debug, trace
	LogLevel string `yaml:"logLevel,omitempty" envconfig:"LOG_LEVEL"  default:"debug"`
	// StateDir defines where services keeps state
	StateDir string `yaml:"stateDir,omitempty" envconfig:"STATE_DIR"  default:"/data"` // where state is stored
	// Default webserver directory in the container to server content from
	WebServerDir string `yaml:"webServerDir,omitempty" envconfig:"WEB_SERVER_DIR"  default:"/www"` // Where web server expects page to be present
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

	if c.ProxyHTTPServerAddr == "" {
		port, err := freeport.GetFreePort()
		if err != nil {
			return nil, fmt.Errorf("failed to allocate free port: %w", err)
		}
		c.ProxyHTTPServerAddr = fmt.Sprintf(":%d", port)
	}
	if c.ProxyHTTPSServerAddr == "" {
		port, err := freeport.GetFreePort()
		if err != nil {
			return nil, fmt.Errorf("failed to allocate free port: %w", err)
		}
		c.ProxyHTTPSServerAddr = fmt.Sprintf(":%d", port)
	}

	// TODO: add check if user provides full bind URL for proxy server address
	c.DefaultHTTPProxyURL = "0.0.0.0" + c.ProxyHTTPServerAddr
	c.DefaultHTTPSProxyURL = "0.0.0.0" + c.ProxyHTTPSServerAddr

	// TODO: add check if user provided full bind URL for webserver
	c.DefaultWebServerURL = "http://0.0.0.0" + c.WebServerAddr

	return c, err
}
