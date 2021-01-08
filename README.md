# bilibili-recording-go
录播服务端，无界面

## 功能
- 定时录制，录制完成自动上传
- 配合BaiduPCS-GO，可以定时上传到百度云 (请提前手动登录BaiduPCS-GO)
- Http Rest API

## 依赖
- [ffmpeg](https://www.gyan.dev/ffmpeg/builds/), 请将ffmpeg放在环境变量里
- [BaiduPCS-Go]https://github.com/qjfoidnh/BaiduPCS-Go
- 还需要你有一个代理，因为里面的连接都走了socks5，端口为1080，这是为了防止登录出验证码，或者触发了B站反爬可以及时切ip

## 运行
```
git clone https://github.com/DarrenIce/bilibili-recording-go.git
pip install -r requirements.txt
修改config_exp.yml，并重命名为config.yml
go run main.go
```
`服务运行在18080端口`

## API

### GET `/api/lives` 获取当前监控房间列表
- Request:
    ```
    method: GET
    url: 127.0.0.1:18080/api/lives
    ```
- Response:
    ```json
    [
        "53849",
        "2603963",
        "10796737"
    ]
    ```

### GET `/api/infos` 获取当前服务端详细信息
- Request:
    ```
    method: GET
    url: 127.0.0.1:18080/api/infos
    ```
- Response:
    ```json
    {
        "BiliInfo": {
            "Username": "",
            "Password": "",
            "Cookies": null
        },
        "RoomInfos": {
            "10796737": {
                "RoomID": "10796737",
                "StartTime": "220000",
                "EndTime": "020000",
                "AutoRecord": true,
                "AutoUpload": false,
                "RealID": "10796737",
                "LiveStatus": 0,
                "LockStatus": 0,
                "Uname": "筱筱莯",
                "UID": "306947837",
                "Title": "我只是一个妹妹~~",
                "LiveStartTime": 0,
                "AreaName": "放松电台",
                "RecordStatus": 0,
                "RecordStartTime": 0,
                "RecordEndTime": 0,
                "DecodeStatus": 0,
                "DecodeStartTime": 0,
                "DecodeEndTime": 0,
                "UploadStatus": 0,
                "UploadStartTime": 0,
                "UploadEndTime": 0,
                "NeedUpload": false,
                "St": "2021-01-07T22:00:00+08:00",
                "Et": "2021-01-08T02:00:00+08:00",
                "State": 1,
                "UploadName": "",
                "FilePath": ""
            },
            "2603963": {
                "RoomID": "2603963",
                "StartTime": "230000",
                "EndTime": "020000",
                "AutoRecord": true,
                "AutoUpload": false,
                "RealID": "2603963",
                "LiveStatus": 1,
                "LockStatus": 0,
                "Uname": "是幼情呀",
                "UID": "44070158",
                "Title": "进来舒服一下吧",
                "LiveStartTime": 1610009355,
                "AreaName": "放松电台",
                "RecordStatus": 0,
                "RecordStartTime": 0,
                "RecordEndTime": 0,
                "DecodeStatus": 0,
                "DecodeStartTime": 0,
                "DecodeEndTime": 0,
                "UploadStatus": 0,
                "UploadStartTime": 0,
                "UploadEndTime": 0,
                "NeedUpload": false,
                "St": "2021-01-07T23:00:00+08:00",
                "Et": "2021-01-08T02:00:00+08:00",
                "State": 1,
                "UploadName": "",
                "FilePath": ""
            },
            "53849": {
                "RoomID": "53849",
                "StartTime": "220000",
                "EndTime": "020000",
                "AutoRecord": true,
                "AutoUpload": false,
                "RealID": "53849",
                "LiveStatus": 0,
                "LockStatus": 0,
                "Uname": "小紗璃",
                "UID": "2323228",
                "Title": "被窝丨枕边细语",
                "LiveStartTime": 0,
                "AreaName": "放松电台",
                "RecordStatus": 0,
                "RecordStartTime": 0,
                "RecordEndTime": 0,
                "DecodeStatus": 0,
                "DecodeStartTime": 0,
                "DecodeEndTime": 0,
                "UploadStatus": 0,
                "UploadStartTime": 0,
                "UploadEndTime": 0,
                "NeedUpload": false,
                "St": "2021-01-07T22:00:00+08:00",
                "Et": "2021-01-08T02:00:00+08:00",
                "State": 1,
                "UploadName": "",
                "FilePath": ""
            }
        }
    }
    ```

### POST '/api/add` 添加新的监控房间
- Request:
    ```
    method: POST
    url: 127.0.0.1:18080/api/delete
    body:
        {
        "493":
            {
            "RoomID":"493",
            "StartTime":"220000",
            "EndTime":"020000",
            "AutoRecord":true,
            "AutoUpload":false
            }
        }
    ```
- Response:
    ```json
    [
        {
            "RoomID": "493",
            "RealID": "360972",
            "Uname": "不要吃咖喱",
            "Title": "温柔的小姐姐在你耳边呢喃",
            "LiveStatus": 2,
            "AutoRecord": true,
            "AutoUpload": false
        }
    ]
    ```

### POST '/api/delete` 删除监控房间
- Request:
    ```
    method: POST
    url: 127.0.0.1:18080/api/delete
    body:
        ["22490788"]
    ```
- Response:
    ```json
    {
        "info": "deleteRooms Success"
    }
    ```

## 参考
- [bililive-go](https://github.com/hr3lxphr6j/bililive-go)
- [Bilibili-Toolkit](https://github.com/Hsury/Bilibili-Toolkit)