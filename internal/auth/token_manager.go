// Package auth provides authentication and secure token storage functionality.
package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/zalando/go-keyring"

	"github.com/eutika/eu-missions-cli/internal/config"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// SecureStorage provides a secure way to store credentials with automatic fallback.
type SecureStorage struct {
	service      string
	useKeyring   bool
	fallbackPath string
	mu           sync.RWMutex
}

var (
	storageInstance *SecureStorage
	storageOnce     sync.Once
)

// getStorage returns a singleton instance of SecureStorage.
func getStorage() *SecureStorage {
	storageOnce.Do(func() {
		storageInstance = newSecureStorage(config.NewConfig().GetKeyringService())
	})
	return storageInstance
}

// newSecureStorage creates a new secure storage instance.
func newSecureStorage(service string) *SecureStorage {
	s := &SecureStorage{
		service:    service,
		useKeyring: true,
	}

	// Test if keyring is available
	if !s.isKeyringAvailable() {
		s.useKeyring = false
		s.fallbackPath = s.getFallbackPath()
	}

	return s
}

// printSecurityWarning displays a warning when using fallback storage.
func (s *SecureStorage) printSecurityWarning() {
	fmt.Println()
	fmt.Println("⚠️  Aviso de Seguridad")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("El keyring del sistema no está disponible en este entorno.")
	fmt.Printf("Los tokens se guardarán cifrados en: %s\n", s.fallbackPath)
	fmt.Println()
	fmt.Println("💡 Para mayor seguridad, considera configurar un keyring del sistema:")

	switch runtime.GOOS {
	case "linux":
		fmt.Println("   • Instala gnome-keyring: sudo apt-get install gnome-keyring")
		fmt.Println("   • O usa KWallet si estás en KDE")
		fmt.Println("   • Luego ejecuta: eval $(dbus-launch --sh-syntax)")
	case "darwin":
		fmt.Println("   • Verifica que Keychain esté funcionando correctamente")
		fmt.Println("   • Si estás en SSH, puede que Keychain no esté accesible")
	case "windows":
		fmt.Println("   • Verifica que Credential Manager esté funcionando correctamente")
	}

	fmt.Println()
	fmt.Println("ℹ️  Más información: https://github.com/eutika/eu-missions-cli#security")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// isKeyringAvailable checks if the system keyring is available.
func (s *SecureStorage) isKeyringAvailable() bool {
	// Allow forcing file storage for testing purposes
	if os.Getenv("MISSIONS_CLI_FORCE_FILE_STORAGE") == "true" {
		return false
	}

	// Try to set and delete a test value
	testKey := "__test_availability__"
	testValue := "test"

	err := keyring.Set(s.service, testKey, testValue)
	if err != nil {
		return false
	}

	// Clean up test value
	_ = keyring.Delete(s.service, testKey)
	return true
}

// getFallbackPath returns the path for the fallback storage file.
func (s *SecureStorage) getFallbackPath() string {
	var baseDir string

	if runtime.GOOS == "windows" {
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			baseDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
	} else {
		// Unix-like systems (Linux, macOS)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = os.Getenv("HOME")
		}
		baseDir = filepath.Join(homeDir, ".config")
	}

	configDir := filepath.Join(baseDir, "missions-cli")
	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0700); err != nil {
		// If we can't create the directory, fall back to temp
		return filepath.Join(os.TempDir(), ".missions-cli-tokens")
	}

	return filepath.Join(configDir, ".tokens")
}

// Set stores a key-value pair securely.
func (s *SecureStorage) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.useKeyring {
		return keyring.Set(s.service, key, value)
	}
	return s.setFallback(key, value)
}

// Get retrieves a value by key.
func (s *SecureStorage) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.useKeyring {
		return keyring.Get(s.service, key)
	}
	return s.getFallback(key)
}

// Delete removes a key-value pair.
func (s *SecureStorage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.useKeyring {
		return keyring.Delete(s.service, key)
	}
	return s.deleteFallback(key)
}

// Fallback storage implementation using encrypted file

type tokenStore struct {
	Data map[string]string `json:"data"`
}

// getEncryptionKey derives an encryption key from machine-specific data.
func (s *SecureStorage) getEncryptionKey() []byte {
	// Use hostname and username as seed for encryption key
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}

	// Create a deterministic key based on machine identity
	seed := fmt.Sprintf("%s:%s:%s", s.service, hostname, username)
	hash := sha256.Sum256([]byte(seed))
	return hash[:]
}

