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

const (
	Salt = "!@ABC)23"

	BarChartType = 1
)

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
	mux.HandleFunc("/get-charts", h.setDefaultHeaders(h.charts))

	err := http.ListenAndServe(":3333", mux)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (h *RouteHandler) charts(w http.ResponseWriter, r *http.Request) {

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var info Auth
	err = json.Unmarshal(raw, &info)
	if err != nil {
		h.response(w, 500, "An error occurred during authentication", nil)
		return
	}

	if info.Token == "" {
		h.response(w, 500, "An error occurred during authentication", nil)
		return
	}

	query, err := h.db.Query(`
		SELECT chart_type, chart_table_name FROM charts_meta
		WHERE user_id = (SELECT id FROM users WHERE password = $1 LIMIT 1)
	`, info.Token)
	if err != nil {
		h.response(w, 500, "An error occurred during authentication", nil)
		return
	}

	var charts = make([]ChartsMeta, 0)
	for query.Next() {
		var chart ChartsMeta
		err := query.Scan(&chart.ChartType, &chart.ChartTableName)
		if err != nil {
			h.response(w, 500, "An error occurred getting login and password from income data", nil)
			return
		}
		charts = append(charts, chart)
	}

	for key, chartData := range charts {

		switch chartData.ChartType {
		case BarChartType:

			rows, err := h.db.Query(fmt.Sprintf("SELECT x_axis, SUM(y_axis) FROM %s GROUP BY x_axis", chartData.ChartTableName))
			if err != nil {
				return
			}

			var barChartData = make(map[string]int)

			for rows.Next() {
				var x string
				var y int
				err := rows.Scan(&x, &y)
				if err != nil {
					return
				}
				barChartData[x] = y
			}

			if len(barChartData) == 0 {
				break
			}

			chartData.BarChartData = barChartData
			charts[key] = chartData

		default:
			fmt.Println("default")
		}

	}

	h.response(w, 200, "", charts)
	return
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
