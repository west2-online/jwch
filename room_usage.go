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
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"

	"github.com/west2-online/jwch/utils"
)

const classroomUsageURL = "https://jwcjwxt2.fzu.edu.cn:81/kkgl/kbcx/kbcx_zb.aspx"

type classroomUsagePageState struct {
	document  *html.Node
	term      string
	buildings []string
}

// GetClassroomBuildings 获取指定校区的全部教学楼列表
func (s *Student) GetClassroomBuildings(ctx context.Context, campus string) ([]string, error) {
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	date := time.Now().In(location).Format(time.DateOnly)
	page, err := s.classroomUsagePage(ctx, date, "", campus)
	if err != nil {
		return nil, err
	}
	return page.buildings, nil
}

// GetClassroomUsage 查询每个教学楼的教室使用总表
func (s *Student) GetClassroomUsage(ctx context.Context, req ClassroomUsageReq) (*ClassroomUsage, error) {
	date, err := time.Parse(time.DateOnly, req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid classroom date: %w", err)
	}
	for _, building := range req.Buildings {
		if strings.TrimSpace(building) == "" {
			return nil, fmt.Errorf("classroom building is required")
		}
	}
	page, err := s.classroomUsagePage(ctx, req.Date, req.Term, req.Campus)
	if err != nil {
		return nil, err
	}
	doc := page.document
	term := page.term
	available := page.buildings
	buildings := req.Buildings
	if len(buildings) == 0 {
		buildings = available
	}
	allowed := make(map[string]bool, len(available))
	for _, building := range available {
		allowed[building] = true
	}
	for _, building := range buildings {
		if !allowed[building] {
			return nil, fmt.Errorf("classroom usage page has no building %s for %s", building, req.Campus)
		}
	}
	usage := &ClassroomUsage{Date: req.Date, Term: term, Rooms: make([]*ClassroomUsageRoom, 0)}
	seen := make(map[string]bool, len(buildings))
	for _, building := range buildings {
		if seen[building] {
			continue
		}
		seen[building] = true
		doc, err = s.postClassroomUsage(ctx, doc, req.Date, term, req.Campus, building, "ctl00$jxldpl")
		if err != nil {
			return nil, fmt.Errorf("select building %s: %w", building, err)
		}
		if selectedValue(doc, "xnxqdpl") != term || selectedValue(doc, "xqdpl") != req.Campus ||
			selectedValue(doc, "jxldpl") != building {
			return nil, fmt.Errorf("classroom usage page did not select building %s", building)
		}
		selector := htmlquery.FindOne(doc, "//*[@id='jsdpl']")
		if selector == nil {
			return nil, fmt.Errorf("classroom usage page has no room selector for %s", building)
		}
		hasRooms := false
		for _, option := range htmlquery.Find(selector, "./option") {
			if htmlquery.SelectAttr(option, "value") != "" {
				hasRooms = true
				break
			}
		}
		if !hasRooms {
			continue
		}
		doc, err = s.postClassroomUsage(ctx, doc, req.Date, term, req.Campus, building, "")
		if err != nil {
			return nil, err
		}
		if selectedValue(doc, "xnxqdpl") != term || selectedValue(doc, "xqdpl") != req.Campus ||
			selectedValue(doc, "jxldpl") != building || safeExtractionValue(doc, "//*[@id='TB_rq']", "value", 0) != req.Date {
			return nil, fmt.Errorf("classroom usage response does not match query for %s", building)
		}
		rooms, err := parseClassroomUsage(doc, date.Weekday())
		if err != nil {
			return nil, fmt.Errorf("parse building %s: %w", building, err)
		}
		for _, room := range rooms {
			room.Building = building
		}
		usage.Rooms = append(usage.Rooms, rooms...)
	}
	return usage, nil
}

