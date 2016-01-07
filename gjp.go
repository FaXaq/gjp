// Copyright 2016 Marin Procureur. All rights reserved.
// Use of gjp source code is governed by a MIT license
// license that can be found in the LICENSE file.

/*
Package gjp stands for Go JobPool, and is willing to be a simple jobpool manager. It maintains a number of queues determined at the init. No priority whatsoever, just every queues are processing one job at a time.
*/

package gjp

import (
	"container/list"
	"fmt"
	"time"
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
		name          string        //Public property retrievable
		element       *list.Element //list element necessary to remove from jobList
		status        int           //Status of the current job
		executionTime time.Duration
	}
)

/*
  SHUTDOWN INSTRUCTIONS
*/
const abord string = "abort"     //stop immediatly cuting jobs immediatly
const classic string = "classic" //stop by non authorizing new jobs in the pool and shutting down after
const long string = "long"       //stop when there is no more job to process

/*
   JOB STATUS
*/
const failed int = 0
const success int = 1
const waiting int = 2

/*
   INTERFACES
*/

type JobRunner interface {
	ExecuteJob() bool
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
	job = &Job{
		JobRunner: jobRunner,
		name:      jobName,
		status:    waiting,
	}
	//Add new job to the queue
	jp.queue[poolNumber].jobsRemaining += 1
	job.element = jp.queue[poolNumber].jobs.PushBack(job)

	return
}

//Loop for processing jobs
//Started at the creation of the pool
func (jp *JobPool) ProcessJobs() {
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
}

//Stop the jobPool and release all memory allowed
//NYI
func (jp *JobPool) StopProcessingJobs() {
	jp.working = false

}

//Dequeue job from priority page
func (jp *JobPool) DequeueJob(p int, e *list.Element) {
	jp.queue[p].dequeueJob(e)

}

/*
   PRIVATE FUNCTIONS
*/

//Remove job from currentQueue
func (jq *JobQueue) dequeueJob(e *list.Element) {
	jq.jobs.Remove(e)
}

//execute current joblist
func (jq *JobQueue) executeJobQueue() {
	//Always take the first job in queue
	for e := jq.jobs.Front(); e != nil; e = jq.jobs.Front() {
		j := e.Value.(*Job)

		//Since job is retrieved remove it from the waiting queue
		jq.dequeueJob(e)

		//start job execution
		go jq.launchJobExecution()

		//put jo in the executionChannel
		jq.executionChannel <- j

		//Retrieve the job report from the reportChannel
		//Waiting until job is finished
		jobReport := <-jq.reportChannel

		//Checking status on report
		switch jobReport.status {
		case failed:
			fmt.Println("Job", jobReport.name, "failed")

			//Push it back at the end of the queue to process it later
			jq.jobs.PushBack(jobReport)
			break

		case success:
			fmt.Println("Job", jobReport.name, "executed in", jobReport.executionTime)

			//Remove one job
			jq.jobsRemaining -= 1
			break
		}

		//Go to the next job
	}

	//unlock queue to allow new jobs to be push to it
	jq.unlockQueue()
}

func (jq *JobQueue) launchJobExecution() {
	j := <-jq.executionChannel
	j.executeJob(time.Now())
	jq.totalExecutionTime += j.executionTime
	j.status = success
	jq.reportChannel <- j
}

func (j *Job) executeJob(start time.Time) bool {
	defer func() {
		j.executionTime = time.Since(start)
	}()
	fmt.Println("Start", j.name, "job")
	return j.ExecuteJob()
}

//lock queue while executing
func (jq *JobQueue) lockQueue() {
	jq.working = true
}

//unlock queue when jobs are done
func (jq *JobQueue) unlockQueue() {
	jq.working = false
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

func (myjob *MyJob) ExecuteJob() bool {
	for i := 0; i < 20; i++ {
		fmt.Println(myjob.name, i)
	}
	return true
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
	time.Sleep(time.Millisecond * 20) //Mandatory otherwise no output will be displayed
}

You should have an output like this :

Start YES job
Start THAT job
YES
THAT
Job THAT executed in 69.631µs
Job YES executed in 69.219µs
Start ROCKS job
ROCKS
Job ROCKS executed in 23.397µs

*/
