package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/y3sh/go143/instagram"

	"github.com/go-chi/cors"
	"github.com/y3sh/go143/twitter"

	"github.com/go-chi/chi"
	"github.com/juju/errors"
)

const (
	SiteRoot            = "/"
	APIRoot             = "/api"
	TweetsURI           = "/api/v1/tweets"
	RandTweetURI        = "/api/v1/randTweet"
	InstagramUserURI    = "/api/v1/instagram/users"
	InstagramSessionURI = "/api/v1/instagram/sessions"
)

var (
	OK         = &struct{}{}
	apiVersion = &APIVersion{"GO143", "v1", []string{
		"https://pure-ridge-19371.herokuapp.com/api/v1/tweets",
		"https://pure-ridge-19371.herokuapp.com/api/v1/randTweet",
		"https://pure-ridge-19371.herokuapp.com/api/v1/instagram/user",
		"https://pure-ridge-19371.herokuapp.com/api/v1/instagram/session",
	}}
)

type API struct {
	Router               Router
	TweetService         TweetService
	InstagramUserService InstagramUserService
}

type Router interface {
	Use(middleware ...func(http.Handler) http.Handler)
	Route(pattern string, fn func(r chi.Router)) chi.Router
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type TweetService interface {
	GetTweets() []*twitter.Tweet
	AddTweet(tweetText string) (*twitter.Tweet, error)
	AddRandTweet() (*twitter.Tweet, error)
}

type InstagramUserService interface {
	AddUser(user instagram.User) error
	IsValidPassword(username instagram.Username, passwordAttempt string) bool
}

type APIVersion struct {
	API     string   `json:"api"`
	Version string   `json:"version"`
	URLS    []string `json:"urls"`
}

func NewAPIRouter(httpRouter Router, tweetService TweetService, instagramUserService InstagramUserService) *API {
	a := &API{
		Router:               httpRouter,
		TweetService:         tweetService,
		InstagramUserService: instagramUserService,
	}

	a.EnableCORS()

	httpRouter.Route(SiteRoot, func(r chi.Router) {
		r.Get("/", a.GetRoot)
	})

	httpRouter.Route(APIRoot, func(r chi.Router) {
		r.Get("/", a.GetRoot)
	})

	httpRouter.Route(TweetsURI, func(r chi.Router) {
		r.Get("/", a.GetTweets)
		r.Post("/", a.PostTweet)
	})

	httpRouter.Route(RandTweetURI, func(r chi.Router) {
		r.Get("/", a.GetRandTweet)
	})

	httpRouter.Route(InstagramUserURI, func(r chi.Router) {
		r.Post("/", a.PostInstagramUser)
	})

	httpRouter.Route(InstagramSessionURI, func(r chi.Router) {
		r.Post("/", a.PostInstagramSession)
	})

	http.Handle(SiteRoot, httpRouter)

	return a
}

func (a *API) PostInstagramUser(w http.ResponseWriter, r *http.Request) {
	var user instagram.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		WriteBadRequest(w, r, "Error invalid user format.")
		return
	}

	err = a.InstagramUserService.AddUser(user)
	if err != nil {
		WriteBadRequest(w, r, fmt.Sprintf("Error: %s.", err.Error()))
		return
	}

	WriteJSON(w, r, OK)
}

func (a *API) PostInstagramSession(w http.ResponseWriter, r *http.Request) {
	var user instagram.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		WriteBadRequest(w, r, "Error invalid user format.")
		return
	}

	validPassword := a.InstagramUserService.IsValidPassword(user.Username, user.Password)
	if !validPassword {
		WriteBadRequest(w, r, "Invalid username and/or password.")
		return
	}

	WriteJSON(w, r, OK)
}

func (a *API) GetRoot(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, r, apiVersion)
}

func (a *API) GetTweets(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, r, a.TweetService.GetTweets())
}

func (a *API) PostTweet(w http.ResponseWriter, r *http.Request) {
	var tweet twitter.Tweet

	err := json.NewDecoder(r.Body).Decode(&tweet)
	if err != nil {
		WriteBadRequest(w, r, "Error invalid tweet format.")
		return
	}

	tweetLen := len(tweet.TweetText)
	if tweetLen < 1 || tweetLen > 280 {
		WriteBadRequest(w, r, "Tweet length must be 1-280 chars.")
		return
	}

	finalTweet, err := a.TweetService.AddTweet(tweet.TweetText)
	if err != nil {
		WriteServerError(w, r, errors.Wrap(err, errors.New("service failed to add tweet")))
		return
	}

	WriteJSON(w, r, finalTweet)
}

func (a *API) GetRandTweet(w http.ResponseWriter, r *http.Request) {
	randTweet, err := a.TweetService.AddRandTweet()
	if err != nil {
		WriteServerError(w, r, errors.Wrap(err, errors.New("service failed to add tweet")))
		return
	}

	WriteJSON(w, r, randTweet)
}

func (a *API) EnableCORS() {
	corsConfig := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           100, // Maximum value not ignored by any of major browsers
	})

	a.Router.Use(corsConfig.Handler)
}
