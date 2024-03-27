package service

type ChartsMeta struct {
	ChartType      int            `json:"chart_type"`
	ChartTableName string         `json:"chart_table_name"`
	BarChartData   map[string]int `json:"bar_chart_data_raw"`
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
