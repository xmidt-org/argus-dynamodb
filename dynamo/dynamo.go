// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package dynamo

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Dynamo struct {
	accessKey string
	secretKey string
	region    string
	endpoint  string

	humanTableName string

	maxWaitForDynamo      time.Duration
	maxDynamoResponseWait time.Duration
	maxDynamoActionWait   time.Duration
	generalDelay          time.Duration

	stdout io.Writer

	verbosity int
	svc       *dynamodb.DynamoDB
}

type Option interface {
	apply(*Dynamo) error
}

type optionFunc func(*Dynamo) error

func (o optionFunc) apply(a *Dynamo) error {
	return o(a)
}

// New creates a new ArgusDynamodb instance
func New(opts ...Option) (*Dynamo, error) {
	var d Dynamo

	defaults := []Option{
		Stdout(os.Stdout),
		MaxWaitForDynamo(10 * time.Second),
		MaxDynamoResponseWait(20 * time.Microsecond),
		MaxDynamoActionWait(5 * time.Second),
		GeneralDelay(10 * time.Millisecond),
	}

	opts = append(defaults, opts...)

	opts = append(opts,
		validateCredentials(),
		validateRegion(),
		validateEndpoint(),
		validateHumanTableName(),
	)

	for _, opt := range opts {
		err := opt.apply(&d)
		if err != nil {
			return nil, err
		}
	}
	return &d, nil
}

// Create creates a new DynamoDB table, or replaces an existing table.
func (d *Dynamo) Create(ctx context.Context, table *dynamodb.CreateTableInput) error {
	start := time.Now()
	if table == nil || table.TableName == nil {
		return fmt.Errorf("table name is required")
	}

	d.logf(1, "setup starting.\n")
	err := d.createSvc(ctx)
	if err != nil {
		d.logf(1, "setup failed.\n")
		return d.wrap(err)
	}

	d.logf(1, "waiting for dynamo to respond.\n")
	err = d.waitForDynamodb(ctx)
	if err != nil {
		return d.wrap(err)
	}
	d.logf(1, "dynamo is ready.\n")

	err = d.truncateTable(ctx, *table.TableName)
	if err != nil {
		return d.wrap(err)
	}

	d.logf(1, "waiting for table to be deleted.\n")

	err = d.confirmTableAbsent(ctx, *table.TableName)
	if err != nil {
		return d.wrap(err)
	}

	d.logf(1, "create table.\n")
	err = d.createTable(ctx, table)
	if err != nil {
		return d.wrap(err)
	}

	d.logf(1, "waiting for table to be confirmed.\n")

	err = d.confirmTablePresent(ctx, *table.TableName)
	if err != nil {
		return d.wrap(err)
	}

	d.logf(1, "setup complete.  Elapsed time: %s\n", time.Since(start))

	return nil
}

func (d *Dynamo) createSvc(ctx context.Context) error {
	// Create a new AWS session
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(d.accessKey, d.secretKey, ""),
		Endpoint:    aws.String(d.endpoint),
		Region:      aws.String(d.region),
	}))

	// Create a DynamoDB service client
	svc := dynamodb.New(sess)
	if svc == nil {
		return fmt.Errorf("failed to create service client")
	}
	d.svc = svc

	return nil
}

func (d *Dynamo) createTable(ctx context.Context, input *dynamodb.CreateTableInput) error {
	for {
		_, err := d.svc.CreateTableWithContext(ctx, input)
		if err == nil {
			d.logf(1, "created table.\n")
			return nil
		}

		if ctx.Err() != nil {
			return fmt.Errorf("failed to create table")
		}

		// Don't spin too fast.
		time.Sleep(d.generalDelay)
	}
}

func (d *Dynamo) confirmTablePresent(ctx context.Context, tableName string) error {
	return d.confirmTable(ctx, tableName, true)
}

func (d *Dynamo) confirmTableAbsent(ctx context.Context, tableName string) error {
	return d.confirmTable(ctx, tableName, false)
}

func (d *Dynamo) confirmTable(ctx context.Context, tableName string, present bool) error {
	for {
		got, _ := d.svc.DescribeTableWithContext(ctx, &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})

		if present {
			if got != nil && got.Table != nil {
				d.logf(1, "table successfully confirmed.\n")
				return nil
			}
		} else {
			if got == nil || got.Table == nil {
				d.logf(1, "table not present.\n")
				return nil
			}
		}

		if ctx.Err() != nil {
			d.logf(1, "table state not confirmed before time ran out.\n")
			return ctx.Err()
		}

		// Don't spin too fast.
		time.Sleep(d.generalDelay)
	}
}

func (d *Dynamo) waitForDynamodb(ctx context.Context) error {
	for {
		short, cancel := context.WithTimeout(ctx, 50*time.Millisecond)

		got, _ := d.svc.ListTablesWithContext(short, &dynamodb.ListTablesInput{})
		cancel()
		if got != nil {
			d.logf(1, "db is ready.\n")
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Don't spin too fast.
		time.Sleep(d.generalDelay)
	}
}

func (d *Dynamo) truncateTable(ctx context.Context, name string) error {
	// Describe the table
	describeInput := &dynamodb.DescribeTableInput{
		TableName: aws.String(name),
	}
	got, _ := d.svc.DescribeTableWithContext(ctx, describeInput)
	if got != nil && got.Table != nil {
		d.logf(1, "existing table found.\n")
		d.logf(1, "removing existing table.\n")
		return d.deleteTable(ctx, name)
	}

	d.logf(1, "no existing table found.\n")
	return nil
}

func (d *Dynamo) deleteTable(ctx context.Context, name string) error {
	// Delete the Kinesis stream
	_, err := d.svc.DeleteTableWithContext(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(name),
	})
	if err != nil {
		return err
	}

	d.logf(1, "deleted table.\n")

	for {
		got, _ := d.svc.DescribeTableWithContext(ctx, &dynamodb.DescribeTableInput{
			TableName: aws.String(name),
		})

		if got == nil || got.Table == nil {
			d.logf(1, "table deleted successfully.\n")
			return nil
		}

		if ctx.Err() != nil {
			return fmt.Errorf("failed to delete existing table")
		}

		// Don't spin too fast.
		time.Sleep(d.generalDelay)
	}
}

// logf prints a message if the verbosity level is greater than or equal to the
// given level.
func (d *Dynamo) logf(level int, format string, a ...any) {
	if d.verbosity >= level {
		var buf strings.Builder
		buf.WriteString(d.humanTableName + " DynamoDB: ")
		fmt.Fprintf(&buf, format, a...)

		_, _ = d.stdout.Write([]byte(buf.String()))
	}
}

func (d *Dynamo) wrap(err error) error {
	return fmt.Errorf("%s DynamoDB: %w", d.humanTableName, err)
}
