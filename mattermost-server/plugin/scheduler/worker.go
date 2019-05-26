// Copyright (c) 2018-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package scheduler

import (
	"github.com/mattermost/mattermost-server/app"
	"github.com/mattermost/mattermost-server/jobs"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
)

type Worker struct {
	name      string
	stop      chan bool
	stopped   chan bool
	jobs      chan model.Job
	jobServer *jobs.JobServer
	app       *app.App
}

func (m *PluginsJobInterfaceImpl) MakeWorker() model.Worker {
	worker := Worker{
		name:      "Plugins",
		stop:      make(chan bool, 1),
		stopped:   make(chan bool, 1),
		jobs:      make(chan model.Job),
		jobServer: m.App.Srv.Jobs,
		app:       m.App,
	}

	return &worker
}

func (worker *Worker) Run() {
	mlog.Debug("Worker started", mlog.String("worker", worker.name))

	defer func() {
		mlog.Debug("Worker finished", mlog.String("worker", worker.name))
		worker.stopped <- true
	}()

	for {
		select {
		case <-worker.stop:
			mlog.Debug("Worker received stop signal", mlog.String("worker", worker.name))
			return
		case job := <-worker.jobs:
			mlog.Debug("Worker received a new candidate job.", mlog.String("worker", worker.name))
			worker.DoJob(&job)
		}
	}
}

func (worker *Worker) Stop() {
	mlog.Debug("Worker stopping", mlog.String("worker", worker.name))
	worker.stop <- true
	<-worker.stopped
}

func (worker *Worker) JobChannel() chan<- model.Job {
	return worker.jobs
}

func (worker *Worker) DoJob(job *model.Job) {
	if claimed, err := worker.jobServer.ClaimJob(job); err != nil {
		mlog.Info("Worker experienced an error while trying to claim job",
			mlog.String("worker", worker.name),
			mlog.String("job_id", job.Id),
			mlog.String("error", err.Error()))
		return
	} else if !claimed {
		return
	}

	err := worker.app.DeleteAllExpiredPluginKeys()
	if err == nil {
		mlog.Info("Worker: Job is complete", mlog.String("worker", worker.name), mlog.String("job_id", job.Id))
		worker.setJobSuccess(job)
		return
	} else {
		mlog.Error("Worker: Failed to delete expired keys", mlog.String("worker", worker.name), mlog.String("job_id", job.Id), mlog.String("error", err.Error()))
		worker.setJobError(job, err)
		return
	}
}

func (worker *Worker) setJobSuccess(job *model.Job) {
	if err := worker.app.Srv.Jobs.SetJobSuccess(job); err != nil {
		mlog.Error("Worker: Failed to set success for job", mlog.String("worker", worker.name), mlog.String("job_id", job.Id), mlog.String("error", err.Error()))
		worker.setJobError(job, err)
	}
}

func (worker *Worker) setJobError(job *model.Job, appError *model.AppError) {
	if err := worker.app.Srv.Jobs.SetJobError(job, appError); err != nil {
		mlog.Error("Worker: Failed to set job error", mlog.String("worker", worker.name), mlog.String("job_id", job.Id), mlog.String("error", err.Error()))
	}
}
