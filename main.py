import requests
import urllib3
import re
import logging
# from captcha import captcha_identify

# 关闭警告
urllib3.disable_warnings()

# 打印日志级别
logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)

# 获取验证码
check_photo_url = "https://jwcjwxt1.fzu.edu.cn/plus/verifycode.asp"
req = requests.get(check_photo_url, verify=False)

# 获取cookies
cookies = req.cookies
logger.info("cookies为: " + str(cookies))


# 保存验证码
photo = req.content
with open("./photo.jpg", "wb") as f:
    f.write(photo)
captcha = input("请输入验证码: ")
# captcha = captcha_identify(img_fp="./photo.jpg")
logging.info("验证码为: " + str(captcha))

# 登录检查
login_check_url = "https://jwcjwxt1.fzu.edu.cn/logincheck.asp"
login_check_header = {
    "Referer": "https://jwch.fzu.edu.cn",
    "Origin": "https://jwch.fzu.edu.cn"
}
login_check_data = {
    "Verifycode": captcha,
    "muser": input("请输入学号:"),
    "passwd": input("请输入密码:")
}
req = requests.post(login_check_url, data=login_check_data, headers=login_check_header, cookies=cookies, verify=False)
cookies.update(req.cookies)

logger.info("Body: " + req.text)



# 获取token
token = re.findall('var token = "(.*?)"', req.text)[0]
logger.info("token为: " + token)

# 获取id和num
success_url = re.findall("window.location.href =\s{2}'(.*?)';", req.text)[0]
id = re.findall("id=(.*?)&", success_url)[0]
num = re.findall("num=(.*?)&", success_url)[0]
logger.info("id为: " + id)
logger.info("num为: " + num)

# sso
sso_url = "https://jwcjwxt2.fzu.edu.cn/Sfrz/SSOLogin"
sso_header = {
    "X-Requested-With": "XMLHttpRequest",
}
sso_data = {
    "token": token,
}
req = requests.post(sso_url, data=sso_data, headers=sso_header, verify=False)

# 更新cookies
cookies.update(req.cookies)
logger.info("cookies为: " + str(cookies))

# 获取session
session_url = "https://jwcjwxt2.fzu.edu.cn:81/loginchk_xs.aspx"
session_header = {
    "Referer": "https://jwcjwxt1.fzu.edu.cn/",
}
session_query = {
    "id": id,
    "num": num,
    "ssourl": "https://jwcjwxt2.fzu.edu.cn",
    "hosturl": "https://jwcjwxt2.fzu.edu.cn:81",
    "ssologin": ""
}

req = requests.get(session_url, params=session_query, headers=session_header, cookies=cookies, verify=False, allow_redirects=False)
user_id = re.findall("id=(.*?)&", req.text)[0]

cookies.update(req.cookies)
logger.info("cookies为: " + str(cookies))
logger.info("req.header: " + req.request.headers.__str__())

info_url = "https://jwcjwxt2.fzu.edu.cn:81/jcxx/xsxx/StudentInformation.aspx"
info_query = {
    "id": user_id,
    "Referer": "https://jwcjwxt1.fzu.edu.cn/",
}

req = requests.get(info_url, params=info_query, cookies=cookies, verify=False)
# print(req.text)
