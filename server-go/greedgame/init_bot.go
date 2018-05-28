package greedgame

import (
	"github.com/strongo-games/greed-game/server-go/greedgame/greedgamebot/platforms/tg"
	"github.com/strongo-games/greed-game/server-go/greedgame/greedgamebot/routing"
	"context"
	"github.com/DebtsTracker/translations/trans"
	"github.com/julienschmidt/httprouter"
	"github.com/strongo/app"
	"github.com/strongo/app/gaestandard"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
)

func newTranslator(c context.Context) strongo.Translator {
	return strongo.NewMapTranslator(c, trans.TRANS)
}

func initBot(httpRouter *httprouter.Router, botHost bots.BotHost, appContext bots.BotAppContext) {

	driver := bots.NewBotDriver( // Orchestrate requests to appropriate handlers
		bots.AnalyticsSettings{GaTrackingID: ""}, // TODO: Refactor to list of analytics providers
		appContext, // Holds User entity kind name, translator, etc.
		botHost,    // Defines how to create context.Context, HttpClient, DB, etc...
		"Please report any issues to @GreedGameGroup", // TODO: Is it wrong place? Router has similar.
	)

	driver.RegisterWebhookHandlers(httpRouter, "/bot",
		telegram.NewTelegramWebhookHandler(
			telegramBotsWithRouter, // Maps of bots by code, language, token, etc...
			newTranslator,          // Creates translator that gets a context.Context (for logging or RPC purpose)
		),
	)
}

func telegramBotsWithRouter(c context.Context) bots.SettingsBy {
	return tg.Bots(gaestandard.GetEnvironment(c), func(profile string) bots.WebhooksRouter {
		switch profile {
		case "greedgame":
			return routing.Router
		default:
			panic("Unknown profile: " + profile) // GreedGame bots use single profile
		}
	})
}
