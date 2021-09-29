const { defineComponent, ref } = Vue;

var Main = {
  data() {
    return {
      tableData: [],
      recordCount: 0,
      cpu: '10%',
      memoryUsage: '7.9G',
      memoryTotal: '16G',
      uploadSpeed: '15.6kb/s↑',
      downloadSpeed: '1.3Mb/s↓',
      diskUsage: '234G',
      diskTotal: '500G'
    }
  },
  mounted() {
    this.timer = setInterval(getData, 1000);
  },
  beforeDestory() {
    clearInterval(this.timer);
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
    }
  }
};

const app = Vue.createApp(Main);
app.use(ElementPlus);

const vm = app.mount("#app");

function getData() {
  var data = Array(0)
  $.ajax({
    type: "POST",
    url: "/live",
    data: {},
    success: function (msg) {
      // console.log(vm.tableData)
      let recording = 0;
      for (var key in msg) {
        let date = new Date(parseInt(msg[key].LiveStartTime) * 1000);
        var item = {
          RoomID: msg[key].RoomID,
          Uname: msg[key].Uname,
          AreaName: msg[key].AreaName,
          Title: msg[key].Title,
          LiveStatus: msg[key].LiveStatus,
          LiveStartTime: `${date.getFullYear()}-${stillTwo(date.getMonth() + 1)}-${stillTwo(date.getDate())} ${stillTwo(date.getHours())}:${stillTwo(date.getMinutes())}:${stillTwo(date.getSeconds())}`,
          LiveTime: getTimeMiuns(msg[key].LiveStartTime, 0),
          RecordTime: getTimeMiuns(msg[key].RecordStartTime, msg[key].RecordEndTime),
          DecodeTime: getTimeMiuns(msg[key].DecodeStartTime, msg[key].DecodeEndTime),
          UploadTime: getTimeMiuns(msg[key].UploadStartTime, msg[key].UploadEndTime),
          State: msg[key].State
        }
        if (item.State == 3) {
          recording++
        }
        data.push(item)
      }
      vm.tableData = data
      vm.recordCount = recording
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

getData()