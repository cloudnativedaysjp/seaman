package main

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"

	"github.com/cloudnativedaysjp/seaman/cmd/seaman/config"
	"github.com/cloudnativedaysjp/seaman/internal/githubwh"
	"github.com/cloudnativedaysjp/seaman/internal/slackbot"
	"github.com/cloudnativedaysjp/seaman/pkg/log"
)

func main() {
	conf, err := config.ParseFlag()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	eg, ctx := errgroup.WithContext(context.Background())

	// for Logger
	loggerOpts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key != log.KeyDetail {
				return a
			}
			err, ok := a.Value.Any().(error)
			if !ok || !conf.StackTrace {
				return slog.Attr{}
			}
			return slog.String(log.KeyDetail, fmt.Sprintf("%+v", err))
		},
	}
	if conf.Debug {
		loggerOpts.Level = slog.LevelDebug
	}
	ctx = log.IntoContext(ctx, slog.New(slog.NewJSONHandler(os.Stdout, loggerOpts)))

	// launch
	eg.Go(func() error { return slackbot.Run(ctx, conf) })
	eg.Go(func() error { return githubwh.Run(ctx, conf) })
	if err := eg.Wait(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
