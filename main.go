package main

import (
	"fmt"
	"test/jwch"
	"test/utils"
)

func main() {
	// 创建学生对象
	stu := jwch.NewStudent().WithUser("username", "passwd")

	// 检查Session活性
	err := stu.CheckSession()

	// 登录账号
	if err != nil {
		fmt.Println("session expired, login...")
		solveErr(stu.Login())
		fmt.Println("session: ", stu.Session)
		// TODO: 加一个设置Cookie到本地的操作，否则光靠session是不够的
	}

	// 需要先获取学期列表才能选择我的选课
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
