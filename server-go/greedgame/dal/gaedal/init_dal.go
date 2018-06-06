package gaedal

import (
	"github.com/prizarena/greed-game/server-go/greedgame/dal"
	"github.com/strongo/db/gaedb"
)

func RegisterDal() {
	dal.DB = gaedb.NewDatabase()

	dal.User = userGaeDal{}
	dal.Play = playGaeDal{}
}
