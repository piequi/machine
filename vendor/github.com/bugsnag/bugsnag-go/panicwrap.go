// +build !appengine

package bugsnag

import (
	"github.com/bugsnag/bugsnag-go/errors"
	"github.com/bugsnag/panicwrap"
)

// NOTE: this function does not return when you call it, instead it
// re-exec()s the current process with panic monitoring.
func defaultPanicHandler() {
	defer defaultNotifier.dontPanic()

	err := panicwrap.BasicMonitor(func(output string) {
		toNotify, err := errors.ParsePanic(output)

		if err != nil {
			defaultNotifier.Config.logf("bugsnag.handleUncaughtPanic: %v", err)
		}
		Notify(toNotify, SeverityError, Configuration{Synchronous: true})
	})

	if err != nil {
		defaultNotifier.Config.logf("bugsnag.handleUncaughtPanic: %v", err)
	}
}
