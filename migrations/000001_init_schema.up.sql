-- Create issues table
CREATE TABLE IF NOT EXISTS issues (
    id SERIAL PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    self VARCHAR(255),
    tracker_id VARCHAR(255) NOT NULL,
    key VARCHAR(255) NOT NULL,
    version NUMERIC(5),
    story_points DECIMAL(15,2),
    summary TEXT,
    status_start_time TIMESTAMP WITH TIME ZONE,
    boards_names TEXT,
    created_at TIMESTAMP WITH TIME ZONE,
    comment_without_external_message_count DECIMAL(15,2),
    votes DECIMAL(15,2),
    comment_with_external_message_count DECIMAL(15,2),
    deadline TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    favorite BOOLEAN,
    updated_by_display VARCHAR(255),
    type_display VARCHAR(255),
    priority_display VARCHAR(255),
    created_by_display VARCHAR(255),
    assignee_display VARCHAR(255),
    queue_key VARCHAR(255),
    queue_display VARCHAR(255),
    status_display VARCHAR(255),
    previous_status_display VARCHAR(255),
    parent_key VARCHAR(255),
    parent_display VARCHAR(255),
    components_display TEXT,
    sprint_display TEXT,
    epic_display VARCHAR(255),
    previous_status_last_assignee_display VARCHAR(255),
    original_estimation DECIMAL(15,2),
    spent VARCHAR,
    tags TEXT,
    estimation VARCHAR,
    checklist_done DECIMAL(15,2),
    checklist_total DECIMAL(15,2),
    email_created_by VARCHAR(255),
    sla TEXT,
    email_to TEXT,
    email_from TEXT,
    last_comment_updated_at TIMESTAMP WITH TIME ZONE,
    followers TEXT,
    pending_reply_from TEXT,
    end_time TIMESTAMP WITH TIME ZONE,
    start_time TIMESTAMP WITH TIME ZONE,
    project_display VARCHAR(255),
    voted_by_display TEXT,
    aliases TEXT,
    previous_queue_display VARCHAR(255),
    access VARCHAR(255),
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by_display VARCHAR(255),
    resolution_display VARCHAR(255),
    last_queue_display VARCHAR(255),
    status_type VARCHAR(255),
    created_at_db TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at_db TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create changelog table
CREATE TABLE IF NOT EXISTS changelog (
    id SERIAL PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    tracker_id VARCHAR(255) NOT NULL,
    issue_key VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE,
    updated_by_display VARCHAR,
    type VARCHAR(255),
    field_display VARCHAR(255),
    from_display TEXT,
    to_display TEXT,
    worklog TEXT,
    created_at_db TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT changelog_tracker_id_field_display_key UNIQUE (tracker_id, field_display)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_issues_tracker_id ON issues(tracker_id);
CREATE INDEX IF NOT EXISTS idx_issues_key ON issues(key);
CREATE INDEX IF NOT EXISTS idx_changelog_issue_key ON changelog(issue_key);
CREATE INDEX IF NOT EXISTS idx_changelog_tracker_id ON changelog(tracker_id);

-- Create views
CREATE OR REPLACE VIEW v_issues AS
SELECT * FROM (
    SELECT *,
           ROW_NUMBER() OVER (PARTITION BY tracker_id ORDER BY updated_at DESC) as rn
    FROM issues
) t WHERE t.rn = 1;

CREATE OR REPLACE VIEW v_changelog AS
SELECT * FROM (
    SELECT *,
           ROW_NUMBER() OVER (PARTITION BY organization_id, tracker_id, field_display ORDER BY updated_at DESC) as rn
    FROM changelog
) t WHERE t.rn = 1;

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