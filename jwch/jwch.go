package jwch

import (
	"bytes"
	"crypto/tls"
	"jwch/errno"
	"jwch/utils"
	"net/http"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
)

func NewStudent() *Student {
	// Disable HTTP/2.0
	// Disable Redirect
	client := resty.New().SetTransport(&http.Transport{TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)}).SetRedirectPolicy(resty.NoRedirectPolicy())

	// client = client.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
	// 	fmt.Println("\n\nRequest:\n", r.Request.RawRequest)
	// 	fmt.Println("Response:\n", r.RawResponse)
	// 	return nil // if its success otherwise return error
	// })

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

func (s *Student) WithSession(session string) *Student {
	s.Session = session
	return s
}

// SaveLoginData save session and cookie to localfile
func (s *Student) SaveLoginData(filePath string) error {
	return utils.SaveData(filePath, []byte(utils.PrintStruct(LoginData{
		Cookies: s.Cookies,
		Session: s.Session,
	})))
}

func (s *Student) ClearCookies() {
	s.Cookies = make([]*http.Cookie, 0)
	s.client.Cookies = s.Cookies
}

func (s *Student) SetCookies(cookies []*http.Cookie) {
	s.AppendCookies(cookies)
	s.client = s.client.SetCookies(cookies)
}

func (s *Student) AppendCookies(cookies []*http.Cookie) {
	s.Cookies = append(s.Cookies, cookies...)
}

func (s *Student) NewRequest() *resty.Request {
	return s.client.R()
}

// GetWithSession returns parse tree for the resp of the request.
func (s *Student) GetWithSession(url string) (*html.Node, error) {
	resp, err := s.NewRequest().SetHeader("Referer", "https://jwcjwxt1.fzu.edu.cn/").SetQueryParam("id", s.Session).Get(url)

	if err != nil {
		return nil, errno.HTTPQueryError.WithErr(err)
	}

	// 会话过期 TODO: 判断条件有点简陋
	if strings.Contains(string(resp.Body()), "重新登录") {
		return nil, errno.SessionExpiredError
	}

	return htmlquery.Parse(bytes.NewReader(resp.Body()))
}

// GetWithSessionRaw returns the raw data of response
func (s *Student) GetWithSessionRaw(url string) (*resty.Response, error) {
	resp, err := s.NewRequest().SetHeader("Referer", "https://jwcjwxt1.fzu.edu.cn/").SetQueryParam("id", s.Session).Get(url)

	if err != nil {
		return nil, errno.HTTPQueryError.WithErr(err)
	}

	// 会话过期 TODO: 判断条件有点简陋
	if strings.Contains(string(resp.Body()), "重新登录") {
		return nil, errno.SessionExpiredError
	}

	return resp, nil
}

// PostWithSession returns parse tree for the resp of the request.
func (s *Student) PostWithSession(url string, formdata map[string]string) (*html.Node, error) {
	resp, err := s.NewRequest().SetHeader("Referer", "https://jwcjwxt1.fzu.edu.cn/").SetQueryParam("id", s.Session).SetFormData(formdata).Post(url)

	s.NewRequest().EnableTrace()

	if err != nil {
		return nil, errno.HTTPQueryError.WithErr(err)
	}

	// 会话过期 TODO: 判断条件有点简陋
	if strings.Contains(string(resp.Body()), "处理URL失败") {
		return nil, errno.SessionExpiredError
	}

	// utils.SaveData("test.html", resp.Body()) // 保存到本地查看网页

	return htmlquery.Parse(strings.NewReader(strings.TrimSpace(string(resp.Body()))))
}
