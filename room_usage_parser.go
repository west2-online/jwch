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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

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
