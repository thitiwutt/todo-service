package exmain

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string
}

// real world use case
type UserHandler struct {
	db *gorm.DB
}

func (u *UserHandler) getUser(c *gin.Context) {
	var user User
	u.db.First(&user)
	c.JSON(200, user)
}

func (u *UserHandler) getListUser(c *gin.Context) {
	var users []User
	u.db.Find(&users)
	c.JSON(200, users)
}

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&User{})

	db.Create(&User{Name: "Earth"})

	userHandler := UserHandler{db: db}
	r := gin.Default()
	r.GET("/user", userHandler.getUser)
	r.GET("/users", userHandler.getListUser)

	r.Run()
}
