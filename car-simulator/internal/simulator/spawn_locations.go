package simulator

import "math/rand"

// SpawnLocation represents a safe vehicle spawn point
type SpawnLocation struct {
	Name string
	Lat  float64
	Lng  float64
}

// GetPortlandSpawnLocations returns safe vehicle spawn points in Portland
func GetPortlandSpawnLocations() []SpawnLocation {
	return []SpawnLocation{
		// Downtown parking areas
		{"Pioneer Courthouse Square", 45.5188, -122.6793},
		{"Union Station Parking", 45.5289, -122.6765},
		{"Portland Building Lot", 45.5145, -122.6794},

		// Shopping center parking lots
		{"Lloyd Center Parking", 45.5311, -122.6536},
		{"Pioneer Place Garage", 45.5188, -122.6746},

		// Hospital/University areas
		{"OHSU Campus Parking", 45.4993, -122.6859},
		{"Portland State Parking", 45.5118, -122.6839},

		// Neighborhood commercial areas
		{"Hawthorne District", 45.5122, -122.6208},
		{"Alberta Arts District", 45.5581, -122.6656},
		{"Mississippi District", 45.5459, -122.6759},
		{"Pearl District", 45.5266, -122.6908},
		{"NW 23rd Avenue", 45.5298, -122.6979},

		// Transit hubs
		{"PDX Airport Pickup", 45.5898, -122.5951},
		{"Eastbank Esplanade", 45.5152, -122.6647},

		// Park and ride locations
		{"Washington Park", 45.5099, -122.7161},
		{"Laurelhurst Park", 45.5162, -122.6295},
		{"Mount Tabor Park", 45.5118, -122.5933},
	}
}

// GetRandomSpawnLocation returns a random safe spawn location
func GetRandomSpawnLocation() SpawnLocation {
	locations := GetPortlandSpawnLocations()
	return locations[rand.Intn(len(locations))]
}
