package service

import (
	"database/sql"
	"fmt"
)

func (h *RouteHandler) getChartData(meta ChartsInfo, payload Payload) (interface{}, error) {

	switch meta.ChartType {
	case BarChartType:

		var groupBy string
		var agg = agg(meta.ChartMeta.Agg, meta.ChartMeta.Y)
		if agg != meta.ChartMeta.Y {
			groupBy = fmt.Sprintf("GROUP BY %s", meta.ChartMeta.X)
		}

		query := fmt.Sprintf("SELECT %s, %s FROM %s %s", meta.ChartMeta.X, agg, meta.ChartTableName, groupBy)
		rows, err := h.db.Query(query)
		if err != nil {
			return nil, err
		}

		var barChartData = BarChartStruct{
			TableName: meta.ChartTableName,
			LabelName: meta.ChartMeta.X,
			Data:      make([]BarUnit, 0),
		}

		for rows.Next() {
			var x string
			var y int
			err := rows.Scan(&x, &y)
			if err != nil {
				return nil, err
			}
			barChartData.Data = append(barChartData.Data, BarUnit{X: x, Y: y})
		}

		return barChartData, nil

	case TimeseriesType:

		var condition = ""
		if payload.StartDate != 0 {
			condition = fmt.Sprintf("WHERE (timestamp >= %d AND timestamp <= %d)", payload.From(), payload.To())
		}

		rows, err := h.db.Query(
			fmt.Sprintf(`SELECT timestamp, %s, %s FROM %s %s ORDER BY timestamp DESC`,
				meta.ChartMeta.X,
				meta.ChartMeta.Y,
				meta.ChartTableName,
				condition,
			))

		if err != nil {
			return nil, err
		}

		var timeseriesData = TimeSeriesStruct{
			TableName: meta.ChartTableName,
			Data:      make([]TimeSeriesUnit, 0),
		}

		for rows.Next() {
			var timestamp sql.NullTime
			var x string
			var y int

			err := rows.Scan(&timestamp, &x, &y)
			if err != nil {
				return nil, err
			}

			timeseriesData.Data = append(timeseriesData.Data, TimeSeriesUnit{
				Tag:   x,
				Value: y,
				Time:  timestamp.Time.Unix() * 1000,
			})
		}

		return timeseriesData, nil

	default:

		return nil, nil
	}

}
