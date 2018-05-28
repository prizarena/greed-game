package api

import (
	"github.com/strongo-games/greed-game/server-go/greedgame/api/dto"
	"github.com/strongo-games/greed-game/server-go/greedgame/dal"
	"github.com/strongo-games/greed-game/server-go/greedgame/models"
	"context"
	firebaseAuth "firebase.google.com/go/auth"
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/SermoDigital/jose/jwt"
	"github.com/pkg/errors"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"google.golang.org/appengine"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

//const SECRET_PREFIX = "eySOMESECRETzzI1NiIsInR5cCI6IkpXVCJ9."

var ErrNoToken = errors.New("No authorization token")

func authenticate(w http.ResponseWriter, r *http.Request, required bool) (userID string, err error) {
	c := appengine.NewContext(r)
	defer func() { // Logs error
		log.Debugf(c, "authenticate() => %v", err)
		if err != nil && required {
			errorText := err.Error()
			log.Debugf(c, errorText)
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("authenticate error:" + errorText))
		}
	}()

	tokenString := r.URL.Query().Get("secret")
	if tokenString == "" {
		if a := r.Header.Get("Authorization"); strings.HasPrefix(a, "Bearer ") {
			tokenString = a[7:]
		}
	}
	if tokenString == "" {
		err = ErrNoToken
		return
	}
	log.Debugf(appengine.NewContext(r), "JWT token: [%v]", tokenString)

	var token jwt.JWT
	if token, err = jws.ParseJWT([]byte(tokenString)); err != nil {
		log.Debugf(c, "Tried to parse: [%v]", tokenString)
		return
	}
	if err = token.Validate(secret, jwtSigningMethod); err != nil {
		return
	}

	claims := token.Claims()
	if sub, ok := claims.Subject(); ok {
		userID = sub
	} else {
		err = errors.New("JWT is missing 'sub' claim.")
		return
	}
	return
}

func issueTokenToFirebaseUser(c context.Context, w http.ResponseWriter, r *http.Request) {
	log.Debugf(c, "issueTokenToFirebaseUser()")
	failed := func(err error, status int) {
		log.Debugf(c, "HTTP StatusForbidden: %v", err)
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
	}
	if idToken, err := ioutil.ReadAll(r.Body); err != nil {
		failed(err, http.StatusInternalServerError)
		return
	} else if firebaseApp, err := newFirebaseApp(c); err != nil {
		failed(err, http.StatusInternalServerError)
		return
	} else if fbAuth, err := firebaseApp.Auth(c); err != nil {
		failed(err, http.StatusInternalServerError)
		return
	} else if firebaseToken, err := fbAuth.VerifyIDToken(string(idToken)); err != nil {
		failed(err, http.StatusForbidden)
		return
	} else if firebaseUser, err := fbAuth.GetUser(c, firebaseToken.UID); err != nil {
		log.Errorf(c, "failed to get Firebase user: ", err)
		failed(err, http.StatusInternalServerError)
		return
	} else {
		log.Debugf(c, "Firebase user: %+v", firebaseUser)
		var user models.User
		userFirebase := new(models.UserFirebase)
		userFirebase.ID = firebaseToken.UID
		now := time.Now()
		err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
			if err = dal.DB.Get(c, userFirebase); err != nil {
				if db.IsNotFound(err) {
					userFirebase.UserFirebaseEntity = &models.UserFirebaseEntity{
						UserID:        firebaseToken.UID,
						Created:       now,
						DisplayName:   firebaseUser.DisplayName,
						Email:         firebaseUser.Email,
						EmailVerified: firebaseUser.EmailVerified,
						PhoneNumber:   firebaseUser.PhoneNumber,
						PhotoURL:      firebaseUser.PhotoURL,
					}
					user = models.User{
						StringID: db.StringID{ID: firebaseToken.UID},
						UserEntity: &models.UserEntity{
							Created:     now,
							Tokens:      1000,
							FirebaseUID: firebaseToken.UID,
							Name:        firebaseUser.DisplayName,
							AvatarURL:   firebaseUser.PhotoURL,
						},
					}
					if err = dal.DB.UpdateMulti(c, []db.EntityHolder{&user, userFirebase}); err != nil {
						return
					}
				}
				return
			} else {
				if err = updateUserFirebase(c, userFirebase, firebaseUser); err != nil {
					failed(err, http.StatusInternalServerError)
				}

				// Update existing user if needed
				if user, err = dal.User.GetUserByID(c, userFirebase.UserID); err != nil {
					return
				}

				if err = updateUserWithFirebaseUser(c, user, firebaseUser); err != nil {
					return
				}
			}
			return
		}, db.CrossGroupTransaction)

		if err != nil {
			failed(err, http.StatusInternalServerError)
			return // Response already formed
		}

		log.Debugf(c, "Firebase ID token verified successfully")
		token := issueToken(userFirebase.UserID, firebaseUser.UID, "Firebase")

		authResponse := dto.AuthResponse{
			Token: string(token),
			User: dto.UserBriefState{
				Balance: user.Tokens,
			},
		}
		jsonToResponse(c, w, authResponse)
	}
}

