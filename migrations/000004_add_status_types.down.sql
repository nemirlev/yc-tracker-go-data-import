-- Revert v_issue_statuses view to original state
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
    EXTRACT(EPOCH FROM (c.updated_at - i.created_at))/60 as from_created_minutes
FROM changelog c
JOIN issues i ON c.issue_key = i.key
LEFT JOIN changelog c2 ON c.issue_key = c2.issue_key
    AND c.type = c2.type
    AND c.field_display = c2.field_display
    AND c.updated_at > c2.updated_at
WHERE c.type = 'IssueWorkflow'
    AND c.field_display = 'Статус'
ORDER BY c.issue_key, c.updated_at;

-- Drop indexes
DROP INDEX IF EXISTS idx_status_types_tracker_id;
DROP INDEX IF EXISTS idx_status_types_status_key;

-- Drop status_types table
DROP TABLE IF EXISTS status_types; 