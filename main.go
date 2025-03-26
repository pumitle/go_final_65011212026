package main

import (
	"fmt"
	"go-final/controller"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	fmt.Println(viper.Get("mysql.dsn"))
	dsn := viper.GetString("mysql.dsn")

	dialactor := mysql.Open(dsn)
	db, err := gorm.Open(dialactor, &gorm.Config{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Connection successful")
	DB = db
	controller.SetDB(DB)

	controller.StartServer()
}
