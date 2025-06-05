package common

import (
	"errors"
	"os/exec"
)

func GetMIMEFromPath(filepath string) (string, error) {
	cmd := exec.Command("file", "-b", "-I", filepath)
	b, err := cmd.Output()
	if err != nil {
		// Return the error message from the file command
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", errors.New(string(exitErr.Stderr))
		}
		return "", err
	}
	return string(b), nil
}
