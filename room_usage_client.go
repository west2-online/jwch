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
	"fmt"
	"net/http"
	"time"

	"github.com/antchfx/htmlquery"

	jwchconstants "github.com/west2-online/jwch/constants"
	"golang.org/x/net/html"
)

const classroomUsageURL = "https://jwcjwxt2.fzu.edu.cn:81/kkgl/kbcx/kbcx_zb.aspx"

// 每次教室查询 GET/POST 前等待 200 毫秒，降低连续请求对教务处的压力。
const classroomUsageRequestDelay = 200 * time.Millisecond

// usageStudentClient 复用 jwch 的请求客户端、代理配置和登录 Cookie。
type usageStudentClient struct {
	Student *Student
}

func (c usageStudentClient) Get(ctx context.Context) (*html.Node, error) {
	return c.request(ctx, nil)
}

func (c usageStudentClient) Post(ctx context.Context, form map[string]string) (*html.Node, error) {
	return c.request(ctx, form)
}

func (c usageStudentClient) request(ctx context.Context, form map[string]string) (*html.Node, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(classroomUsageRequestDelay):
	}
	const requestTimeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	request := c.Student.NewRequest().SetContext(ctx).
		SetHeader("Referer", jwchconstants.JwchReferer).SetQueryParam("id", c.Student.Identifier)
	var body []byte
	if form == nil {
		response, err := request.Get(classroomUsageURL)
		if err != nil {
			return nil, fmt.Errorf("get classroom usage page: %w", err)
		}
		if response.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("get classroom usage page: HTTP %d", response.StatusCode())
		}
		body = response.Body()
	} else {
		response, err := request.SetFormData(form).Post(classroomUsageURL)
		if err != nil {
			return nil, fmt.Errorf("post classroom usage page: %w", err)
		}
		if response.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("post classroom usage page: HTTP %d", response.StatusCode())
		}
		body = response.Body()
	}
	return htmlquery.Parse(bytes.NewReader(body))
}
