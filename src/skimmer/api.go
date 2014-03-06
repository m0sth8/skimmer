package skimmer


import (
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"net/http"
	"strconv"
	"github.com/codegangsta/martini-contrib/sessions"
	"fmt"
)



type ErrorMsg struct{
	Error string `json:"error"`
}

const (
	REQUEST_BODY_SIZE = 1024 * 30
	MAX_REQUEST_COUNT = 20
	BIN_LIFETIME = 60 * 60 * 24 * 2
)

type RedisConfig struct {
	RedisAddr			string
	RedisPassword		string
	RedisPrefix			string
}

type Config struct {
	SessionSecret		string
	Storage				string
	RedisConfig
}

func GetApi(config *Config) *martini.ClassicMartini {
	var storage Storage

	switch config.Storage{
	case "redis":
		redisStorage := NewRedisStorage(config.RedisAddr, config.RedisPassword, config.RedisPassword, MAX_REQUEST_COUNT, BIN_LIFETIME)
		redisStorage.StartCleaning(60)
		storage = redisStorage
	default:
		memoryStorage := NewMemoryStorage(MAX_REQUEST_COUNT, BIN_LIFETIME)
		memoryStorage.StartCleaning(60)
		storage = memoryStorage


	}
	store := sessions.NewCookieStore([]byte(config.SessionSecret))

	api := martini.Classic()

	api.MapTo(storage, (*Storage)(nil))
	api.Use(render.Renderer(render.Options{
		Directory: "public/static/views",
		Extensions: []string{".html"},
		Delims: render.Delims{"{[{", "}]}"},
	}))
	api.Use(sessions.Sessions("my_session", store))
	api.Use(NewSessionHistoryHandler(20, "binHistory"))


	api.Post("/api/v1/bins/", func(r render.Render, storage Storage, history History, session sessions.Session, req *http.Request){
			payload := Bin{}
			if err := DecodeJsonPayload(req, &payload); err != nil {
				r.JSON(400, ErrorMsg{fmt.Sprintf("Decoding payload error: %s", err)})
				return
			}
			bin := NewBin()
			if payload.Private {
				bin.SetPrivate()
			}
			if err := storage.CreateBin(bin); err == nil {
				history.Add(bin.Name)
				if bin.Private {
					session.Set(fmt.Sprintf("pr_%s", bin.Name), bin.SecretKey)
				}
				r.JSON(http.StatusCreated, bin)
			} else {
				r.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			}
		})

	api.Get("/api/v1/bins/", func(r render.Render, storage Storage, history History){
			if bins, err := storage.LookupBins(history.All()); err == nil {
				r.JSON(http.StatusOK, bins)
			} else {
				r.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			}
		})

	api.Get("/api/v1/bins/:bin", func(r render.Render, params martini.Params, session sessions.Session, storage Storage){
			if bin, err := storage.LookupBin(params["bin"]); err == nil{
				if bin.Private && bin.SecretKey != session.Get(fmt.Sprintf("pr_%s", bin.Name)){
					r.JSON(http.StatusForbidden, ErrorMsg{"The bin is private"})
				} else {
					r.JSON(http.StatusOK, bin)
				}
			} else {
				r.JSON(http.StatusNotFound, ErrorMsg{err.Error()})
			}
		})

	api.Get("/api/v1/bins/:bin/requests/", func(r render.Render, storage Storage, session sessions.Session,
			params martini.Params, req *http.Request){
			if bin, error := storage.LookupBin(params["bin"]); error == nil {
				if bin.Private && bin.SecretKey != session.Get(fmt.Sprintf("pr_%s", bin.Name)){
					r.JSON(http.StatusForbidden, ErrorMsg{"The bin is private"})
				} else {
					from := 0
					to := 20
					if fromVal, err := strconv.Atoi(req.FormValue("from")); err == nil {
						from = fromVal
					}
					if toVal, err := strconv.Atoi(req.FormValue("to")); err == nil {
						to = toVal
					}
					if requests, err := storage.LookupRequests(bin.Name, from, to); err == nil {
						r.JSON(http.StatusOK, requests)
					} else {
						r.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
					}
				}
			} else {
				r.Error(http.StatusNotFound)
			}
		})

	api.Get("/api/v1/bins/:bin/requests/:request", func(r render.Render, storage Storage, params martini.Params){
			if request, err := storage.LookupRequest(params["bin"], params["request"]); err == nil {
				r.JSON(http.StatusOK, request)
			} else {
				r.JSON(http.StatusNotFound, ErrorMsg{err.Error()})
			}
		})

	api.Any("/bins/:name", func(r render.Render, storage Storage, params martini.Params,
			req *http.Request){
			if bin, error := storage.LookupBin(params["name"]); error == nil {
				request := NewRequest(req, REQUEST_BODY_SIZE)
				if err := storage.CreateRequest(bin, request); err == nil {
					r.JSON(http.StatusOK, request)
				} else {
					r.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
				}
			} else {
				r.Error(http.StatusNotFound)
			}
		})

	api.Get("**", func(r render.Render){
			r.HTML(200, "index", nil)
		})
	return api
}
