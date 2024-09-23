package jwch

import (
	"bytes"
	"crypto/tls"
	"net/http"
	"strings"

	"github.com/west2-online/jwch/errno"

	"github.com/antchfx/htmlquery"
	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
)

func NewStudent() *Student {
	// Disable HTTP/2.0
	// Disable Redirect
	client := resty.New().SetTransport(&http.Transport{TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}).SetRedirectPolicy(resty.NoRedirectPolicy())

	return &Student{
		client: client,
	}
}

func (s *Student) WithLoginData(identifier string, cookies []*http.Cookie) *Student {
	s.Identifier = identifier
	s.cookies = cookies
	s.client.SetCookies(cookies)
	return s
}

// 携带账号密码，这部分考虑整合到Login中，因为实际上我们不需要这个东西
func (s *Student) WithUser(id, password string) *Student {
	s.ID = id
	s.Password = password
	return s
}

func (s *Student) SetIdentifier(identifier string) {
	s.Identifier = identifier
}

func (s *Student) SetCookies(cookies []*http.Cookie) {
	s.cookies = cookies
	s.client.SetCookies(cookies)
}

func (s *Student) ClearLoginData() {
	s.cookies = []*http.Cookie{}
	s.client.Cookies = []*http.Cookie{}
}

func (s *Student) NewRequest() *resty.Request {
	return s.client.R()
}

func (s *Student) GetWithIdentifier(url string) (*html.Node, error) {
	resp, err := s.NewRequest().SetHeader("Referer", "https://jwcjwxt1.fzu.edu.cn/").SetQueryParam("id", s.Identifier).Get(url)

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
func (s *Student) PostWithIdentifier(url string, formData map[string]string) (*html.Node, error) {
	resp, err := s.NewRequest().SetHeader("Referer", "https://jwcjwxt1.fzu.edu.cn/").SetQueryParam("id", s.Identifier).SetFormData(formData).Post(url)

	s.NewRequest().EnableTrace()

	if err != nil {
		return nil, errno.HTTPQueryError.WithErr(err)
	}

	// Identifier缺失 TODO: 判断条件有点简陋
	if strings.Contains(string(resp.Body()), "处理URL失败") {
		return nil, errno.SessionExpiredError
	}

	return htmlquery.Parse(strings.NewReader(strings.TrimSpace(string(resp.Body()))))
}
