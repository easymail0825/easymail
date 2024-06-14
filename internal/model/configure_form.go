package model

type ConfigureNodeRequest struct {
	DataTableRequest
	SubNodeID uint   `json:"subNodeID"`
	Keyword   string `json:"keyword"`
}

type ConfigureNodeResponse struct {
	ID          uint   `json:"id"`
	TopName     string `json:"topName"`
	SubName     string `json:"subName"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	DataType    uint   `json:"dataType"`
	Private     bool   `json:"private"`
	CreateTime  string `json:"createTime"`
	Description string `json:"description"`
}
