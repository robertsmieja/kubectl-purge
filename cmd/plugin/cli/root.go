package cli

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/robertsmieja/kubectl-purge/pkg/logger"
	"github.com/robertsmieja/kubectl-purge/pkg/plugin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
	"strings"
	"sync"
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

			yes := viper.GetBool("yes")

			if !yes {
				prompt := promptui.Prompt{
					Label:     "This is a destructive operation, are you sure",
					IsConfirm: true,
				}

				result, err := prompt.Run()
				if err != nil {
					return errors.Wrap(err, "Prompt failed")
				}

				confirmation := result == "y" || result == "Y"
				if !confirmation {
					return errors.New("No confirmation was given!")
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()

			logCh := make(chan string, 1)
			errorCh := make(chan error, 1)

			// print messages and errors while running
			logWaitGroup := sync.WaitGroup{}
			logWaitGroup.Add(2)
			logMutex := &sync.Mutex{}
			go func() {
				defer logWaitGroup.Done()
				for logStr := range logCh {
					logMutex.Lock()
					log.Info(logStr)
					logMutex.Unlock()
				}
			}()
			go func() {
				defer logWaitGroup.Done()
				for err := range errorCh {
					logMutex.Lock()
					log.Error(err)
					logMutex.Unlock()
				}
			}()

			log.Info("Running")
			if err := plugin.RunPlugin(KubernetesConfigFlags, logCh, errorCh); err != nil {
				return errors.Cause(err)
			}
			logWaitGroup.Wait()
			log.Info("Finished")

			return nil
		},
	}

	cobra.OnInitialize(initConfig)

	cmd.Flags().BoolP("yes", "y", false, "Delete without confirming")

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
