// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package dynamo

import (
	"io"
	"time"
)

// Verbosity sets the verbosity level for the dynamo package.
func Verbosity(verbosity int) Option {
	return optionFunc(func(a *Dynamo) error {
		a.verbosity = verbosity
		return nil
	})
}

// Credentials sets the access and secret keys for the dynamo package.
func Credentials(accessKey, secretKey string) Option {
	return optionFunc(func(a *Dynamo) error {
		a.accessKey = accessKey
		a.secretKey = secretKey
		return nil
	})
}

// Region sets the region for the dynamo package.
func Region(region string) Option {
	return optionFunc(func(a *Dynamo) error {
		a.region = region
		return nil
	})
}

// Endpoint sets the endpoint for the dynamo package.
func Endpoint(endpoint string) Option {
	return optionFunc(func(a *Dynamo) error {
		a.endpoint = endpoint
		return nil
	})
}

// MaxWaitForDynamo sets the maximum time to wait for DynamoDB to be ready.
//
// The default is 10 seconds.
func MaxWaitForDynamo(max time.Duration) Option {
	return optionFunc(func(a *Dynamo) error {
		a.maxWaitForDynamo = max
		return nil
	})
}

// MaxDynamoResponseWait sets the maximum time to wait for a response from DynamoDB.
//
// The default is 20 microseconds.
func MaxDynamoResponseWait(max time.Duration) Option {
	return optionFunc(func(a *Dynamo) error {
		a.maxDynamoResponseWait = max
		return nil
	})
}

// MaxDynamoActionWait sets the maximum time to wait for an action to complete.
//
// The default is 5 seconds.
func MaxDynamoActionWait(max time.Duration) Option {
	return optionFunc(func(a *Dynamo) error {
		a.maxDynamoActionWait = max
		return nil
	})
}

// HumanTableName sets the human readable table name.
func HumanTableName(humanTableName string) Option {
	return optionFunc(func(a *Dynamo) error {
		a.humanTableName = humanTableName
		return nil
	})
}

// GeneralDelay sets the general delay between actions.
//
// The default is 10 milliseconds.
func GeneralDelay(generalDelay time.Duration) Option {
	return optionFunc(func(a *Dynamo) error {
		a.generalDelay = generalDelay
		return nil
	})
}

// Stdout sets the writer for stdout.
func Stdout(stdout io.Writer) Option {
	return optionFunc(func(a *Dynamo) error {
		if stdout != nil {
			a.stdout = stdout
		}
		return nil
	})
}

func validateCredentials() Option {
	return optionFunc(func(a *Dynamo) error {
		if a.accessKey == "" || a.secretKey == "" {
			a.logf(1, "warning: credentials not set\n")
		}
		return nil
	})
}

func validateRegion() Option {
	return optionFunc(func(a *Dynamo) error {
		if a.region == "" {
			a.logf(1, "warning: region not set\n")
		}
		return nil
	})
}

func validateEndpoint() Option {
	return optionFunc(func(a *Dynamo) error {
		if a.endpoint == "" {
			a.logf(1, "warning: endpoint not set\n")
		}
		return nil
	})
}

func validateHumanTableName() Option {
	return optionFunc(func(a *Dynamo) error {
		if a.humanTableName == "" {
			a.logf(1, "warning: human table name not set\n")
		}
		return nil
	})
}
