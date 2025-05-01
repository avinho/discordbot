package discord

import (
	"discordbot/ai"
	"discordbot/config"
	"discordbot/minecraft"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/generative-ai-go/genai"
)

type Bot struct {
	Session    *discordgo.Session
	Commands   []*discordgo.ApplicationCommand
	AiClient   *genai.Client
	Config     *config.Config
	Registered []*discordgo.ApplicationCommand
}

func NewBot(cfg *config.Config) (*Bot, error) {
	dg, err := discordgo.New("Bot " + cfg.DiscordToken)

	if err != nil {
		return nil, fmt.Errorf("❌ Erro ao criar sessão do Discord: %v", err)
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds

	var aiClient *genai.Client
	if cfg.GeminiAPIKey != "" {
		aiClient, err = ai.InitGemini(cfg.GeminiAPIKey)
		if err != nil {
			log.Printf("⚠️ Aviso: não foi possível inicializar o cliente Gemini: %v", err)
		}
	}

	// Crie o bot
	bot := &Bot{
		Session:  dg,
		Commands: setupCommands(),
		AiClient: aiClient,
		Config:   cfg,
	}

	return bot, nil
}

// Start inicia o bot
func (bot *Bot) Start() error {
	// Adicione manipuladores de eventos
	bot.Session.AddHandler(bot.onReady)
	bot.Session.AddHandler(bot.onInteraction)

	// Abra a conexão
	if err := bot.Session.Open(); err != nil {
		return fmt.Errorf("❌ Erro ao abrir conexão com Discord: %v", err)
	}

	// Registre os comandos
	if err := bot.registerCommands(); err != nil {
		return fmt.Errorf("❌ Erro ao registrar comandos: %v", err)
	}

	return nil
}

// Stop encerra o bot
func (bot *Bot) Stop() {
	if bot.AiClient != nil {
		bot.AiClient.Close()
	}
	bot.Session.Close()
}

// registerCommands registra os comandos no Discord
func (bot *Bot) registerCommands() error {
	bot.Registered = make([]*discordgo.ApplicationCommand, len(bot.Commands))

	for i, cmd := range bot.Commands {
		registeredCmd, err := bot.Session.ApplicationCommandCreate(bot.Session.State.User.ID, "", cmd)
		if err != nil {
			return fmt.Errorf("❌ Erro ao registrar comando %s: %v", cmd.Name, err)
		}
		bot.Registered[i] = registeredCmd
		log.Println("🤖 Comando registrado:", registeredCmd.Name)
	}

	return nil
}

// onReady é chamado quando o bot está pronto
func (bot *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	s.UpdateGameStatus(0, "/online para ver jogadores")
	log.Println("🤖 Bot pronto como", s.State.User.String())

	// Obtenha status inicial do servidor
	count, max := 0, 0
	stats, err := minecraft.GetServerStatus(bot.Config)
	if err == nil && stats.Online {
		count = stats.Players.Online
		max = stats.Players.Max
	}

	// Goroutine para atualizar o status periodicamente
	go bot.updateStatus(count, max)
}

// updateStatus atualiza o status do bot periodicamente
func (bot *Bot) updateStatus(initialCount, initialMax int) {
	count, max := initialCount, initialMax
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		stats, err := minecraft.GetServerStatus(bot.Config)
		if err == nil && stats.Online {
			count = stats.Players.Online
			max = stats.Players.Max
		}
		bot.Session.UpdateGameStatus(0, fmt.Sprintf("/online para ver jogadores. %d/%d", count, max))
	}
}

// respondWithError utilitário para responder com erro
func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &message,
	})
}
