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
)

const classroomUsageURL = "https://jwcjwxt2.fzu.edu.cn:81/kkgl/kbcx/kbcx_zb.aspx"

// 每次教室查询 GET/POST 前等待 200 毫秒，降低连续请求对教务处的压力。
const classroomUsageRequestDelay = 200 * time.Millisecond

func (e *ClassroomUsageError) Error() string {
	return fmt.Sprintf("classroom usage failed for %s buildings %s: %v", e.Campus, strings.Join(e.FailedBuildings, ", "), e.Cause)
}

func (e *ClassroomUsageError) Unwrap() error { return e.Cause }

// GetClassroomUsage 复用学生已有鉴权，按教学楼抓取占用表，
// 无需再遍历教室类型或查询节次组合。
func (s *Student) GetClassroomUsage(ctx context.Context, req ClassroomUsageReq) (*ClassroomUsage, error) {
	term, rooms, err := fetchClassroomUsage(ctx, s, req.Date, req.Term, req.Campus, req.Building)
	usage := &ClassroomUsage{Date: req.Date, Term: term, Rooms: rooms}
	if err != nil {
		return usage, err
	}
	return usage, nil
}

// fetchClassroomUsage 按日期遍历校区内的教学楼，查询每栋楼的教室占用表。
// 切换校区和教学楼时先回发选择事件，取得联动下拉框及最新页面状态。
func fetchClassroomUsage(ctx context.Context, s *Student, date, requestedTerm, campus, requestedBuilding string) (string, []*ClassroomUsageRoom, error) {
	queryDate, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return "", nil, fmt.Errorf("invalid classroom date: %w", err)
	}
	if err := waitForClassroomUsageRequest(ctx); err != nil {
		return "", nil, err
	}
	doc, err := s.GetWithIdentifier(classroomUsageURL)
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
		doc, err = postSelection(ctx, s, doc, date, term, campus, "", "ctl00$xqdpl")
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
		buildingDoc, buildingErr := postSelection(ctx, s, doc, date, term, campus, building, "ctl00$jxldpl")
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
		if err := waitForClassroomUsageRequest(ctx); err != nil {
			return term, rooms, err
		}
		buildingDoc, err = s.PostWithIdentifier(classroomUsageURL, form)
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

