package skimmer


import (
	"fmt"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"net/http"
	"net/http/httputil"
)



type ErrorMsg struct{
	Error string `json:"error"`
}

func GetApi() *martini.ClassicMartini {
	bins := map[string]*Bin{}
	history := []string{}

	api := martini.Classic()

	api.Use(render.Renderer())

	api.Post("/api/v1/bins/", func(r render.Render){
			bin := NewBin()
			bins[bin.Name] = bin
			history = append(history, bin.Name)
			r.JSON(http.StatusCreated, bin)
		})

	api.Get("/api/v1/bins/", func(r render.Render){
			filteredBins := []*Bin{}
			for _, name := range(history) {
				if bin, ok := bins[name]; ok {
					filteredBins = append(filteredBins, bin)
				}
			}
			r.JSON(http.StatusOK, filteredBins)
		})

	api.Get("/api/v1/bins/:bin", func(r render.Render, params martini.Params){
			if bin, ok := bins[params["bin"]]; ok{
				r.JSON(http.StatusOK, bin)
			} else {
				r.Error(http.StatusNotFound)
			}
		})

	api.Any("/", func(res http.ResponseWriter, req *http.Request) {
			if dumped, err := httputil.DumpRequest(req, true); err == nil {
				res.WriteHeader(http.StatusOK)
				res.Write(dumped)
			} else {
				res.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(res, "Error: %v", err)
			}
		})
	return api
}
