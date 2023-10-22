package jwch

import (
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

var (
	buildingMap = map[string]string{
		"x3":  "公共教学楼西3",
		"x2":  "公共教学楼西2",
		"x1":  "公共教学楼西1",
		"zl":  "公共教学楼中楼",
		"d1":  "公共教学楼东1",
		"d2":  "公共教学楼东2",
		"d3":  "公共教学楼东3",
		"wkl": "公共教学楼文科楼",
		"wk":  "公共教学楼文科楼",
	}
)

func (s *Student) GetEmptyRoom(req EmptyRoomReq) (error, []string) {
	err, viewStateMap := s.getEmptyRoomState()
	if err != nil {
		return err, nil
	}
	res, err := s.PostWithSession("https://jwcjwxt2.fzu.edu.cn:81/kkgl/kbcx/kbcx_kjs.aspx",
		map[string]string{
			"__VIEWSTATE":                         viewStateMap["VIEWSTATE"],
			"__EVENTVALIDATION":                   viewStateMap["EVENTVALIDATION"],
			"ctl00$TB_rq":                         req.Time,
			"ctl00$qsjdpl":                        req.Start,
			"ctl00$zzjdpl":                        req.End,
			"ctl00$jxldpl":                        buildingMap[req.Building],
			"ctl00$xnxqdpl":                       "202301",
			"ctl00$xqdpl":                         "旗山校区",
			"ctl00$xz1":                           ">=",
			"ctl00$jsrldpl":                       "0",
			"ctl00$xz2":                           ">=",
			"ctl00$ksrldpl":                       "0",
			"ctl00$ContentPlaceHolder1$BT_search": "查询",
		})
	if err != nil {
		return err, nil
	}
	err, rooms := parseEmptyRoom(res)
	if err != nil {
		return err, nil
	}
	return nil, rooms
}

// 获取VIEWSTATE和EVENTVALIDATION
func (s *Student) getEmptyRoomState() (error, map[string]string) {
	resp, err := s.GetWithSession("https://jwcjwxt2.fzu.edu.cn:81/kkgl/kbcx/kbcx_kjs.aspx")
	if err != nil {
		return err, nil
	}
	viewState := htmlquery.SelectAttr(htmlquery.FindOne(resp, `//*[@id="__VIEWSTATE"]`), "value")
	eventValidation := htmlquery.SelectAttr(htmlquery.FindOne(resp, `//*[@id="__EVENTVALIDATION"]`), "value")
	return nil, map[string]string{
		"VIEWSTATE":       viewState,
		"EVENTVALIDATION": eventValidation,
	}

}

func parseEmptyRoom(doc *html.Node) (error, []string) {
	sel := htmlquery.Find(doc, "//*[@id='jsdpl']//option")
	var res []string
	for _, opt := range sel {
		res = append(res, htmlquery.InnerText(opt))
	}
	return nil, res
}
