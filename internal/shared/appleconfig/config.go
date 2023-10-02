package appleconfig

import (
	"fmt"
	"os"
)

const (
	authKeyPathEnv = "PRIVATE_KEY"
	keyIDEnv       = "KEY_ID"
	teamIDEnv      = "TEAM_ID"
	appBundleIDEnv = "APP_BUNDLE_ID"
)

type Config struct {
	clientID   string
	privateKey string
	keyID      string
	teamID     string
}

func New() (Config, error) {
	const missingConfigMsgFormat = "config field is missing in envs: %q"

	var config Config
	if config.clientID = os.Getenv(appBundleIDEnv); config.clientID == "" {
		return Config{}, fmt.Errorf(missingConfigMsgFormat, appBundleIDEnv)
	}
	if config.privateKey = os.Getenv(authKeyPathEnv); config.privateKey == "" {
		return Config{}, fmt.Errorf(missingConfigMsgFormat, authKeyPathEnv)
	}
	if config.keyID = os.Getenv(keyIDEnv); config.keyID == "" {
		return Config{}, fmt.Errorf(missingConfigMsgFormat, keyIDEnv)
	}
	if config.teamID = os.Getenv(teamIDEnv); config.teamID == "" {
		return Config{}, fmt.Errorf(missingConfigMsgFormat, teamIDEnv)
	}

	return config, nil
}

func (c Config) ClientID() string {
	return c.clientID
}

func (c Config) PrivateKey() string {
	return c.privateKey
}
func (c Config) KeyID() string {
	return c.keyID
}
func (c Config) TeamID() string {
	return c.teamID
}
