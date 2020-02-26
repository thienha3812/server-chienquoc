package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

//C:\Users\Administrator\Desktop\currentnew\taikhoan.ini
var PATH = `/Users/macintoshhd/Documents/MyProject/golang/example.ini` // PATH

// Ghi tài khoản và mật khẩu vào file ini
func writeIniFile(username string, password string, email string) string {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	// regexp kiểm tra ký tự đặc biệt
	var re = regexp.MustCompile(`(?m)[!@#$%^&*(),.?":{}|<>]`)
	regexUsername := re.FindAllIndex([]byte(username), -1)
	regexPassword := re.FindAllIndex([]byte(password), -1)
	if len(regexUsername) > 0 || len(regexPassword) > 0 {
		return "Tài khoản hoặc mật khẩu không được có ký tự đặc biệt!"
	}
	// reexp kiểm tra khoảng cách
	var re1 = regexp.MustCompile(`\s`)
	regex1Password := re1.FindAllIndex([]byte(password), -1)
	regex1Username := re1.FindAllIndex([]byte(username), -1)
	if len(regex1Username) > 0 || len(regex1Password) > 0 {
		return "Tài khoản hoặc mật khẩu không được chứa khoảng trắng!"
	}
	var re2 = regexp.MustCompile(`^[a-z0-9]+$`)
	if re2.MatchString(username) == false || re2.MatchString(password) == false {
		return "Tài khoản và mật khẩu không được có dấu và không được viết hoa!"
	}
	file, err := os.OpenFile(PATH, os.O_APPEND|os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	status := checkUserExist(username, email)
	if len(username) < 3 || len(password) < 6 {
		return "Tài khoản độ dài từ 3 và mật khẩu từ 6 trở lên!"
	}
	if status == 1 {
		return "Tài khoản đã tồn tại!"
	} else if status == 2 {
		return "Email đã tồn tại"
	} else {
		writer.WriteString(username + " " + password + " " + email + "\n")

		return "Tài khoản đăng ký thành công"
	}
}

// Kiểm tra tài khoản có tồn tại ko
func checkUserExist(username string, email string) int {
	file, err := os.OpenFile(PATH, os.O_RDONLY, 0666)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewScanner(file)
	for reader.Scan() {
		s := strings.Fields(reader.Text())
		if s[0] == username {
			return 1 // Tài khoản đã tồn tại
		}
		if s[2] == email {
			return 2 //  Emai đã tồn tại
		}
	}
	return 3 // Code đăng ký thành công
}

// Check mật khẩu và tài khoản đúng
func checkUserAndPasswordCorrect(username string, password string) int {
	file, err := os.OpenFile(PATH, os.O_RDONLY, 0666)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewScanner(file)
	for reader.Scan() {
		s := strings.Fields(reader.Text())
		if s[0] == username && s[1] == password {
			return 1 //Code thành công
		}
	}
	return 0 // Tài khoản hoặc mật khẩu không đúng
}

// controller
//Đăng ký
func registerAccount(c echo.Context) error {
	m := echo.Map{}
	println(m)
	if err := c.Bind(&m); err != nil {
		return err
	}
	textStatus := writeIniFile(m["username"].(string), m["password"].(string), m["email"].(string))
	return c.String(200, textStatus)
}

// Đăng nhập
func loginAccount(c echo.Context) error {
	m := echo.Map{}
	if err := c.Bind(&m); err != nil {
		return err
	}
	codeStatus := checkUserAndPasswordCorrect(m["username"].(string), m["password"].(string))
	if codeStatus == 1 {
		token := jwt.New(jwt.SigningMethodHS256)

		// Set claims
		claims := token.Claims.(jwt.MapClaims)
		claims["username"] = m["username"]
		claims["password"] = m["password"]
		claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

		// Generate encoded token and send it as response.
		t, err := token.SignedString([]byte("secret"))
		if err != nil {
			return err
		}
		return c.JSON(200, map[string]string{
			"token": t,
		})
	}
	return c.String(403, "Lỗi khi request")
}

// Đổi mật khẩu
func changepasswordAccount(c echo.Context) error {
	file, err := os.OpenFile("example.ini", os.O_RDWR|os.O_APPEND, 0644)

	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	m := echo.Map{}
	if err := c.Bind(&m); err != nil {
		return err
	}
	contents, _ := ioutil.ReadAll(reader)
	body := fmt.Sprintf(`(?m)%s %s`, m["username"].(string), m["password"].(string))
	var re = regexp.MustCompile(body)
	s := re.ReplaceAll(contents, []byte(m["username"].(string)+" "+m["newpassword"].(string)))
	file1, _ := os.Create("example.ini")
	file1.Write([]byte(s))
	file.Sync()
	return c.String(200, "1")
}

//Quên mật khẩu
func forgotpasswordAccount(c echo.Context) error {
	m := echo.Map{}
	if err := c.Bind(&m); err != nil {
		log.Fatal(err)
	}
	from := "....email"
	pass := "...password"
	to := m["email"].(string)
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Hỗ trợ tìm lại mật khẩu \n\n" +
		"Xin chào bạn, cám ơn bạn đã ủng hộ game của chúng tôi \n" +
		"Dưới đây là thông tin tài khoản của bạn sau khi nhận được bạn vui lòng đăng nhập để kiểm tra lại nhé \n"
	file, _ := os.OpenFile(PATH, os.O_RDONLY, 0666)
	reader := bufio.NewScanner(file)
	var finded bool
	finded = false
	for reader.Scan() {
		s := strings.Fields(reader.Text())
		if s[2] == m["email"].(string) {
			finded = true
			err1 := smtp.SendMail("smtp.gmail.com:587",
				smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
				from, []string{to}, []byte(msg+"Tài khoản "+s[0]+" có mật khẩu là "+s[1]))
			if err1 != nil {
				c.String(200, "Xảy ra lỗi trên hệ thống.")
			}
		}
	}
	if finded == false {
		return c.String(200, "Email chưa được đăng ký trên hệ thống.")
	}
	return c.String(200, "Mật khẩu đã được gửi về email đăng ký.")
}

// Decode token
func restricted(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["username"].(string)
	return c.String(http.StatusOK, "welcom"+name)
}

//
func main() {
	e := echo.New()
	// Middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://103.127.206.242", "http://chienquoc.online"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//
	r := e.Group("/restricted")
	r.Use(middleware.JWT([]byte("secret")))
	r.GET("", restricted)
	// Routes
	e.POST("/user/register", registerAccount)
	e.POST("/user/login", loginAccount)
	e.POST("/user/changepassword", changepasswordAccount,)
	e.POST("/user/forgotpassword", forgotpasswordAccount)
	e.Logger.Fatal(e.Start(":5000"))
}
