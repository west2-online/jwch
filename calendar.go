package jwch

import (
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	iconv "github.com/djimenez/iconv-go"
	"github.com/west2-online/jwch/constants"
)

func (s *Student) GetSchoolCalendar() (*SchoolCalendar, error) {
	resp, err := s.GetWithIdentifier(constants.SchoolCalendarURL)

	if err != nil {
		return nil, err
	}

	rawCurTerm := htmlquery.InnerText(htmlquery.FindOne(resp, `//html/body/center/div`))
	rawCurTerm, _ = iconv.ConvertString(rawCurTerm, "gb2312", "utf-8")
	curTermRegex := regexp.MustCompile(`当前学期：(\d{6})`)
	curTerm := curTermRegex.FindStringSubmatch(rawCurTerm)[1]

	res := &SchoolCalendar{
		CurrentTerm: curTerm,
	}

	list := htmlquery.Find(resp, `//select[@name="xq"]/option/@value`)

	for _, node := range list {
		rawTerm := htmlquery.SelectAttr(node, "value")
		/*
			2024012024082620250117
			[0] 202401
			[1] 20240826
			[2] 20250117
		*/
		schoolYear := rawTerm[0:4]
		term := rawTerm[0:6]
		startDate := rawTerm[6:14]
		endDate := rawTerm[14:22]

		// convert 20240826 to 2024-08-26
		startDate = startDate[0:4] + "-" + startDate[4:6] + "-" + startDate[6:8]
		endDate = endDate[0:4] + "-" + endDate[4:6] + "-" + endDate[6:8]

		res.Terms = append(res.Terms, CalTerm{
			TermId:     rawTerm,
			SchoolYear: schoolYear,
			Term:       term,
			StartDate:  startDate,
			EndDate:    endDate,
		})
	}

	return res, nil
}

func (s *Student) GetTermEvents(termId string) (*CalTermEvents, error) {
	resp, err := s.PostWithIdentifier(constants.SchoolCalendarURL, map[string]string{
		"xq":     termId,
		"submit": "提交",
	})

	if err != nil {
		return nil, err
	}

	rawTermDetail := htmlquery.InnerText(htmlquery.FindOne(resp, `/html/body/table[2]/tbody/tr`))
	rawTermDetail = strings.ReplaceAll(rawTermDetail, " ", " ")
	rawTermDetail, _ = iconv.ConvertString(rawTermDetail, "gb2312", "utf-8")

	res := &CalTermEvents{
		TermId:     termId,
		Term:       termId[0:6],
		SchoolYear: termId[0:4],
	}

	termDetail := strings.Split(rawTermDetail, "；")

	for _, event := range termDetail {
		event = strings.TrimSpace(event)

		if event == "" {
			continue
		}

		rawData := strings.Split(event, "为")
		rawDate := strings.Split(rawData[0], "至")

		name := strings.TrimSpace(rawData[1])
		startDate := strings.TrimSpace(rawDate[0])
		endDate := strings.TrimSpace(rawDate[1])

		res.Events = append(res.Events, CalTermEvent{
			Name:      name,
			StartDate: startDate,
			EndDate:   endDate,
		})
	}

	return res, nil
}
