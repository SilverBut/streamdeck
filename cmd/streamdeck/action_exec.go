package main

import (
	"os"

	"github.com/Luzifer/go_helpers/v2/env"
	"github.com/pkg/errors"
)

func init() {
	registerAction("exec", actionExec{})
}

type actionExec struct{}

func (actionExec) Execute(attributes attributeCollection) error {
	if attributes.Command.IsEmpty() {
		return errors.New("No command supplied")
	}

	processEnv := env.ListToMap(os.Environ())

	for k, v := range attributes.Env {
		processEnv[k] = v
	}

	command, err := attributes.Command.Build()
	if err != nil {
		return err
	}

	command.Env = env.MapToList(processEnv)

	if attributes.AttachStdout {
		command.Stdout = os.Stdout
	}

	if attributes.AttachStderr {
		command.Stdout = os.Stderr
	}

	if err := command.Start(); err != nil {
		return errors.Wrap(err, "Unable to start command")
	}

	// If "wait" is set to true start command and wait for execution
	if attributes.Wait {
		return errors.Wrap(command.Wait(), "Command was not successful")
	}

	// We don't wait so we release the process and don't care anymore
	return errors.Wrap(command.Process.Release(), "Unable to release process")
}
