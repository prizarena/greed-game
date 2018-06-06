package models

import (
	"fmt"
	"github.com/pkg/errors"
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
	BattlesHandler
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
