package dto

//go:generate ffjson $GOFILE

import (
	"github.com/prizarena/greed-game/server-go/greedgame/models"
	"encoding/json"
	"time"
)

type SponsorDto struct {
	Name          string
	Text          string `json:",omitempty"`
	PrizeValue    int
	PrizeCurrency string
}

type TournamentDto struct {
	ID               string `json:",omitempty"`
	Name             string
	Status           string
	Created          time.Time
	Starts           time.Time
	Ends             time.Time
	ContestantsCount int         `json:",omitempty"`
	Sponsor          *SponsorDto `json:",omitempty"`
}

type RivalDto struct {
	Bid     int    `json:",omitempty"`
	ID      string `json:",omitempty"`
	Name    string `json:",omitempty"`
	Nick    string `json:",omitempty"`
	Balance int    `json:",omitempty"` // Balance of games between user and the rival, not the rival's balance
}

type GameDto struct {
	ID      string    `json:",omitempty"`
	Status  string    `json:",omitempty"`
	UserBid int       `json:",omitempty"`
	Prize   int       `json:",omitempty"`
	Rival   *RivalDto `json:",omitempty"`
}

type ErrorDto struct {
	Code    string `json:",omitempty"`
	Message string `json:",omitempty"`
}

type BidResponse struct {
	UserBalance int            `json:",omitempty"`
	Battle      *models.Battle `json:",omitempty"`
	Error       *ErrorDto      `json:",omitempty"`
}

type UserBriefState struct {
	Balance int
}

type UserFullState struct {
	UserBriefState
	Battles     json.RawMessage          `json:",omitempty"`
	Tournaments map[string]TournamentDto `json:",omitempty"`
}

type AuthResponse struct {
	Token string
	User  UserBriefState
}
