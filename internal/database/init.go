package database

func init() {
	// read config

	var err error
	dsn := "easymail:easymail@tcp(localhost:3306)/easymail?charset=utf8&parseTime=True&loc=Local"
	err = InitDB(dsn)
	if err != nil {
		panic("failed to connect to database")
	}

	InitRedis()
}
