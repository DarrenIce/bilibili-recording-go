const { defineComponent, ref, toRefs, reactive } = Vue;
const { ElNotification } = ElementPlus;

function RoomConfigInfo(room) {
  if (typeof room == "string") {
    this.RoomID = room
    this.RecordMode = false
    this.StartTime = new Date()
    this.EndTime = new Date()
    this.AutoRecord = true
    this.AutoUpload = false
    this.NeedM4a = true
    this.Mp4Compress = true
    this.DivideByTitle = false
    this.CleanUpRegular = false
    this.SaveDuration = "7d"
    this.AreaLock = false
    this.AreaLimit = "放松电台"
  } else {
    this.RoomID = room.RoomID
    this.RecordMode = room.RecordMode
    this.StartTime = date2string(room.StartTime)
    this.EndTime = date2string(room.EndTime)
    this.AutoRecord = room.AutoRecord
    this.AutoUpload = room.AutoUpload
    this.NeedM4a = room.NeedM4a
    this.Mp4Compress = room.Mp4Compress
    this.DivideByTitle = room.DivideByTitle
    this.CleanUpRegular = room.CleanUpRegular
    this.SaveDuration = room.SaveDuration
    this.AreaLock = room.AreaLock
    this.AreaLimit = room.AreaLimit
  }
}

