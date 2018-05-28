package models

import (
	"github.com/strongo/db"
	"time"
)

const (
	UserFirebaseKind = "UserFirebase"
)

type UserFirebaseEntity struct {
	UserID        string `datastore:",omitempty"`
	Created       time.Time
	DisplayName   string `datastore:",noindex,omitempty"`
	Email         string `datastore:",omitempty"`
	EmailVerified bool   `datastore:",noindex,omitempty"`
	PhotoURL      string `datastore:",noindex,omitempty"`
	PhoneNumber   string `datastore:",noindex,omitempty"`
	ProviderID    string `datastore:",noindex,omitempty"`
	FcmToken      string `datastore:",noindex,omitempty"`
}

type UserFirebase struct {
	db.StringID
	*UserFirebaseEntity
}

var _ db.EntityHolder = (*UserFirebase)(nil)

func (UserFirebase) Kind() string {
	return UserFirebaseKind
}

func (UserFirebase) NewEntity() interface{} {
	return new(UserFirebaseEntity)
}

func (u *UserFirebase) Entity() interface{} {
	return u.UserFirebaseEntity
}

func (u *UserFirebase) SetEntity(v interface{}) {
	if v == nil {
		u.UserFirebaseEntity = nil
	} else {
		u.UserFirebaseEntity = v.(*UserFirebaseEntity)
	}

}
