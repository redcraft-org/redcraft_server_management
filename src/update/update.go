package update

import (
	"config"
	"events"
	"fmt"
	"syscall"
	"time"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

// StartUpdateChecks starts a task to check for updates
func StartUpdateChecks() {
	ticker := time.NewTicker(time.Minute * time.Duration(config.AutoUpdateIntervalMinutes))
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
	previous := semver.MustParse(config.Version)
	latest, err := selfupdate.UpdateSelf(previous, config.AutoUpdateRepo)
	if err != nil {
		events.TriggerLogEvent("warn", "updater", fmt.Sprintf("Could not check for updates: %s", err))
		return
	}

	if !previous.Equals(latest.Version) {
		if config.ExitOnAutoUpdate {
			events.TriggerLogEvent("info", "updater", fmt.Sprintf("Updating rcsm to version %s, stopping...", latest.Version))
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		} else {
			events.TriggerLogEvent("info", "updater", fmt.Sprintf("Updated rcsm to version %s, please restart to apply changes", latest.Version))
		}
	}
}
