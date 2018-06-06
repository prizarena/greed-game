package api

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"google.golang.org/appengine"
	"net/http"
	//"github.com/SermoDigital/jose/crypto"
	"fmt"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/log"
	"strings"
)

func InitApi(router *httprouter.Router) {
	POST := func(path string, handle httprouter.Handle) {
		router.POST(path, handle)
		router.OPTIONS(path, handlerWithContext(optionsHandler))
	}
	GET := func(path string, handle httprouter.Handle) {
		router.GET(path, handle)
		router.OPTIONS(path, handlerWithContext(optionsHandler))
	}

	GET("/api/user/state", handlerWithAuthentication(userFullState))
	GET("/api/tournaments/list", handlerWithAuthentication(tournamentsList))
	POST("/api/tournaments/create", handlerWithAuthentication(tournamentsCreate))
	POST("/api/tournaments/archive", handlerWithAuthentication(tournamentsArchive))
	POST("/api/play/place-bid", handlerWithAuthentication(playPlaceBid))
	POST("/api/play/withdraw-bid", handlerWithAuthentication(playWithdrawBid))
	POST("/api/play/quit-battle", handlerWithAuthentication(playQuitBattle))
	POST("/api/play/new-game", handlerWithAuthentication(playNewGame))
	GET("/api/play/battle-state", handlerWithAuthentication(playBattleState))
	POST("/api/auth/issue-token-to-firebase-user", handlerWithContext(issueTokenToFirebaseUser))
}

func handlerWithContext(f func(c context.Context, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		f(appengine.NewContext(r), w, r)
	}
}

func handlerWithAuthentication(f func(c context.Context, userID string, w http.ResponseWriter, r *http.Request)) httprouter.Handle {
	return handlerWithContext(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		log.Debugf(c, "handlerWithAuthentication()")
		if userID, err := authenticate(w, r, true); err != nil {
			log.Errorf(c, "handlerWithAuthentication() => %v", err)
			//w.WriteHeader(http.StatusForbidden)
			//w.Write([]byte(err.Error()))
		} else {
			f(c, userID, w, r)
		}
	})
}

func BadRequestMessage(c context.Context, w http.ResponseWriter, m string) {
	log.Infof(c, m)
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(m))
}

func BadRequestError(c context.Context, w http.ResponseWriter, err error) {
	BadRequestMessage(c, w, err.Error())
}

func markResponseAsJson(header http.Header) {
	header.Add("Content-Type", "application/json")
	header.Add("Access-Control-Allow-Origin", "*")
}

func jsonToResponse(c context.Context, w http.ResponseWriter, v interface{}) {
	header := w.Header()
	if buffer, err := ffjson.Marshal(v); err != nil {
		log.Errorf(c, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		header.Add("Access-Control-Allow-Origin", "*")
		log.Debugf(c, "w.Header(): %v", header)
		w.Write([]byte(err.Error()))
	} else {
		markResponseAsJson(header)
		log.Debugf(c, "w.Header(): %v", header)
		_, err := w.Write(buffer)
		ffjson.Pool(buffer)
		if err != nil {
			InternalError(c, w, err)
		}
	}
}

func InternalError(c context.Context, w http.ResponseWriter, err error) {
	m := err.Error()
	log.Errorf(c, m)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(m))
}

func ErrorAsJson(c context.Context, w http.ResponseWriter, status int, err error) {
	if status == 0 {
		panic("status == 0")
	}
	if status == http.StatusInternalServerError {
		log.Errorf(c, "Error: %v", err.Error())
	} else {
		log.Infof(c, "Error: %v", err.Error())
	}
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	jsonToResponse(c, w, map[string]string{"error": err.Error()})
}

func optionsHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
	log.Debugf(c, "optionsHandler()")
	if r.Method != "OPTIONS" {
		panic("Method != OPTIONS")
	}
	// Pre-flight request
	origin := r.Header.Get("Origin")
	switch origin {
	case "http://localhost:8080":
	case "http://localhost:8100":
	case "https://greed-game.com":
	case "":
		BadRequestMessage(c, w, "Missing required request header: Origin")
		return
	default:
		if !(strings.HasPrefix(origin, "http://") && strings.HasSuffix(origin, ":8100")) {
			err := fmt.Errorf("unknown origin: %v", origin)
			log.Debugf(c, err.Error())
			BadRequestError(c, w, err)
			return
		}
	}
	log.Debugf(c, "Request 'Origin' header: %v", origin)
	responseHeader := w.Header()
	if accessControlRequestMethod := r.Header.Get("Access-Control-Request-Method"); !(accessControlRequestMethod == "GET" || accessControlRequestMethod == "POST") {
		BadRequestMessage(c, w, "Requested method is unsupported: "+accessControlRequestMethod)
		return
	} else {
		responseHeader.Set("Access-Control-Allow-Methods", accessControlRequestMethod)
	}
	if accessControlRequestHeaders := r.Header.Get("Access-Control-Request-Headers"); accessControlRequestHeaders != "" {
		log.Debugf(c, "Request Access-Control-Request-Headers: %v", accessControlRequestHeaders)
		responseHeader.Set("Access-Control-Allow-Headers", accessControlRequestHeaders)
	} else {
		log.Debugf(c, "Request header 'Access-Control-Allow-Headers' is empty or missing")
		// TODO(security): Is it wrong to return 200 in this case?
	}
	responseHeader.Set("Access-Control-Max-Age", "600")
	responseHeader.Set("Cache-Control", "public, max-age=600")
	responseHeader.Set("Access-Control-Allow-Origin", origin)
}
