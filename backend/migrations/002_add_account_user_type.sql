-- Add user_type to accounts table for authorization levels
-- Migration: Add user_type ENUM and default values for Account-based authentication

-- Create user_type ENUM type
CREATE TYPE user_type AS ENUM ('system_admin', 'clinic_admin', 'staff');

-- Add user_type column to accounts
ALTER TABLE accounts
ADD COLUMN user_type user_type NOT NULL DEFAULT 'staff';

-- Update existing accounts based on their relationship to staff
-- System admin (ID=1) and clinic admin (ID=2) are configured via seed data
-- ID=1: admin@noavet.jp → system_admin (managed in seed)
-- ID=2: clinic1@noavet.jp → clinic_admin (managed in seed)
-- ID=3+: staff (default)

-- Create index for authorization queries
CREATE INDEX idx_accounts_user_type ON accounts(user_type);
CREATE INDEX idx_accounts_active_user_type ON accounts(is_active, user_type) WHERE deleted_at IS NULL;
