package main

import (
	"github.com/boichique/movie-reviews/internal/log"

	"github.com/boichique/movie-reviews/scrapper/cmd"
	"github.com/spf13/cobra"
)

type ScrapOptions struct {
	Output string
}

func main() {
	var opts cmd.ScrapOptions
	opts.Output = "./scrapper/output"

	logger, err := log.SetupLogger(true, "debug")
	if err != nil {
		panic(err)
	}

	root := &cobra.Command{
		Use:   "scrapper",
		Short: "Use this tool to scrap movie info",
	}

	root.AddCommand(cmd.NewScrapCmd(logger))
	root.AddCommand(cmd.NewIngestCmd(logger))

	err = root.Execute()
	if err != nil {
		logger.With("err", err).Error("error executing a command")
	}

	err = cmd.RunScrap(&opts, logger)
	if err != nil {
		logger.With("err", err).Error("error executing a command")
	}
}
