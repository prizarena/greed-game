package models

import (
	"github.com/pkg/errors"
	"github.com/strongo/db"
	"strings"
	"time"
)

type GamesSessionEntity struct {
	Created          time.Time `datastore:",noindex,omitempty"`
	LastGameID       string    `datastore:"LgId,noindex,omitempty"`
	LastGameFinished time.Time `datastore:"LgT,noindex,omitempty"`
	GamesPlayed      int       `datastore:",noindex,omitempty"`
	Balance          int       `datastore:",noindex,omitempty"` // Balance for 1st user
}

const GamesSessionKind = "uu"

const gamesSessionIdDelimiter = "-"

func GamesSessionIdFromUserIDs(user1, user2 string) string {
	switch strings.Compare(user1, user2) {
	case -1:
		return user1 + gamesSessionIdDelimiter + user2
	case 1:
		return user2 + gamesSessionIdDelimiter + user1
	default:
		panic("game with self is not allowed: " + user1)
	}
}

func UserIDsFromGameSessionID(id string) (user1, user2 string, err error) {
	userIDs := strings.Split(id, gamesSessionIdDelimiter)
	if len(userIDs) != 2 {
		err = errors.Errorf("games session ID should consist of 2 user IDs delimited '%v' by character", gamesSessionIdDelimiter)
		return
	}
	user1, user2 = userIDs[0], userIDs[1]
	switch strings.Compare(user1, user2) {
	case 1:
		err = errors.New("wrong order of user IDs")
	case 0:
		err = errors.New("same user ID")
	}
	return
}

type GamesSession struct {
	db.StringID
	*GamesSessionEntity
}

var _ db.EntityHolder = (*GamesSession)(nil)

func (GamesSession) Kind() string {
	return GamesSessionKind
}

func (s GamesSession) Entity() interface{} {
	return s.GamesSessionEntity
}

func (s GamesSession) SetEntity(v interface{}) {
	s.GamesSessionEntity = v.(*GamesSessionEntity)
}

func (GamesSession) NewEntity() interface{} {
	return new(GamesSessionEntity)
}

func (s GamesSession) UserBalance(userID string) int {
	u1, u2, err := UserIDsFromGameSessionID(s.ID)
	if err != nil {
		panic(err)
	}
	switch userID {
	case u1:
		return s.Balance
	case u2:
		return -s.Balance
	default:
		panic("unknown user ID: " + userID)
	}
}
