package jwch

import (
	"github.com/go-resty/resty/v2"
)

// 学生对象
type Student struct {
	ID              string        `json:"id"`              // 学号
	Password        string        `json:"password"`        // 密码
	Session         string        `json:"session"`         // 会话ID
	Terms           []string      `json:"terms"`           // 学期
	ViewState       string        `json:"viewstate"`       // 课表必要信息
	EventValidation string        `json:"eventvalidation"` // 课表必要信息
	client          *resty.Client // Request 对象
	// c        *colly.Collector // Colly 对象
}

// SSO登录返回
type SSOLoginResponse struct {
	Code int    `json:"code"` // 状态码
	Info string `json:"info"` // 返回消息
	// Data interface{} `json:"data"`
}

// 课程信息
type Course struct {
	Type          string `json:"type"`          // 修读类别
	Name          string `json:"name"`          // 课程名称
	PaymentStatus string `json:"paymentstatus"` // 缴费状态
	Syllabus      string `json:"syllabus"`      // 课程大纲
	LessonPlan    string `json:"lessonplan"`    // 课程计划
	Credit        string `json:"credit"`        // 学分
	ElectiveType  string `json:"electivetype"`  // 选课类型
	ExamType      string `json:"examtype"`      // 考试类别
	Teacher       string `json:"teacher"`       // 任课教师
	Classroom     string `json:"classroom"`     // 上课时间地点
	ExamTime      string `json:"examtime"`      // 考试时间地点
	Remark        string `json:"remark"`        // 备注
	AdjustInfo    string `json:"adjustinfo"`    // 调课信息
}
