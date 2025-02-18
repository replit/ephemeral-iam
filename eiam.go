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

package main

import (
	"github.com/replit/ephemeral-iam/cmd/eiam"                   //nolint: depguard
	"github.com/replit/ephemeral-iam/internal/appconfig"         //nolint: depguard
	errorsutil "github.com/replit/ephemeral-iam/internal/errors" //nolint: depguard
)

func main() {
	errorsutil.CheckError(appconfig.InitConfig())
	errorsutil.CheckError(appconfig.Setup())

	if appconfig.Version != "v0.0.0" {
		appconfig.CheckForNewRelease()
	}

	rootCmd, err := eiam.NewEphemeralIamCommand()
	errorsutil.CheckError(err)

	// Kill the loaded plugin clients. This is happening here to ensure that
	// Kill is called after the command has finished running, but also accounts
	// for any errors that occur during execution.
	for _, plugin := range rootCmd.Plugins {
		defer plugin.Client.Kill() //nolint: gocritic
	}
	errorsutil.CheckError(rootCmd.Execute())
}
