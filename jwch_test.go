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
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/west2-online/jwch/constants"
	"github.com/west2-online/jwch/utils"
)

var (
	username = os.Getenv("JWCH_USERNAME") // 学号
	password = os.Getenv("JWCH_PASSWORD") // 密码
	// localfile = "logindata.txt"
)

var (
	islogin = false
	stu     = NewStudent().WithUser(username, password)
)

func login() error {
	err := stu.Login()
	if err != nil {
		return err
	}

	err = stu.CheckSession()
	if err != nil {
		return err
	}

	islogin = true
	return nil
}

func TestMain(m *testing.M) {
	err := login()
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		os.Exit(1)
	}

	// 运行测试
	code := m.Run()

	// 在所有测试结束后执行清理
	os.Exit(code)
}

func Test_GetValidateCode(t *testing.T) {
	// 获取验证码图片
	s := NewStudent()
	resp, err := s.NewRequest().Get(constants.VerifyCodeURL)
	if err != nil {
		t.Error(err)
	}
	code, err := GetValidateCode(utils.Base64EncodeHTTPImage(resp.Body()))
	if err != nil {
		t.Error(err)
	}
	fmt.Println(code)
}

func Test_GetIdentifierAndCookies(t *testing.T) {
	_, _, err := stu.GetIdentifierAndCookies()
	if err != nil {
		t.Error(err)
	}
}

func Test_GetCourse(t *testing.T) {
	terms, err := stu.GetTerms()
	if err != nil {
		t.Error(err)
	}

	list, err := stu.GetSemesterCourses(terms.Terms[0], terms.ViewState, terms.EventValidation)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("course num:", len(list))
	// 不允许输出具体课程
	// fmt.Println(utils.PrintStruct(list))
}

func Test_GetInfo(t *testing.T) {
	_, err := stu.GetInfo()
	if err != nil {
		t.Error(err)
	}

	// 不允许输出个人信息
}

func Test_GetMarks(t *testing.T) {
	_, err := stu.GetMarks()
	if err != nil {
		t.Error(err)
	}

	// 不允许输出成绩
}

// 使用并发后似乎快了1s
func Test_GetQiShanEmptyRoom(t *testing.T) {
	rooms, err := stu.GetQiShanEmptyRoom(EmptyRoomReq{
		Campus: "旗山校区",
		Time:   "2024-09-26",
		Start:  "1",
		End:    "8",
	})
	if err != nil {
		t.Error(err)
	}

	// 此处可以输出空教室信息
	fmt.Println(utils.PrintStruct(rooms))
}

func Test_GetJinJiangEmptyRoom(t *testing.T) {
	rooms, err := stu.GetEmptyRoom(EmptyRoomReq{
		Campus: "晋江校区",
		Time:   "2024-09-19",
		Start:  "1",
		End:    "2",
	})
	if err != nil {
		t.Error(err)
	}

	// 此处可以输出空教室信息
	fmt.Println(utils.PrintStruct(rooms))
}

func Test_GetTongPanEmptyRoom(t *testing.T) {
	rooms, err := stu.GetEmptyRoom(EmptyRoomReq{
		Campus: "铜盘校区",
		Time:   "2024-09-19",
		Start:  "1",
		End:    "2",
	})
	if err != nil {
		t.Error(err)
	}

	// 此处可以输出空教室信息
	fmt.Println(utils.PrintStruct(rooms))
}

func Test_GetQuanGangEmptyRoom(t *testing.T) {
	rooms, err := stu.GetEmptyRoom(EmptyRoomReq{
		Campus: "泉港校区",
		Time:   "2024-09-19",
		Start:  "1",
		End:    "2",
	})
	if err != nil {
		t.Error(err)
	}

	// 此处可以输出空教室信息
	fmt.Println(utils.PrintStruct(rooms))
}

func Test_GetYiShanEmptyRoom(t *testing.T) {
	rooms, err := stu.GetEmptyRoom(EmptyRoomReq{
		Campus: "怡山校区",
		Time:   "2024-09-19",
		Start:  "1",
		End:    "2",
	})
	if err != nil {
		t.Error(err)
	}

	// 此处可以输出空教室信息
	fmt.Println(utils.PrintStruct(rooms))
}

func Test_GetXiaMenEmptyRoom(t *testing.T) {
	rooms, err := stu.GetEmptyRoom(EmptyRoomReq{
		Campus: "厦门工艺美院",
		Time:   "2024-09-19",
		Start:  "1",
		End:    "2",
	})
	if err != nil {
		t.Error(err)
	}

	// 此处可以输出空教室信息
	fmt.Println(utils.PrintStruct(rooms))
}

