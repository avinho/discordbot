package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/generative-ai-go/genai"
	mcstatus "github.com/mcstatus-io/go-mcstatus"
	"google.golang.org/api/option"
)

var (
	token     = os.Getenv("DISCORD_TOKEN")
	server    = os.Getenv("SERVER_IP")
	port      = os.Getenv("SERVER_PORT")
	geminiKey = os.Getenv("GEMINI_API_KEY")
	commands  = []*discordgo.ApplicationCommand{
		{
			Name:        "online",
			Description: "Mostra quantos jogadores estão online no servidor",
		},
		{
			Name:        "status",
			Description: "Mostra o status geral do servidor",
		},
		{
			Name:        "perguntar",
			Description: "Faz uma pergunta para o bot sobre Minecraft",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "pergunta",
					Description: "Sua pergunta sobre Minecraft",
					Required:    true,
				},
			},
		},
		{
			Name:        "command",
			Description: "Executa um comando no servidor Minecraft",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "comando",
					Description: "Comando a ser executado no servidor Minecraft",
					Required:    true,
				},
			},
		},
	}
)

func main() {
	if token == "" {
		log.Fatal("DISCORD_TOKEN não definido")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}

	dg.AddHandler(onReady)
	dg.AddHandler(onInteraction)
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds

	if err := dg.Open(); err != nil {
		panic(err)
	}

	defer dg.Close()

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, cmd := range commands {
		registeredCmd, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
		if err != nil {
			log.Fatalf("Erro ao registrar comando %s: %v", cmd.Name, err)
		}
		registeredCommands[i] = registeredCmd
		fmt.Println("Comando registrado:", registeredCmd.Name)
	}

	fmt.Println("Bot está rodando. Pressione CTRL-C para sair.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	s.UpdateGameStatus(0, "/online para ver jogadores")
	fmt.Println("Bot pronto como", s.State.User.String())

	stats, err := mcstatus.GetJavaStatus(server, portStringToInt())
	count := 0
	max := 0
	if err != nil {
		log.Printf("Erro ao obter status do servidor: %v", err)
	}

	if stats.Online {
		count = stats.Players.Online
		max = stats.Players.Max
	}

	// goroutine para atualizar o status a cada 2 minutos
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			s.UpdateGameStatus(0, fmt.Sprintf("/online para ver jogadores. %d/%d", count, max))
		}
	}()
}

func onInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "online":
		handleOnlineCommand(s, i)
	case "status":
		handleStatusCommand(s, i)
	case "perguntar":
		handlePerguntarCommand(s, i)
	case "command":
		handleServerCommand(s, i)
	}
}

func executeServerCommand(command string) error {
	screenCmd := fmt.Sprintf("screen -S minecraft -X stuff '%s\n'", command)

	cmd := exec.Command("bash", "-c", screenCmd)
	return cmd.Run()
}

func initGemini() (*genai.Client, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(geminiKey))
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar Gemini: %v", err)
	}
	return client, nil
}

func handlePerguntarCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	pergunta := options[0].StringValue()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	ctx := context.Background()
	client, err := initGemini()
	if err != nil {
		respondWithError(s, i, "Erro ao inicializar IA")
		return
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")
	prompt := fmt.Sprintf("Responda como um especialista em Minecraft. Sempre responda em português. Ao responder, não mencione que você é um modelo de IA. Adicione Emoji ao final da resposta de acordo com o contexto da resposta. Pergunta:\n\n%s\n\n", pergunta)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		respondWithError(s, i, "Erro ao gerar resposta")
		return
	}

	response := resp.Candidates[0].Content.Parts[0].(genai.Text)
	responseStr := string(response)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &responseStr,
	})
}

func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &message,
	})
}

func handleStatusCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	srv, err := mcstatus.GetJavaStatus(server, portStringToInt())
	msg := ""
	if err != nil {
		msg = "❌ Erro ao obter status do servidor"
	} else {
		msg = fmt.Sprintf("**Status do Servidor** 🖥️\n"+
			"Status: %s\n"+
			"Versão: %s\n"+
			"Jogadores: %d/%d\n"+
			"MOTD: %s\n",
			func() string {
				if srv.Online {
					return "✅ Online"
				}
				return "❌ Offline"
			}(),
			srv.Version.NameClean,
			srv.Players.Online, srv.Players.Max,
			srv.MOTD.Clean)
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &msg,
	})
	if err != nil {
		log.Printf("Erro ao editar resposta: %v", err)
	}
}

func handleOnlineCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	if err != nil {
		log.Printf("Erro ao enviar resposta inicial: %v", err)
		return
	}

	count, players, err := getOnlinePlayers()
	msg := ""
	if err != nil {
		msg = "❌ Erro ao obter contagem de jogadores online. Reclama com o Avinho"
		log.Printf("Erro ao obter players: %v", err)
	} else if count == 0 {
		msg = "Nenhum descupado online. 😃"
	} else {
		msg = fmt.Sprintf("**Descupados online: %d** 🎮\n", count)
		if len(players) > 0 {
			msg += "```\n"
			for _, player := range players {
				msg += fmt.Sprintf("👤 %s\n", player)
			}
			msg += "```"
		}
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &msg,
	})
	if err != nil {
		log.Printf("Erro ao editar resposta: %v", err)
	}
}

func getOnlinePlayers() (int, []string, error) {
	if server == "" {
		return 0, nil, fmt.Errorf("SERVER_IP não definido")
	}

	srv, err := mcstatus.GetJavaStatus(server, portStringToInt())
	if err != nil {
		return 0, nil, err
	}

	var playerNames []string
	for _, player := range srv.Players.List {
		playerNames = append(playerNames, player.NameClean)
	}

	return srv.Players.Online, playerNames, nil
}

func portStringToInt() uint16 {
	if port == "" {
		return 46922
	}
	var p uint16
	fmt.Sscanf(port, "%d", &p)
	if p == 0 {
		p = 46922
	}
	return p
}

func handleServerCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasAdminPermission(i) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Você não tem permissão para executar comandos no servidor",
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// Pega o comando dos argumentos
	options := i.ApplicationCommandData().Options
	command := options[0].StringValue()

	err := executeServerCommand(command)

	var msg string
	if err != nil {
		msg = fmt.Sprintf("❌ Erro ao executar comando: %v", err)
	} else {
		msg = fmt.Sprintf("✅ Comando executado: %s", command)
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &msg,
	})
}

func hasAdminPermission(i *discordgo.InteractionCreate) bool {
	return i.User.ID == "320369059621502997"
}