func updateUserFirebase(c context.Context, userFirebase *models.UserFirebase, firebaseUser *firebaseAuth.UserRecord) (err error) {
	userFirebaseChanged := false

	if firebaseUser.Email != "" && (userFirebase.Email == "" || (userFirebase.Email != userFirebase.Email && !firebaseUser.EmailVerified)) {
		userFirebase.Email = userFirebase.Email
		userFirebase.EmailVerified = userFirebase.EmailVerified
		userFirebaseChanged = true
	}

	if userFirebase.Email == userFirebase.Email && !userFirebase.EmailVerified && firebaseUser.EmailVerified {
		userFirebase.EmailVerified = userFirebase.EmailVerified
		userFirebaseChanged = true
	}

	if userFirebase.PhoneNumber != firebaseUser.PhoneNumber {
		userFirebase.PhoneNumber = firebaseUser.PhoneNumber
		userFirebaseChanged = true
	}

	if userFirebase.PhotoURL != firebaseUser.PhotoURL {
		userFirebase.PhotoURL = firebaseUser.PhotoURL
		userFirebaseChanged = true
	}

	if userFirebase.ProviderID != firebaseUser.ProviderID {
		userFirebase.ProviderID = firebaseUser.ProviderID
		userFirebaseChanged = true
	}

	if userFirebaseChanged {
		if err = dal.DB.Update(c, userFirebase); err != nil {
			return
		}
	}
	return
}

func updateUserWithFirebaseUser(c context.Context, user models.User, firebaseUser *firebaseAuth.UserRecord) (err error) {
	userChanged := false

	if user.FirebaseUID != firebaseUser.UID {
		user.FirebaseUID = firebaseUser.UID
		userChanged = true
	}

	if user.Tokens == 0 {
		user.Tokens = 100
		userChanged = true
	}

	if user.Name == user.ID && firebaseUser.DisplayName != user.Name {
		user.Name = firebaseUser.DisplayName
		userChanged = true
	}

	if userChanged {
		if err = dal.DB.Update(c, &user); err != nil {
			return
		}
	}
	return
}

var (
	secret           = []byte("very-secret-abc")
	jwtSigningMethod = crypto.SigningMethodHS256
)

func issueToken(userID, firebaseUID, issuer string) []byte {
	if userID == "" {
		panic("IssueToken(userID is empty string)")
	}
	claims := jws.Claims{}
	claims.SetIssuedAt(time.Now())
	claims.SetSubject(userID)
	claims.Set("FirebaseUID", firebaseUID)

	if issuer != "" {
		claims.SetIssuer(issuer)
	}

	token := jws.NewJWT(claims, jwtSigningMethod)
	signature, err := token.Serialize(secret)
	if err != nil {
		panic(err.Error())
	}

	return signature
}
