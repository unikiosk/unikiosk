package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GRPCServerURI       string `yaml:"controllerGRPCServerURI,omitempty" envconfig:"GRPC_SERVER_URI"  default:":7000"`
	LogLevel            string `yaml:"logLevel,omitempty" envconfig:"LOG_LEVEL"  default:"debug"`
	DefaultURI          string `yaml:"defaultURI,omitempty" envconfig:"DEFAULT_URI"  default:"https://synpse.net"`
	DefaultPageLocation string `yaml:"defaultPageLocation,omitempty" envconfig:"DEFAULT_PAGE_LOCATIOn"  default:""`
	StateDir            string `yaml:"stateDir,omitempty" envconfig:"STATE_DIR"  default:"/data"`
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
