package dal

import (
	"github.com/prizarena/greed-game/server-go/greedgame/models"
	"context"
	"github.com/strongo/db"
)

type UserDal interface {
	GetUserByID(c context.Context, userID string) (user models.User, err error)
}

//type ContestantDal interface {
//	GetContestant(c context.Context, tournamentID, userID string) (user models.Tournament, err error)
//}

type PlayDal interface {
	GetPlayByID(c context.Context, playID string) (play models.Play, err error)
	NewPlay(c context.Context, entity *models.PlayEntity) (play models.Play, err error)
}

//type GamesSessionDal interface {
//	GetGamesSessionByID(c context.Context, gamesSessionID string) (gamesSession models.GamesSession, err error)
//}

var (
	DB         db.Database
	User UserDal
	Play PlayDal
)
