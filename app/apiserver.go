package apiserver

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bolshagin/shorty-rest-api/model"
	"github.com/bolshagin/shorty-rest-api/tools"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
)

var (
	errNotAuthorized = errors.New("not authorized")
	errInvalidUserID = errors.New("invalid userid")
	errCantDecodeURL = errors.New("cannot decode short url")
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
	return fmt.Sprintf(
		"%s:%s@/%s",
		config.User,
		config.Password,
		config.DBName,
	)
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
	s.router.HandleFunc("/links", checkAuth(s, s.handleLinksCreate())).Methods("POST")
	s.router.HandleFunc("/links/{userid}", checkAuth(s, s.handleGetAllLinks())).Methods("GET")
	s.router.HandleFunc("/link", checkAuth(s, s.handleGetLinkInfo())).Methods("GET")
	s.router.HandleFunc("/link", checkAuth(s, s.handleLinkDelete())).Methods("DELETE")
	s.router.HandleFunc("/{short_url}", s.handleRedirect()).Methods("GET")
	s.router.HandleFunc("/stats/top", checkAuth(s, s.handleGetLinksTop())).Methods("GET")
	s.router.HandleFunc("/me/", checkAuth(s, s.handleMe())).Methods("GET")
}

func checkAuth(s *APIServer, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		s.logger.Info(fmt.Sprintf("try authorizate user with token `%v`", token))
		email, password, err := s.decodeToken(token)

		if err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusUnauthorized, err)
			return
		}

		u := &model.User{
			Email:    email,
			Password: password,
		}

		u, err = u.FindByEmailAndPassword(s.db)
		if err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusUnauthorized, err)
			return
		}

		next(w, r)
	}
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

		token := r.Header.Get("Authorization")
		email, password, _ := s.decodeToken(token)

		u := &model.User{
			Email:    email,
			Password: password,
		}
		u, err := u.FindByEmailAndPassword(s.db)
		if err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		l := &model.Link{}
		l.UserID = u.UserID
		l.LongURL = req.LongURL

		l, err = l.CreateLink(s.db, r)
		if err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		l.ClearUserID()
		s.logger.Info(fmt.Sprintf("link with url '%s' successfully created", l.LongURL))
		s.respond(w, r, http.StatusCreated, l)
	}
}

func (s *APIServer) handleGetAllLinks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["userid"])
		if err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusBadRequest, errInvalidUserID)
			return
		}

		s.logger.Info(fmt.Sprintf("handle /links/{userid} -> try to find all links by given userid '%v'", id))

		u := &model.User{}
		u, err = u.Find(id, s.db)
		if err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		u, err = u.FindAllLinks(id, s.db)
		if err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		u.ClearPassword()
		s.respond(w, r, http.StatusOK, u)
	}
}

func (s *APIServer) handleGetLinksTop() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("handle /stats/top -> try to find top 20 redirecting links")
		links, err := model.FindLinksTop(s.db)
		if err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, links)
	}
}

func (s *APIServer) handleGetLinkInfo() http.HandlerFunc {
	type request struct {
		UserID   int    `json:"userid"`
		ShortURL string `json:"short_url"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		s.logger.Info("handle /link -> try to find link info by userid and short_url")
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		l := &model.Link{}
		l, err := l.FindLinkClicks(req.UserID, req.ShortURL, s.db)

		if err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		l.ClearUserID()

		s.logger.Info(
			fmt.Sprintf("link with short_url '%v' and userid '%v' are founded", req.ShortURL, req.UserID),
		)
		s.respond(w, r, http.StatusOK, l)
	}
}

func (s *APIServer) handleLinkDelete() http.HandlerFunc {
	type request struct {
		ShortURL string `json:"short_url"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		s.logger.Info("handle /link DELETE -> try to delete link by userid and short_url")
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		token := r.Header.Get("Authorization")
		email, password, _ := s.decodeToken(token)
		u := &model.User{
			Email:    email,
			Password: password,
		}
		u, err := u.FindByEmailAndPassword(s.db)
		if err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		l := &model.Link{}
		if err := l.DeleteByUserIDAndShort(u.UserID, req.ShortURL, s.db); err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.logger.Info(
			fmt.Sprintf("link with short_url '%v' and userid '%v' has been deleted", req.ShortURL, u.UserID),
		)
		s.respond(w, r, http.StatusOK, map[string]string{"result": "deleted"})

	}
}

func (s *APIServer) handleRedirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := tools.Decode(vars["short_url"])
		if err != nil {
			s.logger.Error(errCantDecodeURL)
			s.error(w, r, http.StatusBadRequest, errCantDecodeURL)
			return
		}

		l := &model.Link{}
		l, err = l.Find(id, s.db)

		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		c := &model.Click{}
		_, err = c.CreateClick(id, s.db)
		if err != nil {
			s.logger.Error("cannot create click for short url")
		}

		s.logger.Info(fmt.Sprintf("redirecting to '%s'", l.LongURL))
		http.Redirect(w, r, l.LongURL, http.StatusPermanentRedirect)
	}
}

func (s *APIServer) handleMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		email, password, _ := s.decodeToken(token)

		u := &model.User{
			Email:    email,
			Password: password,
		}

		u, _ = u.FindByEmailAndPassword(s.db)
		u.ClearPassword()
		s.respond(w, r, http.StatusOK, u)
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

func (s *APIServer) decodeToken(token string) (string, string, error) {
	auth := strings.SplitN(token, " ", 2)
	if len(auth) != 2 {
		return "", "", errNotAuthorized
	}

	b, err := base64.StdEncoding.DecodeString(auth[1])
	if err != nil {
		return "", "", errNotAuthorized
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return "", "", errNotAuthorized
	}
	return pair[0], pair[1], nil
}
