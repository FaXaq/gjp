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
	"strings"
	"errors"
	"os/exec"
)

/*
   TYPES
*/

type (
	// JobPool principal s²tructure maintaining queues of jobs
	// from a priority range
	JobPool struct {
		poolRange       int `json:"poolRange"`
		queue           []*JobQueue `json:"jobQueues"` //Containing a "poolRange" number of queues
		shutdownChannel chan string `json:"-"`//NYI
		working         bool `json:"Working"`
		jobsRemaining int `json:"Jobs remaining"` //total of jobs remaining
	}
)

/*
  SHUTDOWN INSTRUCTIONS
*/
const (
	abort   string = "abort"   //stop immediatly cuting jobs immediatly
	classic string = "classic" //stop by non authorizing new jobs in the pool and shutting down after
	long    string = "long"    //stop when there is no more job to process
)

/*
   INTERFACES
*/

type JobRunner interface {
	ExecuteJob(job *Job) (*JobError)
	NotifyStart(job *Job)
	NotifyEnd(job *Job)
	GetProgress(id string) (percentage float64, err error)
}

/*
   PUBLIC FUNCTIONS
*/

//Initialization function
func New(poolRange int) (jp *JobPool) {
	//Create the jobPool
	jp = &JobPool{
		poolRange: poolRange,
		working:   false,
		shutdownChannel:  make(chan string),
	}

	//create the queuer ranges
	jp.queue = make([]*JobQueue, jp.poolRange)

	for i := 0; i < len(jp.queue); i++ {
		jp.queue[i] = &JobQueue{
			Jobs:             list.New(),
			executionChannel: make(chan *Job, 2),
			reportChannel:    make(chan *Job, 2),
			working:          false,
		}
	}
	return
}

//generate uuid
func GenerateJobUID() (id string) {
		//generating uuid for the job
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		panic("Couldn't generate uuid for new job")
		fmt.Println("Couldn't generate uuid for new job")
	}

	id = strings.TrimRight(string(out), "\n")

	return
}

//Start jobPool
func (jp *JobPool) Start() {
	go jp.ListenForShutdown()
	go jp.ProcessJobs()
	jp.working = true
}

//List current waiting jobs in each queues
func (jp *JobPool) ListWaitingJobs() (jobList string){
	for i, _ := range jp.queue {
		if jp.queue[i] != nil {
			jobList += jp.queue[i].GetJobsWaiting()
		}
	}
	return
}

//Queue new job to currentJobPool taking on
func (jp *JobPool) QueueJob(id string, jobName string, jobRunner JobRunner, poolNumber int) (job *Job) {
	defer catchPanic(jobName, "QueueJob")

	job, _ = newJob(id, jobRunner, jobName)
	//Add new job to the queue
	jp.queue[poolNumber].jobsRemaining += 1
	jp.queue[poolNumber].Jobs.PushBack(job)

	jp.working = true

	fmt.Println("Adding",jobName,"to Queue", poolNumber,"with id", job.GetJobId())

	go jp.ProcessJobs()

	return
}

//Loop for processing jobs
//Started at the creation of the pool
func (jp *JobPool) ProcessJobs() {
	defer catchPanic("ProcessJobs")

	//While loop
	for jp.working == true {
		//count jobs remaining
		jp.jobsRemaining = 0;
		//iterate through each queue containing jobs
		for i := 0; i < len(jp.queue); i++ {
			if jp.queue[i].jobsRemaining > 0 {
				jp.jobsRemaining += jp.queue[i].jobsRemaining
				if jp.queue[i].working == false {

					//lock queue to avoid double processing
					jp.queue[i].lockQueue()

					//Launch the queue execution in a thread
					go jp.queue[i].executeJobQueue()
				}
			}
		}
		jp.working = false
	}
	fmt.Println("Shutdown")
	return
}

//Stop the jobPool and release all memory allowed
//NYI
func (jp *JobPool) ListenForShutdown() {
	fmt.Println("Waiting for shutdown")
	<- jp.shutdownChannel
	fmt.Println("Shutting Down, waiting for new jobs")
	jp.working = false
}

func (jp *JobPool) ShutdownWorkPool() {
	jp.shutdownChannel <- abort
}

func (jp *JobPool) GetJobFromJobId(jobId string) (j *Job, err error) {
	for i := 0; i < len(jp.queue); i++ {
		j, err = jp.queue[i].GetJobFromJobId(jobId)
		if err == nil {
			return
		}
	}
	if j == nil && err == nil {
		err = errors.New("Job not found in this job pool")
	}
	return
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
