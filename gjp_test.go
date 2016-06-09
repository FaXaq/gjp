package gjp

import (
	"testing"
	"fmt"
)

/*
   Create fake Myjob struct
*/

type (
	MyJob struct {
		id string
	}
)

func (myjob *MyJob) GetProgress(id string) (percent float64, err error){
	fmt.Println("Progress",id)
	return
}

func (myjob *MyJob) NotifyEnd(id string) {
	fmt.Println("End",id)
}

func (myjob *MyJob) NotifyStart(id string) {
	fmt.Println("Start",id)
}

func (myjob *MyJob) ExecuteJob(id string) (err *JobError) {
	fmt.Println("Execute",id)
	return
}

/*
   Ends here
*/


func TestNew(t *testing.T) {
	expectedJobPool := New(4)
	if len(expectedJobPool.queue) != 4 {
                t.Fatalf("Expected %s, got %s", 4, expectedJobPool)
        }
}

func TestGenerateJobUID(t *testing.T) {
	expectedUID := GenerateJobUID()
	expectedUID2 := GenerateJobUID()
	if expectedUID == expectedUID2 {
                t.Fatalf("Expected %s to be an UID", expectedUID)
	}
}

func TestStart(t *testing.T) {
	expectedJobPool := New(4)
	expectedJobPool.Start()
	if expectedJobPool.working != true {
                t.Fatalf("Expected jobpool to be working")
	}
}

func TestQueueJob(t *testing.T) {
	expectedJobPool := New(1) //create new jobPool

	emptyNewMyJob := &MyJob{
		GenerateJobUID(),
	} //create empty new job

	expectedNewJob := expectedJobPool.QueueJob(
		emptyNewMyJob.id,
		"test",
		emptyNewMyJob,
		0,
	)

	if expectedNewJob.Id != emptyNewMyJob.id {
                t.Fatalf("Expected jobpool to be created")
	}
}

func TestGetJobFromJobId(t *testing.T) {
	expectedJobPool := New(1) //create new jobPool

	emptyNewMyJob := &MyJob{
		GenerateJobUID(),
	} //create empty new job

	expectedNewJob := expectedJobPool.QueueJob(
		emptyNewMyJob.id,
		"test",
		emptyNewMyJob,
		0,
	)

	expectedNewJob, err := expectedJobPool.GetJobFromJobId(emptyNewMyJob.id)

	if expectedNewJob.Id != emptyNewMyJob.id ||
		err != nil {
                t.Fatalf("Expected finding job")
	}
}

func TestShutdownWorkPool(t *testing.T) {
	expectedJobPool := New(1)
	expectedJobPool.Start()
	expectedJobPool.ShutdownWorkPool()

	if expectedJobPool.working != false {
                t.Fatalf("Expected finding job")
	}
}
