package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitConfig() {
	// viper
	// viper 去查找名为 app 的配置文件（注意：这里只指定了文件名，不包含扩展名）
	viper.SetConfigName("app")
	viper.AddConfigPath("config")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("config app inited:", viper.Get("app"))
}

func InitMSQL() {
	// 打印 SQL 日志
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // 日志级别
			Colorful:      true,        // 彩色打印
		},
	)
	DB, _ = gorm.Open(mysql.Open(viper.GetString("mysql.dns")), &gorm.Config{Logger: newLogger})
	fmt.Println("config mysql inited:", viper.Get("mysql"))
	// if err != nil {
	// 	fmt.Println("未能连接数据库:", err)
	// }
	// user := models.UserBasic{}
	// db.Find(&user)
	// fmt.Println(user)
}
