"use strict";
// API Configuration - use proxy endpoints to avoid CORS
const API_BASE = {
    fleet: '/api/fleet',
    jobs: '/api/jobs'
};
// Dashboard Class
class FleetDashboard {
    constructor() {
        this.vehicleMarkers = new Map();
        this.chargingStationMarkers = [];
        this.jobMarkers = new Map(); // job_id -> [pickup, destination, line]
        this.updateInterval = 3000; // 3 seconds
        this.initMap();
        this.addChargingStations();
        this.startUpdates();
    }
    initMap() {
        // Initialize map centered on Portland
        this.map = L.map('map').setView([45.5152, -122.6784], 12);
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: '© OpenStreetMap contributors'
        }).addTo(this.map);
    }
    addChargingStations() {
        console.log('Adding charging stations to map...');
        // Portland charging stations - based on real EV charging locations
        const chargingStations = [
            { id: 'pioneer-place', lat: 45.5188, lng: -122.6746, name: 'Pioneer Place Mall' },
            { id: 'lloyd-center', lat: 45.5311, lng: -122.6536, name: 'Lloyd Center' },
            { id: 'ohsu-campus', lat: 45.4993, lng: -122.6859, name: 'OHSU Campus' },
            { id: 'pdx-airport', lat: 45.5898, lng: -122.5951, name: 'PDX Airport' },
            { id: 'hawthorne-whole-foods', lat: 45.5122, lng: -122.6208, name: 'Hawthorne Whole Foods' }
        ];
        chargingStations.forEach(station => {
            const icon = L.divIcon({
                html: `<div style="
                    background: #f39c12; 
                    width: 28px; 
                    height: 28px; 
                    border-radius: 50%; 
                    border: 3px solid white; 
                    box-shadow: 0 2px 6px rgba(0,0,0,0.3);
                    display: flex; 
                    align-items: center; 
                    justify-content: center; 
                    font-size: 16px; 
                    color: white;
                    font-weight: bold;
                    z-index: 1000;
                ">⚡</div>`,
                iconSize: [34, 34],
                className: 'charging-station-marker'
            });
            const marker = L.marker([station.lat, station.lng], {
                icon,
                zIndexOffset: 1000 // Ensure charging stations appear above other markers
            })
                .bindPopup(`
                    <strong>⚡ Charging Station</strong><br>
                    Location: ${station.name}<br>
                    ID: ${station.id}
                `)
                .on('click', () => {
                this.showChargingVehicles(station);
            });
            marker.addTo(this.map);
            this.chargingStationMarkers.push(marker);
            console.log(`Added charging station: ${station.name} at ${station.lat}, ${station.lng}`);
        });
        console.log(`Total charging stations added: ${this.chargingStationMarkers.length}`);
    }
    ensureChargingStations() {
        // Check if charging stations are still on the map
        const visibleStations = this.chargingStationMarkers.filter(marker => this.map.hasLayer(marker));
        // If some are missing, re-add all charging stations
        if (visibleStations.length < this.chargingStationMarkers.length) {
            console.log('Charging stations missing, re-adding...');
            this.chargingStationMarkers.forEach(marker => {
                if (!this.map.hasLayer(marker)) {
                    marker.addTo(this.map);
                }
            });
        }
    }
    async fetchData(url) {
        try {
            const response = await fetch(url);
            if (!response.ok)
                throw new Error(`HTTP ${response.status}`);
            return await response.json();
        }
        catch (error) {
            console.error(`Failed to fetch ${url}:`, error);
            this.updateConnectionStatus(false);
            return null;
        }
    }
    async updateDashboard() {
        const [vehicles, jobs, revenue] = await Promise.all([
            this.fetchData(`${API_BASE.fleet}/vehicles`),
            this.fetchData(`${API_BASE.jobs}/jobs`),
            this.fetchData(`${API_BASE.jobs}/revenue`)
        ]);
        if (vehicles) {
            this.updateConnectionStatus(true);
            const jobsArray = jobs || []; // Handle null jobs response
            this.updateMetrics(vehicles, jobsArray);
            this.updateRevenue(revenue);
            this.updateMap(vehicles);
            this.updateJobMarkers(jobsArray);
            this.updateJobList(jobsArray);
            this.updateVehicleList(vehicles);
            this.updateLastUpdateTime();
            // Ensure charging stations are always visible
            this.ensureChargingStations();
        }
    }
    updateConnectionStatus(connected) {
        const status = document.getElementById('connection-status');
        status.style.color = connected ? '#27ae60' : '#e74c3c';
        status.textContent = connected ? '●' : '●';
    }
    updateMetrics(vehicles, jobs) {
        const activeVehicles = vehicles.filter(v => v.status !== 'charging').length;
        const pendingJobs = jobs.filter(j => j.status === 'pending').length;
        const completedToday = jobs.filter(j => j.status === 'completed').length;
        const avgBattery = vehicles.reduce((sum, v) => sum + v.battery_level, 0) / vehicles.length;
        document.getElementById('active-vehicles').textContent = activeVehicles.toString();
        document.getElementById('pending-jobs').textContent = pendingJobs.toString();
        document.getElementById('completed-jobs').textContent = completedToday.toString();
        document.getElementById('fleet-health').textContent = `${Math.round(avgBattery)}%`;
    }
    updateRevenue(revenue) {
        if (revenue) {
            document.getElementById('total-revenue').textContent = `$${revenue.total_revenue?.toFixed(2) || '0.00'}`;
            document.getElementById('ride-revenue').textContent = `$${revenue.ride_revenue?.toFixed(2) || '0.00'}`;
        }
    }
    updateMap(vehicles) {
        // Clear existing vehicle markers
        this.vehicleMarkers.forEach(marker => this.map.removeLayer(marker));
        this.vehicleMarkers.clear();
        // Add vehicle markers
        vehicles.forEach(vehicle => {
            const icon = this.getVehicleIcon(vehicle.status);
            const marker = L.marker([vehicle.location_lat, vehicle.location_lng], { icon })
                .bindPopup(`
                    <strong>Vehicle ${vehicle.id}</strong><br>
                    Status: ${vehicle.status}<br>
                    Battery: ${vehicle.battery_level}%<br>
                    ${vehicle.current_job_id ? `Job: ${vehicle.current_job_id}` : ''}
                `);
            marker.addTo(this.map);
            this.vehicleMarkers.set(vehicle.id, marker);
        });
    }
    getVehicleIcon(status) {
        const colors = {
            available: '#27ae60',
            busy: '#e74c3c',
            charging: '#f39c12'
        };
        return L.divIcon({
            html: `<div style="background: ${colors[status] || '#666'}; width: 20px; height: 12px; border-radius: 6px; border: 2px solid white; position: relative;">
                     <div style="position: absolute; bottom: -2px; left: 2px; width: 4px; height: 4px; background: #333; border-radius: 50%;"></div>
                     <div style="position: absolute; bottom: -2px; right: 2px; width: 4px; height: 4px; background: #333; border-radius: 50%;"></div>
                   </div>`,
            iconSize: [24, 16],
            className: 'vehicle-marker'
        });
    }
    updateJobMarkers(jobs) {
        // Clear existing job markers
        this.jobMarkers.forEach(markers => {
            markers.forEach(marker => {
                if (this.map.hasLayer(marker)) {
                    this.map.removeLayer(marker);
                }
            });
        });
        this.jobMarkers.clear();
        // Add markers for active jobs (not completed)
        jobs.filter(job => job.status !== 'completed').forEach(job => {
            const markers = [];
            // Pickup marker - different icons for ride vs delivery
            const isRide = job.job_type === 'ride';
            const pickupIcon = L.divIcon({
                html: isRide
                    ? `<div style="background: #3498db; width: 20px; height: 20px; border-radius: 50%; border: 2px solid white; display: flex; align-items: center; justify-content: center; font-size: 12px; color: white;">👤</div>`
                    : `<div style="background: #e67e22; width: 20px; height: 20px; border-radius: 4px; border: 2px solid white; display: flex; align-items: center; justify-content: center; font-size: 12px; color: white;">📦</div>`,
                iconSize: [24, 24],
                className: 'job-pickup-marker'
            });
            const pickupMarker = L.marker([job.pickup_lat, job.pickup_lng], { icon: pickupIcon })
                .bindPopup(`
                    <strong>${isRide ? '👤' : '📦'} ${job.job_type.toUpperCase()} Pickup</strong><br>
                    Job ID: ${job.id}<br>
                    Status: ${job.status}<br>
                    Customer: ${job.customer_id}<br>
                    ${job.assigned_vehicle_id ? `Vehicle: ${job.assigned_vehicle_id}` : 'Unassigned'}
                `);
            pickupMarker.addTo(this.map);
            markers.push(pickupMarker);
            // Destination marker (if exists)
            if (job.destination_lat && job.destination_lng) {
                const destIcon = L.divIcon({
                    html: `<div style="background: #9b59b6; width: 20px; height: 20px; border-radius: 50%; border: 2px solid white; display: flex; align-items: center; justify-content: center; font-size: 12px; color: white;">🎯</div>`,
                    iconSize: [24, 24],
                    className: 'job-destination-marker'
                });
                const destMarker = L.marker([job.destination_lat, job.destination_lng], { icon: destIcon })
                    .bindPopup(`
                        <strong>🎯 ${job.job_type.toUpperCase()} Destination</strong><br>
                        Job ID: ${job.id}<br>
                        Status: ${job.status}<br>
                        Customer: ${job.customer_id}
                    `);
                destMarker.addTo(this.map);
                markers.push(destMarker);
                // Draw line between pickup and destination
                const line = L.polyline([
                    [job.pickup_lat, job.pickup_lng],
                    [job.destination_lat, job.destination_lng]
                ], {
                    color: job.status === 'assigned' ? '#3498db' : '#27ae60',
                    weight: 2,
                    opacity: 0.7,
                    dashArray: '5, 5'
                });
                line.addTo(this.map);
                markers.push(line);
            }
            this.jobMarkers.set(job.id, markers);
        });
    }
    updateJobList(jobs) {
        const jobList = document.getElementById('job-list');
        const jobCount = document.getElementById('job-count');
        // Update job count
        jobCount.textContent = jobs.length.toString();
        // Sort jobs by status priority: pending > assigned > in_progress > completed > failed
        const statusOrder = { 'pending': 0, 'assigned': 1, 'in_progress': 2, 'completed': 3, 'failed': 4 };
        const sortedJobs = jobs.sort((a, b) => {
            const aOrder = statusOrder[a.status] ?? 5;
            const bOrder = statusOrder[b.status] ?? 5;
            if (aOrder !== bOrder)
                return aOrder - bOrder;
            return a.id.localeCompare(b.id); // Secondary sort by ID
        });
        jobList.innerHTML = sortedJobs.map(job => `
            <div class="job-item ${job.status}" onclick="dashboard.focusOnJob('${job.id}')">
                <strong>${job.job_type.toUpperCase()}</strong> - ${job.id}<br>
                Status: ${job.status}<br>
                Customer: ${job.customer_id}
            </div>
        `).join('');
    }
    showChargingVehicles(station) {
        // Get current vehicles data
        fetch(`${API_BASE.fleet}/vehicles`)
            .then(response => response.json())
            .then(vehicles => {
            // Find vehicles that are charging near this station (within 100m)
            const chargingVehicles = vehicles.filter((vehicle) => {
                if (vehicle.status !== 'charging')
                    return false;
                // Calculate distance to station
                const distance = this.calculateDistance(vehicle.location_lat, vehicle.location_lng, station.lat, station.lng);
                return distance < 0.1; // Within 100m
            });
            // Update popup content
            const popupContent = `
                    <strong>⚡ Charging Station</strong><br>
                    Location: ${station.name}<br>
                    ID: ${station.id}<br>
                    <hr>
                    <strong>Vehicles Charging (${chargingVehicles.length}):</strong><br>
                    ${chargingVehicles.length > 0
                ? chargingVehicles.map((v) => `• ${v.id} (${v.battery_level}%)`).join('<br>')
                : 'No vehicles currently charging'}
                `;
            // Find the marker and update its popup
            this.chargingStationMarkers.forEach(marker => {
                const markerLatLng = marker.getLatLng();
                if (Math.abs(markerLatLng.lat - station.lat) < 0.001 &&
                    Math.abs(markerLatLng.lng - station.lng) < 0.001) {
                    marker.setPopupContent(popupContent);
                    marker.openPopup();
                }
            });
        })
            .catch(error => console.error('Error fetching vehicles for charging station:', error));
    }
    calculateDistance(lat1, lng1, lat2, lng2) {
        const R = 6371; // Earth's radius in km
        const dLat = (lat2 - lat1) * Math.PI / 180;
        const dLng = (lng2 - lng1) * Math.PI / 180;
        const a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
            Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) *
                Math.sin(dLng / 2) * Math.sin(dLng / 2);
        const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
        return R * c;
    }
    updateVehicleList(vehicles) {
        const vehicleList = document.getElementById('vehicle-list');
        const vehicleCount = document.getElementById('vehicle-count');
        // Update vehicle count
        vehicleCount.textContent = vehicles.length.toString();
        // Sort vehicles by status for better organization: available > busy > charging
        const statusOrder = { 'available': 0, 'busy': 1, 'charging': 2 };
        const sortedVehicles = vehicles.sort((a, b) => {
            const aOrder = statusOrder[a.status] ?? 3;
            const bOrder = statusOrder[b.status] ?? 3;
            if (aOrder !== bOrder)
                return aOrder - bOrder;
            return a.id.localeCompare(b.id); // Secondary sort by ID
        });
        vehicleList.innerHTML = sortedVehicles.map(vehicle => `
            <div class="vehicle-item ${vehicle.status}" onclick="dashboard.focusOnVehicle('${vehicle.id}')">
                <strong>Vehicle ${vehicle.id}</strong> ${vehicle.status === 'charging' ? '⚡' : ''}<br>
                Status: ${vehicle.status}<br>
                Battery: ${vehicle.battery_level}%${vehicle.current_job_id ? `<br>Job: ${vehicle.current_job_id}` : ''}
            </div>
        `).join('');
    }
    updateLastUpdateTime() {
        const now = new Date().toLocaleTimeString();
        document.getElementById('last-update').textContent = `Last update: ${now}`;
    }
    startUpdates() {
        // Initial update
        this.updateDashboard();
        // Set up periodic updates
        setInterval(() => {
            this.updateDashboard();
        }, this.updateInterval);
    }
    focusOnVehicle(vehicleId) {
        const marker = this.vehicleMarkers.get(vehicleId);
        if (marker) {
            this.map.setView(marker.getLatLng(), 16);
            marker.openPopup();
        }
    }
    focusOnJob(jobId) {
        const markers = this.jobMarkers.get(jobId);
        if (markers && markers.length > 0) {
            const firstMarker = markers[0];
            if (firstMarker.getLatLng) {
                this.map.setView(firstMarker.getLatLng(), 15);
                firstMarker.openPopup();
            }
        }
    }
}
// Global dashboard instance for onclick handlers
let dashboard;
// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    dashboard = new FleetDashboard();
});
