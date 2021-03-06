package api

import (
	"context"

	"firebase.google.com/go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	//"google.golang.org/api/firestore/v1beta1"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/appengine/urlfetch"
	"net/http"
)

func newFirebaseApp(c context.Context) (firebaseApp *firebase.App, err error) {
	var config *jwt.Config
	config, err = google.JWTConfigFromJSON([]byte(`{
  "type": "service_account",
  "project_id": "greedgamealex",
  "private_key_id": "e8124567958051d82b4bfa5b3f951a991401690a",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC8abHOSxZXTlZ4\nmpVkD8avexSkfXTtCg/AO2udks2eLVbh1+DBPatRv6bdh2gvWTi+o9HmeNk6P7kA\nvsbt/MBUqbD9CfkaVceIG3aaK7nQxpJ3yOrDRqaNE+ohvVjFpP64UKtBH6gZAfI8\ntsiCqz1+cVwnDURyAd1AWV2VbclYZzYLSMB/jvJ+px94Ahq/m5xvov+0DZ7txz4c\nexDB1jBb/IxpoS1+lKkXvQoTi5cidUfj/OpOlTLdFU1y5KjOQ9XsxDWw3n5SY2fR\n+8u+JcOj6w9Y7RMV09KXeNLV2/bvdp0k5GJnOZk23ogSmX1bhxaCNXNniky4q8+D\n/DgCbkVTAgMBAAECggEAErFRyKu+ba8B+Tks9R5zkdleNOuVfCbxZRsAFEQKTlUl\nN4bZb5KUuqmO/o9+kKQDczaBjqISuyqzShWjWt0mn7+uJYylwC0efKxs2eLYrpPk\n2CmA0RrjTz/YjLxiYEl8VAD83JstbD27MLbZsc0XbsIEaINydPUmZEn5dOfNgA2h\nmbyd9c33aPFSSzok+XY0LGKw7rvgm77c8Y/m+m8hn4mfSfmt1vsFXV7hPZ7vZU40\n4cmey6169cqFZjgalel7eSi2KLMcM2581tm9VpHBNfvQQgsh+wDO0Kttrly3P7iA\n82Ru3W5vgNlVdJ4TbbR2ARnIUekAxE7NQMvZamry0QKBgQDvz6u33ISP4p88CKCZ\ng15UgL9njDRFjAWH84aIGEPWyHoUOMCSR5uPbH5heb/zlHAxHwU5fs9xIkuu2kcW\nWhy6xuKAWZf/hymPAq9vVmiS8mbsRd2T7LdpaddP78pEU5v+BqrWWQW8sh+HQ4Af\no5NhTZsF226sq/4Ks1IkflzShwKBgQDJIcbH0CoSVSGrCDUp2AeDNHcweJwIXjvl\njDgnjnqlOUwugF8oJG8gHFAdrMQeOuF9I1zvDDzfcfgIPig12VZgk2igGlsDJJMN\n6PJE7VOF5+N82XF0Xf8f90mKEp+IYmE/ww/1qGmvqXywGObPWzHZXd+/D3sx6ZNT\n6RxTrrvN1QKBgCjHH2P8U25EEt+ad/Siqf+khOeOp7TLwoUDm/S4a5CyNlAJ9nTp\nSEJzKGpa0ZERxKIVrEXCknOiaUwqQbxDRm9cMlew5G/HBAIVas972fxiy62Rk8P7\nlJSQMtSc6cAEl5nyeEpKiPc1MrdFexvmLMF2+M1eKsuh02juZSFfe1kxAoGAa6ze\naygg7dGPha2OMImLdA1JZbSb68rvC/OmOF8Jf5yOETL+PlJK/4jIxyovj/N7te+R\nmBQYHpM38sm74yAoIumnkFartKIG6+JymL3pAf3jhnouR9rucyGCyB0yNOReJbF6\nwMvZUIZOz0N1hTrQFAsydmmGTXE7Qye/13jq58UCgYEA5pMvibWYb8fw5Tq7YQtN\nqakkBn3482nmg2m6Nn99vDjW9kirgzQjKhwtIUoVDmfLpi9tKNJE/qNxcG5S7Elp\nODs4Rbi2/zyooD9AHvXKkb1GS0hf9uSFYV9l0OhP8s/oHxXp8cIoQOGbY5oblhA8\n+QiqgnfwKEHH2+jNrNOqnJY=\n-----END PRIVATE KEY-----\n",
  "client_email": "firebase-adminsdk-v5t2w@greedgamealex.iam.gserviceaccount.com",
  "client_id": "105988290052976225701",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://accounts.google.com/o/oauth2/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/firebase-adminsdk-v5t2w%40greedgamealex.iam.gserviceaccount.com"
}`), "https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/datastore", "https://www.googleapis.com/auth/firebase")
	if err != nil {
		return
	}
	transport := &oauth2.Transport{
		Source: config.TokenSource(c),
		Base:   &urlfetch.Transport{Context: c},
	}
	opt := option.WithHTTPClient(&http.Client{Transport: transport})

	return firebase.NewApp(c, nil, opt)
}