let Main = {
  setup() {
    const state = reactive({
      colors: [
        { color: '#5cb87a', percentage: 20 },
        { color: '#1989fa', percentage: 40 },
        { color: '#6f7ad3', percentage: 60 },
        { color: '#e6a23c', percentage: 80 },
        { color: '#f56c6c', percentage: 100 },
      ],
    })
    return toRefs(state)
  },
  data() {
    return {
      tableData: [],
      recordCount: 0,
      cpuPct: 0.0,
      memoryUsage: 0.0,
      memoryTotal: 0.0,
      memoryPct: 0.0,
      uploadSpeed: '0K/s👆',
      downloadSpeed: '0K/s👇',
      diskName: '',
      diskUsage: 0.0,
      diskTotal: 0.0,
      diskPct: 0.0,
      totalDownload: '',
      fileNum: 0,
      editFormVisible: false,
      disabledRecord: false,
      addFormVisible: false,
      disableAddRecord: false,
      form: new RoomConfigInfo(""),
      addForm: new RoomConfigInfo(""),
    }
  },
  mounted() {
    this.timer1 = setInterval(getLowFrqData, 5000);
    this.timer2 = setInterval(flushData, 1000);
    this.timer3 = setInterval(getHighFrqData, 1000);
  },
  beforeDestory() {
    clearInterval(this.timer1);
    clearInterval(this.timer2);
    clearInterval(this.timer3);
  },
  methods: {
    tableRowColor({ row, rowIndex }) {
      if (row.LiveStatus == 1) {
        return 'background-color: #e1f3d8;'
      } else if (row.LiveStatus == 2) {
        return 'background-color: #faecd8;'
      }
    },
    liveSort(obj1, obj2) {
      let v1 = obj1.LiveStatus == 0 ? 0 : -obj1.LiveStatus + 3
      let v2 = obj2.LiveStatus == 0 ? 0 : -obj2.LiveStatus + 3
      return v1 - v2
    },
    state2Type(state) {
      if (state == 1 || state == 2) {
        return
      } else if (state == 3 || state == 5 || state == 8) {
        return 'success'
      } else if (state == 4 || state == 7) {
        return 'warning'
      } else {
        return 'danger'
      }
    },
    state2Name(state) {
      if (state == 1) {
        return '正在监听'
      } else if (state == 2) {
        return '等待重连'
      } else if (state == 3) {
        return '录制中'
      } else if (state == 4) {
        return '等待转码'
      } else if (state == 5) {
        return '转码中'
      } else if (state == 6) {
        return '转码结束'
      } else if (state == 7) {
        return '等待上传'
      } else if (state == 8) {
        return '上传中'
      } else if (state == 9) {
        return '上传结束'
      } else if (state == 10) {
        return '停止监听'
      }
    },
    status2Type(status) {
      if (status == 0) {
        return
      } else if (status == 1) {
        return 'success'
      } else if (status == 2) {
        return 'warning'
      }
    },
    status2Name(status) {
      if (status == 0) {
        return '未开播'
      } else if (status == 1) {
        return '直播中'
      } else if (status == 2) {
        return '录播中'
      }
    },
    handleDelete(row) {
      ElNotification({
        title: '成功',
        message: '删除房间成功',
        type: 'success'
      })
      console.log('Delete!')
      console.log(row)
      $.ajax({
        type: "POST",
        url: "/room-handle",
        dataType: "json",
        data: JSON.stringify({
          handle: "delete",
          data: new RoomConfigInfo(vm.form)
          // data: {
          //   RoomID: vm.form.RoomID,
          //   RecordMode: vm.form.RecordMode,
          //   StartTime: date2string(vm.form.StartTime),
          //   EndTime: date2string(vm.form.EndTime),
          //   AutoRecord: vm.form.AutoRecord,
          //   AutoUpload: vm.form.AutoUpload,
          //   NeedM4a: vm.form.NeedM4a,
          //   Mp4Compress: vm.form.Mp4Compress,
          //   DivideByTitle: vm.form.DivideByTitle,
          //   CleanUpRegular: vm.form.CleanUpRegular,
          //   SaveDuration: vm.form.SaveDuration,
          //   AreaLock: vm.form.AreaLock,
          //   AreaLimit: vm.form.AreaLimit
          // },
        }),
        headers: {
          "Content-Type": "application/json"
        },
        success: function (msg) {
          console.log(msg)
        }
      })
      this.editFormVisible = false
    },
    handleOpen(index, row) {
      console.log(row)
      vm.editFormVisible = true
      vm.form = {
        RoomID: row.RoomID,
        RecordMode: row.RecordMode,
        StartTime: new Date(2021, 9, 29, row.StartTime.slice(0, 2), row.StartTime.slice(2, 4), row.StartTime.slice(4, 6)),
        EndTime: new Date(2021, 9, 29, row.EndTime.slice(0, 2), row.EndTime.slice(2, 4), row.EndTime.slice(4, 6)),
        AutoRecord: row.AutoRecord,
        AutoUpload: row.AutoUpload,
        NeedM4a: row.NeedM4a,
        Mp4Compress: row.Mp4Compress,
        DivideByTitle: row.DivideByTitle,
        CleanUpRegular: row.CleanUpRegular,
        SaveDuration: row.SaveDuration,
        AreaLock: row.AreaLock,
        AreaLimit: row.AreaLimit
      }
      let r = {}
      for (let key in vm.tableData) {
        if (vm.tableData[key].RoomID == row.RoomID) {
          r = vm.tableData[key]
          break
        }
      }
      vm.room = {
        RoomID: r.RoomID,
        Uname: r.Uname,
        AreaName: r.AreaName,
        Title: r.Title,
        LiveStatus: r.LiveStatus,
        LiveStartTime: r.LiveStartTime,
        LiveStartTime2: r.LiveStartTime2,
        State: r.State,
        StartTime: r.StartTime,
        EndTime: r.EndTime,
        RecordMode: r.RecordMode,
        RecordStartTime: r.RecordStartTime,
        RecordEndTime: r.RecordEndTime,
        DecodeStartTime: r.DecodeStartTime,
        DecodeEndTime: r.DecodeEndTime,
        UploadStartTime: r.UploadStartTime,
        UploadEndTime: r.UploadEndTime,
        LiveTime: getTimeMiuns(r.LiveStartTime2, 0),
        RecordTime: getTimeMiuns(r.RecordStartTime, r.RecordEndTime),
        DecodeTime: getTimeMiuns(r.DecodeStartTime, r.DecodeEndTime),
        UploadTime: getTimeMiuns(r.UploadStartTime, r.UploadEndTime),
        NeedM4a: r.NeedM4a,
        Mp4Compress: r.Mp4Compress,
        DivideByTitle: r.DivideByTitle,
        CleanUpRegular: r.CleanUpRegular,
        SaveDuration: r.SaveDuration,
        AreaLock: r.AreaLock,
        AreaLimit: r.AreaLimit
      }
      console.log(vm.room)
    },
    onSubmit() {
      this.editFormVisible = false
      ElNotification({
        title: '成功',
        message: '编辑房间成功',
        type: 'success'
      })
      console.log('submit!')
      console.log(vm.form)
      // TODO: Ajax回传给go处理
      $.ajax({
        type: "POST",
        url: "/room-handle",
        dataType: "json",
        data: JSON.stringify({
          handle: "edit",
          data: new RoomConfigInfo(vm.form)
          // data: {
          //   RoomID: vm.form.RoomID,
          //   RecordMode: vm.form.RecordMode,
          //   StartTime: date2string(vm.form.StartTime),
          //   EndTime: date2string(vm.form.EndTime),
          //   AutoRecord: vm.form.AutoRecord,
          //   AutoUpload: vm.form.AutoUpload,
          //   NeedM4a: vm.form.NeedM4a,
          //   Mp4Compress: vm.form.Mp4Compress,
          //   DivideByTitle: vm.form.DivideByTitle,
          //   CleanUpRegular: vm.form.CleanUpRegular,
          //   SaveDuration: vm.form.SaveDuration,
          //   AreaLock: vm.form.AreaLock,
          //   AreaLimit: vm.form.AreaLimit
          // },
        }),
        headers: {
          "Content-Type": "application/json"
        },
        success: function (msg) {
          console.log(msg)
        }
      })
    },
    changeRecordTimeText(recordMode) {
      this.disabledRecord = recordMode
    },
    changeAddRecordTimeText(recordMode) {
      this.disableAddRecord = recordMode
    },
    addRoom() {
      console.log('addRoom')
      this.addFormVisible = true
    },
    roomID2Url(roomID) {
      return "https://live.bilibili.com/" + roomID
    },
    addOnSubmit() {
      this.addFormVisible = false
      ElNotification({
        title: '成功',
        message: '添加房间成功',
        type: 'success'
      })
      console.log('addSubmit!')
      console.log(
        new RoomConfigInfo(vm.addForm)
      //   {
      //   RoomID: vm.addForm.RoomID,
      //   RecordMode: vm.addForm.RecordMode,
      //   StartTime: date2string(vm.addForm.StartTime),
      //   EndTime: date2string(vm.addForm.EndTime),
      //   AutoRecord: vm.addForm.AutoRecord,
      //   AutoUpload: vm.addForm.AutoUpload,
      //   NeedM4a: vm.addForm.NeedM4a,
      //   Mp4Compress: vm.addForm.Mp4Compress,
      //   DivideByTitle: vm.addForm.DivideByTitle,
      //   CleanUpRegular: vm.addForm.CleanUpRegular,
      //   SaveDuration: vm.addForm.SaveDuration,
      //   AreaLock: vm.addForm.AreaLock,
      //   AreaLimit: vm.addForm.AreaLimit
      // }
      )
      // TODO: Ajax回传给go处理
      $.ajax({
        type: "POST",
        url: "/room-handle",
        dataType: "json",
        data: JSON.stringify({
          handle: "add",
          data: new RoomConfigInfo(vm.addForm),
          // data: {
          //   RoomID: vm.addForm.RoomID,
          //   RecordMode: vm.addForm.RecordMode,
          //   StartTime: date2string(vm.addForm.StartTime),
          //   EndTime: date2string(vm.addForm.EndTime),
          //   AutoRecord: vm.addForm.AutoRecord,
          //   AutoUpload: vm.addForm.AutoUpload,
          //   NeedM4a: vm.addForm.NeedM4a,
          //   Mp4Compress: vm.addForm.Mp4Compress,
          //   DivideByTitle: vm.addForm.DivideByTitle,
          //   CleanUpRegular: vm.addForm.CleanUpRegular,
          //   SaveDuration: vm.addForm.SaveDuration,
          //   AreaLock: vm.addForm.AreaLock,
          //   AreaLimit: vm.addForm.AreaLimit
          // },
        }),
        headers: {
          "Content-Type": "application/json"
        },
        success: function (msg) {
          console.log(msg)
        }
      })
    },
    test(scope) {
      console.log(scope)
    }
  }
};

