package facade

import (
	"github.com/strongo-games/greed-game/server-go/greedgame/dal"
	"github.com/strongo-games/greed-game/server-go/greedgame/models"
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/arena"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"time"
)

type gameFacade struct {
}

var GameFacade = gameFacade{}

func decideWinnerAndUpdateEntities(
	now time.Time,
	game models.Game,
	user1, user2 *models.User,
	contestant1, contestant2 *arena.Contestant,
) {
	if user1 == nil {
		panic("user1 == nil")
	}
	if user2 == nil {
		panic("user2 == nil")
	}
	if contestant1 == nil {
		panic("contestant1 == nil")
	}
	if contestant2 == nil {
		panic("contestant2 == nil")
	}
	bid1, bid2 := game.Bids[0], game.Bids[1]
	user1.Tokens += bid1
	user2.Tokens += bid2
	if bid1 == bid2 {
		return
	}

	var minBid int
	if bid1 < bid2 {
		minBid = bid1
	} else {
		minBid = bid2
	}

	updateGame := func(winnerUserID string) {
		game.WinnerUserID = winnerUserID
	}
	game.Finished = now

	updateUsers := func(winner, loser *models.User) {
		winner.Tokens += minBid
		loser.Tokens -= minBid
	}

	updateContestants := func(winner, loser *arena.Contestant) {
		winner.CountOfGames += 1
		loser.CountOfGames += 1
		winner.RivalUserIDs.Add(loser.UserID)
		loser.RivalUserIDs.Add(winner.UserID)
	}

	bid1doubled, bid2doubled := bid1*2, bid2*2
	switch {
	case bid1 > bid2doubled || (bid1 < bid2 && bid1doubled >= bid2):
		updateGame(user1.ID)
		updateUsers(user1, user2)
		updateContestants(contestant1, contestant2)
	case bid2 > bid1doubled || (bid2 < bid1 && bid2doubled >= bid1):
		updateGame(user2.ID)
		updateUsers(user2, user1)
		updateContestants(contestant2, contestant1)
	default:
		panic("program logic error")
	}
}

func (gameFacade) NewGame(c context.Context, tournamentID, userID, rivalID string) (user models.User, err error) {
	err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		user, err = dal.User.GetUserByID(c, userID)

		battles := user.GetBattles()
		battle := battles.GetBattleByRivalID(arena.NewBattleID(tournamentID, rivalID))

		if battle != nil && battle.LastGame != nil {
			for i, b := range battles {
				if b.ID == battle.ID {
					battles[i].LastGame = nil
				}
				user.SetBattles(battles)
				if err = dal.DB.Update(c, &user); err != nil {
					return
				}
				break
			}
		}
		return
	}, db.SingleGroupTransaction)
	if err != nil {
		return
	}
	return
}