// encrypt encrypts data using AES-GCM.
func (s *SecureStorage) encrypt(plaintext []byte) ([]byte, error) {
	key := s.getEncryptionKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM.
func (s *SecureStorage) decrypt(ciphertext []byte) ([]byte, error) {
	key := s.getEncryptionKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// loadStore loads the token store from disk.
func (s *SecureStorage) loadStore() (*tokenStore, error) {
	data, err := os.ReadFile(s.fallbackPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &tokenStore{Data: make(map[string]string)}, nil
		}
		return nil, err
	}

	// Decode base64
	encrypted, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode token file: %w", err)
	}

	// Decrypt
	decrypted, err := s.decrypt(encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token file: %w", err)
	}

	// Parse JSON
	var store tokenStore
	if err := json.Unmarshal(decrypted, &store); err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	if store.Data == nil {
		store.Data = make(map[string]string)
	}

	return &store, nil
}

// saveStore saves the token store to disk.
func (s *SecureStorage) saveStore(store *tokenStore) error {
	// Marshal to JSON
	jsonData, err := json.Marshal(store)
	if err != nil {
		return err
	}

	// Encrypt
	encrypted, err := s.encrypt(jsonData)
	if err != nil {
		return err
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(encrypted)

	// Write to file with restrictive permissions
	return os.WriteFile(s.fallbackPath, []byte(encoded), 0600)
}

// setFallback stores a value in the fallback storage.
func (s *SecureStorage) setFallback(key, value string) error {
	store, err := s.loadStore()
	if err != nil {
		return err
	}

	store.Data[key] = value
	return s.saveStore(store)
}

// getFallback retrieves a value from the fallback storage.
func (s *SecureStorage) getFallback(key string) (string, error) {
	store, err := s.loadStore()
	if err != nil {
		return "", err
	}

	value, exists := store.Data[key]
	if !exists {
		return "", errors.New("key not found")
	}

	return value, nil
}

// deleteFallback removes a value from the fallback storage.
func (s *SecureStorage) deleteFallback(key string) error {
	store, err := s.loadStore()
	if err != nil {
		return err
	}

	delete(store.Data, key)
	return s.saveStore(store)
}

func SaveTokens(token *TokenResponse) error {
	storage := getStorage()

	// Save access token.
	if err := storage.Set("access_token", token.AccessToken); err != nil {
		return NewTokenSavingError(fmt.Errorf("error saving access token: %w", err))
	}

	// Save refresh token.
	if err := storage.Set("refresh_token", token.RefreshToken); err != nil {
		if delErr := storage.Delete("access_token"); delErr != nil {
			return NewTokenSavingError(fmt.Errorf("error deleting access token: %w", delErr))
		}
		return NewTokenSavingError(fmt.Errorf("error saving refresh token: %w", err))
	}

	// Save token expiration.
	expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	if err := storage.Set("token_expires_at", expiresAt.Format(time.RFC3339)); err != nil {
		if delErr := storage.Delete("access_token"); delErr != nil {
			return NewTokenSavingError(fmt.Errorf("error deleting access token: %w", delErr))
		}
		if delErr := storage.Delete("refresh_token"); delErr != nil {
			return NewTokenSavingError(fmt.Errorf("error deleting refresh token: %w", delErr))
		}
		return NewTokenSavingError(fmt.Errorf("error saving token expiration: %w", err))
	}

	// Show security warning if using fallback storage
	if !storage.useKeyring {
		storage.printSecurityWarning()
	}

	return nil
}

const tokenExpiryBufferMinutes = 5

func IsTokenExpired() (bool, error) {
	storage := getStorage()
	expiresAtStr, err := storage.Get("token_expires_at")
	if err != nil {
		return true, err // Si no podemos obtener la fecha, asumimos que expiró
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return true, err // Si no podemos parsear la fecha, asumimos que expiró
	}

	// Considerar el token expirado tokenExpiryBufferMinutes minutos antes de su expiración real
	return time.Now().Add(tokenExpiryBufferMinutes * time.Minute).After(expiresAt), nil
}

func GetCurrentToken() (string, error) {
	expired, err := IsTokenExpired()
	if err != nil {
		return "", err
	}

	if expired {
		if refreshErr := RefreshToken(); refreshErr != nil {
			return "", err
		}
	}

	storage := getStorage()
	return storage.Get("access_token")
}

func RefreshToken() error {
	storage := getStorage()
	_, err := storage.Get("refresh_token")
	if err != nil {
		return NewTokenRefreshError(errors.New("no refresh token found, please login again"))
	}

	// TODO: Implement refresh token login. For now, we force a new login.
	return NewTokenRefreshError(errors.New("session expired, please login again"))
}
