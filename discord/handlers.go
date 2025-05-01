package discord

import (
	"context"
	"discordbot/minecraft"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/generative-ai-go/genai"
)

func (bot *Bot) handleOnlineCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Responde imediatamente para indicar que o bot está processando
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	if err != nil {
		log.Printf("Erro ao enviar resposta inicial: %v", err)
		return
	}

	count, players, err := minecraft.GetOnlinePlayers(bot.Config)
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

// handleStatusCommand processa o comando /status
func (bot *Bot) handleStatusCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	srv, err := minecraft.GetServerStatus(bot.Config)
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

// handlePerguntarCommand processa o comando /perguntar
func (bot *Bot) handlePerguntarCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	pergunta := options[0].StringValue()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	if bot.AiClient == nil {
		respondWithError(s, i, "❌ Serviço de IA não está disponível")
		return
	}

	ctx := context.Background()
	model := bot.AiClient.GenerativeModel("gemini-2.0-flash")
	prompt := fmt.Sprintf("Responda como um especialista em Minecraft. Sempre responda em português. Ao responder, não mencione que você é um modelo de IA. Adicione Emoji ao final da resposta de acordo com o contexto da resposta. Pergunta:\n\n%s\n\n", pergunta)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		respondWithError(s, i, "❌ Erro ao gerar resposta")
		return
	}

	response := resp.Candidates[0].Content.Parts[0].(genai.Text)
	responseStr := string(response)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &responseStr,
	})
}

// handleServerCommand processa o comando /command
func (bot *Bot) handleServerCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasAdminPermission(i, bot.Config.AdminUserID) {
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

	err := minecraft.ExecuteServerCommand(command, bot.Config)

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

// handleSessionsCommand processa o comando /sessions
func (bot *Bot) handleSessionsCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasAdminPermission(i, bot.Config.AdminUserID) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Você não tem permissão para listar sessões",
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	sessions, err := minecraft.ListScreenSessions(bot.Config)

	var msg string
	if err != nil {
		msg = fmt.Sprintf("❌ Erro ao listar sessões screen: %v", err)
	} else {
		msg = fmt.Sprintf("**Sessões Screen Disponíveis** 📋\n```\n%s\n```", sessions)
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &msg,
	})
}

// handleDiagnoseCommand processa o comando /diagnose
func (bot *Bot) handleDiagnoseCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !hasAdminPermission(i, bot.Config.AdminUserID) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Você não tem permissão para executar diagnósticos",
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// Colete informações de diagnóstico
	var diagnosticInfo strings.Builder

	// 1. Verifique as sessões screen
	screenSessions, err := minecraft.ListScreenSessions(bot.Config)
	if err != nil {
		diagnosticInfo.WriteString(fmt.Sprintf("❌ Erro ao verificar sessões screen: %v\n", err))
	} else {
		diagnosticInfo.WriteString("**Sessões Screen Disponíveis:**\n```\n")
		diagnosticInfo.WriteString(screenSessions)
		diagnosticInfo.WriteString("```\n")
	}

	// 2. Verifique o processo do servidor Minecraft
	psOutput, err := minecraft.ExecuteHostCommand("ps aux | grep -i minecraft | grep -v grep", bot.Config)
	if err != nil {
		diagnosticInfo.WriteString(fmt.Sprintf("❌ Erro ao verificar processos: %v\n", err))
	} else {
		diagnosticInfo.WriteString("**Processos Minecraft:**\n```\n")
		if psOutput != "" {
			diagnosticInfo.WriteString(psOutput)
		} else {
			diagnosticInfo.WriteString("Nenhum processo Minecraft encontrado")
		}
		diagnosticInfo.WriteString("```\n")
	}

	// 3. Verifique o status da conexão com o servidor
	serverStatus, err := minecraft.GetServerStatus(bot.Config)
	if err != nil {
		diagnosticInfo.WriteString(fmt.Sprintf("❌ Erro ao conectar ao servidor: %v\n", err))
	} else {
		diagnosticInfo.WriteString(fmt.Sprintf("**Status do Servidor:** %s\n",
			func() string {
				if serverStatus.Online {
					return "✅ Online"
				}
				return "❌ Offline"
			}()))
	}

	// 4. Verifique a configuração SSH
	diagnosticInfo.WriteString(fmt.Sprintf("**Configuração SSH:**\n```\n"+
		"Usuário: %s\n"+
		"Host: %s\n"+
		"Porta: %s\n"+
		"Sessão Screen: %s\n"+
		"```\n",
		bot.Config.SSHUser,
		bot.Config.SSHHost,
		bot.Config.SSHPort,
		bot.Config.ScreenName))

	diagString := diagnosticInfo.String()
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &diagString,
	})
}
