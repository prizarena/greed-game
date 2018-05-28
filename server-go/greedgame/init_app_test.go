package greedgame

import (
	"github.com/strongo/bots-framework/hosts/appengine"
	"testing"
)

func TestInitGreedGameApp(t *testing.T) {
	InitGreedGameApp(gae_host.GaeBotHost{})
}
