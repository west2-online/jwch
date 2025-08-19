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
	"net/http"
	"net/url"
	"os"

	"github.com/west2-online/jwch/constants"
	"github.com/west2-online/jwch/errno"
)

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() *Config {
	config := &Config{
		Proxy: ProxyConfig{
			Enabled: false,
		},
	}

	// 从环境变量读取代理配置
	if authKey := os.Getenv("QINGGUO_AUTH_KEY"); authKey != "" {
		config.Proxy.AuthKey = authKey
	}
	if authPwd := os.Getenv("QINGGUO_AUTH_PWD"); authPwd != "" {
		config.Proxy.AuthPwd = authPwd
	}
	if enabled := os.Getenv("QINGGUO_PROXY_ENABLED"); enabled == "true" {
		config.Proxy.Enabled = true
	}

	return config
}

// GetProxyAddress 获取青果网络独享代理地址
func (c *Config) GetProxyAddress() (string, error) {
	if !c.Proxy.Enabled || c.Proxy.AuthKey == "" || c.Proxy.AuthPwd == "" {
		return "", fmt.Errorf("代理未启用或认证信息不完整")
	}

	client := &http.Client{}

	// 首先尝试查询现有的代理
	server, err := c.queryExistingProxy(client)
	if err == nil && server != "" {
		c.Proxy.ProxyServer = server
		return server, nil
	}

	// 如果没有现有代理或查询失败，尝试获取新的代理
	return c.getNewProxy(client)
}

// queryExistingProxy 查询现有的代理
func (c *Config) queryExistingProxy(client *http.Client) (string, error) {
	params := url.Values{}
	params.Set("key", c.Proxy.AuthKey)
	params.Set("pwd", c.Proxy.AuthPwd)

	resp, err := client.Get(constants.QingGuoExclusiveQueryURL + "?" + params.Encode())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var queryResp ExclusiveProxyQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return "", err
	}

	if queryResp.Code != "SUCCESS" || queryResp.Data == nil || len(queryResp.Data.Tasks) == 0 {
		return "", fmt.Errorf("没有现有的代理")
	}

	// 使用第一个任务的第一个IP
	task := queryResp.Data.Tasks[0]
	if len(task.IPs) == 0 {
		return "", fmt.Errorf("任务中没有可用的IP")
	}

	return task.IPs[0].Server, nil
}

// getNewProxy 获取新的代理
func (c *Config) getNewProxy(client *http.Client) (string, error) {
	params := url.Values{}
	params.Set("key", c.Proxy.AuthKey)
	params.Set("pwd", c.Proxy.AuthPwd)

	resp, err := client.Get(constants.QingGuoExclusiveGetURL + "?" + params.Encode())
	if err != nil {
		return "", errno.HTTPQueryError.WithMessage("获取代理地址失败").WithErr(err)
	}
	defer resp.Body.Close()

	var getResp ExclusiveProxyGetResponse
	if err := json.NewDecoder(resp.Body).Decode(&getResp); err != nil {
		return "", errno.HTTPQueryError.WithMessage("解析代理地址响应失败").WithErr(err)
	}

	if getResp.Code != "SUCCESS" {
		return "", fmt.Errorf("获取代理地址失败，响应码: %s, 消息: %s", getResp.Code, getResp.Message)
	}

	// 检查是否有可用的数据
	if getResp.Data == nil || len(getResp.Data.IPs) == 0 {
		return "", fmt.Errorf("没有可用的代理地址")
	}

	// 使用第一个可用的代理地址
	proxyServer := getResp.Data.IPs[0].Server
	if proxyServer == "" {
		return "", fmt.Errorf("代理地址为空")
	}

	// 更新配置中的代理服务器地址
	c.Proxy.ProxyServer = proxyServer
	return proxyServer, nil
}

// GetProxyURL 根据青果网络独享代理文档生成代理URL
func (c *Config) GetProxyURL() (*url.URL, error) {
	if !c.Proxy.Enabled {
		return nil, fmt.Errorf("代理未启用")
	}

	if c.Proxy.AuthKey == "" || c.Proxy.AuthPwd == "" || c.Proxy.ProxyServer == "" {
		return nil, fmt.Errorf("代理配置信息不完整")
	}

	// 独享代理模式：使用key作为用户名，pwd作为密码
	link := fmt.Sprintf("http://%s:%s@%s", c.Proxy.AuthKey, c.Proxy.AuthPwd, c.Proxy.ProxyServer)
	return url.Parse(link)
}
