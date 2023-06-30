package rcsm

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// SessionExists is used to check if a session exists
func SessionExists(serverName string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", getSessionName(serverName))

	// If the output error is nil it means the session exists
	return cmd.Run() == nil
}

// WaitForSessionState is used to wait for a wanted session state, such as wait for the server to shutdown
func WaitForSessionState(serverName string, wantedState bool, timeout time.Duration) error {
	counter := 0.0

	for counter < timeout.Seconds() {
		currentState := SessionExists(serverName)

		if currentState == wantedState {
			return nil
		}

		sleepTime := 100 * time.Millisecond
		counter += sleepTime.Seconds()
		time.Sleep(sleepTime)
	}

	return fmt.Errorf("Timeout of %f seconds exceeded for %s", timeout.Seconds(), serverName)
}

// SessionCreate is used to create a tmux session and start a command
func SessionCreate(serverName string, fullPath string, startCommand string) (string, error) {
	sessionName := getSessionName(serverName)
	attachCommand := getAttachCommand(serverName)

	if SessionExists(serverName) {
		return "", fmt.Errorf("Already started, run \"%s\" to see the console", attachCommand)
	}

	javaCommand := strings.Split(startCommand, " ")

	tmuxParams := append([]string{"new", "-d", "-s", sessionName}, javaCommand...)

	cmd := exec.Command("tmux", tmuxParams...)
	cmd.Dir = fullPath
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	if WaitForSessionState(serverName, true, 15*time.Second) != nil {
		return "", fmt.Errorf("Server crashed on start, check server logs")
	}

	return attachCommand, nil
}

// SessionRunCommand is used to run a command on a session
func SessionRunCommand(serverName string, command string) error {
	sessionName := getSessionName(serverName)

	if !SessionExists(serverName) {
		return fmt.Errorf("Server is not running, cannot run \"%s\"", command)
	}

	cmd := exec.Command("tmux", "send-keys", "-t", sessionName, "-l", command)

	err := cmd.Run()

	if err != nil {
		return err
	}

	cmd = exec.Command("tmux", "send-keys", "-t", sessionName, "Enter")

	return cmd.Run()
}

// SessionTerminate is used to terminate a session with an optional instant kill
func SessionTerminate(serverName string, stopCommand string, instantKill bool) error {
	sessionName := getSessionName(serverName)

	if instantKill {
		return killRawSession(sessionName)
	}

	err := SessionRunCommand(serverName, stopCommand)
	if err != nil {
		return err
	}

	timeoutSeconds := AutoRestartCrashTimeoutSec
	timeoutDuration, err := time.ParseDuration(fmt.Sprintf("%ds", timeoutSeconds))
	if err != nil {
		return err
	}

	err = WaitForSessionState(serverName, false, timeoutDuration)
	if err != nil {
		TriggerLogEvent("warn", serverName, fmt.Sprintf("Timeout shutdown of %d seconds reached, killing the server", timeoutSeconds))
		return killRawSession(sessionName)
	}

	return nil
}

func killRawSession(sessionName string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", sessionName)
	TriggerLogEvent("warn", sessionName, "Sending kill")

	return cmd.Run()
}

func getSessionName(serverName string) string {
	return MinecraftTmuxSessionPrefix + serverName
}

func getAttachCommand(serverName string) string {
	return fmt.Sprintf("tmux a -t %s", getSessionName(serverName))
}
