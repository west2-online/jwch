package jwch

import (
	"fmt"
	"strings"

	"github.com/west2-onlin/jwch/errno"
	"github.com/west2-onlin/jwch/utils"

	"github.com/antchfx/htmlquery"
)

// 获取成绩，由于教务处缺陷，这里会返回全部的成绩
func (s *Student) GetMarks() (resp []*Mark, err error) {
	res, err := s.GetWithSession("https://jwcjwxt2.fzu.edu.cn:81/student/xyzk/cjyl/score_sheet.aspx")

	if err != nil {
		return nil, err
	}

	list := htmlquery.Find(htmlquery.FindOne(res, `//*[@id="ContentPlaceHolder1_DataList_xxk"]/tbody`), "tr")

	// 去除第一个元素，第一个元素是标题栏，有个判断文本是“课程名称”
	// TODO: 我们如何确保第一个元素一定是标题栏?
	list = list[2:]

	resp = make([]*Mark, 0)

	for _, node := range list {

		// 教务处的表格HTML是不规范的，因此XPath解析会出现一些BUG
		if strings.TrimSpace(htmlquery.SelectAttr(node, "style")) == "" {
			continue
		}
		info := htmlquery.Find(node, `td`) // 一行的所有信息

		// 这个表格有12栏
		if len(info) < 12 {
			return nil, errno.HTMLParseError.WithMessage("get mark info failed")
		}

		// TODO: performance optimization
		resp = append(resp, &Mark{
			Type:          htmlquery.OutputHTML(info[0], false),
			Semester:      htmlquery.OutputHTML(info[1], false),
			Name:          htmlquery.OutputHTML(info[2], false),
			Credits:       safeExtractionFirst(info[3], "span"),
			Score:         safeExtractionFirst(info[4], "font"),
			GPA:           htmlquery.OutputHTML(info[5], false),
			EarnedCredits: htmlquery.OutputHTML(info[6], false),
			ElectiveType:  utils.GetChineseCharacter(htmlquery.OutputHTML(info[7], false)),
			ExamType:      utils.GetChineseCharacter(htmlquery.OutputHTML(info[8], false)),
			Teacher:       htmlquery.OutputHTML(info[9], false),
			Classroom:     strings.TrimSpace(htmlquery.InnerText(info[10])),
			ExamTime:      strings.TrimSpace(htmlquery.InnerText(info[11])),
		})
	}

	return resp, nil
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
