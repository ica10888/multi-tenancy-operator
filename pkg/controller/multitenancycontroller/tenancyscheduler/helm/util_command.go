/*
Copyright The Helm Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/


package helm

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
)

// Commander defines the interface for a Command
//go:generate pegomock generate github.com/jenkins-x/jx/pkg/util Commander -o mocks/commander.go
type Commander interface {
	DidError() bool
	DidFail() bool
	Error() error
	Run() (string, error)
	RunWithoutRetry() (string, error)
	SetName(string)
	CurrentName() string
	SetDir(string)
	CurrentDir() string
	SetArgs([]string)
	CurrentArgs() []string
	SetTimeout(time.Duration)
	SetExponentialBackOff(*backoff.ExponentialBackOff)
	SetEnv(map[string]string)
	CurrentEnv() map[string]string
	SetEnvVariable(string, string)
}


// Command is a struct containing the details of an external command to be executed
type Command struct {
	attempts           int
	Errors             []error
	Dir                string
	Name               string
	Args               []string
	ExponentialBackOff *backoff.ExponentialBackOff
	Timeout            time.Duration
	Out                io.Writer
	Err                io.Writer
	Env                map[string]string
}

// CommandError is the error object encapsulating an error from a Command
type CommandError struct {
	Command Command
	Output  string
	cause   error
}

func (c CommandError) Error() string {
	// sanitise any password arguments before printing the error string. The actual sensitive argument is still present
	// in the Command object
	sanitisedArgs := make([]string, len(c.Command.Args))
	copy(sanitisedArgs, c.Command.Args)
	for i, arg := range sanitisedArgs {
		if strings.Contains(strings.ToLower(arg), "password") && i < len(sanitisedArgs)-1 {
			// sanitise the subsequent argument to any 'password' fields
			sanitisedArgs[i+1] = "*****"
		}
	}

	return fmt.Sprintf("failed to run '%s %s' command in directory '%s', output: '%s'",
		c.Command.Name, strings.Join(sanitisedArgs, " "), c.Command.Dir, c.Output)
}

// SetName Setter method for Name to enable use of interface instead of Command struct
func (c *Command) SetName(name string) {
	c.Name = name
}

// CurrentName returns the current name of the command
func (c *Command) CurrentName() string {
	return c.Name
}

// SetDir Setter method for Dir to enable use of interface instead of Command struct
func (c *Command) SetDir(dir string) {
	c.Dir = dir
}

// CurrentDir returns the current Dir
func (c *Command) CurrentDir() string {
	return c.Dir
}

// SetArgs Setter method for Args to enable use of interface instead of Command struct
func (c *Command) SetArgs(args []string) {
	c.Args = args
}

// CurrentArgs returns the current command arguments
func (c *Command) CurrentArgs() []string {
	return c.Args
}

// SetTimeout Setter method for Timeout to enable use of interface instead of Command struct
func (c *Command) SetTimeout(timeout time.Duration) {
	c.Timeout = timeout
}

// SetExponentialBackOff Setter method for ExponentialBackOff to enable use of interface instead of Command struct
func (c *Command) SetExponentialBackOff(backoff *backoff.ExponentialBackOff) {
	c.ExponentialBackOff = backoff
}

// SetEnv Setter method for Env to enable use of interface instead of Command struct
func (c *Command) SetEnv(env map[string]string) {
	c.Env = env
}

// CurrentEnv returns the current envrionment variables
func (c *Command) CurrentEnv() map[string]string {
	return c.Env
}

// SetEnvVariable sets an environment variable into the environment
func (c *Command) SetEnvVariable(name string, value string) {
	if c.Env == nil {
		c.Env = map[string]string{}
	}
	c.Env[name] = value
}

// Attempts The number of times the command has been executed
func (c *Command) Attempts() int {
	return c.attempts
}

// DidError returns a boolean if any error occurred in any execution of the command
func (c *Command) DidError() bool {
	if len(c.Errors) > 0 {
		return true
	}
	return false
}

// DidFail returns a boolean if the command could not complete (errored on every attempt)
func (c *Command) DidFail() bool {
	if len(c.Errors) == c.attempts {
		return true
	}
	return false
}

// Error returns the last error
func (c *Command) Error() error {
	if len(c.Errors) > 0 {
		return c.Errors[len(c.Errors)-1]
	}
	return nil
}

// Run Execute the command and block waiting for return values
func (c *Command) Run() (string, error) {
	os.Setenv("PATH", PathWithBinary(c.Dir))
	var r string
	var e error

	f := func() error {
		r, e = c.run()
		c.attempts++
		if e != nil {
			c.Errors = append(c.Errors, e)
			return e
		}
		return nil
	}

	c.ExponentialBackOff = backoff.NewExponentialBackOff()
	if c.Timeout == 0 {
		c.Timeout = 3 * time.Minute
	}
	c.ExponentialBackOff.MaxElapsedTime = c.Timeout
	c.ExponentialBackOff.Reset()
	err := backoff.Retry(f, c.ExponentialBackOff)
	if err != nil {
		return "", err
	}
	return r, nil
}

// RunWithoutRetry Execute the command without retrying on failure and block waiting for return values
func (c *Command) RunWithoutRetry() (string, error) {
	os.Setenv("PATH", PathWithBinary(c.Dir))
	var r string
	var e error
	c.Out = os.Stdout
	c.Err = os.Stderr
	r, e = c.run()
	c.attempts++
	if e != nil {
		c.Errors = append(c.Errors, e)
	}
	return r, e
}

func (c *Command) String() string {
	var builder strings.Builder
	builder.WriteString(c.Name)
	for _, arg := range c.Args {
		builder.WriteString(" ")
		builder.WriteString(arg)
	}
	return builder.String()
}

func (c *Command) run() (string, error) {
	e := exec.Command(c.Name, c.Args...)
	if c.Dir != "" {
		e.Dir = c.Dir
	}
	if len(c.Env) > 0 {
		m := map[string]string{}
		environ := os.Environ()
		for _, kv := range environ {
			paths := strings.SplitN(kv, "=", 2)
			if len(paths) == 2 {
				m[paths[0]] = paths[1]
			}
		}
		for k, v := range c.Env {
			m[k] = v
		}
		envVars := []string{}
		for k, v := range m {
			envVars = append(envVars, k+"="+v)
		}
		e.Env = envVars
	}

	if c.Out != nil {
		e.Stdout = c.Out
	}

	if c.Err != nil {
		e.Stderr = c.Err
	}

	var text string
	var err error

	if c.Out != nil {
		err := e.Run()
		if err != nil {
			return text, CommandError{
				Command: *c,
				cause:   err,
			}
		}
	} else {
		data, err := e.CombinedOutput()
		output := string(data)
		text = strings.TrimSpace(output)
		if err != nil {
			return text, CommandError{
				Command: *c,
				Output:  text,
				cause:   err,
			}
		}
	}

	return text, err
}

// PathWithBinary Sets the $PATH variable. Accepts an optional slice of strings containing paths to add to $PATH
func PathWithBinary(paths ...string) string {
	//path := os.Getenv("PATH")
	//binDir, _ := JXBinLocation()
	//answer := path + string(os.PathListSeparator) + binDir
	//mvnBinDir, _ := MavenBinaryLocation()
	//if mvnBinDir != "" {
	//	answer += string(os.PathListSeparator) + mvnBinDir
	//}
	answer := os.Getenv("PATH")
	for _, p := range paths {
		answer += string(os.PathListSeparator) + p
	}
	return answer
}
