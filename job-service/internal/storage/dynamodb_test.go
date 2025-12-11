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

func TestDynamoDBJobStorage_CreateJob(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	storage := &DynamoDBJobStorage{
		client:    mockClient,
		tableName: "test-jobs",
	}

	job := &Job{
		ID:                  "test-job-1",
		JobType:             "ride",
		Status:              "pending",
		PickupLat:           37.7749,
		PickupLng:           -122.4194,
		DestinationLat:      37.7849,
		DestinationLng:      -122.4094,
		EstimatedDistanceKm: 1.5,
		CustomerID:          "customer-1",
		Region:              "us-west-2",
		CreatedAt:           time.Now(),
		FareAmount:          15.50,
		BaseFare:            5.00,
		DistanceFare:        10.50,
	}

	mockClient.On("PutItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
		return *input.TableName == "test-jobs"
	})).Return(&dynamodb.PutItemOutput{}, nil)

	err := storage.CreateJob(context.Background(), job)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBJobStorage_GetJob_Success(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	storage := &DynamoDBJobStorage{
		client:    mockClient,
		tableName: "test-jobs",
	}

	mockClient.On("GetItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
		return *input.TableName == "test-jobs"
	})).Return(&dynamodb.GetItemOutput{
		Item: map[string]types.AttributeValue{
			"id":                    &types.AttributeValueMemberS{Value: "test-job-1"},
			"job_type":              &types.AttributeValueMemberS{Value: "ride"},
			"status":                &types.AttributeValueMemberS{Value: "pending"},
			"pickup_lat":            &types.AttributeValueMemberN{Value: "37.7749"},
			"pickup_lng":            &types.AttributeValueMemberN{Value: "-122.4194"},
			"destination_lat":       &types.AttributeValueMemberN{Value: "37.7849"},
			"destination_lng":       &types.AttributeValueMemberN{Value: "-122.4094"},
			"estimated_distance_km": &types.AttributeValueMemberN{Value: "1.5"},
			"customer_id":           &types.AttributeValueMemberS{Value: "customer-1"},
			"region":                &types.AttributeValueMemberS{Value: "us-west-2"},
			"fare_amount":           &types.AttributeValueMemberN{Value: "15.50"},
			"base_fare":             &types.AttributeValueMemberN{Value: "5.00"},
			"distance_fare":         &types.AttributeValueMemberN{Value: "10.50"},
		},
	}, nil)

	job, err := storage.GetJob(context.Background(), "test-job-1")

	assert.NoError(t, err)
	assert.Equal(t, "test-job-1", job.ID)
	assert.Equal(t, "ride", job.JobType)
	assert.Equal(t, "pending", job.Status)
	assert.Equal(t, "customer-1", job.CustomerID)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBJobStorage_GetJob_NotFound(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	storage := &DynamoDBJobStorage{
		client:    mockClient,
		tableName: "test-jobs",
	}

	mockClient.On("GetItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
		return *input.TableName == "test-jobs"
	})).Return(&dynamodb.GetItemOutput{
		Item: nil,
	}, nil)

	job, err := storage.GetJob(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Nil(t, job)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBJobStorage_UpdateJobStatus(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	storage := &DynamoDBJobStorage{
		client:    mockClient,
		tableName: "test-jobs",
	}

	mockClient.On("UpdateItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.UpdateItemInput) bool {
		return *input.TableName == "test-jobs"
	})).Return(&dynamodb.UpdateItemOutput{}, nil)

	vehicleID := "vehicle-1"
	err := storage.UpdateJobStatus(context.Background(), "test-job-1", "assigned", &vehicleID)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBJobStorage_GetJobsByStatus(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	storage := &DynamoDBJobStorage{
		client:    mockClient,
		tableName: "test-jobs",
	}

	mockClient.On("Query", mock.Anything, mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
		return *input.TableName == "test-jobs" && *input.IndexName == "status-index"
	})).Return(&dynamodb.QueryOutput{
		Items: []map[string]types.AttributeValue{
			{
				"id":                    &types.AttributeValueMemberS{Value: "test-job-1"},
				"job_type":              &types.AttributeValueMemberS{Value: "ride"},
				"status":                &types.AttributeValueMemberS{Value: "pending"},
				"pickup_lat":            &types.AttributeValueMemberN{Value: "37.7749"},
				"pickup_lng":            &types.AttributeValueMemberN{Value: "-122.4194"},
				"destination_lat":       &types.AttributeValueMemberN{Value: "37.7849"},
				"destination_lng":       &types.AttributeValueMemberN{Value: "-122.4094"},
				"estimated_distance_km": &types.AttributeValueMemberN{Value: "1.5"},
				"customer_id":           &types.AttributeValueMemberS{Value: "customer-1"},
				"region":                &types.AttributeValueMemberS{Value: "us-west-2"},
			},
		},
	}, nil)

	jobs, err := storage.GetJobsByStatus(context.Background(), "pending")

	assert.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "test-job-1", jobs[0].ID)
	assert.Equal(t, "pending", jobs[0].Status)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBJobStorage_GetAllJobs(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	storage := &DynamoDBJobStorage{
		client:    mockClient,
		tableName: "test-jobs",
	}

	mockClient.On("Scan", mock.Anything, mock.MatchedBy(func(input *dynamodb.ScanInput) bool {
		return *input.TableName == "test-jobs"
	})).Return(&dynamodb.ScanOutput{
		Items: []map[string]types.AttributeValue{
			{
				"id":              &types.AttributeValueMemberS{Value: "test-job-1"},
				"job_type":        &types.AttributeValueMemberS{Value: "ride"},
				"status":          &types.AttributeValueMemberS{Value: "pending"},
				"pickup_lat":      &types.AttributeValueMemberN{Value: "37.7749"},
				"pickup_lng":      &types.AttributeValueMemberN{Value: "-122.4194"},
				"destination_lat": &types.AttributeValueMemberN{Value: "37.7849"},
				"destination_lng": &types.AttributeValueMemberN{Value: "-122.4094"},
				"customer_id":     &types.AttributeValueMemberS{Value: "customer-1"},
			},
			{
				"id":                  &types.AttributeValueMemberS{Value: "test-job-2"},
				"job_type":            &types.AttributeValueMemberS{Value: "delivery"},
				"status":              &types.AttributeValueMemberS{Value: "assigned"},
				"pickup_lat":          &types.AttributeValueMemberN{Value: "37.7649"},
				"pickup_lng":          &types.AttributeValueMemberN{Value: "-122.4294"},
				"destination_lat":     &types.AttributeValueMemberN{Value: "37.7949"},
				"destination_lng":     &types.AttributeValueMemberN{Value: "-122.3994"},
				"customer_id":         &types.AttributeValueMemberS{Value: "customer-2"},
				"assigned_vehicle_id": &types.AttributeValueMemberS{Value: "vehicle-1"},
			},
		},
	}, nil)

	jobs, err := storage.GetAllJobs(context.Background())

	assert.NoError(t, err)
	assert.Len(t, jobs, 2)
	assert.Equal(t, "test-job-1", jobs[0].ID)
	assert.Equal(t, "test-job-2", jobs[1].ID)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBJobStorage_GetJobsByVehicle(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	storage := &DynamoDBJobStorage{
		client:    mockClient,
		tableName: "test-jobs",
	}

	mockClient.On("Query", mock.Anything, mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
		return *input.TableName == "test-jobs" && *input.IndexName == "assigned-vehicle-index"
	})).Return(&dynamodb.QueryOutput{
		Items: []map[string]types.AttributeValue{
			{
				"id":                  &types.AttributeValueMemberS{Value: "test-job-1"},
				"job_type":            &types.AttributeValueMemberS{Value: "ride"},
				"status":              &types.AttributeValueMemberS{Value: "assigned"},
				"assigned_vehicle_id": &types.AttributeValueMemberS{Value: "vehicle-1"},
				"pickup_lat":          &types.AttributeValueMemberN{Value: "37.7749"},
				"pickup_lng":          &types.AttributeValueMemberN{Value: "-122.4194"},
				"destination_lat":     &types.AttributeValueMemberN{Value: "37.7849"},
				"destination_lng":     &types.AttributeValueMemberN{Value: "-122.4094"},
				"customer_id":         &types.AttributeValueMemberS{Value: "customer-1"},
			},
		},
	}, nil)

	jobs, err := storage.GetJobsByVehicle(context.Background(), "vehicle-1")

	assert.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "test-job-1", jobs[0].ID)
	assert.Equal(t, "vehicle-1", *jobs[0].AssignedVehicleID)
	mockClient.AssertExpectations(t)
}
