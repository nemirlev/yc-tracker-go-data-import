package tracker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nemirlev/yc-tracker-go-data-import/internal/config"
	"golang.org/x/time/rate"
)

type Service struct {
	cfg    *config.Config
	client *http.Client
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Time represents a custom time type that can handle various time formats
type Time time.Time

// UnmarshalJSON implements the json.Unmarshaler interface
func (t *Time) UnmarshalJSON(data []byte) error {
	// Remove quotes from the input string
	str := string(data)
	str = strings.Trim(str, "\"")

	if str == "" || str == "null" {
		return nil
	}

	// Try parsing as RFC3339 first
	parsed, err := time.Parse(time.RFC3339, str)
	if err == nil {
		*t = Time(parsed.UTC())
		return nil
	}

	// Try parsing tracker's format (2025-03-29T19:16:33.418+0000)
	parsed, err = time.Parse("2006-01-02T15:04:05.999-0700", str)
	if err == nil {
		*t = Time(parsed.UTC())
		return nil
	}

	// If that fails, try parsing as date only
	parsed, err = time.Parse("2006-01-02", str)
	if err == nil {
		*t = Time(parsed.UTC())
		return nil
	}

	// If all attempts fail, return the error
	return fmt.Errorf("failed to parse time %q: unsupported format", str)
}

// Time returns the underlying time.Time value
func (t Time) Time() time.Time {
	return time.Time(t)
}

// FromTime creates a Time from a time.Time
func FromTime(t time.Time) Time {
	return Time(t)
}

// String returns the string representation of the time value
func (t Time) String() string {
	return time.Time(t).Format(time.RFC3339)
}

// User represents a Tracker user
type User struct {
	Self        string `json:"self"`
	ID          string `json:"id"`
	Display     string `json:"display"`
	CloudUID    string `json:"cloudUid"`
	PassportUID int64  `json:"passportUid"`
}

// Entity represents a Tracker entity with Self, ID, Key and Display fields
type Entity struct {
	Self    string `json:"self"`
	ID      string `json:"id"`
	Key     string `json:"key,omitempty"`
	Display string `json:"display"`
}

// Issue represents a Tracker issue
type Issue struct {
	Self            string   `json:"self"`
	ID              string   `json:"id"`
	Key             string   `json:"key"`
	Version         int      `json:"version"`
	StatusStartTime Time     `json:"statusStartTime"`
	Aliases         []string `json:"aliases"`
	PreviousQueue   Entity   `json:"previousQueue"`
	StatusType      Entity   `json:"statusType"`
	Sprint          []Entity `json:"sprint"`
	ResolvedBy      User     `json:"resolvedBy"`
	Project         Entity   `json:"project"`
	Description     string   `json:"description"`
	Boards          []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"boards"`
	Type                               Entity   `json:"type"`
	Resolution                         Entity   `json:"resolution"`
	PreviousStatusLastAssignee         User     `json:"previousStatusLastAssignee"`
	CreatedAt                          Time     `json:"createdAt"`
	TeamNumber                         string   `json:"teamNumber"`
	CommentWithExternalMessageCount    int      `json:"commentWithExternalMessageCount"`
	UpdatedAt                          Time     `json:"updatedAt"`
	LastCommentUpdatedAt               Time     `json:"lastCommentUpdatedAt"`
	LastQueue                          Entity   `json:"lastQueue"`
	Summary                            string   `json:"summary"`
	UpdatedBy                          User     `json:"updatedBy"`
	ResolvedAt                         string   `json:"resolvedAt"`
	Start                              string   `json:"start"`
	QAEngineer                         User     `json:"qaEngineer"`
	Priority                           Entity   `json:"priority"`
	TypeOfWork                         string   `json:"typeOfWork"`
	Tags                               []string `json:"tags"`
	Environment                        []string `json:"environment"`
	Followers                          []User   `json:"followers"`
	CreatedBy                          User     `json:"createdBy"`
	CommentWithoutExternalMessageCount int      `json:"commentWithoutExternalMessageCount"`
	Votes                              int      `json:"votes"`
	Assignee                           User     `json:"assignee"`
	Queue                              Entity   `json:"queue"`
	Status                             Entity   `json:"status"`
	PreviousStatus                     Entity   `json:"previousStatus"`
	Transitions                        []Entity `json:"transitions"`
	Favorite                           bool     `json:"favorite"`

	// Additional fields for storage
	OrganizationID     string   `json:"organization_id"`
	StoryPoints        float64  `json:"story_points"`
	BoardsNames        string   `json:"boards_names"`
	Deadline           Time     `json:"deadline"`
	Parent             Entity   `json:"parent"`
	Components         []Entity `json:"components"`
	Epic               Entity   `json:"epic"`
	OriginalEstimation float64  `json:"original_estimation"`
	Spent              string   `json:"spent"`
	Estimation         string   `json:"estimation"`
	ChecklistDone      int      `json:"checklist_done"`
	ChecklistTotal     int      `json:"checklist_total"`
	EmailCreatedBy     string   `json:"email_created_by"`
	SLA                []Entity `json:"sla"`
	EmailTo            string   `json:"email_to"`
	EmailFrom          string   `json:"email_from"`
	PendingReplyFrom   string   `json:"pending_reply_from"`
	End                string   `json:"end"`
	VotedBy            User     `json:"voted_by"`
	Access             []Entity `json:"access"`
}

// Changelog represents the data we store in the database
type Changelog struct {
	OrganizationID   string `json:"organization_id"`
	ID               string `json:"id"`
	IssueKey         string `json:"issueKey"`
	UpdatedAt        Time   `json:"updatedAt"`
	UpdatedByDisplay string `json:"updatedBy_display"`
	Type             string `json:"type"`
	FieldDisplay     string `json:"field_display"`
	FromDisplay      string `json:"from_display"`
	ToDisplay        string `json:"to_display"`
	Worklog          string `json:"worklog"`
}

// Field represents a field in a changelog entry
type Field struct {
	Self    string `json:"self"`
	ID      string `json:"id"`
	Key     string `json:"key,omitempty"`
	Display string `json:"display"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (f *Field) UnmarshalJSON(data []byte) error {
	type Alias Field
	aux := &struct {
		ID interface{} `json:"id"`
		*Alias
	}{
		Alias: (*Alias)(f),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert ID to string if it's a number
	if id, ok := aux.ID.(float64); ok {
		f.ID = strconv.FormatFloat(id, 'f', -1, 64)
	} else if id, ok := aux.ID.(int); ok {
		f.ID = strconv.Itoa(id)
	} else if id, ok := aux.ID.(string); ok {
		f.ID = id
	}

	return nil
}

// ChangelogComment represents a comment in a changelog entry
type ChangelogComment struct {
	Self    string `json:"self"`
	ID      string `json:"id"`
	Display string `json:"display"`
}

// ChangelogComments represents comments in a changelog entry
type ChangelogComments struct {
	Added []ChangelogComment `json:"added"`
}

// ChangelogTrigger represents a trigger in a changelog entry
type ChangelogTrigger struct {
	Trigger struct {
		Self    string `json:"self"`
		ID      string `json:"id"`
		Display string `json:"display"`
	} `json:"trigger"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Board represents a board in a changelog field
type Board struct {
	ID int `json:"id"`
}

// ChangelogField represents a field change in a changelog entry
type ChangelogField struct {
	Field struct {
		Self    string `json:"self"`
		ID      string `json:"id"`
		Display string `json:"display"`
	} `json:"field"`
	From interface{} `json:"from"`
	To   interface{} `json:"to"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (f *ChangelogField) UnmarshalJSON(data []byte) error {
	type Alias ChangelogField
	aux := &struct {
		Field struct {
			Self    string `json:"self"`
			ID      string `json:"id"`
			Display string `json:"display"`
		} `json:"field"`
		From json.RawMessage `json:"from"`
		To   json.RawMessage `json:"to"`
		*Alias
	}{
		Alias: (*Alias)(f),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	f.Field = aux.Field

	// Handle From field
	if len(aux.From) > 0 {
		if aux.From[0] == '[' {
			// Try to unmarshal as array of boards first
			var boards []Board
			if err := json.Unmarshal(aux.From, &boards); err == nil {
				f.From = boards
			} else {
				// If not boards, try as array of users
				var users []User
				if err := json.Unmarshal(aux.From, &users); err != nil {
					return err
				}
				f.From = users
			}
		} else if aux.From[0] == '{' {
			// Entity object
			var entity Entity
			if err := json.Unmarshal(aux.From, &entity); err != nil {
				// If it's not an Entity, try as a string
				var str string
				if err := json.Unmarshal(aux.From, &str); err != nil {
					return err
				}
				f.From = str
			} else {
				f.From = entity
			}
		} else if aux.From[0] == '"' {
			// String
			var str string
			if err := json.Unmarshal(aux.From, &str); err != nil {
				return err
			}
			f.From = str
		} else if string(aux.From) == "null" {
			f.From = nil
		}
	}

	// Handle To field
	if len(aux.To) > 0 {
		if aux.To[0] == '[' {
			// Try to unmarshal as array of boards first
			var boards []Board
			if err := json.Unmarshal(aux.To, &boards); err == nil {
				f.To = boards
			} else {
				// If not boards, try as array of users
				var users []User
				if err := json.Unmarshal(aux.To, &users); err != nil {
					return err
				}
				f.To = users
			}
		} else if aux.To[0] == '{' {
			// Entity object
			var entity Entity
			if err := json.Unmarshal(aux.To, &entity); err != nil {
				// If it's not an Entity, try as a string
				var str string
				if err := json.Unmarshal(aux.To, &str); err != nil {
					return err
				}
				f.To = str
			} else {
				f.To = entity
			}
		} else if aux.To[0] == '"' {
			// String
			var str string
			if err := json.Unmarshal(aux.To, &str); err != nil {
				return err
			}
			f.To = str
		} else if string(aux.To) == "null" {
			f.To = nil
		}
	}

	return nil
}

// ChangelogEntry represents a single changelog entry from the API
type ChangelogEntry struct {
	ID    string `json:"id"`
	Issue struct {
		Key string `json:"key"`
	} `json:"issue"`
	UpdatedAt Time `json:"updatedAt"`
	UpdatedBy struct {
		Display string `json:"display"`
	} `json:"updatedBy"`
	Type   string `json:"type"`
	Fields []struct {
		Field struct {
			Display string `json:"display"`
		} `json:"field"`
		From interface{} `json:"from"`
		To   interface{} `json:"to"`
	} `json:"fields"`
}

// StatusType represents a status type in Tracker
type StatusType struct {
	Self           string `json:"self"`
	ID             int    `json:"id"`
	Version        int    `json:"version"`
	Key            string `json:"key"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Order          int    `json:"order"`
	Type           string `json:"type"`
	OrganizationID string `json:"organization_id"`
	TrackerID      string `json:"tracker_id"`
}

// GetStatusTypes fetches all status types from the Tracker API
func (s *Service) GetStatusTypes(ctx context.Context) ([]StatusType, error) {
	url := fmt.Sprintf("%s/statuses/", s.cfg.Tracker.APIIssuesURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", s.cfg.Tracker.OAuthToken))
	req.Header.Set("X-Org-ID", s.cfg.Tracker.OrgID)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var statusTypes []StatusType
	if err := json.NewDecoder(resp.Body).Decode(&statusTypes); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Set organization and tracker IDs
	for i := range statusTypes {
		statusTypes[i].OrganizationID = s.cfg.Tracker.OrgID
		statusTypes[i].TrackerID = strconv.Itoa(statusTypes[i].ID)
	}

	return statusTypes, nil
}

// GetIssuesCount returns the number of issues matching the filter
func (s *Service) GetIssuesCount(ctx context.Context, query string) (int, error) {
	url := fmt.Sprintf("%s/issues/_count", s.cfg.Tracker.APIIssuesURL)

	reqBody := map[string]string{"query": query}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Org-ID", s.cfg.Tracker.OrgID)
	req.Header.Set("Authorization", "OAuth "+s.cfg.Tracker.OAuthToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("tracker API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var count int
	if err := json.NewDecoder(resp.Body).Decode(&count); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return count, nil
}

// GetIssues retrieves all issues from Tracker using scroll API
func (s *Service) GetIssues(ctx context.Context, query string) ([]Issue, error) {
	slog.Info("Starting getting issues", "query", query)

	// If InitialHistoryDepth is set, modify the query to limit history
	if s.cfg.Tracker.InitialHistoryDepth != "" {
		// Add time filter to the query
		timeFilter := fmt.Sprintf("updated: >now()-%s", s.cfg.Tracker.InitialHistoryDepth)
		if query != "" {
			query = fmt.Sprintf("%s %s", timeFilter, query)
		} else {
			query = timeFilter
		}
		slog.Info("Modified query with history depth",
			"original_query", query,
			"history_depth", s.cfg.Tracker.InitialHistoryDepth)
	}

	url := fmt.Sprintf("%s/issues/_search?scrollType=unsorted&perScroll=500&scrollTTLMillis=60000", s.cfg.Tracker.APIIssuesURL)

	reqBody := map[string]string{"query": query}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Org-ID", s.cfg.Tracker.OrgID)
	req.Header.Set("Authorization", "OAuth "+s.cfg.Tracker.OAuthToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tracker API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var issues []Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Get total count from header
	totalCount, err := strconv.Atoi(resp.Header.Get("X-Total-Count"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse total count: %w", err)
	}

	slog.Info("Initial response received",
		"issues_count", len(issues),
		"total_count", totalCount)

	// Continue fetching if we haven't got all records
	for len(issues) < totalCount {
		scrollID := resp.Header.Get("X-Scroll-Id")
		scrollToken := resp.Header.Get("X-Scroll-Token")
		scrollURL := fmt.Sprintf("%s/issues/_search?scrollId=%s&scrollToken=%s", s.cfg.Tracker.APIIssuesURL, scrollID, scrollToken)

		req, err = http.NewRequestWithContext(ctx, "POST", scrollURL, strings.NewReader(string(bodyBytes)))
		if err != nil {
			return nil, fmt.Errorf("failed to create scroll request: %w", err)
		}

		req.Header.Set("X-Org-ID", s.cfg.Tracker.OrgID)
		req.Header.Set("Authorization", "OAuth "+s.cfg.Tracker.OAuthToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err = s.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send scroll request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("tracker API error during scroll: status=%d, body=%s", resp.StatusCode, string(body))
		}

		var moreIssues []Issue
		if err := json.NewDecoder(resp.Body).Decode(&moreIssues); err != nil {
			return nil, fmt.Errorf("failed to decode scroll response: %w", err)
		}

		issues = append(issues, moreIssues...)
		slog.Info("Scroll response received",
			"current_count", len(issues),
			"total_count", totalCount)
	}

	slog.Info("Successfully retrieved all issues",
		"total_issues", len(issues),
		"query", query)
	return issues, nil
}

// GetChangelogsConcurrently retrieves changelogs for multiple issues in parallel with rate limiting
func (s *Service) GetChangelogsConcurrently(ctx context.Context, issues []Issue) ([]Changelog, error) {
	var wg sync.WaitGroup
	changelogChan := make(chan []Changelog, len(issues))
	errChan := make(chan error, len(issues))
	progressChan := make(chan string, len(issues))

	// Create a rate limiter that allows 10 requests per second
	limiter := rate.NewLimiter(rate.Limit(20), 5)

	// Start progress tracking goroutine
	go func() {
		processed := 0
		total := len(issues)
		for range progressChan {
			processed++
			if processed%10 == 0 || processed == total {
				slog.Debug("Progress update",
					"processed", processed,
					"total", total,
					"percentage", (processed*100)/total)
			}
		}
	}()

	// Start fetching changelogs
	for _, issue := range issues {
		wg.Add(1)
		go func(issueKey string) {
			defer wg.Done()
			defer func() { progressChan <- issueKey }()

			maxRetries := 3
			retryDelay := 2 * time.Second
			var lastErr error

			for attempt := 1; attempt <= maxRetries; attempt++ {
				// Wait for rate limiter
				if err := limiter.Wait(ctx); err != nil {
					lastErr = fmt.Errorf("rate limiter error for issue %s: %w", issueKey, err)
					continue
				}

				// Fetch changelog for the issue
				url := fmt.Sprintf("%s/issues/%s/changelog?type=IssueWorkflow", s.cfg.Tracker.APIIssuesURL, issueKey)

				req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				if err != nil {
					lastErr = fmt.Errorf("failed to create request for %s: %w", issueKey, err)
					continue
				}

				req.Header.Set("X-Org-ID", s.cfg.Tracker.OrgID)
				req.Header.Set("Authorization", "OAuth "+s.cfg.Tracker.OAuthToken)

				resp, err := s.client.Do(req)
				if err != nil {
					lastErr = fmt.Errorf("failed to send request for %s: %w", issueKey, err)
					continue
				}

				if resp.StatusCode == http.StatusTooManyRequests {
					resp.Body.Close()
					slog.Warn("Rate limit exceeded, retrying",
						"issue_key", issueKey,
						"attempt", attempt,
						"max_retries", maxRetries)
					time.Sleep(retryDelay * time.Duration(attempt))
					continue
				}

				if resp.StatusCode != http.StatusOK {
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					lastErr = fmt.Errorf("tracker API error for %s: status=%d, body=%s", issueKey, resp.StatusCode, string(body))
					continue
				}

				var entries []ChangelogEntry
				if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
					resp.Body.Close()
					lastErr = fmt.Errorf("failed to decode response for %s: %w", issueKey, err)
					continue
				}
				resp.Body.Close()

				var changelogs []Changelog
				for _, entry := range entries {
					for _, field := range entry.Fields {
						// Skip if field is empty
						if field.Field.Display == "" {
							continue
						}
						changelog := Changelog{
							OrganizationID:   s.cfg.Tracker.OrgID,
							ID:               entry.ID,
							IssueKey:         entry.Issue.Key,
							UpdatedAt:        entry.UpdatedAt,
							UpdatedByDisplay: entry.UpdatedBy.Display,
							Type:             entry.Type,
							FieldDisplay:     field.Field.Display,
							FromDisplay:      getDisplayValue(field.From),
							ToDisplay:        getDisplayValue(field.To),
							Worklog:          "",
						}
						changelogs = append(changelogs, changelog)
					}
				}

				changelogChan <- changelogs
				return
			}

			errChan <- fmt.Errorf("max retries exceeded for issue %s: %w", issueKey, lastErr)
		}(issue.Key)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(changelogChan)
	close(errChan)
	close(progressChan)

	// Check for errors
	if err, ok := <-errChan; ok {
		return nil, err
	}

	// Collect all changelogs
	var allChangelogs []Changelog
	for cl := range changelogChan {
		allChangelogs = append(allChangelogs, cl...)
	}

	slog.Info("Finished fetching all changelogs",
		"total_changelogs", len(allChangelogs))

	return allChangelogs, nil
}

// getDisplayValue extracts the display value from various types
func getDisplayValue(v interface{}) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case map[string]interface{}:
		if display, ok := val["display"].(string); ok {
			return display
		}
	case []interface{}:
		var displays []string
		for _, item := range val {
			if m, ok := item.(map[string]interface{}); ok {
				if display, ok := m["display"].(string); ok {
					displays = append(displays, display)
				}
			}
		}
		return strings.Join(displays, ", ")
	}
	return fmt.Sprintf("%v", v)
}
