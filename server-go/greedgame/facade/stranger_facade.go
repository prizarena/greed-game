package facade

import (
	"github.com/prizarena/greed-game/server-go/greedgame/dal"
	"github.com/prizarena/greed-game/server-go/greedgame/models"
	"context"
	"github.com/pkg/errors"
	"github.com/prizarena/arena/arena-go"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"time"
	"github.com/prizarena/prizarena-public/prizarena-client-go"
)

type strangerFacade struct {
}

var (
	StrangerFacade = strangerFacade{}
)

func (sf strangerFacade) PlaceBidAgainstStranger(c context.Context, now time.Time, userID, tournamentID string, bid int) (bidOutput BidOutput, err error) {
	log.Debugf(c, "strangerFacade.PlaceBidAgainstStranger(userID=%v, tournamentID=%v, bid=%v)", userID, tournamentID, bid)
	if bid <= 0 || bid > 100 {
		err = errors.New("bid must be in range 1-100")
		return
	}

	onRivalFound := func(rivalUserID string) (err error) {
		log.Debugf(c, "strangerFacade.PlaceBidAgainstStranger() => will link 2 strangers")
		bidOutput, err = GreedGameFacade.PlaceBidAgainstRival(c, now, userID, tournamentID, rivalUserID, true, bid)
		return
	}

	onStranger := func() error {
		err = sf.registerNewStranger(c, now, bid, &bidOutput, tournamentID, userID)
		return err
	}

	user := models.User{StringID: db.NewStrID(userID)}

	if err = prizarena.NewFacade(nil).MakeMoveAgainstStranger(c, tournamentID, userID, onRivalFound, onStranger); err != nil {
		return
	}

	return
}

func (strangerFacade) registerNewStranger(c context.Context, now time.Time, bid int, bidOutput *BidOutput, tournamentID, userID string) (err error) {
	var (
		user        models.User
		userBattles []models.Battle
	)
	updateUser := func(tc context.Context, strangerRivalKey arena.BattleID) (userEntityHolder db.EntityHolder, err error) {
		if user, err = dal.User.GetUserByID(c, userID); err != nil {
			return
		}
		userEntityHolder = &user
		bidOutput.User = user
		if _, userBattles, err = user.RecordBid(strangerRivalKey, bid, now); err != nil {
			return
		}
		user.SetBattles(userBattles)
		return
	}


	return arena.RegisterStranger(c, now, tournamentID, userID, contestant, updateUser)
}
