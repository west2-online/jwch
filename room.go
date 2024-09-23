package jwch

import (
	"github.com/antchfx/htmlquery"
	"github.com/west2-online/jwch/constants"
	"golang.org/x/net/html"
)

func (s *Student) GetEmptyRoom(req EmptyRoomReq) ([]string, error) {
	viewStateMap, err := s.getEmptyRoomState()
	if err != nil {
		return nil, err
	}
	roomTypes, emptyRoomState, err := s.getEmptyRoomTypes(viewStateMap, "", req)
	if err != nil {
		return nil, err
	}
	//按照教室类型进行并发访问
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
		}) {
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
	viewStateMap, err := s.getEmptyRoomState()
	if err != nil {
		return nil, err
	}
	var rooms []string
	//这里按照building的顺序进行并发爬取
	//创建channel数组
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
		}) {
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
func (s *Student) getEmptyRoomState() (map[string]string, error) {
	resp, err := s.GetWithIdentifier(constants.ClassroomQueryURL)
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
	if building != "" {
		res, _ = s.PostWithIdentifier(constants.ClassroomQueryURL, map[string]string{
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
		res, _ = s.PostWithIdentifier(constants.ClassroomQueryURL, map[string]string{
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
	sel := htmlquery.Find(doc, "//*[@id='jsdpl']//option")
	var res []string
	for _, opt := range sel {
		res = append(res, htmlquery.InnerText(opt))
	}
	return res, nil
}
