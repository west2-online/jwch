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
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/west2-online/jwch/constants"
	"github.com/west2-online/jwch/errno"
)

// 全局缓存，避免重复获取隧道地址
var (
	globalConfig     *Config
	globalConfigOnce sync.Once
)

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() *Config {
	globalConfigOnce.Do(func() {
		globalConfig = &Config{
			Proxy: ProxyConfig{
				Enabled: false,
			},
		}

		// 从环境变量读取代理配置
		if authKey := os.Getenv("QINGGUO_AUTH_KEY"); authKey != "" {
			globalConfig.Proxy.AuthKey = authKey
		}
		if authPwd := os.Getenv("QINGGUO_AUTH_PWD"); authPwd != "" {
			globalConfig.Proxy.AuthPwd = authPwd
		}
		if shortterm := os.Getenv("QINGGUO_SHORTTERM"); shortterm == "true" {
			globalConfig.Proxy.Shortterm = true
		}
		if enabled := os.Getenv("QINGGUO_PROXY_ENABLED"); enabled == "true" {
			globalConfig.Proxy.Enabled = true
		}
	})

	return globalConfig
}

// GetTunnelAddress 获取青果网络隧道地址
func (c *Config) GetTunnelAddress() ([]string, error) {
	if !c.Proxy.Enabled || c.Proxy.AuthKey == "" || c.Proxy.AuthPwd == "" {
		return nil, fmt.Errorf("代理未启用或认证信息不完整")
	}

	// 1) 优先使用当前Config中已存在的代理地址
	if servers := validTunnelServers(c.Proxy.ProxyServers); len(servers) != 0 {
		c.Proxy.ProxyServers = servers
		return append([]string(nil), servers...), nil
	}

	client := &http.Client{}

	// 构建请求参数
	params := url.Values{}
	params.Set("key", c.Proxy.AuthKey)
	params.Set("pwd", c.Proxy.AuthPwd)

	QingGuoTunnelURL := constants.QingGuoLongtermTunnelURL
	if c.Proxy.Shortterm {
		QingGuoTunnelURL = constants.QingGuoShareTunnelURL
	}

	// 发送GET请求
	resp, err := client.Get(QingGuoTunnelURL + "?" + params.Encode())
	if err != nil {
		return nil, errno.HTTPQueryError.WithMessage("获取隧道地址失败").WithErr(err)
	}
	defer resp.Body.Close()

	var tunnelResp TunnelResponse
	if err := json.NewDecoder(resp.Body).Decode(&tunnelResp); err != nil {
		return nil, errno.HTTPQueryError.WithMessage("解析隧道地址响应失败").WithErr(err)
	}

	if tunnelResp.Code != "SUCCESS" {
		return nil, fmt.Errorf("获取隧道地址失败，响应码: %s", tunnelResp.Code)
	}

	// 检查是否有可用的隧道数据
	if len(tunnelResp.Data) == 0 {
		return nil, fmt.Errorf("没有可用的隧道地址")
	}

	// 更新配置中的代理服务器地址，并过滤无效地址。
	tunnelServers := make([]string, 0, len(tunnelResp.Data))

	for _, server := range tunnelResp.Data {
		if server := strings.TrimSpace(server.Server); server != "" {
			tunnelServers = append(tunnelServers, server)
		}
	}
	if len(tunnelServers) == 0 {
		return nil, fmt.Errorf("没有可用的隧道地址")
	}

	c.Proxy.ProxyServers = tunnelServers
	return append([]string(nil), tunnelServers...), nil
}

// GetProxyURL 根据青果网络文档生成代理URL
func (c *Config) GetProxyURL() (*url.URL, error) {
	if !c.Proxy.Enabled {
		return nil, fmt.Errorf("代理未启用")
	}

	if c.Proxy.AuthKey == "" || c.Proxy.AuthPwd == "" {
		return nil, fmt.Errorf("代理配置信息不完整")
	}

	servers := validTunnelServers(c.Proxy.ProxyServers)
	if len(servers) == 0 {
		return nil, fmt.Errorf("代理配置信息不完整")
	}

	// 负载均衡
	server := servers[rand.Intn(len(servers))]
	u := &url.URL{
		Scheme: "http",
		User:   url.UserPassword(c.Proxy.AuthKey, c.Proxy.AuthPwd),
		Host:   server,
	}
	return u, nil
}

func validTunnelServers(servers []string) []string {
	validServers := make([]string, 0, len(servers))
	for _, server := range servers {
		if server = strings.TrimSpace(server); server != "" {
			validServers = append(validServers, server)
		}
	}
	return validServers
}
