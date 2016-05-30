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
	"errors"
	"encoding/json"
)

/*
   TYPES
*/

type (
	// JobQueue is structure to control jobs queues
	JobQueue struct {
		Jobs               *list.List `json:"jobsWaiting"`//list of waiting jobs
		executionChannel   chan *Job  `json:"-"` //Channel to contain current job to execute in queue
		reportChannel      chan *Job  `json:"-"`//Channel taking job back when its execution has finished
		working            bool       `json:"-"`//Indicate whether or not the queue is working
		jobsRemaining      int        `json:"-"`//Remaining jobs in the queue
		Done *list.List `json:"jobsDone"` //jobs done
		totalExecutionTime time.Duration `json:"-"`
	}
)


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
	jq.Jobs.Remove(e)
}

//execute current joblist
func (jq *JobQueue) executeJobQueue() {
	defer catchPanic("executeJobQueue")

	for jq.jobsRemaining > 0 {
		//Always take the first job in queue
		j := jq.Jobs.Front().Value.(*Job)

		//Since job is retrieved remove it from the waiting queue
		jq.dequeueJob(jq.Jobs.Front())

		//start job execution
		go jq.launchJobExecution()

		//put jo in the executionChannel
		jq.executionChannel <- j

		//Retrieve the job report from the reportChannel
		//Waiting until job is finished
		jobReport := <-jq.reportChannel

		//Checking status on report
		switch jobReport.Status {
			//Through an error if failed
		case failed:
			if jobReport.HasJobErrored() {
				fmt.Println(jobReport.GetJobError())
			} else {
				fmt.Println(jobReport.GetJobError())
				fmt.Println(jobReport.GetJobName(),
					"panicked after an execution of",
					jobReport.getExecutionTime())
			}
			break
		case success:
			fmt.Println("Job",
				jobReport.GetJobName(),
				"executed in",
				jobReport.getExecutionTime())
			break
		}
		jq.Done.PushBack(jobReport)
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
	j.executeJob(time.Now())
	//add this time to the queue execution time
	jq.totalExecutionTime += j.getExecutionTime()

	//Send job to the report channel
	jq.reportChannel <- j
	return
}


	/*
  GETTERS & SETTERS
*/

func (jq *JobQueue) GetJobFromJobId(jobId string) (j *Job, err error) {
	if jq.Jobs.Len() == 0 && jq.Done.Len() == 0 {
		err = errors.New("No job in queue")
	}
	for e := jq.Jobs.Front(); e != nil; e = e.Next() {
		job := e.Value.(*Job)
		if strings.Compare(job.getJobStringId(), jobId) == 1 {
			j = job
			return
		}
	}
	//check in the done stack if not present in the waiting one
	for e := jq.Done.Front(); e != nil; e = e.Next() {
		job := e.Value.(*Job)
		if strings.Compare(job.getJobStringId(), jobId) == 1 {
			j = job
			return
		}
	}
	err = errors.New("Job not found")
	return
}

func (jq *JobQueue) GetJobsWaiting() (jobList string) {
	jlArray, err := json.Marshal(jq)
	if err != nil {
		fmt.Println("Error while processing serialization on jobs waiting :",
			err.Error())
	}
	fmt.Println(jlArray)
	jobList = string(jlArray[:])
	return
}
