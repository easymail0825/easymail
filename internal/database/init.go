package database

func init() {
	appConfig, err := ReadAppConfig("easymail.yaml")
	if err != nil {
		panic(err)
	}
	// initialize database first
	err = initMySQL(appConfig.Mysql)
	if err != nil {
		panic("failed to connect to database")
	}
	initRedis(appConfig.Redis)
}
