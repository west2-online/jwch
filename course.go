package jwch

import (
	"strings"

	"github.com/west2-online/jwch/errno"
	"github.com/west2-online/jwch/utils"

	"github.com/antchfx/htmlquery"
)

// 获取我的学期
func (s *Student) GetTerms() (*Term, error) {
	resp, err := s.GetWithSession("https://jwcjwxt2.fzu.edu.cn:81/student/xkjg/wdxk/xkjg_list.aspx")

	if err != nil {
		return nil, err
	}

	res := &Term{}

	res.ViewState = htmlquery.SelectAttr(htmlquery.FindOne(resp, `//*[@id="__VIEWSTATE"]`), "value")
	res.EventValidation = htmlquery.SelectAttr(htmlquery.FindOne(resp, `//*[@id="__EVENTVALIDATION"]`), "value")

	// 获取学年学期，例如 202202/202201/202102/202101 需要获取value
	list := htmlquery.Find(resp, `//*[@id="ContentPlaceHolder1_DDL_xnxq"]/option/@value`)

	// 这里考虑过使用 len(list) < 1，但是实际上这没必要，因为小于1那么它必定是0
	if len(list) == 0 {
		return nil, errno.HTMLParseError.WithMessage("empty terms")
	}

	for _, node := range list {
		res.Terms = append(res.Terms, htmlquery.SelectAttr(node, "value"))
	}

	return res, nil
}

// 获取我的选课
func (s *Student) GetSemesterCourses(term, viewState, eventValidation string) ([]*Course, error) {

	resp, err := s.PostWithSession("https://jwcjwxt2.fzu.edu.cn:81/student/xkjg/wdxk/xkjg_list.aspx", map[string]string{
		"ctl00$ContentPlaceHolder1$DDL_xnxq":  term,
		"ctl00$ContentPlaceHolder1$BT_submit": "确定",
		"__VIEWSTATE":                         viewState,
		"__EVENTVALIDATION":                   eventValidation,
	})

	if err != nil {
		return nil, err
	}

	list := htmlquery.Find(htmlquery.FindOne(resp, `//*[@id="ContentPlaceHolder1_DataList_xxk"]/tbody`), "tr")

	// 去除第一个元素，第一个元素是标题栏，有个判断文本是“课程名称”
	// TODO: 我们如何确保第一个元素一定是标题栏?
	list = list[2:]

	res := make([]*Course, 0)

	for _, node := range list {

		// 教务处的表格HTML是不规范的，因此XPath解析会出现一些BUG
		if strings.TrimSpace(htmlquery.SelectAttr(node, "style")) == "" {
			continue
		}
		info := htmlquery.Find(node, `td`) // 一行的所有信息

		// 这个表格有12栏
		if len(info) < 12 {
			return nil, errno.HTMLParseError.WithMessage("get course info failed")
		}

		// TODO: performance optimization
		res = append(res, &Course{
			Type:          htmlquery.OutputHTML(info[0], false),
			Name:          htmlquery.OutputHTML(info[1], false),
			Syllabus:      "https://jwcjwxt2.fzu.edu.cn:81" + safeExtractRegex(`javascript:pop1\('(.*?)&`, safeExtractionValue(info[2], "a", "href", 0)),
			LessonPlan:    "https://jwcjwxt2.fzu.edu.cn:81" + safeExtractRegex(`javascript:pop1\('(.*?)&`, safeExtractionValue(info[2], "a", "href", 1)),
			PaymentStatus: safeExtractionFirst(info[3], "font"),
			Credits:       safeExtractionFirst(info[4], "span"),
			ElectiveType:  utils.GetChineseCharacter(htmlquery.OutputHTML(info[5], false)),
			ExamType:      utils.GetChineseCharacter(htmlquery.OutputHTML(info[6], false)),
			Teacher:       htmlquery.OutputHTML(info[7], false),
			Classroom:     strings.TrimSpace(htmlquery.InnerText(info[8])),
			ExamTime:      strings.TrimSpace(htmlquery.InnerText(info[9])),
			Remark:        htmlquery.OutputHTML(info[10], false),
			Adjust:        htmlquery.OutputHTML(info[11], false),
		})
	}

	return res, nil
}
