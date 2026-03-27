package model

type Column struct {
	Data       string `json:"data"`
	Name       string `json:"name"`
	Searchable bool   `json:"searchable"`
	Orderable  bool   `json:"orderable"`
	Search     Search `json:"search"`
}

type Order struct {
	Column int    `json:"column"`
	Dir    string `json:"dir"`
}

type Search struct {
	Value string `json:"value"`
	Regex bool   `json:"regex"`
}

type DataTableRequest struct {
	Draw    int      `json:"draw"`
	Columns []Column `json:"columns"`
	Orders  []Order  `json:"order"`
	Start   int      `json:"start"`
	Length  int      `json:"length"`
	Search  Search   `json:"search"`
}