func (gameFacade) Withdraw(c context.Context, tournamentID, userID, rivalID string, quitBattle bool) (bidOutput BidOutput, err error) {
	log.Debugf(c, "gameFacade.Withdraw(tournamentID=%v, userID=%v, rivalID=%v, quitBattle=%v)", tournamentID, userID, rivalID, quitBattle)
	if tournamentID == "" {
		err = errors.New("gameFacade.Withdraw() => Invalid tournamentID")
		return
	}
	if userID == "" {
		err = errors.New("gameFacade.Withdraw() => Invalid userID")
		return
	}
	if rivalID == "" {
		err = errors.New("gameFacade.Withdraw() => Invalid rivalID")
		return
	}
	isStranger := rivalID == arena.RivalKeyStranger

	rivalKey := arena.NewBattleID(tournamentID, rivalID)

	err = dal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
		var user models.User
		if user, err = dal.User.GetUserByID(c, userID); err != nil {
			return
		}
		bidOutput.User = user

		battles := user.GetBattles()

		battleUpdated := false

		for i, battle := range battles {
			if battle.ID == rivalKey {
				if battle.Bid != nil {
					user.Tokens += battle.Bid.Value
				}
				if isStranger || quitBattle { // No need to keep battle with stranger
					battles = append(battles[:i], battles[i+1:]...)
					log.Debugf(c, "removing record for battle with ID=%v", rivalKey)
				} else {
					battles[i].Bid = nil
					log.Debugf(c, "removing bid for battle with ID=%v", rivalKey)
				}
				user.SetBattles(battles)
				if err = dal.DB.Update(c, &user); err != nil {
					return
				}
				battleUpdated = true
				break
			}
		}
		if !battleUpdated {
			log.Debugf(c, "User is not updated as no battle not found with ID=%v", rivalKey)
		}

		contestant := new(arena.Contestant)
		contestant.ID = arena.NewContestantID(tournamentID, userID)
		if err = dal.DB.Get(c, contestant); db.IsNotFound(err) {
			log.Debugf(c, errors.WithMessage(err, "Contestant record not found").Error())
			err = nil
		} else if err != nil {
			return
		} else {
			if contestant.Score == 0 && contestant.CountOfGames == 0 && contestant.RivalUserIDs == "" {
				if err = dal.DB.Delete(c, contestant); err != nil {
					return
				}
			} else if !contestant.Stranger.IsZero() {
				contestant.Stranger = time.Time{}
				if err = dal.DB.Update(c, contestant); err != nil {
					return
				}
			}
		}
		return
	}, db.CrossGroupTransaction)
	return
}

