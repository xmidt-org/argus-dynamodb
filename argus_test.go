// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package argusdynamodb_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	argusdynamodb "github.com/xmidt-org/argus-dynamodb"
	"github.com/xmidt-org/argus-dynamodb/dynamo"
	"github.com/xmidt-org/idock"
)

func TestEnd2End(t *testing.T) {
	for i := 0; i < 2; i++ {
		t.Run(fmt.Sprintf("run %d", i), func(t *testing.T) {
			assert := assert.New(t)

			a, err := argusdynamodb.New(
				dynamo.Credentials("accesKey", "secretKey"),
				dynamo.Region("local"),
				dynamo.Endpoint("http://localhost:7804"),
				dynamo.Verbosity(99),
			)
			assert.NoError(err)
			if !assert.NotNil(a) {
				return
			}

			err = a.Create(context.Background())
			assert.NoError(err)
		})
	}
}

func TestMain(m *testing.M) {
	infra := idock.New(
		idock.DockerComposeFile("docker-compose.yml"),
		idock.RequireDockerTCPPorts(7800, 7801, 7802, 7803, 7804),
	)

	err := infra.Start()
	if err != nil {
		panic(err)
	}

	returnCode := m.Run()

	infra.Stop()

	if returnCode != 0 {
		os.Exit(returnCode)
	}
}
