package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

func Load(file string, config interface{}) error {
	v := viper.New()
	m := make(map[string]interface{})

	if err := mapstructure.Decode(config, &m); err != nil {
		return fmt.Errorf("mapstructure: %v", err)
	}

	if err := v.MergeConfigMap(m); err != nil {
		return fmt.Errorf("merge config map: %v", err)
	}

	v.SetConfigFile(file)
	if err := v.ReadInConfig(); err != nil {
		if e := new(os.PathError); !errors.As(err, &e) {
			return fmt.Errorf("%v", err)
		}
		log.Printf("[WARN] Config file %q not found. Use default and environment variables", file)
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.Unmarshal(config); err != nil {
		return fmt.Errorf("unmarshal config: %v", err)
	}

	return nil
}
