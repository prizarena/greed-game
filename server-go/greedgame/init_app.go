package greedgame

import (
	"github.com/strongo-games/greed-game/server-go/greedgame/api"
	"github.com/strongo-games/greed-game/server-go/greedgame/dal/gaedal"
	"github.com/julienschmidt/httprouter"
	"github.com/strongo/bots-framework/core"
	"net/http"
)

func InitGreedGameApp(botHost bots.BotHost) {
	gaedal.RegisterDal()
	router := httprouter.New()
	http.Handle("/", router)
	router.GET("/", homepage)
	api.InitApi(router)
	initBot(router, botHost, appContext{})
}

func homepage(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	w.Write([]byte(`
<!doctype html>
<html lang="en">
  <head>
    <title>The Greed Game</title>
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-beta.2/css/bootstrap.min.css" integrity="sha384-PsH8R72JQ3SOdhVi3uxftmaW6Vc51MKb0q5P2rRUpPvrszuE4W1povHYgTpBfshb" crossorigin="anonymous">
  </head>
  <body>

<div class="container">
    <h1>The Greed Game</h1>

<h2>Rules</h2>
<p>
Two players make hidden bids against each other.
</p>
<p>
You get back 2x of your bet if other opponent was too greedy and is bidding more then you.
</p>
<p>
Unless he was bold enough to bet more than twice than you. In this case he gets your bet.
</p>

<button onclick=signInWithGoogle()>Sign in with Google</button>
</div>

    <!-- Optional JavaScript -->
    <!-- jQuery first, then Popper.js, then Bootstrap JS -->
<!--
    <script src="https://code.jquery.com/jquery-3.2.1.slim.min.js" integrity="sha384-KJ3o2DKtIkvYIK3UENzmM7KCkRr/rE9/Qpg6aAZGJwFDMVNA/GpGFF93hXpG5KkN" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.12.3/umd/popper.min.js" integrity="sha384-vFJXuSJphROIrBnz7yo7oB41mKfc8JzQZiCq4NCceLEaO4IHwicKwpJf9c9IpFgh" crossorigin="anonymous"></script>
    <script src="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-beta.2/js/bootstrap.min.js" integrity="sha384-alpBpkh1PFOepccYVYDB4do5UnbKysX5WZXm3XxPqe5iKTfUKjNkCk9SaVuEZflJ" crossorigin="anonymous"></script>
-->
  </body>

<script src="https://www.gstatic.com/firebasejs/4.8.1/firebase-app.js"></script>
<script src="https://www.gstatic.com/firebasejs/4.8.1/firebase-auth.js"></script>
<script src="https://www.gstatic.com/firebasejs/4.8.1/firebase-firestore.js"></script>
<!--script src="https://www.gstatic.com/firebasejs/4.8.1/firebase-database.js"></script-->
<!--script src="https://www.gstatic.com/firebasejs/4.8.1/firebase-messaging.js"></script-->

<script>
  // Initialize Firebase
	// apiKey: "AIzaSyCYXFaNW9AgukMY-kX7-Q5TlrpS8iAXLMU",
	//     messagingSenderId: "206654700263"
  var config = {
    authDomain: "greedgamealex.firebaseapp.com",
    databaseURL: "https://greedgamealex.firebaseio.com",
    projectId: "greedgamealex",
  };
  firebase.initializeApp(config);
</script>
<script>
function signInWithGoogle() {
	var provider = new firebase.auth.GoogleAuthProvider();

	firebase.auth().signInWithPopup(provider).then(function(result) {
	  // This gives you a Google Access Token. You can use it to access the Google API.
	  //var token = result.credential.accessToken;
	  // The signed-in user info.
		console.log('user', result.user);
	  //var user = result.user;
	  // ...
	}).catch(function(error) {
	  // Handle Errors here.
	  console.log('error', error);
	  //var errorCode = error.code;
	  //var errorMessage = error.message;
	  // The email of the user's account used.
	  //var email = error.email;
	  // The firebase.auth.AuthCredential type that was used.
	  //var credential = error.credential;
	  // ...
	});
}
</script>
</html>
`))
}