func (gameFacade) PlaceBidAgainstRival(c context.Context, now time.Time, userID, tournamentID, rivalUserID string, isStrangersGame bool, bid int) (bidOutput BidOutput, err error) {
	log.Debugf(c, "gameFacade.PlaceBidAgainstRival(userID=%v, tournamentID=%v, rivalUserID=%v, isStrangersGame=%v, bid=%v)", userID, tournamentID, rivalUserID, isStrangersGame, bid)
	if userID == "" {
		err = errors.New("Parameter userID is empty string")
		return
	}
	if rivalUserID == "" {
		err = errors.New("Parameter rivalUserID is empty string")
		return
	}
	if userID == rivalUserID {
		err = errors.New("Parameter rivalUserID is equal to userID: " + userID)
		return
	}
	if bid <= 0 || bid > 100 {
		err = errors.Errorf("Parameter bid is out of range (1-100), got %v", bid)
		return
	}
	if now.IsZero() {
		err = errors.New("Parameter now is empty")
		return
	}

	gamesSessionID := models.GamesSessionIdFromUserIDs(userID, rivalUserID)
	//panic(gamesSessionID)
	var (
		user, rivalUser                     models.User
		userContestant, rivalUserContestant arena.Contestant
	)
	user1, user2 := new(models.User), new(models.User)
	if user1.ID, user2.ID, err = models.UserIDsFromGameSessionID(gamesSessionID); err != nil {
		return
	} else if userID != user1.ID && userID != user2.ID {
		err = errors.New("attempt to bid for someone's else games session")
		return
	}

	contestant1, contestant2 := new(arena.Contestant), new(arena.Contestant)
	contestant1.ID = arena.NewContestantID(tournamentID, user1.ID)
	contestant2.ID = arena.NewContestantID(tournamentID, user2.ID)

	err = dal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
		if err = dal.DB.GetMulti(c, []db.EntityHolder{user1, user2}); err != nil {
			return
		}

		var entitiesToUpdate []db.EntityHolder

		getContestant := func(contestant *arena.Contestant) (err error) {
			if err = dal.DB.Get(c, contestant); db.IsNotFound(err) {
				err = nil
				contestant.ContestantEntity = &arena.ContestantEntity{
					TimeJoined:   now,
					TournamentID: tournamentID,
					UserID:       arena.ContestantID(contestant.ID).UserID(),
				}
			} else if isStrangersGame {
				contestant.Stranger = time.Time{}
			}
			return
		}
		if err = getContestant(contestant1); err != nil {
			return
		}
		if err = getContestant(contestant2); err != nil {
			return
		}
		switch userID {
		case user1.ID:
			user = *user1
			rivalUser = *user2
			userContestant = *contestant1
			rivalUserContestant = *contestant2
		case user2.ID:
			user = *user2
			rivalUser = *user1
			userContestant = *contestant2
			rivalUserContestant = *contestant1
			//
		default:
			panic("userID is not equal either it user1.ID or user2.ID")
		}
		bidOutput.User = user
		bidOutput.RivalUser = rivalUser
		bidOutput.UserContestant = userContestant
		bidOutput.UserContestant = rivalUserContestant

		var (
			rivalPrevKey, rivalNextKey arena.BattleID
			rivalUserRivalPrevKey      arena.BattleID
			rivalUserRivalNextKey      arena.BattleID
			userBattle                 models.Battle
			userBattles                []models.Battle
		)

		if isStrangersGame {
			rivalPrevKey = arena.NewStrangerBattleID(tournamentID)
			rivalUserRivalPrevKey = arena.NewStrangerBattleID(tournamentID)
		} else {
			rivalPrevKey = arena.NewBattleID(tournamentID, rivalUserID)
			rivalUserRivalPrevKey = arena.NewBattleID(tournamentID, userID)
		}
		rivalNextKey = arena.NewBattleID(tournamentID, rivalUserID)
		rivalUserRivalNextKey = arena.NewBattleID(tournamentID, userID)

		bidOutput.RivalKey = rivalNextKey

		if userBattle, userBattles, err = user.RecordBid(rivalPrevKey, bid, now); err != nil {
			return
		}

		rivalUserBattles := rivalUser.GetBattles()

		var (
			rivalUserBattleIndex int
			rivalUserBattle      models.Battle
			rivalUserBid         *models.Bid
		)
		for rivalUserBattleIndex, rivalUserBattle = range rivalUserBattles {
			if rivalUserBattle.ID == rivalUserRivalPrevKey {
				if rivalUserBattle.Bid == nil {
					if isStrangersGame {
						return errors.WithMessage(arena.ErrRivalUserIsNotBiddingAgainstStranger, "rivalUserBattle.Bid == nil")
					}
					// rival has no bid yet
				} else {
					rivalUserBid = rivalUserBattle.Bid
				}
				break
			}
		}

		if rivalUserBid == nil || rivalUserBid.Value == 0 {
			if isStrangersGame { // Check the rival user has an open bid for a stranger
				return errors.WithMessage(arena.ErrRivalUserIsNotBiddingAgainstStranger, "rivalBid == 0 && isStrangersGame")
			}
			// No matching bids
			entitiesToUpdate = append(entitiesToUpdate, &bidOutput.User)
			rivalUserBattles = append(rivalUserBattles, models.Battle{
				ID:   rivalUserRivalNextKey,
				Name: user.Name,
			})
		} else {
			// RivalStat user already placed a bid
			entitiesToUpdate = append(entitiesToUpdate, user1, user2)

			bidOutput.Game.GameEntity = &models.GameEntity{
				Strangers: isStrangersGame,
				Created:   rivalUserBid.Time,
				UserIDs: []string{
					user1.ID,
					user2.ID,
				},
			}
			bidOutput.Game.SetBid(rivalUserID, rivalUserBid.Value)
			bidOutput.Game.SetBid(userID, bid)

			decideWinnerAndUpdateEntities(now, bidOutput.Game, user1, user2, contestant1, contestant2)

			entitiesToUpdate = append(entitiesToUpdate, contestant1, contestant2)

			if bidOutput.Game, err = dal.Game.NewGame(c, bidOutput.Game.GameEntity); err != nil {
				return
			}

			updateUserGameStats(bidOutput.Game, user, tournamentID, rivalUserID)
			updateUserGameStats(bidOutput.Game, rivalUser, tournamentID, userID)

			rivalUserBattle.ID = arena.NewBattleID(tournamentID, userID)
			rivalUserBattles[rivalUserBattleIndex] = updateBattleBidLastGameRivalNameAndOldID(rivalUserBattle, bidOutput.Game, rivalUser, user, rivalUserBid, userBattle.Bid)
			rivalUser.SetBattles(rivalUserBattles)
		}

		{ // Update user battles (either remove bid or replace stranger battle with rival battle
			rivalBattleFound := false
			for i, battle := range userBattles {
				if battle.ID == rivalNextKey {
					if isStrangersGame {
						err = errors.Errorf("an attempt to play with known rival as with stranger")
						return
					}
					rivalBattleFound = true
					userBattle = updateBattleBidLastGameRivalNameAndOldID(battle, bidOutput.Game, user, rivalUser, userBattle.Bid, rivalUserBid)
					userBattles[i] = userBattle
					break
				} else if isStrangersGame && battle.ID == rivalPrevKey {
					rivalBattleFound = true
					battle.ID = rivalNextKey
					userBattle = updateBattleBidLastGameRivalNameAndOldID(battle, bidOutput.Game, user, rivalUser, userBattle.Bid, rivalUserBid)
					userBattles[i] = userBattle
					log.Debugf(c, "Stranger battle %v replaced with rival one: %v", rivalPrevKey, rivalNextKey)
				}
			}

			if !rivalBattleFound {
				userBattle = updateBattleBidLastGameRivalNameAndOldID(models.Battle{ID: rivalNextKey}, bidOutput.Game, user, rivalUser, userBattle.Bid, rivalUserBid)
				userBattles = append([]models.Battle{userBattle}, userBattles[:]...)
				log.Debugf(c, "Battle added for user: %v", rivalNextKey)
			}

			user.SetBattles(userBattles)
		}

		if len(entitiesToUpdate) == 1 {
			if err = dal.DB.Update(c, entitiesToUpdate[0]); err != nil {
				return
			}
		} else {
			if err = dal.DB.UpdateMulti(c, entitiesToUpdate); err != nil {
				return
			}
		}
		return
	}, db.CrossGroupTransaction)

	if err != nil {
		err = errors.WithMessage(err, "failed to place a bid")
		return
	}

	log.Debugf(c, "A bid successfully placed on game #"+bidOutput.Game.ID)
	return
}

