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
	"encoding/json"
	"fmt"
	"time"
)

/*
   TYPES
*/

type (
	// Job
	Job struct {
		JobRunner `json:"-"` //skip the field for json
		Id        string     `json:"id"`
		Name      string     `json:"name"`   //Public property retrievable
		Status    string        `json:"status"` //Status of the current job
		Error     *JobError  `json:"-"`
		Start     time.Time  `json:"start"`
		End       time.Time  `json:"end"`
	}
)

/*
   JOB STATUS
*/
const (
	failed     string = "failed"
	success    string = "done"
	waiting    string = "queued"
	processing string = "proceeded"
)

func newJob(id string, jobRunner JobRunner, jobName string) (job *Job, jobId string) {
	job = &Job{
		JobRunner: jobRunner,
		Name:      jobName,
		Status:    waiting,
		Id:        id,
	}

	fmt.Println("New job with Id : ", job.Id)

	jobId = job.Id

	return
}

//execute the job safely and set the status back for the reportChannel
func (j *Job) executeJob(start time.Time) {
	defer catchPanic("Job", j.Name, "failed in executeJob")
	//Set the execution time for this job

	j.Start = start

	j.NotifyStart(j.Id)

	j.setJobToProcessing()

	defer func() {
		j.End = time.Now()
	}()

	j.Error = j.ExecuteJob(j.Id)

	//Set the job status
	switch j.Error {
	case nil:
		j.setJobToSuccess()
		break
	default:
		j.setJobToError()
		break
	}

	j.NotifyEnd(j.Id)

	return
}

/*
 GETTERS & SETTERS
*/

func (j *Job) HasJobErrored() (errored bool) {
	fmt.Println("Has job", j.GetJobName(), "errored ? ", j.Error != nil)
	if j.Error != nil {
		errored = true
	} else {
		errored = false
	}
	return
}

//create an error well formated
func (j *Job) GetJobError() (errorString string) {
	errorString = j.Error.FmtError()
	return
}

func (j *Job) getJobStringId() (jobId string) {
	jobId = j.Id
	return
}

func (j *Job) jobErrored() (jobError bool, error string) {
	if j.Status == failed {
		jobError = true
		error = j.GetJobError()
	}
	return
}

func (j *Job) GetJobStatus() (jobStatus string) {
	jobStatus = j.Status
	return
}

func (j *Job) getExecutionTime() (executionTime time.Duration) {
	nullTime := time.Time{}
	if j.End == nullTime {
		executionTime = j.Start.Sub(j.End)
	} else {
		executionTime = time.Since(j.Start)
	}
	return
}

func (j *Job) GetJobInfos() (jobjson []byte, err error) {
	jobjson, err = json.Marshal(j)
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}

func (j *Job) GetJobId() (jobId string) {
	jobId = j.Id
	return
}

func (j *Job) GetJobName() (jobName string) {
	jobName = j.Name
	return
}

func (j *Job) setJobToWaiting() {
	j.Status = waiting
}

func (j *Job) setJobToError() {
	j.Status = failed
}

func (j *Job) setJobToSuccess() {
	j.Status = success
}

func (j *Job) setJobToProcessing() {
	j.Status = processing
}
