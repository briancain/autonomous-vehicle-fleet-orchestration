package simulator

// ChargingStation represents a charging location
type ChargingStation struct {
	ID  string  `json:"id"`
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// GetChargingStations returns available charging stations for a region
func GetChargingStations(region string) []ChargingStation {
	// Portland area charging stations - based on real EV charging locations
	if region == "us-west-2" {
		return []ChargingStation{
			// Pioneer Place Mall (downtown)
			{ID: "pioneer-place", Lat: 45.5188, Lng: -122.6746},
			// Lloyd Center (northeast Portland)
			{ID: "lloyd-center", Lat: 45.5311, Lng: -122.6536},
			// OHSU (southwest hills)
			{ID: "ohsu-campus", Lat: 45.4993, Lng: -122.6859},
			// Portland International Airport
			{ID: "pdx-airport", Lat: 45.5898, Lng: -122.5951},
			// Whole Foods Hawthorne (southeast)
			{ID: "hawthorne-whole-foods", Lat: 45.5122, Lng: -122.6208},
		}
	}

	// Default fallback stations
	return []ChargingStation{
		{ID: "default-station-1", Lat: 37.7749, Lng: -122.4194},
		{ID: "default-station-2", Lat: 37.7849, Lng: -122.4094},
	}
}

// FindNearestChargingStation finds the closest charging station to a vehicle
func FindNearestChargingStation(vehicleLat, vehicleLng float64, region string) ChargingStation {
	stations := GetChargingStations(region)

	if len(stations) == 0 {
		// Fallback to a default station
		return ChargingStation{ID: "emergency-station", Lat: vehicleLat, Lng: vehicleLng}
	}

	// Find the actual nearest station
	nearest := stations[0]
	minDistance := haversineDistance(vehicleLat, vehicleLng, nearest.Lat, nearest.Lng)

	for _, station := range stations[1:] {
		distance := haversineDistance(vehicleLat, vehicleLng, station.Lat, station.Lng)
		if distance < minDistance {
			minDistance = distance
			nearest = station
		}
	}
	return nearest
}
