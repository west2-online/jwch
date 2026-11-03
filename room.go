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
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"

	"github.com/west2-online/jwch/constants"
	"github.com/west2-online/jwch/errno"
)

func (s *Student) GetEmptyRoom(req EmptyRoomReq) ([]string, error) {
	viewStateMap, err := s.getState(constants.ClassroomQueryURL)
	if err != nil {
		return nil, err
	}
	roomTypes, emptyRoomState, err := s.getEmptyRoomTypes(viewStateMap, "", req)
	if err != nil {
		return nil, err
	}
	// 按照教室类型进行并发访问
	channels := make([]chan struct {
		res []string
		err error
	}, len(roomTypes))
	var rooms []string
	for i, t := range roomTypes {
		channels[i] = make(chan struct {
			res []string
			err error
		})
		go func(t string, ch chan struct {
			res []string
			err error
		},
		) {
			res, err := s.PostWithIdentifier(constants.ClassroomQueryURL,
				map[string]string{
					"__VIEWSTATE":                         emptyRoomState["VIEWSTATE"],
					"__EVENTVALIDATION":                   emptyRoomState["EVENTVALIDATION"],
					"ctl00$TB_rq":                         req.Time,
					"ctl00$qsjdpl":                        req.Start,
					"ctl00$zzjdpl":                        req.End,
					"ctl00$jslxdpl":                       t,
					"ctl00$xqdpl":                         req.Campus,
					"ctl00$xz1":                           ">=",
					"ctl00$jsrldpl":                       "0",
					"ctl00$xz2":                           ">=",
					"ctl00$ksrldpl":                       "0",
					"ctl00$ContentPlaceHolder1$BT_search": "查询",
				})
			if err != nil {
				ch <- struct {
					res []string
					err error
				}{res: nil, err: err}
				return
			}
			rooms, err := parseEmptyRoom(res)
			if err != nil {
				ch <- struct {
					res []string
					err error
				}{res: nil, err: err}
				return
			}
			ch <- struct {
				res []string
				err error
			}{res: rooms, err: nil}
		}(t, channels[i])
	}
	for _, ch := range channels {
		temp := <-ch
		if temp.err != nil {
			return nil, temp.err
		}
		rooms = append(rooms, temp.res...)
	}
	return rooms, err
}

func (s *Student) GetQiShanEmptyRoom(req EmptyRoomReq) ([]string, error) {
	viewStateMap, err := s.getState(constants.ClassroomQueryURL)
	if err != nil {
		return nil, err
	}
	var rooms []string
	// 这里按照building的顺序进行并发爬取
	// 创建channel数组
	channels := make([]chan struct {
		res []string
		err error
	}, len(constants.BuildingArray))

	for i, building := range constants.BuildingArray {
		channels[i] = make(chan struct {
			res []string
			err error
		})
		go func(index int, building string, ch chan struct {
			res []string
			err error
		},
		) {
			roomTypes, emptyRoomState, err := s.getEmptyRoomTypes(viewStateMap, building, req)
			if err != nil {
				ans := struct {
					res []string
					err error
				}{res: nil, err: err}
				ch <- ans
			}
			var rooms []string
			for _, t := range roomTypes {
				res, err := s.PostWithIdentifier(constants.ClassroomQueryURL,
					map[string]string{
						"__VIEWSTATE":                         emptyRoomState["VIEWSTATE"],
						"__EVENTVALIDATION":                   emptyRoomState["EVENTVALIDATION"],
						"ctl00$TB_rq":                         req.Time,
						"ctl00$qsjdpl":                        req.Start,
						"ctl00$zzjdpl":                        req.End,
						"ctl00$jxldpl":                        building,
						"ctl00$jslxdpl":                       t,
						"ctl00$xqdpl":                         req.Campus,
						"ctl00$xz1":                           ">=",
						"ctl00$jsrldpl":                       "0",
						"ctl00$xz2":                           ">=",
						"ctl00$ksrldpl":                       "0",
						"ctl00$ContentPlaceHolder1$BT_search": "查询",
					})
				if err != nil {
					ch <- struct {
						res []string
						err error
					}{res: nil, err: err}
					return
				}

				roomList, err := parseEmptyRoom(res)
				if err != nil {
					ch <- struct {
						res []string
						err error
					}{res: roomList, err: err}
					return
				}

				rooms = append(rooms, roomList...)
			}
			ch <- struct {
				res []string
				err error
			}{res: rooms, err: nil}
		}(i, building, channels[i])
	}

	// 按顺序合并结果
	for _, ch := range channels {
		temp := <-ch
		if temp.err != nil {
			return nil, temp.err
		}
		rooms = append(rooms, temp.res...)
	}

	return rooms, nil
}

// 获取VIEWSTATE和EVENTVALIDATION
// 抽象成一个函数, 因为基本上每个请求都需要这两个参数
func (s *Student) getState(url string) (map[string]string, error) {
	resp, err := s.GetWithIdentifier(url)
	if err != nil {
		return nil, err
	}
	viewState := htmlquery.SelectAttr(htmlquery.FindOne(resp, `//*[@id="__VIEWSTATE"]`), "value")
	eventValidation := htmlquery.SelectAttr(htmlquery.FindOne(resp, `//*[@id="__EVENTVALIDATION"]`), "value")
	return map[string]string{
		"VIEWSTATE":       viewState,
		"EVENTVALIDATION": eventValidation,
	}, nil
}

