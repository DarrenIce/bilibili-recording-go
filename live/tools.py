# -*- coding: utf-8 -*-

import base64
import chardet
import functools
import hashlib
import json
import os
import platform
import random
import requests
import rsa
import shutil
import subprocess
import sys
import threading
import time
import toml
from multiprocessing import freeze_support, Manager, Pool, Process
from urllib import parse
import logging
import yaml

logger = logging.getLogger()
# logger.setLevel('DEBUG')
# uploadLogPath = os.path.join(os.path.dirname(os.path.dirname(os.path.realpath(__file__))), 'log', 'upload.log')
# if not os.path.exists(uploadLogPath):
#     with open(uploadLogPath, 'w', encoding='utf-8') as a:
#         pass
# info_fh = logging.FileHandler(uploadLogPath,mode='a', encoding='utf-8')
# formatter = logging.Formatter(
#     '[%(levelname)s]\t%(asctime)s\t%(filename)s:%(lineno)d\tpid:%(thread)d\t%(message)s')
# info_fh.setFormatter(formatter)
# logger.addHandler(info_fh)

class Bilibili:
    app_key = "bca7e84c2d947ac6"

    def __init__(self, https=True, username = "", password = ""):
        self._session = requests.Session()
        self._session.headers.update({
            'User-Agent': "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66"})
        self.get_cookies = lambda: self._session.cookies.get_dict(domain=".bilibili.com")
        self.get_csrf = lambda: self.get_cookies().get("bili_jct", "")
        self.get_sid = lambda: self.get_cookies().get("sid", "")
        self.get_uid = lambda: self.get_cookies().get("DedeUserID", "")
        self.access_token = ""
        self.refresh_token = ""
        self.username = username
        self.password = password
        self.protocol = "https" if https else "http"
        self.proxy = {'http': 'socks5://127.0.0.1:1083', 'https': 'socks5://127.0.0.1:1083'}

    def _requests(self, method, url, decode_level=2, retry=10, timeout=15, **kwargs):
        if method in ["get", "post"]:
            for _ in range(retry + 1):
                try:
                    response = getattr(self._session, method)(url, timeout=timeout,
                                                              proxies=self.proxy, **kwargs)
                    return response.json() if decode_level == 2 else response.content if decode_level == 1 else response
                except:
                    continue
        return None

    def _solve_captcha(self, image):
        url = "https://bili.dev:2233/captcha"
        payload = {'image': base64.b64encode(image).decode("utf-8")}
        response = self._requests("post", url, json=payload)
        return response['message'] if response and response.get("code") == 0 else None

    @staticmethod
    def calc_sign(param):
        salt = "60698ba2f68e01ce44738920a0ffe768"
        sign_hash = hashlib.md5()
        sign_hash.update(f"{param}{salt}".encode())
        return sign_hash.hexdigest()

    # 登录
    def login(self, **kwargs):
        def by_password():
            def get_key():
                url = f"{self.protocol}://passport.bilibili.com/api/oauth2/getKey"
                payload = {
                    'appkey': Bilibili.app_key,
                    'sign': self.calc_sign(f"appkey={Bilibili.app_key}"),
                }
                while True:
                    response = self._requests("post", url, data=payload)
                    if response and response.get("code") == 0:
                        return {
                            'key_hash': response['data']['hash'],
                            'pub_key': rsa.PublicKey.load_pkcs1_openssl_pem(response['data']['key'].encode()),
                        }
                    else:
                        time.sleep(1)

            while True:
                key = get_key()
                key_hash, pub_key = key['key_hash'], key['pub_key']
                url = f"{self.protocol}://passport.bilibili.com/api/v2/oauth2/login"
                param = f"appkey={Bilibili.app_key}&password={parse.quote_plus(base64.b64encode(rsa.encrypt(f'{key_hash}{self.password}'.encode(), pub_key)))}&username={parse.quote_plus(self.username)}"
                payload = f"{param}&sign={self.calc_sign(param)}"
                headers = {'Content-type': "application/x-www-form-urlencoded"}
                response = self._requests("post", url, data=payload, headers=headers)
                logger.debug(response)
                while True:
                    if response and response.get("code") is not None:
                        if response['code'] == -105:
                            url = f"{self.protocol}://passport.bilibili.com/captcha"
                            headers = {'Host': "passport.bilibili.com"}
                            response = self._requests("get", url, headers=headers, decode_level=1)
                            captcha = self._solve_captcha(response)
                            if captcha:
                                logger.info(f"登录验证码识别结果: {captcha}")
                                key = get_key()
                                key_hash, pub_key = key['key_hash'], key['pub_key']
                                url = f"{self.protocol}://passport.bilibili.com/api/v2/oauth2/login"
                                param = f"appkey={Bilibili.app_key}&captcha={captcha}&password={parse.quote_plus(base64.b64encode(rsa.encrypt(f'{key_hash}{self.password}'.encode(), pub_key)))}&username={parse.quote_plus(self.username)}"
                                payload = f"{param}&sign={self.calc_sign(param)}"
                                headers = {'Content-type': "application/x-www-form-urlencoded"}
                                response = self._requests("post", url, data=payload, headers=headers)
                                logger.debug(response)
                            else:
                                break
                        elif response['code'] == -449:
                            logger.info("服务繁忙, 尝试使用V3接口登录")
                            url = f"{self.protocol}://passport.bilibili.com/api/v3/oauth2/login"
                            param = f"access_key=&actionKey=appkey&appkey={Bilibili.app_key}&build=6040500&captcha=&challenge=&channel=bili&cookies=&device=phone&mobi_app=android&password={parse.quote_plus(base64.b64encode(rsa.encrypt(f'{key_hash}{self.password}'.encode(), pub_key)))}&permission=ALL&platform=android&seccode=&subid=1&ts={int(time.time())}&username={parse.quote_plus(self.username)}&validate="
                            payload = f"{param}&sign={self.calc_sign(param)}"
                            headers = {'Content-type': "application/x-www-form-urlencoded"}
                            response = self._requests("post", url, data=payload, headers=headers)
                            logger.debug(response)
                        elif response['code'] == 0 and response['data']['status'] == 0:
                            for cookie in response['data']['cookie_info']['cookies']:
                                self._session.cookies.set(cookie['name'], cookie['value'], domain=".bilibili.com")
                            self.access_token = response['data']['token_info']['access_token']
                            self.refresh_token = response['data']['token_info']['refresh_token']
                            logger.info("登录成功")
                            return True
                        else:
                            logger.info(f"登录失败 {response}")
                            return False
                    else:
                        logger.info(f"当前IP登录过于频繁, {'尝试更换代理' if self.proxy else '1分钟后重试'}")
                        time.sleep(60)
                        break

        self._session.cookies.clear()
        for name in ["bili_jct", "DedeUserID", "DedeUserID__ckMd5", "sid", "SESSDATA"]:
            value = kwargs.get(name)
            if value:
                self._session.cookies.set(name, value, domain=".bilibili.com")
        if by_password():
            return True
        else:
            self._session.cookies.clear()
            return False

def login():
    f = open(os.path.join(os.path.dirname(os.path.dirname(os.path.realpath(__file__))), 'config.yml'), 'r')
    y = yaml.load(f, Loader=yaml.FullLoader)
    print(y['bilibili']['cookies'])
    if y['bilibili']['cookies'] != {}:
        return y['bilibili']['cookies']
    bili = Bilibili(username=y['bilibili']['user'], password=y['bilibili']['password'])
    bili.login()
    y['bilibili']['cookies'] = bili.get_cookies()
    f = open(os.path.join(os.path.dirname(os.path.dirname(os.path.realpath(__file__))), 'config.yml'), 'w')
    yaml.dump(y, f)
    return bili.get_cookies()


if __name__ == '__main__':
    login()