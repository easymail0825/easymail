package filter

import (
	"easymail/internal/model"
)

func featureSwitch(names []string) bool {
	if len(names) == 0 {
		return false
	}
	c, err := model.GetConfigureByNames(names...)
	if err != nil {
		return false
	}
	if c.DataType != model.DataTypeBool {
		return false
	}
	return c.Value == "true"
}
