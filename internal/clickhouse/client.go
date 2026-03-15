package clickhouse

import (
	"context"
	"fmt"
	"log"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/learning/flight-intel-pipeline/pkg/models"
)

type Client struct {
	conn driver.Conn
}

func NewClient(addr []string) (*Client, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: addr,
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "password",
		},
	})
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &Client{conn: conn}, nil
}

func (c *Client) InsertFlights(ctx context.Context, flights []models.Flight) error {
	if len(flights) == 0 {
		return nil
	}

	batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO flights (Icao24, Callsign, OriginCountry, TimePosition, LastContact, Longitude, Latitude, BaroAltitude, OnGround, Velocity, TrueTrack, VerticalRate, Sensors, GeoAltitude, Squawk, Spi, PositionSource, IngestedAt)")
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, f := range flights {
		err := batch.Append(
			f.Icao24,
			f.Callsign,
			f.OriginCountry,
			f.TimePosition,
			f.LastContact,
			f.Longitude,
			f.Latitude,
			f.BaroAltitude,
			f.OnGround,
			f.Velocity,
			f.TrueTrack,
			f.VerticalRate,
			f.Sensors,
			f.GeoAltitude,
			f.Squawk,
			f.Spi,
			f.PositionSource,
			f.IngestedAt,
		)
		if err != nil {
			log.Printf("Failed to append to batch: %v", err)
			return err
		}
	}

	return batch.Send()
}

func (c *Client) Close() error {
	return c.conn.Close()
}
