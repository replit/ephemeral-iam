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
	"strings"
	"text/tabwriter"

	"github.com/lithammer/dedent"
	"github.com/mitchellh/go-wordwrap"
	"github.com/spf13/cobra"
	"google.golang.org/api/iam/v1"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
	"github.com/rigup/ephemeral-iam/internal/gcpclient"
	"github.com/rigup/ephemeral-iam/pkg/options"
)

var listCmdConfig options.CmdConfig

func NewCmdListServiceAccounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "list-service-accounts",
		Aliases:    []string{"list"},
		Short:      "List service accounts that can be impersonated [alias: list]",
		SuggestFor: []string{"ls"},
		Long: dedent.Dedent(`
			The "list-service-accounts" command fetches all Cloud IAM Service Accounts in the current
			GCP project (as determined by the activated gcloud config) and checks each of them to see
			which ones the current user has access to impersonate.`),
		Example: dedent.Dedent(`
			$ eiam list-service-accounts
			$ eiam list`),
		PreRun: func(cmd *cobra.Command, args []string) {
			cmd.Flags().VisitAll(options.CheckRequired)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			availableSAs, err := gcpclient.FetchAvailableServiceAccounts(listCmdConfig.Project)
			if err != nil {
				return err
			}
			if len(availableSAs) == 0 {
				util.Logger.Warning("You do not have access to impersonate any accounts in this project")
				return nil
			}
			return printColumns(availableSAs)
		},
	}
	options.AddProjectFlag(cmd.Flags(), &listCmdConfig.Project)

	return cmd
}

func printColumns(serviceAccounts []*iam.ServiceAccount) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
	fmt.Fprintln(w, "\nEMAIL\tDESCRIPTION")
	for _, sa := range serviceAccounts {
		desc := strings.Split(wordwrap.WrapString(sa.Description, 75), "\n")
		if len(desc) == 1 {
			fmt.Fprintf(w, "%s\t%s\n", sa.Email, desc[0])
		} else {
			firstLine, remaining := desc[0], desc[1:]
			fmt.Fprintf(w, "%s\t%s\n", sa.Email, firstLine)
			for _, line := range remaining {
				fmt.Fprintf(w, "%s\t%s\n", " ", line)
			}
		}
	}
	if err := w.Flush(); err != nil {
		return errorsutil.New("Failed to flush tabwriter", err)
	}
	return nil
}
