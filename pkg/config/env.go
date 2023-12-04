package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Domain struct {
	Name           string
	ZoneIdentifier string `mapstructure:"zone_identifier"`
	DNS            []DNS  `mapstructure:"record_name" validate:"required"`
}

type DNS struct {
	Name     string `validate:"required"`
	TTL      int    `mapstructure:"ttl"`
	UseProxy bool   `mapstructure:"use_proxy"`
}

type EnvConfigMap struct {
	Cloudflare struct {
		API struct {
			ID string `mapstructure:"id" validate:"required"`
		} `mapstructure:"api"`
		Domain []Domain `mapstructure:"Domain" validate:"required"`
	} `mapstructure:"cloudflare"`
	SeekIPURL string `mapstructure:"seek_ip_url"`
}

// ENV is global variable for using config in other place
var ENV EnvConfigMap

// LoadConfig read env file and loaded to environment and global ENV variable
func LoadConfig(cfgFile string) error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		//viper.SetConfigName(".obm-bot-crawler")
		viper.SetConfigFile("ddns.yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err == nil {
		_, _ = fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		return err
	}

	err = viper.Unmarshal(&ENV)
	if err != nil {
		return err
	}

	err = validateConfig(ENV)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Validate error: %v", err)
		return err
	}

	return nil
}

func validateConfig(config EnvConfigMap) error {
	if config.Cloudflare.API.ID == "" {
		return errors.New("cloudflare.api.id must be not null or empty")
	}
	if len(config.Cloudflare.Domain) == 0 {
		return errors.New("cloudflare.api.domain must be not null or empty")
	}

	for index, domain := range config.Cloudflare.Domain {
		if domain.ZoneIdentifier == "" {
			return fmt.Errorf("cloudflare_domain[%v].zone_identifier_must_be_not_null_or_empty", index)
		}
		if len(domain.DNS) == 0 {
			return fmt.Errorf("cloudflare_domain[%v].domain must be not null or empty", index)
		}
		for dnsIndex, dns := range domain.DNS {
			if dns.Name == "" {
				return fmt.Errorf("cloudflare_domain[%v].domain[%v].name must be not null or empty", index, dnsIndex)
			}
		}
	}
	return nil
}
