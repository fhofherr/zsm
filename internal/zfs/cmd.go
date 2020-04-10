package zfs

import (
	"fmt"
	"os/exec"
)

// CmdFunc is a function that creates an *exec.Cmd.
//
// It allows to abstract from an installed program for testing purposes.
//
// When the Run method of the returned Cmd is called the program is executed
// with all args appended to its call.
//
// args must not contain the program to execute.
type CmdFunc func(args ...string) *exec.Cmd

// NewCmdFunc creates a new CmdFunc for the named program.
//
// The returned CmdFunc takes the args passed to NewCmdFunc as well as the args
// passed to the CmdFunc itself into account.
//
// Internally the returned CmdFunc uses exec.Command. Everything about
// exec.Command applies here as well.
func NewCmdFunc(name string, args ...string) CmdFunc {
	return func(moreArgs ...string) *exec.Cmd {
		allArgs := make([]string, 0, len(args)+len(moreArgs))
		allArgs = append(allArgs, args...)
		allArgs = append(allArgs, moreArgs...)
		return exec.Command(name, allArgs...)
	}
}

// WithEnv modifies the command returned by the CmdFunc f to execute with the
// passed environment.
//
// env is a slice of strings of the form KEY=VALUE.
func WithEnv(f CmdFunc, env map[string]string) CmdFunc {
	cmdEnv := make([]string, 0, len(env))
	for k, v := range env {
		cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", k, v))
	}
	return func(args ...string) *exec.Cmd {
		cmd := f(args...)
		cmd.Env = cmdEnv
		return cmd
	}
}

// SwallowFurtherArgs returns a CmdFunc that does not pass any args it
// it received on to the program, but instead writes them to swallowed.
func SwallowFurtherArgs(f CmdFunc, swallowed *[]string) CmdFunc {
	return func(args ...string) *exec.Cmd {
		if *swallowed == nil {
			*swallowed = make([]string, 0, len(args))
		}
		*swallowed = append(*swallowed, args...)
		return f()
	}
}
