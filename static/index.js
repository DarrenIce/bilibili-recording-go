const { defineComponent, ref } = Vue;

var Main = {
  data() {
    return {
      tableData: [],
      tabs: [
        {
          'id': 'overview',
          'name': '总览',
        },
        {
          'id': 'rooms',
          'name': '房间管理',
        },
      ],
      currentTab: 'overview'
    }
  },
  mounted() {
    this.timer = setInterval(getData, 1000);
  },
  beforeDestory() {
    clearInterval(this.timer);
  },
  computed: {
    currentView() {
      return 'view-' + this.currentTab
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
    url: "/",
    data: {},
    success: function (msg) {
      // console.log(msg)
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
          DecodeTime: getTimeMiuns(msg[key].DecodeEndTime, msg[key].DecodeEndTime),
          UploadTime: getTimeMiuns(msg[key].UploadEndTime, msg[key].UploadEndTime),
          State: msg[key].State
        }
        data.push(item)
      }
      vm.tableData = data
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

app.component('view-overview', {
  template: `<el-main>
  <el-table :data="tableData">
      <el-table-column prop="RoomID" label="房间ID" width="140">
      </el-table-column>
      <el-table-column prop="Uname" label="主播" width="120">
      </el-table-column>
      <el-table-column prop="AreaName" label="分区">
      </el-table-column>
      <el-table-column prop="Title" label="直播标题">
      </el-table-column>
      <el-table-column prop="LiveStatus" label="直播状态">
      </el-table-column>
      <el-table-column prop="LiveStartTime" label="开播时间">
      </el-table-column>
      <el-table-column prop="LiveTime" label="开播时长">
      </el-table-column>
      <el-table-column prop="RecordTime" label="录制时间">
      </el-table-column>
      <el-table-column prop="DecodeTime" label="转码用时">
      </el-table-column>
      <el-table-column prop="UploadTime" label="上传用时">
      </el-table-column>
      <el-table-column prop="State" label="当前状态">
      </el-table-column>
  </el-table>
</el-main>`
})

app.component('view-rooms', {
  template: ``
})