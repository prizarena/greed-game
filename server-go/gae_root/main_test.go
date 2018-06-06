package appengine

import (
	"testing"
	"github.com/strongo/log"
	"github.com/prizarena/greed-game/server-go/greedgame/dal"
)

func TestMain1(t *testing.T) {
	if dal.DB == nil {
		t.Error("btttdal.DB == nil")
	}
	if log.NumberOfLoggers() == 0 {
		t.Error("NumberOfLoggers() == 0")
	}
}
