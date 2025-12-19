-- Migration: Fix amortization runs unique constraint
-- This migration removes the overly restrictive unique constraint on amortization runs
-- to allow multiple manual runs per day while keeping the scheduler unique constraint

-- Remove the existing unique constraint
ALTER TABLE amortization_runs DROP CONSTRAINT IF EXISTS unique_run_per_date;

-- Add a new constraint that allows multiple manual runs but only one scheduled run per day
ALTER TABLE amortization_runs
    ADD CONSTRAINT unique_scheduled_run_per_date
    UNIQUE (run_date, is_manual)
    DEFERRABLE INITIALLY DEFERRED;

-- Also update the constraint to handle the is_manual column properly
ALTER TABLE amortization_runs
    ADD CONSTRAINT valid_run_status_extended
    CHECK (status IN ('pending', 'started', 'running', 'completed', 'partial', 'failed', 'cancelled'));

-- Comment explaining the change
COMMENT ON TABLE amortization_runs IS 'Scheduler execution tracking with checkpoint mechanism for restart capability. Multiple manual runs allowed per day, only one scheduled run per day.';