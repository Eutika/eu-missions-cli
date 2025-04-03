package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/eutika/eu-missions-cli/internal/services"
)

type ValidateCommand struct {
	remoteService *services.RemoteService
	executor      *CommandExecutor
}

func NewValidateCommand(remoteService *services.RemoteService, executor *CommandExecutor) *cobra.Command {
	vc := &ValidateCommand{
		remoteService: remoteService,
		executor:      executor,
	}

	return &cobra.Command{
		Use:   "validate [id]",
		Short: "Validate a command from remote service",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Fetch command from remote
			commands, err := vc.remoteService.FetchCommands(args[0])
			if err != nil {
				cmd.PrintErrf("Error retrieving command: %v\n", err)
				os.Exit(1)
			}

			// Confirm execution
			if !vc.executor.ConfirmExecution(commands) {
				fmt.Println("Command execution cancelled.")
				return
			}

			// Execute command and capture output
			output, err := vc.executor.ExecuteCommand(commands)
			if err != nil {
				cmd.PrintErrf("Error executing command: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("")
			fmt.Println("💻 Resultado en local:")
			fmt.Println(output)
			fmt.Println("------")

			// Send result back to remote endpoint
			if response, sendErr := vc.remoteService.SendCommandResult("validate", args[0], output); sendErr != nil {
				cmd.PrintErrf("Error sending command result: %v\n", sendErr)
				os.Exit(1)
			} else {
				commands, ok := response["commands"].([]interface{})
				if !ok {
					cmd.PrintErrf("Invalid response format: commands field is not an array\n")
					os.Exit(1)
				}

				fmt.Println("\nValidation Summary:")
				fmt.Println("- Commands:")
				for _, cmdData := range commands {
					if cmdMap, ok := cmdData.(map[string]interface{}); ok {
						fmt.Printf("  - Command: %s, Correct: %v\n", cmdMap["command"], cmdMap["isCorrect"])
					}
				}
				fmt.Printf("- Overall Validation: %v\n", response["isValid"])
				fmt.Printf("- Percentage Correct: %.0f%%\n", response["percentageCorrect"])
				fmt.Printf("- Required Correct Percentage: %.0f%%\n\n", response["requiredCorrectPercentage"])
			}
		},
	}
}
