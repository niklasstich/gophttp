package common

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var mimeTypes = map[string]string{
	"txt":   "text/plain",
	"html":  "text/html",
	"htm":   "text/html",
	"css":   "text/css",
	"js":    "application/javascript",
	"json":  "application/json",
	"xml":   "application/xml",
	"csv":   "text/csv",
	"yaml":  "text/yaml",
	"yml":   "text/yaml",
	"md":    "text/markdown",
	"ini":   "text/plain",
	"log":   "text/plain",
	"sh":    "application/x-sh",
	"py":    "text/x-python",
	"java":  "text/x-java-source",
	"c":     "text/x-c",
	"cpp":   "text/x-c++",
	"h":     "text/x-c",
	"hpp":   "text/x-c++",
	"ts":    "application/typescript",
	"tsx":   "text/tsx",
	"jsx":   "text/jsx",
	"php":   "application/x-httpd-php",
	"rb":    "text/x-ruby",
	"pl":    "text/x-perl",
	"go":    "text/x-go",
	"rs":    "text/x-rustsrc",
	"swift": "text/x-swift",
}

func GetMIMEFromPath(filepath string) (string, error) {
	cmd := exec.Command("file", "-b", "--mime-type", filepath)
	b, err := cmd.Output()
	if err != nil {
		return "", getExitErrorFromCommand(err)
	}
	mime_type := strings.TrimSpace(string(b))
	if mime_type == "text/plain" {
		splits := strings.Split(filepath, ".")
		mime_type = lookupTextFormatFromFileEnding(splits[len(splits)-1])
	}

	cmd = exec.Command("file", "-b", "--mime-encoding", filepath)
	b, err = cmd.Output()
	if err != nil {
		return "", getExitErrorFromCommand(err)
	}
	mime_encoding := strings.TrimSpace(string(b))

	return fmt.Sprintf("%s; charset=%s", mime_type, mime_encoding), nil
}

func lookupTextFormatFromFileEnding(ending string) string {
	if s, ok := mimeTypes[ending]; ok {
		return s
	}
	return "text/plain"
}

func getExitErrorFromCommand(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return errors.New(string(exitErr.Stderr))
	}
	return err
}
