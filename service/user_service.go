package service

import (
	"fmt"
	"ginchat/models"
	"ginchat/utils"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// GetUserList
// @Summary 用户列表
// @Tags 用户模块
// @Success 200 {string} json{"code", "message"}
// @Router /user/getUserList [get]
func GetUserList(c *gin.Context) {
	data := make([]*models.UserBasic, 10)
	data = models.GetUserList()

	c.JSON(http.StatusOK, gin.H{
		"code":    0, // 0表示成功，-1表示失败
		"message": "获取用户列表成功!",
		"data":    data,
	})
}

// CreateUser
// @Summary 新增用户
// @Tags 用户模块
// @param name query string false "用户名"
// @param password query string false "密码"
// @param repassword query string false "确认密码"
// @Success 200 {string} json{"code", "message"}
// @Router /user/createUser [get]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{}
	user.Name = c.Query("name")
	password := c.Query("password")
	repassword := c.Query("repassword")

	// 密码加密
	salt := fmt.Sprintf("%06d", rand.Int31())

	// 注册用户名检测
	data := models.FindByName(user.Name)
	if data.Name != "" {
		c.JSON(-1, gin.H{
			"code":    -1, // 0表示成功，-1表示失败
			"message": "用户名已存在，请重新输入!",
			"data":    user,
		})
		return
	}
	if password != repassword {
		c.JSON(-1, gin.H{
			"code":    -1, // 0表示成功，-1表示失败
			"message": "两次密码不一致，请重新输入!",
			"data":    user,
		})
		return
	}
	// user.Password = password
	user.Password = utils.MakePassword(password, salt)
	user.Salt = salt
	models.CreateUser(user)
	c.JSON(http.StatusOK, gin.H{
		"code":    0, // 0表示成功，-1表示失败
		"message": "新增用户成功!",
		"data":    user,
	})
}

// DeleteUser
// @Summary 删除用户
// @Tags 用户模块
// @param id query string false "id"
// @Success 200 {string} json{"code", "message"}
// @Router /user/deleteUser [get]
func DeleteUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.Query("id"))
	user.ID = uint(id)

	models.DeleteUser(user)
	c.JSON(http.StatusOK, gin.H{
		"code":    0, // 0表示成功，-1表示失败
		"message": "删除用户成功!",
		"data":    user,
	})
}

// UpdateUser
// @Summary 修改用户
// @Tags 用户模块
// @param id formData string false "id"
// @param name formData string false "name"
// @param password formData string false "password"
// @param phone formData string false "phone"
// @param email formData string false "email"
// @Success 200 {string} json{"code", "message"}
// @Router /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.PostForm("id"))
	user.ID = uint(id)
	user.Name = c.PostForm("name")
	user.Password = c.PostForm("password")
	user.Phone = c.PostForm("phone")
	user.Email = c.PostForm("email")

	_, err := govalidator.ValidateStruct(user)
	if err != nil {
		fmt.Println(err)
		c.JSON(200, gin.H{
			"code":    -1, // 0表示成功，-1表示失败
			"message": "修改参数不匹配!",
			"data":    user,
		})
	} else {
		models.UpdateUser(user)
		c.JSON(http.StatusOK, gin.H{
			"code":    0, // 0表示成功，-1表示失败
			"message": "修改用户成功!",
			"data":    user,
		})
	}
}

// FindByNameAndPwd
// @Summary 根据用户名和密码登录
// @Tags 用户模块
// @param name formData string false "用户名"
// @param password formData string false "密码"
// @Success 200 {string} json{"code", "message"}
// @Router /user/findByNameAndPwd [post]
func FindByNameAndPwd(c *gin.Context) {
	data := models.UserBasic{}

	name := c.PostForm("name")
	password := c.PostForm("password")
	user := models.FindByName(name)
	if user.Name == "" {
		c.JSON(-1, gin.H{
			"code":    -1, // 0表示成功，-1表示失败
			"message": "用户名不存在，请重新输入!",
			"data":    data,
		})
		return
	}

	flag := utils.ValidPassword(password, user.Salt, user.Password)
	if !flag {
		c.JSON(-1, gin.H{
			"code":    -1, // 0表示成功，-1表示失败
			"message": "密码错误，请重新输入!",
			"data":    data,
		})
		return
	}
	// pwd := utils.MakePassword(password, user.Salt)
	data = models.FindByNameAndPwd(name, user.Password)

	c.JSON(http.StatusOK, gin.H{
		"code":    0, // 0表示成功，-1表示失败
		"message": "登录成功!",
		"data":    data,
	})
}

// 防止跨域站点伪造请求
var upGrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SendMsg(c *gin.Context) {
	ws, err := upGrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(ws *websocket.Conn) {
		err := ws.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(ws)
	MsgHandler(ws, c)
}

func MsgHandler(ws *websocket.Conn, c *gin.Context) {
	for {
		msg, err := utils.Subscribe(c, utils.PublishKey)
		if err != nil {
			fmt.Println(err)
			return
		}
		tm := time.Now().Format("2006-01-02 15:04:05")
		m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
		err = ws.WriteMessage(1, []byte(m))
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func SendUserMsg(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}
