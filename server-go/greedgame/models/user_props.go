package models

//go:generate ffjson $GOFILE

import (
	"github.com/pquerna/ffjson/ffjson"
	"github.com/prizarena/arena/arena-go"
	"time"
)

type LastGame struct {
	ID       string
	Status   string
	UserBid  int
	RivalBid int
	Prize    int
	Time     time.Time `json:"t"`
}

type Bid struct {
	Time  time.Time
	Value int
}

func (j *Bid) String() string {
	if b, err := ffjson.MarshalFast(j); err != nil {
		panic(err)
	} else {
		return string(b)
	}
}

type Battle struct {
	ID       arena.BattleID
	OldID    arena.BattleID `json:",omitempty"`
	Name     string         `json:",omitempty"`
	Nick     string         `json:",omitempty"`
	Bid      *Bid           `json:",omitempty"`
	LastGame *LastGame      `json:",omitempty"`
}

type Battles []Battle

func (battles Battles) GetBattleByRivalID(k arena.BattleID) *Battle {
	for _, row := range battles {
		if row.ID == k || row.OldID == k {
			return &row
		}
	}
	return nil
}

func (j *Battle) String() string {
	if b, err := ffjson.MarshalFast(j); err != nil {
		panic(err)
	} else {
		return string(b)
	}
}

func (b *Battle) IsWithStranger() bool {
	return b.ID.IsStranger()
}

//type UserGames struct {
//	RivalStats map[BattleID]RivalStat // Map bids by rival user ID
//}
//
//func (v UserGames) String() string {
//	b, err := ffjson.MarshalFast(&v)
//	if err != nil {
//		panic(err)
//	}
//	return string(b)
//}
