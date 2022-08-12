'use strict';

let TerminalApp = angular.module('TerminalApp', ['f2c.common','f2c.filter']);

TerminalApp.controller("BaseTerminalController", function ($scope,HttpUtils) {
    $scope.sessionId = window.sessionId
    $scope.podDetail = window.podDetail
    $scope.shell = window.shell
});

TerminalApp.directive("bashTerminal", function (Notification, HttpUtils, Translator, $timeout, $window) {
    return {
        restrict: 'A',
        templateUrl: '/static/app/html/pod-terminal.html' + '?_t=' + Math.random(),
        scope: {
            shell: "=",
            sessionId: "=",
            podDetail: "="
        },
        link: function ($scope) {
            $scope.socket = null;

            $scope.init = function (){
                $scope.podDetail = angular.fromJson($scope.podDetail);
                $scope.initContainer();
                $scope.initCommand();
                $scope.initSocket();
            }

            $scope.initContainer = function (){
                $scope.containerList = [];
                $scope.podDetail.spec.containers.forEach(item => {
                    $scope.containerList.push(item.name);
                });
                if($scope.containerList.length > 0 && !$scope.containerName){
                    $scope.containerName = $scope.containerList[0];
                }
            }

            $scope.initCommand = function (){
                if($scope.shell === 'exec'){
                    $scope.commandName = "bash";
                    $scope.commandList = ["bash","sh"];
                }else{
                    $scope.log = {
                        follow: true
                    }
                    $scope.tailLines = 20;
                    $scope.logRowList = [20, 100, 200, 500]
                }
            }

            $scope.initSocket = function () {

                let url = window.location.origin
                if($scope.shell === 'exec'){
                    url += "/terminal/sockjs?sessionId=" + $scope.sessionId + "&container=" + $scope.containerName + "&cmd=" + $scope.commandName;
                } else {
                    url += "/logging/sockjs?sessionId=" + $scope.sessionId + "&container=" + $scope.containerName + "&follow=" + $scope.log.follow + "&tailLines=" + $scope.tailLines;
                }
                $scope.socket = new SockJS(url);
                $scope.socketFunc();
            }

            $scope.socketFunc = function () {
                $scope.socket.onopen = () => {
                    $scope.initTerm();
                };

                $scope.socket.onerror = (e) => {
                    Notification.warn("socket 链接失败");
                };
                $scope.socket.onmessage = function (msg) {
                    $scope.term.write(msg.data);
                };
            }

            $scope.initTerm = function (errorMsg) {
                $scope.term = new Terminal({
                    rendererType: "canvas", //渲染类型
                    // rows: rows, //行数
                    // cols: cols,// 设置之后会输入多行之后覆盖现象
                    convertEol: true, //启用时，光标将设置为下一行的开头
                    // scrollback: 10,//终端中的回滚量
                    fontSize: 14, //字体大小
                    windowsMode: true,
                    disableStdin: false, //是否应禁用输入。
                    cursorStyle: "underline", //光标样式
                    cursorBlink: true, //光标闪烁
                    scrollback: 30,
                    tabStopWidth: 4,
                    theme: {
                        foreground: "#06ff06", //字体
                        background: "#060101", //背景色
                        cursor: "help" //设置光标
                    }
                });
                $scope.term.open(document.getElementById("terminal"));
                $scope.term.focus();
                let fitAddon = new FitAddon.FitAddon();
                $scope.term.loadAddon(fitAddon);
                fitAddon.fit();
                if(errorMsg){
                    $scope.term.write(errorMsg);
                } else {
                    $scope.term.onData(function (key) {
                        $scope.socket.send(key);
                    });
                }
            }

            $scope.changeActive = function (){
                $timeout(()=>{
                    $scope.reloadTerm();
                }, 200);
            }
            $scope.containerChange = function (obj){
                $scope.containerName = obj.containerName;
                $scope.reloadTerm();
            }
            $scope.commandChange = function (obj){
                $scope.commandName = obj.commandName;
                $scope.reloadTerm();
            }
            $scope.logRowChange = function (obj){
                $scope.tailLines = obj.tailLines;
                $scope.reloadTerm();
            }

            $scope.reloadTerm = function (){
                if($scope.socket){
                    $scope.socket.close();
                }
                $scope.terminal = document.getElementById("terminal");
                $scope.terminal.innerHTML = "";
                $scope.initSocket();
            }

            $scope.init();
        }
    }
});
