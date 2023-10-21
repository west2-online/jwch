package jwch

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

// 本地数据
type LoginData struct {
	Cookies []*http.Cookie `json:"cookies"`
	Session string         `json:"session"`
}

// 学生对象
type Student struct {
	ID        string        `json:"id"`         // 学号
	Password  string        `json:"password"`   // 密码
	LoginData LoginData     `json:"login_data"` // 登录凭证
	client    *resty.Client // Request对象
}

// 学生信息详情
type StudentDetail struct {
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
	PoliticalStatus  string `json:"political_status"`  // 政治面貌
	Source           string `json:"source"`            // 生源地
}

// 学期信息
type Term struct {
	Terms           []string `json:"terms"`           // 学期数量
	ViewState       string   `json:"viewstate"`       // 课表必要信息
	EventValidation string   `json:"eventvalidation"` // 课表必要信息
}

// 课程信息
type Course struct {
	Type          string `json:"type"`          // 修读类别
	Name          string `json:"name"`          // 课程名称
	PaymentStatus string `json:"paymentstatus"` // 缴费状态
	Syllabus      string `json:"syllabus"`      // 课程大纲
	LessonPlan    string `json:"lessonplan"`    // 课程计划
	Credits       string `json:"credit"`        // 学分
	ElectiveType  string `json:"electivetype"`  // 选课类型
	ExamType      string `json:"examtype"`      // 考试类别
	Teacher       string `json:"teacher"`       // 任课教师
	Classroom     string `json:"classroom"`     // 上课时间地点
	ExamTime      string `json:"examtime"`      // 考试时间地点
	Remark        string `json:"remark"`        // 备注
	Adjust        string `json:"adjust"`        // 调课信息
}

type Mark struct {
	Type          string `json:"type"`           // 修读类别
	Semester      string `json:"semester"`       // 开课学期
	Name          string `json:"name"`           // 课程名称
	Credits       string `json:"credit"`         // 计划学分
	Score         string `json:"score"`          // 得分
	GPA           string `json:"GPA"`            // 绩点
	EarnedCredits string `json:"earned_credits"` // 得到学分
	ElectiveType  string `json:"electivetype"`   // 选课类型
	ExamType      string `json:"examtype"`       // 考试类别
	Teacher       string `json:"teacher"`        // 任课教师
	Classroom     string `json:"classroom"`      // 上课时间地点
	ExamTime      string `json:"examtime"`       // 考试时间地点
}

// 空教室请求
type EmptyRoomReq struct {
	Time     string `form:"time" binding:"required"` // 日期 格式:2023-09-22
	Start    string `form:"start" binding:"required"`
	End      string `form:"end" binding:"required"`   // 查询第Start节到第End节
	Building string `form:"build" binding:"required"` // 教学楼名
}
