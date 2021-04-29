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

package gcpclient

import (
	"context"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"google.golang.org/api/option"

	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
)

// ClientWithReason creates a client SDK with the provided reason field.
func ClientWithReason(reason string) (*credentials.IamCredentialsClient, error) {
	ctx := context.Background()
	gcpClientWithReason, err := credentials.NewIamCredentialsClient(ctx, option.WithRequestReason(reason))
	if err != nil {
		return nil, &errorsutil.SDKClientCreateError{Err: err, ResourceType: "Credentials"}
	}
	return gcpClientWithReason, nil
}