package main

import (
	"discordbot/config"
	"discordbot/discord"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Carregue a configuração
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Erro ao carregar configuração: %v", err)
	}

	// Inicialize o bot
	bot, err := discord.NewBot(cfg)
	if err != nil {
		log.Fatalf("❌ Erro ao inicializar o bot: %v", err)
	}

	// Inicie o bot
	if err := bot.Start(); err != nil {
		log.Fatalf("❌ Erro ao iniciar o bot: %v", err)
	}
	defer bot.Stop()

	// Aguarde sinal para encerrar
	fmt.Println("🤖 Bot está rodando. Pressione CTRL-C para sair.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
