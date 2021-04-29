// Copyright 2021 Workrise Technologies Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package appconfig

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	archutil "github.com/rigup/ephemeral-iam/internal/appconfig/arch_util"
	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
)

const (
	AuthProxyAddress       = "authproxy.proxyaddress"
	AuthProxyPort          = "authproxy.proxyport"
	AuthProxyVerbose       = "authproxy.verbose"
	AuthProxyLogDir        = "authproxy.logdir"
	AuthProxyCertFile      = "authproxy.certfile"
	AuthProxyKeyFile       = "authproxy.keyfile"
	CloudSQLProxyPath      = "binarypaths.cloudsqlproxy"
	GcloudPath             = "binarypaths.gcloud"
	KubectlPath            = "binarypaths.kubectl"
	LoggingFormat          = "logging.format"
	LoggingLevel           = "logging.level"
	LoggingLevelTruncation = "logging.disableleveltruncation"
	LoggingPadLevelText    = "logging.padleveltext"
)

var (
	configDir string
	once      sync.Once

	binPaths = map[string]string{
		CloudSQLProxyPath: "cloud_sql_proxy",
		GcloudPath:        "gcloud",
		KubectlPath:       "kubectl",
	}
)

func InitConfig() error {
	viper.SetConfigName("config")
	viper.AddConfigPath(GetConfigDir())
	viper.AutomaticEnv()
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return errorsutil.EiamError{
				Log: logrus.StandardLogger().WithError(err),
				Msg: "Failed to initialize configuration",
				Err: err,
			}
		}
		initConfig()
	}

	// Instantiate logger now that the config is loaded.
	util.Logger = util.NewLogger()

	// Find the paths to gcloud, kubectl, and cloud_sql_proxy and write them to the config.
	if err := getBinPaths(); err != nil {
		return err
	}

	return nil
}

func initConfig() {
	viper.SetDefault(AuthProxyAddress, "127.0.0.1")
	viper.SetDefault(AuthProxyPort, "8084")
	viper.SetDefault(AuthProxyVerbose, false)
	viper.SetDefault(AuthProxyLogDir, filepath.Join(GetConfigDir(), "log"))
	viper.SetDefault(AuthProxyCertFile, filepath.Join(GetConfigDir(), "server.pem"))
	viper.SetDefault(AuthProxyKeyFile, filepath.Join(GetConfigDir(), "server.key"))
	viper.SetDefault(LoggingFormat, "text")
	viper.SetDefault(LoggingLevel, "info")
	viper.SetDefault(LoggingLevelTruncation, true)
	viper.SetDefault(LoggingPadLevelText, true)

	if err := viper.SafeWriteConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileAlreadyExistsError); !ok {
			log.Fatalf("failed to write config file %s/config.yml: %v", GetConfigDir(), err)
		}
	}
}

func getBinPaths() error {
	updated := false
	for configKey, binName := range binPaths {
		if viper.GetString(configKey) == "" {
			updated = true
			binPath, err := CheckCommandExists(binName)
			if err != nil {
				if configKey != CloudSQLProxyPath {
					// Exit if kubectl or gcloud aren't installed, but continue if
					// cloud_sql_proxy isn't.
					return err
				}
				util.Logger.Debug("Could not find path to cloud_sql_proxy binary")
			}
			viper.Set(configKey, binPath)
		}
	}

	if updated {
		if err := viper.WriteConfig(); err != nil {
			return errorsutil.EiamError{
				Log: util.Logger.WithError(err),
				Msg: "Failed to write binary paths to configuration file",
				Err: err,
			}
		}
	}
	return nil
}

// CheckCommandExists tries to find the location of a given binary.
func CheckCommandExists(command string) (string, error) {
	cmdPath, err := exec.LookPath(command)
	if err != nil {
		util.Logger.Debugf("Error while checking for %s binary", command)
		return "", err
	}
	util.Logger.Debugf("Found binary %s at %s\n", command, cmdPath)
	return cmdPath, nil
}

// GetConfigDir returns the directory to use for the ephemeral-iam configurations.
func GetConfigDir() string {
	once.Do(func() {
		dir, err := getConfigDir()
		errorsutil.CheckError(err)
		configDir = dir
	})
	return configDir
}

func getConfigDir() (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errorsutil.EiamError{
			Log: logrus.StandardLogger().WithError(err),
			Msg: "Failed to get user's home directory",
			Err: err,
		}
	}
	confPath := filepath.Join(userHomeDir, archutil.ConfigPath)
	if err = os.MkdirAll(confPath, 0o755); err != nil {
		return "", errorsutil.EiamError{
			Log: logrus.StandardLogger().WithError(err),
			Msg: fmt.Sprintf("Failed to create config directory: %s", confPath),
			Err: err,
		}
	}
	return confPath, nil
}