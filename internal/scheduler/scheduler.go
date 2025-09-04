package scheduler

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Jobs []Job
}

type Job struct {
	Schedule    string
	FullCommand string

	minutes     map[int]bool
	hours       map[int]bool
	daysOfMonth map[int]bool
	months      map[int]bool
	daysOfWeek  map[int]bool
}

const (
	minMinute = 0
	maxMinute = 59
	minHour   = 0
	maxHour   = 23
	minDom    = 1
	maxDom    = 31
	minMonth  = 1
	maxMonth  = 12
	minDow    = 0 // Sunday
	maxDow    = 6 // Saturday
)

func (j *Job) parseSchedule() error {
	fields := strings.Fields(j.Schedule)
	if len(fields) != 5 {
		return fmt.Errorf("invalid schedule format: expected 5 fields, got %d for schedule '%s'", len(fields), j.Schedule)
	}

	var err error
	j.minutes, err = parseField(fields[0], minMinute, maxMinute)
	if err != nil {
		return fmt.Errorf("invalid minute field: %w", err)
	}
	j.hours, err = parseField(fields[1], minHour, maxHour)
	if err != nil {
		return fmt.Errorf("invalid hour field: %w", err)
	}
	j.daysOfMonth, err = parseField(fields[2], minDom, maxDom)
	if err != nil {
		return fmt.Errorf("invalid day-of-month field: %w", err)
	}
	j.months, err = parseField(fields[3], minMonth, maxMonth)
	if err != nil {
		return fmt.Errorf("invalid month field: %w", err)
	}
	j.daysOfWeek, err = parseField(fields[4], minDow, maxDow)
	if err != nil {
		return fmt.Errorf("invalid day-of-week field: %w", err)
	}

	return nil
}

func parseField(field string, min, max int) (map[int]bool, error) {
	parts := make(map[int]bool)
	if field == "*" {
		for i := min; i <= max; i++ {
			parts[i] = true
		}
		return parts, nil
	}

	for _, item := range strings.Split(field, ",") {
		val, err := strconv.Atoi(item)
		if err != nil {
			return nil, fmt.Errorf("invalid value '%s'", item)
		}
		if val < min || val > max {
			return nil, fmt.Errorf("value %d out of range [%d, %d]", val, min, max)
		}
		parts[val] = true
	}
	return parts, nil
}

func (j *Job) ShouldRun(t time.Time) bool {
	// Truncate the time to the minute for accurate comparison
	t = t.Truncate(time.Minute)

	if !j.minutes[t.Minute()] {
		return false
	}
	if !j.hours[t.Hour()] {
		return false
	}
	if !j.daysOfMonth[t.Day()] {
		return false
	}
	if !j.months[int(t.Month())] {
		return false
	}
	if !j.daysOfWeek[int(t.Weekday())] {
		return false
	}

	return true
}

func (j *Job) Run() {
	log.Printf("INFO: Starting job: %s", j.FullCommand)

	cmd := exec.Command("cmd", "/C", j.FullCommand)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("ERROR: Failed to get stdout pipe for job '%s': %v", j.FullCommand, err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("ERROR: Failed to get stderr pipe for job '%s': %v", j.FullCommand, err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("ERROR: Failed to start job '%s': %v", j.FullCommand, err)
		return
	}

	go logOutput(stdout, j.FullCommand, "STDOUT")
	go logOutput(stderr, j.FullCommand, "STDERR")

	if err := cmd.Wait(); err != nil {
		log.Printf("ERROR: Job '%s' finished with error: %v", j.FullCommand, err)
	} else {
		log.Printf("INFO: Job '%s' finished successfully", j.FullCommand)
	}
}

func logOutput(pipe io.Reader, jobName, streamName string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		log.Printf("JOB_OUTPUT [%s] [%s]: %s", jobName, streamName, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Printf("ERROR: Error reading %s for job '%s': %v", streamName, jobName, err)
	}
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open config file: %w", err)
	}
	defer file.Close()

	var jobs []Job
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			log.Printf("WARN: Skipping invalid line %d in %s: expected at least 6 fields", lineNumber, path)
			continue
		}

		schedulePart := strings.Join(fields[0:5], " ")
		commandPart := strings.Join(fields[5:], " ")

		job := Job{
			Schedule:    schedulePart,
			FullCommand: commandPart,
		}

		if err := job.parseSchedule(); err != nil {
			log.Printf("WARN: Skipping line %d due to schedule parse error: %v", lineNumber, err)
			continue
		}

		jobs = append(jobs, job)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	return &Config{Jobs: jobs}, nil
}
