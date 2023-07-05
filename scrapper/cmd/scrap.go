package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/boichique/movie-reviews/scrapper/collectors"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

type ScrapOptions struct {
	Output string
}

func NewScrapCmd(logger *slog.Logger) *cobra.Command {
	var opts ScrapOptions

	cmd := &cobra.Command{
		Use:   "scrap",
		Short: "Scrap movie info",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunScrap(&opts, logger)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output directory")

	_ = cmd.MarkFlagRequired("output")

	return cmd
}

func RunScrap(opts *ScrapOptions, logger *slog.Logger) error {
	baseCollector := collectors.NewBaseCollector()

	bioCollector := collectors.NewBioCollector(collectors.Derive(baseCollector), logger)
	starCollector := collectors.NewStarCollector(collectors.Derive(baseCollector), bioCollector, logger)
	castCollector := collectors.NewCastCollector(collectors.Derive(baseCollector), starCollector, logger)
	movieCollector := collectors.NewMovieCollector(collectors.Derive(baseCollector), castCollector, logger)
	topMoviesCollector := collectors.NewTopMoviesCollector(collectors.Derive(baseCollector), movieCollector, logger)

	topMoviesCollector.Start()
	topMoviesCollector.Wait()
	movieCollector.Wait()
	castCollector.Wait()
	starCollector.Wait()
	bioCollector.Wait()

	writes := []struct {
		data any
		path string
	}{
		{data: movieCollector.Movies(), path: filepath.Join(opts.Output, "movies.json")},
		{data: movieCollector.Genres(), path: filepath.Join(opts.Output, "genres.json")},
		{data: castCollector.Cast(), path: filepath.Join(opts.Output, "cast.json")},
		{data: starCollector.Stars(), path: filepath.Join(opts.Output, "stars.json")},
		{data: bioCollector.Bios(), path: filepath.Join(opts.Output, "bios.json")},
	}

	if err := os.MkdirAll(opts.Output, os.ModePerm); err != nil {
		logger.With("err", err).Error("error creating output dir")
		return err
	}

	for _, w := range writes {
		content, err := json.Marshal(w.data)
		if err != nil {
			return fmt.Errorf("failed to marshal %s: %w", w.path, err)
		}

		if err = os.WriteFile(w.path, content, os.ModePerm); err != nil {
			return fmt.Errorf("failed to write %s: %w", w.path, err)
		}
	}

	return nil
}
