// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build darwin

package launchd

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/keybase/client/go/libkb"
)

// Service defines a service
type Service struct {
	label string
	log   Log
}

// NewService constructs a launchd service.
func NewService(label string) Service {
	return Service{
		label: label,
		log:   emptyLog{},
	}
}

// SetLogger sets the logger
func (s *Service) SetLogger(log Log) {
	if log != nil {
		s.log = log
	} else {
		s.log = emptyLog{}
	}
}

// Label for service
func (s Service) Label() string { return s.label }

// EnvVar defines and environment variable for the Plist
type EnvVar struct {
	key   string
	value string
}

// NewEnvVar creates a new environment variable
func NewEnvVar(key string, value string) EnvVar {
	return EnvVar{key, value}
}

// Plist defines a launchd plist
type Plist struct {
	label       string
	binPath     string
	args        []string
	envVars     []EnvVar
	keepAlive   bool
	logFileName string
	comment     string
}

// NewPlist constructs a launchd service plist
func NewPlist(label string, binPath string, args []string, envVars []EnvVar, logFileName string, comment string) Plist {
	return Plist{
		label:       label,
		binPath:     binPath,
		args:        args,
		envVars:     envVars,
		keepAlive:   true,
		logFileName: logFileName,
		comment:     comment,
	}
}

// Start will start the service.
func (s Service) Start(wait time.Duration) error {
	if !s.HasPlist() {
		return fmt.Errorf("No service (plist) installed with label: %s", s.label)
	}

	plistDest := s.plistDestination()
	s.log.Info("Starting %s", s.label)
	// We start using load -w on plist file
	output, err := exec.Command("/bin/launchctl", "load", "-w", plistDest).Output()
	s.log.Debug("Output (launchctl): %s", string(output))

	if wait > 0 {
		status, waitErr := s.WaitForStatus(wait, 500*time.Millisecond)
		if waitErr != nil {
			return waitErr
		}
		if status != nil {
			s.log.Debug("Service status: %#v", status)
		}
	}
	return err
}

// HasPlist returns true if service has plist installed
func (s Service) HasPlist() bool {
	plistDest := s.plistDestination()
	if _, err := os.Stat(plistDest); err == nil {
		return true
	}
	return false
}

// Stop a service.
func (s Service) Stop(wait time.Duration) error {
	s.log.Info("Stopping %s", s.label)
	// We stop by removing the job. This works for non-demand and demand jobs.
	output, err := exec.Command("/bin/launchctl", "remove", s.label).Output()
	s.log.Debug("Output (launchctl): %s", string(output))
	if wait > 0 {
		// The docs say launchd ExitTimeOut defaults to 20 seconds, but in practice
		// it seems more like 5 seconds before it resorts to a SIGKILL.
		// Because of the SIGKILL fallback we can use a large timeout here of 25
		// seconds, which we'll likely never reach unless the process is zombied.
		err = s.WaitForExit(wait)
		if err != nil {
			return err
		}
	}
	return err
}

// Restart a service.
func (s Service) Restart(wait time.Duration) error {
	return Restart(s.Label(), wait, s.log)
}

// WaitForStatus waits for service status to be available
func (s Service) WaitForStatus(wait time.Duration, delay time.Duration) (*ServiceStatus, error) {
	t := time.Now()
	i := 1
	for time.Now().Sub(t) < wait {
		status, err := s.LoadStatus()
		if err != nil {
			return nil, err
		}
		if status != nil {
			return status, nil
		}
		// Tell user we're waiting for status after 4 seconds, every 4 seconds
		if i%4 == 0 {
			s.log.Info("Waiting for %s to be loaded...", s.label)
		}
		time.Sleep(delay)
		i++
	}
	return nil, nil
}

// WaitForExit waits for service to exit
func (s Service) WaitForExit(wait time.Duration) error {
	running := true
	t := time.Now()
	i := 1
	for time.Now().Sub(t) < wait {
		status, err := s.LoadStatus()
		if err != nil {
			return err
		}
		if status == nil || !status.IsRunning() {
			running = false
			break
		}
		// Tell user we're waiting for exit after 4 seconds, every 4 seconds
		if i%4 == 0 {
			s.log.Info("Waiting for %s to exit...", s.label)
		}
		time.Sleep(time.Second)
		i++
	}
	if running {
		return fmt.Errorf("Waiting for service exit timed out")
	}
	return nil
}

// Install will install the launchd service
func (s Service) Install(p Plist, wait time.Duration) error {
	plistDest := s.plistDestination()
	return s.install(p, plistDest, wait)
}

