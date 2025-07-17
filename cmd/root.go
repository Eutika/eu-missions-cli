package cmd

import (
	"github.com/spf13/cobra"

	"github.com/eutika/eu-missions-cli/internal/auth"
	"github.com/eutika/eu-missions-cli/internal/commands"
	"github.com/eutika/eu-missions-cli/internal/config"
	"github.com/eutika/eu-missions-cli/internal/services"
)

type CommandDependencies struct {
	Config        *config.Config
	RemoteService *services.RemoteService
	CmdExecutor   *commands.CommandExecutor
	AuthService   *auth.AuthService
}

func NewRootCommand(deps *CommandDependencies) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "missions",
		Short:   "Missions CLI (Command Line Interface)",
		Version: "1.0.8",
	}

	rootCmd.AddCommand(
		commands.NewLoginCommand(deps.AuthService),
		commands.NewExecuteCommand(deps.RemoteService, deps.CmdExecutor),
		commands.NewValidateCommand(deps.RemoteService, deps.CmdExecutor),
	)
	rootCmd.SetVersionTemplate("missions version {{.Version}}\n")

	return rootCmd
}

func Execute() error {
	cfg := config.NewConfig()
	deps := &CommandDependencies{
		Config:        cfg,
		RemoteService: services.NewRemoteService(cfg),
		CmdExecutor:   commands.NewCommandExecutor(cfg),
		AuthService:   auth.NewAuthService(cfg),
	}

	return NewRootCommand(deps).Execute()
}
