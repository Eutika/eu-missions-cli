package commands

import (
	"github.com/spf13/cobra"

	"github.com/eutika/eu-missions-cli/internal/auth"
)

func NewLoginCommand(authService *auth.AuthService) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Autentica la CLI con tu cuenta en Missions",
		Long: "Este comando inicia un proceso de autenticación basado en OAuth2 para conectar tu cuenta de Missions " +
			"con la CLI",
		RunE: func(_ *cobra.Command, _ []string) error {
			return authService.Login()
		},
	}
}
