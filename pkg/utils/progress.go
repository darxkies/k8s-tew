package utils

import (
	"fmt"
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

var _spinner *spinner.Spinner
var _progressSteps int
var _progressStep int
var _progressShow bool
var _mutex sync.Mutex
var _supressProgress bool

func init() {
	_mutex = sync.Mutex{}

	_spinner = spinner.New(spinner.CharSets[9], 100*time.Millisecond)
}

func SupressProgress(hide bool) {
	_supressProgress = hide
}

func ShowProgress() {
	if _supressProgress {
		return
	}

	_progressPercentage := 0

	if _progressSteps > 0 {
		_progressPercentage = int(float32(_progressStep+1) * 100.0 / float32(_progressSteps))

		if _progressPercentage > 100 {
			_progressPercentage = 100
		}
	}

	_spinner.Prefix = "["
	_spinner.Suffix = fmt.Sprintf("] Progress: %d%%", _progressPercentage)

	_spinner.Start()

	_progressShow = true
}

func HideProgress() {
	_spinner.Stop()

	_progressShow = false
}

func IncreaseProgressStep() {
	_mutex.Lock()
	defer _mutex.Unlock()

	_progressStep++
}

func SetProgressSteps(steps int) {
	_progressSteps = steps
}
