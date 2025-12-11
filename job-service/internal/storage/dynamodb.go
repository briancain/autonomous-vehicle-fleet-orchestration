package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBAPI interface for mocking
type DynamoDBAPI interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

type DynamoDBJobStorage struct {
	client    DynamoDBAPI
	tableName string
}

func NewDynamoDBJobStorage(client DynamoDBAPI, tableName string) *DynamoDBJobStorage {
	return &DynamoDBJobStorage{
		client:    client,
		tableName: tableName,
	}
}

func (d *DynamoDBJobStorage) CreateJob(ctx context.Context, job *Job) error {
	item, err := attributevalue.MarshalMap(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to put job: %w", err)
	}

	return nil
}

func (d *DynamoDBJobStorage) GetJob(ctx context.Context, jobID string) (*Job, error) {
	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: jobID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("job %s not found", jobID)
	}

	var job Job
	err = attributevalue.UnmarshalMap(result.Item, &job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

func (d *DynamoDBJobStorage) UpdateJob(ctx context.Context, job *Job) error {
	item, err := attributevalue.MarshalMap(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

func (d *DynamoDBJobStorage) UpdateJobStatus(ctx context.Context, jobID, status string, vehicleID *string) error {
	updateExpression := "SET #status = :status"
	expressionAttributeValues := map[string]types.AttributeValue{
		":status": &types.AttributeValueMemberS{Value: status},
	}

	if vehicleID != nil {
		updateExpression += ", assigned_vehicle_id = :vehicleID"
		expressionAttributeValues[":vehicleID"] = &types.AttributeValueMemberS{Value: *vehicleID}
	}

	_, err := d.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: jobID},
		},
		UpdateExpression: aws.String(updateExpression),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: expressionAttributeValues,
	})
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

func (d *DynamoDBJobStorage) GetJobsByStatus(ctx context.Context, status string) ([]*Job, error) {
	result, err := d.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(d.tableName),
		IndexName:              aws.String("status-index"),
		KeyConditionExpression: aws.String("#status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: status},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs by status: %w", err)
	}

	var jobs []*Job
	for _, item := range result.Items {
		var job Job
		err = attributevalue.UnmarshalMap(item, &job)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal job: %w", err)
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

func (d *DynamoDBJobStorage) GetAllJobs(ctx context.Context) ([]*Job, error) {
	result, err := d.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(d.tableName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan jobs: %w", err)
	}

	var jobs []*Job
	for _, item := range result.Items {
		var job Job
		err = attributevalue.UnmarshalMap(item, &job)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal job: %w", err)
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

func (d *DynamoDBJobStorage) GetJobsByVehicle(ctx context.Context, vehicleID string) ([]*Job, error) {
	result, err := d.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(d.tableName),
		IndexName:              aws.String("assigned-vehicle-index"),
		KeyConditionExpression: aws.String("assigned_vehicle_id = :vehicleID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":vehicleID": &types.AttributeValueMemberS{Value: vehicleID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs by vehicle: %w", err)
	}

	var jobs []*Job
	for _, item := range result.Items {
		var job Job
		err = attributevalue.UnmarshalMap(item, &job)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal job: %w", err)
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}
