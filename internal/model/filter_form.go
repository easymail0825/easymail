package model

type IndexFilterFieldRequest struct {
	DataTableRequest
}

type IndexFilterMetricRequest struct {
	DataTableRequest
	OrderField string `json:"orderField;omitempty"`
	OrderDir   string `json:"orderDir;omitempty"`
}

type IndexFilterMetricResponse struct {
	FilterMetric
	PrimaryFieldName   string `json:"primary_field_name"`
	SecondaryFieldName string `json:"secondary_field_name"`
}

type CreateFilterMetricRequest struct {
	ID             int64           `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	PrimaryField   int64           `json:"primary_field"`
	SecondaryField int64           `json:"secondary_field"`
	Operation      MetricOperation `json:"operation"`
	Category       FilterCategory  `json:"category"`
	Interval       int             `json:"interval"`
	Unit           MetricUnit      `json:"unit"`
	AccountID      int64           `json:"account_id"`
}
