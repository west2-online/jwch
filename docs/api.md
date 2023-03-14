# Overview

```go
// Init
func NewStudent() *Student {}
func (s *Student) WithUser(id, password string) *Student {}
func (s *Student) WithSession(session string) *Student {}

// LoginData
func (s *Student) SaveLoginData(filePath string) error {}
func (s *Student) GetLoginDataJSON() string {}
func (s *Student) ClearLoginData() {}
func (s *Student) SetLoginData(data LoginData) {}
func (s *Student) AppendCookies(cookies []*http.Cookie) {}

// Internal Request
func (s *Student) NewRequest() *resty.Request {}
func (s *Student) GetWithSession(url string) (*html.Node, error) {}
func (s *Student) GetWithSessionRaw(url string) (*resty.Response, error) {}
func (s *Student) PostWithSession(url string, formdata map[string]string) (*html.Node, error) {}

// User
func (s *Student) Login() error {}
func (s *Student) CheckSession() error {}
func (s *Student) GetInfo() (resp *StudentDetail, err error) {}

// Course
func (s *Student) GetTerms() (*Term, error) {}
func (s *Student) GetSemesterCourses(term, viewState, eventValidation string) ([]*Course, error)

// Mark
func (s *Student) GetMarks() (resp []*Mark, err error) {}
func (s *Student) GetCET() error {}
```

# Usage example

```go
package main

import (
	"fmt"
	"jwch/jwch"
	"jwch/utils"
	"log"
)

var res jwch.LoginData
var localfile string = "logindata.txt"


func main() {
	// 创建学生对象
	stu := jwch.NewStudent().WithUser("id", "password")

	// 读取本地数据
	solveErr(utils.JSONUnmarshalFromFile(localfile, &res))
	stu.SetLoginData(res)

	// 登录账号
	err := stu.CheckSession()
	if err != nil {
		log.Println(err.Error())    // 输出错误
		solveErr(stu.Login())        // 登录
		solveErr(stu.CheckSession()) // 检查session
		stu.SaveLoginData(localfile) // 保存到本地文件
	}

	// 需要先获取学期列表才能选择我的选课
	// 对于客户端，可以GetTerms后存在本地，这样可以直接进行获取课程的请求
	term, err := stu.GetTerms()
	solveErr(err)

	// 获取最新一学期的课程，按照terms排列可以获取各个学期的
	list, err := stu.GetSemesterCourses(term.Terms[0], term.ViewState, term.EventValidation)
	solveErr(err)

	// 输出课程数量与信息
	fmt.Println("course num:", len(list))
	for _, v := range list {
		fmt.Println(utils.PrintStruct(v))
	}

	// 获取个人信息
	detail, err := stu.GetInfo()
	solveErr(err)
	fmt.Println(utils.PrintStruct(detail))

	// 获取成绩
	marks, err := stu.GetMarks()
	solveErr(err)
	fmt.Println(utils.PrintStruct(marks))

}

func solveErr(err error) {
	if err != nil {
		panic(err)
	}
}
```
