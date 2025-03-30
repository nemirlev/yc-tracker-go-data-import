-- Create status_types table
CREATE TABLE IF NOT EXISTS status_types (
    id SERIAL PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    tracker_id VARCHAR(255) NOT NULL,
    status_key VARCHAR(255) NOT NULL,
    status_type VARCHAR(255) NOT NULL,
    created_at_db TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at_db TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT status_types_tracker_id_status_key_key UNIQUE (tracker_id, status_key)
);

-- Create index for status_types
CREATE INDEX IF NOT EXISTS idx_status_types_tracker_id ON status_types(tracker_id);
CREATE INDEX IF NOT EXISTS idx_status_types_status_key ON status_types(status_key);

-- Modify v_issue_statuses view to include status types
CREATE OR REPLACE VIEW v_issue_statuses AS
SELECT 
    c.issue_key,
    i.created_at as issue_created,
    c2.updated_at as from_status_timestamp,
    c.updated_at as to_status_timestamp,
    CASE 
        WHEN c.from_display = 'Открыт' AND c2.updated_at IS NULL THEN
            EXTRACT(EPOCH FROM (c.updated_at - i.created_at))/60
        ELSE
            EXTRACT(EPOCH FROM (c.updated_at - c2.updated_at))/60
    END as from_previous_minutes,
    c.from_display as from_status,
    c.to_display as to_status,
    EXTRACT(EPOCH FROM (c.updated_at - i.created_at))/60 as from_created_minutes,
    st_from.status_type as from_status_type,
    st_to.status_type as to_status_type
FROM changelog c
JOIN issues i ON c.issue_key = i.key
LEFT JOIN changelog c2 ON c.issue_key = c2.issue_key
    AND c.type = c2.type
    AND c.field_display = c2.field_display
    AND c.updated_at > c2.updated_at
LEFT JOIN status_types st_from ON c.tracker_id = st_from.tracker_id 
    AND c.from_display = st_from.status_key
LEFT JOIN status_types st_to ON c.tracker_id = st_to.tracker_id 
    AND c.to_display = st_to.status_key
WHERE c.type = 'IssueWorkflow'
    AND c.field_display = 'Статус'
ORDER BY c.issue_key, c.updated_at; 