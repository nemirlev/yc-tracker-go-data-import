-- Remove unique constraint on tracker_id column
ALTER TABLE issues DROP CONSTRAINT issues_tracker_id_key; 