// 获取教室类型
func (s *Student) getEmptyRoomTypes(viewStateMap map[string]string, building string, req EmptyRoomReq) ([]string, map[string]string, error) {
	var res *html.Node
	var err error
	if building != "" {
		res, err = s.PostWithIdentifier(constants.ClassroomQueryURL, map[string]string{
			"__VIEWSTATE":                         viewStateMap["VIEWSTATE"],
			"__EVENTVALIDATION":                   viewStateMap["EVENTVALIDATION"],
			"ctl00$TB_rq":                         req.Time,
			"ctl00$qsjdpl":                        req.Start,
			"ctl00$zzjdpl":                        req.End,
			"ctl00$jxldpl":                        building,
			"ctl00$xqdpl":                         req.Campus,
			"ctl00$xz1":                           ">=",
			"ctl00$jsrldpl":                       "0",
			"ctl00$xz2":                           ">=",
			"ctl00$ksrldpl":                       "0",
			"ctl00$ContentPlaceHolder1$BT_search": "查询",
		})
	} else {
		res, err = s.PostWithIdentifier(constants.ClassroomQueryURL, map[string]string{
			"__VIEWSTATE":                         viewStateMap["VIEWSTATE"],
			"__EVENTVALIDATION":                   viewStateMap["EVENTVALIDATION"],
			"ctl00$TB_rq":                         req.Time,
			"ctl00$qsjdpl":                        req.Start,
			"ctl00$zzjdpl":                        req.End,
			"ctl00$xqdpl":                         req.Campus,
			"ctl00$xz1":                           ">=",
			"ctl00$jsrldpl":                       "0",
			"ctl00$xz2":                           ">=",
			"ctl00$ksrldpl":                       "0",
			"ctl00$ContentPlaceHolder1$BT_search": "查询",
		})
	}
	if err != nil {
		return nil, nil, err
	}
	if res == nil {
		return nil, nil, nil
	}

	sel := htmlquery.Find(res, "//*[@id='jslxdpl']//option")
	var types []string
	for _, opt := range sel {
		types = append(types, htmlquery.InnerText(opt))
	}

	viewState := htmlquery.SelectAttr(htmlquery.FindOne(res, `//*[@id="__VIEWSTATE"]`), "value")
	eventValidation := htmlquery.SelectAttr(htmlquery.FindOne(res, `//*[@id="__EVENTVALIDATION"]`), "value")

	return types, map[string]string{
		"VIEWSTATE":       viewState,
		"EVENTVALIDATION": eventValidation,
	}, nil
}

func parseEmptyRoom(doc *html.Node) ([]string, error) {
	if doc == nil {
		return nil, nil
	}
	sel := htmlquery.Find(doc, "//*[@id='jsdpl']//option")
	var res []string
	for _, opt := range sel {
		res = append(res, htmlquery.InnerText(opt))
	}
	return res, nil
}

// 考场查询
// 与「我的选课」使用同一个数据源，但只解析考场需要的四列，避免完整解析课表。
func (s *Student) GetExamRoom(req ExamRoomReq) ([]*ExamRoomInfo, error) {
	state, err := s.getState(constants.CourseURL)
	if err != nil {
		return nil, err
	}

	resp, err := s.PostWithIdentifier(constants.CourseURL, map[string]string{
		"ctl00$ContentPlaceHolder1$DDL_xnxq":  req.Term,
		"ctl00$ContentPlaceHolder1$BT_submit": "确定",
		"__VIEWSTATE":                         state["VIEWSTATE"],
		"__EVENTVALIDATION":                   state["EVENTVALIDATION"],
	})
	if err != nil {
		return nil, err
	}

	return parseExamRoomFromCourses(resp)
}

// parseExamRoomFromCourses 解析「我的选课」表格中的课程名、学分、教师和考试时间地点四列。
func parseExamRoomFromCourses(resp *html.Node) ([]*ExamRoomInfo, error) {
	list := htmlquery.Find(htmlquery.FindOne(resp, `//*[@id="ContentPlaceHolder1_DataList_xxk"]/tbody`), "tr")

	// 去除第一个元素，第一个元素是标题栏，有个判断文本是“课程名称”
	// TODO: 我们如何确保第一个元素一定是标题栏?
	list = list[2:]

	res := make([]*ExamRoomInfo, 0)

	for _, node := range list {
		// 教务处的表格HTML是不规范的，因此XPath解析会出现一些BUG
		if strings.TrimSpace(htmlquery.SelectAttr(node, "style")) == "" {
			continue
		}
		info := htmlquery.Find(node, `td`) // 一行的所有信息

		// 这个表格有12栏
		if len(info) < 12 {
			return nil, errno.HTMLParseError.WithMessage("get exam room info failed")
		}

		// 考场信息只需要课程名、学分、教师和考试时间地点
		examTime := parseExamTime(strings.TrimSpace(htmlquery.InnerText(info[9])))
		res = append(res, &ExamRoomInfo{
			CourseName: htmlquery.OutputHTML(info[1], false),
			Credit:     safeExtractionFirst(info[4], "span"),
			Teacher:    htmlquery.OutputHTML(info[7], false),
			Date:       examTime.Date,
			Time:       examTime.Time,
			Location:   examTime.Location,
		})
	}

	return res, nil
}
