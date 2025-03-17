-- Create sequences table
CREATE TABLE IF NOT EXISTS sequences (
    unit_code VARCHAR(10) NOT NULL,
    type VARCHAR(10) NOT NULL,
    provider VARCHAR(20) NOT NULL,
    region VARCHAR(10) NOT NULL,
    environment VARCHAR(10) NOT NULL,
    function VARCHAR(20) NOT NULL,
    current_value INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (unit_code, type, provider, region, environment, function)
);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);