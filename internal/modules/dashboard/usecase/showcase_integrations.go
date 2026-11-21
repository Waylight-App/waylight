package usecase

// integrationShowcaseConfigs configures the showcase widgets of example
// integrations, whose answers come from the demo fixtures, keyed by widget
// kind. The stop ID and entity IDs are those of internal/platform/demo; the
// API widget gets its endpoint after the demo created it.
var integrationShowcaseConfigs = map[string]string{
	"transit": `{"stopId":"demo_de:03403:16750","stopName":"Oldenburg Hbf","count":6}`,
	"radar":   `{"lat":53.14,"lon":8.21,"label":"Oldenburg"}`,
	// The example Grafana has no image renderer, so the panel shows its values (#579).
	"grafana-panel": `{"dashboardUid":"homeserver","dashboardTitle":"Home server","panelId":1,"panelTitle":"CPU","range":"6h","refreshMinutes":5}`,
	"smarthome":     `{"entities":["sensor.living_room_temperature","sensor.balcony_solar_power","light.desk_lamp","switch.coffee_machine","sensor.washing_machine_power","binary_sensor.front_door"]}`,
}
