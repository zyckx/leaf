package main

import (
	"fmt"

	"kuukaa.fun/leaf/cache"
	"kuukaa.fun/leaf/db/mysql"
	"kuukaa.fun/leaf/initialize"
	"kuukaa.fun/leaf/logger"
	"kuukaa.fun/leaf/routes"
	"kuukaa.fun/leaf/service"
)

func main() {
	// 初始化配置文件
	initialize.ConfigFiles()
	// 初始化日志
	logger.InitLogger()
	// 初始化mysql
	mysql.Init()
	// 初始化数据库表
	mysql.InitTables()
	// 初始化缓存
	cache.Init()
	// 初始化mysql客户端
	service.InitMysqlClient()

	// fmt.Println(jwt.GenerateAccessToken(1))

	count := cache.GetLoginTryCount("2")
	fmt.Printf("count: %v\n", count)

	// 初始化路由
	routes.InitRouter()
}
