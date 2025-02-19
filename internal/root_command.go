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

package eiamplugin

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"text/tabwriter"

	hcplugin "github.com/hashicorp/go-plugin"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/status"

	"github.com/replit/ephemeral-iam/internal/appconfig"
	util "github.com/replit/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/replit/ephemeral-iam/internal/errors"
	"github.com/replit/ephemeral-iam/internal/plugins"
	eiamplugin "github.com/replit/ephemeral-iam/pkg/plugins"
)

// RootCommand is a struct that holds the loaded plugins and the top level cobra command.
type RootCommand struct {
	Plugins []*plugins.EphemeralIamPlugin
	cobra.Command
}

// LoadPlugins searches for files in the plugin directory and attempts to load them.
func (rc *RootCommand) LoadPlugins() error {
	configDir := appconfig.GetConfigDir()
	pluginsDir := path.Join(configDir, "plugins")

	files, err := os.ReadDir(pluginsDir)
	if err != nil {
		return errorsutil.New("Failed to read plugins directory", err)
	}

	for _, f := range files {
		pl, plClient, err := loadPlugin(f.Name(), pluginsDir)
		if err != nil {
			util.Logger.WithError(err).Errorf("Failed to load plugin: %s", f.Name())
			continue
		}
		pluginCmd, name, desc, version, err := addPluginCmd(pl)
		if err != nil {
			return err
		}
		rc.AddCommand(pluginCmd)
		rc.Plugins = append(rc.Plugins, &plugins.EphemeralIamPlugin{
			Name:        name,
			Description: desc,
			Version:     version,
			Client:      plClient,
			Path:        path.Join(pluginsDir, f.Name()),
		})
	}
	return nil
}

func loadPlugin(pf, pluginsDir string) (plugins.EIAMPlugin, *hcplugin.Client, error) {
	args := []string{}
	if len(os.Args) >= 2 {
		args = os.Args[2:]
	}
	client := hcplugin.NewClient(&hcplugin.ClientConfig{
		HandshakeConfig: eiamplugin.Handshake,
		Plugins: map[string]hcplugin.Plugin{
			"run-command": &eiamplugin.Command{},
		},
		Cmd:              exec.Command(path.Join(pluginsDir, pf), args...), //nolint:gosec // Single string with no args
		AllowedProtocols: []hcplugin.Protocol{hcplugin.ProtocolGRPC},
		SyncStderr:       os.Stderr,
		SyncStdout:       os.Stdout,
		Logger:           plugins.NewHCLogAdapter(util.Logger, ""),
		// AutoMTLS:         true,  // For some reason, enabling this breaks the plugin commands.
	})

	rpcClient, err := client.Client()
	if err != nil {
		return nil, nil, err
	}

	raw, err := rpcClient.Dispense("run-command")
	if err != nil {
		return nil, nil, err
	}
	return raw.(plugins.EIAMPlugin), client, nil //nolint: errcheck
}

func addPluginCmd(p plugins.EIAMPlugin) (cmd *cobra.Command, name, desc, version string, err error) {
	name, desc, version, err = p.GetInfo()
	if err != nil {
		return nil, "", "", "", errorsutil.New("Failed to fetch plugin information", err)
	}

	cmd = &cobra.Command{
		Use:                name,
		Short:              fmt.Sprintf("%s %s: %s", name, version, desc),
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.Run(); err != nil {
				if serr, ok := status.FromError(err); ok {
					return errors.New(serr.Message())
				}
				return err
			}
			return nil
		},
	}
	return cmd, name, desc, version, nil
}

// PrintPlugins formats the list of loaded plugins as a table and prints them.
func (rc *RootCommand) PrintPlugins() {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 4, 4, ' ', 0)
	fmt.Fprintln(w, "\nPLUGIN\tVERSION\tDESCRIPTION")
	for _, p := range rc.Plugins {
		fmt.Fprintf(w, "%s\t%s\t%s\n", p.Name, p.Version, p.Description)
	}
	fmt.Fprintln(w)
	w.Flush()

	fmt.Println(buf.String())
}
