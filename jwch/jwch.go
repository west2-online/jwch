package jwch

import (
	"bytes"
	"crypto/tls"
	"net/http"
	"strings"
	"test/errno"

	"github.com/antchfx/htmlquery"
	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
)

func NewStudent() *Student {
	// Disable HTTP/2.0
	client := resty.New().SetTransport(&http.Transport{TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)})

	return &Student{
		client: client,
	}
}

// 携带账号密码，这部分考虑整合到Login中，因为实际上我们不需要这个东西
func (s *Student) WithUser(id, password string) *Student {
	s.ID = id
	s.Password = password
	return s
}

// Add Session to student object
func (s *Student) WithSession(session string) *Student {
	s.Session = session
	return s
}

func (s *Student) GetCookies() []*http.Cookie {
	return s.client.Cookies
}

// GetWithSession returns parse tree for the resp of the request.
func (s *Student) GetWithSession(url string) (*html.Node, error) {
	resp, err := s.client.R().SetHeader("Referer", "https://jwcjwxt1.fzu.edu.cn/").SetQueryParam("id", s.Session).Get(url)

	if err != nil {
		return nil, errno.HTTPQueryError.WithErr(err)
	}

	// 会话过期 TODO: 判断条件有点简陋
	if strings.Contains(string(resp.Body()), "重新登录") {
		return nil, errno.SessionExpiredError
	}

	return htmlquery.Parse(bytes.NewReader(resp.Body()))
}

// PostWithSession returns parse tree for the resp of the request.
func (s *Student) PostWithSession(url string, formdata map[string]string) (*html.Node, error) {
	resp, err := s.client.R().SetHeader("Referer", "https://jwcjwxt1.fzu.edu.cn/").SetQueryParam("id", s.Session).SetFormData(formdata).Post(url)

	if err != nil {
		return nil, errno.HTTPQueryError.WithErr(err)
	}

	// 会话过期 TODO: 判断条件有点简陋
	if strings.Contains(string(resp.Body()), "处理URL失败") {
		return nil, errno.SessionExpiredError
	}

	// utils.SaveData("test.html", resp.Body()) // 保存到本地查看网页

	// return htmlquery.Parse(bytes.NewReader(resp.Body()))

	return htmlquery.Parse(strings.NewReader(strings.TrimSpace(string(resp.Body()))))
}
