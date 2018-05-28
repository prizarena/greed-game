package tg

import (
	"fmt"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
)

var _bots bots.SettingsBy

func Bots(environment strongo.Environment, router func(profile string) bots.WebhooksRouter) bots.SettingsBy { //TODO: Consider to do pre-deployment replace
	if len(_bots.ByCode) == 0 || (!_bots.HasRouter && router != nil) {
		//log.Debugf(c, "Bots() => hostname:%v, environment:%v:%v", hostname, environment, strongo.EnvironmentNames[environment])
		switch environment {
		case strongo.EnvProduction:
			_bots = bots.NewBotSettingsBy(router,
				// Production bots
				telegram.NewTelegramBot(strongo.EnvProduction, "greedgame", "GreedGameBot", "TODO:get-code", "", "", "", strongo.LocaleEnUS),
			)
		case strongo.EnvDevTest:
		case strongo.EnvStaging:
		case strongo.EnvLocal:
			_bots = bots.NewBotSettingsBy(router,
				// Staging bots
				telegram.NewTelegramBot(strongo.EnvLocal, "greedgame", "GreedGameBot", "TODO:get-code", "", "", "", strongo.LocaleEnUS),
			)
		default:
			panic(fmt.Sprintf("Unknown environment => %v:%v", environment, strongo.EnvironmentNames[environment]))
		}
	}
	return _bots
}
