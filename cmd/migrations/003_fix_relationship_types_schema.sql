-- Fix relationship_types table schema to match the code expectations

-- Add missing columns
ALTER TABLE relationship_types
ADD COLUMN display_name VARCHAR(100),
ADD COLUMN reverse_label VARCHAR(100) NOT NULL DEFAULT 'relates to',
ADD COLUMN color VARCHAR(20),
ADD COLUMN icon VARCHAR(50),
ADD COLUMN category VARCHAR(50),
ADD COLUMN is_active BOOLEAN DEFAULT true,
ADD COLUMN is_system BOOLEAN DEFAULT false,
ADD COLUMN allowed_source_types TEXT[] DEFAULT '{}',
ADD COLUMN allowed_target_types TEXT[] DEFAULT '{}',
ADD COLUMN cardinality_source VARCHAR(20) DEFAULT 'many',
ADD COLUMN cardinality_target VARCHAR(20) DEFAULT 'many',
ADD COLUMN bidirectional BOOLEAN DEFAULT false,
ADD COLUMN allow_self_reference BOOLEAN DEFAULT false;

-- Rename backward_label to reverse_label (temporarily keep both for compatibility)
ALTER TABLE relationship_types RENAME COLUMN backward_label TO old_backward_label;

-- Update existing data to populate new columns
UPDATE relationship_types SET
    display_name = INITCAP(REPLACE(name, '_', ' ')),
    reverse_label = old_backward_label,
    cardinality_source = 'many',
    cardinality_target = 'many',
    bidirectional = true,
    is_active = true,
    is_system = true,
    allowed_source_types = source_types,
    allowed_target_types = target_types,
    allow_self_reference = allow_same_type;

-- Drop the temporary column
ALTER TABLE relationship_types DROP COLUMN old_backward_label;

-- Add indexes for new columns
CREATE INDEX idx_relationship_types_display_name ON relationship_types(display_name);
CREATE INDEX idx_relationship_types_category ON relationship_types(category);
CREATE INDEX idx_relationship_types_is_active ON relationship_types(is_active);

-- Add constraints
ALTER TABLE relationship_types
ADD CONSTRAINT chk_cardinality_source CHECK (cardinality_source IN ('one', 'many')),
ADD CONSTRAINT chk_cardinality_target CHECK (cardinality_target IN ('one', 'many')),
ADD CONSTRAINT chk_display_name_length CHECK (LENGTH(display_name) >= 2),
ADD CONSTRAINT chk_reverse_label_length CHECK (LENGTH(reverse_label) >= 1);

-- Update default values for new columns
ALTER TABLE relationship_types
ALTER COLUMN display_name SET DEFAULT NULL,
ALTER COLUMN reverse_label SET DEFAULT 'relates to',
ALTER COLUMN color SET DEFAULT NULL,
ALTER COLUMN icon SET DEFAULT NULL,
ALTER COLUMN category SET DEFAULT NULL,
ALTER COLUMN is_active SET DEFAULT true,
ALTER COLUMN is_system SET DEFAULT false,
ALTER COLUMN allowed_source_types SET DEFAULT '{}',
ALTER COLUMN allowed_target_types SET DEFAULT '{}',
ALTER COLUMN cardinality_source SET DEFAULT 'many',
ALTER COLUMN cardinality_target SET DEFAULT 'many',
ALTER COLUMN bidirectional SET DEFAULT false,
ALTER COLUMN allow_self_reference SET DEFAULT false;