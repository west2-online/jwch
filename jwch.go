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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/west2-online/jwch/constants"
	"github.com/west2-online/jwch/errno"

	"github.com/antchfx/htmlquery"
	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
)

func NewStudent() *Student {
	// Disable HTTP/2.0
	// Disable Redirect
	client := resty.New().SetTransport(&http.Transport{
		TLSNextProto:    make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}).SetRedirectPolicy(resty.NoRedirectPolicy())

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

// WithUser 携带账号密码，这部分考虑整合到Login中，因为实际上我们不需要这个东西
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
	resp, err := s.NewRequest().SetHeader("Referer", constants.JwchReferer).SetQueryParam("id", s.Identifier).Get(url)
	if err != nil {
		return nil, errno.IdentifierExpiredError.WithErr(err)
	}

	// 会话过期 TODO: 判断条件有点简陋
	if strings.Contains(string(resp.Body()), "重新登录") {
		return nil, errno.IdentifierExpiredError
	}

	return htmlquery.Parse(bytes.NewReader(resp.Body()))
}

// PostWithIdentifier returns parse tree for the resp of the request.
func (s *Student) PostWithIdentifier(url string, formData map[string]string) (*html.Node, error) {
	resp, err := s.NewRequest().SetHeader("Referer", constants.JwchReferer).SetQueryParam("id", s.Identifier).SetFormData(formData).Post(url)

	s.NewRequest().EnableTrace()

	if err != nil {
		return nil, errno.IdentifierExpiredError.WithErr(err)
	}

	// Identifier缺失 TODO: 判断条件有点简陋
	if strings.Contains(string(resp.Body()), "处理URL失败") {
		return nil, errno.IdentifierExpiredError
	}

	return htmlquery.Parse(strings.NewReader(strings.TrimSpace(string(resp.Body()))))
}

// GetValidateCode 获取验证码
func GetValidateCode(image string) (string, error) {
	// 请求西二服务器，自动识别验证码
	code := verifyCodeResponse{}

	s := NewStudent()
	resp, err := s.NewRequest().SetFormData(map[string]string{
		"validateCode": image,
	}).Post("https://statistics.fzuhelper.w2fzu.com/api/login/validateCode?validateCode")
	if err != nil {
		return "", errno.HTTPQueryError.WithMessage("automatic code identification failed")
	}

	err = json.Unmarshal(resp.Body(), &code)
	if err != nil {
		return "", errno.HTTPQueryError.WithErr(err)
	}
	return code.Message, nil
}
