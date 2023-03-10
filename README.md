# JWCH

这个项目是教务处类的半成品，欢迎大家**提出疑问/todolist**

# How to use

We should clone this repo

```bash
❯ git clone https://github.com/fzuhelper/jwch
```

Then we just need to modify **main.go** to test any func.

```go
func main() {
	// 创建学生对象
	stu := jwch.NewStudent().WithUser("username", "passwd") // 更改为学号与密码

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
```

Finally we can use `go run main.go` to test.

# Current Progress

- [x] User login
- [x] Get course selections for each semester 
- [x] Set any apis but not implement
- [ ] Complete all apis
- [ ] Benchmark test
- [ ] Bug check & fix
- [ ] ...

# FileTree

```
.
├── README.md			// 文档
├── cookies.txt
├── errno				// 错误处理
│   ├── code.go
│   ├── default.go
│   └── errno.go
├── go.mod
├── go.sum
├── jwch				// 教务处类
│   ├── course.go		// 课程
│   ├── jwch.go			// 类主函数
│   ├── mark.go			// 成绩
│   ├── model.go		// 自定义结构体
│   ├── user.go			// 用户
│   └── xpath.go		// xpath优化函数
├── main.go
├── main.py				// python脚本
├── photo.jpg			// 验证码图片(python use)
├── test.html			// 测试html
├── test.png			// 验证码图片
└── utils				// 通用函数
    └── utils.go
```

