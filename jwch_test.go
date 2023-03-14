package jwch

import (
	"fmt"
	"testing"

	"github.com/west2-onlin/jwch/utils"
)

const (
	username = "username" // 学号
	password = "password" // 密码
)

var (
	islogin   bool     = false
	localfile string   = "logindata.txt"
	stu       *Student = NewStudent().WithUser(username, password)
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

func Test_Login(t *testing.T) {
	err := login()
	if err != nil {
		t.Error(err)
	}
}

func Test_LoginFromLocal(t *testing.T) {
	var res LoginData
	err := utils.JSONUnmarshalFromFile(localfile, &res)
	if err != nil {
		t.Error(err)
	}
	stu.SetLoginData(res)

	err = stu.CheckSession()

	if err != nil {
		t.Log("session expire, relogin")
		err = stu.Login()
		if err != nil {
			t.Error(err)
		}
		err = stu.CheckSession()

		if err != nil {
			t.Error(err)
		}

		stu.SaveLoginData(localfile)
	}
}

func Test_GetCourse(t *testing.T) {
	if !islogin {
		err := login()

		if err != nil {
			t.Error(err)
		}
	}

	terms, err := stu.GetTerms()

	if err != nil {
		t.Error(err)
	}

	list, err := stu.GetSemesterCourses(terms.Terms[0], terms.ViewState, terms.EventValidation)

	if err != nil {
		t.Error(err)
	}

	fmt.Println("course num:", len(list))

	for _, v := range list {
		fmt.Println(utils.PrintStruct(v))
	}
}

func Test_GetInfo(t *testing.T) {
	if !islogin {
		err := login()

		if err != nil {
			t.Error(err)
		}
	}

	detail, err := stu.GetInfo()

	if err != nil {
		t.Error(err)
	}

	fmt.Println(utils.PrintStruct(detail))
}

func Test_GetMarks(t *testing.T) {
	if !islogin {
		err := login()

		if err != nil {
			t.Error(err)
		}
	}

	marks, err := stu.GetMarks()

	if err != nil {
		t.Error(err)
	}

	fmt.Println(utils.PrintStruct(marks))
}
