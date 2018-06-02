package greedgame

import (
	"reflect"
	"time"

	"github.com/prizarena/greed-game/server-go/greedgame/models"
	"context"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
)

type appContext struct {
}

var _ bots.BotAppContext = (*appContext)(nil)

func (appCtx appContext) AppUserEntityKind() string {
	return models.UserKind
}

func (appCtx appContext) AppUserEntityType() reflect.Type {
	return reflect.TypeOf(&models.UserEntity{})
}

func (appCtx appContext) NewBotAppUserEntity() bots.BotAppUser {
	return &models.UserEntity{
		Created: time.Now(),
	}
}

func (appCtx appContext) GetBotChatEntityFactory(platform string) func() bots.BotChat {
	switch platform {
	case "telegram":
		return func() bots.BotChat {
			return &models.TelegramChatEntity{
				TgChatEntityBase: *telegram.NewTelegramChatEntity(),
			}
		}
	default:
		panic("Unknown platform: " + platform)
	}
}

func (appCtx appContext) NewAppUserEntity() strongo.AppUser {
	return appCtx.NewBotAppUserEntity()
}

func (appCtx appContext) GetTranslator(c context.Context) strongo.Translator {
	return strongo.NewMapTranslator(c, trans.TRANS)
}

func (appCtx appContext) SupportedLocales() strongo.LocalesProvider {
	return trans.DebtsTrackerLocales{}
}

var _ bots.BotAppContext = (*appContext)(nil)
