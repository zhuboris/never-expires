package httpmux

import (
	"html/template"
	"net/http"
	"strings"
)

const (
	publicDirPath = "./web/swagger/"

	swaggerInitFileName = "swagger-initializer.js"
)

const (
	swaggerRoute     = "/swagger/"
	swaggerSpecRoute = "/.swagger"
)

func (m *Mux) HandleSwaggerBySpecification(pathToSpec string) {
	m.HandleFunc(swaggerSpecRoute, m.handleSwaggerSpec(pathToSpec))

	m.HandleFunc(swaggerRoute, m.handleSwagger())
}

func (m *Mux) handleSwaggerSpec(pathToSpec string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, pathToSpec)
	}
}

func (m *Mux) handleSwagger() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, swaggerRoute)
		if name != swaggerInitFileName {
			http.StripPrefix(swaggerRoute, http.FileServer(http.Dir(publicDirPath))).ServeHTTP(w, r)
			return
		}

		initTemplate, err := template.ParseFiles(publicDirPath + swaggerInitFileName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		url, err := EndpointURL(swaggerSpecRoute, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		data := struct {
			SpecificationURL string
		}{
			SpecificationURL: url,
		}

		err = initTemplate.Execute(w, data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