func Test_GetSchoolCalendar(t *testing.T) {
	calendar, err := stu.GetSchoolCalendar()
	if err != nil {
		t.Error(err)
	}

	// 此处可以输出校历信息
	fmt.Println(utils.PrintStruct(calendar))
}

func Test_GetTermEvents(t *testing.T) {
	calendar, err := stu.GetSchoolCalendar()
	if err != nil {
		t.Error(err)
	}

	events, err := stu.GetTermEvents(calendar.Terms[0].TermId)
	if err != nil {
		t.Error(err)
	}

	// 此处可以输出学期信息
	fmt.Println(utils.PrintStruct(events))
}

func Test_GetCredit(t *testing.T) {
	_, err := stu.GetCredit()
	if err != nil {
		t.Error(err)
	}

	// 不允许输出学分信息
}

func Test_GetGPA(t *testing.T) {
	_, err := stu.GetGPA()
	if err != nil {
		t.Error(err)
	}

	// 不允许输出 GPA 信息
}

func TestGetUnifiedExam(t *testing.T) {
	_, err := stu.GetCET()
	if err != nil {
		t.Error(err)
	}

	_, err = stu.GetJS()
	if err != nil {
		t.Error(err)
	}

	// 不允许输出考试成绩信息
}

// 考场信息
func TestGetExamRoomInfo(t *testing.T) {
	_, err := stu.GetExamRoom(ExamRoomReq{
		Term: "202401",
	})
	if err != nil {
		t.Error(err)
	}
}

func TestGetNoticesInfo(t *testing.T) {
	content, totalPages, err := stu.GetNoticeInfo(&NoticeInfoReq{PageNum: 2})
	fmt.Println(totalPages)
	if err != nil {
		t.Error(err)
	}
	if content == nil || totalPages == 0 {
		t.Error("content is nil")
	}
}

func TestGetCultivatePlan(t *testing.T) {
	url, err := stu.GetCultivatePlan()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(url)
}

func TestGetLocateDate(t *testing.T) {
	date, err := stu.GetLocateDate()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(utils.PrintStruct(date))
}

func TestGetLectures(t *testing.T) {
	_, err := stu.GetLectures()
	if err != nil {
		t.Error(err)
	}
}

