/*
Copyright © 2021 Jesse Somerville <jssomerville2@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package main

import (
	"emperror.dev/emperror"

	"github.com/jessesomerville/gcp-iam-escalate/appconfig"
	"github.com/jessesomerville/gcp-iam-escalate/cmd"
	"github.com/jessesomerville/gcp-iam-escalate/errorhandler"
	"github.com/jessesomerville/gcp-iam-escalate/gcpclient"
	"github.com/jessesomerville/gcp-iam-escalate/loghandler"
)

func main() {
	config := &appconfig.Config

	logger := loghandler.GetLogger(&config.Logging)
	errorHandler := errorhandler.GetErrorHandler(logger)
	credentialsClient := gcpclient.GetGCPClient()

	defer emperror.HandleRecover(errorHandler)
	defer credentialsClient.Close()

	err := cmd.Execute()
	emperror.Panic(err)
}