func updateUserGameStats(game models.Game, user models.User, tournamentID, rivalUserID string) {
	var balanceDiff int
	if game.WinnerUserID == user.ID {
		balanceDiff = game.Prize()
	} else if game.WinnerUserID != "" {
		balanceDiff = -game.Prize()
	}
	user.UpdateArenaStats(tournamentID, rivalUserID, game.ID, balanceDiff)
}

func updateBattleBidLastGameRivalNameAndOldID(battle models.Battle, game models.Game, user, rivalUser models.User, userBid, rivalBid *models.Bid) models.Battle {
	if rivalBid != nil && rivalBid.Value != 0 { // If both users have bids
		battle.LastGame = &models.LastGame{
			ID:       game.ID,
			UserBid:  userBid.Value,
			RivalBid: rivalBid.Value,
			Prize:    game.Prize(),
			Time:     game.Finished,
			Status:   GetGameStatus(user.ID, game.WinnerUserID),
		}
		battle.Bid = nil // Clear user's battle from bid
		if game.Strangers {
			battle.OldID = arena.NewBattleID(game.TournamentID, "*")
		}
	} else if battle.LastGame != nil {
		battle.LastGame = nil
	}

	if battle.Name != rivalUser.Name && rivalUser.Name != "" {
		battle.Name = rivalUser.Name
	}
	return battle
}

func GetGameStatus(userID, winnerID string) string {
	switch winnerID {
	case userID:
		return "win"
	case "":
		return "tie"
	default:
		return "loss"
	}
}
