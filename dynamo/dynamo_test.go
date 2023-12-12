// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package dynamo_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/argus-dynamodb/dynamo"
	"github.com/xmidt-org/idock"
)

var argusTable = dynamodb.CreateTableInput{
	TableName: aws.String("dynamo-tests"),
	AttributeDefinitions: []*dynamodb.AttributeDefinition{
		{
			AttributeName: aws.String("bucket"),
			AttributeType: aws.String(dynamodb.ScalarAttributeTypeS),
		}, {
			AttributeName: aws.String("expires"),
			AttributeType: aws.String(dynamodb.ScalarAttributeTypeN),
		}, {
			AttributeName: aws.String("id"),
			AttributeType: aws.String(dynamodb.ScalarAttributeTypeS),
		},
	},
	KeySchema: []*dynamodb.KeySchemaElement{
		{
			AttributeName: aws.String("bucket"),
			KeyType:       aws.String(dynamodb.KeyTypeHash),
		}, {
			AttributeName: aws.String("id"),
			KeyType:       aws.String(dynamodb.KeyTypeRange),
		},
	},
	GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
		{
			IndexName: aws.String("Expires-index"),
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String("bucket"),
					KeyType:       aws.String(dynamodb.KeyTypeHash),
				}, {
					AttributeName: aws.String("expires"),
					KeyType:       aws.String(dynamodb.KeyTypeRange),
				},
			},
			Projection: &dynamodb.Projection{
				ProjectionType: aws.String(dynamodb.ProjectionTypeAll),
			},
			ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(10),
				WriteCapacityUnits: aws.Int64(5),
			},
		},
	},
	ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
		ReadCapacityUnits:  aws.Int64(10),
		WriteCapacityUnits: aws.Int64(5),
	},
	StreamSpecification: &dynamodb.StreamSpecification{
		StreamEnabled:  aws.Bool(true),
		StreamViewType: aws.String(dynamodb.StreamViewTypeNewAndOldImages),
	},
}

func TestEnd2End(t *testing.T) {
	tests := []struct {
		description string
		opts        []dynamo.Option
	}{
		{
			description: "base",
		}, {
			description: "second time",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			opts := []dynamo.Option{
				dynamo.Credentials("accesKey", "secretKey"),
				dynamo.Region("local"),
				dynamo.Endpoint("http://localhost:7805"),
				dynamo.Verbosity(99),
				dynamo.HumanTableName("dynamo-tests"),
			}

			opts = append(opts, tc.opts...)

			a, err := dynamo.New(opts...)
			assert.NoError(err)
			if !assert.NotNil(a) {
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			err = a.Create(ctx, &argusTable)
			assert.NoError(err)
		})
	}
}

func TestWarnings(t *testing.T) {
	tests := []struct {
		description string
		opts        []dynamo.Option
		expect      []string
	}{
		{
			description: "no credentials",
			opts: []dynamo.Option{
				dynamo.Region("local"),
				dynamo.Endpoint("http://localhost:7805"),
				dynamo.Verbosity(99),
				dynamo.HumanTableName("dynamo-tests"),
			},
			expect: []string{
				"dynamo-tests DynamoDB: warning: credentials not set\n",
			},
		}, {
			description: "no region",
			opts: []dynamo.Option{
				dynamo.Credentials("accesKey", "secretKey"),
				dynamo.Endpoint("http://localhost:7805"),
				dynamo.Verbosity(99),
				dynamo.HumanTableName("dynamo-tests"),
			},
			expect: []string{
				"dynamo-tests DynamoDB: warning: region not set\n",
			},
		}, {
			description: "no endpoint",
			opts: []dynamo.Option{
				dynamo.Credentials("accesKey", "secretKey"),
				dynamo.Region("local"),
				dynamo.Verbosity(99),
				dynamo.HumanTableName("dynamo-tests"),
			},
			expect: []string{
				"dynamo-tests DynamoDB: warning: endpoint not set\n",
			},
		}, {
			description: "no human table name",
			opts: []dynamo.Option{
				dynamo.Credentials("accesKey", "secretKey"),
				dynamo.Region("local"),
				dynamo.Endpoint("http://localhost:7805"),
				dynamo.Verbosity(99),
			},
			expect: []string{
				" DynamoDB: warning: human table name not set\n",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			var buf strings.Builder

			opts := []dynamo.Option{
				dynamo.Stdout(&buf),
			}

			opts = append(opts, tc.opts...)

			a, err := dynamo.New(opts...)
			assert.NoError(err)
			if !assert.NotNil(a) {
				return
			}

			s := strings.Join(tc.expect, "")

			assert.Equal(s, buf.String())
		})
	}
}

func TestMain(m *testing.M) {
	infra := idock.New(
		idock.DockerComposeFile("docker-compose.yml"),
		idock.RequireDockerTCPPorts(7805),
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
