-- Add unique constraint on tracker_id column
ALTER TABLE issues ADD CONSTRAINT issues_tracker_id_key UNIQUE (tracker_id); 