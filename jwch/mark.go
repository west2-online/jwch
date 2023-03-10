package jwch

import "fmt"

// 获取成绩，由于教务处缺陷，这里会返回全部的成绩
func (s *Student) GetMarks() error {
	resp, err := s.GetWithSession("https://jwcjwxt2.fzu.edu.cn:81/student/xyzk/cjyl/score_sheet.aspx")

	if err != nil {
		return err
	}

	fmt.Println(resp)
	return nil
}

// 获取CET成绩
func (s *Student) GetCET() error {
	resp, err := s.GetWithSession("https://jwcjwxt2.fzu.edu.cn:81/student/glbm/cet/cet_cszt.aspx")

	if err != nil {
		return err
	}

	fmt.Println(resp)
	return nil
}
