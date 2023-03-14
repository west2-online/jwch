package main

import (
	"fmt"
	"jwch/jwch"
	"jwch/utils"
	"log"

	"github.com/otiai10/gosseract/v2"
)

var res jwch.LoginData
var localfile string = "logindata.txt"

// test Image
func test() {
	client := gosseract.NewClient()
	defer client.Close()

	// client.SetImage("./photo.jpg")
	// client.SetLanguage("chi_sim")
	// text, err := client.Text()

	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(text)
}

func main() {
	test()
	// 创建学生对象
	stu := jwch.NewStudent().WithUser("", "")

	// 读取本地数据
	solveErr(utils.JSONUnmarshalFromFile(localfile, &res))
	stu.SetLoginData(res)

	// 登录账号
	err := stu.CheckSession()
	if err != nil {
		log.Println(err.Error())
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
