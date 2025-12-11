package storage

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDynamoDBClient mocks the DynamoDB client
type MockDynamoDBClient struct {
	mock.Mock
}

func (m *MockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*dynamodb.UpdateItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}

func (m *MockDynamoDBClient) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*dynamodb.ScanOutput), args.Error(1)
}

func TestDynamoDBVehicleStorage_CreateVehicle(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	storage := &DynamoDBVehicleStorage{
		client:    mockClient,
		tableName: "test-vehicles",
	}

	vehicle := &Vehicle{
		ID:           "test-vehicle-1",
		Region:       "us-west-2",
		Status:       "available",
		BatteryLevel: 100,
		LocationLat:  37.7749,
		LocationLng:  -122.4194,
		LastUpdated:  time.Now(),
		VehicleType:  "sedan",
	}

	mockClient.On("PutItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
		return *input.TableName == "test-vehicles"
	})).Return(&dynamodb.PutItemOutput{}, nil)

	err := storage.CreateVehicle(context.Background(), vehicle)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBVehicleStorage_GetVehicle_Success(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	storage := &DynamoDBVehicleStorage{
		client:    mockClient,
		tableName: "test-vehicles",
	}

	mockClient.On("GetItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
		return *input.TableName == "test-vehicles"
	})).Return(&dynamodb.GetItemOutput{
		Item: map[string]types.AttributeValue{
			"id":            &types.AttributeValueMemberS{Value: "test-vehicle-1"},
			"region":        &types.AttributeValueMemberS{Value: "us-west-2"},
			"status":        &types.AttributeValueMemberS{Value: "available"},
			"battery_level": &types.AttributeValueMemberN{Value: "100"},
			"location_lat":  &types.AttributeValueMemberN{Value: "37.7749"},
			"location_lng":  &types.AttributeValueMemberN{Value: "-122.4194"},
			"vehicle_type":  &types.AttributeValueMemberS{Value: "sedan"},
		},
	}, nil)

	vehicle, err := storage.GetVehicle(context.Background(), "test-vehicle-1")

	assert.NoError(t, err)
	assert.Equal(t, "test-vehicle-1", vehicle.ID)
	assert.Equal(t, "us-west-2", vehicle.Region)
	assert.Equal(t, "available", vehicle.Status)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBVehicleStorage_GetVehicle_NotFound(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	storage := &DynamoDBVehicleStorage{
		client:    mockClient,
		tableName: "test-vehicles",
	}

	mockClient.On("GetItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
		return *input.TableName == "test-vehicles"
	})).Return(&dynamodb.GetItemOutput{
		Item: nil,
	}, nil)

	vehicle, err := storage.GetVehicle(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Nil(t, vehicle)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBVehicleStorage_UpdateVehicleLocation(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	storage := &DynamoDBVehicleStorage{
		client:    mockClient,
		tableName: "test-vehicles",
	}

	mockClient.On("UpdateItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.UpdateItemInput) bool {
		return *input.TableName == "test-vehicles"
	})).Return(&dynamodb.UpdateItemOutput{}, nil)

	err := storage.UpdateVehicleLocation(context.Background(), "test-vehicle-1", 40.7128, -74.0060)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBVehicleStorage_UpdateVehicleStatus(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	storage := &DynamoDBVehicleStorage{
		client:    mockClient,
		tableName: "test-vehicles",
	}

	mockClient.On("UpdateItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.UpdateItemInput) bool {
		return *input.TableName == "test-vehicles"
	})).Return(&dynamodb.UpdateItemOutput{}, nil)

	jobID := "job-123"
	err := storage.UpdateVehicleStatus(context.Background(), "test-vehicle-1", "busy", &jobID)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBVehicleStorage_GetVehiclesByRegionAndStatus(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	storage := &DynamoDBVehicleStorage{
		client:    mockClient,
		tableName: "test-vehicles",
	}

	mockClient.On("Query", mock.Anything, mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
		return *input.TableName == "test-vehicles" && *input.IndexName == "region-status-index"
	})).Return(&dynamodb.QueryOutput{
		Items: []map[string]types.AttributeValue{
			{
				"id":            &types.AttributeValueMemberS{Value: "test-vehicle-1"},
				"region":        &types.AttributeValueMemberS{Value: "us-west-2"},
				"status":        &types.AttributeValueMemberS{Value: "available"},
				"battery_level": &types.AttributeValueMemberN{Value: "100"},
				"location_lat":  &types.AttributeValueMemberN{Value: "37.7749"},
				"location_lng":  &types.AttributeValueMemberN{Value: "-122.4194"},
				"vehicle_type":  &types.AttributeValueMemberS{Value: "sedan"},
			},
		},
	}, nil)

	vehicles, err := storage.GetVehiclesByRegionAndStatus(context.Background(), "us-west-2", "available")

	assert.NoError(t, err)
	assert.Len(t, vehicles, 1)
	assert.Equal(t, "test-vehicle-1", vehicles[0].ID)
	assert.Equal(t, "available", vehicles[0].Status)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBVehicleStorage_GetVehiclesByRegionAndStatus_ReservedKeyword(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	storage := &DynamoDBVehicleStorage{
		client:    mockClient,
		tableName: "test-vehicles",
	}

	// Test that both 'region' and 'status' are properly handled as reserved keywords
	mockClient.On("Query", mock.Anything, mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
		// Verify KeyConditionExpression uses ExpressionAttributeNames for both region and status
		expectedKeyCondition := "#region = :region AND #status = :status"
		if *input.KeyConditionExpression != expectedKeyCondition {
			t.Errorf("Expected KeyConditionExpression '%s', got '%s'", expectedKeyCondition, *input.KeyConditionExpression)
			return false
		}

		// Verify ExpressionAttributeNames maps both reserved keywords
		if input.ExpressionAttributeNames["#region"] != "region" {
			t.Errorf("Expected ExpressionAttributeNames['#region'] = 'region', got '%s'", input.ExpressionAttributeNames["#region"])
			return false
		}
		if input.ExpressionAttributeNames["#status"] != "status" {
			t.Errorf("Expected ExpressionAttributeNames['#status'] = 'status', got '%s'", input.ExpressionAttributeNames["#status"])
			return false
		}

		return true
	})).Return(&dynamodb.QueryOutput{
		Items: []map[string]types.AttributeValue{
			{
				"id":            &types.AttributeValueMemberS{Value: "test-vehicle-reserved"},
				"region":        &types.AttributeValueMemberS{Value: "us-west-2"},
				"status":        &types.AttributeValueMemberS{Value: "available"},
				"battery_level": &types.AttributeValueMemberN{Value: "85"},
				"location_lat":  &types.AttributeValueMemberN{Value: "45.5152"},
				"location_lng":  &types.AttributeValueMemberN{Value: "-122.6784"},
				"vehicle_type":  &types.AttributeValueMemberS{Value: "sedan"},
			},
		},
	}, nil)

	vehicles, err := storage.GetVehiclesByRegionAndStatus(context.Background(), "us-west-2", "available")

	assert.NoError(t, err)
	assert.Len(t, vehicles, 1)
	assert.Equal(t, "test-vehicle-reserved", vehicles[0].ID)
	assert.Equal(t, "us-west-2", vehicles[0].Region)
	assert.Equal(t, "available", vehicles[0].Status)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBVehicleStorage_GetAllVehicles(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	storage := &DynamoDBVehicleStorage{
		client:    mockClient,
		tableName: "test-vehicles",
	}

	mockClient.On("Scan", mock.Anything, mock.MatchedBy(func(input *dynamodb.ScanInput) bool {
		return *input.TableName == "test-vehicles"
	})).Return(&dynamodb.ScanOutput{
		Items: []map[string]types.AttributeValue{
			{
				"id":            &types.AttributeValueMemberS{Value: "test-vehicle-1"},
				"region":        &types.AttributeValueMemberS{Value: "us-west-2"},
				"status":        &types.AttributeValueMemberS{Value: "available"},
				"battery_level": &types.AttributeValueMemberN{Value: "100"},
				"location_lat":  &types.AttributeValueMemberN{Value: "37.7749"},
				"location_lng":  &types.AttributeValueMemberN{Value: "-122.4194"},
				"vehicle_type":  &types.AttributeValueMemberS{Value: "sedan"},
			},
			{
				"id":            &types.AttributeValueMemberS{Value: "test-vehicle-2"},
				"region":        &types.AttributeValueMemberS{Value: "us-west-2"},
				"status":        &types.AttributeValueMemberS{Value: "busy"},
				"battery_level": &types.AttributeValueMemberN{Value: "80"},
				"location_lat":  &types.AttributeValueMemberN{Value: "37.7849"},
				"location_lng":  &types.AttributeValueMemberN{Value: "-122.4094"},
				"vehicle_type":  &types.AttributeValueMemberS{Value: "sedan"},
			},
		},
	}, nil)

	vehicles, err := storage.GetAllVehicles(context.Background())

	assert.NoError(t, err)
	assert.Len(t, vehicles, 2)
	assert.Equal(t, "test-vehicle-1", vehicles[0].ID)
	assert.Equal(t, "test-vehicle-2", vehicles[1].ID)
	mockClient.AssertExpectations(t)
}
