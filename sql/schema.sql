CREATE TABLE IF NOT EXISTS flights
(
    Icao24 String,
    Callsign String,
    OriginCountry String,
    TimePosition Nullable(Int64),
    LastContact Int64,
    Longitude Nullable(Float64),
    Latitude Nullable(Float64),
    BaroAltitude Nullable(Float64),
    OnGround Bool,
    Velocity Nullable(Float64),
    TrueTrack Nullable(Float64),
    VerticalRate Nullable(Float64),
    Sensors Array(Int32),
    GeoAltitude Nullable(Float64),
    Squawk Nullable(String),
    Spi Bool,
    PositionSource Int32,
    IngestedAt DateTime DEFAULT now()
)
ENGINE = MergeTree
ORDER BY (Icao24, LastContact);
