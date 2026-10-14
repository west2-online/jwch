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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type usagePageClient interface {
	Get(context.Context) (*html.Node, error)
	Post(context.Context, map[string]string) (*html.Node, error)
}

// fetchClassroomUsage 按日期遍历校区内的教学楼，查询每栋楼的教室占用表。
// 切换校区和教学楼时先回发选择事件，取得联动下拉框及最新页面状态。
func fetchClassroomUsage(ctx context.Context, client usagePageClient, date, requestedTerm, campus, requestedBuilding string) (string, []*ClassroomUsageRoom, error) {
	queryDate, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return "", nil, fmt.Errorf("invalid classroom date: %w", err)
	}
	doc, err := client.Get(ctx)
	if err != nil {
		return "", nil, err
	}
	term := requestedTerm
	if term == "" {
		term = selectedValue(doc, "xnxqdpl")
	}
	if term == "" {
		return "", nil, fmt.Errorf("classroom usage page has no selected academic term")
	}
	if selectedValue(doc, "xqdpl") != campus {
		doc, err = postSelection(ctx, client, doc, date, term, campus, "", "ctl00$xqdpl")
		if err != nil {
			return "", nil, fmt.Errorf("select campus %s: %w", campus, err)
		}
	}
	if selectedValue(doc, "xqdpl") != campus {
		return "", nil, fmt.Errorf("classroom usage page did not select campus %s", campus)
	}
	var buildings []string
	for _, option := range htmlquery.Find(doc, "//*[@id='jxldpl']/option") {
		if building := htmlquery.SelectAttr(option, "value"); building != "" {
			buildings = append(buildings, building)
		}
	}
	if len(buildings) == 0 {
		return "", nil, fmt.Errorf("classroom usage page has no buildings for %s", campus)
	}
	rooms := make([]*ClassroomUsageRoom, 0)
	seen := make(map[string]bool)
	if requestedBuilding != "" {
		found := false
		for _, building := range buildings {
			if building == requestedBuilding {
				found = true
				break
			}
		}
		if !found {
			return term, nil, fmt.Errorf("classroom usage page has no building %s for %s", requestedBuilding, campus)
		}
		buildings = []string{requestedBuilding}
	}
	failedBuildings := make([]string, 0)
	failedSet := make(map[string]bool)
	var firstFailure error
	markFailed := func(building string, cause error) {
		if !failedSet[building] {
			failedBuildings = append(failedBuildings, building)
			failedSet[building] = true
		}
		if firstFailure == nil {
			firstFailure = cause
		}
	}
	for _, building := range buildings {
		buildingDoc, buildingErr := postSelection(ctx, client, doc, date, term, campus, building, "ctl00$jxldpl")
		if buildingErr != nil {
			markFailed(building, fmt.Errorf("select building %s: %w", building, buildingErr))
			continue
		}
		doc = buildingDoc
		var err error
		if selectedValue(doc, "xnxqdpl") != term || selectedValue(doc, "xqdpl") != campus || selectedValue(doc, "jxldpl") != building {
			err = fmt.Errorf("classroom usage page did not select building %s: got term=%q campus=%q building=%q",
				building, selectedValue(doc, "xnxqdpl"), selectedValue(doc, "xqdpl"), selectedValue(doc, "jxldpl"))
		}
		if err != nil {
			markFailed(building, err)
			continue
		}
		roomOptions := htmlquery.FindOne(doc, "//*[@id='jsdpl']")
		if roomOptions == nil {
			markFailed(building, fmt.Errorf("classroom usage page has no room selector for %s", building))
			continue
		}
		hasRooms := false
		for _, option := range htmlquery.Find(roomOptions, "./option") {
			if htmlquery.SelectAttr(option, "value") != "" {
				hasRooms = true
				break
			}
		}
		if !hasRooms {
			continue
		}
		form, err := queryForm(doc, date, term, campus, building)
		if err != nil {
			markFailed(building, fmt.Errorf("prepare building %s: %w", building, err))
			continue
		}
		form["ctl00$ContentPlaceHolder1$BT_search"] = "查询"
		buildingDoc, err = client.Post(ctx, form)
		if err != nil {
			markFailed(building, fmt.Errorf("query building %s: %w", building, err))
			continue
		}
		doc = buildingDoc
		if selectedValue(doc, "xnxqdpl") != term || selectedValue(doc, "xqdpl") != campus ||
			selectedValue(doc, "jxldpl") != building || inputValue(doc, "TB_rq") != date {
			markFailed(building, fmt.Errorf("classroom usage response does not match query for %s: got date=%q term=%q campus=%q building=%q",
				building, inputValue(doc, "TB_rq"), selectedValue(doc, "xnxqdpl"), selectedValue(doc, "xqdpl"), selectedValue(doc, "jxldpl")))
			continue
		}
		buildingRooms, err := parseClassroomUsage(doc, queryDate.Weekday())
		if err != nil {
			markFailed(building, fmt.Errorf("parse building %s: %w", building, err))
			continue
		}
		// 先完整校验当前楼的数据，再追加到总结果，避免重复教室导致半栋楼数据混入快照。
		buildingSeen := make(map[string]bool, len(buildingRooms))
		duplicate := ""
		invalidRoom := false
		for _, room := range buildingRooms {
			if room == nil || room.Location == "" || seen[room.Location] || buildingSeen[room.Location] {
				if room != nil {
					duplicate = room.Location
				}
				invalidRoom = true
				break
			}
			buildingSeen[room.Location] = true
		}
		if invalidRoom {
			markFailed(building, fmt.Errorf("duplicate or invalid classroom %s in campus %s", duplicate, campus))
			continue
		}
		for _, room := range buildingRooms {
			seen[room.Location] = true
			room.Building = building
			rooms = append(rooms, room)
		}
	}
	if len(failedBuildings) > 0 {
		return term, rooms, &ClassroomUsageError{Campus: campus, FailedBuildings: failedBuildings, Cause: firstFailure}
	}
	return term, rooms, nil
}

