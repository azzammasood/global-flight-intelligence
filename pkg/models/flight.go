package models

import "time"

// Flight represents a single aircraft state at a given time.
// Sent over Kafka to ClickHouse.
type Flight struct {
	Icao24         string    `json:"icao24"`
	Callsign       string    `json:"callsign"`
	OriginCountry  string    `json:"origin_country"`
	TimePosition   *int64    `json:"time_position"`
	LastContact    int64     `json:"last_contact"`
	Longitude      *float64  `json:"longitude"`
	Latitude       *float64  `json:"latitude"`
	BaroAltitude   *float64  `json:"baro_altitude"`
	OnGround       bool      `json:"on_ground"`
	Velocity       *float64  `json:"velocity"`
	TrueTrack      *float64  `json:"true_track"`
	VerticalRate   *float64  `json:"vertical_rate"`
	Sensors        []int32   `json:"sensors"`
	GeoAltitude    *float64  `json:"geo_altitude"`
	Squawk         *string   `json:"squawk"`
	Spi            bool      `json:"spi"`
	PositionSource int32     `json:"position_source"`

	IngestedAt time.Time `json:"ingested_at"`
}
