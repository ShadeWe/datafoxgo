package service

type ChartsInfo struct {
	ChartType      int       `json:"chart_type"`
	ChartTableName string    `json:"chart_table_name"`
	ChartMeta      ChartMeta `json:"meta"`
}

type Auth struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

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

type CreatChartDetails struct {
	BarType         int    `json:"barType"`
	TableName       string `json:"tableName"`
	XAxisColumnName string `json:"xAxisColumnName"`
	YAxisColumnName string `json:"yAxisColumnName"`
	YAxisColumnType string `json:"yAxisColumnType"`
	YAxisColumnAgg  string `json:"yAxisColumnAgg"`
}

type CreateChart struct {
	User Auth              `json:"user"`
	Data CreatChartDetails `json:"data"`
}

type ChartMeta struct {
	X   string `json:"x"`
	Y   string `json:"y"`
	Agg string `json:"agg"`
}

type ChartFill struct {
	Table     string `json:"table"`
	Tag       string `json:"tag"`
	Value     int    `json:"value"`
	Timestamp int    `json:"timestamp,omitempty"`
}

type DoubleUnit struct {
	X string  `json:"x"`
	Y float64 `json:"y"`
}

type BarUnit struct {
	X string `json:"x"`
	Y int64  `json:"y"`
}

type BarChartStruct struct {
	TableName string        `json:"table_name"`
	LabelName string        `json:"label_name"`
	Data      []interface{} `json:"data"`
}

type TimeSeriesUnit struct {
	Time  int64  `json:"time"`
	Tag   string `json:"tag"`
	Value int    `json:"value"`
}

type TimeSeriesStruct struct {
	TableName string           `json:"table_name"`
	Data      []TimeSeriesUnit `json:"data"`
}

type Payload struct {
	StartDate int64 `json:"start_date"`
	EndDate   int64 `json:"end_date"`
}

func (g Payload) From() int64 {
	return g.StartDate
}

func (g Payload) To() int64 {
	return g.EndDate
}

type GetChartStruct struct {
	Auth
	Payload `json:"payload"`
}