func postSelection(ctx context.Context, client usagePageClient, doc *html.Node,
	date, term, campus, building, target string,
) (*html.Node, error) {
	form, err := queryForm(doc, date, term, campus, building)
	if err != nil {
		return nil, err
	}
	form["__EVENTTARGET"] = target
	return client.Post(ctx, form)
}

func queryForm(doc *html.Node, date, term, campus, building string) (map[string]string, error) {
	form := map[string]string{
		"__EVENTTARGET": "", "__EVENTARGUMENT": "", "__LASTFOCUS": "",
		"ctl00$TB_rq": date, "ctl00$xnxqdpl": term, "ctl00$xqdpl": campus, "ctl00$jxldpl": building,
		"ctl00$qszdpl": "", "ctl00$qsxqdpl": "", "ctl00$qsjdpl": "", "ctl00$zzjdpl": "",
	}
	// 下拉框没有选项时，浏览器不会提交该字段。
	// jsdpl 初始没有选项，主动提交空值会触发 ASP.NET 联动控件校验，
	// 因此仅当页面提供空选项时才提交空值。
	for _, option := range htmlquery.Find(doc, "//*[@id='jsdpl']/option") {
		if htmlquery.SelectAttr(option, "value") == "" {
			form["ctl00$jsdpl"] = ""
			break
		}
	}
	// 同时保留拆分的 VIEWSTATE 字段，兼容教务处启用状态分片的情况。
	for _, input := range htmlquery.Find(doc, "//input[@type='hidden']") {
		name := htmlquery.SelectAttr(input, "name")
		if strings.HasPrefix(name, "__VIEWSTATE") || name == "__EVENTVALIDATION" {
			form[name] = htmlquery.SelectAttr(input, "value")
		}
	}
	if form["__VIEWSTATE"] == "" || form["__EVENTVALIDATION"] == "" {
		return nil, fmt.Errorf("classroom usage page is missing ASP.NET state (session may have expired)")
	}
	return form, nil
}

func selectedValue(doc *html.Node, id string) string {
	options := htmlquery.Find(doc, fmt.Sprintf("//*[@id='%s']/option", id))
	for _, option := range options {
		for _, attr := range option.Attr {
			if attr.Key == "selected" {
				return htmlquery.SelectAttr(option, "value")
			}
		}
	}
	if len(options) > 0 {
		return htmlquery.SelectAttr(options[0], "value")
	}
	return ""
}

func inputValue(doc *html.Node, id string) string {
	input := htmlquery.FindOne(doc, fmt.Sprintf("//*[@id='%s']", id))
	if input == nil {
		return ""
	}
	return htmlquery.SelectAttr(input, "value")
}
