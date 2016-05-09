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
		jobsRemaining int //total of jobs remaining
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
	ExecuteJob() (*JobError)
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
			jobs:             list.New(),
			executionChannel: make(chan *Job, 2),
			reportChannel:    make(chan *Job, 2),
			working:          false,
		}
	}
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
			fmt.Println("in place", i, "there is", jp.queue[i].jobs.Len(), "job waiting")
			jobList += fmt.Sprintf("in place %d there is %d job waiting", i, jp.queue[i].jobs.Len())
		}
	}
	return
}

//Queue new job to currentJobPool taking on
func (jp *JobPool) QueueJob(jobName string, jobRunner JobRunner, poolNumber int) (job *Job, jobId string) {
	defer catchPanic(jobName, "QueueJob")

	job, jobId = newJob(jobRunner, jobName)
	//Add new job to the queue
	jp.queue[poolNumber].jobsRemaining += 1
	jp.queue[poolNumber].jobs.PushBack(job)

	fmt.Println("Adding",jobName,"to Queue", poolNumber,"with id", jobId)

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
	}
	return
}

//Stop the jobPool and release all memory allowed
//NYI
func (jp *JobPool) ListenForShutdown() {
	fmt.Println("Waiting for shutdown")
	<- jp.shutdownChannel
	fmt.Println("Shutting Down")
	jp.working = false
	fmt.Println("Shutdown")
}

func (jp *JobPool) ShutdownWorkPool() {
	jp.shutdownChannel <- abort
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
