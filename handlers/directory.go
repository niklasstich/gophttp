package handlers

import (
	_ "embed"
	"gophttp/common"
	"gophttp/http"
	"os"
	"slices"
	"strings"
)
import "github.com/cbroglie/mustache"

//go:embed directory.template.mustache
var directoryTemplateString string

var directoryTemplate *mustache.Template

func init() {
	var err error
	directoryTemplate, err = mustache.ParseString(directoryTemplateString)
	if err != nil {
		panic(err)
	}
}

type directoryHandler struct {
	HtmlPage string
}

func (d directoryHandler) HandleRequest(ctx http.Context) error {
	ctx.Response.Body = d.HtmlPage
	ctx.Response.Status = http.StatusOK
	ctx.Response.AddHeader(http.Header{
		Name:  "MIME",
		Value: "text/html",
	})
	return nil
}

func NewDirectoryHandler(dirPath string) (Handler, error) {
	h := directoryHandler{}

	//get all directories first, then append all files to the list
	directories, err := common.DirsInDirectory(dirPath)
	if err != nil {
		return nil, err
	}
	//build response body once, reuse it on requests
	files, err := common.FilesInDirectory(dirPath)
	if err != nil {
		return nil, err
	}

	files = slices.Concat(directories, files)

	//calculate the http path to the file
	var filesWithPaths []struct{ Filename, HttpPath string }
	for _, file := range files {
		fp := strings.Join([]string{dirPath, file}, string(os.PathSeparator))
		httpP := http.GetHttpPathForFilepath(fp)
		filesWithPaths = append(filesWithPaths, struct{ Filename, HttpPath string }{Filename: file, HttpPath: httpP})
	}

	page, err := directoryTemplate.Render(map[string]interface{}{"files": filesWithPaths, "path": http.GetHttpPathForFilepath(dirPath)})
	if err != nil {
		return nil, err
	}

	h.HtmlPage = page

	return h, nil
}
