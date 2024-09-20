package jwch

import (
	"strconv"
	"strings"

	"github.com/west2-online/jwch/constants"
	"github.com/west2-online/jwch/errno"
	"github.com/west2-online/jwch/utils"

	"github.com/antchfx/htmlquery"
)

// 获取我的学期
func (s *Student) GetTerms() (*Term, error) {
	resp, err := s.GetWithSession(constants.CourseURL)

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

	resp, err := s.PostWithSession(constants.CourseURL, map[string]string{
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

		// 解析上课时间、地点
		/*
			05-18 星期1:3-4节 铜盘A110
			05-17 星期3:1-2节 铜盘A110
			05-17 星期5:3-4节 铜盘A110
		*/
		courseInfo8 := strings.Split(utils.InnerTextWithBr(info[8]), "\n")
		scheduleRules := []CourseScheduleRule{}

		for i := 0; i < len(courseInfo8); i++ {
			courseInfo8[i] = strings.TrimSpace(courseInfo8[i])

			if courseInfo8[i] == "" { // 空行
				continue
			}

			lineData := strings.Fields(courseInfo8[i])

			if len(lineData) < 3 {
				return nil, errno.HTMLParseError.WithMessage("get course info failed")
			}

			if strings.Contains(lineData[0], "周") { // 处理整周的课程，比如军训
				/*
					03周  星期1  -  04周  星期7
					[0] 03周
					[1] 星期1
					[2] -
					[3] 04周
					[4] 星期7
				*/
				startWeek, _ := strconv.Atoi(strings.TrimSuffix(lineData[0], "周"))
				endWeek, _ := strconv.Atoi(strings.TrimSuffix(lineData[3], "周"))
				startWeekday, _ := strconv.Atoi(strings.TrimPrefix(lineData[1], "星期"))
				endWeekday, _ := strconv.Atoi(strings.TrimPrefix(lineData[4], "星期"))

				/*
					目前对于这种课程的解析有两种猜测:
						1. 第3周的周一到第4周的周日
						2. 第3周到第4周，每周的周一到周日
					福uu客户端现在采用的是猜测2，所以现在先按照猜测2来解析
				*/
				for i := startWeekday; i <= endWeekday; i++ {
					scheduleRules = append(scheduleRules, CourseScheduleRule{
						Location:   "",
						StartClass: 1,
						EndClass:   8,
						StartWeek:  startWeek,
						EndWeek:    endWeek,
						Weekday:    i,
						Single:     true,
						Double:     true,
					})
				}

				continue
			}

			/*
				08-16 星期5:7-8节 铜盘A508
				[0] 08-16
				[1] 星期5:7-8节
				[2] 铜盘A508
			*/
			/*
				02-14 星期1:1-2节(双) 旗山西1-206
				[0] 02-14
				[1] 星期1:1-2节(双)
				[2] 旗山西1-206
			*/
			/*
				01-13 星期1:3-4节(单) 旗山西1-206
				[0] 01-13
				[1] 星期1:3-4节(单)
				[2] 旗山西1-206
			*/

			weekInfo := strings.SplitN(lineData[0], "-", 2)    // [8, 16]
			dayInfo := strings.SplitN(lineData[1], ":", 2)     // ["星期5", "7-8节"] or ["星期1", "1-2节(双)"]
			classBasicInfo := strings.Split(dayInfo[1], "节")   // ["7-8", ""] or ["1-2", "(双)"]
			classInfo := strings.Split(classBasicInfo[0], "-") // ["7", "8"]

			scheduleRules = append(scheduleRules, CourseScheduleRule{
				Location:   lineData[2],
				StartClass: utils.SafeAtoi(classInfo[0]),
				EndClass:   utils.SafeAtoi(classInfo[1]),
				StartWeek:  utils.SafeAtoi(weekInfo[0]),
				EndWeek:    utils.SafeAtoi(weekInfo[1]),
				Weekday:    utils.SafeAtoi(strings.TrimPrefix(dayInfo[0], "星期")),
				Single:     !strings.Contains(classBasicInfo[1], "双"),
				Double:     !strings.Contains(classBasicInfo[1], "单"),
			})
		}

		// TODO: performance optimization
		res = append(res, &Course{
			Type:             htmlquery.OutputHTML(info[0], false),
			Name:             htmlquery.OutputHTML(info[1], false),
			Syllabus:         "https://jwcjwxt2.fzu.edu.cn:81" + safeExtractRegex(`javascript:pop1\('(.*?)&`, safeExtractionValue(info[2], "a", "href", 0)),
			LessonPlan:       "https://jwcjwxt2.fzu.edu.cn:81" + safeExtractRegex(`javascript:pop1\('(.*?)&`, safeExtractionValue(info[2], "a", "href", 1)),
			PaymentStatus:    safeExtractionFirst(info[3], "font"),
			Credits:          safeExtractionFirst(info[4], "span"),
			ElectiveType:     utils.GetChineseCharacter(htmlquery.OutputHTML(info[5], false)),
			ExamType:         utils.GetChineseCharacter(htmlquery.OutputHTML(info[6], false)),
			Teacher:          htmlquery.OutputHTML(info[7], false),
			ScheduleRules:    scheduleRules,
			RawScheduleRules: strings.Join(courseInfo8, "\n"),
			ExamTime:         strings.TrimSpace(htmlquery.InnerText(info[9])),
			Remark:           htmlquery.OutputHTML(info[10], false),
			Adjust:           htmlquery.OutputHTML(info[11], false),
		})
	}

	return res, nil
}
