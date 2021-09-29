const { defineComponent, ref } = Vue;
const { ElNotification } = ElementPlus;

let Main = {
  data() {
    return {
      tableData: [],
      recordCount: 0,
      cpu: '10%',
      memoryUsage: '7.9G',
      memoryTotal: '16G',
      uploadSpeed: '15.6K/s↑',
      downloadSpeed: '1.3M/s↓',
      diskUsage: '234G',
      diskTotal: '500G',
      totalDownload: '',
      fileNum: 0,
      editFormVisible: false,
      form: {
        RoomID: '123',
        RecordMode: false,
        StartTime: new Date(),
        EndTime: new Date(),
        AutoRecord: true,
        AutoUpload: false,
      },
      disabledRecord: false,
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
      return v1-v2
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
    handleDelete(index, row) {
      console.log(row)
      ElNotification({
        title: '成功',
        message: '删除房间成功',
        type: 'success'
      })
    },
    handleEdit(index, row) {
      vm.editFormVisible = true
      vm.form = {
        RoomID: row.RoomID,
        RecordMode: false,
        StartTime: new Date(2021,9,29,row.StartTime.slice(0,2), row.StartTime.slice(2,4), row.StartTime.slice(4,6)),
        EndTime: new Date(2021,9,29,row.EndTime.slice(0,2), row.EndTime.slice(2,4), row.EndTime.slice(4,6)),
        AutoRecord: row.AutoRecord,
        AutoUpload: row.AutoUpload,
      }
      console.log(vm.form)
    },
    onSubmit() {
      this.editFormVisible = false
      console.log('submit!')
      // TODO: Ajax回传给go处理
    },
    changeRecordTimeText(recordMode) {
      this.disabledRecord = recordMode
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
      // console.log(vm.tableData)
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
          LiveStartTime: `${date.getFullYear()}-${stillTwo(date.getMonth() + 1)}-${stillTwo(date.getDate())} ${stillTwo(date.getHours())}:${stillTwo(date.getMinutes())}:${stillTwo(date.getSeconds())}`,
          LiveStartTime2: msg[key].LiveStartTime,
          // LiveTime: getTimeMiuns(msg[key].LiveStartTime, 0),
          // RecordTime: getTimeMiuns(msg[key].RecordStartTime, msg[key].RecordEndTime),
          // DecodeTime: getTimeMiuns(msg[key].DecodeStartTime, msg[key].DecodeEndTime),
          // UploadTime: getTimeMiuns(msg[key].UploadStartTime, msg[key].UploadEndTime),
          State: msg[key].State,
          StartTime: msg[key].StartTime,
          EndTime: msg[key].EndTime,
          RecordMode: false,
          RecordStartTime: msg[key].RecordStartTime,
          RecordEndTime: msg[key].RecordEndTime,
          DecodeStartTime: msg[key].DecodeStartTime,
          DecodeEndTime: msg[key].DecodeEndTime,
          UploadStartTime: msg[key].UploadStartTime,
          UploadEndTime: msg[key].UploadEndTime,
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
    success: function(msg) {
      vm.totalDownload = getReadableFileSizeString(msg.TotalDownload)
      vm.fileNum = msg.FileNum
    }
  })
}

function stillTwo(num) {
  return ("0" + num).substr(-2);
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

function getReadableFileSizeString(fileSizeInBytes) {
  var i = -1;
  var byteUnits = [' kB', ' MB', ' GB', ' TB', 'PB', 'EB', 'ZB', 'YB'];
  do {
      fileSizeInBytes = fileSizeInBytes / 1024;
      i++;
  } while (fileSizeInBytes > 1024);

  return Math.max(fileSizeInBytes, 0.1).toFixed(2) + byteUnits[i];
};

getLowFrqData()