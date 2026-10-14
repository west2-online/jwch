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
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
)

// 测试页面按实际 kbcx_zb.aspx 表格构造：三行表头、
// 77 列星期和节次、跨行教室标签，以及课程行和详情行。
func usageFixture() string {
	var body strings.Builder
	body.WriteString(`<select id="jsdpl"><option value=""></option>`)
	for _, room := range []string{"旗山东2-100", "旗山东2-101", "旗山东2-102"} {
		fmt.Fprintf(&body, `<option value="%s">%s 318(159) 智慧教室普通型</option>`, room, room)
	}
	body.WriteString(`</select><span id="ContentPlaceHolder1_LB_kb"><table><tr><td rowspan="3">时间课程<br>班级教室</td>`)
	for day := 1; day <= 7; day++ {
		fmt.Fprintf(&body, `<td colspan="11">周%d</td>`, day)
	}
	body.WriteString(`</tr><tr>`)
	for day := 1; day <= 7; day++ {
		body.WriteString(`<td colspan="4">上午</td><td colspan="4">下午</td><td colspan="3">晚上</td>`)
	}
	body.WriteString(`</tr><tr>`)
	for day := 1; day <= 7; day++ {
		for section := 1; section <= 11; section++ {
			fmt.Fprintf(&body, `<td>%d</td>`, section)
		}
	}
	body.WriteString(`</tr>`)
	for _, room := range []string{"旗山东2-100", "旗山东2-101", "旗山东2-102"} {
		fmt.Fprintf(&body, `<tr><td rowspan="2">%s<br>(318)</td>`, room)
		for day := 1; day <= 7; day++ {
			for section := 1; section <= 11; section++ {
				content := "&nbsp;"
				if room == "旗山东2-100" && day == 7 && section <= 4 {
					content = `<font color="blue">讲座()</font>(06-06)`
				}
				if room == "旗山东2-101" && day == 1 && section == 11 {
					content = `<font color="blue">课程(教师)</font>(01-20)`
				}
				fmt.Fprintf(&body, `<td>%s</td>`, content)
			}
		}
		body.WriteString(`</tr><tr>`)
		for day := 1; day <= 7; day++ {
			for section := 1; section <= 11; section++ {
				content := "&nbsp;"
				if room == "旗山东2-100" && day == 7 && section == 1 {
					content = "先造"
				}
				if room == "旗山东2-101" && day == 1 && section == 11 {
					content = "2026级某班"
				}
				fmt.Fprintf(&body, `<td>%s</td>`, content)
			}
		}
		body.WriteString(`</tr>`)
	}
	body.WriteString(`</table></span>`)
	return body.String()
}

func parseFixture(t *testing.T, source string) *html.Node {
	t.Helper()
	doc, err := htmlquery.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestClassroomUsageSelectedDate(t *testing.T) {
	doc := parseFixture(t, usageFixture())
	rooms, err := parseClassroomUsage(doc, time.Sunday)
	if err != nil || len(rooms) != 3 {
		t.Fatalf("rooms=%v, err=%v", rooms, err)
	}
	want := []*ClassroomUsageOccupancy{{Start: 1, End: 4, Reason: "讲座", Detail: "先造", RawText: "讲座()(06-06)"}}
	if !reflect.DeepEqual(rooms[0].Occupancies, want) {
		t.Fatalf("Sunday activity: got %+v, want %+v", rooms[0].Occupancies, want)
	}
	if len(rooms[1].Occupancies) != 0 || len(rooms[2].Occupancies) != 0 {
		t.Fatal("other weekday columns must not contribute to Sunday's availability")
	}
	if rooms[0].Location != "旗山东2-100" || rooms[0].Capacity != "318" || rooms[0].Type != "智慧教室普通型" {
		t.Fatalf("missing legacy metadata: %+v", rooms[0])
	}
	monday, err := parseClassroomUsage(doc, time.Monday)
	if err != nil || len(monday[0].Occupancies) != 0 || len(monday[1].Occupancies) != 1 {
		t.Fatalf("Monday: rooms=%v, err=%v", monday, err)
	}
	if monday[1].Occupancies[0].Start != 11 || monday[1].Occupancies[0].Detail != "2026级某班" {
		t.Fatalf("Monday last section: %+v", monday[1].Occupancies[0])
	}
}

func TestClassroomUsageRejectsIncompletePages(t *testing.T) {
	for name, source := range map[string]string{
		"missing table":  "<html>重新登录</html>",
		"wrong header":   strings.Replace(usageFixture(), "周7", "周8", 1),
		"missing cell":   strings.Replace(usageFixture(), "<td>&nbsp;</td>", "", 1),
		"wrong metadata": strings.Replace(usageFixture(), "318(159) 智慧教室普通型", "318(159)", 1),
		"wrong span":     strings.Replace(usageFixture(), `rowspan="2"`, `rowspan="999"`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseClassroomUsage(parseFixture(t, source), time.Sunday); err == nil {
				t.Fatal("malformed upstream data must not be published as free classrooms")
			}
		})
	}
}

