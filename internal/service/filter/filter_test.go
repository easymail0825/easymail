package filter

import (
	"easymail/internal/model"
	"easymail/internal/service/milter"
	"encoding/json"
	"net"
	"strconv"
	"testing"
)

func TestQueryRegion(t *testing.T) {
	country, province, city, err := QueryRegion(net.ParseIP("120.8.3.3"))
	if err != nil {
		t.Error(err)
	}
	t.Log(country, province, city)
}

func TestFeature2Json(t *testing.T) {
	feature := make([]milter.Feature, 0)
	feature = append(feature, milter.Feature{
		Name:      "testStr",
		Value:     "123",
		ValueType: milter.DataTypeString,
	})
	feature = append(feature, milter.Feature{
		Name:      "testInt",
		Value:     "123",
		ValueType: milter.DataTypeInt,
	})
	feature = append(feature, milter.Feature{
		Name:      "testFloat",
		Value:     "123.45",
		ValueType: milter.DataTypeFloat,
	})
	feature = append(feature, milter.Feature{
		Name:      "testBool",
		Value:     "true",
		ValueType: milter.DataTypeBool,
	})

	FeatureMap := make(map[string]any)
	for _, f := range feature {
		t.Log(f.Name, f.Value, f.ValueType)
		switch f.ValueType {
		case milter.DataTypeString:
			FeatureMap[f.Name] = f.Value
		case milter.DataTypeInt:
			FeatureMap[f.Name], _ = strconv.Atoi(f.Value)
		case milter.DataTypeFloat:
			FeatureMap[f.Name], _ = strconv.ParseFloat(f.Value, 64)
		case milter.DataTypeBool:
			FeatureMap[f.Name], _ = strconv.ParseBool(f.Value)
		}
	}

	jsonStr, err := json.Marshal(FeatureMap)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(jsonStr))
}

func TestFormatTime(t *testing.T) {
	//t.Log(truncateTimeByMinutes(time.Now(), 10))
	fields, err := model.GetFilterFieldByStage(model.FilterStageHeader)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fields {
		t.Log(f.ID, f.Name)
	}

	//feature, err := model.GetFilterMetricByID(1)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//t.Log(feature.Name, feature.ID)

	metrics, err := model.GetFilterMetricByStage(model.FilterStageHeader)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range metrics {
		t.Log(m.Unit, m.MakeFilterMetricKey(), m.Operation, m.PrimaryField.Stage, m.SecondaryField.Stage)
	}

}