func (s Service) install(p Plist, plistDest string, wait time.Duration) error {
	if _, ferr := os.Stat(p.binPath); os.IsNotExist(ferr) {
		return fmt.Errorf("%s doesn't exist", p.binPath)
	}
	plist := p.plistXML()

	// See GH issue: https://github.com/keybase/client/pull/1399#issuecomment-164810645
	if err := libkb.MakeParentDirs(plistDest); err != nil {
		return err
	}

	s.log.Info("Saving %s", plistDest)
	file := libkb.NewFile(plistDest, []byte(plist), 0644)
	if err := file.Save(); err != nil {
		return err
	}

	return s.Start(wait)
}

// Uninstall will uninstall the launchd service
func (s Service) Uninstall(wait time.Duration) error {
	if err := s.Stop(wait); err != nil {
		return err
	}

	plistDest := s.plistDestination()
	if _, err := os.Stat(plistDest); err == nil {
		s.log.Info("Removing %s", plistDest)
		return os.Remove(plistDest)
	}
	return nil
}

// ListServices will return service with label that starts with a filter string.
func ListServices(filters []string) (services []Service, err error) {
	launchAgentDir := launchAgentDir()
	if _, derr := os.Stat(launchAgentDir); os.IsNotExist(derr) {
		return
	}
	files, err := ioutil.ReadDir(launchAgentDir)
	if err != nil {
		return
	}
	for _, f := range files {
		fileName := f.Name()
		suffix := ".plist"
		// We care about services that contain the filter word and end in .plist
		for _, filter := range filters {
			if strings.HasPrefix(fileName, filter) && strings.HasSuffix(fileName, suffix) {
				label := fileName[0 : len(fileName)-len(suffix)]
				service := NewService(label)
				services = append(services, service)
			}
		}
	}
	return
}

// ServiceStatus defines status for a service
type ServiceStatus struct {
	label          string
	pid            string // May be blank if not set, or a number "123"
	lastExitStatus string // Will be blank if pid > 0, or a number "123"
}

// Label for status
func (s ServiceStatus) Label() string { return s.label }

// Pid for status (empty string if not running)
func (s ServiceStatus) Pid() string { return s.pid }

// LastExitStatus will be blank if pid > 0, or a number "123"
func (s ServiceStatus) LastExitStatus() string { return s.lastExitStatus }

// Description returns service status info
func (s ServiceStatus) Description() string {
	var status string
	infos := []string{}
	if s.IsRunning() {
		status = "Running"
		infos = append(infos, fmt.Sprintf("(pid=%s)", s.pid))
	} else {
		status = "Not Running"
	}
	if s.lastExitStatus != "" {
		infos = append(infos, fmt.Sprintf("exit=%s", s.lastExitStatus))
	}
	return status + " " + strings.Join(infos, ", ")
}

// IsRunning is true if the service is running (with a pid)
func (s ServiceStatus) IsRunning() bool {
	return s.pid != ""
}

// StatusDescription returns the service status description
func (s Service) StatusDescription() string {
	status, err := s.LoadStatus()
	if status == nil {
		return fmt.Sprintf("%s: Not Running", s.label)
	}
	if err != nil {
		return fmt.Sprintf("%s: %v", s.label, err)
	}
	return fmt.Sprintf("%s: %s", s.label, status.Description())
}

// LoadStatus returns service status
func (s Service) LoadStatus() (*ServiceStatus, error) {
	out, err := exec.Command("/bin/launchctl", "list").Output()
	if err != nil {
		return nil, err
	}

	var pid, lastExitStatus string
	var found bool
	scanner := bufio.NewScanner(bytes.NewBuffer(out))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[2] == s.label {
			found = true
			if fields[0] != "-" {
				pid = fields[0]
			}
			if fields[1] != "-" {
				lastExitStatus = fields[1]
			}
		}
	}

	if found {
		// If pid is set and > 0, then clear lastExitStatus which is the
		// exit status of the previous run and doesn't mean anything for
		// the current state. Clearing it to avoid confusion.
		pidInt, _ := strconv.ParseInt(pid, 0, 64)
		if pid != "" && pidInt > 0 {
			lastExitStatus = ""
		}
		return &ServiceStatus{label: s.label, pid: pid, lastExitStatus: lastExitStatus}, nil
	}

	return nil, nil
}

// CheckPlist returns false, if the plist destination doesn't match what we
// would install. This means the plist is old and we need to update it.
func (s Service) CheckPlist(plist Plist) (bool, error) {
	plistDest := s.plistDestination()
	return plist.Check(plistDest)
}

