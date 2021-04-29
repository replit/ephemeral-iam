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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
)

// EphemeralIamPlugin represents an ephemeral-iam command plugin.
type EphemeralIamPlugin struct {
	*cobra.Command

	// Name is the canonical name of the plugin.
	Name string
	// Desc is the description of the plugin.
	Desc string
	// Version is the version of the plugin.
	Version string
	// Path is the path that the plugin was loaded from.
	Path string
}

// Logger allows plugins to fetch an instance of the ephemeral-iam logger.
func Logger() *logrus.Logger {
	return util.Logger
}
