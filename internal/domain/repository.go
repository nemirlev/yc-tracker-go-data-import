package domain

import (
	"context"
	"time"

	"github.com/nemirlev/yc-tracker-go-data-import/pkg/tracker"
)

// IssueRepository defines the interface for issue storage operations
type IssueRepository interface {
	SaveIssues(ctx context.Context, issues []tracker.Issue) error
	GetLastUpdateTime(ctx context.Context, issueKey string) (*time.Time, error)
}

// ChangelogRepository defines the interface for changelog storage operations
type ChangelogRepository interface {
	SaveChangelogs(ctx context.Context, changelogs []tracker.Changelog) error
}

// StatusTypeRepository defines the interface for status type storage operations
type StatusTypeRepository interface {
	SaveStatusTypes(ctx context.Context, statusTypes []tracker.StatusType) error
}

// Repository combines all repository interfaces
type Repository interface {
	IssueRepository
	ChangelogRepository
	StatusTypeRepository
}