func (s *Student) classroomUsagePage(ctx context.Context, date, requestedTerm, campus string) (*classroomUsagePageState, error) {
	if strings.TrimSpace(campus) == "" {
		return nil, fmt.Errorf("classroom campus is required")
	}
	if err := s.waitForRequest(ctx); err != nil {
		return nil, err
	}
	doc, err := s.GetWithIdentifier(classroomUsageURL)
	if err != nil {
		return nil, err
	}
	term := requestedTerm
	if term == "" {
		term = selectedValue(doc, "xnxqdpl")
	}
	if term == "" {
		return nil, fmt.Errorf("classroom usage page has no selected academic term")
	}
	if selectedValue(doc, "xnxqdpl") != term {
		doc, err = s.postClassroomUsage(ctx, doc, date, term, selectedValue(doc, "xqdpl"), "", "ctl00$xnxqdpl")
		if err != nil {
			return nil, err
		}
	}
	if selectedValue(doc, "xqdpl") != campus || safeExtractionValue(doc, "//*[@id='TB_rq']", "value", 0) != date {
		doc, err = s.postClassroomUsage(ctx, doc, date, term, campus, "", "ctl00$xqdpl")
		if err != nil {
			return nil, fmt.Errorf("select campus %s: %w", campus, err)
		}
	}
	if selectedValue(doc, "xqdpl") != campus ||
		selectedValue(doc, "xnxqdpl") != term ||
		safeExtractionValue(doc, "//*[@id='TB_rq']", "value", 0) != date {
		return nil, fmt.Errorf("classroom usage page did not select date %s, term %s, campus %s", date, term, campus)
	}
	buildings := make([]string, 0)
	seen := make(map[string]bool)
	for _, option := range htmlquery.Find(doc, "//*[@id='jxldpl']/option") {
		building := htmlquery.SelectAttr(option, "value")
		if building != "" && !seen[building] {
			buildings = append(buildings, building)
			seen[building] = true
		}
	}
	if len(buildings) == 0 {
		return nil, fmt.Errorf("classroom usage page has no buildings for %s", campus)
	}
	return &classroomUsagePageState{
		document:  doc,
		term:      term,
		buildings: buildings,
	}, nil
}

func (s *Student) postClassroomUsage(ctx context.Context, doc *html.Node, date, term, campus, building, target string) (*html.Node, error) {
	form := map[string]string{
		"__EVENTTARGET": target, "__EVENTARGUMENT": "", "__LASTFOCUS": "",
		"ctl00$TB_rq": date, "ctl00$xnxqdpl": term, "ctl00$xqdpl": campus, "ctl00$jxldpl": building,
		"ctl00$qszdpl": "", "ctl00$qsxqdpl": "", "ctl00$qsjdpl": "", "ctl00$zzjdpl": "",
	}
	if target == "" {
		form["ctl00$ContentPlaceHolder1$BT_search"] = "查询"
	}
	for _, option := range htmlquery.Find(doc, "//*[@id='jsdpl']/option") {
		if htmlquery.SelectAttr(option, "value") == "" {
			form["ctl00$jsdpl"] = ""
			break
		}
	}
	for _, input := range htmlquery.Find(doc, "//input[@type='hidden']") {
		name := htmlquery.SelectAttr(input, "name")
		if strings.HasPrefix(name, "__VIEWSTATE") || name == "__EVENTVALIDATION" {
			form[name] = htmlquery.SelectAttr(input, "value")
		}
	}
	if form["__VIEWSTATE"] == "" || form["__EVENTVALIDATION"] == "" {
		return nil, fmt.Errorf("classroom usage page is missing ASP.NET state (session may have expired)")
	}
	if err := s.waitForRequest(ctx); err != nil {
		return nil, err
	}
	doc, err := s.PostWithIdentifier(classroomUsageURL, form)
	if err != nil && target == "" {
		return nil, fmt.Errorf("query building %s: %w", building, err)
	}
	return doc, err
}

const (
	sectionsPerDay = 11
	daysPerWeek    = 7
	headerRows     = 3
	gridColumns    = 1 + sectionsPerDay*daysPerWeek
)

