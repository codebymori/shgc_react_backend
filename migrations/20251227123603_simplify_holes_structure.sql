-- Migration: Simplify Holes Structure
-- Date: 2025-12-27
-- Description: Drop hole_details table and add par/distance fields to holes table

-- Step 1: Drop hole_details table
DROP TABLE IF EXISTS hole_details;

-- Step 2: Add new columns to holes table with default values (to handle existing data)
ALTER TABLE holes 
ADD COLUMN IF NOT EXISTS par INT DEFAULT 0,
ADD COLUMN IF NOT EXISTS distance INT DEFAULT 0;

-- Step 3: (IMPORTANT) Update existing holes with actual values
-- You MUST update existing holes before removing defaults!
-- Example:
-- UPDATE holes SET par = 4, distance = 350 WHERE name = 'Hole 1';
-- UPDATE holes SET par = 3, distance = 180 WHERE name = 'Hole 2';

-- Step 4: After updating all existing holes, you can optionally make columns NOT NULL
-- ALTER TABLE holes ALTER COLUMN par SET NOT NULL;
-- ALTER TABLE holes ALTER COLUMN distance SET NOT NULL;
