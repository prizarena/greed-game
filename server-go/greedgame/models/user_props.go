package models

//go:generate ffjson $GOFILE

import (
	"github.com/pquerna/ffjson/ffjson"
	"github.com/prizarena/arena/arena-go"
	"time"
	"fmt"
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

// Holds information about bids made by a user against other user in some tournament
// BattleID has form of userId@tournamentID
// User or some other dedicated entity will hold list of battles for a specific user
type Battle struct {
	ID       arena.BattleID
	OldID    arena.BattleID `json:",omitempty"`
	Name     string         `json:",omitempty"`
	Nick     string         `json:",omitempty"`
	Bid      *Bid           `json:",omitempty"`
	Move     string         `json:",omitempty"` // For GreedGame it's empty. For BiddingTTT it's cell, for RPS its either rock|paper|scissors
	LastGame *LastGame      `json:",omitempty"`
}

type Battles []Battle

type BattlesHandler struct {
	BattlesJson   string   `datastore:",noindex,omitempty"`
	BattlesCount  int      `datastore:",noindex,omitempty"`
}

func (entity *BattlesHandler) GetBattles() (battles Battles) {
	if entity.BattlesJson == "" {
		battles = make(Battles, 0, 1)
		return
	}
	battles = make(Battles, 0, entity.BattlesCount)
	if err := ffjson.Unmarshal([]byte(entity.BattlesJson), &battles); err != nil {
		panic(err)
	}
	return
}

func (entity *BattlesHandler) SetBattles(battles []Battle) {
	if len(battles) == 0 {
		entity.BattlesJson = ""
		entity.BattlesCount = 0
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
		entity.BattlesJson = string(b)
		entity.BattlesCount = len(battles)
	}
	return
}


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

func (j *Battle) IsWithStranger() bool {
	return j.ID.IsStranger()
}

// type UserGames struct {
// 	RivalStats map[BattleID]RivalStat // Map bids by rival user ID
// }
//
// func (v UserGames) String() string {
// 	b, err := ffjson.MarshalFast(&v)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return string(b)
// }
