package clicommand

import (
	"github.com/buildkite/agent/agent"
	"github.com/buildkite/agent/cliconfig"
	"github.com/buildkite/agent/logger"
	"github.com/urfave/cli"
)

var DownloadHelpDescription = `Usage:

   buildkite-agent artifact download [arguments...]

Description:

   Downloads artifacts from Buildkite to the local machine.

   Note: You need to ensure that your search query is surrounded by quotes if
   using a wild card as the built-in shell path globbing will provide files,
   which will break the download.

Example:

   $ buildkite-agent artifact download "pkg/*.tar.gz" . --build xxx

   This will search across all the artifacts for the build with files that match that part.
   The first argument is the search query, and the second argument is the download destination.

   If you're trying to download a specific file, and there are multiple artifacts from different
   jobs, you can target the particular job you want to download the artifact from:

   $ buildkite-agent artifact download "pkg/*.tar.gz" . --step "tests" --build xxx

   You can also use the step's jobs id (provided by the environment variable $BUILDKITE_JOB_ID)`

type ArtifactDownloadConfig struct {
	Query       string `cli:"arg:0" label:"artifact search query" validate:"required"`
	Destination string `cli:"arg:1" label:"artifact download path" validate:"required"`
	Step        string `cli:"step"`
	Build       string `cli:"build" validate:"required"`
	AgentSocket string `cli:"agent-socket" validate:"required"`
	NoColor     bool   `cli:"no-color"`
	Debug       bool   `cli:"debug"`
	DebugHTTP   bool   `cli:"debug-http"`
}

var ArtifactDownloadCommand = cli.Command{
	Name:        "download",
	Usage:       "Downloads artifacts from Buildkite to the local machine",
	Description: DownloadHelpDescription,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "step",
			Value: "",
			Usage: "Scope the search to a paticular step by using either it's name or job ID",
		},
		cli.StringFlag{
			Name:   "build",
			Value:  "",
			EnvVar: "BUILDKITE_BUILD_ID",
			Usage:  "The build that the artifacts were uploaded to",
		},
		AgentSocketFlag,
		NoColorFlag,
		DebugFlag,
		DebugHTTPFlag,
	},
	Action: func(c *cli.Context) {
		// The configuration will be loaded into this struct
		cfg := ArtifactDownloadConfig{}

		// Load the configuration
		if err := cliconfig.Load(c, &cfg); err != nil {
			logger.Fatal("%s", err)
		}

		// Setup the any global configuration options
		HandleGlobalFlags(cfg)

		// Create the API client
		client := agent.APIClient{}.CreateFromSocket(cfg.AgentSocket)

		// Setup the downloader
		downloader := agent.ArtifactDownloader{
			APIClient:   client,
			Query:       cfg.Query,
			Destination: cfg.Destination,
			BuildID:     cfg.Build,
			Step:        cfg.Step,
		}

		// Download the artifacts
		if err := downloader.Download(); err != nil {
			logger.Fatal("Failed to download artifacts: %s", err)
		}
	},
}
