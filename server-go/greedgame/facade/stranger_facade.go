package facade

import (
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/log"
	"time"
	"github.com/prizarena/prizarena-public/prizarena-client-go"
	"github.com/prizarena/prizarena-public/prizarena-interfaces"
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

	prizarenaFacade := prizarena.NewFacade(nil)
	move := prizarena_interfaces.MoveDto{Bid: bid}
	onRivalFound := func(rivalUserID string, move *prizarena_interfaces.MoveDto) (err error) {
		log.Debugf(c, "strangerFacade.PlaceBidAgainstStranger() => will link 2 strangers")
		bidOutput, err = GreedGameFacade.PlaceBidAgainstRival(c, now, userID, tournamentID, rivalUserID, true, bid)
		return
	}
	if err = prizarenaFacade.MakeMoveAgainstStranger(c, tournamentID, userID, &move, onRivalFound, nil); err != nil {
		return
	}

	return
}
