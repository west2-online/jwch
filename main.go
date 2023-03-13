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
	stu := jwch.NewStudent().WithUser("", "")

	// 读取本地数据
	solveErr(utils.JSONUnmarshalFromFile(localfile, &res))
	stu.SetCookies(res.Cookies)
	stu.WithSession(res.Session)

	// 登录账号
	if stu.CheckSession() != nil {
		log.Println("session expire, relogin")
		solveErr(stu.Login()) // 登录
		log.Println("check session")
		solveErr(stu.CheckSession()) // 检查session
		stu.SaveLoginData(localfile) // 保存到本地文件
	}

	// 需要先获取学期列表才能选择我的选课
	// 对于客户端，可以GetTerms后存在本地，这样可以直接进行获取课程的请求
	solveErr(stu.GetTerms())

	// 获取最新一学期的课程，按照terms排列可以获取各个学期的
	list, err := stu.GetSelectedCourse(stu.Terms[0])
	solveErr(err)

	// 输出课程数量
	fmt.Println("course len:", len(list))

	// 输出课程信息
	for _, v := range list {
		fmt.Println(utils.PrintStruct(v))
	}

}

func solveErr(err error) {
	if err != nil {
		panic(err)
	}
}
