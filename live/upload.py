import requests
import os
import asyncio
import json
import math
import aiohttp
import re
import copy
import logging
import sys

logger = logging.getLogger()
logger.setLevel('DEBUG')
uploadLogPath = os.path.join(os.path.dirname(os.path.dirname(os.path.realpath(__file__))), 'log', 'upload.log')
if not os.path.exists(uploadLogPath):
    with open(uploadLogPath, 'w', encoding='utf-8') as a:
        pass
info_fh = logging.FileHandler(uploadLogPath,mode='a', encoding='utf-8')
formatter = logging.Formatter(
    '[%(levelname)s]\t%(asctime)s\t%(filename)s:%(lineno)d\tpid:%(thread)d\t%(message)s')
info_fh.setFormatter(formatter)
logger.addHandler(info_fh)

DEFAULT_HEADERS = {
    'User-Agent':'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36 Edg/86.0.622.69',
    'accept':'application/json, text/plain, */*',
    'accept-encoding':'gzip, deflate, br',
    'accept-language':'zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6,zh-TW;q=0.5',
    'referer':'https://live.bilibili.com/'
}

class BilibiliApiException(Exception):
    def __init__(self, msg):
        self.msg = msg

    def __str__(self):
        return self.msg

class UploadException(BilibiliApiException):
    def __init__(self, msg: str):
        super().__init__(msg)

class BilibiliException(BilibiliApiException):
    def __init__(self, code, msg):
        self.code = code
        self.msg = msg

    def __str__(self):
        return "错误代码：%s, 信息：%s" % (self.code, self.msg)

class NetworkException(BilibiliApiException):
    def __init__(self, code):
        self.code = code

    def __str__(self):
        return "网络错误。状态码：%s" % self.code

def request(method: str, url: str, params=None, data=None, cookies=None, headers=None, data_type: str = "form", **kwargs):
    if params is None:
        params = {}
    if data is None:
        data = {}
    if cookies is None:
        cookies = {}
    if headers is None:
        headers = copy.deepcopy(DEFAULT_HEADERS)
    if data_type.lower() == "json":
        headers['Content-Type'] = "application/json"
    st = {
        "url": url,
        "params": params,
        "cookies": cookies,
        "headers": headers,
        "verify": True,
        "data": data,
        "proxies": None
    }
    st.update(kwargs)

    req = requests.request(method, **st)

    if req.ok:
        content = req.content.decode("utf8")
        if req.headers.get("content-length") == 0:
            return None
        if 'jsonp' in params and 'callback' in params:
            con = json.loads(re.match(".*?({.*}).*", content, re.S).group(1))
        else:
            con = json.loads(content)
        if con["code"] != 0:
            if "message" in con:
                msg = con["message"]
            elif "msg" in con:
                msg = con["msg"]
            else:
                msg = "请求失败，服务器未返回失败原因"
            raise BilibiliException(con["code"], msg)
        else:
            if 'data' in con.keys():
                return con['data']
            else:
                if 'result' in con.keys():
                    return con["result"]
                else:
                    return None
    else:
        raise NetworkException(req.status_code)

def post(url, cookies, data=None, headers=None, data_type: str = "form", **kwargs):
    """
    专用POST请求
    :param data_type:
    :param url:
    :param cookies:
    :param data:
    :param headers:
    :param kwargs:
    :return:
    """
    resp = request("POST", url=url, data=data, cookies=cookies, headers=headers, data_type=data_type, **kwargs)
    return resp

