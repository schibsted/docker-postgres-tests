package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type healthLog struct {
	ExitCode int
	Output   string
}

type health struct {
	Status string
	Log    []healthLog
}

type state struct {
	Health health
}

type inspect struct {
	State state
}

func testCommandOut(t testing.TB, c string, args ...string) string {
	out := &bytes.Buffer{}
	cmd := exec.Command(c, args...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Errorf("failed to run command %s: %v", strings.Join(append([]string{c}, args...), " "), err)
	}
	return out.String()
}

func testCommand(t testing.TB, c string, args ...string) {
	cmd := exec.Command(c, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Errorf("failed to run command %s: %v", strings.Join(append([]string{c}, args...), " "), err)
	}
}

func TestBuildable(t *testing.T) {
	testCommand(t, "docker", "build", ".")
}
func TestHealthcheck(t *testing.T) {
	testCommand(t, "docker", "build", "-t", "test-postgres", ".")
	container := strings.Trim(testCommandOut(t, "docker", "run", "-d", "test-postgres"), " \n")
	defer func() {
		if t.Failed() {
			testCommand(t, "docker", "logs", container)
			testCommand(t, "docker", "inspect", container)
		}
		testCommand(t, "docker", "kill", container)
		testCommand(t, "docker", "rm", "-v", container)
	}()

	inspect := []inspect{}
	for i := 0; i < 10; i++ {
		err := json.Unmarshal([]byte(testCommandOut(t, "docker", "inspect", container)), &inspect)
		if err != nil {
			t.Errorf("failed to decode inspect json: %s", err)
		}
		if len(inspect) > 0 && inspect[0].State.Health.Status == "healthy" {
			for _, log := range inspect[0].State.Health.Log {
				if log.ExitCode == 0 {
					// Check that debug information is logged in the healthcheck logs
					if !strings.Contains(log.Output, "Health query succeed") {
						t.Errorf("with success health query, query should be executed")
					}
					if !strings.Contains(log.Output, "uptime:") {
						t.Errorf("default healthcheck should display uptime from query")
					}
					if t.Failed() {
						fmt.Println("health:", log.Output)
					}
				}
			}
			break
		}
		time.Sleep(1 * time.Second)
	}
	if len(inspect) == 0 || inspect[0].State.Health.Status != "healthy" {
		fmt.Println("decoded inspect: ", inspect)
		t.Errorf("with default command, the container should be healthy")
	}
}
func TestHealthcheckFails(t *testing.T) {
	testCommand(t, "docker", "build", "-t", "test-postgres", ".")
	container := strings.Trim(testCommandOut(
		t, "docker", "run", "-d", "-e", "POSTGRES_HEALTH_QUERY=SELECT 1 from non_existing_table;", "--health-retries", "2", "test-postgres",
	), " \n")
	defer func() {
		if t.Failed() {
			testCommand(t, "docker", "logs", container)
			testCommand(t, "docker", "inspect", container)
		}
		testCommand(t, "docker", "kill", container)
		testCommand(t, "docker", "rm", "-v", container)
	}()

	inspect := []inspect{}
	for i := 0; i < 10; i++ {
		err := json.Unmarshal([]byte(testCommandOut(t, "docker", "inspect", container)), &inspect)
		if err != nil {
			t.Errorf("failed to decode inspect json: %s", err)
		}
		if len(inspect) > 0 && inspect[0].State.Health.Status == "unhealthy" {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if len(inspect) == 0 || inspect[0].State.Health.Status != "unhealthy" {
		fmt.Println("decoded inspect: ", inspect)
		t.Errorf("with an invalid SQL command the container the container must not turn healthy")
	}
}
func TestHealthcheckNoQuery(t *testing.T) {
	testCommand(t, "docker", "build", "-t", "test-postgres", ".")
	container := strings.Trim(testCommandOut(
		t, "docker", "run", "-d", "-e", "POSTGRES_NO_HEALTH_QUERY=true", "--health-retries", "2", "test-postgres",
	), " \n")
	defer func() {
		if t.Failed() {
			testCommand(t, "docker", "logs", container)
			testCommand(t, "docker", "inspect", container)
		}
		testCommand(t, "docker", "kill", container)
		testCommand(t, "docker", "rm", "-v", container)
	}()

	inspect := []inspect{}
	for i := 0; i < 10; i++ {
		err := json.Unmarshal([]byte(testCommandOut(t, "docker", "inspect", container)), &inspect)
		if err != nil {
			t.Errorf("failed to decode inspect json: %s", err)
		}
		if len(inspect) > 0 && inspect[0].State.Health.Status == "healthy" {
			for _, log := range inspect[0].State.Health.Log {
				if log.ExitCode == 0 {
					// Check that debug information is logged in the healthcheck logs
					if strings.Contains(log.Output, "Health query succeed") {
						t.Errorf("without health query, no query should be executed")
					}
					if strings.Contains(log.Output, "uptime:") {
						t.Errorf("without health query, no query should be executed")
					}
				}
			}
			break
		}
		time.Sleep(1 * time.Second)
	}
	if len(inspect) == 0 || inspect[0].State.Health.Status != "healthy" {
		fmt.Println("decoded inspect: ", inspect)
		t.Errorf("with default command, the container should be healthy")
	}
}
