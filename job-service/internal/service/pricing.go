package service

import "job-service/internal/storage"

// PricingConfig holds pricing parameters
type PricingConfig struct {
	// Ride pricing (distance-based like taxi)
	RideBaseFare float64 // Base fare for rides
	RidePerKm    float64 // Per kilometer rate for rides

	// Delivery pricing (flat rate)
	DeliveryFlatRate float64 // Flat rate for deliveries
}

// DefaultPricingConfig returns standard Portland pricing
func DefaultPricingConfig() *PricingConfig {
	return &PricingConfig{
		RideBaseFare:     2.50, // $2.50 base fare
		RidePerKm:        1.80, // $1.80 per km (similar to Portland taxi rates)
		DeliveryFlatRate: 8.99, // $8.99 flat delivery fee
	}
}

// CalculateFare calculates the fare for a job based on type and distance
func (p *PricingConfig) CalculateFare(job *storage.Job) {
	if job.JobType == "ride" {
		// Distance-based pricing for rides
		job.BaseFare = p.RideBaseFare
		job.DistanceFare = job.EstimatedDistanceKm * p.RidePerKm
		job.FareAmount = job.BaseFare + job.DistanceFare
	} else {
		// Flat rate for deliveries
		job.BaseFare = p.DeliveryFlatRate
		job.DistanceFare = 0.0
		job.FareAmount = job.BaseFare
	}
}
