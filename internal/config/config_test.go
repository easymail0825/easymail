package config

import (
	"log"
	"testing"
)

func TestRootConfig(t *testing.T) {
	db.AutoMigrate(&Configure{})

	err := CreateRoot("postfix")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestNodeConfig(t *testing.T) {
	c, err := CreateNode("postfix", "log", "", DataTypeNull)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	log.Println(c)
}

func TestNode2Config(t *testing.T) {
	c, err := CreateNode("log", "mail", "/var/log/mail.log", DataTypeString)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	log.Println(c)
}

func TestGetConfigure(t *testing.T) {
	c, err := GetConfigure("postfix", "log", "mail")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	log.Println(c)
}

func TestCreateFullConfig(t *testing.T) {
	// top level
	top, err := GetConfigure("postfix")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log(top)

	// second level
	//second, err := CreateNodeFromParent(top, "execute", "", DataTypeNull)
	//if err != nil {
	//	t.Log(err)
	//	t.Fail()
	//}
	//t.Log(second)

	second, err := GetConfigure("postfix", "execute")
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	// third level
	// second level
	third, err := CreateNodeFromParent(second, "postcat", "/usr/sbin/postcat", DataTypeString)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log(third)
}
