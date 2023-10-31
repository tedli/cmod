package main

import (
	"bufio"
	"github.com/go-logr/logr"
	"io"
	"k8s.io/apimachinery/pkg/util/sets"
	"os"
	"path"
	"regexp"
)

func fileExist(fullPath string) bool {
	info, err := os.Stat(fullPath)
	return err == nil && !info.IsDir()
}

var (
	includePattern = regexp.MustCompile(`#[ \t]*include[ \t]*<[ \t]*(?P<header>\S+)[ \t]*>`)
)

func matchCaptureGroup(pattern *regexp.Regexp, input string) (match bool, group map[string]string) {
	mg := pattern.FindStringSubmatch(input)
	if match = mg != nil; !match {
		return
	}
	group = make(map[string]string)
	for i, name := range pattern.SubexpNames() {
		if i != 0 && name != "" {
			group[name] = mg[i]
		}
	}
	return
}

func extractInternal(logger logr.Logger, inputBasePath, currentFile string,
	sourceExtensions []string, index sets.Set[string]) (err error) {

	fullPath := path.Join(inputBasePath, currentFile)
	if !fileExist(fullPath) || index.Has(currentFile) {
		return
	}
	index.Insert(currentFile)
	logger.V(4).Info("indexing one file", "file", currentFile)
	currentFileExtension := path.Ext(fullPath)
	if len(currentFileExtension) > 0 {
		currentFileExtension = currentFileExtension[1:]
	}
	for _, extension := range sourceExtensions {
		if currentFileExtension == extension {
			continue
		}
		sourceFile := currentFile[:len(currentFile)-len(currentFileExtension)] + extension
		if err = extractInternal(logger, inputBasePath, sourceFile, sourceExtensions, index); err != nil {
			logger.Error(err, "extract failed", "sourceFile", sourceFile)
			return
		}
	}
	var current *os.File
	if current, err = os.Open(fullPath); err != nil {
		logger.Error(err, "open current failed", "currentFile", current)
		return
	}
	defer func() {
		if e := current.Close(); e != nil {
			logger.Error(e, "close current failed", "currentFile", current)
		}
	}()
	scanner := bufio.NewScanner(current)
	for scanner.Scan() {
		if match, group := matchCaptureGroup(includePattern, scanner.Text()); match {
			header := group["header"]
			if err = extractInternal(logger, inputBasePath, header, sourceExtensions, index); err != nil {
				logger.Error(err, "extract failed", "header", header)
				return
			}
		}
	}
	if err = scanner.Err(); err != nil {
		logger.Error(err, "parse file to extract header file failed", "currentFile", currentFile)
	}
	return
}

func doOutput(logger logr.Logger, inputPath, outputPath, boilerplatePath string, index sets.Set[string]) (err error) {
	var boilerplate string
	if boilerplatePath != "" {
		var content []byte
		if content, err = os.ReadFile(boilerplatePath); err != nil {
			logger.Error(err, "read boilerplate failed", "path", boilerplatePath)
			return
		}
		boilerplate = string(content)
	}
	if outputPath == "" {
		outputPath = path.Join(inputPath, "dist")
		logger.Info("output path not specified", "default", outputPath)
	}
	for _, file := range index.UnsortedList() {
		logger.V(4).Info("copying file", "file", file)
		outputFullPath := path.Join(outputPath, file)
		dir := path.Dir(outputFullPath)
		if err = os.MkdirAll(dir, 0644); err != nil {
			logger.Error(err, "create output directory failed", "path", dir)
			return
		}
		var output *os.File
		if output, err = os.OpenFile(outputFullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
			logger.Error(err, "create output file failed", "file", file)
			return
		}
		closeOutput := func() {
			if e := output.Close(); e != nil {
				logger.Error(e, "close output failed failed", "file", file)
			}
		}
		if boilerplate != "" {
			if _, err = output.WriteString(boilerplate); err != nil {
				logger.Error(err, "write boilerplate failed", "file", file)
				closeOutput()
				return
			}
		}
		var input *os.File
		if input, err = os.Open(path.Join(inputPath, file)); err != nil {
			logger.Error(err, "open input file failed", "file", file)
			closeOutput()
			return
		}
		closeInput := func() {
			if e := input.Close(); e != nil {
				logger.Error(e, "close input file failed", "file", file)
			}
		}
		if _, err = io.Copy(output, input); err != nil {
			logger.Error(err, "copy file failed", "file", file)
			closeOutput()
			closeInput()
			return
		}
		closeOutput()
		closeInput()
		logger.V(4).Info("file copied", "file", file)
	}
	return
}

func extract(rootLogger logr.Logger, inputPath, outputPath, headerFile, boilerplatePath string,
	sourceExtensions []string) (err error) {

	logger := rootLogger.WithName("ext")
	index := sets.New[string]()
	if err = extractInternal(logger, inputPath, headerFile, sourceExtensions, index); err != nil {
		return
	}
	if err = doOutput(logger, inputPath, outputPath, boilerplatePath, index); err != nil {
		return
	}

	return
}
