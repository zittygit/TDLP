(function() {
    'use strict';

    function csrfSafeMethod(method) {
        // these HTTP methods do not require CSRF protection
        return (/^(GET|HEAD|OPTIONS|TRACE)$/.test(method));
    }

    var csrftoken = $("[name=csrfmiddlewaretoken]").val();
    $.ajaxSetup({
        beforeSend: function(xhr, settings) {
            if (!csrfSafeMethod(settings.type) && !this.crossDomain) {
                xhr.setRequestHeader("X-CSRFToken", csrftoken);
            }
        }
    });


    function closeNotify() {
        $.notifyClose();
    }

    // system_usage.html
    if ( $( "#system_resource" ).length > 0 ) {
        var cpu_src, mem_src, net_src, cpu_solo_src, mem_solo_src, file_src;
        var nodename;
        var nodeactive = "overall";
        var time_from, time_to;
        var time_interval = 60*60*1000;//default: an hour
        time_to = new Date().getTime();
        time_from = time_to - time_from;
        cpu_solo_src = "../grafana_data/dashboard-solo/db/overall?&panelId=1&theme=light&width=150&height=150";
        mem_solo_src = "../grafana_data/dashboard-solo/db/overall?&panelId=2&theme=light&width=150&height=150";
        $('#cpu_usage').attr('src', cpu_solo_src);
        $('#mem_usage').attr('src', mem_solo_src);
        updateCharts();
    }

    $("#toggleNode").click(function(e){
        nodeactive  = e.target.id;
        updateCharts();
    });

    $("#cpu").on("load", function(e){
        $("#refresh-vis").find("i").removeClass("fa-spin");
    });


    $("#refresh-vis").click(function(e){
        updateCharts();
    });

    $("#toggleTime").click(function(e){
        var interval =  e.target.id;
        switch (interval){
            case "5min":
                time_interval = 5*60*1000;
                break;
            case "30min":
                time_interval = 30*60*1000;
                break;
            case "1hour":
                time_interval = 60*60*1000;
                break;
            case "3hour":
                time_interval = 3*60*60*1000;
                break;
            case "12hour":
                time_interval = 12*60*60*1000;
                break;
            case "sofar":
                time_from = new Date().setHours(0,0,0,0);
                time_to = new Date();
                time_interval = time_to - time_from;
                break;
        }
        updateCharts();
    });

    $("#toggleTime li a").click(function(){
        var html = '<i class="fa fa-clock-o"></i> '+ $(this).text() + ' <i class="fa fa-angle-down"></i>';
        $(this).parents(".btn-group").find('#time-vis').html(html);
    });

    // home.html
    var timer;
    if ( $("#app_status").length > 0 ) {
        updateAppTable();
        timer = setInterval(function(){
            updateAppTable();
        }, 2000);
    }




    // workspace.html
    $('.workspace_tab').click(function(e){
            var pod_name = e.target.id;
            $.ajax({
                url: "/home/",
                method: 'POST',
                data: {
                    appid: pod_name
                }
            }).then(function(res){
                location.href = "/workspace/";
            });
        });
    if($('#app-tab').children().length <= 1) {
        $('.full-screen-btn').css('display','none');
    }else {
        var full_screen_flag = false;
        $('.full-screen-btn').click(function(){
            full_screen_flag = !full_screen_flag;
            if (full_screen_flag) {
                $('.full-screen-btn').attr('title','Exit Full Screen').addClass('full-screen-active');
                $('#workspace').addClass('full-screen');
            } else {
                $('.full-screen-btn').attr('title','Full Screen').removeClass('full-screen-active');
                $('#workspace').removeClass('full-screen');
            }
        });
    }


    function calAge(objectMeta){
        var cur = new Date();
        var start = new Date(objectMeta.creationTimestamp);
        var diff = cur.getTime() - start.getTime();
        var age;
        if (Math.floor(diff / (24 * 60 * 60 * 1000)) > 1) {
            var days = Math.floor(diff / (24 * 60 * 60 * 1000));
            age = days > 1 ? days + " days" : "a day";
        } else if (Math.floor(diff / (60 * 60 * 1000)) > 1) {
            var hours = Math.floor(diff / (60 * 60 * 1000));
            age = hours > 1 ? hours + " hours" : "an hour";
        } else {
            var minutes = Math.floor(diff / (60 * 1000));
            age = minutes > 1 ? minutes + " minutes" : "a minute";
        }
        return age
    }

    function createAppData(app_data){
        var app_table_data = [];
        var apps = app_data.pods;
        for (var i = 0; i<apps.length; i++) {
            var app = apps[i];
            var objectMeta = app.objectMeta;
            var podStatus = app.podStatus;
            var age = calAge(objectMeta);
            var cpu_usage, mem_usage;
            if (podStatus.status === "success") {
                var cpuUsage = app.metrics.cpuUsage;
                cpuUsage = cpuUsage === null ? 0 : cpuUsage;
                cpu_usage = app.metrics.cpuUsageHistory.map(function (history) {
                    return history.value;
                });

                cpu_usage = JSON.stringify(cpu_usage) + ";" + cpuUsage / 1000;

                mem_usage = app.metrics.memoryUsageHistory.map(function (history) {
                    return history.value;
                });
                var memoryUsage = app.metrics.memoryUsage;
                memoryUsage = memoryUsage === null ? 0 : memoryUsage;
                memoryUsage = memoryUsage > 1024 * 1024 ? (memoryUsage / 1024 / 1024).toFixed(3) + " MB" : memoryUsage > 1024 ? (memoryUsage / 1024 / 1024).toFixed(3) + " KB" : memoryUsage;
                mem_usage = JSON.stringify(mem_usage) + ";" + memoryUsage;
            } else {
                cpu_usage = "-";
                mem_usage = "-";
            }

            var success_content = '<i class="fa fa-check-circle-o fa-lg" title = "Running" aria-hidden="true" style="text-align: center;width: 100%;color: mediumseagreen"></i>';
            var pending_content = '<i class="fa fa-clock-o fa-lg" title="Pending" aria-hidden="true" style="text-align: center;width: 100%;color: #666;"></i>';
            var fail_content = '<i class="fa fa-times-circle-o fa-lg" title="Error" aria-hidden="true" style="text-align: center;width: 100%;color: #ff5722;"></i>';
            var status, reason, msg;
            if (podStatus.podPhase === "Running") {
                status = success_content;
            } else {
                if (!podStatus.containerStates){
                    status = fail_content;
                    if (app.warnings && app.warnings.length > 0) {
                        var fail_message = app.warnings[0].message;
                        fail_content = '<i class="fa fa-times-circle-o fa-lg" title="' + fail_message + '" aria-hidden="true" style="text-align: center;width: 100%;color: #ff5722;"></i>';
                        status = fail_content;
                    }
                } else {
                    reason = podStatus.containerStates[0].waiting.reason;
                    msg = podStatus.containerStates[0].waiting.message;
                    if (!msg) {
                        status = pending_content;
                    } else {
                        fail_content = '<i class="fa fa-times-circle-o fa-lg" title="' + msg + '" aria-hidden="true" style="text-align: center;width: 100%;color: #ff5722;"></i>';
                        status = fail_content;
                    }
                }
            }
            app_table_data.push({
                "status": status,
                "app_name": '<a>' + app.appMeta.name + '</a>',
                "app_name_hidden": app.appMeta.name,
                "podPhase": podStatus.podPhase,
                "age": age,
                "cpu_usage": cpu_usage,
                "mem_usage": mem_usage,
                "cpu": app.appMeta.cpu,
                "mem": app.appMeta.mem + " MB",
                "gpu": app.appMeta.gpu,
                "pod_ip": app.appMeta.podIP,
                "pod_name_hidden": objectMeta.name,
                "gpu_mode": app.appMeta.gpu_mode===1 ? 'Dedicated': 'Shared',
                "manage": '<button class="btn btn-danger btn-sm delete_app" >Delete</button>'
            });
        }
        return app_table_data
    }

    function createALLothersAppData(app_data){
        var app_table_data = [];
        var apps = app_data.pods;
        for (var i = 0; i<apps.length; i++) {
            var app = apps[i];
            var objectMeta = app.objectMeta;
            var podStatus = app.podStatus;
            var age = calAge(objectMeta);

            var cpu_usage, mem_usage;
            if (podStatus.status === "success") {
                var cpuUsage = app.metrics.cpuUsage;
                cpuUsage = cpuUsage === null ? 0 : cpuUsage;
                cpu_usage = app.metrics.cpuUsageHistory.map(function (history) {
                    return history.value;
                });

                cpu_usage = JSON.stringify(cpu_usage) + ";" + cpuUsage / 1000;

                mem_usage = app.metrics.memoryUsageHistory.map(function (history) {
                    return history.value;
                });
                var memoryUsage = app.metrics.memoryUsage;
                memoryUsage = memoryUsage === null ? 0 : memoryUsage;
                memoryUsage = memoryUsage > 1024 * 1024 ? (memoryUsage / 1024 / 1024).toFixed(3) + " MB" : memoryUsage > 1024 ? (memoryUsage / 1024 / 1024).toFixed(3) + " KB" : memoryUsage;
                mem_usage = JSON.stringify(mem_usage) + ";" + memoryUsage;
            } else {
                cpu_usage = "-";
                mem_usage = "-";
            }

            var success_content = '<i class="fa fa-check-circle-o fa-lg" title = "Running" aria-hidden="true" style="text-align: center;width: 100%;color: mediumseagreen"></i>';
            var pending_content = '<i class="fa fa-clock-o fa-lg" title="Pending" aria-hidden="true" style="text-align: center;width: 100%;color: #666;"></i>';
            var fail_content = '<i class="fa fa-times-circle-o fa-lg" title="Error" aria-hidden="true" style="text-align: center;width: 100%;color: #ff5722;"></i>';
            var status, reason, msg;
            if (podStatus.podPhase === "Running") {
                status = success_content;
            } else {
                if (!podStatus.containerStates){
                    status = fail_content;
                } else {
                    reason = podStatus.containerStates[0].waiting.reason;
                    msg = podStatus.containerStates[0].waiting.message;
                    if (!msg) {
                        status = pending_content;
                    } else {
                        status = fail_content;
                    }
                }
            }
            app_table_data.push({
                "status": status,
                "app_name": app.appMeta.name,
                "podPhase": podStatus.podPhase,
                "age": age,
                "cpu_usage": cpu_usage,
                "mem_usage": mem_usage,
                "cpu": app.appMeta.cpu,
                "mem": app.appMeta.mem + " MB",
                "gpu": app.appMeta.gpu,
                "pod_ip": app.appMeta.podIP,
                "username": app.appMeta.username,
                "gpu_mode": app.appMeta.gpu_mode === 1? 'Dedicated': 'Shared'
            });
        }
        return app_table_data
    }
    function updateAppTable(){
        if ($("#app-table").length > 0){
            $.get( "../app_data/api/v1/pod/default", function( app_data ) {
                var app_table_data = createAppData(app_data);
                var app_table = $('#app-table');
                app_table.bootstrapTable('destroy');
                app_table.bootstrapTable({
                    data: app_table_data,
                    onPostBody: function(){
                        var type = "all";
                        drawSparkline(type);
                        //closeNotify();
                    }
                });

                if (app_table_data.length === 0) {
                    $("#no_app_warning").removeClass("hide");
                } else {
                    $("#no_app_warning").addClass("hide");
                }

                $('.delete_app').click(function(e){
                    var app_name = $(e.target.parentElement.parentElement).find('.app_name_hidden')[0].textContent;
                    closeNotify();
                    $('#appdelete-modal').modal('toggle');
                    $('.delete_app2').click(function() {
                        $('#appdelete-modal').modal('hide');
                        $('.delete_app2').unbind();
                        closeNotify();
                        $.notify("Deleting App <strong>" + app_name + "</strong> ...", {
                            element: ".page-content",
                            type: 'success',
                            placement: {
                                from: "top",
                                align: "center"
                            }
                        });
                        setTimeout(function(){$.notifyClose();},2000);
                        $.ajax({
                            url: "/app_manage",
                            method: 'POST',
                            data: {
                                appname: app_name,
                                action: 'Delete'
                            }
                        }).then(function(res){
                            if (res.status) {
                                updateAppTable();
                            }else{
                                $.notify(res.message + " <strong>" + app_name + "</strong>", {
                                    element: ".page-content",
                                    type: 'warning',
                                    placement: {
                                        from: "top",
                                        align: "center"
                                    },
                                    timer: 2000
                                });
                            }
                        });
                    });

                });

                $('.app_name').click(function(e){
                    var pod_name = $(e.target.parentElement.parentElement).find('.pod_name_hidden')[0].textContent;
                    $.ajax({
                        url: "/home/",
                        method: 'POST',
                        data: {
                            appid: pod_name
                        }
                    }).then(function(res){
                        location.href = "/workspace/";
                    });
                });
            });
        }

        if ($("#allothers-app-table").length > 0){
            $.get( "../app_data_allothers/api/v1/pod/default", function( app_data ) {
                var app_table_data = createALLothersAppData(app_data);
                var allothers_app_table =  $('#allothers-app-table');
                allothers_app_table.bootstrapTable('destroy');
                allothers_app_table.bootstrapTable({
                    data: app_table_data,
                    onPostBody: function(){
                        var type = "allothers";
                        drawSparkline(type);
                    }
                });
            });
        }
        }


    $('.create_app').click(function(e){
        $('#new_app_form').validator('validate');
        if ($('#appname').val()) {
            var form = $('#new_app_form').serializeArray().reduce(function (obj, item) {
                obj[item.name] = item.value;
                return obj;
            }, {});
            var app_name = form.appname;
            form.action = 'Create';

            $.notifyClose();

            setTimeout(function(){$.notifyClose();},2000);

            $.ajax({
                url: "/app_manage",
                method: 'POST',
                data: form
            }).then(function (res) {
                $('#appcreate-modal').modal('hide');
                $.notify("Creating App <strong>" + app_name + "</strong> ...", {
                    element: ".page-content",
                    type: 'success',
                    placement: {
                        from: "top",
                        align: "center"
                    }
                });
                $('#new_app_form')[0].reset();

                if (res.status) {
                    updateAppTable();
                } else {
                    $.notifyClose();
                    $.notify(res.message + " App name: <strong>" + app_name + "</strong>", {
                        element: ".page-content",
                        type: 'warning',
                        placement: {
                            from: "top",
                            align: "center"
                        },
                        // timer: 5000
                    });
                }
            });
        }
    });

    $('.cancel_btn').click(function() {
        var app_form = $('#new_app_form');
        app_form.validator('destroy');
        app_form[0].reset();
    });
    $('.cancel_profile').click(function() {
        var user_form = $('#user_profile_form');
        user_form.validator('destroy');
        user_form[0].reset();
    });


    $('.user_profile').click(function(e) {
        if(!$("#user_profile_form").validator().data('bs.validator').hasErrors()) {
            var user_form = $('#user_profile_form').serializeArray().reduce(function(obj, item) {
                obj[item.name] = item.value;
                return obj;
            }, {});
            $.ajax({
                url:"/user_profile",
                method: 'POST',
                data: user_form
            }).then(function(res) {
                if(res.status) {
                    $('#userprofile-modal').modal('hide');
                    $.notify(res.message, {
                        element: ".page-content",
                        type: 'warning',
                        placement: {
                            from: "top",
                            align: "center"
                        }
                    });
                    setTimeout(function(){closeNotify();}, 1000);

                }else {
                    $.notify(res.message, {
                        element:"#user_profile_form",
                        type: 'warning',
                        placement: {
                            from: "top",
                            align: "center"
                        },
                        offset: -60,
                        width: 200
                    });
                    setTimeout(function(){closeNotify();}, 1000);
                }

            });
        }
    });


    function drawSparkline(type){
        $('.cpu-spark.'+ type).each(function(idx, ele){
            if ($(ele).get(0).tagName == "TD"){
                if (ele.textContent === "-") return;
                var usage = ele.textContent.split(";");
                var values = JSON.parse("[" + usage[0] + "]")[0];
                var sum = usage[1];
                $(ele).sparkline(values, {
                    width: 80,
                    lineColor: "#00c752",
                    fillColor: "#00c752",
                    lineWidth: 0,
                    spotColor: false,
                    minSpotColor: false,
                    maxSpotColor: false,
                    chartRangeMin: 0,
                    disableInteraction: true
                });
                $(ele).append(' '+sum);
            }
        });

        $('.mem-spark.' + type).each(function(idx, ele){
            if ($(ele).get(0).tagName == "TD"){
                if (ele.textContent === "-") return;
                var usage = ele.textContent.split(";");
                var values = JSON.parse("[" + usage[0] + "]")[0];
                var sum = usage[1];
                $(ele).sparkline(values, {
                    width: 80,
                    lineColor: "#326de6",
                    fillColor: "#326de6",
                    lineWidth: 0,
                    spotColor: false,
                    minSpotColor: false,
                    maxSpotColor: false,
                    chartRangeMin: 0,
                    disableInteraction: true
                });
                $(ele).append(' '+sum);
            }
        });
    }

    function updateCharts(){
        $("#refresh-vis").find("i").addClass("fa-spin");

        time_to = new Date();
        time_from = time_to - time_interval;
        if (nodeactive === "overall"){
            cpu_src = "../grafana_data/dashboard-solo/db/cluster?from="+time_from+"&to="+time_to+"&var-nodename=node1&theme=light&panelId=3";
            mem_src = "../grafana_data/dashboard-solo/db/cluster?from="+time_from+"&to="+time_to+"&var-nodename=node1&theme=light&panelId=1";
            net_src = "../grafana_data/dashboard-solo/db/cluster?from="+time_from+"&to="+time_to+"&var-nodename=node1&theme=light&panelId=7";
            file_src = "../grafana_data/dashboard-solo/db/cluster?from="+time_from+"&to="+time_to+"&var-nodename=node1&theme=light&panelId=10"
        } else {
            cpu_src = "../grafana_data/dashboard-solo/db/cluster?from="+time_from+"&to="+time_to+"&var-nodename="+ nodeactive+ "&theme=light&panelId=6";
            mem_src = "../grafana_data/dashboard-solo/db/cluster?from="+time_from+"&to="+time_to+"&var-nodename="+ nodeactive+ "&theme=light&panelId=5";
            net_src = "../grafana_data/dashboard-solo/db/cluster?from="+time_from+"&to="+time_to+"&var-nodename="+ nodeactive+ "&theme=light&panelId=9";
            file_src = "../grafana_data/dashboard-solo/db/cluster?from="+time_from+"&to="+time_to+"&var-nodename="+ nodeactive+ "&theme=light&panelId=12"
        }
        $('#cpu').attr('src', cpu_src);
        $('#mem').attr('src', mem_src);
        $('#net').attr('src', net_src);
        $('#file').attr('src', file_src);
    }
})();