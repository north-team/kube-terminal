/**
 * 加载自定义模块
 */
(function () {
    angular.module('f2c.module', [
        'ngAnimate',
        'ngMaterial',
        'ngMessages',
        'pascalprecht.translate'
    ]);

    let F2CCommon = angular.module('f2c.common', ['f2c.filter', 'f2c.service', 'f2c.module']);

    F2CCommon.config(function ($mdAriaProvider, $httpProvider) {
        $mdAriaProvider.disableWarnings();
        $httpProvider.interceptors.push(['$q', '$injector', function ($q, $injector) {
            return {
                response: function (response) {
                    let deferred = $q.defer();
                    let Translator = $injector.get('Translator');
                    if (response.headers("Authentication-Status") === "invalid" || (typeof (response.data) === "string" && response.data.indexOf("action=\"") > -1 && response.data.indexOf("<form") > -1 && response.data.indexOf("/form>") > -1)) {
                        deferred.reject('invalid session');
                        if (!window.parent.sessionInvalid) {
                            window.parent.sessionInvalid = true;
                            Translator.wait(function () {
                                let $mdDialog = $injector.get("$mdDialog");
                                $mdDialog.show($mdDialog.alert()
                                    .clickOutsideToClose(true)
                                    .title(Translator.get("i18n_warn"))
                                    .textContent(Translator.get("i18n_login_expired"))
                                    .ok(Translator.get("i18n_ok"))
                                ).then(function () {
                                    window.parent.parent.location.href = "/logout";
                                });
                            });
                        }
                    } else {
                        deferred.resolve(response);
                    }
                    return deferred.promise;
                }
                ,
                request: function (request) {
                    //resolve ie cache issue
                    if (request.method === "GET" && request.headers.Accept.indexOf("json") !== -1 && request.url.indexOf(".") === -1) {
                        let d = new Date();
                        if (request.url.indexOf("?") === -1) {
                            request.url = request.url + '?_nocache=' + d.getTime();
                        } else {
                            request.url = request.url + '&_nocache=' + d.getTime();
                        }
                    }
                    return request;
                },
                responseError: function (err) {
                    let Notification = $injector.get("Notification");
                    if (-1 === err.status) {
                        // 远程服务器无响应
                        Notification.danger("No response from remote server", null, {delay: 5000});
                    }
                    if (502 === err.status) {
                        // Bad Gateway
                        Notification.danger("Bad Gateway", null, {delay: 5000});
                    }
                    return $q.reject(err);
                }
            };
        }]);
    });

    F2CCommon.config(function ($mdThemingProvider) {

        if (window.parent.IndexConstants) {
            $mdThemingProvider.definePalette('primary', Palette.primary(window.parent.IndexConstants['ui.theme.primary']));
            $mdThemingProvider.theme('default').primaryPalette('primary');

            $mdThemingProvider.definePalette('accent', Palette.accent(window.parent.IndexConstants['ui.theme.accent']));
            $mdThemingProvider.theme('default').accentPalette('accent');
        }

        $mdThemingProvider.definePalette('white', {
            '50': '#ffffff',
            '100': '#f5f5f5',
            '200': '#eeeeee',
            '300': '#e0e0e0',
            '400': '#bdbdbd',
            '500': '#9e9e9e',
            '600': '#757575',
            '700': '#616161',
            '800': '#424242',
            '900': '#212121',
            'A100': '#fafafa',
            'A200': '#000000',
            'A400': '#303030',
            'A700': '#616161',
            'contrastDefaultColor': 'dark',
            'contrastLightColors': '600 700 800 900 A200 A400 A700'
        });

        $mdThemingProvider.theme('default').backgroundPalette('white');
    });

    F2CCommon.config(function ($mdDateLocaleProvider) {
        let Translator = window.parent[window.parent.userLocale];
        if (Translator) {
            $mdDateLocaleProvider.months = Translator.months;
            $mdDateLocaleProvider.shortMonths = Translator.shortMonths;
            $mdDateLocaleProvider.days = Translator.days;
            $mdDateLocaleProvider.shortDays = Translator.shortDays;
            $mdDateLocaleProvider.msgCalendar = Translator.msgCalendar;
            $mdDateLocaleProvider.msgOpenCalendar = Translator.msgOpenCalendar;
        }

        $mdDateLocaleProvider.firstDayOfWeek = 0;

        $mdDateLocaleProvider.parseDate = function (dateString) {
            let m = moment(dateString, 'YYYY-MM-DD', true);
            return m.isValid() ? m.toDate() : new Date(NaN);
        };

        $mdDateLocaleProvider.formatDate = function (date) {
            if (!date) return '';
            let m = moment(date);
            return m.isValid() ? m.format('YYYY-MM-DD') : '';
        };

        $mdDateLocaleProvider.monthHeaderFormatter = function (date) {
            return date.getFullYear() + ' ' + $mdDateLocaleProvider.months[date.getMonth()];
        };
    });

}());

