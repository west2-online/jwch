package jwch

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"test/errno"
	"test/utils"

	"github.com/go-resty/resty/v2"
)

// 模拟教务处登录/刷新Session
func (s *Student) Login() error {
	var code string // 验证码
	loginResp := SSOLoginResponse{}

	// 获取验证码图片
	resp, err := s.client.R().Get("https://jwcjwxt1.fzu.edu.cn/plus/verifycode.asp")

	if err != nil {
		return errno.HTTPQueryError.WithErr(err)
	}

	// TODO: 这里做的不好，实际上没必要保存图片吧
	utils.SaveData("test.png", resp.Body())
	fmt.Print("Please Input the code:")
	fmt.Scanln(&code)

	// 登录验证
	resp, err = s.client.R().SetHeaders(map[string]string{
		"Referer": "https://jwch.fzu.edu.cn",
		"Origin":  "https://jwch.fzu.edu.cn",
	}).SetFormData(map[string]string{
		"Verifycode": code,
		"muser":      s.ID,
		"passwd":     s.Password,
	}).Post("https://jwcjwxt1.fzu.edu.cn/logincheck.asp")

	if err != nil {
		return errno.HTTPQueryError.WithErr(err)
	}

	// 获取token，第一个是匹配的全部字符，第二个是我们需要的
	token := regexp.MustCompile(`var token = "(.*?)"`).FindStringSubmatch(string(resp.Body()))
	if len(token) < 1 {
		return errno.LoginCheckFailedError
	}

	// 获取session的id和num
	sessionURL := regexp.MustCompile(`window.location.href =\s{2}'(.*?)';`).FindStringSubmatch(string(resp.Body()))
	if len(token) < 1 {
		return errno.LoginCheckFailedError
	}
	id := regexp.MustCompile(`id=(.*?)&`).FindStringSubmatch(sessionURL[1])[1]
	num := regexp.MustCompile(`num=(.*?)&`).FindStringSubmatch(sessionURL[1])[1]

	// SSO登录
	resp, err = s.client.R().SetHeaders(map[string]string{
		"X-Requested-With": "XMLHttpRequest",
	}).SetFormData(map[string]string{
		"token": token[1],
	}).Post("https://jwcjwxt2.fzu.edu.cn/Sfrz/SSOLogin")

	if err != nil {
		return errno.HTTPQueryError.WithErr(err)
	}

	err = json.Unmarshal(resp.Body(), &loginResp)

	if err != nil {
		return errno.HTTPQueryError.WithErr(err)
	}

	// 获取account不存在是400，登录成功是200
	if loginResp.Code != 200 {
		return errno.SSOLoginFailedError
	}

	// 获取session
	s.client = s.client.SetRedirectPolicy(resty.NoRedirectPolicy())
	_, err = s.client.R().SetHeaders(map[string]string{
		"Referer": "https://jwcjwxt1.fzu.edu.cn/",
		"Origin":  "https://jwcjwxt2.fzu.edu.cn/",
	}).SetQueryParams(map[string]string{
		"id":       id,
		"num":      num,
		"ssourl":   "https://jwcjwxt2.fzu.edu.cn",
		"hosturl":  "https://jwcjwxt2.fzu.edu.cn:81",
		"ssologin": "",
	}).Get("https://jwcjwxt2.fzu.edu.cn:81/loginchk_xs.aspx")

	// 这里是err == nil 因为禁止了重定向，如果没有出现异常，那么就不对了
	if err == nil {
		return errno.GetSessionFailedError
	}

	session := regexp.MustCompile(`id=(.*?)&`).FindStringSubmatch(err.Error())[1]

	if len(session) < 1 {
		return errno.GetSessionFailedError
	}

	s.Session = session

	return nil
}

// 检查Session状态
func (s *Student) CheckSession() error {
	resp, err := s.client.R().SetQueryParam("id", s.Session).Get("https://jwcjwxt2.fzu.edu.cn:81/top.aspx")

	if err != nil {
		return err
	}

	// TODO: 这里的判断不太好
	if !strings.Contains(string(resp.Body()), "当前用户") {
		return errno.SessionExpiredError
	}

	return nil
}

// 获取学生个人信息
func (s *Student) GetInfo() error {
	resp, err := s.GetWithSession("https://jwcjwxt2.fzu.edu.cn:81/jcxx/xsxx/StudentInformation.aspx")

	if err != nil {
		return err
	}

	fmt.Println(resp)
	return nil
}
