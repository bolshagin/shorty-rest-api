package apiserver

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/bolshagin/shorty-rest-api/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
)

type APIServer struct {
	config *Config
	logger *logrus.Logger
	router *mux.Router
	db     *sql.DB
}

func New(config *Config) *APIServer {
	return &APIServer{
		config: config,
		logger: logrus.New(),
		router: mux.NewRouter(),
		db:     &sql.DB{},
	}
}

func (s *APIServer) Start() error {
	level, err := logrus.ParseLevel(s.config.LogLevel)
	if err != nil {
		return err
	}

	s.logger.SetLevel(level)
	s.configureRouter()
	if err := s.configureDB(); err != nil {
		return err
	}

	s.logger.Info("starting api server")

	return http.ListenAndServe(s.config.BindAddr, s.router)
}

func getConnectionString(config *Config) string {
	return fmt.Sprintf("%s:%s@/%s",
		config.User,
		config.Password,
		config.DBName)
}

func (s *APIServer) configureDB() error {
	cs := getConnectionString(s.config)
	db, err := sql.Open("mysql", cs)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *APIServer) configureRouter() {
	s.router.HandleFunc("/users", s.handleUsersCreate()).Methods("POST")
	s.router.HandleFunc("/links", s.handleLinksCreate()).Methods("POST")
}

func (s *APIServer) handleUsersCreate() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		s.logger.Info("handle /users -> create user")
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u := &model.User{}
		u.Email = req.Email
		u.Password = req.Password

		u, err := u.CreateUser(s.db)
		if err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		u.ClearPassword()
		s.logger.Info(fmt.Sprintf("user with email '%s' successfully created", u.Email))
		s.respond(w, r, http.StatusCreated, u)
	}
}

func (s *APIServer) handleLinksCreate() http.HandlerFunc {
	type request struct {
		LongURL string `json:"long_url"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		s.logger.Info("handle /links -> create link")
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		l := &model.Link{}
		l.LongURL = req.LongURL

		l, err := l.CreateLink(s.db)
		if err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.logger.Info(fmt.Sprintf("link with url '%s' successfully created", l.LongURL))
		s.respond(w, r, http.StatusCreated, l)
	}
}

func (s *APIServer) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *APIServer) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
