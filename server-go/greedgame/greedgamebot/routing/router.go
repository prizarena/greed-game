package routing

import "github.com/strongo/bots-framework/core"

var Router = bots.NewWebhookRouter(
	map[bots.WebhookInputType][]bots.Command{
		bots.WebhookInputText:          {},
		bots.WebhookInputContact:       {},
		bots.WebhookInputCallbackQuery: {},
		//
		bots.WebhookInputInlineQuery:        {},
		bots.WebhookInputChosenInlineResult: {},
		bots.WebhookInputNewChatMembers:     {},
	},
	func() string { return "Please report any errors to @GreedGameGroup" },
)
