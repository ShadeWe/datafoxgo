package service

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const Salt = "!@ABC)23"

type UserData struct {
	Username string `json:"username"`
	Token    string `json:"token"`
	UserId   int    `json:"user_id"`
}

type LoginInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

type ResponseFunction func(w http.ResponseWriter, r *http.Request)
type RouteHandler struct {
	db *sql.DB
}

func NewRouteHandler(db *sql.DB) *RouteHandler {
	return &RouteHandler{
		db,
	}
}

func (h *RouteHandler) Run() error {

	mux := http.NewServeMux()
	mux.HandleFunc("/login", h.setDefaultHeaders(h.login))

	err := http.ListenAndServe(":3333", mux)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (h *RouteHandler) login(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		h.response(w, 200, "", nil)
	}

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var info LoginInfo
	err = json.Unmarshal(raw, &info)
	if err != nil {
		h.response(w, 500, "An error occurred getting login and password from income data", nil)
		return
	}

	var saltedPassword = fmt.Sprintf("%x", md5.Sum([]byte(info.Password+Salt)))

	var userId int
	err = h.db.QueryRow("SELECT id FROM users WHERE username = $1 AND password = $2", info.Username, saltedPassword).Scan(&userId)
	if err != nil || userId == 0 {
		h.response(w, 500, "Incorrect user or password, try again.", nil)
		return
	}

	h.response(w, 200, "login success", UserData{
		UserId:   userId,
		Username: info.Username,
		Token:    saltedPassword,
	})
}
