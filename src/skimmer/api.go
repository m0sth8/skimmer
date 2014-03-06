package skimmer


import (
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"net/http"
	"strconv"
)



type ErrorMsg struct{
	Error string `json:"error"`
}

const (
	REQUEST_BODY_SIZE = 1024 * 30
	MAX_REQUEST_COUNT = 20
)

func GetApi() *martini.ClassicMartini {
	history := []string{}
	memoryStorage := NewMemoryStorage(MAX_REQUEST_COUNT)


	api := martini.Classic()

	api.Use(render.Renderer())
	api.MapTo(memoryStorage, (*Storage)(nil))

	api.Post("/api/v1/bins/", func(r render.Render, storage Storage){
			bin := NewBin()
			if err := storage.CreateBin(bin); err == nil {
				history = append(history, bin.Name)
				r.JSON(http.StatusCreated, bin)
			} else {
				r.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			}
		})

	api.Get("/api/v1/bins/", func(r render.Render, storage Storage){
			if bins, err := storage.LookupBins(history); err == nil {
				r.JSON(http.StatusOK, bins)
			} else {
				r.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			}
		})

	api.Get("/api/v1/bins/:bin", func(r render.Render, params martini.Params, storage Storage){
			if bin, err := storage.LookupBin(params["bin"]); err == nil{
				r.JSON(http.StatusOK, bin)
			} else {
				r.JSON(http.StatusNotFound, ErrorMsg{err.Error()})
			}
		})

	api.Get("/api/v1/bins/:bin/requests/", func(r render.Render, storage Storage, params martini.Params,
			req *http.Request){
			if bin, error := storage.LookupBin(params["bin"]); error == nil {
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
			req *http.Request, res http.ResponseWriter){
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
	return api
}
