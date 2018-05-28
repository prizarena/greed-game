package models

import (
	"testing"
	"time"
)

func TestUser_RecordBid(t *testing.T) {
	u := User{
		UserEntity: &UserEntity{
			Tokens: 1000,
		},
	}

	verify := func(err error, now time.Time, expectedBalance, expectedBid int, battle *Battle, battles []Battle) {
		t.Helper()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if u.Tokens != expectedBalance {
			t.Errorf("Expected balance is %v, got: %v", expectedBalance, u.Tokens)
		}

		if battle == nil {
			t.Fatalf("battle == nil")
		}

		if battle.ID != "*@*" {
			t.Errorf("Unexpected battle.ID: %v", battle.ID)
		}

		if battle.Name != "" {
			t.Errorf("battle.Name should be empty string for a stranger")
		}

		if battle.Nick != "" {
			t.Errorf("battle.Nick should be empty string for a stranger")
		}

		if battle.Bid.Value != expectedBid {
			t.Errorf("battle.Bid.Value expected to be %v, got: %v", expectedBid, battle.Bid.Value)
		}

		if battle.Bid.Time != now {
			t.Errorf("battle.Bid.Time has value %v, expected: %v", battle.Bid.Time, now)
		}

		if len(battles) != 1 {
			t.Errorf("Unexpected length of battles: %v", len(battles))
		}

		if battle != battles[0] {
			t.Errorf("battle != battles[0]:\n\tbattle:%v\n\t%v", battle, battles[0])
		}
		u.SetBattles(battles)
	}

	now := time.Now()
	battle, battles, err := u.RecordBid("*@*", 10, now)
	verify(err, now, 990, 10, battle, battles)

	now = time.Now()
	battle, battles, err = u.RecordBid("*@*", 20, now)
	verify(err, now, 980, 20, battle, battles)

}
