package gaedal

import (
	"github.com/prizarena/greed-game/server-go/greedgame/dal"
	"github.com/strongo/db/gaedb"
)

func RegisterDal() {
	dal.DB = gaedb.NewDatabase()

	dal.Tournament = tournamentGaeDal{}
	dal.User = userGaeDal{}
	dal.Game = gameGaeDal{}
}
