package gaedal

import (
	"github.com/strongo-games/greed-game/server-go/greedgame/dal"
	"github.com/strongo/db/gaedb"
)

func RegisterDal() {
	dal.DB = gaedb.NewDatabase()

	dal.Tournament = tournamentGaeDal{}
	dal.User = userGaeDal{}
	dal.Game = gameGaeDal{}
}
