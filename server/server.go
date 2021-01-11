package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kataras/golog"
)

const (
	apiRouterPrefix = "/api"
)

// Server server
type Server struct {
	server *http.Server
}

func initMux() *mux.Router {
	m := mux.NewRouter()
	apiRoute := m.PathPrefix(apiRouterPrefix).Subrouter()
	apiRoute.Use(mux.CORSMethodMiddleware(apiRoute))
	apiRoute.HandleFunc("/lives", getAllLives).Methods("GET")
	apiRoute.HandleFunc("/infos", getAllInfos).Methods("GET")
	apiRoute.HandleFunc("/save", saveConfig).Methods("GET")
	apiRoute.HandleFunc("/add", addRooms).Methods("POST")
	apiRoute.HandleFunc("/delete", deleteRooms).Methods("POST")
	apiRoute.HandleFunc("/decode", manualDecode).Methods("POST")
	apiRoute.HandleFunc("/upload", manualUpload).Methods("POST")
	return m
}

// New new
func New() *Server {
	httpServer := &http.Server{
		Addr:    "127.0.0.1:18080",
		Handler: initMux(),
	}
	server := &Server{server: httpServer}
	return server
}

// Start start
func (s *Server) Start() error {
	go func() {
		switch err := s.server.ListenAndServe(); err {
		case nil, http.ErrServerClosed:
		default:
			golog.Error(err)
		}
	}()
	golog.Info("Server start at ", s.server.Addr)
	return nil
}