// Install will install a service
func Install(plist Plist, wait time.Duration, log Log) error {
	service := NewService(plist.label)
	service.SetLogger(log)
	return service.Install(plist, wait)
}

// Uninstall will uninstall a service
func Uninstall(label string, wait time.Duration, log Log) error {
	service := NewService(label)
	service.SetLogger(log)
	return service.Uninstall(wait)
}

// Start will start a service
func Start(label string, wait time.Duration, log Log) error {
	service := NewService(label)
	service.SetLogger(log)
	return service.Start(wait)
}

// Stop will stop a service
func Stop(label string, wait time.Duration, log Log) error {
	service := NewService(label)
	service.SetLogger(log)
	return service.Stop(wait)
}

// ShowStatus shows status info for a service
func ShowStatus(label string, log Log) error {
	service := NewService(label)
	service.SetLogger(log)
	status, err := service.LoadStatus()
	if err != nil {
		return err
	}
	if status != nil {
		log.Info("%s", status.Description())
	} else {
		log.Info("No service found with label: %s", label)
	}
	return nil
}

// Restart restarts a service
func Restart(label string, wait time.Duration, log Log) error {
	service := NewService(label)
	service.SetLogger(log)
	err := service.Stop(wait)
	if err != nil {
		return err
	}
	return service.Start(wait)
}

func launchAgentDir() string {
	return filepath.Join(launchdHomeDir(), "Library", "LaunchAgents")
}

// PlistDestination is the plist path for a label
func PlistDestination(label string) string {
	return filepath.Join(launchAgentDir(), label+".plist")
}

// PlistDestination is the service plist path
func (s Service) PlistDestination() string {
	return s.plistDestination()
}

func (s Service) plistDestination() string {
	return PlistDestination(s.label)
}

func launchdHomeDir() string {
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	return currentUser.HomeDir
}

// LogDir is the directory for logs
func LogDir() string {
	return filepath.Join(launchdHomeDir(), "Library", "Logs")
}

func ensureDirectoryExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0700)
		return err
	}
	return nil
}

// Check if plist matches plist at path
func (p Plist) Check(path string) (bool, error) {
	if p.binPath == "" {
		return false, fmt.Errorf("Invalid ProgramArguments")
	}

	// If path doesn't exist, we don't match
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	}

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}

	plistXML := p.plistXML()
	if string(buf) == plistXML {
		return true, nil
	}

	return false, nil
}

// TODO Use go-plist library
func (p Plist) plistXML() string {
	logFile := filepath.Join(LogDir(), p.logFileName)

	encodeTag := func(name, val string) string {
		return fmt.Sprintf("<%s>%s</%s>", name, val, name)
	}

	encodeBool := func(val bool) string {
		sval := "false"
		if val {
			sval = "true"
		}
		return fmt.Sprintf("<%s/>", sval)
	}

	pargs := []string{}
	// First arg is the executable
	pargs = append(pargs, encodeTag("string", p.binPath))
	for _, arg := range p.args {
		pargs = append(pargs, encodeTag("string", arg))
	}

	envVars := []string{}
	for _, envVar := range p.envVars {
		envVars = append(envVars, encodeTag("key", envVar.key))
		envVars = append(envVars, encodeTag("string", envVar.value))
	}

	options := []string{}
	if p.keepAlive {
		options = append(options, encodeTag("key", "KeepAlive"), encodeBool(true))
		options = append(options, encodeTag("key", "RunAtLoad"), encodeBool(true))
	}

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>` + p.label + `</string>
  <key>EnvironmentVariables</key>
  <dict>` + "\n    " + strings.Join(envVars, "\n    ") + `
  </dict>
  <key>ProgramArguments</key>
  <array>` + "\n    " + strings.Join(pargs, "\n    ") + `
  </array>` +
		"\n  " + strings.Join(options, "\n  ") + `
  <key>StandardErrorPath</key>
  <string>` + logFile + `</string>
  <key>StandardOutPath</key>
  <string>` + logFile + `</string>
  <key>WorkingDirectory</key>
  <string>/tmp</string>
</dict>
</plist>
`

	if p.comment != "" {
		xml = fmt.Sprintf("<!-- %s -->\n%s", p.comment, xml)
	}

	return xml
}

// Log is the logging interface for this package
type Log interface {
	Debug(s string, args ...interface{})
	Info(s string, args ...interface{})
}

type emptyLog struct{}

func (l emptyLog) Debug(s string, args ...interface{}) {}
func (l emptyLog) Info(s string, args ...interface{})  {}
