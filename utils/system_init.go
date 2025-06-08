package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB  *gorm.DB
	Red *redis.Client
)

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

func InitMySQL() {
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

func InitRedis() {

	Red = redis.NewClient(&redis.Options{
		Addr:         viper.GetString("redis.addr"),
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetInt("redis.db"),
		PoolSize:     viper.GetInt("redis.poolSize"),
		MinIdleConns: viper.GetInt("redis.minIdleConns"),
	})
	pong, err := Red.Ping(context.Background()).Result()
	if err != nil {
		fmt.Println("init redis...", err)
	} else {
		fmt.Println("config redis inited:", pong)
	}
}

const (
	PublishKey = "websocket"
)

// Publish 发布消息到 Redis
func Publish(ctx context.Context, channel string, msg string) error {
	fmt.Println("Publish to channel:", channel, "message:", msg)
	err := Red.Publish(ctx, channel, msg).Err()
	if err != nil {
		fmt.Println("Publish error:", err)
		return err
	}
	return err
}

// Subscribe 订阅 Redis 消息
func Subscribe(ctx context.Context, channel string) (string, error) {
	sub := Red.Subscribe(ctx, channel)
	fmt.Println("Subscribed ...", ctx)
	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		fmt.Println("Subscribe error:", err)
		return "", err
	}
	fmt.Println("Subscribed ...", msg.Payload)
	return msg.Payload, err
}
