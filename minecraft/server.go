package minecraft

import (
	"bytes"
	"discordbot/config"
	"fmt"
	"os/exec"

	mcstatus "github.com/mcstatus-io/go-mcstatus"
)

// GetServerStatus obtém o status do servidor Minecraft
func GetServerStatus(cfg *config.Config) (*mcstatus.JavaStatusResponse, error) {
	if cfg.ServerIP == "" {
		return nil, fmt.Errorf("❌ SERVER_IP não definido")
	}

	port := config.ParsePort(cfg.ServerPort, cfg.DefaultPort)
	return mcstatus.GetJavaStatus(cfg.ServerIP, port)
}

// GetOnlinePlayers obtém a lista de jogadores online
func GetOnlinePlayers(cfg *config.Config) (int, []string, error) {
	if cfg.ServerIP == "" {
		return 0, nil, fmt.Errorf("❌ SERVER_IP não definido")
	}

	port := config.ParsePort(cfg.ServerPort, cfg.DefaultPort)
	srv, err := mcstatus.GetJavaStatus(cfg.ServerIP, port)
	if err != nil {
		return 0, nil, err
	}

	var playerNames []string
	for _, player := range srv.Players.List {
		playerNames = append(playerNames, player.NameClean)
	}

	return srv.Players.Online, playerNames, nil
}

// ExecuteServerCommand executa um comando no servidor Minecraft via SSH e screen
func ExecuteServerCommand(command string, cfg *config.Config) error {
	// Construa o comando SSH
	sshCmd := fmt.Sprintf("ssh -p %s %s@%s 'screen -S %s -X stuff \"%s\n\"'",
		cfg.SSHPort, cfg.SSHUser, cfg.SSHHost, cfg.ScreenName, command)

	// Execute o comando
	cmd := exec.Command("bash", "-c", sshCmd)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("falha ao executar comando via SSH: %v - %s", err, stderr.String())
	}

	return nil
}

// ListScreenSessions lista as sessões screen disponíveis no host
func ListScreenSessions(cfg *config.Config) (string, error) {
	// Construa o comando SSH
	sshCmd := fmt.Sprintf("ssh -p %s %s@%s 'screen -ls'",
		cfg.SSHPort, cfg.SSHUser, cfg.SSHHost)

	// Execute o comando
	cmd := exec.Command("bash", "-c", sshCmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("falha ao listar sessões screen via SSH: %v - %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// ExecuteHostCommand executa um comando qualquer no host via SSH
func ExecuteHostCommand(command string, cfg *config.Config) (string, error) {
	// Construa o comando SSH
	sshCmd := fmt.Sprintf("ssh -p %s %s@%s '%s'",
		cfg.SSHPort, cfg.SSHUser, cfg.SSHHost, command)

	// Execute o comando
	cmd := exec.Command("bash", "-c", sshCmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("falha ao executar comando via SSH: %v - %s", err, stderr.String())
	}

	return stdout.String(), nil
}
