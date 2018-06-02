package models

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/app"
	"github.com/strongo/app/user"
	"github.com/prizarena/arena/arena-go"
	"github.com/strongo/db"
	"time"
)

const (
	UserKind = "User"
	//strangerRivalBidKey = "$tranger"
)

type UserEntity struct {
	strongo.AppUserBase
	user.AccountsOfUser

	Name        string `datastore:",noindex,omitempty"`
	Created     time.Time
	AvatarURL   string `datastore:",noindex,omitempty"`
	FirebaseUID string `datastore:",omitempty"`
	Tokens      int
	//
	//
	TournamentIDs []string `datastore:",noindex"`
	BattlesJson   string   `datastore:",noindex,omitempty"`
	BattlesCount  int      `datastore:",noindex,omitempty"`
	RivalStats    string   `datastore:",noindex,omitempty"`
	//
	arena.UserContestantEntity
}

type User struct {
	db.StringID
	*UserEntity
}

var _ db.EntityHolder = (*User)(nil)

func (User) Kind() string {
	return UserKind
}

func (User) NewEntity() interface{} {
	return new(UserEntity)
}

func (u User) Entity() interface{} {
	return u.UserEntity
}

func (u *User) SetEntity(v interface{}) {
	if v == nil {
		u.UserEntity = nil
	} else {
		u.UserEntity = v.(*UserEntity)
	}

}

func (u *UserEntity) SetBotUserID(platform, botID, botUserID string) {
	u.AddAccount(user.Account{
		Provider: platform,
		App:      botID,
		ID:       botUserID,
	})
}

func (u *User) GetBattles() (battles Battles) {
	if u.BattlesJson == "" {
		battles = make(Battles, 0, 1)
		return
	}
	battles = make(Battles, 0, u.BattlesCount)
	if err := ffjson.Unmarshal([]byte(u.BattlesJson), &battles); err != nil {
		panic(err)
	}
	return
}

func (u *User) SetBattles(battles []Battle) {
	if len(battles) == 0 {
		u.BattlesJson = ""
		u.BattlesCount = 0
		return
	}
	{ // Perform data integrity checks
		idsCount := make(map[arena.BattleID]int, len(battles)) // Check for duplicates
		for i, b := range battles {
			if b.ID == "" {
				panic(fmt.Sprintf("battle has no ID, index=%v, battles: %+v", i, battles))
			}
			if count := idsCount[b.ID] + 1; count > 1 {
				panic(fmt.Sprintf("Multiple battles with same ID=%v", b.ID))
			} else {
				idsCount[b.ID] = count
			}
			if !b.IsWithStranger() && (b.Name == "" && b.Nick == "") {
				panic(fmt.Sprintf("battle has no Name or Nickname, ID=%v", b.ID))
			}
		}
	}
	if b, err := ffjson.Marshal(&battles); err != nil {
		panic(err)
	} else {
		u.BattlesJson = string(b)
		u.BattlesCount = len(battles)
	}
	return
}

var ErrNotEnoughTokens = errors.New("not enough tokens")

func (u User) RecordBid(rivalKey arena.BattleID, bid int, now time.Time) (battle Battle, battles []Battle, err error) {
	if bid <= 0 {
		panic("bid < 0")
	}
	if rivalKey == "" {
		panic("rivalKey is empty string")
	}
	if now.IsZero() {
		panic("now is zero")
	}
	battles = u.GetBattles()
	var (
		i int
	)

	// Let's find the battle if already exists
	for i, battle = range battles {
		if battle.ID == rivalKey {
			// Battle found and we need to change to new bid
			if battle.Bid == nil || bid != battle.Bid.Value {
				var prevBidValue int
				if battle.Bid != nil { // Existing battle has no bid yet
					prevBidValue = battle.Bid.Value
				}
				if u.Tokens+prevBidValue-bid < 0 {
					err = errors.WithMessage(ErrNotEnoughTokens, fmt.Sprintf("not able to increase bid, balance is %v tokens", u.Tokens))
					return
				}
				battle.Bid = &Bid{Value: bid, Time: now}
				u.Tokens += prevBidValue - bid
				battles[i] = battle
			}
			return
		}
	}

	// A bid for a new battle
	if newBalance := u.Tokens - bid; newBalance < 0 {
		err = errors.WithMessage(ErrNotEnoughTokens, fmt.Sprintf("balance is %v tokens, bid %v", u.Tokens, bid))
		return
	}
	u.Tokens -= bid
	battle = Battle{ID: rivalKey, Bid: &Bid{Value: bid, Time: now}}
	battles = append(battles, battle)

	return
}
