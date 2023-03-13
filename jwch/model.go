package jwch

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

// 本地数据
type LoginData struct {
	Cookies []*http.Cookie
	Session string
}

// 学生对象
type Student struct {
	ID              string         `json:"id"`              // 学号
	Password        string         `json:"password"`        // 密码
	Session         string         `json:"session"`         // 会话ID
	Terms           []string       `json:"terms"`           // 学期数量
	ViewState       string         `json:"viewstate"`       // 课表必要信息
	EventValidation string         `json:"eventvalidation"` // 课表必要信息
	client          *resty.Client  // Request 对象
	Cookies         []*http.Cookie // Cookies
	// c        *colly.Collector // Colly 对象
}

// 学生信息详情
type StudentDetail struct {
	Nickname         string `json:"nickname"`          // 昵称
	Signature        string `json:"signature"`         // 签名
	Sex              string `json:"sex"`               // 性别
	Birthday         string `json:"birthday"`          // 出生日期
	Phone            string `json:"phont"`             // 手机号
	Email            string `json:"email"`             // 邮箱
	College          string `json:"college"`           // 学院
	Grade            string `json:"grade"`             // 年级
	StatusChanges    string `json:"status_change"`     // 学籍异动与奖励
	Major            string `json:"major"`             // 专业
	Counselor        string `json:"counselor"`         // 辅导员
	ExamineeCategory string `json:"examinee_category"` // 考生类别
	Nationality      string `json:"nationality"`       // 民族
	Country          string `json:"country"`           // 国别
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
