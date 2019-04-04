package at

import (
	"testing"
	"time"
)

func TestAfter5Seconds(t *testing.T) {
	job := After(2*time.Second, func() {})

	if !job.Wait(3 * time.Second) {
		t.Fail()
	}
}

func TestJobStop(t *testing.T) {
	job := After(2*time.Second, func() { t.Fail() })
	job.Cancel()

	job.Wait(3 * 1000)
	if job.IsWorked() {
		t.Fail()
	}
}

func TestJobStopAll(t *testing.T) {
	jobs := make([]*Job, 20)
	for i := range jobs {
		jobs[i] = After(time.Duration(i+2)*time.Second, func() { t.Fail() })
	}
	CancelAllJobs()

	for _, job := range jobs {
		job.Wait(0)

		if job.IsWorked() {
			t.Fail()
		}
	}
}
