package service

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"job-service/internal/storage"
)

// DemoJobGenerator handles automatic job creation for demo purposes
type DemoJobGenerator struct {
	jobService *JobService
	isRunning  bool
	stopChan   chan bool
	interval   time.Duration
	maxJobs    int
}

// NewDemoJobGenerator creates a new demo job generator
func NewDemoJobGenerator(jobService *JobService, interval time.Duration) *DemoJobGenerator {
	return &DemoJobGenerator{
		jobService: jobService,
		interval:   interval,
		stopChan:   make(chan bool),
		maxJobs:    25, // Limit to 25 active jobs for demo
	}
}

// Start begins generating random jobs
func (d *DemoJobGenerator) Start() {
	if d.isRunning {
		return
	}

	d.isRunning = true
	slog.Info("Demo job generator started", "max_jobs", d.maxJobs, "interval", d.interval)

	go func() {
		for {
			select {
			case <-d.stopChan:
				return
			default:
				// Check active job limit (pending + in-progress only)
				activeJobs, err := d.jobService.GetActiveJobCount()
				if err != nil {
					slog.Error("Failed to get active job count", "error", err)
					time.Sleep(d.interval)
					continue
				}
				if activeJobs >= d.maxJobs {
					slog.Info("Demo active job limit reached, pausing generation", "active_jobs", activeJobs, "max_jobs", d.maxJobs)
					time.Sleep(d.interval)
					continue
				}

				d.createRandomJob()

				// Add jitter: random interval between 10-30 seconds
				jitter := time.Duration(10+rand.Intn(20)) * time.Second
				time.Sleep(jitter)
			}
		}
	}()
}

// Stop stops generating jobs
func (d *DemoJobGenerator) Stop() {
	if !d.isRunning {
		return
	}

	d.isRunning = false
	d.stopChan <- true
	slog.Info("Demo job generator stopped")
}

// IsRunning returns whether the generator is active
func (d *DemoJobGenerator) IsRunning() bool {
	return d.isRunning
}

// createRandomJob generates a realistic job
func (d *DemoJobGenerator) createRandomJob() {
	locations := getPortlandLocations()
	customers := getRandomCustomers()

	// 70% rides, 30% deliveries
	jobType := "ride"
	if rand.Float64() < 0.3 {
		jobType = "delivery"
	}

	// Random pickup location
	pickup := locations[rand.Intn(len(locations))]

	// Random destination (different from pickup)
	var destination Location
	for {
		destination = locations[rand.Intn(len(locations))]
		if destination.Name != pickup.Name {
			break
		}
	}

	// Random customer
	customer := customers[rand.Intn(len(customers))]

	ctx := context.Background()
	var createdJob *storage.Job
	var err error

	if jobType == "ride" {
		createdJob, err = d.jobService.CreateRideJob(ctx, customer, "us-west-2",
			pickup.Lat, pickup.Lng, destination.Lat, destination.Lng)
	} else {
		// Create simple delivery details
		deliveryDetails := &storage.DeliveryDetails{
			RestaurantName: "Demo Restaurant",
			Items:          []string{"Demo Package"},
			Instructions:   "Demo delivery - handle with care",
		}
		createdJob, err = d.jobService.CreateDeliveryJob(ctx, customer, "us-west-2",
			pickup.Lat, pickup.Lng, destination.Lat, destination.Lng, deliveryDetails)
	}

	if err != nil {
		slog.Error("Failed to create demo job", "error", err)
		return
	}

	slog.Info("Created demo job",
		"job_type", jobType,
		"job_id", createdJob.ID,
		"pickup", pickup.Name,
		"destination", destination.Name,
		"customer", customer,
		"max_jobs", d.maxJobs)
}

// Location represents a Portland location
type Location struct {
	Name string
	Lat  float64
	Lng  float64
}

