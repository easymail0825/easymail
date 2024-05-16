package controller

import (
	"easymail/internal/model"
)

type menu struct {
	ConfigureNodes []model.Configure
}

func createMenu() *menu {
	m := &menu{}
	nodes, err := model.GetRootConfigureRootNodes()
	if err != nil {
		return m
	}
	m.ConfigureNodes = nodes
	return m
}
