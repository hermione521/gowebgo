// load compare image
$("#compare").attr("src","./image/" + originData[originData.length - 1].ClusterId + ".png");
// generate url
$("#more").attr("href", wptUrl + "/testlog.php?days=30&filter=&all=on&label=Script");

var browserList = ["Chrome", "Firefox", "IE"];
var viewList = ["FirstView", "RepeatView"];
var labelList = ["Backend", "StartRender", "DomDelta", "FullyLoaded", "SpeedIndex"];
var titleList = ["BACKEND", "START RENDER", "DOM", "FULLY LOADED", "SPEEDINDEX"];
var yAxis = ["Time (s)", "Time (s)", "Time (s)", "Time (s)", "speedindex (index)"];
var valueSuffix = ["s", "s", "s", "s", ""];

labelList.forEach(function (label) {
	$("#" + label + "Tag").click(function() {
		$(".selected").removeClass("selected");
		$("#" + label + "Tag").addClass("selected");
		$(".highchart").hide();
		$("#" + label).show();
	});
});

labelList.forEach(function (label, index) {
	var series = [];
	var xAxis = [];
	viewList.forEach(function (view) {
		browserList.forEach(function (browser) {
			var item = {
				name: view + "-" + browser,
				data: []
			};
			originData.forEach(function (curData) {
				xAxis.push(curData.ClusterId);
				item.data.push({
					y: curData[browser + "Data"][view][label],
					url: "./" + curData[browser + "Data"].Id + ".html"
				});
			});
			series.push(item);
		});
	});
	
	var options = {
		chart: {
			type: 'line',
			spacingTop: 140,
			spacingBottom: 250
		},
		title: {
			text: titleList[index],
			x: -20, //center
			y: 20
		},
		xAxis: {
			categories: xAxis
		},
		yAxis: {
			title: {
				text: yAxis[index]
			}
		},
		tooltip: {
			valueSuffix: valueSuffix[index]
		},
		plotOptions: {
			series: {
				cursor: 'pointer',
				point: {
					events: {
						click: function () {
							location.href = this.options.url;
						}
					}
				}
			}
		},
		legend: {
			layout: 'vertical',
			align: 'right',
			verticalAlign: 'middle',
			borderWidth: 0
		},
		series: series
	};

	$("#" + label).highcharts(options);
});