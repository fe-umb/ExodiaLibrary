package web

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/lcmps/ExodiaLibrary/app"
	"github.com/lcmps/ExodiaLibrary/db"
	"github.com/lcmps/ExodiaLibrary/model"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var store = sessions.NewCookieStore([]byte("secret"))
var conf *oauth2.Config
var state string

type WebApp struct {
	Config *app.Config
	Logger *logrus.Logger
}

func NewAuth(client_id, client_secret, redirect_uri string) *oauth2.Config {

	conf = &oauth2.Config{
		ClientID:     client_id,
		ClientSecret: client_secret,
		RedirectURL:  redirect_uri,
		Scopes:       []string{"email", "profile"},
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://oauth2.googleapis.com/token",
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
		},
	}

	return conf
}

func getLoginURL(state string) string {
	return conf.AuthCodeURL(state)
}

func New() (*WebApp, error) {
	var App WebApp
	l := logrus.New()

	appData, err := app.InitConfig()
	if err != nil {
		return &App, err
	}

	App.Config = appData
	App.Logger = l

	return &App, nil
}

func (App *WebApp) Host() {

	r := gin.Default()
	r.Use(sessions.Sessions("exodialib", store))
	// Assets
	r.Static("/css", "./pages/assets/css")
	r.Static("/js", "./pages/assets/js")
	r.Static("/fvc", "./pages/assets/fvc")
	r.Static("/img", "./pages/assets/img")
	r.Static("/card-img", "./pages/card-img")
	r.Static("/fonts", "./pages/assets/fonts")

	// Auth
	conf = NewAuth(App.Config.Client_id, App.Config.Client_secret, App.Config.Redirect_url)

	// API
	r.GET("/card", getAllCards)
	r.GET("/random", GetRandomCards)
	r.GET("/auth/google/callback", authHandler)
	r.GET("/login", loginHandler)

	// Pages
	r.GET("/", home)

	// HTML
	r.LoadHTMLGlob("./pages/html/*.html")

	gin.SetMode(App.Config.Env)
	err := r.Run(":" + App.Config.Web_Port)
	if err != nil {
		App.Logger.Error(err.Error())
	}
}

func home(ctx *gin.Context) {
	ctx.HTML(
		http.StatusOK,
		"index.html",
		gin.H{
			"title": "𓂀 Exodia Library 𓂀",
		},
	)
}

func getAllCards(ctx *gin.Context) {
	var q model.CardQuery

	err := ctx.ShouldBindQuery(&q)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cn, err := db.InitConnection()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	m := cn.GetCardsByFilter(q)
	ctx.JSON(http.StatusOK, m)
}

func GetRandomCards(ctx *gin.Context) {

	var lim struct {
		Limit int `json:"limit"`
	}

	err := ctx.ShouldBindQuery(&lim)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cn, err := db.InitConnection()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m := cn.GetRandomCards(lim.Limit)
	ctx.JSON(http.StatusOK, m)
}

func loginHandler(ctx *gin.Context) {
	state = app.RandToken()
	session := sessions.Default(ctx)
	session.Set("state", state)
	session.Save()

	ctx.HTML(
		http.StatusOK,
		"login.html",
		gin.H{
			"title":    "𓂀 Exodia Library 𓂀",
			"callback": getLoginURL(state),
		},
	)
}

func authHandler(ctx *gin.Context) {
	session := sessions.Default(ctx)
	retrievedState := session.Get("state")
	if retrievedState != ctx.Query("state") {
		ctx.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid session state: %s", retrievedState))
		return
	}

	tok, err := conf.Exchange(context.Background(), ctx.Query("code"))
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	client := conf.Client(context.Background(), tok)
	email, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer email.Body.Close()
	data, _ := ioutil.ReadAll(email.Body)
	log.Println("Email body: ", string(data))
	ctx.Status(http.StatusOK)
}
