package utils

import "sync"

type Task func() error
type Tasks []Task
type Errors []error

func RunParallelTasks(tasks Tasks, parallel bool) (errors Errors) {
	if !parallel {
		for _, task := range tasks {
			if error := task(); error != nil {
				errors = append(errors, error)

				return
			}
		}

		return
	}

	waitGroup := sync.WaitGroup{}

	waitGroup.Add(len(tasks))

	errorChannel := make(chan error, 1)
	finishedChannel := make(chan bool, 1)

	// Schedule tasks to be executed
	for _, task := range tasks {
		go func(_task Task) {
			if error := _task(); error != nil {
				errorChannel <- error
			}

			waitGroup.Done()
		}(task)
	}

	// Wait for all tasks to be done and send notification
	go func() {
		waitGroup.Wait()

		close(finishedChannel)
	}()

	done := false

	// Collect errors and wait for all tasks to be done
	for !done {
		select {
		case <-finishedChannel:
			done = true
		case error := <-errorChannel:
			errors = append(errors, error)
		}
	}

	return
}
