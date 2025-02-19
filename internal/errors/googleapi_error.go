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

package errors

import (
	"errors"
	"fmt"

	"google.golang.org/api/googleapi"
)

// See https://cloud.google.com/apis/design/errors#handling_errors
var googleErrorCodes = map[int]string{
	400: "Invalid argument",
	401: "Invalid authentication credentials",
	403: "Permission denied",
	404: "Resource not found",
	409: "Resource conflict",
	429: "Quota limit exceeded",
	499: "Request canceled by client",
	500: "Internal server error",
	501: "Unimplemented method",
	503: "Server unavailable",
	504: "Server deadline exceeded",
}

func checkGoogleAPIError(err error) EiamError {
	if serr, ok := err.(EiamError); ok {
		err = serr.Err
	}
	if gerr, ok := err.(*googleapi.Error); ok {
		errStatusMsg, ok := googleErrorCodes[gerr.Code]
		if !ok {
			errStatusMsg = "Unknown error"
		}
		errMsg := gerr.Message
		if errMsg == "" {
			// TODO Check if message can be parsed from body.
			errMsg = gerr.Body
		}
		return New(fmt.Sprintf("[Google API Error] %s", errStatusMsg), errors.New(errMsg)).(EiamError) //nolint: errcheck
	}
	return EiamError{}
}
