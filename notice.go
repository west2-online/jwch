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
	"net/url"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"

	"github.com/west2-online/jwch/constants"
)

func (s *Student) GetNoticeInfo(req *NoticeInfoReq) (list []*NoticeInfo, totalPages int, err error) {
	// 获取通知公告页面的总页数
	res, err := s.NewRequest().
		SetHeader("User-Agent", constants.UserAgent).
		Get(constants.NoticeInfoQueryURL)
	if err != nil {
		return nil, 0, err
	}

	doc, err := htmlquery.Parse(strings.NewReader(string(res.Body())))
	if err != nil {
		return nil, 0, err
	}

	// 获取总页数
	lastPageNum, err := getTotalPages(doc)
	if err != nil {
		return nil, 0, err
	}
	// 判断是否超出总页数
	if req.PageNum > lastPageNum {
		return nil, lastPageNum, fmt.Errorf("超出总页数")
	}
	// 首页直接爬取
	if req.PageNum == 1 {
		list, err = parseNoticeInfo(doc)
		if err != nil {
			return nil, lastPageNum, err
		}
		return list, lastPageNum, nil
	}
	// 根据总页数计算 url
	num := lastPageNum - req.PageNum + 1
	url := fmt.Sprintf("https://jwch.fzu.edu.cn/jxtz/%d.htm", num)
	resp, err := s.NewRequest().
		SetHeader("User-Agent", constants.UserAgent).
		Get(url)
	if err != nil {
		return nil, lastPageNum, err
	}

	doc, err = htmlquery.Parse(strings.NewReader(string(resp.Body())))
	if err != nil {
		return nil, lastPageNum, err
	}
	list, err = parseNoticeInfo(doc)
	if err != nil {
		return nil, lastPageNum, err
	}
	// 3. 返回结果
	return list, lastPageNum, nil
}

// 获取当前页面的所有数据信息
func parseNoticeInfo(doc *html.Node) ([]*NoticeInfo, error) {
	// 解析通知公告页面
	var list []*NoticeInfo

	sel := htmlquery.FindOne(doc, "//div[@class='box-gl clearfix']")
	if sel == nil {
		return nil, fmt.Errorf("cannot find the notice list")
	}

	rows := htmlquery.Find(sel, ".//ul[@class='list-gl']/li")

	for _, row := range rows {
		// 提取日期
		dateNode := htmlquery.FindOne(row, ".//span[@class='doclist_time']")
		if dateNode == nil {
			return nil, fmt.Errorf("cannot find the date")
		}
		date := strings.TrimSpace(htmlquery.InnerText(dateNode))

		// 提取标题
		titleNode := htmlquery.FindOne(row, ".//a")

		title := strings.TrimSpace(htmlquery.SelectAttr(titleNode, "title"))

		// 提取 URL
		rawURL := strings.TrimSpace(htmlquery.SelectAttr(titleNode, "href"))
		rawURL = constants.JwchNoticeURLPrefix + rawURL

		convertedURL, wbTreeId, wbNewsId := convertURL(rawURL)

		noticeInfo := &NoticeInfo{
			Title:    title,
			URL:      convertedURL,
			Date:     date,
			WbTreeId: wbTreeId,
			WbNewsId: wbNewsId,
		}
		list = append(list, noticeInfo)
	}

	return list, nil
}

// 获取总页数
func getTotalPages(doc *html.Node) (int, error) {
	totalPagesNode := htmlquery.FindOne(doc, "//span[@class='p_pages']//a[@href='jxtz/1.htm']")
	if totalPagesNode == nil {
		return 0, fmt.Errorf("未找到总页数")
	}

	totalPagesStr := htmlquery.InnerText(totalPagesNode)
	var totalPages int
	_, err := fmt.Sscanf(totalPagesStr, "%d", &totalPages)
	if err != nil {
		return 0, fmt.Errorf("解析总页数失败: %v", err)
	}
	return totalPages, nil
}

