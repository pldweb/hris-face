-- Geofence radius per work location (docs/PRD.md 3.1 originally cancelled this
-- for desktop-only attendance -- browser Geolocation on desktop is WiFi/IP
-- position, 20m-few km accurate, spoofable from devtools. Reintroduced for the
-- phone-based deployment, where the device has real GPS. Still a soft signal,
-- not a substitute for the face+liveness check: lat/lng/radius are all
-- nullable, so a location with none set enforces no radius at all.
ALTER TABLE work_locations
    ADD COLUMN lat           double precision,
    ADD COLUMN lng           double precision,
    ADD COLUMN radius_meters integer;
