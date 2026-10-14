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
)

// ClassroomUsageReq 查询指定日期和校区内所有教学楼的教室占用情况。
// Term 可省略，默认使用教务处页面当前选中的学年学期。
type ClassroomUsageReq struct {
	Date     string
	Campus   string
	Term     string
	Building string // 可选；为空时查询校区内全部教学楼。
}

// ClassroomUsageError 表示部分教学楼抓取失败，但其他教学楼可能已有结果。
type ClassroomUsageError struct {
	Campus          string
	FailedBuildings []string
	Cause           error
}

func (e *ClassroomUsageError) Error() string {
	return fmt.Sprintf("classroom usage failed for %s buildings %s: %v", e.Campus, strings.Join(e.FailedBuildings, ", "), e.Cause)
}

func (e *ClassroomUsageError) Unwrap() error { return e.Cause }

type ClassroomUsage struct {
	Date  string                `json:"date"`
	Term  string                `json:"term"`
	Rooms []*ClassroomUsageRoom `json:"rooms"`
}

type ClassroomUsageRoom struct {
	Building    string                     `json:"building"`
	Location    string                     `json:"location"`
	Capacity    string                     `json:"capacity"`
	Type        string                     `json:"type"`
	Occupancies []*ClassroomUsageOccupancy `json:"occupancies"`
}

// Reason、Detail 分别保留教务处的课程或活动名称、班级或使用单位。
// RawText 还保留原始周次标注；页面已按 Date 筛选，
// 无需再根据周次推算课程是否发生在查询日期。
type ClassroomUsageOccupancy struct {
	Start   int32  `json:"start"`
	End     int32  `json:"end"`
	Reason  string `json:"reason"`
	Detail  string `json:"detail"`
	RawText string `json:"raw_text"`
}

// GetClassroomUsage 复用学生已有鉴权，按教学楼抓取占用表，
// 无需再遍历教室类型或查询节次组合。
func (s *Student) GetClassroomUsage(ctx context.Context, req ClassroomUsageReq) (*ClassroomUsage, error) {
	term, rooms, err := fetchClassroomUsage(ctx, usageStudentClient{Student: s}, req.Date, req.Term, req.Campus, req.Building)
	usage := &ClassroomUsage{Date: req.Date, Term: term, Rooms: rooms}
	if err != nil {
		return usage, err
	}
	return usage, nil
}
