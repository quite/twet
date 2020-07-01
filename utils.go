package main

import (
	"io"
	"os/exec"
	"syscall"
)

type execResult struct {
	io.ReadCloser
	Status    int
	Output    []byte
	readIndex int64
}

func (res *execResult) Close() error {
	return nil
}

func (res *execResult) Read(p []byte) (n int, err error) {
	if res.readIndex >= int64(len(res.Output)) {
		err = io.EOF
		return
	}

	n = copy(p, res.Output[res.readIndex:])
	res.readIndex += int64(n)
	return
}

func execShell(dir, cmd string) (res *execResult, err error) {
	res = &execResult{}

	sh := exec.Command("/bin/sh", "-c", cmd)
	if dir != "" {
		sh.Dir = dir
	}

	res.Output, err = sh.CombinedOutput()
	if err != nil {

		// Shamelessly borrowed from https://github.com/prologic/je/blob/master/job.go#L247
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				res.Status = status.ExitStatus()
			}
		}
	}

	return
}
