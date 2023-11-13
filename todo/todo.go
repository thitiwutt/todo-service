package todo

import (
	"net/http"
	"strconv"

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

func (t *TodoHandler) NewTask(ctx *gin.Context) {
	var todo Todo
	// .BindJSON (must bind) forces error with status 400
	// .ShouldBindJSON manual error
	if err := ctx.ShouldBindJSON(&todo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}

	// Create inserts value, returning the inserted data's primary key in value's id
	r := t.db.Create(&todo)
	if err := r.Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"ID":   todo.Model.ID,
		"text": todo.Title,
	})
}

func (t *TodoHandler) ListTodos(ctx *gin.Context) {
	var todos []Todo
	r := t.db.Find(&todos)
	if err := r.Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, todos)
}

func (t *TodoHandler) DeleteTodo(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	r := t.db.Delete(&Todo{}, id)
	if err := r.Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"ID":     id,
		"status": "success",
	})
}
