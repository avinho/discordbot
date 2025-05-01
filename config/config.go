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
}

// Load carrega a configuração das variáveis de ambiente
func Load() (*Config, error) {
	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN não definido")
	}

	config := &Config{
		DiscordToken: discordToken,
		ServerIP:     os.Getenv("SERVER_IP"),
		ServerPort:   os.Getenv("SERVER_PORT"),
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		AdminUserID:  os.Getenv("ADMIN_USER_ID"),
		DefaultPort:  46922,
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
