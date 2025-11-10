package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/eutika/eu-missions-cli/internal/config"
)

type CommandExecutor struct {
	config *config.Config
}

func NewCommandExecutor(cfg *config.Config) *CommandExecutor {
	return &CommandExecutor{
		config: cfg,
	}
}

func (e *CommandExecutor) ValidateCommand(command string) error {
	for _, pattern := range e.config.GetDangerousPatterns() {
		// Detectar comandos de formato más precisamente
		if pattern == "format " {
			// Evitar falsos positivos con --format
			if strings.Contains(command, "format ") && !strings.Contains(command, "--format") {
				return fmt.Errorf("este comando no está permitido en Missions: detectado patrón '%s' en '%s'", pattern, command)
			}
		} else if strings.Contains(command, pattern) {
			return fmt.Errorf("este comando no está permitido en Missions: detectado patrón '%s' en '%s'", pattern, command)
		}
	}
	return nil
}

func (e *CommandExecutor) ExecuteCommand(commands []string) ([]string, error) {
	results := make([]string, 0, len(commands))

	for _, command := range commands {
		// Validate each command before execution
		if err := e.ValidateCommand(command); err != nil {
			return nil, fmt.Errorf("👮 : '%s': %w", command, err)
		}

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", command)
		} else {
			// Get the current user's shell to ensure environment consistency
			userShell := os.Getenv("SHELL")
			if userShell == "" {
				userShell = "/bin/bash" // Fallback to bash
			}

			// Use login shell to ensure PATH and environment are properly loaded
			cmd = exec.Command(userShell, "-l", "-c", command)
		}

		// Capture both stdout and stderr
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Agrega el error y la salida al array de resultados
			results = append(results, fmt.Sprintf("🔥 Error ejecutando '%s': %v\nSalida: %s", command, err, string(output)))
			break // Si quieres continuar con el resto de comandos, elimina este break
		}
		results = append(results, string(output))
	}

	return results, nil
}

func (e *CommandExecutor) ConfirmExecution(commands []string) bool {
	fmt.Println("\n👀 Se van a ejecutar los siguientes comandos:")
	fmt.Println("─────────────────────────────────────────")
	for _, command := range commands {
		fmt.Printf("  ▶️  %s\n", command)
	}
	fmt.Println()

	for {
		fmt.Print("¿Quieres continuar? (si/no): ")

		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			fmt.Printf("❌ Error al leer la entrada: %v\n", err)
			continue
		}

		// Normalize input
		response = strings.ToLower(strings.TrimSpace(response))

		switch response {
		case "sí", "si", "s", "":
			return true
		case "no", "n":
			return false
		default:
			fmt.Println("⚠️ Por favor, contesta 'sí' o 'no'. También puedes pulsar 'Enter' para confirmar.")
		}
	}
}