func TestApplyAdjustRules(t *testing.T) {
	cases := []struct {
		name     string
		rules    []CourseScheduleRule
		adjusts  []CourseAdjustRule
		expected []CourseScheduleRule
	}{
		{
			name: "NoAdjust",
			rules: []CourseScheduleRule{
				{Location: "旗山西1-206", StartClass: 3, EndClass: 4, StartWeek: 1, EndWeek: 16, Weekday: 1, Single: true, Double: true},
			},
			adjusts: nil,
			expected: []CourseScheduleRule{
				{Location: "旗山西1-206", StartClass: 3, EndClass: 4, StartWeek: 1, EndWeek: 16, Weekday: 1, Single: true, Double: true},
			},
		},
		{
			name: "EmptyAdjustRules",
			rules: []CourseScheduleRule{
				{Location: "旗山西1-206", StartClass: 3, EndClass: 4, StartWeek: 1, EndWeek: 16, Weekday: 1, Single: true, Double: true},
			},
			adjusts: []CourseAdjustRule{},
			expected: []CourseScheduleRule{
				{Location: "旗山西1-206", StartClass: 3, EndClass: 4, StartWeek: 1, EndWeek: 16, Weekday: 1, Single: true, Double: true},
			},
		},
		{
			name: "SingleAdjust",
			rules: []CourseScheduleRule{
				{Location: "铜盘A110", StartClass: 5, EndClass: 6, StartWeek: 5, EndWeek: 18, Weekday: 3, Single: true, Double: true},
			},
			adjusts: []CourseAdjustRule{
				{OldWeek: 6, OldWeekday: 3, OldStartClass: 5, OldEndClass: 6, NewWeek: 9, NewWeekday: 1, NewStartClass: 7, NewEndClass: 8, NewLocation: "旗山西1-206"},
			},
			expected: []CourseScheduleRule{
				{Location: "旗山西1-206", StartClass: 7, EndClass: 8, StartWeek: 9, EndWeek: 9, Weekday: 1, Single: true, Double: true, Adjust: true},
				{Location: "铜盘A110", StartClass: 5, EndClass: 6, StartWeek: 5, EndWeek: 5, Weekday: 3, Single: true, Double: true},
				{Location: "铜盘A110", StartClass: 5, EndClass: 6, StartWeek: 7, EndWeek: 18, Weekday: 3, Single: true, Double: true},
			},
		},
		{
			name: "AdjustFirstWeek",
			rules: []CourseScheduleRule{
				{Location: "铜盘A110", StartClass: 3, EndClass: 4, StartWeek: 1, EndWeek: 8, Weekday: 2, Single: true, Double: true},
			},
			adjusts: []CourseAdjustRule{
				{OldWeek: 1, OldWeekday: 2, OldStartClass: 3, OldEndClass: 4, NewWeek: 10, NewWeekday: 5, NewStartClass: 3, NewEndClass: 4, NewLocation: "旗山东3-101"},
			},
			expected: []CourseScheduleRule{
				{Location: "旗山东3-101", StartClass: 3, EndClass: 4, StartWeek: 10, EndWeek: 10, Weekday: 5, Single: true, Double: true, Adjust: true},
				{Location: "铜盘A110", StartClass: 3, EndClass: 4, StartWeek: 2, EndWeek: 8, Weekday: 2, Single: true, Double: true},
			},
		},
		{
			name: "AdjustLastWeek",
			rules: []CourseScheduleRule{
				{Location: "铜盘A110", StartClass: 1, EndClass: 2, StartWeek: 5, EndWeek: 10, Weekday: 4, Single: true, Double: true},
			},
			adjusts: []CourseAdjustRule{
				{OldWeek: 10, OldWeekday: 4, OldStartClass: 1, OldEndClass: 2, NewWeek: 12, NewWeekday: 3, NewStartClass: 1, NewEndClass: 2, NewLocation: "旗山西1-206"},
			},
			expected: []CourseScheduleRule{
				{Location: "旗山西1-206", StartClass: 1, EndClass: 2, StartWeek: 12, EndWeek: 12, Weekday: 3, Single: true, Double: true, Adjust: true},
				{Location: "铜盘A110", StartClass: 1, EndClass: 2, StartWeek: 5, EndWeek: 9, Weekday: 4, Single: true, Double: true},
			},
		},
		{
			name: "MultipleAdjusts",
			rules: []CourseScheduleRule{
				{Location: "铜盘A110", StartClass: 5, EndClass: 6, StartWeek: 5, EndWeek: 18, Weekday: 3, Single: true, Double: true},
			},
			adjusts: []CourseAdjustRule{
				{OldWeek: 6, OldWeekday: 3, OldStartClass: 5, OldEndClass: 6, NewWeek: 9, NewWeekday: 1, NewStartClass: 7, NewEndClass: 8, NewLocation: "旗山西1-206"},
				{OldWeek: 10, OldWeekday: 3, OldStartClass: 5, OldEndClass: 6, NewWeek: 12, NewWeekday: 2, NewStartClass: 5, NewEndClass: 6, NewLocation: "旗山东3-101"},
			},
			expected: []CourseScheduleRule{
				{Location: "旗山西1-206", StartClass: 7, EndClass: 8, StartWeek: 9, EndWeek: 9, Weekday: 1, Single: true, Double: true, Adjust: true},
				{Location: "旗山东3-101", StartClass: 5, EndClass: 6, StartWeek: 12, EndWeek: 12, Weekday: 2, Single: true, Double: true, Adjust: true},
				{Location: "铜盘A110", StartClass: 5, EndClass: 6, StartWeek: 5, EndWeek: 5, Weekday: 3, Single: true, Double: true},
				{Location: "铜盘A110", StartClass: 5, EndClass: 6, StartWeek: 7, EndWeek: 9, Weekday: 3, Single: true, Double: true},
				{Location: "铜盘A110", StartClass: 5, EndClass: 6, StartWeek: 11, EndWeek: 18, Weekday: 3, Single: true, Double: true},
			},
		},
		{
			name: "NoMatchingAdjust",
			rules: []CourseScheduleRule{
				{Location: "铜盘A110", StartClass: 3, EndClass: 4, StartWeek: 1, EndWeek: 16, Weekday: 1, Single: true, Double: true},
			},
			adjusts: []CourseAdjustRule{
				{OldWeek: 6, OldWeekday: 3, OldStartClass: 3, OldEndClass: 4, NewWeek: 9, NewWeekday: 1, NewStartClass: 7, NewEndClass: 8, NewLocation: "旗山西1-206"},
			},
			expected: []CourseScheduleRule{
				{Location: "铜盘A110", StartClass: 3, EndClass: 4, StartWeek: 1, EndWeek: 16, Weekday: 1, Single: true, Double: true},
			},
		},
		{
			name: "MultipleScheduleRules",
			rules: []CourseScheduleRule{
				{Location: "铜盘A110", StartClass: 3, EndClass: 4, StartWeek: 1, EndWeek: 16, Weekday: 1, Single: true, Double: true},
				{Location: "铜盘A508", StartClass: 7, EndClass: 8, StartWeek: 1, EndWeek: 16, Weekday: 5, Single: true, Double: true},
			},
			adjusts: []CourseAdjustRule{
				{OldWeek: 4, OldWeekday: 5, OldStartClass: 7, OldEndClass: 8, NewWeek: 5, NewWeekday: 2, NewStartClass: 7, NewEndClass: 8, NewLocation: "旗山东3-101"},
			},
			expected: []CourseScheduleRule{
				{Location: "铜盘A110", StartClass: 3, EndClass: 4, StartWeek: 1, EndWeek: 16, Weekday: 1, Single: true, Double: true},
				{Location: "旗山东3-101", StartClass: 7, EndClass: 8, StartWeek: 5, EndWeek: 5, Weekday: 2, Single: true, Double: true, Adjust: true},
				{Location: "铜盘A508", StartClass: 7, EndClass: 8, StartWeek: 1, EndWeek: 3, Weekday: 5, Single: true, Double: true},
				{Location: "铜盘A508", StartClass: 7, EndClass: 8, StartWeek: 5, EndWeek: 16, Weekday: 5, Single: true, Double: true},
			},
		},
		{
			name: "ConsecutiveWeeksRemoved",
			rules: []CourseScheduleRule{
				{Location: "铜盘A110", StartClass: 1, EndClass: 2, StartWeek: 1, EndWeek: 10, Weekday: 3, Single: true, Double: true},
			},
			adjusts: []CourseAdjustRule{
				{OldWeek: 5, OldWeekday: 3, OldStartClass: 1, OldEndClass: 2, NewWeek: 11, NewWeekday: 4, NewStartClass: 1, NewEndClass: 2, NewLocation: "旗山西1-206"},
				{OldWeek: 6, OldWeekday: 3, OldStartClass: 1, OldEndClass: 2, NewWeek: 12, NewWeekday: 4, NewStartClass: 1, NewEndClass: 2, NewLocation: "旗山西1-206"},
			},
			expected: []CourseScheduleRule{
				{Location: "旗山西1-206", StartClass: 1, EndClass: 2, StartWeek: 11, EndWeek: 11, Weekday: 4, Single: true, Double: true, Adjust: true},
				{Location: "旗山西1-206", StartClass: 1, EndClass: 2, StartWeek: 12, EndWeek: 12, Weekday: 4, Single: true, Double: true, Adjust: true},
				{Location: "铜盘A110", StartClass: 1, EndClass: 2, StartWeek: 1, EndWeek: 4, Weekday: 3, Single: true, Double: true},
				{Location: "铜盘A110", StartClass: 1, EndClass: 2, StartWeek: 7, EndWeek: 10, Weekday: 3, Single: true, Double: true},
			},
		},
		{
			name: "PreservesFromFullWeek",
			rules: []CourseScheduleRule{
				{Location: "", StartClass: 1, EndClass: 8, StartWeek: 3, EndWeek: 4, Weekday: 1, Single: true, Double: true, FromFullWeek: true},
			},
			adjusts: []CourseAdjustRule{
				{OldWeek: 3, OldWeekday: 1, OldStartClass: 1, OldEndClass: 8, NewWeek: 5, NewWeekday: 1, NewStartClass: 1, NewEndClass: 8, NewLocation: ""},
			},
			expected: []CourseScheduleRule{
				{Location: "", StartClass: 1, EndClass: 8, StartWeek: 5, EndWeek: 5, Weekday: 1, Single: true, Double: true, Adjust: true},
				{Location: "", StartClass: 1, EndClass: 8, StartWeek: 4, EndWeek: 4, Weekday: 1, Single: true, Double: true, FromFullWeek: true},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ApplyAdjustRules(tc.rules, tc.adjusts)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("result mismatch\ngot:      %+v\nexpected: %+v", result, tc.expected)
			}
		})
	}
}
