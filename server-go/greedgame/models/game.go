package models

import (
	"fmt"
	"github.com/strongo/db"
	"strconv"
	"github.com/prizarena/turn-based"
)

type PlayEntity struct {
	turnbased.PlayEntity
	Bids []int `datastore:",noindex"`
}

const PlayKind = "Play"

type Play struct {
	db.StringID
	*PlayEntity
}

var _ db.EntityHolder = (*Play)(nil)

func (Play) Kind() string {
	return PlayKind
}

func (Play) NewEntity() interface{} {
	return new(PlayEntity)
}

func (g Play) Entity() interface{} {
	return g.PlayEntity
}

func (g *Play) SetEntity(v interface{}) {
	if v == nil {
		g.PlayEntity = nil
	} else {
		g.PlayEntity = v.(*PlayEntity)
	}
}

func (g *PlayEntity) GetBid(userID string) (bid int) {
	switch len(g.Bids) {
	case 0:
		return
	case 1:
		if g.UserIDs[0] == userID {
			return g.Bids[0]
		}
		panic("unknown user ID")
	case 2:
		switch userID {
		case g.UserIDs[0]:
			return g.Bids[0]
		case g.UserIDs[1]:
			return g.Bids[1]
		default:
			panic("unknown user ID")
		}
	default:
		panic("too many user IDs")
	}
}

func (g *PlayEntity) SetBid(userID string, bid int) (change int) {
	if bid <= 0 {
		panic(fmt.Sprintf("bid should be > 0, got %v", bid))
	}
	if len(g.Bids) == 0 {
		g.Bids = []int{0, 0}
	}
	switch userID {
	case g.UserIDs[0]:
		if g.Bids[0] != int(bid) {
			change = bid - g.Bids[0]
			g.Bids[0] = bid
		}
	case g.UserIDs[1]:
		if g.Bids[1] != int(bid) {
			change = bid - g.Bids[1]
			g.Bids[1] = bid
		}
	default:
		panic("unknown user ID: " + userID)
	}
	return
}

func (g PlayEntity) HasBothBids() bool {
	if bidsCount := len(g.Bids); bidsCount > 2 {
		panic("maximum 2 bids allowed, got " + strconv.FormatInt(int64(bidsCount), 10))
	} else {
		return bidsCount == 2 && g.Bids[0] > 0 && g.Bids[1] > 0
	}
}

func (g PlayEntity) Prize() int {
	if !g.HasBothBids() {
		panic("not enough bids")
	}
	if bid0, bid1 := g.Bids[0], g.Bids[1]; bid0 < bid1 {
		return bid0
	} else {
		return bid1
	}
}
