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
	_, err := stu.GetCultivatePlan()
	if err != nil {
		t.Error(err)
	}
}
