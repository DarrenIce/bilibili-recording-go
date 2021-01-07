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
from bilibili_api import Verify
from bilibili_api import video
import tools

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

now = 0

def on_progress(up_data):
    global now
    if up_data['event'] != 'UPLOAD_CHUNK':
        logger.debug(up_data)
    else:
        # logger.debug(up_data)
        now += up_data['data']['size']
        logger.debug("完成度: %.2f" % (float(now) / float(up_data['data']['total'])))

def upload(uname, roomID, uploadName, filePath, cookies):
    global now
    now = 0
    logger.info('%s[RoomID:%s]开始本次上传，投稿名称: %s, 本地位置: %s' % (uname, roomID, uploadName, filePath))
    verify = Verify(sessdata=cookies['SESSDATA'], csrf=cookies['bili_jct'])
    try:
        filename = video.video_upload(filePath, verify, on_progress)
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
        result = video.video_submit(data, verify)
        logger.info('上传结果: %s' % (result))
    except:
        logger.error('%s[RoomID:%s]投稿失败' % (uname, roomID))
        return None

if __name__ == '__main__':
    uname, roomID, uploadName, filePath= sys.argv[1:5]
    upload(uname,roomID,uploadName,filePath,tools.login())
