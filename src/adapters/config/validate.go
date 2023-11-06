package config

import (
	"fmt"
	"strings"

	"net.kopias.oscbridge/app/pkg/slicetools"
)

// nolint:cyclop,revive,nolintlint
func validateConfig(cfg *MainConfig) error {
	err := validateAPPConfig(cfg)
	if err != nil {
		return err
	}
	return nil
}

func validateAPPConfig(cfg *MainConfig) error {
	oscImpls := []string{"l" /* "s" */}

	for _, cd := range cfg.OSCSources.ConsoleBridges {
		if slicetools.IndexOf(oscImpls, cd.OSCImplementation) == -1 {
			return fmt.Errorf("invalid osc implementation at %s: %s, valid values: %s", cd.Name, cd.OSCImplementation, strings.Join(oscImpls, ","))
		}
	}

	// @TODO add checks for connection-name integrity
	return nil
}
