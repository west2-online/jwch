/*
Copyright 2024 The west2-online Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package jwch

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

// 学生对象
type Student struct {
	ID       string         `json:"id"`       // 学号
	Password string         `json:"password"` // 密码
	cookies  []*http.Cookie // cookies中将包含session_id和其他数据
	// 如果我们使用client进行登陆的话，此时该字段失效，因为client会在登录时自动保存登陆凭证（session）
	// 所以该字段用于其他服务调用时传递登陆凭证
	Identifier string        // 位于url上id=....的一个标识符，主要用于组成url
	client     *resty.Client // Request对象
}

// 学生信息详情
type StudentDetail struct {
	Name             string `json:"name"`              // 姓名
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
	Type       string `json:"type"`       // 修读类别
	Name       string `json:"name"`       // 课程名称
	Syllabus   string `json:"syllabus"`   // 课程大纲
	LessonPlan string `json:"lessonplan"` // 课程计划
	// PaymentStatus string `json:"paymentstatus"` // 缴费状态
	Credits               string                       `json:"credit"`                // 学分
	ElectiveType          string                       `json:"electivetype"`          // 选课类型
	ExamType              string                       `json:"examtype"`              // 考试类别
	Teacher               string                       `json:"teacher"`               // 任课教师
	ScheduleRules         []CourseScheduleRule         `json:"scheduleRules"`         // 上课时间地点规则
	FullWeekScheduleRules []CourseFullWeekScheduleRule `json:"fullWeekScheduleRules"` // 整周课程上课时间地点规则
	RawScheduleRules      string                       `json:"rawScheduleRules"`      // 上课时间地点（原始文本）
	RawExamTime           string                       `json:"rawExamTime"`           // 考试时间地点（原始文本）
	RawAdjust             string                       `json:"rawAdjust"`             // 调课信息（原始文本）
	Remark                string                       `json:"remark"`                // 备注
}

// 周内课程的上课时间地点规则
type CourseScheduleRule struct {
	Location     string `json:"location"`     // 上课地点
	StartClass   int    `json:"startClass"`   // 开始节数
	EndClass     int    `json:"endClass"`     // 结束节数
	StartWeek    int    `json:"startWeek"`    // 开始周
	EndWeek      int    `json:"endWeek"`      // 结束周
	Weekday      int    `json:"weekday"`      // 星期几
	Single       bool   `json:"single"`       // 单周 (PS: 为啥不用 odd)
	Double       bool   `json:"double"`       // 双周 (PS: 为啥不用 even)
	Adjust       bool   `json:"adjust"`       // 调课
	FromFullWeek bool   `json:"fromFullWeek"` // 是否来自整周课程
}

// 整周课程的上课时间地点规则
type CourseFullWeekScheduleRule struct {
	StartWeek    int `json:"startWeek"`    // 开始周
	StartWeekDay int `json:"startWeekDay"` // 开始周在星期几
	EndWeek      int `json:"endWeek"`      // 结束周
	EndWeekDay   int `json:"endWeekDay"`   // 结束周在星期几
}

// 调课规则
type CourseAdjustRule struct {
	OldWeek       int `json:"oldWeek"`       // 原-周次
	OldWeekday    int `json:"oldWeekday"`    // 原-星期几
	OldStartClass int `json:"oldStartClass"` // 原-开始节数
	OldEndClass   int `json:"oldEndClass"`   // 原-结束节数

	NewWeek       int    `json:"newWeek"`       // 新-周次
	NewWeekday    int    `json:"newWeekday"`    // 新-星期几
	NewStartClass int    `json:"newStartClass"` // 新-开始节数
	NewEndClass   int    `json:"newEndClass"`   // 新-结束节数
	NewLocation   string `json:"newLocation"`   // 新-上课地点
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
	Campus   string `form:"campus" binding:"required"` // 校区
	Time     string `form:"time" binding:"required"`   // 日期 格式:2023-09-22
	Start    string `form:"start" binding:"required"`
	End      string `form:"end" binding:"required"`   // 查询第Start节到第End节
	Building string `form:"build" binding:"required"` // 教学楼名
}

// 校历
type SchoolCalendar struct {
	CurrentTerm string    `json:"currentTerm"` // 当前学期
	Terms       []CalTerm `json:"terms"`       // 学期信息
}

type CalTerm struct {
	TermId     string `json:"termId"`     // 学期ID
	SchoolYear string `json:"schoolYear"` // 学年
	Term       string `json:"term"`       // 学期
	StartDate  string `json:"startDate"`  // 开始日期 格式:2024-08-26
	EndDate    string `json:"endDate"`    // 结束日期 格式:2025-01-17
}

type CalTermEvents struct {
	TermId     string         `json:"termId"`     // 学期ID
	Term       string         `json:"term"`       // 学期
	SchoolYear string         `json:"schoolYear"` // 学年
	Events     []CalTermEvent `json:"events"`     // 事件
}

type CalTermEvent struct {
	Name      string `json:"name"`      // 事件名称
	StartDate string `json:"startDate"` // 开始日期 格式:2024-08-26
	EndDate   string `json:"endDate"`   // 结束日期 格式:2025-01-17
}

type CreditStatistics struct {
	Type  string // 学分类型
	Gain  string // 已获得
	Total string // 应获学分
}

type GPAData struct {
	Type  string
	Value string
}

type GPABean struct {
	Time string // 绩点计算时间
	Data []GPAData
}

type UnifiedExam struct {
	Name  string
	Score string
	Term  string
}

type ExamRoomReq struct {
	Term string
}

type ExamRoomInfo struct {
	CourseName string // 课程名称
	Credit     string // 学分
	Teacher    string // 任课教师
	Date       string // 考试日期
	Time       string // 考试时间
	Location   string // 考试地点
}

type NoticeInfo struct {
	Title string // 通知标题
	URL   string // 通知链接
	Date  string // 通知日期
}

type NoticeInfoReq struct {
	PageNum int // 获取第几页的数据，从 1 开始
}

// LocateDate 当前时间
type LocateDate struct {
	Week string
	Year string
	Term string
}
