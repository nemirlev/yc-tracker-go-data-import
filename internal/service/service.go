package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nemirlev/yc-tracker-go-data-import/internal/config"
	"github.com/nemirlev/yc-tracker-go-data-import/internal/domain"
	"github.com/nemirlev/yc-tracker-go-data-import/pkg/tracker"
)

type Service struct {
	cfg     *config.Config
	tracker *tracker.Service
	storage domain.Repository
	workers int
}

func NewService(cfg *config.Config, storage domain.Repository) *Service {
	return &Service{
		cfg:     cfg,
		tracker: tracker.NewService(cfg),
		storage: storage,
		workers: 5, // Number of concurrent workers for processing issues
	}
}

// Sync synchronizes issues and their changelogs from Tracker to the database
func (s *Service) Sync(ctx context.Context) error {
	// Get all issues from Tracker
	issues, err := s.tracker.GetIssues(ctx, s.cfg.Tracker.Filter)
	if err != nil {
		return fmt.Errorf("failed to get issues from tracker: %w", err)
	}

	slog.Info("Retrieved issues from Tracker", "count", len(issues))

	// Save issues to database
	if err := s.storage.SaveIssues(ctx, issues); err != nil {
		return fmt.Errorf("failed to save issues to database: %w", err)
	}

	slog.Info("Saved issues to database")

	// Get all statuses from Tracker
	statusTypes, err := s.tracker.GetStatusTypes(ctx)
	if err != nil {
		return fmt.Errorf("failed to get status types from tracker: %w", err)
	}

	slog.Info("Retrieved status types from Tracker", "count", len(statusTypes))

	// Save status types to database
	if err := s.storage.SaveStatusTypes(ctx, statusTypes); err != nil {
		return fmt.Errorf("failed to save status types to database: %w", err)
	}

	// Get changelogs concurrently
	changelogEntries, err := s.tracker.GetChangelogsConcurrently(ctx, issues)
	if err != nil {
		return fmt.Errorf("failed to get changelogs concurrently: %w", err)
	}

	// Save changelogs to database
	if err := s.storage.SaveChangelogs(ctx, changelogEntries); err != nil {
		return fmt.Errorf("failed to save changelogs to database: %w", err)
	}

	slog.Info("Successfully synchronized data from Tracker")
	return nil
}

// GetTracker returns the tracker service
func (s *Service) GetTracker() *tracker.Service {
	return s.tracker
}
