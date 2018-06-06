package facade

import (
	"github.com/prizarena/greed-game/server-go/greedgame/dal"
	"github.com/prizarena/greed-game/server-go/greedgame/dal/gaedal"
	"github.com/prizarena/greed-game/server-go/greedgame/defaults"
	"github.com/prizarena/greed-game/server-go/greedgame/models"
	"context"
	"github.com/strongo/db"
	"github.com/strongo/db/mockdb"
	"testing"
	"time"
)

type mockStrangerDal struct {
	//originalDal        dal.TournamentDal
	findStrangerResult string
}

func (mock mockStrangerDal) FindStranger(c context.Context, tournamentID, userID string, friends []string) (strangerID string, err error) {
	return mock.findStrangerResult, nil
}

func TestStrangerFacade_PlaceBidAgainstStranger(t *testing.T) {
	gaedal.RegisterDal()
	dal.DB = mockdb.NewMockDB(nil, nil)
	c := context.Background()
	f := strangerFacade{}

	now := time.Now()

	const tournamentID = "tournament1"

	t.Run("zero-bid", func(t *testing.T) {
		if _, err := f.PlaceBidAgainstStranger(c, now, "test-user", tournamentID, 0); err == nil {
			t.Fatal("zero bids are not allowed")
		}
	})

	t.Run("too-big-bid", func(t *testing.T) {
		if _, err := f.PlaceBidAgainstStranger(c, now, "test-user", tournamentID, 101); err == nil {
			t.Fatal("zero bids are not allowed")
		}
	})

	t.Run("unknown-user", func(t *testing.T) {
		if _, err := f.PlaceBidAgainstStranger(c, now, "test-user", tournamentID, 20); err == nil {
			t.Fatal("unknown users are not allowed")
		}
	})

	t.Run("known-user-no-strangers", func(t *testing.T) {
		user1 := models.User{
			StringID: db.StringID{ID: "user1"},
			UserEntity: &models.UserEntity{
				Name:   "First user",
				Tokens: defaults.UserStartingTokensBalance,
			},
		}
		if err := dal.DB.Update(c, &user1); err != nil {
			t.Fatal(err)
		}

		dal.Tournament = mockStrangerDal{
			//originalDal:        dal.Tournament,
			findStrangerResult: "",
		}

		const (
			bid1 = 20
		)
		if _, err := f.PlaceBidAgainstStranger(c, now, "user1", tournamentID, bid1); err != nil {
			t.Error(err)
		} else {
			user1, err = dal.User.GetUserByID(c, "user1")
			if err = dal.DB.Get(c, &user1); err != nil {
				t.Error(err)
				return
			}

			if expected := defaults.UserStartingTokensBalance - bid1; user1.Tokens != expected {
				t.Errorf("Expected user balance to be %v, got: %v", expected, user1.Tokens)
			}

			if rivals := user1.GetBattles(); len(rivals) != 1 {
				t.Errorf("len(rivals): %v != 1: %+v", len(rivals), rivals)
			} else if rival := rivals.GetBattleByRivalID(models.NewStrangerBattleKey(tournamentID)); rival == nil {
				t.Error("not stranger")
			} else if rival.Bid.Value != bid1 {
				t.Errorf("unexpected bid: %v", rival.Bid.Value)
			}
		}
	})

	t.Run("known-user-and-valid-stranger", func(t *testing.T) {
		user1 := models.User{
			StringID: db.StringID{ID: "user1"},
			UserEntity: &models.UserEntity{
				Name:   "First User",
				Tokens: defaults.UserStartingTokensBalance,
			},
		}
		if err := dal.DB.Update(c, &user1); err != nil {
			t.Fatal(err)
		} else if user1.Tokens != defaults.UserStartingTokensBalance {
			t.Fatal("user1.Tokens !+ defaults.UserStartingTokensBalance")
		}

		const bid1, bid2 = 20, 25
		user2 := models.User{
			StringID: db.StringID{ID: "user2"},
			UserEntity: &models.UserEntity{
				Name:   "Second User",
				Tokens: defaults.UserStartingTokensBalance - bid2,
			},
		}
		user2rows := user2.GetBattles()
		user2rows = append(user2rows, models.Battle{
			ID: models.NewStrangerBattleKey(tournamentID),
			Bid: &models.Bid{
				Value: bid2,
				Time:  now,
			},
		})
		user2.SetBattles(user2rows)
		if err := dal.DB.Update(c, &user2); err != nil {
			t.Fatal(err)
		}

		dal.Tournament = mockStrangerDal{
			//originalDal:        dal.Tournament, //
			findStrangerResult: "user2",
		}

		if bidOutput, err := f.PlaceBidAgainstStranger(c, now, "user1", tournamentID, bid1); err != nil {
			t.Error(err)
		} else if bidOutput.User.ID != user1.ID {
			t.Errorf("Updated wrong user expected '%v', got: '%v'", bidOutput.User.ID, user1.ID)
		} else {
			user1, err = dal.User.GetUserByID(c, "user1")
			if err = dal.DB.Get(c, &user1); err != nil {
				t.Error(err)
				return
			}
			if expected := defaults.UserStartingTokensBalance + bid1; user1.Tokens != expected {
				t.Errorf("Expected user1 balance to be %v, got: %v", expected, user1.Tokens)
			}

			user2, err = dal.User.GetUserByID(c, "user2")
			if err = dal.DB.Get(c, &user2); err != nil {
				t.Error(err)
				return
			}
			if expected := defaults.UserStartingTokensBalance - bid1; user2.Tokens != expected {
				t.Errorf("Expected user2 balance to be %v, got: %v", expected, user2.Tokens)
			}

			if bidOutput.Play.ID == "" {
				t.Error("bidOutput.Game.ID is empty")
			}

			if !bidOutput.Play.HasBothBids() {
				t.Error("bidOutput.Game.HasBothBids(): false")
			}

			validateUser := func(user models.User, name, rivalUserID string) {
				rivalKey := models.NewBattleKey(tournamentID, rivalUserID)
				if rivals := user.GetRivalStats(); len(rivals) != 1 {
					t.Error(name + " unexpected rivals")
				} else if rival, ok := rivals[models.NewBattleKey(tournamentID, rivalUserID)]; !ok {
					t.Errorf("%v{ID: %v} misses rival with ID=%v. RivalStats: %+v", name, user.ID, rivalKey, rivals)
				} else if rival.GamesCount != 1 {
					t.Errorf("%v.rivals[%v].CountOfPlaysCompleted != 1", name, user.ID)
				} else if rival.Balance == 0 {
					t.Errorf("%v.rivals[%v].Balance == 0", name, user.ID)
				}
			}
			validateUser(bidOutput.User, "bidOutput.User", bidOutput.RivalUser.ID)
			validateUser(bidOutput.RivalUser, "bidOutput.RivalUser", bidOutput.User.ID)

			validateBattle := func(user models.User, rivalUser models.User) {
				battles := user.GetBattles()
				if len(battles) != 1 {
					t.Errorf("Expected 1 battle, got %v: %+v", len(battles), battles)
				}
				battle := battles.GetBattleByRivalID(models.NewBattleKey(tournamentID, rivalUser.ID))
				if battle == nil {
					t.Errorf("User %v has no battle with %v", user.ID, rivalUser.ID)
					return
				}
				if battle.IsWithStranger() {
					t.Errorf("battle ID should be agains rival, not stranger, got: %v", battle.ID)
				}
				if battle.OldID != "*@*" {
					t.Errorf("user(%v): battle should have OldID equal to '*@*'", user.ID)
				}
				if battle.Bid != nil {
					t.Errorf("After game played battle's bid should be nil, user: %v", user.ID)
				}

				{ // verify battle's last game
					lastGame := battle.LastGame

					if lastGame == nil {
						t.Errorf("Battle should have info on last game, user: %v", user.ID)
					}
					if lastGame.Prize != 20 {
						t.Errorf("user(%v): lastGame.Prize expected to be 20, got: %v", user.ID, lastGame.Prize)
					}
					if lastGame.ID == "" {
						t.Errorf("battle.LastGame.ID is empty: %v", user.ID)
					} else if lastGame.ID != bidOutput.Play.ID {
						t.Errorf("battle.LastGame.ID != bidOutput.Game.ID: %v != %v", lastGame.ID, bidOutput.Play.ID)
					}
					if lastGame.Time.IsZero() {
						t.Errorf("battle.LastGame.Time is zero")
					}
					if lastGame.Status == "" {
						t.Errorf("battle.LastGame.Status is empty")
					}
					if lastGame.UserBid <= 0 {
						t.Errorf("lastGame.UserBid <= 0: %v", lastGame.UserBid)
					}
					if lastGame.RivalBid <= 0 {
						t.Errorf("lastGame.RivalBid <= 0: %v", lastGame.RivalBid)
					}
				}
			}

			validateBattle(bidOutput.User, bidOutput.RivalUser)
			validateBattle(bidOutput.RivalUser, bidOutput.User)

			//if !strings.Contains(gamesSession.ID, user1.ID) || !strings.Contains(gamesSession.ID, user2.ID) {
			//	t.Error("gamesSession.ID should contains both player's user IDs")
			//}
			//
			//if gamesSession.LastGameID == "" {
			//	t.Errorf("gamesSession.LastGameID is not set")
			//}
			//
			//if gamesSession.ActiveGameID != "" {
			//	t.Errorf("gamesSession.ActiveGameID is not empty")
			//}
			//
			//if gamesSession.GamesPlayed != 1 {
			//	t.Errorf("gamesSession.CountOfPlaysCompleted expected to be 1, got %v", gamesSession.GamesPlayed)
			//}
			//
			//if game.ID == "" {
			//	t.Error("game.ID is not set")
			//}
			//if game.WinnerUserID != "user1" {
			//	t.Errorf("Unexpected games.WinnerUserID: %v", game.WinnerUserID)
			//}
			//if len(user1.RivalStats) != 1 || user1.RivalStats[0] != user2.ID {
			//	t.Error("user2 is not a rival of user1")
			//}
			//if len(user2.RivalStats) != 1 || user2.RivalStats[0] != user1.ID {
			//	t.Error("user1 is not a rival of user2")
			//}
		}
	})
}
