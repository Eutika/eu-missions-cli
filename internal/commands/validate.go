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

func (vc *ValidateCommand) handleCommandResponse(cmd *cobra.Command, response map[string]interface{}) {
	commands, ok := response["commands"].([]interface{})
	if !ok {
		cmd.PrintErrf("❌ Formato de respuesta inválido: el campo 'commands' no es un array\n")
		os.Exit(1)
	}

	fmt.Println("\n📊 Detalle de comandos:")
	fmt.Println("─────────────────────")
	for _, cmdData := range commands {
		if cmdMap, ok := cmdData.(map[string]interface{}); ok {
			isCorrect := cmdMap["isCorrect"].(bool)
			statusIcon := "✅"
			if !isCorrect {
				statusIcon = "❌"
			}
			fmt.Printf("  %s  %s\n", statusIcon, cmdMap["command"])
		}
	}

	fmt.Println("\n🏁 Resultado final:")
	fmt.Println("────────────────")

	isValid := response["isValid"].(bool)
	percentageCorrect := response["percentageCorrect"].(float64)
	requiredPercentage := response["requiredCorrectPercentage"].(float64)

	resultIcon := "🎉"
	resultText := "VALIDACIÓN SUPERADA"
	if !isValid {
		resultIcon = "❌"
		resultText = "VALIDACIÓN NO SUPERADA"
	}

	fmt.Printf("  %s %s\n", resultIcon, resultText)
	fmt.Printf("  ➡️ Porcentaje de acierto: %.0f%% (requerido: %.0f%%)\n\n",
		percentageCorrect, requiredPercentage)
}

func NewValidateCommand(remoteService *services.RemoteService, executor *CommandExecutor) *cobra.Command {
	vc := &ValidateCommand{
		remoteService: remoteService,
		executor:      executor,
	}

	return &cobra.Command{
		Use:   "validate [id]",
		Short: "Valida una ordre des d'un servei remot",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Fetch command from remote
			commands, err := vc.remoteService.FetchCommands(args[0])
			if err != nil {
				cmd.PrintErrf("❌ Error en recuperar l'ordre: %v\n", err)
				os.Exit(1)
			}

			if len(commands) == 0 {
				cmd.PrintErrf("❌ No se ha encontrado el comando con id: %s\n", args[0])
				os.Exit(1)
			}

			// Confirm execution
			if !vc.executor.ConfirmExecution(commands) {
				fmt.Println("⚠️ Execució de l'ordre cancel·lada.")
				return
			}

			// Execute command and capture output
			output, err := vc.executor.ExecuteCommand(commands)
			if err != nil {
				cmd.PrintErrf("❌ Error executing command: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("\n📋 RESULTATS DE LA VALIDACIÓ")
			fmt.Println("══════════════════════════════")

			fmt.Println("\n💻 Resultat d'execució local:")
			fmt.Println("─────────────────────────────")
			fmt.Println(output)

			// Send result back to remote endpoint
			response, sendErr := vc.remoteService.SendCommandResult("validate", args[0], output)
			if sendErr != nil {
				cmd.PrintErrf("❌ Error enviando resultado del comando: %v\n", sendErr)
				os.Exit(1)
			}

			// Handle the command response
			vc.handleCommandResponse(cmd, response)
		},
	}
}
