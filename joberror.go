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
)

/*
   TYPES
*/

type (
	//Error handling structure
	JobError struct {
		ErrorString string //error as a string
	}
)

func NewJobError(err error) (jobError *JobError) {
	jobError = &JobError{
		err.Error(),
	}
	return
}


//create an error well formated
func (je *JobError) fmtError() (errorString string) {
	errorString = fmt.Sprintln("error :",
		je.ErrorString)
	return
}
