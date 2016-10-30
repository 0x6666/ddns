

function get_script(url, callback) {
	var head = document.getElementsByTagName('head')[0];
	var script = document.createElement('script');
	script.src = url;
	var done = false;
	// Attach handlers for all browsers
	script.onload = script.onreadystatechange = function () {
		if (!done && (!this.readyState ||
			this.readyState == 'loaded' || this.readyState == 'complete')) {
			done = true;
			if (callback)
				callback();
			// Handle memory leak in IE
			script.onload = script.onreadystatechange = null;
		}
	};
	head.appendChild(script);
	// We handle everything using the script element injection
	return undefined;
}

function load_series(arr, callback) {
	callback = callback || function () { };
	if (!arr.length) return callback();

	var completed = 0;
	var iterate = function () {
		get_script(arr[completed], function (err) {
			if (err) {
				callback(err);
				callback = function () { };
			}
			else {
				completed += 1;
				if (completed >= arr.length) {
					callback(null);
				}
				else {
					iterate();
				}
			}
		});
	};
	iterate();
};

function on_init_recode_lists() {

	var $table = $('#tb_recodes'),
		$remove = $('#remove'),
		selections = [];

	function initTable() {
		$table.bootstrapTable({
			height: getHeight(),
			columns: [
				[
					{
						title: 'Item ID',
						field: 'id',
						align: 'center',
						valign: 'middle',
						//sortable: true,
						footerFormatter: totalTextFormatter
					}, {
						field: 'name',
						title: 'Item Name',
						//sortable: true,
						editable: true,
						footerFormatter: totalNameFormatter,
						align: 'center'
					}, {
						field: 'price',
						title: 'Item Price',
						//sortable: true,
						align: 'center',
						editable: {
							type: 'text',
							title: 'Item Price',
							validate: function (value) {
								value = $.trim(value);
								if (!value) {
									return 'This field is required';
								}
								if (!/^\$/.test(value)) {
									return 'This field needs to start width $.'
								}
								var data = $table.bootstrapTable('getData'),
									index = $(this).parents('tr').data('index');
								console.log(data[index]);
								return '';
							}
						},
						footerFormatter: totalPriceFormatter
					}, {
						field: 'operate',
						title: 'Item Operate',
						align: 'center',
						events: operateEvents,
						formatter: operateFormatter
					}
				]
			],
			queryParams: function (params) {
				return params;
			}
		});
		// sometimes footer render error.
		//setTimeout(function () {
		//	$table.bootstrapTable('resetView');
		//}, 200);

		$table.on('check.bs.table uncheck.bs.table ' +
			'check-all.bs.table uncheck-all.bs.table', function () {
				$remove.prop('disabled', !$table.bootstrapTable('getSelections').length);
				// save your data, here just save the current page
				selections = getIdSelections();
				// push or splice the selections if you want to save all data selections
			});
		$table.on('expand-row.bs.table', function (e, index, row, $detail) {
			if (index % 2 == 1) {
				$detail.html('Loading from ajax request...');
				$.get('LICENSE', function (res) {
					$detail.html(res.replace(/\n/g, '<br>'));
				});
			}
		});
		$table.on('all.bs.table', function (e, name, args) {
			console.log(name, args);
		});
		$remove.click(function () {
			var ids = getIdSelections();
			$table.bootstrapTable('remove', {
				field: 'id',
				values: ids
			});
			$remove.prop('disabled', true);
		});
		$(window).resize(function () {
			$table.bootstrapTable('resetView', {
				height: getHeight()
			});
		});
	}
	function getIdSelections() {
		return $.map($table.bootstrapTable('getSelections'), function (row) {
			return row.id
		});
	}
	function responseHandler(res) {
		$.each(res.rows, function (i, row) {
			row.state = $.inArray(row.id, selections) !== -1;
		});
		return res;
	}
	function detailFormatter(index, row) {
		var html = [];
		$.each(row, function (key, value) {
			html.push('<p><b>' + key + ':</b> ' + value + '</p>');
		});
		return html.join('');
	}
	function operateFormatter(value, row, index) {
		return [
			'<a class="like" href="javascript:void(0)" title="Like">',
			'<i class="glyphicon glyphicon-heart"></i>',
			'</a>  ',
			'<a class="remove" href="javascript:void(0)" title="Remove">',
			'<i class="glyphicon glyphicon-remove"></i>',
			'</a>'
		].join('');
	}
	window.operateEvents = {
		'click .like': function (e, value, row, index) {
			alert('You click like action, row: ' + JSON.stringify(row));
		},
		'click .remove': function (e, value, row, index) {
			$table.bootstrapTable('remove', {
				field: 'id',
				values: [row.id]
			});
		}
	};
	function totalTextFormatter(data) {
		return 'Total';
	}
	function totalNameFormatter(data) {
		return data.length;
	}
	function totalPriceFormatter(data) {
		var total = 0;
		$.each(data, function (i, row) {
			total += +(row.price.substring(1));
		});
		return '$' + total;
	}
	function getHeight() {
		return $(window).height() - $('h1').outerHeight(true);
	}
	$(function () {
		var scripts = [
			location.search.substring(1) ||
			'/vendors/bootstrap-table-1.11.0/extensions/export/bootstrap-table-export.js',
			'http://rawgit.com/hhurz/tableExport.jquery.plugin/master/tableExport.js',
			'/vendors/bootstrap-table-1.11.0/extensions/editable/bootstrap-table-editable.js',
			'http://rawgit.com/vitalets/x-editable/master/dist/bootstrap3-editable/js/bootstrap-editable.js'
		];

		load_series(scripts, initTable);
	});
}