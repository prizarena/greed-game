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

type GameDal interface {
	GetGameByID(c context.Context, gameID string) (game models.Game, err error)
	NewGame(c context.Context, entity *models.GameEntity) (game models.Game, err error)
}

//type GamesSessionDal interface {
//	GetGamesSessionByID(c context.Context, gamesSessionID string) (gamesSession models.GamesSession, err error)
//}

var (
	DB         db.Database
	//Tournament pa.TournamentDal
	//Contestant   TournamentDal
	User UserDal
	Game GameDal
	//GamesSession GamesSessionDal
)
