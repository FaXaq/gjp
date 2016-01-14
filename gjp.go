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
	"container/list"
	"fmt"
	"time"
	"strings"
)

/*
   TYPES
*/

type (
	// JobPool principal s²tructure maintaining queues of jobs
	// from a priority range
	JobPool struct {
		poolRange       int
		queue           []*JobQueue //Containing a "poolRange" number of queues
		shutdownChannel chan string //NYI
		working         bool
	}

	// JobQueue is structure to control jobs queues
	JobQueue struct {
		jobs               *list.List //list of waiting jobs
		executionChannel   chan *Job  //Channel to contain current job to execute in queue
		reportChannel      chan *Job  //Channel taking job back when its execution has finished
		working            bool       //Indicate whether or not the queue is working
		jobsRemaining      int        //Remaining jobs in the queue
		totalExecutionTime time.Duration
	}

	// Job
	Job struct {
		JobRunner
		name          string //Public property retrievable
		status        int    //Status of the current job
		error         *JobError
		executionTime time.Duration
	}

	//Error handling structure
	JobError struct {
		JobName     string //job name
		ErrorString string //error as a string
	}
)

/*
  SHUTDOWN INSTRUCTIONS
*/
const (
	abord   string = "abort"   //stop immediatly cuting jobs immediatly
	classic string = "classic" //stop by non authorizing new jobs in the pool and shutting down after
	long    string = "long"    //stop when there is no more job to process
)

/*
   JOB STATUS
*/
const (
	failed  int = 0
	success int = 1
	waiting int = 2
)

/*
   INTERFACES
*/

type JobRunner interface {
	ExecuteJob() (*JobError)
}

/*
   PUBLIC FUNCTIONS
*/

//Initialization function
func New(poolRange int) (jp *JobPool) {
	jp = &JobPool{
		poolRange: poolRange,
		working:   false,
	}
	jp.queue = make([]*JobQueue, jp.poolRange)
	for i := 0; i < len(jp.queue); i++ {
		jp.queue[i] = &JobQueue{
			jobs:             list.New(),
			executionChannel: make(chan *Job, 2),
			reportChannel:    make(chan *Job, 2),
			working:          false,
		}
	}

	//launch the loop
	go jp.ProcessJobs()
	go jp.StopProcessingJobs()

	return
}

//List current waiting jobs in each queues
func (jp *JobPool) ListWaitingJobs() {
	for i, _ := range jp.queue {
		if jp.queue[i] != nil {
			fmt.Println("in place", i, "there is", jp.queue[i].jobs.Len(), "job waiting")
		}
	}
}

//Queue new job to currentJobPool taking on
func (jp *JobPool) QueueJob(jobName string, jobRunner JobRunner, poolNumber int) (job *Job) {
	defer catchPanic(jobName, "QueueJob")

	job = &Job{
		JobRunner: jobRunner,
		name:      jobName,
		status:    waiting,
	}
	//Add new job to the queue
	jp.queue[poolNumber].jobsRemaining += 1
	jp.queue[poolNumber].jobs.PushBack(job)

	return
}

//Loop for processing jobs
//Started at the creation of the pool
func (jp *JobPool) ProcessJobs() {
	defer catchPanic("ProcessJobs")

	jp.working = true

	//While loop
	for jp.working == true {

		//iterate through each queue containing jobs
		for i := 0; i < len(jp.queue); i++ {
			if jp.queue[i].jobsRemaining > 0 &&
				jp.queue[i].working == false {

				//lock queue to avoid double processing
				jp.queue[i].lockQueue()

				//Launch the queue execution in a thread
				go jp.queue[i].executeJobQueue()
			}
		}
	}
	return
}

//Stop the jobPool and release all memory allowed
//NYI
func (jp *JobPool) StopProcessingJobs() {
	jp.working = false
}

/*
   PRIVATE FUNCTIONS
*/

//Handle error
func catchPanic(v ...string) {
	if r := recover(); r != nil {
		fmt.Println("Recover", r)
		fmt.Println("Details :", strings.Join(v, " "))
	}
}

