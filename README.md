# bilibili-recording-go
录播后台+简单的后台页面管理

## 功能
- 定时录制，<del>录制完成自动上传</del>(自动上传功能后面再维护吧，目前不考虑了)
- 配合BaiduPCS-GO，可以定时上传到百度云并导出秒链方便分享 (请提前手动登录BaiduPCS-GO)
- Http Rest API
- 简单的后台管理，不用搭配那个python界面了

## 依赖
- [ffmpeg](https://www.gyan.dev/ffmpeg/builds/), 请将ffmpeg放在环境变量里
- (可选)[BaiduPCS-Go](https://github.com/qjfoidnh/BaiduPCS-Go)，请提前登录好
- [streamlink](https://streamlink.github.io/)，建议用[Chocolatey](https://chocolatey.org/packages/streamlink)安装: ```choco install streamlink```
- 推荐有一个代理可以切Ip，虽然频率已经拉的很低了，但还是以防万一。

## 运行
```
git clone https://github.com/DarrenIce/bilibili-recording-go.git
pip install -r requirements.txt
修改config_exp.yml，并重命名为config.yml
go run main.go
```
`服务运行在18080端口`

## TODO

- 录制名字可以一定程度自定义
- beego页面
- 解决线程安全问题
- 增加更多自定义配置：是否转m4a，视频是否压缩画质，是否根据直播标题分P，定期删除

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