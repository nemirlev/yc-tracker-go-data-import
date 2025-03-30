package repository

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nemirlev/yc-tracker-go-data-import/internal/domain"
	"github.com/nemirlev/yc-tracker-go-data-import/pkg/tracker"
)

// Service represents the repository service
type Service struct {
	db *pgxpool.Pool
}

// NewService creates a new repository service
func NewService(db *pgxpool.Pool) domain.Repository {
	return &Service{db: db}
}

// SaveIssues saves issues to the database
func (s *Service) SaveIssues(ctx context.Context, issues []tracker.Issue) error {
	slog.Info("Starting save issues", "total_issues", len(issues))

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Prepare the insert statement
	query := `
		INSERT INTO issues (
			organization_id, self, tracker_id, key, version, story_points,
			summary, status_start_time, boards_names, created_at,
			comment_without_external_message_count, votes,
			comment_with_external_message_count, deadline, updated_at,
			favorite, updated_by_display, type_display, priority_display,
			created_by_display, assignee_display, queue_key, queue_display,
			status_display, previous_status_display, parent_key, parent_display,
			components_display, sprint_display, epic_display,
			previous_status_last_assignee_display, original_estimation, spent,
			tags, estimation, checklist_done, checklist_total, email_created_by,
			sla, email_to, email_from, last_comment_updated_at, followers,
			pending_reply_from, end_time, start_time, project_display,
			voted_by_display, aliases, previous_queue_display, access,
			resolved_at, resolved_by_display, resolution_display,
			last_queue_display, status_type, team_number
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
			$31, $32, $33, $34, $35, $36, $37, $38, $39, $40,
			$41, $42, $43, $44, $45, $46, $47, $48, $49, $50, $51,
			$52, $53, $54, $55, $56, $57
		) ON CONFLICT (tracker_id) DO UPDATE SET
			updated_at = EXCLUDED.updated_at,
			updated_at_db = CURRENT_TIMESTAMP,
			team_number = EXCLUDED.team_number
	`

	_, err = tx.Prepare(ctx, "insert_issue", query)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}

	// Insert each issue
	for _, issue := range issues {

		params := []interface{}{
			issue.OrganizationID,
			issue.Self,
			issue.ID,
			issue.Key,
			func() interface{} {
				if issue.Version == 0 {
					return nil
				}
				return issue.Version
			}(),
			func() interface{} {
				if issue.StoryPoints == 0 {
					return nil
				}
				return issue.StoryPoints
			}(),
			issue.Summary,
			func() interface{} {
				if issue.StatusStartTime.Time().IsZero() {
					return nil
				}
				return issue.StatusStartTime.Time()
			}(),
			strings.Join(getBoardNames(issue.Boards), ", "),
			func() interface{} {
				if issue.CreatedAt.Time().IsZero() {
					return nil
				}
				return issue.CreatedAt.Time()
			}(),
			func() interface{} {
				if issue.CommentWithoutExternalMessageCount == 0 {
					return nil
				}
				return issue.CommentWithoutExternalMessageCount
			}(),
			func() interface{} {
				if issue.Votes == 0 {
					return nil
				}
				return issue.Votes
			}(),
			func() interface{} {
				if issue.CommentWithExternalMessageCount == 0 {
					return nil
				}
				return issue.CommentWithExternalMessageCount
			}(),
			func() interface{} {
				if issue.Deadline.Time().IsZero() {
					return nil
				}
				return issue.Deadline.Time()
			}(),
			func() interface{} {
				if issue.UpdatedAt.Time().IsZero() {
					return nil
				}
				return issue.UpdatedAt.Time()
			}(),
			issue.Favorite,
			issue.UpdatedBy.Display,
			issue.Type.Display,
			issue.Priority.Display,
			issue.CreatedBy.Display,
			issue.Assignee.Display,
			issue.Queue.Key,
			issue.Queue.Display,
			issue.Status.Display,
			issue.PreviousStatus.Display,
			issue.Parent.Key,
			issue.Parent.Display,
			strings.Join(getEntityDisplays(issue.Components), ", "),
			strings.Join(getEntityDisplays(issue.Sprint), ", "),
			issue.Epic.Display,
			issue.PreviousStatusLastAssignee.Display,
			func() interface{} {
				if issue.OriginalEstimation == 0 {
					return nil
				}
				return issue.OriginalEstimation
			}(),
			func() interface{} {
				if issue.Spent == "" {
					return nil
				}
				return issue.Spent
			}(),
			strings.Join(issue.Tags, ", "),
			func() interface{} {
				if issue.Estimation == "" {
					return nil
				}
				return issue.Estimation
			}(),
			func() interface{} {
				if issue.ChecklistDone == 0 {
					return nil
				}
				return issue.ChecklistDone
			}(),
			func() interface{} {
				if issue.ChecklistTotal == 0 {
					return nil
				}
				return issue.ChecklistTotal
			}(),
			issue.EmailCreatedBy,
			strings.Join(getEntityDisplays(issue.SLA), ", "),
			issue.EmailTo,
			issue.EmailFrom,
			func() interface{} {
				if issue.LastCommentUpdatedAt.Time().IsZero() {
					return nil
				}
				return issue.LastCommentUpdatedAt.Time()
			}(),
			strings.Join(getUserDisplays(issue.Followers), ", "),
			issue.PendingReplyFrom,
			func() interface{} {
				if issue.End == "" {
					return nil
				}
				return issue.End
			}(),
			func() interface{} {
				if issue.Start == "" {
					return nil
				}
				return issue.Start
			}(),
			issue.Project.Display,
			issue.VotedBy.Display,
			strings.Join(issue.Aliases, ", "),
			issue.PreviousQueue.Display,
			strings.Join(getEntityDisplays(issue.Access), ", "),
			func() interface{} {
				if issue.ResolvedAt == "" {
					return nil
				}
				return issue.ResolvedAt
			}(),
			issue.ResolvedBy.Display,
			issue.Resolution.Display,
			issue.LastQueue.Display,
			issue.StatusType.Display,
			issue.TeamNumber,
		}

		_, err = tx.Exec(ctx, "insert_issue", params...)
		if err != nil {
			return fmt.Errorf("failed to insert issue %s: %w", issue.Key, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("Successfully saved all issues", "total_issues", len(issues))
	return nil
}

// SaveChangelogs saves changelog entries to the database
func (s *Service) SaveChangelogs(ctx context.Context, changelogs []tracker.Changelog) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Prepare the insert statement
	_, err = tx.Prepare(ctx, "insert_changelog", `
		INSERT INTO changelog (
			organization_id, tracker_id, issue_key, updated_at,
			updated_by_display, type, field_display, from_display,
			to_display, worklog
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) ON CONFLICT (tracker_id, field_display) DO UPDATE SET
			updated_at = EXCLUDED.updated_at,
			updated_by_display = EXCLUDED.updated_by_display,
			from_display = EXCLUDED.from_display,
			to_display = EXCLUDED.to_display,
			worklog = EXCLUDED.worklog
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}

	// Insert each changelog entry
	for _, changelog := range changelogs {
		_, err = tx.Exec(ctx, "insert_changelog",
			changelog.OrganizationID,
			changelog.ID,
			changelog.IssueKey,
			changelog.UpdatedAt,
			changelog.UpdatedByDisplay,
			changelog.Type,
			changelog.FieldDisplay,
			changelog.FromDisplay,
			changelog.ToDisplay,
			changelog.Worklog,
		)
		if err != nil {
			return fmt.Errorf("failed to insert changelog for issue %s: %w", changelog.IssueKey, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetLastUpdateTime returns the timestamp of the last update for a given issue
func (s *Service) GetLastUpdateTime(ctx context.Context, issueKey string) (*time.Time, error) {
	var lastUpdate time.Time
	err := s.db.QueryRow(ctx, `
		SELECT updated_at
		FROM issues
		WHERE key = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`, issueKey).Scan(&lastUpdate)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last update time: %w", err)
	}
	return &lastUpdate, nil
}

// SaveStatusTypes saves status types to the database
func (s *Service) SaveStatusTypes(ctx context.Context, statusTypes []tracker.StatusType) error {
	slog.Info("Starting save status types", "total_status_types", len(statusTypes))

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Prepare the insert statement
	query := `
		INSERT INTO status_types (
			organization_id, tracker_id, status_key, status_type, status_name
		) VALUES (
			$1, $2, $3, $4, $5
		) ON CONFLICT (tracker_id, status_key) DO UPDATE SET
			status_type = EXCLUDED.status_type,
			status_name = EXCLUDED.status_name,
			updated_at_db = CURRENT_TIMESTAMP
	`

	_, err = tx.Prepare(ctx, "insert_status_type", query)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}

	// Insert each status type
	for _, st := range statusTypes {
		params := []interface{}{
			st.OrganizationID,
			st.TrackerID,
			st.Key,
			st.Type,
			st.Name,
		}

		_, err = tx.Exec(ctx, "insert_status_type", params...)
		if err != nil {
			return fmt.Errorf("failed to insert status type %s: %w", st.Key, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("Successfully saved all status types", "total_status_types", len(statusTypes))
	return nil
}

// Helper functions to get display values
func getEntityDisplays(entities []tracker.Entity) []string {
	var displays []string
	for _, e := range entities {
		displays = append(displays, e.Display)
	}
	return displays
}

func getUserDisplays(users []tracker.User) []string {
	var displays []string
	for _, u := range users {
		displays = append(displays, u.Display)
	}
	return displays
}

// Helper function to get board names
func getBoardNames(boards []struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}) []string {
	var names []string
	for _, b := range boards {
		names = append(names, b.Name)
	}
	return names
}
