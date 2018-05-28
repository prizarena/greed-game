package gaedal

import (
	"github.com/strongo-games/greed-game/server-go/greedgame/dal"
	"github.com/strongo-games/greed-game/server-go/greedgame/models"
	"context"
)

type gameGaeDal struct {
}

var _ dal.GameDal = (*gameGaeDal)(nil)

func (gameGaeDal) GetGameByID(c context.Context, gameID string) (game models.Game, err error) {
	game.ID = gameID
	err = dal.DB.Get(c, &game)
	return
}

func (gameGaeDal) NewGame(c context.Context, entity *models.GameEntity) (game models.Game, err error) {
	game.GameEntity = entity
	err = dal.DB.InsertWithRandomStrID(c, &game, 8, 10)
	return
}