func TestClassroomUsageFormOmitsUnpopulatedRoomSelect(t *testing.T) {
	for _, tc := range []struct {
		name, options string
		wantRoomField bool
	}{
		{name: "initial page without options"},
		{name: "building selected", options: `<option value=""></option><option value="教室1">教室1</option>`, wantRoomField: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc := parseFixture(t, `<input type="hidden" name="__VIEWSTATE" value="state">
<input type="hidden" name="__EVENTVALIDATION" value="validation"><select id="jsdpl">`+tc.options+`</select>`)
			form, err := queryForm(doc, "2026-10-11", "202601", "旗山校区", "第1学科群")
			if err != nil {
				t.Fatal(err)
			}
			if _, present := form["ctl00$jsdpl"]; present != tc.wantRoomField {
				t.Fatalf("room field present=%v, want %v", present, tc.wantRoomField)
			}
		})
	}
}

type fixtureUsageClient struct {
	t             *testing.T
	state         int
	campus        string
	building      string
	queries       int
	fail          bool
	emptyBuilding bool
}

func (c *fixtureUsageClient) page(date, term, table string) *html.Node {
	return parseFixture(c.t, fmt.Sprintf(`<input type="hidden" name="__VIEWSTATE" value="state%d">
<input type="hidden" name="__VIEWSTATEGENERATOR" value="generator%d">
<input type="hidden" name="__EVENTVALIDATION" value="validation%d">
<input id="TB_rq" value="%s"><select id="xnxqdpl"><option selected value="%s">%s</option></select>
<select id="xqdpl"><option selected value="%s">%s</option></select>
<select id="jxldpl"><option selected value="%s">%s</option><option value="楼A">楼A</option><option value="楼B">楼B</option></select>%s`,
		c.state, c.state, c.state, date, term, term, c.campus, c.campus, c.building, c.building, table))
}

func (c *fixtureUsageClient) Get(context.Context) (*html.Node, error) {
	c.campus = "旗山校区"
	return c.page("", "202601", ""), nil
}

func (c *fixtureUsageClient) Post(_ context.Context, form map[string]string) (*html.Node, error) {
	if form["__VIEWSTATE"] != fmt.Sprintf("state%d", c.state) ||
		form["__EVENTVALIDATION"] != fmt.Sprintf("validation%d", c.state) ||
		form["__VIEWSTATEGENERATOR"] != fmt.Sprintf("generator%d", c.state) {
		c.t.Fatal("ASP.NET state from the previous response must be reused")
	}
	for _, key := range []string{"ctl00$qszdpl", "ctl00$qsxqdpl", "ctl00$qsjdpl", "ctl00$zzjdpl", "ctl00$jsdpl", "ctl00$jslxdpl"} {
		if form[key] != "" {
			c.t.Fatalf("unexpected query dimension %s=%s", key, form[key])
		}
	}
	c.state++
	c.campus = form["ctl00$xqdpl"]
	c.building = form["ctl00$jxldpl"]
	table := ""
	if c.building != "" {
		table = strings.Split(usageFixture(), `<span id="ContentPlaceHolder1_LB_kb">`)[0]
	}
	if c.emptyBuilding && c.building == "楼B" {
		table = `<select id="jsdpl"><option value=""></option></select>`
	}
	if form["ctl00$ContentPlaceHolder1$BT_search"] == "查询" {
		if form["__EVENTTARGET"] != "" || form["ctl00$TB_rq"] != "2026-10-11" || form["ctl00$xnxqdpl"] != "202601" {
			c.t.Fatal("wrong query condition")
		}
		c.queries++
		table = strings.ReplaceAll(usageFixture(), "旗山东2-", c.building+"-")
		if c.fail {
			table = ""
		}
	}
	return c.page(form["ctl00$TB_rq"], form["ctl00$xnxqdpl"], table), nil
}

