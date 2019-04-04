package at

import (
	"sync"
	"sync/atomic"
	"time"
)

type Action func()

type Job struct {
	c      chan struct{}
	worked int32
	timer  *time.Timer
	time   time.Time
	action Action
}

var (
	reset         chan struct{}
	tickerWorking int32
	lock          sync.Mutex
	jobCur        *Job
	jobs          []*Job
)

func init() {
	reset = make(chan struct{})
}

// Execute action after a given time.Duration
func After(d time.Duration, action Action) (job *Job) {
	return addJob(d, time.Now().Add(d), action)
}

// Execute action after a given time.Time
func At(t time.Time, action Action) (job *Job) {
	return addJob(t.Sub(time.Now()), t, action)
}

func addJob(d time.Duration, t time.Time, action Action) (job *Job) {
	job = &Job{
		c:      make(chan struct{}, 1),
		time:   t,
		action: action,
	}

	if d < 0 {
		job.c <- struct{}{}
		job.worked = 1
		go action()
	} else {
		lock.Lock()
		if atomic.LoadInt32(&tickerWorking) == 1 {
			reset <- struct{}{}
		}

		i := 0
		k := len(jobs)
		for i < k {
			m := int(uint(i+k) >> 1)
			if jobs[m].time.Equal(job.time) {
				i = m
				break
			} else if jobs[m].time.Before(job.time) {
				i = m + 1
			} else {
				k = m
			}
		}

		jobs = append(jobs, nil)
		if i < len(jobs)-1 {
			copy(jobs[i+1:], jobs[i:])
		}
		jobs[i] = job

		job.timer = time.NewTimer(d)

		tickerReset()

		lock.Unlock()
	}

	return
}

// Cancel job.
func (job *Job) Cancel() {
	if atomic.LoadInt32(&tickerWorking) == 0 {
		return
	}

	lock.Lock()
	if atomic.LoadInt32(&tickerWorking) == 1 {
		reset <- struct{}{}
	}

	remove(job)

	job.c <- struct{}{}

	job.timer.Stop()

	tickerReset()

	lock.Unlock()
}

// Cancel all jobs
func CancelAllJobs() {
	if atomic.LoadInt32(&tickerWorking) == 0 {
		return
	}

	lock.Lock()
	reset <- struct{}{}

	for i, job := range jobs {
		job.c <- struct{}{}
		job.timer.Stop()
		jobs[i] = nil
	}

	jobs = jobs[:0]

	tickerReset()

	lock.Unlock()
}

func remove(job *Job) {
	i := 0
	k := len(jobs)
	for i < k {
		m := int(uint(i+k) >> 1)
		if jobs[m].time.Equal(job.time) {
			if jobs[m] == job {
				jobs[m] = nil
				copy(jobs[m:], jobs[m+1:])
				jobs = jobs[:len(jobs)-1]
				break
			}
		} else if jobs[m].time.Before(job.time) {
			i = m + 1
		} else {
			k = m
		}
	}
}

func tickerReset() {
	// no lock
	if len(jobs) == 0 {
		jobCur = nil
	} else {
		jobCur = jobs[len(jobs)-1]
	}

	if atomic.CompareAndSwapInt32(&tickerWorking, 0, 1) {
		go ticker()
	}
}

func ticker() {
	var job *Job
	for {
		select {
		case <-reset:
		default:
			lock.Lock()
			job = jobCur
			lock.Unlock()
		}

		if job == nil {
			atomic.StoreInt32(&tickerWorking, 0)
			break
		}

		select {
		case <-reset:
		case <-job.timer.C:
			lock.Lock()

			// pop
			remove(job)

			atomic.StoreInt32(&job.worked, 1)
			job.c <- struct{}{}
			go job.action()

			tickerReset()

			lock.Unlock()
		}
	}
}

// Wait job for a given time.Duration.
// If job is executed within time, returned true.
// If d is 0, waited for the job to execute.
func (job *Job) Wait(d time.Duration) bool {
	if d == 0 {
		<-job.c
		return true
	} else {
		select {
		case <-job.c:
			return true
		case <-time.After(d):
			return false
		}
	}
}

// If job is executed and not canceled, returned true.
func (job *Job) IsWorked() bool {
	return atomic.LoadInt32(&job.worked) == 1
}
