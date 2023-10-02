package applesignin

import (
	"errors"
	"os"
)

type config struct {
	privateKey string
	keyID      string
	teamID     string
	bundleID   string
}

var errMissingConfiguration = errors.New("missing required configurations env variable")

func newConfig() (config, error) {
	const (
		privateKeyEnv  = "PRIVATE_KEY"
		teamIDKeyEnv   = "TEAM_ID"
		keyIDKeyEnv    = "KEY_ID"
		appBundleIDEnv = "APP_BUNDLE_ID"
	)

	privateKey := os.Getenv(privateKeyEnv)
	if privateKey == "" {
		return config{}, errMissingConfiguration
	}

	teamID := os.Getenv(teamIDKeyEnv)
	if teamID == "" {
		return config{}, errMissingConfiguration
	}

	keyID := os.Getenv(keyIDKeyEnv)
	if keyID == "" {
		return config{}, errMissingConfiguration
	}

	bundleID := os.Getenv(appBundleIDEnv)
	if bundleID == "" {
		return config{}, errMissingConfiguration
	}

	return config{
		privateKey: privateKey,
		keyID:      keyID,
		teamID:     teamID,
		bundleID:   bundleID,
	}, nil
}