func TestClassroomUsageCrawlsBuildings(t *testing.T) {
	client := &fixtureUsageClient{t: t}
	term, rooms, err := fetchClassroomUsage(context.Background(), client, "2026-10-11", "", "铜盘校区", "")
	if err != nil || term != "202601" || len(rooms) != 6 || client.queries != 2 {
		t.Fatalf("term=%s rooms=%d queries=%d err=%v", term, len(rooms), client.queries, err)
	}
	if rooms[0].Building != "楼A" || rooms[3].Building != "楼B" {
		t.Fatal("building identity missing")
	}
	client = &fixtureUsageClient{t: t, emptyBuilding: true}
	if _, rooms, err := fetchClassroomUsage(context.Background(), client, "2026-10-11", "", "旗山校区", ""); err != nil || len(rooms) != 3 {
		t.Fatalf("building with no classrooms should be skipped: rooms=%d err=%v", len(rooms), err)
	}
	client = &fixtureUsageClient{t: t, fail: true}
	_, rooms, err = fetchClassroomUsage(context.Background(), client, "2026-10-11", "", "旗山校区", "")
	var usageErr *ClassroomUsageError
	if !errors.As(err, &usageErr) || rooms == nil || len(usageErr.FailedBuildings) == 0 {
		t.Fatalf("failed buildings should be reported with partial rooms: rooms=%d err=%v", len(rooms), err)
	}
	client = &fixtureUsageClient{t: t}
	_, rooms, err = fetchClassroomUsage(context.Background(), client, "2026-10-11", "", "铜盘校区", "楼B")
	if err != nil || len(rooms) != 3 || client.queries != 1 {
		t.Fatalf("single building query should only request that building: rooms=%d queries=%d err=%v", len(rooms), client.queries, err)
	}
}

type usageRoundTripper func(*http.Request) (*http.Response, error)

func (f usageRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestGetClassroomUsageReusesAuthentication(t *testing.T) {
	pages := &fixtureUsageClient{t: t}
	client := resty.New().SetTransport(usageRoundTripper(func(request *http.Request) (*http.Response, error) {
		if request.URL.Query().Get("id") != "existing-session" || request.URL.Path != "/kkgl/kbcx/kbcx_zb.aspx" {
			t.Fatal("request did not use the usage page with the existing session id")
		}
		cookie, err := request.Cookie("session")
		if err != nil || cookie.Value != "existing-cookie" || request.Header.Get("Referer") == "" {
			t.Fatal("request did not reuse cookies and referer")
		}
		var doc *html.Node
		if request.Method == http.MethodGet {
			doc, err = pages.Get(request.Context())
		} else {
			if err := request.ParseForm(); err != nil {
				t.Fatal(err)
			}
			form := make(map[string]string)
			for key := range request.PostForm {
				form[key] = request.PostForm.Get(key)
			}
			doc, err = pages.Post(request.Context(), form)
		}
		if err != nil {
			return nil, err
		}
		var body strings.Builder
		if err := html.Render(&body, doc); err != nil {
			return nil, err
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body.String())), Request: request}, nil
	}))
	student := (&Student{client: client}).WithLoginData("existing-session", []*http.Cookie{{Name: "session", Value: "existing-cookie"}})
	usage, err := student.GetClassroomUsage(context.Background(), ClassroomUsageReq{Date: "2026-10-11", Campus: "旗山校区"})
	if err != nil || usage == nil || usage.Term != "202601" || len(usage.Rooms) != 6 {
		t.Fatalf("usage=%v err=%v", usage, err)
	}
}
