package gaedal

import (
	"github.com/prizarena/greed-game/server-go/greedgame/dal"
	"github.com/prizarena/greed-game/server-go/greedgame/models"
	"context"
)

type playGaeDal struct {
}

var _ dal.PlayDal = (*playGaeDal)(nil)

func (playGaeDal) GetPlayByID(c context.Context, playID string) (play models.Play, err error) {
	play.ID = playID
	err = dal.DB.Get(c, &play)
	return
}

func (playGaeDal) NewPlay(c context.Context, entity *models.PlayEntity) (play models.Play, err error) {
	play.PlayEntity = entity
	err = dal.DB.InsertWithRandomStrID(c, &play, 8, 10, "")
	return
}