func postSelection(ctx context.Context, s *Student, doc *html.Node,
	date, term, campus, building, target string,
) (*html.Node, error) {
	form, err := queryForm(doc, date, term, campus, building)
	if err != nil {
		return nil, err
	}
	form["__EVENTTARGET"] = target
	if err := waitForClassroomUsageRequest(ctx); err != nil {
		return nil, err
	}
	return s.PostWithIdentifier(classroomUsageURL, form)
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

// waitForClassroomUsageRequest 只控制请求间隔；请求和鉴权使用 Student 的现有方法。
func waitForClassroomUsageRequest(ctx context.Context) error {
	timer := time.NewTimer(classroomUsageRequestDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return ctx.Err()
	}
}

const (
	sectionsPerDay = 11
	daysPerWeek    = 7
	headerRows     = 3
	gridColumns    = 1 + sectionsPerDay*daysPerWeek
)

// parseClassroomUsage 只读取查询日期对应的星期列。教务处已按日期筛选，
// 其他星期列为空不代表教室空闲，不能用来判断其他日期的占用情况。
func parseClassroomUsage(doc *html.Node, weekday time.Weekday) ([]*ClassroomUsageRoom, error) {
	table := htmlquery.FindOne(doc, "//*[@id='ContentPlaceHolder1_LB_kb']/table")
	if table == nil {
		return nil, fmt.Errorf("classroom usage table is missing")
	}
	grid, err := expandTable(table)
	if err != nil {
		return nil, err
	}
	if len(grid) < headerRows {
		return nil, fmt.Errorf("classroom usage table has no complete header")
	}
	for day := 0; day < daysPerWeek; day++ {
		for section := 1; section <= sectionsPerDay; section++ {
			column := day*sectionsPerDay + section
			if cellText(grid[0][column]) != fmt.Sprintf("周%d", day+1) ||
				cellText(grid[headerRows-1][column]) != strconv.Itoa(section) {
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
		room := parseUsageRoomMetadata(htmlquery.InnerText(option))
		if room == nil || room.Location != location {
			return nil, fmt.Errorf("invalid classroom metadata for %s", location)
		}
		metadata[location] = room
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
		fields := strings.Fields(cellText(label))
		if len(fields) == 0 {
			return nil, fmt.Errorf("empty classroom label at row %d", row)
		}
		room, ok := metadata[fields[0]]
		if !ok || seen[room.Location] {
			return nil, fmt.Errorf("missing or duplicate classroom metadata for %s", fields[0])
		}
		span, err := cellSpan(label, "rowspan")
		if err != nil || row+span > len(grid) {
			return nil, fmt.Errorf("invalid classroom row span at row %d", row)
		}
		seen[room.Location] = true
		room.Occupancies = parseOccupancies(grid[row:row+span], day)

		rooms = append(rooms, room)
		row += span
	}
	if len(rooms) != len(metadata) {
		return nil, fmt.Errorf("classroom table is incomplete: got %d of %d rooms", len(rooms), len(metadata))
	}
	return rooms, nil
}

func parseOccupancies(rows [][]*html.Node, day int) []*ClassroomUsageOccupancy {
	occupancies := make([]*ClassroomUsageOccupancy, 0)
	// 同一教室可能有多组课程行和班级行，分别保留重叠的占用记录。
	for row := 0; row < len(rows); row += 2 {
		var previous *ClassroomUsageOccupancy
		for section := 1; section <= sectionsPerDay; section++ {
			column := (day-1)*sectionsPerDay + section
			raw := cellText(rows[row][column])
			detail := ""
			if row+1 < len(rows) {
				detail = cellText(rows[row+1][column])
			}
			if raw == "" && detail == "" {
				previous = nil
				continue
			}
			// 讲座、补课等连续活动可能只在起始节次写班级或使用单位，
			// 后续节次详情为空时仍可合并到前一条占用记录。
			if previous != nil && previous.RawText == raw && (detail == "" || detail == previous.Detail) {
				previous.End = int32(section)
				continue
			}
			reason := raw
			if font := htmlquery.FindOne(rows[row][column], ".//font"); font != nil {
				reason = strings.TrimSuffix(cellText(font), "()")
			}
			if reason == "" {
				reason = detail
			}
			previous = &ClassroomUsageOccupancy{
				Start: int32(section), End: int32(section), Reason: reason, Detail: detail, RawText: raw,
			}
			occupancies = append(occupancies, previous)
		}
	}
	return occupancies
}

// expandTable 先展开跨行、跨列单元格，再按星期和节次定位，
// 避免详情行省略跨行教室标签导致列下标错位。
func expandTable(table *html.Node) ([][]*html.Node, error) {
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
			rowSpan, err := cellSpan(cell, "rowspan")
			if err != nil {
				return nil, err
			}
			colSpan, err := cellSpan(cell, "colspan")
			if err != nil {
				return nil, err
			}
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
	return grid, nil
}

func cellSpan(cell *html.Node, name string) (int, error) {
	value := htmlquery.SelectAttr(cell, name)
	if value == "" {
		return 1, nil
	}
	span, err := strconv.Atoi(value)
	if err != nil || span < 1 {
		return 0, fmt.Errorf("invalid classroom table %s", name)
	}
	return span, nil
}

func cellText(cell *html.Node) string {
	if cell == nil {
		return ""
	}
	var text strings.Builder
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
		} else if node.Type == html.ElementNode && node.Data == "br" {
			text.WriteByte(' ')
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(cell)
	return strings.Join(strings.Fields(text.String()), " ")
}

func parseUsageRoomMetadata(text string) *ClassroomUsageRoom {
	const metadataFields = 3
	fields := strings.Fields(text)
	if len(fields) < metadataFields {
		return nil
	}
	return &ClassroomUsageRoom{
		Location: fields[0],
		Capacity: strings.Split(fields[1], "(")[0],
		Type:     strings.Join(fields[2:], " "),
	}
}
