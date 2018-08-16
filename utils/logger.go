package utils

import (
	"path"

	log "github.com/sirupsen/logrus"
)

var debug bool

func SetDebug(_debug bool) {
	debug = _debug

	// Turn logging debug info on/off
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

type logrusHook struct{}

func (hook logrusHook) Fire(entry *log.Entry) error {
	show := _progressShow

	HideProgress()

	fields := log.Fields{}

	for key, value := range entry.Data {
		// Keys that start with _ denote debug information
		// and are hidden for info messages as long as debugging
		// is turned off. The underscore is removed.
		if len(key) > 0 && key[0] == '_' {
			if entry.Level != log.InfoLevel || debug {
				fields[key[1:]] = value
			}
		} else {
			fields[key] = value
		}
	}

	entry.Data = fields

	if show {
		ShowProgress()
	}

	return nil
}

func (hook logrusHook) Levels() []log.Level {
	return log.AllLevels
}

func SetupLogger() {
	log.AddHook(logrusHook{})
}

func LogFilename(message, filename string) {
	name := path.Base(filename)

	log.WithFields(log.Fields{"name": name, "_filename": filename}).Info(message)
}

func LogURL(message, url string) {
	name := path.Base(url)

	log.WithFields(log.Fields{"name": name, "_url": url}).Info(message)
}
