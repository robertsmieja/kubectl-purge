package cli

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/robertsmieja/kubectl-purge/pkg/logger"
	"github.com/robertsmieja/kubectl-purge/pkg/plugin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
	"strings"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "kubectl-purge",
		Short:         "",
		Long:          `.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := viper.BindPFlags(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to bind flags")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()

			logCh := make(chan string, 1)
			errorCh := make(chan error, 1)

			// print errors while running
			go func() {
				for {
					select {
					case logStr := <-logCh:
						log.Info(logStr)
					case err := <-errorCh:
						log.Error(err)
					}
				}
			}()

			log.Info("Running")
			if err := plugin.RunPlugin(KubernetesConfigFlags, logCh, errorCh); err != nil {
				return errors.Cause(err)
			}
			log.Info("Finished")

			return nil
		},
	}

	cobra.OnInitialize(initConfig)

	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	KubernetesConfigFlags.AddFlags(cmd.Flags())

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	return cmd
}

func InitAndExecute() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.AutomaticEnv()
}
