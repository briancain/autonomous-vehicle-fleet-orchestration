package storage

import (
	"context"
	"fmt"
	"strconv"
	"time"

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

type DynamoDBVehicleStorage struct {
	client    DynamoDBAPI
	tableName string
}

func NewDynamoDBVehicleStorage(client DynamoDBAPI, tableName string) *DynamoDBVehicleStorage {
	return &DynamoDBVehicleStorage{
		client:    client,
		tableName: tableName,
	}
}

func (d *DynamoDBVehicleStorage) CreateVehicle(ctx context.Context, vehicle *Vehicle) error {
	item, err := attributevalue.MarshalMap(vehicle)
	if err != nil {
		return fmt.Errorf("failed to marshal vehicle: %w", err)
	}

	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to put vehicle: %w", err)
	}

	return nil
}

func (d *DynamoDBVehicleStorage) GetVehicle(ctx context.Context, vehicleID string) (*Vehicle, error) {
	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: vehicleID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicle: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("vehicle %s not found", vehicleID)
	}

	var vehicle Vehicle
	err = attributevalue.UnmarshalMap(result.Item, &vehicle)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal vehicle: %w", err)
	}

	return &vehicle, nil
}

func (d *DynamoDBVehicleStorage) UpdateVehicle(ctx context.Context, vehicle *Vehicle) error {
	item, err := attributevalue.MarshalMap(vehicle)
	if err != nil {
		return fmt.Errorf("failed to marshal vehicle: %w", err)
	}

	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update vehicle: %w", err)
	}

	return nil
}

func (d *DynamoDBVehicleStorage) UpdateVehicleLocationAndStatus(ctx context.Context, vehicleID string, lat, lng float64, status string) error {
	_, err := d.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: vehicleID},
		},
		UpdateExpression: aws.String("SET location_lat = :lat, location_lng = :lng, #status = :status, last_updated = :timestamp"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":lat":       &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", lat)},
			":lng":       &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", lng)},
			":status":    &types.AttributeValueMemberS{Value: status},
			":timestamp": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})
	return err
}

func (d *DynamoDBVehicleStorage) UpdateVehicleLocation(ctx context.Context, vehicleID string, lat, lng float64) error {
	_, err := d.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: vehicleID},
		},
		UpdateExpression: aws.String("SET location_lat = :lat, location_lng = :lng, last_updated = :timestamp"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":lat":       &types.AttributeValueMemberN{Value: strconv.FormatFloat(lat, 'f', -1, 64)},
			":lng":       &types.AttributeValueMemberN{Value: strconv.FormatFloat(lng, 'f', -1, 64)},
			":timestamp": &types.AttributeValueMemberS{Value: "2024-01-01T00:00:00Z"}, // TODO: use actual timestamp
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update vehicle location: %w", err)
	}

	return nil
}

func (d *DynamoDBVehicleStorage) UpdateVehicleStatus(ctx context.Context, vehicleID string, status string, jobID *string) error {
	updateExpression := "SET #status = :status, last_updated = :timestamp"
	expressionAttributeValues := map[string]types.AttributeValue{
		":status":    &types.AttributeValueMemberS{Value: status},
		":timestamp": &types.AttributeValueMemberS{Value: "2024-01-01T00:00:00Z"}, // TODO: use actual timestamp
	}

	if jobID != nil {
		updateExpression += ", current_job_id = :jobID"
		expressionAttributeValues[":jobID"] = &types.AttributeValueMemberS{Value: *jobID}
	} else {
		updateExpression += " REMOVE current_job_id"
	}

	_, err := d.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: vehicleID},
		},
		UpdateExpression: aws.String(updateExpression),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: expressionAttributeValues,
	})
	if err != nil {
		return fmt.Errorf("failed to update vehicle status: %w", err)
	}

	return nil
}

func (d *DynamoDBVehicleStorage) GetVehiclesByRegionAndStatus(ctx context.Context, region, status string) ([]*Vehicle, error) {
	result, err := d.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(d.tableName),
		IndexName:              aws.String("region-status-index"),
		KeyConditionExpression: aws.String("#region = :region AND #status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#region": "region",
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":region": &types.AttributeValueMemberS{Value: region},
			":status": &types.AttributeValueMemberS{Value: status},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query vehicles by region and status: %w", err)
	}

	var vehicles []*Vehicle
	for _, item := range result.Items {
		var vehicle Vehicle
		err = attributevalue.UnmarshalMap(item, &vehicle)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal vehicle: %w", err)
		}
		vehicles = append(vehicles, &vehicle)
	}

	return vehicles, nil
}

func (d *DynamoDBVehicleStorage) GetAllVehicles(ctx context.Context) ([]*Vehicle, error) {
	result, err := d.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(d.tableName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan vehicles: %w", err)
	}

	var vehicles []*Vehicle
	for _, item := range result.Items {
		var vehicle Vehicle
		err = attributevalue.UnmarshalMap(item, &vehicle)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal vehicle: %w", err)
		}
		vehicles = append(vehicles, &vehicle)
	}

	return vehicles, nil
}