def video_upload(path: str, cookies):
    """
    上传视频
    :param on_progress: 进度回调，数据格式：{"event": "事件名", "ok": "是否成功", "data": "附加数据"}
                        事件名：PRE_UPLOAD，GET_UPLOAD_ID，UPLOAD_CHUNK，VERIFY
    :param path: 视频路径
    :param cookies:
    :return: 该视频的filename，用于后续提交投稿用
    """
    
    session = requests.session()
    session.headers = DEFAULT_HEADERS
    requests.utils.add_dict_to_cookiejar(session.cookies, cookies)
    if not os.path.exists(path):
        raise UploadException("视频路径不存在")
    total_size = os.stat(path).st_size
    # 上传设置
    params = {
        'name': os.path.basename(path),
        'size': total_size,
        'r': 'upos',
        'profile': 'ugcupos/bup'
    }
    try:
        resp = session.get('https://member.bilibili.com/preupload', params=params)
        settings = resp.json()
        upload_url = 'https:' + settings['endpoint'] + '/' + settings['upos_uri'].replace('upos://', '')
        headers = {
            'X-Upos-Auth': settings['auth']
        }
    except Exception as e:
        raise e
    try:
        resp = session.post(upload_url + "?uploads&output=json", headers=headers)
        settings['upload_id'] = resp.json()['upload_id']
        filename = os.path.splitext(resp.json()['key'].lstrip('/'))[0]
    except Exception as e:
        raise e
    # 分配任务
    chunks_settings = []
    i = 0
    total_chunks = math.ceil(total_size / settings['chunk_size'])
    offset = 0
    remain = total_size
    while True:
        s = {
            'partNumber': i + 1,
            'uploadId': settings['upload_id'],
            'chunk': i,
            'chunks': total_chunks,
            'start': offset,
            'end': offset + settings['chunk_size'] if remain >= settings['chunk_size'] else total_size,
            'total': total_size
        }
        s['size'] = s['end'] - s['start']
        chunks_settings.append(s)
        i += 1
        offset = s['end']
        remain -= settings['chunk_size']
        if remain <= 0:
            break

    async def upload(chunks, sess):
        failed_chunks = []
        with open(path, 'rb') as f:
            for chunk in chunks:
                f.seek(chunk['start'], 0)
                async with sess.put(upload_url, params=chunk, data=f.read(chunk['size']),
                                    headers=DEFAULT_HEADERS) as r:
                    if r.status != 200:
                        failed_chunks.append(chunk)
        return failed_chunks

    async def main():
        chunks_per_thread = len(chunks_settings) // settings['threads']
        remain = len(chunks_settings) % settings['threads']
        task_chunks = []
        for i in range(settings['threads']):
            this_task_chunks = chunks_settings[i * chunks_per_thread:(i + 1) * chunks_per_thread]
            task_chunks.append(this_task_chunks)
        task_chunks[-1] += (chunks_settings[-remain:])

        async with aiohttp.ClientSession(headers={'X-Upos-Auth': settings['auth']}, cookies=cookies) as sess:
            while True:
                # 循环上传
                coroutines = []
                chs = task_chunks
                for chunks in chs:
                    coroutines.append(upload(chunks, sess))
                results = await asyncio.gather(*coroutines)
                failed_chunks = []
                for result in results:
                    failed_chunks += result
                chs = failed_chunks
                if len(chs) == 0:
                    break
            # 验证是否上传成功
            params = {
                'output': 'json',
                'name': os.path.basename(path),
                'profile': 'ugcupos/bup',
                'uploadId': settings['upload_id'],
                'biz_id': settings['biz_id']
            }
            payload = {
                'parts': []
            }
            for chunk in chunks_settings:
                payload['parts'].append({
                    'eTag': 'eTag',
                    'partNumber': chunk['partNumber']
                })
            async with sess.post(upload_url, params=params, data=payload) as resp:
                result = await resp.read()
                result = json.loads(result)
                ok = result.get('OK', 0)
                if ok == 1:
                    return filename
                else:
                    raise UploadException('视频上传失败')
    r = asyncio.run(main())
    return r

def video_submit(data: dict, cookies):
    """
    提交投稿信息
    :param data: 投稿信息
    {
        "copyright": 1自制2转载,
        "source": "类型为转载时注明来源",
        "cover": "封面URL",
        "desc": "简介",
        "desc_format_id": 0,
        "dynamic": "动态信息",
        "interactive": 0,
        "no_reprint": 1为显示禁止转载,
        "subtitles": {
            // 字幕格式，请自行研究
            "lan": "语言",
            "open": 0
        },
        "tag": "标签1,标签2,标签3（英文半角逗号分隔）",
        "tid": 分区ID,
        "title": "标题",
        "videos": [
            {
                "desc": "描述",
                "filename": "video_upload(返回值)",
                "title": "分P标题"
            }
        ]
    }
    :param verify:
    :return:
    """
    url = "https://member.bilibili.com/x/vu/web/add"
    params = {
        "csrf": cookies['bili_jct']
    }
    payload = json.dumps(data, ensure_ascii=False).encode()
    resp = post(url, params=params, data=payload, data_type="json", cookies=cookies)
    return resp

def upload(uname, roomID, uploadName, filePath, cookies):
    logger.info('%s[RoomID:%s]开始本次上传，投稿名称: %s, 本地位置: %s' % (uname, roomID, uploadName, filePath))
    try:
        filename = video_upload(filePath, cookies=cookies)
    except Exception as e:
        logger.error('%s[RoomID:%s]上传失败 %s' % (uname, roomID, e))
        return
    logger.info('%s[RoomID:%s]上传成功' % (uname, roomID))
    data = {
        "copyright": 2,
        "source": "https://live.bilibili.com/%s" % roomID,
        "cover": "",
        "desc": "",
        "desc_format_id": 0,
        "dynamic": "",
        "interactive": 0,
        "no_reprint": 0,
        "subtitles": {
            "lan": "",
            "open": 0
        },
        "tag": "录播,%s" % uname,
        "tid": 174,
        "title": uploadName,
        "videos": [
            {
                "desc": "",
                "filename": filename,
                "title": "P1"
            }
        ]
    }
    try:
        result = video_submit(data, cookies=cookies)
        logger.info('上传结果: %s' % (result))
    except:
        logger.error('%s[RoomID:%s]投稿失败' % (uname, roomID))
        return None

if __name__ == '__main__':
    uname, roomID, uploadName, filePath, DedeUserID, DedeUserID__ckMd5, SESSDATA, bili_jct, sid = sys.argv[1:10]
    cookies = {
        'DedeUserID':DedeUserID,
        'DedeUserID__ckMd5':DedeUserID__ckMd5,
        'SESSDATA':SESSDATA,
        'bili_jct':bili_jct,
        'sid':sid
    }
    upload(uname,roomID,uploadName,filePath,cookies)
