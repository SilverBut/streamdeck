package main

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type commandValue struct {
	Args        []string
	ShellString string
	UseShell    bool
}

func (c commandValue) IsEmpty() bool {
	if len(c.Args) == 0 && c.ShellString == "" {
		return true
	}
	if c.UseShell {
		return strings.TrimSpace(c.ShellString) == ""
	}
	return len(c.Args) == 0
}

func (c commandValue) Build() (*exec.Cmd, error) {
	if c.UseShell {
		shellCmd := strings.TrimSpace(c.ShellString)
		if shellCmd == "" {
			return nil, errors.New("No command supplied")
		}
		return exec.Command("sh", "-c", shellCmd), nil
	}

	if len(c.Args) == 0 {
		return nil, errors.New("No command supplied")
	}

	return exec.Command(c.Args[0], c.Args[1:]...), nil
}

func (c *commandValue) UnmarshalJSON(data []byte) error {
	var shell string
	if err := json.Unmarshal(data, &shell); err == nil {
		c.Args = nil
		c.ShellString = shell
		c.UseShell = true
		return nil
	}

	var args []string
	if err := json.Unmarshal(data, &args); err == nil {
		c.Args = args
		c.ShellString = ""
		c.UseShell = false
		return nil
	}

	return errors.New("command must be a string or an array of strings")
}

func (c *commandValue) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var shell string
	if err := unmarshal(&shell); err == nil {
		c.Args = nil
		c.ShellString = shell
		c.UseShell = true
		return nil
	}

	var args []string
	if err := unmarshal(&args); err == nil {
		c.Args = args
		c.ShellString = ""
		c.UseShell = false
		return nil
	}

	return errors.New("command must be a string or an array of strings")
}
