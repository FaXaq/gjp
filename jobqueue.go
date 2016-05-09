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
)

/*
   TYPES
*/

type (
	// JobQueue is structure to control jobs queues
	JobQueue struct {
		jobs               *list.List //list of waiting jobs
		executionChannel   chan *Job  //Channel to contain current job to execute in queue
		reportChannel      chan *Job  //Channel taking job back when its execution has finished
		working            bool       //Indicate whether or not the queue is working
		jobsRemaining      int        //Remaining jobs in the queue
		totalExecutionTime time.Duration
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
				fmt.Println(jobReport.name,
					"panicked after an execution of",
					jobReport.getExecutionTime())
			}
			break
		case success:
			fmt.Println("Job",
				jobReport.name,
				"executed in",
				jobReport.getExecutionTime())
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
	j.executeJob(time.Now())
	//add this time to the queue execution time
	jq.totalExecutionTime += j.getExecutionTime()

	//Send job to the report channel
	jq.reportChannel <- j
	return
}
