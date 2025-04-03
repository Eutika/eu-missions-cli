package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/eutika/eu-missions-cli/internal/services"
)

type ExecuteCommand struct {
	remoteService *services.RemoteService
	executor      *CommandExecutor
}

func NewExecuteCommand(remoteService *services.RemoteService, executor *CommandExecutor) *cobra.Command {
	ec := &ExecuteCommand{
		remoteService: remoteService,
		executor:      executor,
	}

	return &cobra.Command{
		Use:   "submit [id]",
		Short: "Envía el resultado de la validación de una etapa de una misión",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Fetch command from remote
			command, err := ec.remoteService.FetchCommands(args[0])
			if err != nil {
				cmd.PrintErrf("❌ Error al recuperar el comando: %v\n", err)
				os.Exit(1)
			}

			// Confirm execution
			if !ec.executor.ConfirmExecution(command) {
				fmt.Println("⚠️ Ejecución del comando cancelada.")
				return
			}

			// Execute command and capture output
			output, err := ec.executor.ExecuteCommand(command)
			if err != nil {
				cmd.PrintErrf("❌ Error al ejecutar el comando: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("\n📋 RESULTADOS DE LA MISIÓN")
			fmt.Println("═════════════════════════")

			fmt.Println("\n💻 Resultado de ejecución local:")
			fmt.Println("─────────────────────────────")
			fmt.Println(output)

			// Send result back to remote endpoint
			response, sendErr := ec.remoteService.SendCommandResult("submit", args[0], output)
			if sendErr != nil {
				cmd.PrintErrf("❌ Error al enviar el resultado del comando: %v\n", sendErr)
				os.Exit(1)
			}
			commands, isArray := response["commands"].([]interface{})
			if !isArray {
				cmd.PrintErrf("❌ Formato de respuesta inválido: el campo commands no es un array\n")
				os.Exit(1)
			}

			fmt.Println("\n📊 Detalle de comandos:")
			fmt.Println("─────────────────────")
			for _, cmdData := range commands {
				if cmdMap, isMap := cmdData.(map[string]interface{}); isMap {
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
			resultText := "ETAPA COMPLETADA"
			if !isValid {
				resultIcon = "❌"
				resultText = "ETAPA NO COMPLETADA"
			}

			fmt.Printf("  %s %s\n", resultIcon, resultText)
			fmt.Printf("  ➡️ Porcentaje de acierto: %.0f%% (requerido: %.0f%%)\n\n",
				percentageCorrect, requiredPercentage)
		},
	}
}
