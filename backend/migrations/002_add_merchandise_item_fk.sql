-- Migration: Add merchandise_item_id FK to billing_items and estimate_items (BUG-109)
-- Purpose: Enable dependency checking for merchandise item deletion
-- Created: 2026-04-03

-- Step 1: Add nullable merchandise_item_id column to billing_items
ALTER TABLE billing_items
ADD COLUMN merchandise_item_id BIGINT
REFERENCES merchandise_items(id) ON DELETE SET NULL;

-- Step 2: Add index for FK lookup performance
CREATE INDEX idx_billing_items_merchandise_item_id
ON billing_items(merchandise_item_id)
WHERE deleted_at IS NULL;

-- Step 3: Add nullable merchandise_item_id column to estimate_items
ALTER TABLE estimate_items
ADD COLUMN merchandise_item_id BIGINT
REFERENCES merchandise_items(id) ON DELETE SET NULL;

-- Step 4: Add index for FK lookup performance
-- Note: estimate_items does not have deleted_at column, so index without WHERE clause
CREATE INDEX idx_estimate_items_merchandise_item_id
ON estimate_items(merchandise_item_id);

-- Note: Existing rows will have NULL merchandise_item_id (value-copied from before migration)
-- New rows created after migration will have merchandise_item_id populated
-- This allows the CountUsageByMerchandiseItemID query to work going forward
