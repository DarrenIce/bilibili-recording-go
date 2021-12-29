# bilibili-recording-go
完全重构的哔站录播系统

## 功能
- 定时录制，<del>录制完成自动上传</del>(自动上传功能后面再维护吧，目前不考虑了)
- 配合BaiduPCS-GO，可以定时上传到百度云并导出秒链方便分享 (请提前手动登录BaiduPCS-GO)
- 后台管理页面，操作简单
- 尽可能的开箱即用

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

## TODO

- 增加更多自定义配置：
  - [x] 是否转m4a
  - [x] 视频是否压缩画质
  - [ ] 是否根据直播标题分P
  - [ ] 定期删除
  - [ ] 分区锁定
  - [x] 全天录制or指定时间段录制
- [ ] 可以监控整个放松分区
  API：https://api.live.bilibili.com/xlive/web-interface/v1/second/getList?platform=web&parent_area_id=5&area_id=339&sort_type=&page=1
  可以标记黑名单up主，将不在监控界面展示
- [ ] 记录主播的开播和下播时间
