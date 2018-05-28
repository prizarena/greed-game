package appengine

import (
	"github.com/strongo-games/greed-game/server-go/greedgame"
	"github.com/strongo/bots-framework/hosts/appengine"
	"github.com/strongo/log"
)

func init() {
	log.AddLogger(gaehost.GaeLogger)
	greedgame.InitGreedGameApp(gaehost.GaeBotHost{})
}
