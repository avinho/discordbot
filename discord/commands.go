package discord

import (
	"github.com/bwmarrin/discordgo"
)

func setupCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
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
}

// onInteraction é chamado quando uma interação é recebida
func (bot *Bot) onInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	cmdName := i.ApplicationCommandData().Name
	cmdHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"online":    bot.handleOnlineCommand,
		"status":    bot.handleStatusCommand,
		"perguntar": bot.handlePerguntarCommand,
		"command":   bot.handleServerCommand,
	}

	if handler, exists := cmdHandlers[cmdName]; exists {
		handler(s, i)
	}
}

func hasAdminPermission(i *discordgo.InteractionCreate, adminID string) bool {
	return i.User.ID == adminID
}