// 将通知公告列表中的 URL 转换成 content.jsp 格式，并提取 wbtreeid 和 wbnewsid
//
// 例：将
//   - https://jwch.fzu.edu.cn/../info/1040/13769.htm
//   - https://jwch.fzu.edu.cn/info/1040/13769.htm
//   - https://jwch.fzu.edu.cn/../content.jsp?urltype=news.NewsContentUrl&wbtreeid=1040&wbnewsid=13769
//
// 转换成
// - https://jwch.fzu.edu.cn/content.jsp?urltype=news.NewsContentUrl&wbtreeid=1040&wbnewsid=13769
//
// Returns:
//   - finalURL
//   - wbTreeId
//   - wbNewsId
func convertURL(original string) (string, string, string) {
	// 去除 "../"
	cleaned := strings.ReplaceAll(original, "../", "")

	// 正则提取 wbtreeid 和 wbnewsid（info/TREE/NEWS.htm 格式）
	re := regexp.MustCompile(`info/(\d+)/(\d+)\.htm`)
	matches := re.FindStringSubmatch(cleaned)
	if len(matches) == 3 {
		wbtreeid := matches[1]
		wbnewsid := matches[2]
		newURL := fmt.Sprintf("https://jwch.fzu.edu.cn/content.jsp?urltype=news.NewsContentUrl&wbtreeid=%s&wbnewsid=%s", wbtreeid, wbnewsid)
		return newURL, wbtreeid, wbnewsid
	}

	// 已经是 content.jsp 格式，从 query string 中提取 wbtreeid 和 wbnewsid
	parsed, err := url.Parse(cleaned)
	if err == nil {
		q := parsed.Query()
		wbtreeid := q.Get("wbtreeid")
		wbnewsid := q.Get("wbnewsid")
		if wbtreeid != "" && wbnewsid != "" {
			return cleaned, wbtreeid, wbnewsid
		}
	}

	return cleaned, "", ""
}

// GetNoticeDetail 获取通知正文内容
func (s *Student) GetNoticeDetail(req *NoticeDetailReq) (*NoticeDetail, error) {
	targetURL := fmt.Sprintf("https://jwch.fzu.edu.cn/content.jsp?urltype=news.NewsContentUrl&wbtreeid=%s&wbnewsid=%s", req.WbTreeId, req.WbNewsId)

	res, err := s.NewRequest().
		SetHeader("User-Agent", constants.UserAgent).
		Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("GetNoticeDetail: failed to fetch url %s: %w", targetURL, err)
	}

	doc, err := htmlquery.Parse(strings.NewReader(string(res.Body())))
	if err != nil {
		return nil, fmt.Errorf("GetNoticeDetail: failed to parse html: %w", err)
	}

	// 主容器
	mainNode := htmlquery.FindOne(doc, "//div[contains(@class,'xl_main')]")
	if mainNode == nil {
		return nil, fmt.Errorf("GetNoticeDetail: .xl_main not found, url=%s", targetURL)
	}

	// 提取标题
	titleNode := htmlquery.FindOne(mainNode, ".//*[contains(@class,'xl_tit')]/h4")
	if titleNode == nil {
		return nil, fmt.Errorf("GetNoticeDetail: .xl_tit h4 not found, url=%s", targetURL)
	}
	title := strings.TrimSpace(htmlquery.InnerText(titleNode))

	// 提取发布时间
	timeNode := htmlquery.FindOne(mainNode, ".//*[contains(@class,'xl_sj')]//span[1]")
	if timeNode == nil {
		return nil, fmt.Errorf("GetNoticeDetail: .xl_sj span not found, url=%s", targetURL)
	}
	date := strings.TrimPrefix(strings.TrimSpace(htmlquery.InnerText(timeNode)), "发布时间：")

	// 提取内容
	contentNode := htmlquery.FindOne(mainNode, ".//*[@id='vsb_content']")
	if contentNode == nil {
		return nil, fmt.Errorf("GetNoticeDetail: #vsb_content not found, url=%s", targetURL)
	}

	return &NoticeDetail{
		NoticeInfo: NoticeInfo{
			Title:    title,
			Date:     date,
			URL:      targetURL,
			WbTreeId: req.WbTreeId,
			WbNewsId: req.WbNewsId,
		},
		Content: htmlquery.InnerText(contentNode),
	}, nil
}
