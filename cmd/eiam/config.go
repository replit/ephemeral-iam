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

package eiam

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/replit/ephemeral-iam/internal/appconfig"
	util "github.com/replit/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/replit/ephemeral-iam/internal/errors"
)

var (
	loggingLevels    = []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	loggingFormats   = []string{"text", "json", "debug"}
	boolConfigFields = []string{
		appconfig.AuthProxyVerbose,
		appconfig.GithubAuth,
		appconfig.LoggingLevelTruncation,
		appconfig.LoggingPadLevelText,
	}
)

var configInfo = dedent.Dedent(`
		┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
		┃ Key                            ┃ Description                                 ┃
		┡━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┩
		│ authproxy.certfile             │ The path to the auth proxy's TLS            │
		│                                │ certificate                                 │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ authproxy.keyfile              │ The path to the auth proxy's x509 key       │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ authproxy.logdir               │ The directory that auth proxy logs will be  │
		│                                │ written to                                  │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ authproxy.proxyaddress         │ The address that the auth proxy is hosted   │
		│                                │ on                                          │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ authproxy.proxyport            │ The port that the auth proxy runs on        │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ authproxy.verbose              │ When set to 'true', verbose output for      │
		│                                │ proxy logs will be enabled                  │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ binarypaths.cloudsqlproxy      │ The path to the cloud_sql_proxy binary on   │
		│                                │ your filesystem                             │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ binarypaths.gcloud             │ The path to the gcloud binary on your       │
		│                                │ filesystem                                  │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ binarypaths.kubectl            │ The path to the kubectl binary on your      │
		│                                │ filesystem                                  │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ github.auth                    │ When set to 'true', the "plugins install"   │
		│                                │ command will use a configured personal      │
		│                                │ access token to authenticate to the Github  │
		│                                │ API.                                        │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ github.tokens                  │ The configured Github personal access       │
		│                                │ tokens                                      │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ logging.format                 │ The format for which to write console logs  │
		│                                │ Can be 'json', 'text', or 'debug'           │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ logging.level                  │ The logging level to write to the console   │
		│                                │ Can be one of 'trace', 'debug', 'info',     │
		│                                │ 'warn', 'error', 'fatal', or 'panic'        │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ logging.disableleveltruncation │ When set to 'true', the level indicator for │
		│                                │ logs will not be trucated                   │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ logging.padleveltext           │ When set to 'true', output logs will align  │
		│                                │ evenly with their output level indicator    │
		├────────────────────────────────┼─────────────────────────────────────────────┤
		│ serviceaccounts                │ The default service accounts set via the    │
		│                                │ 'default-service-accounts' command          │
		└────────────────────────────────┴─────────────────────────────────────────────┘
`)

func newCmdConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration values",
	}

	cmd.AddCommand(newCmdConfigPrint())
	cmd.AddCommand(newCmdConfigView())
	cmd.AddCommand(newCmdConfigSet())
	cmd.AddCommand(newCmdConfigInfo())

	return cmd
}

func newCmdConfigPrint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print",
		Short: "Print the current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := viper.ConfigFileUsed()
			data, err := ioutil.ReadFile(configFile)
			if err != nil {
				return errorsutil.New("Failed to read configuration file", err)
			}
			lines := strings.Split(string(data), "\n")
			inTokens := false
			fmt.Printf("\n")
			for i := 0; i < len(lines); i++ {
				if lines[i] == "  tokens:" {
					inTokens = true
				} else if inTokens {
					// Mask Github personal access tokens.
					tokenConfig := viper.GetStringMapString(appconfig.GithubTokens)
					for tokenName := range tokenConfig {
						fmt.Printf("    %s: **********\n", tokenName)
						i++
					}
					i--
					inTokens = false
					continue
				}
				fmt.Println(lines[i])
			}
			fmt.Printf("\n")
			return nil
		},
	}
	return cmd
}

func newCmdConfigInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Print information about config fields",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(configInfo)
		},
	}
	return cmd
}

func newCmdConfigView() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "view",
		Short:     "View the value of a provided config item",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: viper.AllKeys(),
		Run: func(cmd *cobra.Command, args []string) {
			val := viper.Get(args[0])
			util.Logger.Infof("%s: %v\n", args[0], val)
		},
	}
	return cmd
}

func newCmdConfigSet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set the value of a provided config item",
		Args:  checkSetArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			oldVal := viper.Get(args[0])

			if oldVal == args[1] {
				util.Logger.Warn("New value is the same as the current one")
				return nil
			}
			if util.Contains(boolConfigFields, args[0]) {
				newValue, err := strconv.ParseBool(args[1])
				if err != nil {
					return argsError(fmt.Errorf("the %s value must be either true or false", args[0]))
				}
				viper.Set(args[0], newValue)
			} else {
				viper.Set(args[0], args[1])
			}
			// Update the logger (for testing).
			switch args[0] {
			case appconfig.LoggingLevel:
				level, err := logrus.ParseLevel(args[1])
				if err != nil {
					return argsError(err)
				}
				util.Logger.Level = level

			case appconfig.LoggingFormat:
				switch args[1] {
				case "debug":
					util.Logger.Formatter = util.NewRuntimeFormatter()
				case "json":
					util.Logger.Formatter = util.NewJSONFormatter()
				default:
					util.Logger.Formatter = util.NewTextFormatter()
				}
			}
			if err := viper.WriteConfig(); err != nil {
				return errorsutil.New("Failed to write updated configuration", err)
			}
			util.Logger.Infof("Updated %s from %v to %s", args[0], oldVal, args[1])
			return nil
		},
	}
	return cmd
}

func checkSetArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return argsError(errors.New("requires both a config key and a new value"))
	}

	switch args[0] {
	case appconfig.LoggingLevel:
		if !util.Contains(loggingLevels, args[1]) {
			return argsError(fmt.Errorf("logging level must be one of %v", loggingLevels))
		}
		return nil
	case appconfig.LoggingFormat:
		if !util.Contains(loggingFormats, args[1]) {
			return argsError(fmt.Errorf("logging format must be one of %v", loggingFormats))
		}
		return nil
	case appconfig.GithubTokens:
		return errors.New("please use the 'plugins auth' commands to edit configured Github access tokens")
	case appconfig.DefaultServiceAccounts:
		return errors.New("please use the 'default-service-accounts' commands to edit configured default service accounts")
	}

	if util.Contains(boolConfigFields, args[0]) {
		if _, err := strconv.ParseBool(args[1]); err != nil {
			return argsError(fmt.Errorf("the %s value must be either true or false", args[0]))
		}
		return nil
	}

	if !util.Contains(viper.AllKeys(), args[0]) {
		return argsError(fmt.Errorf("invalid config key %s", args[0]))
	}

	return nil
}

func argsError(err error) error {
	return errorsutil.New("Invalid command arguments", err)
}
