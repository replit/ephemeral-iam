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

//go:build linux
// +build linux

package archutil

const (
	// FormattedOS is the string representation of the GOOS used in
	// ephemeral-iam release tarballs.
	FormattedOS = "Linux"
	// ConfigPath is the path relative to the users home directory to store the
	// ephemeral-iam config.
	ConfigPath = ".config/ephemeral-iam"
)