func parseClassroomUsage(doc *html.Node, weekday time.Weekday) ([]*ClassroomUsageRoom, error) {
	table := htmlquery.FindOne(doc, "//*[@id='ContentPlaceHolder1_LB_kb']/table")
	if table == nil {
		return nil, fmt.Errorf("classroom usage table is missing")
	}
	// 先展开跨行、跨列单元格
	rows := htmlquery.Find(table, "./tr | ./tbody/tr")
	grid := make([][]*html.Node, len(rows))
	for row := range grid {
		grid[row] = make([]*html.Node, gridColumns)
	}
	for row, tr := range rows {
		column := 0
		for _, cell := range htmlquery.Find(tr, "./td | ./th") {
			for column < gridColumns && grid[row][column] != nil {
				column++
			}
			spans := [2]int{1, 1}
			for i, name := range [...]string{"rowspan", "colspan"} {
				if value := htmlquery.SelectAttr(cell, name); value != "" {
					spans[i] = utils.SafeAtoi(value)
					if spans[i] < 1 {
						return nil, fmt.Errorf("invalid classroom table %s", name)
					}
				}
			}
			rowSpan, colSpan := spans[0], spans[1]
			if row+rowSpan > len(rows) || column+colSpan > gridColumns {
				return nil, fmt.Errorf("classroom table span exceeds grid at row %d", row)
			}
			for r := row; r < row+rowSpan; r++ {
				for c := column; c < column+colSpan; c++ {
					if grid[r][c] != nil {
						return nil, fmt.Errorf("overlapping classroom table spans at row %d", r)
					}
					grid[r][c] = cell
				}
			}
			column += colSpan
		}
		for _, cell := range grid[row] {
			if cell == nil {
				return nil, fmt.Errorf("incomplete classroom table row %d", row)
			}
		}
	}
	// 校验表头，确保展开后的列与星期、节次对应
	if len(grid) < headerRows {
		return nil, fmt.Errorf("classroom usage table has no complete header")
	}
	for day := 0; day < daysPerWeek; day++ {
		for section := 1; section <= sectionsPerDay; section++ {
			column := day*sectionsPerDay + section
			if normalizedHTMLText(grid[0][column]) != fmt.Sprintf("周%d", day+1) ||
				normalizedHTMLText(grid[headerRows-1][column]) != strconv.Itoa(section) {
				return nil, fmt.Errorf("unexpected classroom usage header at column %d", column)
			}
		}
	}
	metadata := make(map[string]*ClassroomUsageRoom)
	for _, option := range htmlquery.Find(doc, "//*[@id='jsdpl']/option") {
		location := htmlquery.SelectAttr(option, "value")
		if location == "" {
			continue
		}
		fields := strings.Fields(htmlquery.InnerText(option))
		if len(fields) < 3 || fields[0] != location {
			return nil, fmt.Errorf("invalid classroom metadata for %s", location)
		}
		metadata[location] = &ClassroomUsageRoom{
			Location: fields[0],
			Capacity: strings.Split(fields[1], "(")[0],
			Type:     strings.Join(fields[2:], " "),
		}
	}
	day := int(weekday)
	if weekday == time.Sunday {
		day = daysPerWeek
	}
	if day < 1 || day > daysPerWeek {
		return nil, fmt.Errorf("invalid weekday %d", weekday)
	}
	rooms := make([]*ClassroomUsageRoom, 0, len(metadata))
	seen := make(map[string]bool)
	for row := headerRows; row < len(grid); {
		label := grid[row][0]
		if label == nil {
			return nil, fmt.Errorf("missing classroom label at row %d", row)
		}
		fields := strings.Fields(normalizedHTMLText(label))
		if len(fields) == 0 {
			return nil, fmt.Errorf("empty classroom label at row %d", row)
		}
		room, ok := metadata[fields[0]]
		if !ok || seen[room.Location] {
			return nil, fmt.Errorf("missing or duplicate classroom metadata for %s", fields[0])
		}
		span := 1
		if value := htmlquery.SelectAttr(label, "rowspan"); value != "" {
			span = utils.SafeAtoi(value)
		}
		if span < 1 || row+span > len(grid) {
			return nil, fmt.Errorf("invalid classroom row span at row %d", row)
		}
		seen[room.Location] = true
		room.Occupancies = make([]*ClassroomUsageOccupancy, 0)
		roomRows := grid[row : row+span]
		// 同一教室可能有多组课程行和班级行，分别保留重叠的占用记录
		for i := 0; i < len(roomRows); i += 2 {
			var previous *ClassroomUsageOccupancy
			for section := 1; section <= sectionsPerDay; section++ {
				column := (day-1)*sectionsPerDay + section
				raw := normalizedHTMLText(roomRows[i][column])
				detail := ""
				if i+1 < len(roomRows) {
					detail = normalizedHTMLText(roomRows[i+1][column])
				}
				if raw == "" && detail == "" {
					previous = nil
					continue
				}
				// 讲座、补课等连续活动可能只在起始节次写班级或使用单位，后续节次详情为空时仍可合并到前一条占用记录
				if previous != nil && previous.RawText == raw && (detail == "" || detail == previous.Detail) {
					previous.End = int32(section)
					continue
				}
				reason := raw
				if font := htmlquery.FindOne(roomRows[i][column], ".//font"); font != nil {
					reason = strings.TrimSuffix(normalizedHTMLText(font), "()")
				}
				if reason == "" {
					reason = detail
				}
				previous = &ClassroomUsageOccupancy{
					Start: int32(section), End: int32(section), Reason: reason, Detail: detail, RawText: raw,
				}
				room.Occupancies = append(room.Occupancies, previous)
			}
		}

		rooms = append(rooms, room)
		row += span
	}
	if len(rooms) != len(metadata) {
		return nil, fmt.Errorf("classroom table is incomplete: got %d of %d rooms", len(rooms), len(metadata))
	}
	return rooms, nil
}
