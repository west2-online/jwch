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
	s.LoginData.Session = session
	return s
}

// SaveLoginData save session and cookie to localfile
func (s *Student) SaveLoginData(filePath string) error {
	return utils.SaveData(filePath, []byte(utils.PrintStruct(s.LoginData)))
}

func (s *Student) GetLoginDataJSON() string {
	return utils.PrintStruct(s.LoginData)
}

func (s *Student) ClearLoginData() {
	s.LoginData = LoginData{}
	s.client.Cookies = s.LoginData.Cookies
}

func (s *Student) SetLoginData(data LoginData) {
	s.LoginData = data
	s.client = s.client.SetCookies(s.LoginData.Cookies)
}

func (s *Student) AppendCookies(cookies []*http.Cookie) {
	s.LoginData.Cookies = append(s.LoginData.Cookies, cookies...)
}

func (s *Student) NewRequest() *resty.Request {
	return s.client.R()
}

// GetWithSession returns parse tree for the resp of the request.
func (s *Student) GetWithSession(url string) (*html.Node, error) {
	resp, err := s.NewRequest().SetHeader("Referer", "https://jwcjwxt1.fzu.edu.cn/").SetQueryParam("id", s.LoginData.Session).Get(url)

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
	resp, err := s.NewRequest().SetHeader("Referer", "https://jwcjwxt1.fzu.edu.cn/").SetQueryParam("id", s.LoginData.Session).Get(url)

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
	resp, err := s.NewRequest().SetHeader("Referer", "https://jwcjwxt1.fzu.edu.cn/").SetQueryParam("id", s.LoginData.Session).SetFormData(formdata).Post(url)

	s.NewRequest().EnableTrace()

	if err != nil {
		return nil, errno.HTTPQueryError.WithErr(err)
	}

	// 会话过期 TODO: 判断条件有点简陋
	if strings.Contains(string(resp.Body()), "处理URL失败") {
		return nil, errno.SessionExpiredError
	}

	return htmlquery.Parse(strings.NewReader(strings.TrimSpace(string(resp.Body()))))
}
