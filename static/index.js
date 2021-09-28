const { defineComponent, ref } = Vue;

var Main = {
  data() {
    const tableData = Array(0)
    return {
      tableData,
    }
  },
};

const app = Vue.createApp(Main);
app.use(ElementPlus);
const vm = app.mount("#app");

var data = Array(0)
$.ajax({
  type: "POST",
  url: "/",
  data: {},
  success: function (msg) {
    console.log(msg)
    for (var key in msg) {
      let date = new Date(parseInt(msg[key].LiveStartTime) * 1000);
      var item = {
        RoomID: msg[key].RoomID,
        Uname: msg[key].Uname,
        AreaName: msg[key].AreaName,
        Title: msg[key].Title,
        LiveStatus: msg[key].LiveStatus,
        LiveStartTime: `${date.getFullYear()}-${stillTwo(date.getMonth() + 1)}-${stillTwo(date.getDate())} ${stillTwo(date.getHours())}:${stillTwo(date.getMinutes())}:${stillTwo(date.getSeconds())}`,
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

function stillTwo(num) {
  return ("0" + num).substr(-2);
}

function getTimeMiuns(st, et) {
  let nTime = 0;
  if (parseInt(st) < parseInt(et)) {
    nTime = parseInt(et) - parseInt(st);
  } else if (parseInt(st) >0){
    let dn = new Date();
    let start = new Date(parseInt(st) * 1000);
    nTime = dn.getTime() - start.getTime();
  }
  let day = Math.floor(nTime / 86400);
  let hour = Math.floor(nTime % 86400 / 3600);
  let minute = Math.floor(nTime % 86400 % 3600 / 60);
  let second = nTime % 60;
  return `${day}天 ${hour}时 ${minute}分 ${second} 秒`
}