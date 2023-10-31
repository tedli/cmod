package main

import (
	"errors"
	"flag"
	"github.com/spf13/cobra"
)

func newRootCommand() (cmd *cobra.Command) {
	var (
		inputPath        string
		outputPath       string
		headerFile       string
		boilerplatePath  string
		sourceExtensions []string
	)
	cmd = &cobra.Command{
		Use:   "cmod",
		Short: "Extract used only sources from huge c/c++ project.",
		Long:  "Extract used only sources from huge c/c++ project to embedded into other project as in-tree dependency.",
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			rootLogger := newLogger()
			logger := rootLogger.WithName("cmd")
			defer logger.Info("finished")
			if inputPath == "" || headerFile == "" {
				e := errors.New("no input specified")
				logger.Error(e, "input path and header file should be specified")
				return e
			}
			return extract(rootLogger, inputPath, outputPath, headerFile, boilerplatePath, sourceExtensions)
		},
	}
	fs := flag.NewFlagSet("", flag.PanicOnError)
	pfs := cmd.PersistentFlags()
	initLogging(fs, pfs)
	cmd.Flags().AddGoFlagSet(fs)
	pfs.StringVarP(&inputPath, "input-path", "p", "", "Input project root path.")
	pfs.StringVarP(&outputPath, "output-path", "o", "", "Extracted used only sources output path.")
	pfs.StringVarP(&headerFile, "header-file", "i", "", "The header file to extract, like \"folly/concurrency/ConcurrentHashMap.h\" for #include<folly/concurrency/ConcurrentHashMap.h>")
	pfs.StringVarP(&boilerplatePath, "boilerplate-path", "b", "", "License header file path.")
	pfs.StringSliceVarP(&sourceExtensions, "source-extensions", "s", []string{"cc", "cpp", "S"}, "The source file extensions.")
	return
}
