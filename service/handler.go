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
	Salt           = "!@ABC)23"
	BarChartType   = 1
	TimeseriesType = 2
	PieChartType   = 3
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
	mux.HandleFunc("/get-charts/{table}", h.setDefaultHeaders(h.chart))
	mux.HandleFunc("/create", h.setDefaultHeaders(h.create))
	mux.HandleFunc("/delete/{table}", h.setDefaultHeaders(h.delete))
	mux.HandleFunc("/fill", h.setDefaultHeaders(h.fill))

	err := http.ListenAndServe(":3333", mux)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (h *RouteHandler) delete(w http.ResponseWriter, r *http.Request) {
	var table = r.PathValue("table")
	var credits Auth

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		h.response(w, 500, "Error when reading body, try again.", nil)
		return
	}

	err = json.Unmarshal(raw, &credits)
	if err != nil {
		h.response(w, 500, "Error when reading credits, try again.", nil)
		return
	}

	var userId int
	err = h.db.QueryRow("SELECT id FROM users WHERE username = $1 AND password = $2", credits.Username, credits.Token).Scan(&userId)
	if err != nil || userId == 0 {
		h.response(w, 500, "Incorrect user or password, try again.", nil)
		return
	}

	stmt, err := h.db.Begin()
	defer stmt.Rollback()
	_, err = stmt.Exec(fmt.Sprintf("DROP TABLE %s", table))
	if err != nil {
		return
	}

	_, err = stmt.Exec("DELETE FROM charts_meta WHERE user_id = $1 AND chart_table_name = $2", userId, table)
	if err != nil {
		return
	}

	err = stmt.Commit()
	if err != nil {
		return
	}

	h.response(w, 200, "", nil)
}

func (h *RouteHandler) create(w http.ResponseWriter, r *http.Request) {

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var info CreateChart
	err = json.Unmarshal(raw, &info)

	var userId int
	err = h.db.QueryRow("SELECT id FROM users WHERE username = $1 AND password = $2", info.User.Username, info.User.Token).Scan(&userId)
	if err != nil || userId == 0 {
		h.response(w, 500, "Incorrect user or password, try again.", nil)
		return
	}

	stmt, err := h.db.Begin()
	if err != nil {
		return
	}
	defer stmt.Rollback()

	metaData := ChartMeta{
		X:   info.Data.XAxisColumnName,
		Y:   info.Data.YAxisColumnName,
		Agg: info.Data.YAxisColumnAgg,
	}

	metaRaw, _ := json.Marshal(metaData)

	_, err = stmt.Exec("INSERT INTO charts_meta (user_id, chart_type, chart_table_name, meta) VALUES ($1,$2,$3,$4)", userId, info.Data.BarType, info.Data.TableName, metaRaw)
	if err != nil {
		h.response(w, 500, "Incorrect user or password, try again.", nil)
	}

	switch info.Data.BarType {
	case BarChartType, PieChartType:

		query := fmt.Sprintf("CREATE TABLE %s (%s TEXT not null, %s %s not null);",
			info.Data.TableName,
			info.Data.XAxisColumnName,
			info.Data.YAxisColumnName,
			info.Data.YAxisColumnType,
		)

		_, err = stmt.Exec(query)
		if err != nil {
			return
		}

		err = stmt.Commit()
		if err != nil {
			return
		}

		break
	case TimeseriesType:

		query := fmt.Sprintf("CREATE TABLE %s (timestamp DATETIME not null, %s TEXT not null, %s %s not null);",
			info.Data.TableName,
			info.Data.XAxisColumnName,
			info.Data.YAxisColumnName,
			info.Data.YAxisColumnType,
		)

		_, err = stmt.Exec(query)
		if err != nil {
			return
		}

		err = stmt.Commit()
		if err != nil {
			return
		}

		break
	default:
		break
	}

	h.response(w, 200, "", ChartsInfo{
		ChartType:      info.Data.BarType,
		ChartTableName: info.Data.TableName,
		ChartMeta:      metaData,
	})
}

func (h *RouteHandler) chart(w http.ResponseWriter, r *http.Request) {

	var table = r.PathValue("table")

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var info GetChartStruct
	err = json.Unmarshal(raw, &info)
	if err != nil {
		h.response(w, 500, "An error occurred during authentication", nil)
		return
	}

	var meta ChartsInfo
	var metaRaw []byte
	h.db.QueryRow(`
		SELECT chart_type, meta, chart_table_name FROM charts_meta 
		WHERE chart_table_name = $1 
		AND user_id = (SELECT id FROM users WHERE password = $2)
	`, table, info.Token).Scan(&meta.ChartType, &metaRaw, &meta.ChartTableName)

	err = json.Unmarshal(metaRaw, &meta.ChartMeta)
	if err != nil {
		h.response(w, 500, "An error occurred during getting data", nil)
		return
	}

	if meta.ChartType == 0 {
		h.response(w, 500, "No access for the table: "+table, nil)
		return
	}

	object, err := h.getChartData(meta, info.Payload)
	if err != nil {
		h.response(w, 500, "An error occurred during getting data", nil)
		return
	}

	h.response(w, 200, "success", object)
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
		SELECT chart_type, chart_table_name, meta FROM charts_meta
		WHERE user_id = (SELECT id FROM users WHERE password = $1 LIMIT 1)
	`, info.Token)
	if err != nil {
		h.response(w, 500, "An error occurred during authentication", nil)
		return
	}

	var charts = make([]ChartsInfo, 0)
	for query.Next() {
		var chart ChartsInfo
		var metaRaw []byte
		err := query.Scan(&chart.ChartType, &chart.ChartTableName, &metaRaw)
		if err != nil {
			h.response(w, 500, "An error occurred getting login and password from income data", nil)
			return
		}
		err = json.Unmarshal(metaRaw, &chart.ChartMeta)
		if err != nil {
			return
		}
		charts = append(charts, chart)
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
