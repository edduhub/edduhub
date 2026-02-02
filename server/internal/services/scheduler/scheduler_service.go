package scheduler

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// SchedulerService defines the interface for scheduled task operations
type SchedulerService interface {
	Start() error
	Stop() error
	ScheduleFeeReminders(cronExpression string, job func() error) error
	ScheduleAttendanceAlerts(cronExpression string, job func() error) error
	ScheduleGradeNotifications(cronExpression string, job func() error) error
	ScheduleExamReminders(cronExpression string, job func() error) error
	ScheduleDailyReports(cronExpression string, job func() error) error
	ScheduleWeeklyDigest(cronExpression string, job func() error) error
	ScheduleBackup(cronExpression string, job func() error) error
	AddJob(cronExpression string, job func() error) (int, error)
	RemoveJob(entryID int) error
	GetScheduledJobs() []JobInfo
}

// JobInfo represents information about a scheduled job
type JobInfo struct {
	ID       int       `json:"id"`
	Schedule string    `json:"schedule"`
	NextRun  time.Time `json:"next_run"`
	PrevRun  time.Time `json:"prev_run,omitempty"`
}

// cronSchedulerService implements SchedulerService using robfig/cron
type cronSchedulerService struct {
	cron *cron.Cron
	jobs map[int]cron.EntryID
}

// NewSchedulerService creates a new scheduler service instance
func NewSchedulerService() SchedulerService {
	return &cronSchedulerService{
		cron: cron.New(cron.WithSeconds()),
		jobs: make(map[int]cron.EntryID),
	}
}

// Start begins the scheduler
func (s *cronSchedulerService) Start() error {
	s.cron.Start()
	return nil
}

// Stop halts the scheduler
func (s *cronSchedulerService) Stop() error {
	ctx := s.cron.Stop()
	<-ctx.Done()
	return nil
}

// ScheduleFeeReminders schedules fee reminder jobs
func (s *cronSchedulerService) ScheduleFeeReminders(cronExpression string, job func() error) error {
	if cronExpression == "" {
		cronExpression = "0 9 * * *" // Daily at 9 AM
	}

	_, err := s.cron.AddFunc(cronExpression, func() {
		if err := job(); err != nil {
			fmt.Printf("Fee reminder job failed: %v\n", err)
		}
	})

	return err
}

// ScheduleAttendanceAlerts schedules attendance alert jobs
func (s *cronSchedulerService) ScheduleAttendanceAlerts(cronExpression string, job func() error) error {
	if cronExpression == "" {
		cronExpression = "0 18 * * *" // Daily at 6 PM
	}

	_, err := s.cron.AddFunc(cronExpression, func() {
		if err := job(); err != nil {
			fmt.Printf("Attendance alert job failed: %v\n", err)
		}
	})

	return err
}

// ScheduleGradeNotifications schedules grade notification jobs
func (s *cronSchedulerService) ScheduleGradeNotifications(cronExpression string, job func() error) error {
	if cronExpression == "" {
		cronExpression = "0 */6 * * *" // Every 6 hours
	}

	_, err := s.cron.AddFunc(cronExpression, func() {
		if err := job(); err != nil {
			fmt.Printf("Grade notification job failed: %v\n", err)
		}
	})

	return err
}

// ScheduleExamReminders schedules exam reminder jobs
func (s *cronSchedulerService) ScheduleExamReminders(cronExpression string, job func() error) error {
	if cronExpression == "" {
		cronExpression = "0 8 * * *" // Daily at 8 AM
	}

	_, err := s.cron.AddFunc(cronExpression, func() {
		if err := job(); err != nil {
			fmt.Printf("Exam reminder job failed: %v\n", err)
		}
	})

	return err
}

// ScheduleDailyReports schedules daily report generation jobs
func (s *cronSchedulerService) ScheduleDailyReports(cronExpression string, job func() error) error {
	if cronExpression == "" {
		cronExpression = "0 23 * * *" // Daily at 11 PM
	}

	_, err := s.cron.AddFunc(cronExpression, func() {
		if err := job(); err != nil {
			fmt.Printf("Daily report job failed: %v\n", err)
		}
	})

	return err
}

// ScheduleWeeklyDigest schedules weekly digest jobs
func (s *cronSchedulerService) ScheduleWeeklyDigest(cronExpression string, job func() error) error {
	if cronExpression == "" {
		cronExpression = "0 9 * * 1" // Mondays at 9 AM
	}

	_, err := s.cron.AddFunc(cronExpression, func() {
		if err := job(); err != nil {
			fmt.Printf("Weekly digest job failed: %v\n", err)
		}
	})

	return err
}

// ScheduleBackup schedules database backup jobs
func (s *cronSchedulerService) ScheduleBackup(cronExpression string, job func() error) error {
	if cronExpression == "" {
		cronExpression = "0 2 * * *" // Daily at 2 AM
	}

	_, err := s.cron.AddFunc(cronExpression, func() {
		if err := job(); err != nil {
			fmt.Printf("Backup job failed: %v\n", err)
		}
	})

	return err
}

// AddJob adds a custom scheduled job
func (s *cronSchedulerService) AddJob(cronExpression string, job func() error) (int, error) {
	entryID, err := s.cron.AddFunc(cronExpression, func() {
		if err := job(); err != nil {
			fmt.Printf("Custom job failed: %v\n", err)
		}
	})

	if err != nil {
		return 0, err
	}

	intID := int(entryID)
	s.jobs[intID] = entryID
	return intID, nil
}

// RemoveJob removes a scheduled job
func (s *cronSchedulerService) RemoveJob(entryID int) error {
	if id, exists := s.jobs[entryID]; exists {
		s.cron.Remove(id)
		delete(s.jobs, entryID)
	}
	return nil
}

// GetScheduledJobs returns information about all scheduled jobs
func (s *cronSchedulerService) GetScheduledJobs() []JobInfo {
	var jobs []JobInfo

	for id, entryID := range s.jobs {
		entry := s.cron.Entry(entryID)
		if entry.ID != 0 {
			jobs = append(jobs, JobInfo{
				ID:       id,
				Schedule: fmt.Sprintf("%v", entry.Schedule),
				NextRun:  entry.Next,
				PrevRun:  entry.Prev,
			})
		}
	}

	return jobs
}

// Common cron expressions for reference:
// "0 9 * * *"     - Daily at 9:00 AM
// "0 */6 * * *"   - Every 6 hours
// "0 0 * * 0"     - Weekly on Sunday at midnight
// "0 0 1 * *"     - Monthly on the 1st at midnight
// "*/5 * * * *"   - Every 5 minutes
// "0 8 * * 1-5"   - Weekdays at 8:00 AM
