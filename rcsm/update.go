package rcsm

import (
	"fmt"
	"syscall"
	"time"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

// StartUpdateChecks starts a task to check for updates
func StartUpdateChecks() {
	ticker := time.NewTicker(time.Minute * time.Duration(AutoUpdateIntervalMinutes))
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				runUpdateChecks()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	runUpdateChecks()
}

func runUpdateChecks() {
	previous := semver.MustParse(Version)
	latest, err := selfupdate.UpdateSelf(previous, AutoUpdateRepo)
	if err != nil {
		TriggerLogEvent("warn", "updater", fmt.Sprintf("Could not check for updates: %s", err))
		return
	}

	if !previous.Equals(latest.Version) {
		if ExitOnAutoUpdate {
			TriggerLogEvent("info", "updater", fmt.Sprintf("Updating rcsm to version %s, stopping...", latest.Version))
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		} else {
			TriggerLogEvent("info", "updater", fmt.Sprintf("Updated rcsm to version %s, please restart to apply changes", latest.Version))
		}
	}
}
