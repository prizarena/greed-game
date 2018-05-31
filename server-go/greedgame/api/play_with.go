package api

import (
	"github.com/strongo-games/greed-game/server-go/greedgame/api/dto"
	"github.com/strongo-games/greed-game/server-go/greedgame/dal"
	"github.com/strongo-games/greed-game/server-go/greedgame/facade"
	"github.com/strongo-games/greed-game/server-go/greedgame/models"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo-games/arena/arena-go"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"net/http"
	"strconv"
	"time"
)

func verifyStrParam(c context.Context, w http.ResponseWriter, v string, name string) (err error) {
	if v == "undefined" || v == "null" {
		err = errors.New("Parameter is either 'null' or 'undefined': " + name)
		ErrorAsJson(c, w, http.StatusBadRequest, err)
	}
	return
}

func getAndVerifyRequestParam(c context.Context, w http.ResponseWriter, r *http.Request, name, defaultVal string) (v string, err error) {
	switch r.Method {
	case "POST":
		v = r.PostForm.Get(name)
	case "GET":
		v = r.URL.Query().Get(name)
	default:
		err = errors.New("Unsupported method (expected GET or POST): " + r.Method)
	}
	if err = verifyStrParam(c, w, v, name); err != nil {
		return
	} else if v == "" {
		v = defaultVal
	}
	return
}

func getTournamentAndRivalIDs(c context.Context, w http.ResponseWriter, r *http.Request) (tournamentID string, rivalID string, err error) {
	if err = r.ParseForm(); err != nil {
		err = errors.New("unable to parse form")
		ErrorAsJson(c, w, http.StatusBadRequest, err)
		return
	}

	if tournamentID, err = getAndVerifyRequestParam(c, w, r, "tournament", arena.TournamentStarID); err != nil {
		return
	}

	if rivalID, err = getAndVerifyRequestParam(c, w, r, "rival", ""); err != nil {
		return
	}

	return
}

func playPlaceBid(c context.Context, userID string, w http.ResponseWriter, r *http.Request) {
	log.Debugf(c, "playPlaceBid()")

	tournamentID, rivalID, err := getTournamentAndRivalIDs(c, w, r)
	if err != nil {
		return
	}

	bid, err := strconv.Atoi(r.PostForm.Get("bid"))
	if err != nil || bid <= 0 || bid > 100 {
		ErrorAsJson(c, w, http.StatusBadRequest, errors.WithMessage(err, "bid should be a number in 1-100 range"))
		return
	}

	now := time.Now()

	var bidOutput facade.BidOutput
	switch rivalID {
	case arena.RivalKeyStranger:
		if bidOutput, err = facade.StrangerFacade.PlaceBidAgainstStranger(c, now, userID, tournamentID, bid); err != nil {
			err = errors.WithMessage(err, "failed to place a bid against stranger")
		}
	default:
		if bidOutput, err = facade.GreedGameFacade.PlaceBidAgainstRival(c, now, userID, tournamentID, rivalID, false, bid); err != nil {
			err = errors.WithMessage(err, "failed to place a bid against rival")
		}
	}

	bidOutputToResponse(c, bidOutput, err, w)
}

func playWithdrawBid(c context.Context, userID string, w http.ResponseWriter, r *http.Request) {
	log.Debugf(c, "playWithdrawBid()")

	tournamentID, rivalID, err := getTournamentAndRivalIDs(c, w, r)
	if err != nil {
		return
	}

	bidOutput, err := facade.GreedGameFacade.Withdraw(c, tournamentID, userID, rivalID, false)
	bidOutputToResponse(c, bidOutput, err, w)
	return
}

func playNewGame(c context.Context, userID string, w http.ResponseWriter, r *http.Request) {
	log.Debugf(c, "playNewGame()")

	tournamentID, rivalID, err := getTournamentAndRivalIDs(c, w, r)
	if err != nil {
		return
	}

	user, err := facade.GreedGameFacade.NewGame(c, tournamentID, userID, rivalID)
	jsonToResponse(c, w, &dto.BidResponse{
		UserBalance: user.Tokens,
	})
}

func playQuitBattle(c context.Context, userID string, w http.ResponseWriter, r *http.Request) {
	log.Debugf(c, "playQuitBattle()")

	tournamentID, rivalID, err := getTournamentAndRivalIDs(c, w, r)
	if err != nil {
		return
	}

	bidOutput, err := facade.GreedGameFacade.Withdraw(c, tournamentID, userID, rivalID, true)
	bidOutputToResponse(c, bidOutput, err, w)
	return
}

func playBattleState(c context.Context, userID string, w http.ResponseWriter, r *http.Request) {
	log.Debugf(c, "playBattleState()")

	tournamentID, rivalID, err := getTournamentAndRivalIDs(c, w, r)
	if err != nil {
		return
	}

	user, err := dal.User.GetUserByID(c, userID)
	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}

	battles := user.GetBattles()

	battleID := arena.NewBattleID(tournamentID, rivalID)

	log.Debugf(c, "battleID: %v", battleID)

	bidResponse := dto.BidResponse{
		UserBalance: user.Tokens,
		Battle:      battles.GetBattleByRivalID(battleID),
	}
	jsonToResponse(c, w, &bidResponse)
	return
}

func currentTournamentID(t time.Time) string {
	return fmt.Sprintf("%v%v", t.Year(), t.Month())
}

func bidOutputToResponse(c context.Context, bidOutput facade.BidOutput, err error, w http.ResponseWriter) {
	log.Debugf(c, "bidOutputToResponse(err=%v) bidOutput: %+v", err, bidOutput)

	bidResponse := dto.BidResponse{}

	if err != nil {
		if db.IsNotFound(err) {
			BadRequestError(c, w, err)
			return
		} else if errors.Cause(err) == models.ErrNotEnoughTokens {
			bidResponse.Error = &dto.ErrorDto{
				Code: models.ErrNotEnoughTokens.Error(),
			}
			err = nil
		} else {
			InternalError(c, w, err)
			return
		}
	}

	user := bidOutput.User
	if user.UserEntity != nil {
		bidResponse.UserBalance = user.Tokens
		bidResponse.Battle = user.GetBattles().GetBattleByRivalID(bidOutput.RivalKey)
	}

	jsonToResponse(c, w, &bidResponse)
}
