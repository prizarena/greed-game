package facade

import (
	"github.com/strongo-games/greed-game/server-go/greedgame/dal"
	"github.com/strongo-games/greed-game/server-go/greedgame/models"
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/arena"
	"github.com/strongo/db"
	"github.com/strongo/log"
)

type userFacade struct {
}

var User = userFacade{}

func (userFacade) GetUserTournaments(c context.Context, user models.User) (tournaments []arena.Tournament, err error) {
	if user.UserEntity == nil {
		if user, err = dal.User.GetUserByID(c, user.ID); err != nil {
			return
		}
	}
	tournaments = make([]arena.Tournament, len(user.TournamentIDs), len(user.TournamentIDs)+1)
	entityHolders := make([]db.EntityHolder, len(tournaments))

	for i, tournamentID := range user.TournamentIDs {
		tournaments[i] = arena.Tournament{StringID: db.StringID{ID: tournamentID}}
		entityHolders[i] = &tournaments[i]
	}

	if err = dal.DB.GetMulti(c, entityHolders); err != nil {
		err = errors.WithMessage(err, "Failed to load tournament entities")
		return
	}
	if len(entityHolders) > 0 {
		log.Debugf(c, "entityHolders[0].Entity(): %v", entityHolders[0].Entity())
		log.Debugf(c, "tournaments[0].TournamentEntity: %v", tournaments[0].TournamentEntity)
	}
	return
}