(function () {
    let F2CFilter = angular.module('f2c.filter', []);
    F2CFilter.filter('translator', function ($translate) {
        // 翻译过滤器，兼容没有使用国际化的情况
        function translator(input, defaultStr) {
            try {
                let str = $translate.instant(input);
                // let str = input;
                // $translate(input, defaultStr).then(function (resp) {
                //     // console.log("test" + resp);
                //     str = resp;
                // }, function (err) {
                //     // console.log(err)
                // });
                if (str === input && defaultStr) {
                    return defaultStr;
                }
                if (str === input && input.startsWith("i18n_")) {
                    return str === input ? "" : str;
                }
                return str;
            } catch (e) {
                return defaultStr ? defaultStr : input;
            }
        }

        if ($translate.statefulFilter()) {
            translator.$stateful = true;
        }

        return translator;
    });

    F2CFilter.service('Translator', function ($filter, $translate) {
        this.get = function (key) {
            return $filter('translator')(key);
        };

        this.gets = function (key) {
            return $filter('translators')(key);
        };

        this.wait = function (func) {
            $translate("i18n_i18n").then(func);
        };

        this.setLang = function (lang) {
            $translate.use(lang);
        }
    });
}());


(function () {
    let F2CService = angular.module('f2c.service', []);
    F2CService.service('Notification', function ($mdToast, $mdDialog, $compile, $rootScope, $document, Translator) {
        let template = "<notice messages='messages'></notice>";
        let scope = $rootScope.$new();
        scope.messages = {
            left: [],
            center: [],
            right: []
        };

        function add(msg, type, callback, option) {
            let item = angular.extend({
                msg: msg,
                close: callback,
                position: "notify-right",
                type: type,
                delay: 2000
            }, option);

            switch (item.position) {
                case "notify-left":
                    scope.messages.left.unshift(item);
                    break;
                case "notify-center":
                    scope.messages.center.unshift(item);
                    break;
                case "notify-right":
                    scope.messages.right.unshift(item);
                    break;
                default:
                    scope.messages.right.unshift(item);
            }

            let notify = $("#_notification");
            if (notify.length === 0) {
                angular.element("md-content[ui-view]").append($compile(template)(scope));
            }
        }

        this.show = function (msg, callback, option) {
            add(msg, "notify-default", callback, option);
        };

        this.info = function (msg, callback, option) {
            add(msg, "notify-primary", callback, option);
        };

        this.success = function (msg, callback, option) {
            add(msg, "notify-success", callback, option);
        };

        this.warn = function (msg, callback, option) {
            add(msg, "notify-warn", callback, angular.extend({
                delay: 10000
            }, option));
        };

        this.danger = function (msg, callback, option) {
            add(msg, "notify-danger", callback, angular.extend({
                delay: 30000
            }, option));
        };

        this.alert = function (msg) {
            $mdDialog.show($mdDialog.alert()
                .multiple(true)
                .clickOutsideToClose(true)
                .title(Translator.get("i18n_warn"))
                .textContent(msg)
                .ok(Translator.get("i18n_ok"))
            );
        };

        this.confirm = function (msg, success, cancel) {
            let confirm = $mdDialog.confirm()
                .multiple(true)
                .title(Translator.get("i18n_confirm"))
                .textContent(msg)
                .ok(Translator.get("i18n_ok"))
                .cancel(Translator.get("i18n_cancel"));

            $mdDialog.show(confirm).then(function () {
                if (angular.isFunction(success)) success();
            }, function () {
                if (angular.isFunction(cancel)) cancel();
            });
        };

        this.prompt = function (obj, success, cancel) {
            let locals = {
                title: obj.title,
                text: obj.text,
                placeholder: obj.placeholder,
                init: obj.init,
                required: angular.isUndefined(obj.required) ? true : obj.required,
                showInput: angular.isUndefined(obj.showInput) ? true : obj.showInput,

                selectValue: obj.selectValue,
                selectKey: obj.selectKey,
                selectItems: obj.selectItems,
                selectText: obj.selectText,
                selectRequired: obj.selectRequired || false,

                multiSelectValue: obj.multiSelectValue,
                multiSelectKey: obj.multiSelectKey,
                multiSelectItems: obj.multiSelectItems,
                multiSelectText: obj.multiSelectText,
                multiSelectRequired: obj.multiSelectRequired || false
            };
            $mdDialog.show({
                multiple: true,
                templateUrl: "/web-public/fit2cloud/html/notice/prompt.html" + '?_t=' + window.appversion,
                parent: angular.element($document[0].body),
                controller: function ($scope, $mdDialog, title, text, placeholder, init, required, showInput, selectRequired, selectValue, selectKey, selectItems, selectText, multiSelectRequired, multiSelectValue, multiSelectKey, multiSelectItems, multiSelectText) {
                    $scope.title = title;
                    $scope.text = text;
                    $scope.placeholder = placeholder;
                    $scope.init = init;
                    $scope.required = required;
                    $scope.showInput = showInput;
                    $scope.value = null;

                    $scope.selected = null;
                    $scope.selectValue =  selectValue;
                    $scope.selectKey =  selectKey;
                    $scope.selectItems =  selectItems;
                    $scope.selectText =  selectText;
                    $scope.showSelect = angular.isArray(selectItems) ? true : false;
                    $scope.selectRequired = selectRequired;

                    $scope.multiSelected = null;
                    $scope.multiSelectValue =  multiSelectValue;
                    $scope.multiSelectKey =  multiSelectKey;
                    $scope.multiSelectItems =  multiSelectItems;
                    $scope.multiSelectText =  multiSelectText;
                    $scope.showMultiSelect = angular.isArray(multiSelectItems) ? true : false;
                    $scope.multiSelectRequired = multiSelectRequired;


                    $scope.close = function () {
                        $mdDialog.cancel();
                    };

                    $scope.ok = function () {
                        if($scope.selectItems || $scope.multiSelectItems){
                            let result = {
                                value: $scope.value,
                                selected: $scope.selected,
                                multiSelected: $scope.multiSelected
                            };
                            $mdDialog.hide(result);
                        }else {
                            $mdDialog.hide($scope.value);
                        }
                    }
                },
                locals: locals,
                clickOutsideToClose: false
            }).then(function (result) {
                if (angular.isFunction(success)) success(result);
            }, function () {
                if (angular.isFunction(cancel)) cancel();
            });
        };

    });

    F2CService.service('Loading', function ($q) {
        let promises = [];
        this.add = function (promise) {
            promises.push(promise);
        };

        this.load = function () {
            let promises_q = $q.all(promises);
            promises = [];
            return promises_q;
        }
    });

    F2CService.service('HttpUtils', function ($http, Notification, $log) {
        this.get = function (url, success, error, config) {
            return $http.get(url, config).then(function (response) {
                if (!response.data) {
                    //处理不是ResultHolder的结果
                    return success(response);
                } else if (response.data.success) {
                    return success(response.data);
                } else {
                    if (error) {
                        error(response.data);
                    } else {
                        Notification.danger(response.data.message);
                        $log.error(response);
                    }
                }
            }, function (response) {
                if (error) {
                    error(response.data);
                } else {
                    Notification.danger(response.data.message);
                    $log.error(response);
                }
            });
        };

        this.post = function (url, data, success, error, config) {
            return $http.post(url, data, config).then(function (response) {
                if (!response.data) {
                    //处理不是ResultHolder的结果
                    return success(response);
                } else if (response.data.success) {
                    return success(response.data);
                } else {
                    if (error) {
                        error(response.data);
                    } else {
                        Notification.danger(response.data.message);
                        $log.error(response);
                    }
                }
            }, function (response) {
                if (error) {
                    error(response.data);
                } else {
                    Notification.danger(response.data.message);
                    $log.error(response);
                }
            });
        };

        this.delete = function (url, success, error, config) {
            return $http.delete(url, config).then(function (response) {
                if (!response.data) {
                    //处理不是ResultHolder的结果
                    success(response);
                } else if (response.data.success) {
                    success(response.data);
                } else {
                    if (error) {
                        error(response.data);
                    } else {
                        Notification.danger(response.data.message);
                        $log.error(response);
                    }
                }
            }, function (response) {
                if (error) {
                    error(response.data);
                } else {
                    Notification.danger(response.data.message);
                    $log.error(response);
                }
            });
        };

        /**
         * 前端分页， 因为返回值的限制，目前只适用于安全组
         * @param $scope
         * @param url
         * @param data
         * @param callBack
         * @param error
         */
        this.frontPaging = function ($scope, url, data, callBack, error) {
            let self = this;
            let method = callBack;
            let result = function (response) {
                $scope.itemHasData = true; // 标识是否请求数据
                response.data["dataList"].forEach(function(value){
                    value.itemUniqueId = (window.crypto.getRandomValues(new Uint32Array(1))).toString(36);
                });
                $scope.originItems = response.data["dataList"];
                $scope.frontPagination.itemCount = response.data["dataList"].length;
                if (response.data["dataList"].length > 0) {
                    $scope.frontPagination.show = true;
                } else {
                    $scope.frontPagination.show = false;
                }
                self.frontPopulatePagination($scope, $scope.originItems, response, method);
            };
            $scope.frontPagination = angular.extend({
                page: 1,
                limit: 10,
                limitOptions: [10, 20, 50, 100]
            }, $scope.frontPagination);

            $scope.frontPagination.onPaginate = function () {
                $scope.frontItems = [];
                if(!$scope.itemHasData){
                    if (data !== undefined && data !== null) {
                        $scope.loadingLayer = self.post(url, data, result, error);
                    } else {
                        $scope.loadingLayer = self.get(url, result, error);
                    }
                }else {
                    self.frontPopulatePagination($scope, $scope.originItems, result);
                }
            };
            $scope.frontPagination.onPaginate();
        };

        /**
         * 前端分页 拆分数据组件
         * @param $scope
         * @param data
         */
        this.frontPopulatePagination = function ($scope, data, response, callBack) {
            let self = this;
            let method = callBack;
            let startRow = $scope.frontPagination.page > 0 ? ($scope.frontPagination.page - 1) * $scope.frontPagination.limit : 0;
            let endRow = startRow + $scope.frontPagination.limit * ($scope.frontPagination.page > 0 ? 1 : 0);
            $scope.frontItems = data.slice(startRow, endRow);
            if (data.length > 0) {
                $scope.frontPagination.show = true;
                $scope.frontPagination.pageCount = ($scope.frontPagination.itemCount / $scope.frontPagination.limit + (($scope.frontPagination.itemCount %  $scope.pagination.limit == 0) ? 0 : 1)) ;
            } else {
                $scope.frontPagination.show = false;
            }
            if (angular.isFunction(method)) {
                method(response);
            }
        };

        this.paging = function ($scope, url, data, callBack, error) {
            let self = this;
            let method = callBack;
            let result = function (response) {
                $scope.items = response.data["listObject"];
                $scope.pagination.formData = response.data["param"];
                $scope.pagination.itemCount = response.data["itemCount"];
                $scope.pagination.pageCount = response.data['pageCount'];
                if ($scope.pagination.pageCount > 0) {
                    $scope.pagination.show = true;
                    $scope.pagination.pageCount = response.data['pageCount'];
                } else {
                    $scope.pagination.show = false;
                }
                if (angular.isFunction(method)) {
                    method(response);
                }
            };
            $scope.pagination = angular.extend({
                page: 1,
                limit: 10,
                limitOptions: [10, 20, 50, 100]
            }, $scope.pagination);

            $scope.pagination.onPaginate = function () {
                $scope.items = [];
                if (data !== undefined && data !== null) {
                    $scope.loadingLayer = self.post(url + "/" + $scope.pagination.page + "/" + $scope.pagination.limit, data, result, error);
                } else {
                    $scope.loadingLayer = self.get(url + "/" + $scope.pagination.page + "/" + $scope.pagination.limit, result, error);
                }

            };

            $scope.pagination.onPaginate();
        };

        this.download = function (url, data, filename, mime) {
            return $http({
                url: url,
                method: 'post',
                responseType: 'arraybuffer',
                data: data
            }).then(function (response) {
                let blob = new Blob([response.data], {type: mime});
                saveAs(blob, filename);
            }, function (response) {
                Notification.danger('file download error.');
                $log.error(response);
            });
        }
    });
}());