const app = Vue.createApp(Main);
app.use(ElementPlus);

const vm = app.mount("#app");

// 获取低频数据
function getLowFrqData() {
  let data = Array(0)
  $.ajax({
    type: "POST",
    url: "/live-info",
    data: {},
    success: function (msg) {
      console.log(msg)
      // TODO: 扩展接口，获取其他信息
      let recording = 0;
      for (let key in msg) {
        let date = new Date(parseInt(msg[key].LiveStartTime) * 1000);
        let item = {
          RoomID: msg[key].RoomID,
          Uname: msg[key].Uname,
          AreaName: msg[key].AreaName,
          Title: msg[key].Title,
          LiveStatus: msg[key].LiveStatus,
          LiveStartTime: date2string(date),
          LiveStartTime2: msg[key].LiveStartTime,
          State: msg[key].State,
          StartTime: msg[key].StartTime,
          EndTime: msg[key].EndTime,
          RecordMode: msg[key].RecordMode,
          AutoRecord: msg[key].AutoRecord,
          AutoUpload: msg[key].AutoUpload,
          RecordStartTime: msg[key].RecordStartTime,
          RecordEndTime: msg[key].RecordEndTime,
          DecodeStartTime: msg[key].DecodeStartTime,
          DecodeEndTime: msg[key].DecodeEndTime,
          UploadStartTime: msg[key].UploadStartTime,
          UploadEndTime: msg[key].UploadEndTime,
          NeedM4a: msg[key].NeedM4a,
          Mp4Compress: msg[key].Mp4Compress,
          DivideByTitle: msg[key].DivideByTitle,
          CleanUpRegular: msg[key].CleanUpRegular,
          SaveDuration: msg[key].SaveDuration,
          AreaLock: msg[key].AreaLock,
          AreaLimit: msg[key].AreaLimit,
        }
        if (item.State == 3) {
          recording++
        }
        data.push(item)
      }
      vm.tableData = data
      vm.recordCount = recording
      flushData()
    }
  })
}

