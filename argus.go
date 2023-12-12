// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package argusdynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/xmidt-org/argus-dynamodb/dynamo"
)

type ArgusDynamodb struct {
	db *dynamo.Dynamo
}

// New creates a new ArgusDynamodb instance
func New(opts ...dynamo.Option) *ArgusDynamodb {
	var ad ArgusDynamodb

	opts = append(opts, dynamo.HumanTableName("Argus"))
	ad.db = dynamo.New(opts...)

	if ad.db == nil {
		return nil
	}

	return &ad
}

func (ad *ArgusDynamodb) Create(ctx context.Context) error {
	return ad.db.Create(ctx,
		&dynamodb.CreateTableInput{
			TableName: aws.String("gifnoc"),
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
		})
}
