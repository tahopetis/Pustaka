-- Add missing updated_by column to relationship_types table
ALTER TABLE relationship_types
ADD COLUMN updated_by UUID REFERENCES users(id);

-- Create index for updated_by
CREATE INDEX idx_relationship_types_updated_by ON relationship_types(updated_by);