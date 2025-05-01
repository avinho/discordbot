package config

import (
	"fmt"
	"os"
)

// Config armazena a configuração do bot
type Config struct {
	DiscordToken string
	ServerIP     string
	ServerPort   string
	GeminiAPIKey string
	AdminUserID  string
	DefaultPort  uint16

	// Configurações SSH
	SSHUser    string
	SSHHost    string
	SSHPort    string
	SSHKeyPath string
	ScreenName string
}

// Load carrega a configuração das variáveis de ambiente
func Load() (*Config, error) {
	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		return nil, fmt.Errorf("❌ DISCORD_TOKEN não definido")
	}

	config := &Config{
		DiscordToken: discordToken,
		ServerIP:     os.Getenv("SERVER_IP"),
		ServerPort:   os.Getenv("SERVER_PORT"),
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		AdminUserID:  os.Getenv("ADMIN_USER_ID"),
		DefaultPort:  46922,

		// Configurações SSH
		SSHUser:    os.Getenv("SSH_USER"),
		SSHHost:    os.Getenv("SSH_HOST"),
		SSHPort:    os.Getenv("SSH_PORT"),
		SSHKeyPath: os.Getenv("SSH_KEY_PATH"),
		ScreenName: os.Getenv("SCREEN_NAME"),
	}

	// Valores padrão
	if config.SSHUser == "" {
		config.SSHUser = "ubuntu" // Usuário padrão
	}
	if config.SSHHost == "" {
		config.SSHHost = "localhost" // Host padrão
	}
	if config.SSHPort == "" {
		config.SSHPort = "22" // Porta SSH padrão
	}
	if config.SSHKeyPath == "" {
		config.SSHKeyPath = "/home/ubuntu/.ssh/id_rsa_mine" // Caminho padrão da chave
	}
	if config.ScreenName == "" {
		config.ScreenName = "minecraft" // Nome padrão da sessão screen
	}

	return config, nil
}

func ParsePort(portStr string, defaultPort uint16) uint16 {
	if portStr == "" {
		return defaultPort
	}

	var port uint16
	fmt.Sscanf(portStr, "%d", &port)
	if port == 0 {
		return defaultPort
	}
	return port
}
