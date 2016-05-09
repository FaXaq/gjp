// Copyright 2016 Marin Procureur. All rights reserved.
// Use of gjp source code is governed by a MIT license
// license that can be found in the LICENSE file.

/*
Package gjp stands for Go JobPool, and is willing to be a simple jobpool manager. It maintains
a number of queues determined at the init. No priority whatsoever, just every queues are
processing one job at a time.
*/

package gjp

import (
	"fmt"
	"time"
	"os/exec"
)

/*
   TYPES
*/

type (
	// Job
	Job struct {
		JobRunner
		id []byte
		name          string //Public property retrievable
		status        int    //Status of the current job
		error         *JobError
		start time.Time
		end time.Time
	}
)


/*
   JOB STATUS
*/
const (
	failed  int = 0
	success int = 1
	waiting int = 2
	processing int = 3
)

func newJob (jobRunner JobRunner, jobName string) (job *Job, jobId string){
	//generating uuid for the job
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		fmt.Println("Couldn't generate uuid for new job")
	}

	job = &Job{
		JobRunner: jobRunner,
		name:      jobName,
		status:    waiting,
		id: out,
	}

	jobId = string(job.id[:])

	return
}

//execute the job safely and set the status back for the reportChannel
func (j *Job) executeJob(start time.Time) {
	defer catchPanic("Job", j.name, "failed in executeJob")
	//Set the execution time for this job

	j.start = start
	j.setJobToProcessing()

	defer func() {
		j.end = time.Now()
	}()

	j.error = j.ExecuteJob()

	//Set the job status
	switch j.error {
	case nil:
		j.setJobToSuccess()
		break
	default:
		j.setJobToError()
		break
	}
	return
}
/*
 GETTERS & SETTERS
*/

//create an error well formated
func (j *Job) getJobError() (errorString string) {
	errorString = j.error.fmtError()
	return
}

func (j *Job) getJobStringId() (jobId string) {
	jobId = string(j.id[:])
	return
}

func (j *Job) jobErrored() (jobError bool, error string) {
	if j.status == 0 {
		jobError = true
		error = j.getJobError()
	}
	return
}

func (j *Job) getJobStatus() (jobStatus string) {
	switch j.status {
	case 0:
		jobStatus = "failed"
		break;
	case 1:
		jobStatus = "success"
		break;
	case 2:
		jobStatus = "waiting"
		break;
	case 3:
		jobStatus = "processing"
		break;
	default:
		jobStatus = "error"
		break;
	}
	return
}

func (j *Job) getExecutionTime() (executionTime time.Duration) {
	nullTime := time.Time{}
	if j.end == nullTime {
		executionTime = j.start.Sub(j.end)
	} else {
		executionTime = time.Since(j.start)
	}
	return
}

func (j *Job) getJobInfos() (jobName string, jobId string, jobStatus string) {
	jobName = j.name
	jobId = j.getJobStringId()
	jobStatus = j.getJobStatus()

	return
}

func (j *Job) setJobToWaiting() {
	j.status = waiting
}

func (j *Job) setJobToError() {
	j.status = failed
}

func (j *Job) setJobToSuccess() {
	j.status = success
}

func (j *Job) setJobToProcessing() {
	j.status = processing
}
