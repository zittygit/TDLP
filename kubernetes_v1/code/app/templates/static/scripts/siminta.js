
/*====================================
 Free To Use For Personal And Commercial Usage
Author: http://binarytheme.com
 Share Us if You Like our work 
 Enjoy Our Codes For Free always.
======================================*/

$(function () {
    // tooltip demo
    $('.tooltip-demo').tooltip({
        selector: "[data-toggle=tooltip]",
        container: "body"
    })

    // popover demo
    $("[data-toggle=popover]")
        .popover()
    ///calling side menu

    $('#side-menu').metisMenu();


    function autoResizeDiv()
    {
        // var body = document.body,
        //     html = document.documentElement;
        //
        // var height = Math.max( body.scrollHeight, body.offsetHeight,
        //                        html.clientHeight, html.scrollHeight, html.offsetHeight );
        // if ($("#page-wrapper").length){
        //     $("#page-wrapper").height((height - 70) +'px');
        // }
        if ($("#workspace").length){
            $("#workspace").height((window.innerHeight - 70 - 115) +'px');
        }
        if ($("#bar-control").length){
            $("#bar-control").css('margin-top', (window.innerHeight/2) + 'px');
        }
        //document.getElementById('').style.height = (window.innerHeight - 70) +'px';
        //document.getElementById('workspace').style.height = (window.innerHeight - 70 - 115) +'px';
    }
    window.onresize = autoResizeDiv;
    autoResizeDiv();

    Pace.on('hide', function () {
        console.log('done');
    });
    paceOptions = {
        elements: true,
        restartOnRequestAfter: false
    };
   

});

//Loads the correct sidebar on window load, collapses the sidebar on window resize.
$(function() {
    $(window).bind("load resize", function() {
        console.log($(this).width())
        if ($(this).width() < 768) {
            $('div.sidebar-collapse').addClass('collapse')
        } else {
            $('div.sidebar-collapse').removeClass('collapse')
        }
    })
})

$(function() {
    if ( $( "#chart_div" ).length ) {
        // Load the Visualization API and the piechart package.
        google.charts.load('current', {'packages': ['corechart']});

        // Set a callback to run when the Google Visualization API is loaded.
        google.charts.setOnLoadCallback(drawChart);

        // Callback that creates and populates a data table,
        // instantiates the pie chart, passes in the data and
        // draws it.
        function drawChart() {

            // Create the data table.
            var data = new google.visualization.DataTable();
            data.addColumn('string', 'Topping');
            data.addColumn('number', 'Slices');
            data.addRows([
                ['Mushrooms', 3],
                ['Onions', 1],
                ['Olives', 1],
                ['Zucchini', 1],
                ['Pepperoni', 2]
            ]);

            // Set chart options
            var options = {
                'title': 'How Much Pizza I Ate Last Night',
                'width': 400,
                'height': 300
            };

            // Instantiate and draw our chart, passing in some options.
            var chart = new google.visualization.PieChart(document.getElementById('chart_div'));

            function selectHandler() {
                var selectedItem = chart.getSelection()[0];
                if (selectedItem) {
                    var topping = data.getValue(selectedItem.row, 0);
                    alert('The user selected ' + topping);
                }
            }

            google.visualization.events.addListener(chart, 'select', selectHandler);
            chart.draw(data, options);
        }
    }
})



$(".sidebar-control").click(function(){
    if ($('div.sidebar-collapse').hasClass('collapse')){
        $('#bar-control').attr('src', '/static/img/DoubleChevronLeft.svg');
        $('div.sidebar-collapse').removeClass('collapse');
        $('.sidebar-control').css('left', '235px');
        $('.navbar-static-side').css('width', '235px');
        $('#page-wrapper').css('margin-left', '250px');
    } else {
        $('#bar-control').attr('src', '/static/img/DoubleChevronRight.svg');
        $('div.sidebar-collapse').addClass('collapse');
        $('.sidebar-control').css('left', '0px');
        $('.navbar-static-side').css('width', '0px');
        $('#page-wrapper').css('margin-left', '15px');
    }
});

function popup(mylink, windowname) {
    if (! window.focus)return true;
    var href;
    if (typeof(mylink) == 'string') href=mylink;
    else href=mylink.href;
    window.open(href, windowname, 'location=no,type=fullWindow,fullscreen,scrollbars=yes');
    return false;
}