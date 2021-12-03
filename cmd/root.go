package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/choffmeister/kustomization-helm/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type rootCmd struct {
	cmd       *cobra.Command
	directory string
}

func newRootCmd(version FullVersion) *rootCmd {
	result := &rootCmd{}
	cmd := &cobra.Command{
		Version:      version.Version,
		Use:          "kustomization-helm",
		Short:        "An converter from helm charts to kustomizations",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, viperInstance, err := initConfig(result)
			if err != nil {
				return fmt.Errorf("unable to initialize: %v", err)
			}
			config, err := internal.LoadConfig(*viperInstance, path.Join(*dir, "kustomization-helm.yaml"))
			if err != nil {
				return fmt.Errorf("unable to load configuration: %v", err)
			}
			err = internal.Run(*dir, *config)
			if err != nil {
				return fmt.Errorf("unable to run: %v", err)
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&result.directory, "dir", ".", "dir")

	result.cmd = cmd
	return result
}

func initConfig(rootCmd *rootCmd) (*string, *viper.Viper, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}
	if !filepath.IsAbs(rootCmd.directory) {
		dir = filepath.Join(dir, rootCmd.directory)
	} else {
		dir = rootCmd.directory
	}

	viperInstance := viper.New()
	viperInstance.AddConfigPath(dir)
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viperInstance.SetEnvPrefix("HELM_KUSTOMIZE")
	viperInstance.AutomaticEnv()

	return &dir, viperInstance, nil
}

func Execute(version FullVersion) error {
	rootCmd := newRootCmd(version)
	initConfig(rootCmd)
	return rootCmd.cmd.Execute()
}

type FullVersion struct {
	Version string
	Commit  string
	Date    string
	BuiltBy string
}

func (v FullVersion) ToString() string {
	result := v.Version
	if v.Commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, v.Commit)
	}
	if v.Date != "" {
		result = fmt.Sprintf("%s\nbuilt at: %s", result, v.Date)
	}
	if v.BuiltBy != "" {
		result = fmt.Sprintf("%s\nbuilt by: %s", result, v.BuiltBy)
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result = fmt.Sprintf("%s\nmodule version: %s, checksum: %s", result, info.Main.Version, info.Main.Sum)
	}
	return result
}