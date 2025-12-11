package service

import (
	"testing"

	"job-service/internal/storage"
)

func TestPricingConfig_CalculateFare_Ride(t *testing.T) {
	pricing := DefaultPricingConfig()

	job := &storage.Job{
		JobType:             "ride",
		EstimatedDistanceKm: 5.0, // 5km ride
	}

	pricing.CalculateFare(job)

	expectedBaseFare := 2.50
	expectedDistanceFare := 5.0 * 1.80                       // 5km * $1.80/km = $9.00
	expectedTotal := expectedBaseFare + expectedDistanceFare // $2.50 + $9.00 = $11.50

	if job.BaseFare != expectedBaseFare {
		t.Errorf("Expected base fare %.2f, got %.2f", expectedBaseFare, job.BaseFare)
	}

	if job.DistanceFare != expectedDistanceFare {
		t.Errorf("Expected distance fare %.2f, got %.2f", expectedDistanceFare, job.DistanceFare)
	}

	if job.FareAmount != expectedTotal {
		t.Errorf("Expected total fare %.2f, got %.2f", expectedTotal, job.FareAmount)
	}
}

func TestPricingConfig_CalculateFare_Delivery(t *testing.T) {
	pricing := DefaultPricingConfig()

	job := &storage.Job{
		JobType:             "delivery",
		EstimatedDistanceKm: 10.0, // Distance shouldn't matter for deliveries
	}

	pricing.CalculateFare(job)

	expectedBaseFare := 8.99
	expectedDistanceFare := 0.0 // Flat rate for deliveries
	expectedTotal := expectedBaseFare

	if job.BaseFare != expectedBaseFare {
		t.Errorf("Expected base fare %.2f, got %.2f", expectedBaseFare, job.BaseFare)
	}

	if job.DistanceFare != expectedDistanceFare {
		t.Errorf("Expected distance fare %.2f, got %.2f", expectedDistanceFare, job.DistanceFare)
	}

	if job.FareAmount != expectedTotal {
		t.Errorf("Expected total fare %.2f, got %.2f", expectedTotal, job.FareAmount)
	}
}

func TestDefaultPricingConfig(t *testing.T) {
	pricing := DefaultPricingConfig()

	if pricing.RideBaseFare != 2.50 {
		t.Errorf("Expected ride base fare 2.50, got %.2f", pricing.RideBaseFare)
	}

	if pricing.RidePerKm != 1.80 {
		t.Errorf("Expected ride per km 1.80, got %.2f", pricing.RidePerKm)
	}

	if pricing.DeliveryFlatRate != 8.99 {
		t.Errorf("Expected delivery flat rate 8.99, got %.2f", pricing.DeliveryFlatRate)
	}
}