// getPortlandLocations returns 100+ realistic street-level Portland locations
func getPortlandLocations() []Location {
	return []Location{
		// Downtown Portland - Financial/Business District (15 locations)
		{"Pioneer Courthouse Square", 45.5188, -122.6793},
		{"Powell's City of Books", 45.5230, -122.6814},
		{"Portland Building", 45.5145, -122.6794},
		{"Union Station", 45.5289, -122.6765},
		{"Multnomah Athletic Club", 45.5197, -122.6925},
		{"Director Park", 45.5181, -122.6850},
		{"Tom McCall Waterfront Park", 45.5152, -122.6647},
		{"Salmon Street Springs", 45.5142, -122.6647},
		{"Oregon Convention Center", 45.5289, -122.6633},
		{"Moda Center", 45.5316, -122.6668},
		{"Pioneer Place Mall", 45.5188, -122.6746},
		{"Crystal Ballroom", 45.5230, -122.6814},
		{"Keller Auditorium", 45.5145, -122.6794},
		{"Arlene Schnitzer Concert Hall", 45.5145, -122.6850},
		{"Portland Art Museum", 45.5181, -122.6850},

		// Pearl District (10 locations)
		{"Pearl District Whole Foods", 45.5266, -122.6908},
		{"Jamison Square", 45.5263, -122.6919},
		{"Tanner Springs Park", 45.5284, -122.6925},
		{"Fields Park", 45.5298, -122.6944},
		{"Ecotrust Building", 45.5263, -122.6908},
		{"Brewery Blocks", 45.5230, -122.6908},
		{"Pearl District Safeway", 45.5284, -122.6908},
		{"Lovejoy Fountain Park", 45.5230, -122.6925},
		{"Pearl District Starbucks", 45.5263, -122.6925},
		{"Anthropologie Pearl", 45.5263, -122.6908},

		// Southeast Portland (20 locations)
		{"Hawthorne Bridge East End", 45.5122, -122.6687},
		{"Division/Clinton Food Carts", 45.5048, -122.6540},
		{"Laurelhurst Park", 45.5162, -122.6295},
		{"Mount Tabor Summit", 45.5118, -122.5933},
		{"Sellwood Bridge", 45.4632, -122.6681},
		{"OHSU Waterfront Campus", 45.4983, -122.6739},
		{"Tilikum Crossing", 45.5017, -122.6656},
		{"Eastbank Esplanade", 45.5152, -122.6647},
		{"Hawthorne District", 45.5122, -122.6208},
		{"Division District", 45.5048, -122.6540},
		{"Richmond District", 45.4764, -122.6540},
		{"Woodstock District", 45.4764, -122.6319},
		{"Reed College", 45.4823, -122.6319},
		{"Ladd's Addition", 45.5048, -122.6540},
		{"Hosford-Abernethy", 45.4983, -122.6540},
		{"Brooklyn District", 45.4983, -122.6540},
		{"Creston Park", 45.4764, -122.6319},
		{"Mt. Tabor Park", 45.5118, -122.5933},
		{"Laurelhurst Theater", 45.5162, -122.6295},
		{"New Seasons Hawthorne", 45.5122, -122.6208},

		// Northeast Portland (15 locations)
		{"Lloyd Center Mall", 45.5311, -122.6536},
		{"Alberta Arts District", 45.5581, -122.6656},
		{"Mississippi District", 45.5459, -122.6759},
		{"Williams Avenue", 45.5459, -122.6656},
		{"Fremont District", 45.5581, -122.6759},
		{"Beaumont Village", 45.5311, -122.6208},
		{"Irvington District", 45.5459, -122.6536},
		{"Rose Quarter", 45.5316, -122.6668},
		{"Lloyd District", 45.5311, -122.6536},
		{"Holladay Park", 45.5311, -122.6536},
		{"Grant Park", 45.5459, -122.6536},
		{"Klickitat Street", 45.5581, -122.6656},
		{"Going Street", 45.5581, -122.6759},
		{"Concordia District", 45.5581, -122.6656},
		{"Sabin District", 45.5581, -122.6656},

		// Southwest Portland (12 locations)
		{"OHSU Main Campus", 45.4993, -122.6859},
		{"Portland State University", 45.5118, -122.6839},
		{"South Waterfront", 45.4983, -122.6739},
		{"Marquam Hill", 45.4993, -122.6859},
		{"Burlingame District", 45.4632, -122.6908},
		{"Johns Landing", 45.4764, -122.6739},
		{"Corbett District", 45.4993, -122.6739},
		{"Lair Hill", 45.5048, -122.6739},
		{"OHSU Tram", 45.4993, -122.6859},
		{"Duniway Park", 45.5048, -122.6739},
		{"Willamette Park", 45.4764, -122.6739},
		{"Gabriel Park", 45.4511, -122.6908},

		// Northwest Portland (10 locations)
		{"Forest Park Entrance", 45.5701, -122.7603},
		{"NW 23rd Avenue", 45.5298, -122.6979},
		{"Nob Hill District", 45.5298, -122.6979},
		{"Wallace Park", 45.5298, -122.7025},
		{"Alphabet District", 45.5263, -122.6979},
		{"Couch Park", 45.5263, -122.6979},
		{"Thurman Street", 45.5298, -122.6979},
		{"Burnside Street", 45.5230, -122.6979},
		{"NW Industrial District", 45.5459, -122.7025},
		{"Slabtown District", 45.5263, -122.6979},

		// North Portland (8 locations)
		{"St. Johns Bridge", 45.5816, -122.7603},
		{"Cathedral Park", 45.5816, -122.7603},
		{"Kenton District", 45.5816, -122.6908},
		{"Piedmont District", 45.5581, -122.6908},
		{"Overlook Park", 45.5459, -122.6908},
		{"Arbor Lodge", 45.5816, -122.6908},
		{"University Park", 45.5816, -122.7161},
		{"St. Johns Town Center", 45.5816, -122.7603},

		// West Hills & Washington Park (8 locations)
		{"Oregon Zoo", 45.5099, -122.7161},
		{"Washington Park", 45.5099, -122.7161},
		{"International Rose Garden", 45.5188, -122.7161},
		{"Japanese Garden", 45.5188, -122.7161},
		{"Hoyt Arboretum", 45.5099, -122.7161},
		{"World Forestry Center", 45.5099, -122.7161},
		{"Vietnam Veterans Memorial", 45.5099, -122.7161},
		{"Pittock Mansion", 45.5230, -122.7161},

		// Airport and Outer Areas (5 locations)
		{"PDX Departures", 45.5898, -122.5951},
		{"PDX Arrivals", 45.5881, -122.5975},
		{"Columbia River", 45.5898, -122.5633},
		{"Jantzen Beach", 45.6062, -122.6908},
		{"Hayden Island", 45.6062, -122.6908},

		// Shopping Centers & Malls (7 locations)
		{"IKEA Portland", 45.5533, -122.6789},
		{"Bridgeport Village", 45.3816, -122.7603},
		{"Washington Square", 45.4511, -122.7603},
		{"Clackamas Town Center", 45.4511, -122.5633},
		{"Eastport Plaza", 45.5311, -122.5633},
		{"Hollywood District", 45.5311, -122.6208},
		{"Woodstock Neighborhood", 45.4764, -122.6319},

		// Entertainment & Culture (5 locations)
		{"Hawthorne Theater", 45.5122, -122.6208},
		{"McMenamins Kennedy School", 45.5581, -122.6656},
		{"Bagdad Theater", 45.5122, -122.6208},
		{"Academy Theater", 45.5048, -122.6540},
		{"Clinton Street Theater", 45.5048, -122.6540},

		// Hospitals & Medical (5 locations)
		{"Legacy Emanuel Hospital", 45.5459, -122.6656},
		{"Providence Portland", 45.5230, -122.6319},
		{"Kaiser Permanente", 45.5311, -122.6319},
		{"Adventist Medical Center", 45.4511, -122.5633},
		{"Shriners Hospital", 45.4511, -122.6908},

		// Universities & Schools (5 locations)
		{"University of Portland", 45.5701, -122.7161},
		{"Lewis & Clark College", 45.4511, -122.6681},
		{"Concordia University", 45.5581, -122.6656},
		{"Warner Pacific University", 45.4764, -122.6319},
		{"Marylhurst University", 45.3816, -122.7603},
	}
}

// getRandomCustomers returns realistic customer names
func getRandomCustomers() []string {
	return []string{
		"alex-chen", "sarah-johnson", "mike-rodriguez", "emma-davis",
		"james-wilson", "lisa-anderson", "david-brown", "maria-garcia",
		"chris-taylor", "jennifer-white", "robert-lee", "amanda-clark",
		"kevin-martinez", "stephanie-lewis", "brian-hall", "nicole-young",
		"portland-tourist-1", "business-traveler", "medical-patient",
		"airport-shuttle", "food-delivery", "package-express",
		"urgent-courier", "grocery-delivery", "pharmacy-run",
		"zoo-visitor", "concert-goer", "hospital-visitor", "student-rider",
		"shopping-trip", "date-night", "family-outing", "work-commute",
	}
}
