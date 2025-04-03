package auth

import (
	"errors"
	"fmt"

	"github.com/eutika/eu-missions-cli/internal/config"
)

type AuthService struct {
	config *config.Config
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		config: cfg,
	}
}

// Login handles the device code authentication flow.
func (s *AuthService) Login() error {
	deviceCode, err := RequestDeviceCode()
	if err != nil {
		fmt.Printf("🚫 No ha sido posible solicitar el código de dispositivo a Missions: %v\n", err)
		return errors.New("🚫 No sido posible solicitar el código de dispositivo a Missions")
	}

	fmt.Println("\n🔐 Iniciando el proceso de autenticación con Missions...")
	fmt.Printf("\n   1. Accede con tu navegador a: %s\n", deviceCode.VerificationURI)
	fmt.Printf("   2. Introduce el código: %s\n", deviceCode.UserCode)
	fmt.Println("\n⏳ Esperando autenticación...")

	token, err := PollForToken(deviceCode)
	if err != nil {
		fmt.Printf("🚫 No ha sido posible obtener el token de autenticación de Missions %s", err)
		return errors.New("🚫 No ha sido posible obtener el token de autenticación de Missions")
	}

	if saveErr := SaveTokens(token); saveErr != nil {
		return errors.New("🚫 No ha sido posible guardar el token de autenticación de Missions en tu sistema")
	}

	fmt.Println("\n✅ ¡Enhorabuena, te has autenticado con Missions!")
	return nil
}
