# GJP

Go job pool is an easy to use library willing to help users implement a performant a minimal job pool.

Through its simple JobRunner interface containing 4 methods : 
``` go
type JobRunner interface {
	ExecuteJob(job *Job) (*JobError) //Job error has to be set if the job errored
	NotifyStart(job *Job)
	NotifyEnd(job *Job)
	GetProgress(id string) (percentage float64, err error)
}
```

All you will need is to initiate the JobPool, start and queue jobs. Regarding the priority, it's your implementation of each queues, and each job that will set it up.

## Example

``` go
  //create job pool with three queues
	jobPool := gjp.New(3)
	//start it
	jobPool.Start()
	//----------------------------------------------------v Queue number
	//------------------------------------v Job name      v
	//-----------------------v  Job ID    v               v 
	j := jobPool.QueueJob(newJob.Id, newJob.Name, newJob, 0)  
```

## Job API

``` go
func (j *Job) HasJobErrored() (errored bool) {}

func (j *Job) GetJobError() (errorString string) {}

func (j *Job) GetJobStatus() (jobStatus string) {}

func (j *Job) GetJobInfos() (jobjson []byte, err error) {}

func (j *Job) GetJobId() (jobId string) {}

func (j *Job) GetJobName() (jobName string) {}
```

## Job Error API

``` go
func NewJobError(err error, desc string) (jobError *JobError) {}

func (je *JobError) FmtError() (errorString string) {}
```

## Job pool API

``` go
func New(poolRange int) (jp *JobPool) {}

func GenerateJobUID() (id string) {}

func (jp *JobPool) Start() {}

func (jp *JobPool) QueueJob(id string, jobName string, jobRunner JobRunner, poolNumber int) (job *Job) {}

func (jp *JobPool) ProcessJobs() {}

func (jp *JobPool) ListenForShutdown() {}

func (jp *JobPool) ShutdownWorkPool() {}

func (jp *JobPool) GetJobFromJobId(jobId string) (j *Job, err error) {}
```
