package opensky

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/learning/flight-intel-pipeline/pkg/models"
)

const openSkyAPIURL = "https://opensky-network.org/api/states/all"

type Client struct {
	httpClient *http.Client
	Username   string
	Password   string
}

func NewClient(username, password string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		Username:   username,
		Password:   password,
	}
}

type openSkyResponse struct {
	Time   int64           `json:"time"`
	States [][]interface{} `json:"states"`
}

func (c *Client) FetchFlights() ([]models.Flight, error) {
	req, err := http.NewRequest(http.MethodGet, openSkyAPIURL, nil)
	if err != nil {
		return nil, err
	}

	// Removing Basic Auth completely because OpenSky anonymous access is sufficient 
	// and provided credentials caused a 401 Unauthorized

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var osResp openSkyResponse
	if err := json.NewDecoder(res.Body).Decode(&osResp); err != nil {
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	var flights []models.Flight
	for _, state := range osResp.States {
		if len(state) < 17 {
			continue
		}

		flight := parseOpenSkyState(state)
		flight.IngestedAt = time.Now()
		flights = append(flights, flight)
	}

	return flights, nil
}

func parseOpenSkyState(state []interface{}) models.Flight {
	var f models.Flight

	if v, ok := state[0].(string); ok { f.Icao24 = v }
	if v, ok := state[1].(string); ok { f.Callsign = strings.TrimSpace(v) }
	if v, ok := state[2].(string); ok { f.OriginCountry = v }
	
	if v, ok := state[3].(float64); ok { 
		val := int64(v)
		f.TimePosition = &val
	}
	if v, ok := state[4].(float64); ok { f.LastContact = int64(v) }
	
	if v, ok := state[5].(float64); ok { f.Longitude = &v }
	if v, ok := state[6].(float64); ok { f.Latitude = &v }
	if v, ok := state[7].(float64); ok { f.BaroAltitude = &v }
	
	if v, ok := state[8].(bool); ok { f.OnGround = v }
	
	if v, ok := state[9].(float64); ok { f.Velocity = &v }
	if v, ok := state[10].(float64); ok { f.TrueTrack = &v }
	if v, ok := state[11].(float64); ok { f.VerticalRate = &v }
	
	// sensors is a bit tricky, it can be []int32 but opensky often returns null or float64 arrays
	if vArr, ok := state[12].([]interface{}); ok {
		var sensors []int32
		for _, s := range vArr {
			if sf, ok := s.(float64); ok {
				sensors = append(sensors, int32(sf))
			}
		}
		f.Sensors = sensors
	}

	if v, ok := state[13].(float64); ok { f.GeoAltitude = &v }
	if v, ok := state[14].(string); ok { f.Squawk = &v }
	if v, ok := state[15].(bool); ok { f.Spi = v }
	if v, ok := state[16].(float64); ok { f.PositionSource = int32(v) }

	return f
}
