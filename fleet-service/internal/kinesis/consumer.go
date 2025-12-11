package kinesis

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"fleet-service/internal/service"

	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
)

type Consumer struct {
	client       *kinesis.Client
	streamName   string
	fleetService *service.FleetService
}

type VehicleTelemetry struct {
	VehicleID string  `json:"vehicle_id"`
	Timestamp string  `json:"timestamp"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Status    string  `json:"status"`
	Battery   float64 `json:"battery"`
	JobID     *string `json:"job_id,omitempty"`
}

func NewConsumer(client *kinesis.Client, streamName string, fleetService *service.FleetService) *Consumer {
	return &Consumer{
		client:       client,
		streamName:   streamName,
		fleetService: fleetService,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	slog.Info("Starting Kinesis consumer", "stream", c.streamName)

	// Get stream description to find shards
	describeOutput, err := c.client.DescribeStream(ctx, &kinesis.DescribeStreamInput{
		StreamName: &c.streamName,
	})
	if err != nil {
		slog.Error("Failed to describe Kinesis stream", "error", err)
		return
	}

	// Process each shard
	for _, shard := range describeOutput.StreamDescription.Shards {
		go c.processShard(ctx, *shard.ShardId)
	}
}

func (c *Consumer) processShard(ctx context.Context, shardID string) {
	slog.Info("Processing shard", "shard_id", shardID)

	// Get shard iterator
	iteratorOutput, err := c.client.GetShardIterator(ctx, &kinesis.GetShardIteratorInput{
		StreamName:        &c.streamName,
		ShardId:           &shardID,
		ShardIteratorType: types.ShardIteratorTypeLatest,
	})
	if err != nil {
		slog.Error("Failed to get shard iterator", "error", err, "shard_id", shardID)
		return
	}

	shardIterator := iteratorOutput.ShardIterator

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping shard processing", "shard_id", shardID)
			return
		default:
			if shardIterator == nil {
				slog.Warn("Shard iterator is nil, stopping", "shard_id", shardID)
				return
			}

			// Get records
			recordsOutput, err := c.client.GetRecords(ctx, &kinesis.GetRecordsInput{
				ShardIterator: shardIterator,
			})
			if err != nil {
				slog.Error("Failed to get records", "error", err, "shard_id", shardID)
				time.Sleep(1 * time.Second)
				continue
			}

			// Process records
			for _, record := range recordsOutput.Records {
				c.processRecord(record)
			}

			shardIterator = recordsOutput.NextShardIterator
			time.Sleep(1 * time.Second) // Avoid aggressive polling
		}
	}
}

func (c *Consumer) processRecord(record types.Record) {
	var telemetry VehicleTelemetry
	if err := json.Unmarshal(record.Data, &telemetry); err != nil {
		slog.Error("Failed to unmarshal telemetry record", "error", err)
		return
	}

	slog.Debug("Processing vehicle telemetry",
		"vehicle_id", telemetry.VehicleID,
		"lat", telemetry.Latitude,
		"lng", telemetry.Longitude,
		"status", telemetry.Status,
		"battery", telemetry.Battery)

	// This is supplemental analytics - we don't update the primary data store
	// In a real implementation, this could feed into analytics dashboards,
	// ML models for route optimization, or real-time monitoring systems
}