//lock queue while executing
func (jq *JobQueue) lockQueue() {
	jq.working = true
}

//unlock queue when jobs are done
func (jq *JobQueue) unlockQueue() {
	jq.working = false
}

//Remove job from currentQueue
func (jq *JobQueue) dequeueJob(e *list.Element) {
	jq.jobs.Remove(e)
}

//execute current joblist
func (jq *JobQueue) executeJobQueue() {
	defer catchPanic("executeJobQueue")
	for jq.jobsRemaining > 0 {
		//Always take the first job in queue
		j := jq.jobs.Front().Value.(*Job)

		//Since job is retrieved remove it from the waiting queue
		jq.dequeueJob(jq.jobs.Front())

		//start job execution
		go jq.launchJobExecution()

		//put jo in the executionChannel
		jq.executionChannel <- j

		//Retrieve the job report from the reportChannel
		//Waiting until job is finished
		jobReport := <-jq.reportChannel

		//Checking status on report
		switch jobReport.status {
		//Through an error if failed
		case failed:
			if jobReport.error != nil {
				fmt.Println(jobReport.error.fmtError())
			} else {
				fmt.Println(jobReport.name, "panicked after an execution of", jobReport.executionTime)
			}
			break
		case success:
			fmt.Println("Job",
				jobReport.name,
				"executed in",
				jobReport.executionTime)
			break
		}
		jq.jobsRemaining -= 1
		//Go to the next job
	}
	//unlock queue to allow new jobs to be push to it
	jq.unlockQueue()
	return
}

//Launch the JobExecution
func (jq *JobQueue) launchJobExecution() {
	defer catchPanic("launchJobExecution")

	//Retrieve job from execution channel of the queue
	j := <-jq.executionChannel

	//execute the job synchronously with time starter
	j.status = j.executeJob(time.Now())
	//add this time to the queue execution time
	jq.totalExecutionTime += j.executionTime

	//Send job to the report channel
	jq.reportChannel <- j
}

//execute the job safely and set the status back for the reportChannel
func (j *Job) executeJob(start time.Time) (jobStatus int) {
	defer catchPanic("Job", j.name, "failed in executeJob")
	//Set the execution time for this job
	defer func() {
		j.executionTime = time.Since(start)
	}()

	j.error = j.ExecuteJob()
	//Set the status required for the job
	switch j.error {
	case nil:
		jobStatus = success
		break
	default:
		jobStatus = failed
		break
	}
	return
}

//create an error well formated
func (je *JobError) fmtError() (errorString string) {
	errorString = fmt.Sprintln("Job",
		je.JobName,
		"has failed with a error :",
		je.ErrorString)
	return
}

/*
Implementation example :

package main

import (
	"fmt"
	"github.com/FaXaq/gjp"
	"time"
)

type (
	MyJob struct {
		name string
	}
)

func (myjob *MyJob) ExecuteJob() (err *gjp.JobError) {
	if myjob.name == "YES" {
		defer panic("plz send haelp")
		err = &gjp.JobError{
			myjob.name,
			"nooooooooo",
		}
		return
	}
	for i := 0; i < 1; i++ {
		fmt.Println(myjob.name)
	}
	return
}

func main(){
	jobPool := gjp.New(2)
	newJob := &MyJob{
		name: "YES",
	}
	newJob2 := &MyJob{
		name: "THAT",
	}
	newJob3 := &MyJob{
		name: "ROCKS",
	}
	jobPool.QueueJob(newJob.name, newJob, 0)
	jobPool.QueueJob(newJob2.name, newJob2, 1)
	jobPool.QueueJob(newJob3.name, newJob3, 1)
	time.Sleep(time.Millisecond * 30)
}


The execution should retrieve something like :

Recover plz send haelp
Details : Job YES failed in executeJob
YES panicked after an execution of 14.046µs
THAT
Job THAT executed in 24.399µs
ROCKS
Job ROCKS executed in 21.57µs


*/
