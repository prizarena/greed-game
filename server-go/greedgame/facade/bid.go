package facade

import (
	"github.com/prizarena/greed-game/server-go/greedgame/models"
	"github.com/prizarena/arena/arena-go"
)

type BidOutput struct {
	RivalKey            arena.BattleID
	Play                models.Play
	User                models.User
	RivalUser           models.User
	UserContestant      arena.Contestant
	RivalUserContestant arena.Contestant
}
