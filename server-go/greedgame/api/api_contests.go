package api

import (
	"github.com/strongo-games/greed-game/server-go/greedgame/api/dto"
	"github.com/strongo-games/greed-game/server-go/greedgame/dal"
	"github.com/strongo-games/greed-game/server-go/greedgame/facade"
	"github.com/strongo-games/greed-game/server-go/greedgame/models"
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/arena"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"net/http"
	"strconv"
	"time"
)

func tournamentsList(c context.Context, userID string, w http.ResponseWriter, _ *http.Request) {
	log.Debugf(c, "tournamentsList(userID=%v)", userID)
	tournaments, err := facade.User.GetUserTournaments(c, models.User{StringID: db.StringID{ID: userID}})
	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
	jsonToResponse(c, w, tournamentsToDto(tournaments))
}

func tournamentToDto(t arena.Tournament) dto.TournamentDto {
	return dto.TournamentDto{
		Name:             t.Name,
		Status:           t.Status,
		Created:          t.Created,
		Starts:           t.Starts,
		Ends:             t.Ends,
		ContestantsCount: t.ContestantsCount,
	}
}

func tournamentsToDto(tournaments []arena.Tournament) (tornamentsDto map[string]dto.TournamentDto) {
	tornamentsDto = make(map[string]dto.TournamentDto, len(tournaments))
	for _, t := range tournaments {
		tornamentsDto[t.ID] = tournamentToDto(t)
	}
	return
}

func tournamentsCreate(c context.Context, userID string, w http.ResponseWriter, r *http.Request) {
	var err error

	tournament := arena.Tournament{
		TournamentEntity: new(arena.TournamentEntity),
	}
	if err = r.ParseForm(); err != nil {
		ErrorAsJson(c, w, http.StatusBadRequest, err)
		return
	}
	if tournament.Name = r.Form.Get("name"); tournament.Name == "" {
		ErrorAsJson(c, w, http.StatusBadRequest, errors.New("Name is required"))
		return
	} else if len(tournament.Name) > 100 {
		ErrorAsJson(c, w, http.StatusBadRequest, errors.New("Valeu of parameer 'name' is too long, 100 chars max."))
		return
	}

	if tournament.Starts, err = time.Parse("2006-01-02", r.Form.Get("starts")); err != nil {
		ErrorAsJson(c, w, http.StatusBadRequest, errors.WithMessage(err, "parameter start is invalid"))
		return
	}

	var duration int

	if duration, err = strconv.Atoi(r.Form.Get("duration")); err != nil || duration == 0 {
		ErrorAsJson(c, w, http.StatusBadRequest, errors.WithMessage(err, "parameter period is invalid"))
		return
	} else {
		tournament.Ends = tournament.Starts.Add(time.Duration(duration) * 24 * time.Hour)
	}

	if tournament.Note = r.Form.Get("note"); len(tournament.Note) > 1000 {
		ErrorAsJson(c, w, http.StatusBadRequest, errors.WithMessage(err, "parameter note is too long"))
		return
	}

	tournament.CreatorUserID = userID
	tournament.Created = time.Now()

	if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		var user models.User
		if user, err = dal.User.GetUserByID(c, userID); err != nil {
			return
		}
		if err = dal.DB.InsertWithRandomStrID(c, &tournament, 5, 3); err != nil {
			ErrorAsJson(c, w, http.StatusInternalServerError, err)
			return
		}
		user.TournamentIDs = append(user.TournamentIDs, tournament.ID)
		if err = dal.DB.Update(c, &user); err != nil {
			return
		}
		return
	}, db.CrossGroupTransaction); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}

	if tournament.ID == "" {
		ErrorAsJson(c, w, http.StatusInternalServerError, errors.New("no tournament ID"))
		return
	}

	tournamentDto := tournamentToDto(tournament)
	tournamentDto.ID = tournament.ID
	jsonToResponse(c, w, tournamentDto)
}

var errNotAuthorized = errors.New("not authorized")

func tournamentsArchive(c context.Context, userID string, w http.ResponseWriter, r *http.Request) {
	if userID == "" {
		ErrorAsJson(c, w, http.StatusForbidden, errors.New("userID is required"))
		return
	}
	if err := r.ParseForm(); err != nil {
		ErrorAsJson(c, w, http.StatusBadRequest, err)
		return
	}
	tournament := arena.Tournament{StringID: db.StringID{ID: r.Form.Get("id")}}
	if tournament.ID == "" {
		ErrorAsJson(c, w, http.StatusBadRequest, errors.New("Parameter 'id' is required."))
		return
	}

	err := dal.DB.RunInTransaction(c, func(c context.Context) (err error) {

		user := models.User{StringID: db.StringID{ID: userID}}

		if err = dal.DB.GetMulti(c, []db.EntityHolder{&tournament, &user}); err != nil {
			return
		}
		if tournament.CreatorUserID != userID {
			err = errNotAuthorized
			return
		}

		changedEntities := make([]db.EntityHolder, 0, 2)

		for i, id := range user.TournamentIDs {
			if id == tournament.ID {
				user.TournamentIDs = append(user.TournamentIDs[:i], user.TournamentIDs[i+1:]...)
				changedEntities = append(changedEntities, &user)
				break
			}
		}

		if tournament.Status != "archived" {
			tournament.Status = "archived"
			changedEntities = append(changedEntities, &tournament)
		}

		if len(changedEntities) > 0 {
			err = dal.DB.UpdateMulti(c, changedEntities)
		}

		return
	}, db.CrossGroupTransaction)

	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
	}
}
