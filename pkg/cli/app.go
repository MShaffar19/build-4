package cli

import (
	"context"

	"github.com/pkg/errors"

	"github.com/solo-io/build/pkg/envutils"

	"github.com/solo-io/build/pkg/ingest"

	"github.com/solo-io/build/pkg/version"

	"github.com/solo-io/go-utils/clicore"

	"github.com/solo-io/go-utils/contextutils"

	v1 "github.com/solo-io/build/pkg/api/v1"
	"github.com/spf13/cobra"
)

type Options struct {
	Internal Internal
	Input    Input
	BuildRun v1.BuildRun
}

type Internal struct {
	ctx context.Context
}
type Input struct {
	Debug bool
}

var AppConfig = clicore.CommandConfig{
	Command:             App,
	Version:             version.Version,
	FileLogPathElements: FileLogPathElements,
	OutputModeEnvVar:    OutputModeEnvVar,
	RootErrorMessage:    RootErrorMessage,
	LoggingContext:      []interface{}{"version", version.Version},
}

func Run() {
	AppConfig.Run()
}

func App(ctx context.Context, version string) *cobra.Command {
	o := &Options{
		Internal: Internal{ctx: ctx},
	}
	app := &cobra.Command{
		Use:     "build",
		Short:   "CLI for solo.io's build tool",
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			buildRun, err := ingest.InitializeBuildRun()
			if err != nil {
				return err
			}
			o.BuildRun = buildRun

			if o.Input.Debug {
				config := o.BuildRun.Config.BuildEnvVars
				contextutils.CliLogInfow(o.Internal.ctx, "logging build env vars to debug file",
					"build_id", config.BuildId,
					"tag_version", config.TaggedVersion)
			}
			return nil
		},
	}

	app.AddCommand(
		o.reportComputedValues(),
		o.validateOperatingParameters(),
	)
	app.PersistentFlags().BoolVar(&o.Input.Debug, "debug", false, "enable verbose debug output")
	return app
}

//------------------------------------------------------------------------------
// parse-env
//------------------------------------------------------------------------------

func (o *Options) reportComputedValues() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse-env",
		Short: "read environment variables and return corresponding build values",
	}
	cmd.AddCommand(
		o.reportRelease(),
		o.reportImageTag(),
		o.reportContainerPrefix(),
		o.reportVersion())
	return cmd
}

func (o *Options) reportRelease() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "reports if a build is a release build",
		RunE: func(cmd *cobra.Command, args []string) error {
			cbv := o.BuildRun.Config.ComputedBuildVars
			contextutils.CliLogInfow(o.Internal.ctx, envutils.StringForBoolToEnv(cbv.Release), "config", cbv)
			return nil
		},
	}
	return cmd
}

func (o *Options) reportImageTag() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image-tag",
		Short: "reports the image tag to use for this build",
		RunE: func(cmd *cobra.Command, args []string) error {
			cbv := o.BuildRun.Config.ComputedBuildVars
			contextutils.CliLogInfow(o.Internal.ctx, cbv.ImageTag, "config", cbv)
			return nil
		},
	}
	return cmd
}

func (o *Options) reportContainerPrefix() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "container-prefix",
		Short: "reports the container repo and org spec (ex: gcr.io/solo-projects/)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cbv := o.BuildRun.Config.ComputedBuildVars
			contextutils.CliLogInfow(o.Internal.ctx, cbv.ContainerPrefix, "config", cbv)
			return nil
		},
	}
	return cmd
}

func (o *Options) reportVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "reports the version of the source being build (tag during release, build id during test)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cbv := o.BuildRun.Config.ComputedBuildVars
			contextutils.CliLogInfow(o.Internal.ctx, cbv.Version, "config", cbv)
			return nil
		},
	}
	return cmd
}

//------------------------------------------------------------------------------
// validate-operating-parameters
//------------------------------------------------------------------------------

func (o *Options) validateOperatingParameters() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-operating-parameters",
		Short: "for use by scripts for closed-loop communication: exits gracefully if provided arguments match computed values, exits with error otherwise",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := ingest.ValidateOperatingParameters(args, o.BuildRun.Config.ComputedBuildVars)
			if err != nil {
				return errors.Wrapf(err, "did not receive the expected computed variables")
			}
			contextutils.CliLogInfow(o.Internal.ctx, "build parameters are valid", "computed_build_vars", o.BuildRun.Config.ComputedBuildVars)
			return nil
		},
	}
	return cmd
}
