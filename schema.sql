-- GCIT Ground Booking System — complete schema
-- Run this on a fresh PostgreSQL database:
--   psql -U postgres -d gcit_booking -f schema.sql

-- ============================================================
-- Users
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    student_id VARCHAR(36)  NOT NULL UNIQUE,
    first_name VARCHAR(100) NOT NULL,
    last_name  VARCHAR(100) NOT NULL,
    phone      VARCHAR(20)  NOT NULL,
    email      VARCHAR(255) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL
);

-- ============================================================
-- Admins
-- ============================================================
CREATE TABLE IF NOT EXISTS admins (
    id       SERIAL PRIMARY KEY,
    email    VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL
);

-- Seed a default admin (change the password before going live)
INSERT INTO admins (email, password)
VALUES ('admin@gcit.edu.bt', 'admin123')
ON CONFLICT (email) DO NOTHING;

-- ============================================================
-- Profile  (one row per user, keyed by email)
-- ============================================================
CREATE TABLE IF NOT EXISTS profile (
    id         SERIAL PRIMARY KEY,
    email      VARCHAR(255) NOT NULL UNIQUE REFERENCES users(email) ON DELETE CASCADE,
    department VARCHAR(100) NOT NULL,
    phone      BIGINT       NOT NULL
);

-- ============================================================
-- Slots  (fixed time windows admin can define)
-- ============================================================
CREATE TABLE IF NOT EXISTS slots (
    id         SERIAL PRIMARY KEY,
    start_time VARCHAR(5) NOT NULL,   -- "HH:MM"
    end_time   VARCHAR(5) NOT NULL,   -- "HH:MM"
    status     VARCHAR(20) NOT NULL DEFAULT 'available'
);

-- Seed default slots (06:00 – 22:00 in 2-hour blocks)
INSERT INTO slots (start_time, end_time, status) VALUES
    ('06:00', '08:00', 'available'),
    ('08:00', '10:00', 'available'),
    ('10:00', '12:00', 'available'),
    ('12:00', '14:00', 'available'),
    ('14:00', '16:00', 'available'),
    ('16:00', '18:00', 'available'),
    ('18:00', '20:00', 'available'),
    ('20:00', '22:00', 'available')
ON CONFLICT DO NOTHING;

-- ============================================================
-- Bookings
-- ============================================================
CREATE TABLE IF NOT EXISTS booking (
    id             SERIAL PRIMARY KEY,
    student_id     VARCHAR(36)  NOT NULL,
    match_type     VARCHAR(50)  NOT NULL,
    date           DATE         NOT NULL,
    starting_time  VARCHAR(5)   NOT NULL,   -- "HH:MM"
    ending_time    VARCHAR(5)   NOT NULL,   -- "HH:MM"
    notes          TEXT,
    status         VARCHAR(20)  NOT NULL DEFAULT 'pending',
    created_at     TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_booking_date        ON booking (date);
CREATE INDEX IF NOT EXISTS idx_booking_student_id  ON booking (student_id);
CREATE INDEX IF NOT EXISTS idx_booking_status      ON booking (status);
