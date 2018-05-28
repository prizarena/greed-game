package gaedal

import (
	"github.com/strongo-games/greed-game/server-go/greedgame/dal"
	"github.com/strongo-games/greed-game/server-go/greedgame/models"
	"context"
)

type userGaeDal struct {
}

func (userGaeDal) GetUserByID(c context.Context, userID string) (user models.User, err error) {
	user.ID = userID
	err = dal.DB.Get(c, &user)
	return
}
