package todo

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Todo struct {
	Title string `json:"text" binding:"required"`
	gorm.Model
}

func (Todo) TableName() string {
	return "todos"
}

// handler type
type TodoHandler struct {
	db *gorm.DB
}

// init handler
func NewTodoHandler(db *gorm.DB) *TodoHandler {
	return &TodoHandler{db: db}
}

func (t *TodoHandler) NewTask(context *gin.Context) {
	var todo Todo
	// .BindJSON (must bind) forces error with status 400
	// .ShouldBindJSON manual error
	if err := context.ShouldBindJSON(&todo); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}

	// Create inserts value, returning the inserted data's primary key in value's id
	r := t.db.Create(&todo)
	if err := r.Error; err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}

	context.JSON(http.StatusCreated, gin.H{
		"ID":   todo.Model.ID,
		"text": todo.Title,
	})
}
