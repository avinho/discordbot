# Discord Bot para Servidor de Minecraft

Um bot Discord desenvolvido em Go que permite monitorar e interagir com um servidor de Minecraft, além de responder perguntas sobre Minecraft usando a API Gemini da Google.

## Funcionalidades

- **Comando `/online`**: Mostra quantos jogadores estão online no servidor de Minecraft e lista seus nomes
- **Comando `/status`**: Exibe informações detalhadas sobre o status do servidor, incluindo versão, MOTD e capacidade
- **Comando `/perguntar`**: Permite fazer perguntas sobre Minecraft que são respondidas pela IA Gemini da Google

## Tecnologias Utilizadas

- [Go](https://golang.org/) - Linguagem de programação
- [DiscordGo](https://github.com/bwmarrin/discordgo) - Biblioteca para interação com a API do Discord
- [go-mcstatus](https://github.com/mcstatus-io/go-mcstatus) - Biblioteca para verificar o status de servidores Minecraft
- [Google Gemini API](https://github.com/google/generative-ai-go) - API de IA generativa para responder perguntas

## Requisitos

- Go 1.23+
- Token de bot do Discord
- Endereço IP e porta do servidor Minecraft
- Chave de API do Google Gemini

## Configuração

1. Clone o repositório
2. Configure as variáveis de ambiente necessárias:
   - `DISCORD_TOKEN`: Token do seu bot Discord
   - `SERVER_IP`: Endereço IP do servidor Minecraft
   - `SERVER_PORT`: Porta do servidor Minecraft (padrão: 46922)
   - `GEMINI_API_KEY`: Chave de API do Google Gemini

## Executando o Bot

### Localmente

```bash
go run cmd/main.go
```

### Com Docker

```bash
# Construir e iniciar o contêiner
docker-compose up -d

# Visualizar logs
docker-compose logs -f
```

## Estrutura do Projeto

```
.
├── cmd/
│   └── main.go       # Código principal do bot
├── docker-compose.yaml
├── Dockerfile
├── go.mod
└── go.sum
```
