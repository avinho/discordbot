package minecraft

import (
	"discordbot/config"
	"fmt"
	"os/exec"

	mcstatus "github.com/mcstatus-io/go-mcstatus"
)

// GetServerStatus obtém o status do servidor Minecraft
func GetServerStatus(cfg *config.Config) (*mcstatus.JavaStatusResponse, error) {
	if cfg.ServerIP == "" {
		return nil, fmt.Errorf("SERVER_IP não definido")
	}

	port := config.ParsePort(cfg.ServerPort, cfg.DefaultPort)
	return mcstatus.GetJavaStatus(cfg.ServerIP, port)
}

// GetOnlinePlayers obtém a lista de jogadores online
func GetOnlinePlayers(cfg *config.Config) (int, []string, error) {
	if cfg.ServerIP == "" {
		return 0, nil, fmt.Errorf("SERVER_IP não definido")
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

// ExecuteServerCommand executa um comando no servidor Minecraft via screen
func ExecuteServerCommand(command string) error {
	// Primeiro, verifique se a sessão screen existe
	checkCmd := exec.Command("bash", "-c", "screen -ls | grep minecraft")
	checkOutput, err := checkCmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("sessão screen 'minecraft' não encontrada: %v - %s", err, string(checkOutput))
	}

	// Execute o comando na sessão screen
	screenCmd := fmt.Sprintf("screen -S minecraft -X stuff '%s\n'", command)
	cmd := exec.Command("bash", "-c", screenCmd)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("falha ao executar comando: %v - %s", err, string(output))
	}

	return nil
}
