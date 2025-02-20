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
	"fmt"
	"os"
	"regexp"

	"github.com/lithammer/dedent"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	util "github.com/replit/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/replit/ephemeral-iam/internal/errors"
	"github.com/replit/ephemeral-iam/internal/plugins"
)

func newCmdPlugins() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugins",
		Short: "Manage ephemeral-iam plugins",
		Long: dedent.Dedent(`
			Plugins for ephemeral-iam allow you to extend eiam's functionality in the form of commands.
			Plugins are stored in the 'plugins' directory of your eiam configuration folder.
			
			-------------------------------     Installing a plugin     -------------------------------
			Plugins are loaded from the '/path/to/config/ephemeral-iam/plugins' directory. To install a
			plugin, you just place the plugin's binary in that directory and eiam will automatically
			discover and load it.
			
			If the plugin you want to install is hosted in a Github repo and the binary is published as
			a release in the repository, you can install the plugin using the 'eiam plugin install'
			command.
		`),
	}

	cmd.AddCommand(newCmdPluginsList())
	cmd.AddCommand(newCmdPluginsInstall())
	cmd.AddCommand(newCmdPluginsRemove())
	cmd.AddCommand(newCmdPluginsAuth())
	return cmd
}

func newCmdPluginsList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show the list of loaded plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(RootCommand.Plugins) == 0 {
				util.Logger.Warn("No plugins are currently installed")
				return nil
			}
			RootCommand.PrintPlugins()
			return nil
		},
	}
	return cmd
}

func newCmdPluginsInstall() *cobra.Command {
	var (
		url       string
		tokenName string
		repoOwner string
		repoName  string
	)
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install a new eiam plugin",
		Long: dedent.Dedent(`
			The "plugins install" command installs a plugin from a Github repository.

			The latest release in the provided repository is downloaded, extracted, and
			the binary files are moved to the "plugins" directory.

			If the plugin is hosted in a private repository, you need to provide
			ephemeral-iam with a Github personal access token to authenticate
			with. See 'eiam plugins auth --help' for more details.
		`),
		Args: func(cmd *cobra.Command, args []string) error {
			urlRegex := regexp.MustCompile(`^github\.com/(?P<user>[[:alnum:]\-]+)/(?P<repo>[[:alnum:]\.\-_]+)$`)
			match := urlRegex.FindStringSubmatch(url)
			if match == nil {
				err := fmt.Errorf("%s is not a valid Github repo URL", url)
				return errorsutil.New("Invalid input parameter", err)
			}
			for i, grpName := range urlRegex.SubexpNames() {
				if grpName == "user" {
					repoOwner = match[i]
				} else if grpName == "repo" {
					repoName = match[i]
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return plugins.InstallPlugin(repoOwner, repoName, tokenName)
		},
	}
	cmd.Flags().StringVarP(&url, "url", "u", "", "The URL of the plugin's Github repo")
	if err := cmd.MarkFlagRequired("url"); err != nil {
		util.Logger.Fatal(err.Error())
	}
	cmd.Flags().StringVarP(&tokenName, "token", "t", "", "The name of the Github access token to use for private repos")
	return cmd
}

func newCmdPluginsRemove() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove an installed eiam plugin",
		Long: dedent.Dedent(`
			The "plugins remove" command removes a currently installed plugin.
			
			You will be prompted to select the plugin to uninstall from the list of plugins loaded
			by eiam. If no plugins are currently installed, a warning is shown.`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(RootCommand.Plugins) == 0 {
				util.Logger.Warn("No plugins are currently installed")
				return nil
			}

			plugin, err := selectPlugin()
			if err != nil {
				return err
			}

			if err := os.Remove(plugin.Path); err != nil {
				return errorsutil.New(fmt.Sprintf("Failed to remove plugin file %s", plugin.Path), err)
			}
			util.Logger.Infof("Successfully removed %s", plugin.Name)
			return nil
		},
	}
	return cmd
}

func selectPlugin() (*plugins.EphemeralIamPlugin, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   " ►  {{ .Name | red }}",
		Inactive: "  {{ .Name | red }}",
		Selected: " ►  {{ .Name | red | cyan }}",
		Details: `
--------- Plugin ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Description:" | faint }}	{{ .Description }}`,
	}

	prompt := promptui.Select{
		Label:     "Plugin to remove",
		Items:     RootCommand.Plugins,
		Templates: templates,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return nil, errorsutil.New("Select-plugin prompt failed", err)
	}
	return RootCommand.Plugins[i], nil
}
