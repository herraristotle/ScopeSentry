package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/Autumn-27/ScopeSentry/internal/scheduler/jobs"
	"github.com/Autumn-27/ScopeSentry/internal/utils/helper"

	"github.com/go-co-op/gocron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// JobHandler is the type of a job handler function
type JobHandler func(params string, nextTime string) error

// Job represents a scheduled task
type Job struct {
	ID          string    `bson:"_id"`
	Name        string    `bson:"name"`
	Schedule    string    `bson:"schedule"` // cron expression
	Handler     string    `bson:"handler"`  // handler function name
	Params      string    `bson:"params"`
	LastRunTime time.Time `bson:"last_run_time"`
	NextRunTime time.Time `bson:"next_run_time"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
	CycleType   string    `bson:"cycle_type"` // cycle type: cron or interval
	Interval    struct {
		Days    int `bson:"days"`
		Hours   int `bson:"hours"`
		Minutes int `bson:"minutes"`
	} `bson:"interval"`
}

// Scheduler represents the scheduled task scheduler
type Scheduler struct {
	scheduler      *gocron.Scheduler
	db             *mongo.Database
	jobsCollection *mongo.Collection
	logger         *zap.Logger
	jobs           map[string]*gocron.Job
	handlers       map[string]JobHandler // registered handler functions
}

var (
	globalScheduler *Scheduler
	once            sync.Once
)

// GetGlobalScheduler returns the global scheduler instance
func GetGlobalScheduler() *Scheduler {
	return globalScheduler
}

// InitializeGlobalScheduler initializes the global scheduler
func InitializeGlobalScheduler() {
	once.Do(func() {
		globalScheduler = NewScheduler()
		globalScheduler.RegisterHandler("scan", func(params string, nextTime string) error {
			return jobs.Scan(params, nextTime)
		})
		globalScheduler.RegisterHandler("page_monitoring", func(params string, nextTime string) error {
			return jobs.CreatePageMonitoringTask()
		})
		globalScheduler.RegisterHandler("server_plugin", func(params string, nextTime string) error {
			return jobs.ServerPluginRunner(params, nextTime)
		})
	})
}

// NewScheduler creates a new scheduler instance
func NewScheduler() *Scheduler {
	loc, err := time.LoadLocation(config.GlobalConfig.System.Timezone)
	if err != nil {
		// Handle the error, e.g. by using the default timezone
		loc = time.Local
	}
	return &Scheduler{
		scheduler:      gocron.NewScheduler(loc),
		db:             mongodb.DB,
		jobsCollection: mongodb.DB.Collection("scheduled"),
		logger:         logger.Logger,
		jobs:           make(map[string]*gocron.Job),
		handlers:       make(map[string]JobHandler),
	}
}

// RegisterHandler registers a job handler function
func (s *Scheduler) RegisterHandler(name string, handler JobHandler) {
	s.handlers[name] = handler
	s.logger.Info("Registered job handler", zap.String("handler_name", name))
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	s.scheduler.StartAsync()
	s.loadJobs()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.scheduler.Stop()
}

// AddJob adds a new job
func (s *Scheduler) AddJob(job *Job) error {
	// Check whether the handler function is registered
	if _, exists := s.handlers[job.Handler]; !exists {
		return fmt.Errorf("handler %s not registered", job.Handler)
	}

	// Check whether the job already exists (in memory)
	if _, exists := s.jobs[job.ID]; exists {
		// Job already exists, return success directly
		return nil
	}

	// Check whether the job already exists in MongoDB
	var existingJob Job
	err := s.jobsCollection.FindOne(context.Background(), bson.M{"_id": job.ID}).Decode(&existingJob)
	if err == nil {
		// Job already exists in the database, return success directly
		return nil
	} else if err != mongo.ErrNoDocuments {
		// Query error (not a "document not found" error)
		return fmt.Errorf("failed to check if job exists: %w", err)
	}

	// Save the job to MongoDB
	job.CreatedAt = helper.GetNowTime()
	job.UpdatedAt = helper.GetNowTime()

	_, err = s.jobsCollection.InsertOne(context.Background(), job)
	if err != nil {
		return err
	}

	var schedJob *gocron.Job
	var err2 error

	// Add the job according to its cycle type
	if job.CycleType == "interval" {
		// Add the job using an interval
		schedJob, err2 = s.scheduler.Every(job.Interval.Days).Days().
			At(fmt.Sprintf("%02d:%02d", job.Interval.Hours, job.Interval.Minutes)).
			Do(s.executeJob, job)
	} else {
		// Add the job using a cron expression
		schedJob, err2 = s.scheduler.Cron(job.Schedule).Do(s.executeJob, job)
	}

	if err2 != nil {
		return err2
	}

	s.jobs[job.ID] = schedJob

	// Add logging
	s.logger.Info("Successfully added new job",
		zap.String("job_id", job.ID),
		zap.String("job_name", job.Name),
		zap.String("cycle_type", job.CycleType),
		zap.Time("next_run_time", schedJob.NextRun()),
	)

	return nil
}

// RemoveJob removes a job
func (s *Scheduler) RemoveJob(jobID string) error {
	// Remove the job from the scheduler
	if job, exists := s.jobs[jobID]; exists {
		s.scheduler.RemoveByReference(job)
		delete(s.jobs, jobID)
	}

	// Delete the job from MongoDB
	_, err := s.jobsCollection.DeleteOne(context.Background(), bson.M{"_id": jobID})
	if err != nil {
		return err
	}
	s.logger.Info("Successfully remove job")
	return nil
}

// UpdateJob updates a job
func (s *Scheduler) UpdateJob(job *Job) error {
	// Check whether the handler function is registered
	if _, exists := s.handlers[job.Handler]; !exists {
		return fmt.Errorf("handler %s not registered", job.Handler)
	}

	job.UpdatedAt = helper.GetNowTime()

	// Update the job in MongoDB
	_, err := s.jobsCollection.ReplaceOne(context.Background(), bson.M{"_id": job.ID}, job)
	if err != nil {
		return err
	}

	// Remove the old job from the scheduler
	if oldJob, exists := s.jobs[job.ID]; exists {
		s.scheduler.RemoveByReference(oldJob)
	}

	// Add the new job to the scheduler
	schedJob, err := s.scheduler.Cron(job.Schedule).Do(s.executeJob, job)
	if err != nil {
		return err
	}

	s.jobs[job.ID] = schedJob
	return nil
}

// loadJobs loads all jobs from MongoDB
func (s *Scheduler) loadJobs() {
	cursor, err := s.jobsCollection.Find(context.Background(), bson.M{})
	if err != nil {
		s.logger.Error("Failed to load jobs from MongoDB", zap.Error(err))
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var job Job
		if err := cursor.Decode(&job); err != nil {
			s.logger.Error("Failed to decode job", zap.Error(err))
			continue
		}

		// Check whether the handler function is registered
		if _, exists := s.handlers[job.Handler]; !exists {
			s.logger.Error("Handler not registered for job",
				zap.String("job_id", job.ID),
				zap.String("handler", job.Handler),
			)
			continue
		}

		schedJob, err := s.scheduler.Cron(job.Schedule).Do(s.executeJob, &job)
		if err != nil {
			s.logger.Error("Failed to schedule job", zap.Error(err))
			continue
		}
		s.jobs[job.ID] = schedJob
	}
}

// executeJob executes a job
func (s *Scheduler) executeJob(job *Job) {
	s.logger.Info("Executing job",
		zap.String("job_id", job.ID),
		zap.String("job_name", job.Name),
		zap.String("handler", job.Handler),
	)

	// Get and execute the handler function
	handler, exists := s.handlers[job.Handler]
	if !exists {
		s.logger.Error("Handler not found",
			zap.String("job_id", job.ID),
			zap.String("handler", job.Handler),
		)
		return
	}

	// Update the last execution time
	job.LastRunTime = helper.GetNowTime()
	job.NextRunTime = s.jobs[job.ID].NextRun()

	// Execute the handler function
	if err := handler(job.Params, helper.FormatTime(job.NextRunTime)); err != nil {
		s.logger.Error("Job execution failed",
			zap.String("job_id", job.ID),
			zap.String("job_name", job.Name),
			zap.Error(err),
		)
	}

	_, err := s.jobsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": job.ID},
		bson.M{
			"$set": bson.M{
				"last_run_time": job.LastRunTime,
				"next_run_time": job.NextRunTime,
			},
		},
	)
	if err != nil {
		s.logger.Error("Failed to update job execution time", zap.Error(err))
	}
}
