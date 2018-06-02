package api

import (
	"github.com/prizarena/greed-game/server-go/greedgame/api/dto"
	"github.com/prizarena/greed-game/server-go/greedgame/dal"
	"github.com/prizarena/greed-game/server-go/greedgame/facade"
	"context"
	"encoding/json"
	"net/http"
)

func userFullState(c context.Context, userID string, w http.ResponseWriter, _ *http.Request) {
	user, err := dal.User.GetUserByID(c, userID)
	if err != nil {
		ErrorAsJson(c, w, http.StatusGone, err)
		return
	}
	userFullState := dto.UserFullState{
		UserBriefState: dto.UserBriefState{
			Balance: user.Tokens,
		},
		Battles: json.RawMessage(user.BattlesJson),
	}
	{ // populate tournaments
		tournaments, err := facade.User.GetUserTournaments(c, user)
		if err != nil {
			ErrorAsJson(c, w, http.StatusGone, err)
			return
		}
		userFullState.Tournaments = tournamentsToDto(tournaments)
	}
	jsonToResponse(c, w, userFullState)
}
