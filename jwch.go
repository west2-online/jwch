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
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/west2-online/jwch/constants"
	"github.com/west2-online/jwch/errno"

	"github.com/antchfx/htmlquery"
	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
)

const defaultRequestDelay = 200 * time.Millisecond

func NewStudent() *Student {
	// 从环境变量加载配置
	config := LoadConfigFromEnv()

	// Disable HTTP/2.0
	// Disable Redirect
	transport := &http.Transport{
		TLSNextProto:    make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// 如果启用了代理，先获取隧道地址再设置代理
	if config.Proxy.Enabled {
		servers, err := config.GetTunnelAddress()
		if err == nil && len(servers) != 0 {
			proxyURL, err := config.GetProxyURL()
			if err == nil {
				transport.Proxy = http.ProxyURL(proxyURL)
			}
		}
	}

	client := resty.New().
		SetTransport(transport).
		SetRedirectPolicy(resty.NoRedirectPolicy())

	return &Student{
		client:       client,
		requestDelay: defaultRequestDelay,
	}
}

// WithRequestDelay 设置请求前的等待时间，等于 0 时不等待
func (s *Student) WithRequestDelay(delay time.Duration) *Student {
	s.requestDelay = max(delay, 0)
	return s
}

// waitForRequest 等待配置的请求间隔，并在上下文取消时提前返回。
func (s *Student) waitForRequest(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.requestDelay <= 0 {
		return nil
	}
	timer := time.NewTimer(s.requestDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return ctx.Err()
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
	// 会话过期：会直接重定向，但我们禁用了重定向，所以会有error
	if err != nil {
		// 由于评议在重定向后的页面上，所以我们需要处理重定向
		if resp != nil && resp.StatusCode() == 302 {
			redirectURL := resp.Header().Get("Location")
			// 再次访问重定向后的URL(带prefix)
			respRedirected, errRedirected := s.NewRequest().
				SetHeader("Referer", constants.JwchReferer).
				SetQueryParam("id", s.Identifier).
				Get(constants.JwchPrefix + redirectURL)

			if errRedirected != nil {
				return nil, errno.CookieError
			}
			if strings.Contains(string(respRedirected.Body()), "请先对任课教师和教材进行测评") {
				return nil, errno.EvaluationNotFoundError
			}
		}
		return nil, errno.CookieError
	}

	text := string(resp.Body())
	// 正常页面可能已是 UTF-8，教务处错误页仍使用 GB2312 (Fuck GBK)
	if !utf8.Valid(resp.Body()) {
		text, err = utils.ConvertGB2312ToUTF8(resp.Body())
		if err != nil {
			return nil, errno.HTMLParseError.WithErr(err)
		}
	}
	if strings.Contains(text, "重新登录") {
		return nil, errno.CookieError
	}
	if strings.Contains(string(resp.Body()), "请先对任课教师进行测评") {
		return nil, errno.EvaluationNotFoundError
	}
	doc, err := htmlquery.Parse(strings.NewReader(strings.TrimSpace(text)))
	if err != nil {
		return nil, err
	}
	if strings.Contains(text, "处理URL失败") {
		return nil, errno.HTTPQueryError.WithMessage(normalizedHTMLText(doc))
	}
	return doc, nil
}

// PostWithIdentifier returns parse tree for the resp of the request.
func (s *Student) PostWithIdentifier(url string, formData map[string]string) (*html.Node, error) {
	resp, err := s.NewRequest().SetHeader("Referer", constants.JwchReferer).SetQueryParam("id", s.Identifier).SetFormData(formData).Post(url)

	s.NewRequest().EnableTrace()
	// 会话过期：会直接重定向，但我们禁用了重定向，所以会有error
	if err != nil {
		// 由于评议在重定向后的页面上，所以我们需要处理重定向
		if resp != nil && resp.StatusCode() == 302 {
			redirectURL := resp.Header().Get("Location")
			// 再次访问重定向后的URL(带prefix)
			respRedirected, errRedirected := s.NewRequest().
				SetHeader("Referer", constants.JwchReferer).
				SetQueryParam("id", s.Identifier).
				// 这里不确定应该Get还是Post，但目前Post Method没有会被评议卡的
				Get(constants.JwchPrefix + redirectURL)

			if errRedirected != nil {
				return nil, errno.JwchNetworkError.WithErr(err)
			}
			if strings.Contains(string(respRedirected.Body()), "请先对任课教师进行测评") {
				return nil, errno.EvaluationNotFoundError
			}
		}
		return nil, errno.JwchNetworkError.WithErr(err)
	}

	text := string(resp.Body())
	if !utf8.Valid(resp.Body()) {
		text, err = utils.ConvertGB2312ToUTF8(resp.Body())
		if err != nil {
			return nil, errno.HTMLParseError.WithErr(err)
		}
	}
	if strings.Contains(text, "重新登录") {
		return nil, errno.CookieError
	}
	if strings.Contains(string(resp.Body()), "请先对任课教师进行测评") {
		return nil, errno.EvaluationNotFoundError
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
