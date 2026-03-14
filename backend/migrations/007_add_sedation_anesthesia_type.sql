-- Add sedation to anesthesia_type enum
ALTER TYPE anesthesia_type ADD VALUE IF NOT EXISTS 'sedation' AFTER 'local';
