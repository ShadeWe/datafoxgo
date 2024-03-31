package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (h *RouteHandler) fill(w http.ResponseWriter, r *http.Request) {

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var info BarChartFill
	err = json.Unmarshal(raw, &info)
	if err != nil {
		h.response(w, 500, "An error occurred during authentication", nil)
		return
	}

	var metaRaw []byte
	err = h.db.QueryRow("SELECT meta FROM charts_meta WHERE chart_table_name = $1", info.Table).Scan(&metaRaw)
	if err != nil {
		return
	}

	var meta ChartMeta
	err = json.Unmarshal(metaRaw, &meta)
	if err != nil {
		return
	}

	_, err = h.db.Exec(fmt.Sprintf("INSERT INTO %s VALUES ($1, $2)", info.Table), info.Tag, info.Value)
	if err != nil {
		return
	}
}