// 实时刷新
function flushData() {
  for (let key in vm.tableData) {
    vm.tableData[key].LiveTime = getTimeMiuns(vm.tableData[key].LiveStartTime2, 0)
    vm.tableData[key].RecordTime = getTimeMiuns(vm.tableData[key].RecordStartTime, vm.tableData[key].RecordEndTime)
    vm.tableData[key].DecodeTime = getTimeMiuns(vm.tableData[key].DecodeStartTime, vm.tableData[key].DecodeEndTime)
    vm.tableData[key].UploadTime = getTimeMiuns(vm.tableData[key].UploadStartTime, vm.tableData[key].UploadEndTime)
  }
}

// 获取低频数据
function getHighFrqData() {
  let data = Array(0)
  $.ajax({
    type: "POST",
    url: "/base-info",
    data: {},
    success: function (msg) {
      vm.totalDownload = getReadableSizeString(msg.TotalDownload)
      vm.fileNum = msg.FileNum
      vm.cpuPct = parseFloat(msg.DeviceInfo.TotalCpuUsage.toFixed(2))
      vm.memoryUsage = getReadableSizeString(msg.DeviceInfo.MemUsage)
      vm.memoryTotal = getReadableSizeString(msg.DeviceInfo.MemTotal)
      vm.memoryPct = parseFloat((msg.DeviceInfo.MemUsage * 100 / msg.DeviceInfo.MemTotal).toFixed(2))
      vm.diskUsage = getReadableSizeString(msg.DeviceInfo.DiskUsage)
      vm.diskTotal = getReadableSizeString(msg.DeviceInfo.DiskTotal)
      vm.diskName = msg.DeviceInfo.DiskName
      vm.diskPct = parseFloat((msg.DeviceInfo.DiskUsage * 100 / msg.DeviceInfo.DiskTotal).toFixed(2))
      vm.uploadSpeed = getReadableSizeString(msg.DeviceInfo.NetUpPerSec) + "/s👆"
      vm.downloadSpeed = getReadableSizeString(msg.DeviceInfo.NetDownPerSec) + "/s👇"
    }
  })
}

function stillTwo(num) {
  return ("0" + num).substr(-2);
}

function date2string(date) {
  return `${date.getFullYear()}-${stillTwo(date.getMonth() + 1)}-${stillTwo(date.getDate())} ${stillTwo(date.getHours())}:${stillTwo(date.getMinutes())}:${stillTwo(date.getSeconds())}`
}

function getTimeMiuns(st, et) {
  let nTime = 0;
  if (parseInt(st) < parseInt(et)) {
    nTime = parseInt(et) - parseInt(st);
  } else if (parseInt(st) > 0) {
    let dn = new Date();
    let start = new Date(parseInt(st) * 1000);
    nTime = dn.getTime() - start.getTime();
  }
  nTime = Math.floor(nTime / 1000);
  let day = Math.floor(nTime / 86400);
  let hour = Math.floor(nTime % 86400 / 3600);
  let minute = Math.floor(nTime % 86400 % 3600 / 60);
  let second = nTime % 60;
  return `${day}天 ${hour}时 ${minute}分 ${second} 秒`
}

function getReadableSizeString(fileSizeInBytes) {
  var i = -1;
  var byteUnits = [' K', ' M', ' G', ' T', 'P', 'E', 'Z', 'Y'];
  do {
    fileSizeInBytes = fileSizeInBytes / 1024;
    i++;
  } while (fileSizeInBytes > 1024);

  return Math.max(fileSizeInBytes, 0.1).toFixed(1) + byteUnits[i];
};

getLowFrqData()
getHighFrqData()