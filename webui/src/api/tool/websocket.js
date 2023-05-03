import { getToken } from '@/utils/auth'
import {showNotify} from "@/api/default";

let webSocket = null;
let open = false;

/**
 * 发送消息
 * @param msg
 */
export function webSocketSend(msg){
  if (open){
    webSocket.send(msg);
  }
}

/**
 * 连接webSocket
 */
export function webSocketConnect(that){
  console.info("连接socket")
  let token = getToken();
  let hostname = window.location.hostname;
  if ("WebSocket" in window) {
    webSocket = new WebSocket("ws://" + hostname + ":8080/msg?userId=" + token);
    webSocket.onopen = function () {
      console.log("连接成功！");
      open = true
    };
    webSocket.onmessage = function (evt) {
      let data = evt.data;
      console.log("接受消息：" + data);
      let socketMsg = JSON.parse(data);
      showNotify(that, "系统消息",  socketMsg.msg, socketMsg.type, 0);
    };
    webSocket.onclose = function () {
      console.log("连接关闭！");
      setTimeout(function () {
        console.log("重新连接")
        webSocketConnect(that);
      }, 2000);
    };
    webSocket.onerror = function () {
      console.log("连接异常！");
    };
  }
}
