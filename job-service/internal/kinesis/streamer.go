package kinesis

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"job-service/internal/storage"

	"github.com/aws/aws-sdk-go-v2/service/kinesis"
)

type Streamer struct {
	client     *kinesis.Client
	streamName string
}

type JobEvent struct {
	JobID      string    `json:"job_id"`
	EventType  string    `json:"event_type"` // created, assigned, completed
	Timestamp  time.Time `json:"timestamp"`
	VehicleID  *string   `json:"vehicle_id,omitempty"`
	JobType    string    `json:"job_type"`
	CustomerID string    `json:"customer_id"`
	Region     string    `json:"region"`
	PickupLat  float64   `json:"pickup_lat"`
	PickupLng  float64   `json:"pickup_lng"`
	DestLat    float64   `json:"dest_lat"`
	DestLng    float64   `json:"dest_lng"`
}

func NewStreamer(client *kinesis.Client, streamName string) *Streamer {
	return &Streamer{
		client:     client,
		streamName: streamName,
	}
}

func (s *Streamer) StreamJobEvent(eventType string, job *storage.Job) {
	if s.client == nil {
		return // Kinesis not enabled
	}

	event := JobEvent{
		JobID:      job.ID,
		EventType:  eventType,
		Timestamp:  time.Now().UTC(),
		VehicleID:  job.AssignedVehicleID,
		JobType:    job.JobType,
		CustomerID: job.CustomerID,
		Region:     job.Region,
		PickupLat:  job.PickupLat,
		PickupLng:  job.PickupLng,
		DestLat:    job.DestinationLat,
		DestLng:    job.DestinationLng,
	}

	data, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to marshal job event", "job_id", job.ID, "error", err)
		return
	}

	_, err = s.client.PutRecord(context.TODO(), &kinesis.PutRecordInput{
		StreamName:   &s.streamName,
		Data:         data,
		PartitionKey: &job.ID,
	})

	if err != nil {
		slog.Error("Failed to stream job event", "job_id", job.ID, "event_type", eventType, "error", err)
	} else {
		slog.Debug("Streamed job event", "job_id", job.ID, "event_type", eventType)
	}
}
