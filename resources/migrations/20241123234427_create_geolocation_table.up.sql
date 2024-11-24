CREATE TABLE IF NOT EXISTS "geolocation" (
    id serial PRIMARY KEY,

    ip_address      INET NOT NULL,
    country_code    CHAR(2) NOT NULL,
    country         VARCHAR(100) NOT NULL,
    city            VARCHAR(100) NOT NULL,
    latitude        NUMERIC(17, 15) NOT NULL,
    longitude       NUMERIC(18, 15) NOT NULL,
    mystery_value   BIGINT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);