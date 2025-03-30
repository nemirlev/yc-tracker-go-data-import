-- Add status_name column to status_types table
ALTER TABLE status_types ADD COLUMN status_name VARCHAR(255);

-- Update status_types table with status names
UPDATE status_types SET status_name = status_key;

-- Make status_name NOT NULL after populating it
ALTER TABLE status_types ALTER COLUMN status_name SET NOT NULL;

-- Update v_issue_statuses view to use status_name instead of status_key
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
LEFT JOIN status_types st_from ON c.from_display = st_from.status_name
LEFT JOIN status_types st_to ON c.to_display = st_to.status_name
WHERE c.type = 'IssueWorkflow'
    AND c.field_display = 'Статус'
ORDER BY c.issue_key, c.updated_at; 