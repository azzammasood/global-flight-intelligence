-- Basic analytics to verify the pipeline

-- 1. Total records ingested
SELECT count(*) AS total_flights FROM flights;

-- 2. Top 10 countries with most active flights
SELECT OriginCountry, count(*) AS active_flights
FROM flights
GROUP BY OriginCountry
ORDER BY active_flights DESC
LIMIT 10;

-- 3. Highest flying aircraft
SELECT Callsign, OriginCountry, GeoAltitude
FROM flights
WHERE GeoAltitude IS NOT NULL
ORDER BY GeoAltitude DESC
LIMIT 5;

-- 4. Fastest moving aircraft
SELECT Callsign, OriginCountry, Velocity
FROM flights
WHERE Velocity IS NOT NULL
ORDER BY Velocity DESC
LIMIT 5;
