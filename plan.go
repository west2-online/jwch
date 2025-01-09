package jwch

import (
	"fmt"
	"strings"

	"github.com/antchfx/htmlquery"

	"github.com/west2-online/jwch/constants"
)

func (s *Student) GetCultivatePlan() (string, error) {
	info, err := s.GetInfo()
	if err != nil {
		return "", err
	}
	viewStateMap, err := s.getState(constants.CultivatePlanURL)
	if err != nil {
		return "", err
	}
	res, err := s.PostWithIdentifier(constants.CultivatePlanURL,
		map[string]string{
			"__VIEWSTATE":       viewStateMap["VIEWSTATE"],
			"__EVENTVALIDATION": viewStateMap["EVENTVALIDATION"],
			"ctl00$njdpl":       info.Grade, // 年级
			// "ctl00$xymcdpl":                       "010",// 学院名称
			"ctl00$dldpl":                         "<-全部->", // 大类
			"ctl00$zymcdpl":                       "<-全部->", // 专业代码(无法获取)
			"ctl00$zylbdpl":                       "本专业",   // 本专业、辅修
			"ctl00$ContentPlaceHolder1$DDL_syxw":  "<-全部->",
			"ctl00$ContentPlaceHolder1$BT_submit": "确定",
		})
	if err != nil {
		return "", err
	}
	xpathExpr := strings.Join([]string{"//tr[td[contains(., '", info.Major, "')]]/td/a[contains(@href, 'pyfa')]/@href"}, "")
	node := htmlquery.FindOne(res, xpathExpr)
	if node == nil {
		return "", fmt.Errorf("%s", "cultivate plan not found")
	}

	url := htmlquery.SelectAttr(node, "href")
	formatUrl := constants.JwchPrefix + "/pyfa/pyjh/" + strings.TrimPrefix(strings.TrimSuffix(url, "')"), "javascript:pop1('")
	return formatUrl, nil
}
