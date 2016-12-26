
function on_init_recode_lists() {

	var $table = $('#tb_recodes'),
		$remove = $('#remove'),
		$add = $('#add'),
		selections = [];

	function initTable() {
		$table.bootstrapTable({
			height: getHeight(),
			columns: [
				[
					{
						title: 'ID',
						field: 'id',
						align: 'center',
						valign: 'middle',
						//sortable: true,
						footerFormatter: totalTextFormatter
					}, {
						field: 'name',
						title: 'Domain',
						//sortable: true,
						editable: true,
						footerFormatter: totalNameFormatter,
						align: 'center'
					}, {
						field: 'value',
						title: 'Value',
						//sortable: true,
						editable: true,
						footerFormatter: totalNameFormatter,
						align: 'center'
					}, {
						field: 'ttl',
						title: 'TTL',
						//sortable: true,
						editable: true,
						footerFormatter: totalNameFormatter,
						align: 'center'
					}, {
						field: 'dynamic',
						title: 'Dynamic',
						//sortable: true,
						editable: true,
						footerFormatter: totalNameFormatter,
						align: 'center'
					}, {
						field: 'key',
						title: 'Update Key',
						//sortable: true,
						align: 'center',
						footerFormatter: totalPriceFormatter
					}, {
						field: 'operate',
						title: 'Operate',
						align: 'center',
						events: operateEvents,
						formatter: operateFormatter
					}
				]
			],
			queryParams: function (params) {
				return params;
			},
			responseHandler: responseHandler
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
		$add.click(function () {
			var randomId = Number(new Date());
			$table.bootstrapTable('insertRow', {
				index: 0,
				row: {
					id: '',
					name: 'test.site',
					key: '',
					value: '',
					ttl: 600,
					dynamic: true,
					randomId: randomId
				}
			});
		});
		$(window).resize(function () {
			$table.bootstrapTable('resetView', {
				height: getHeight()
			});
		});
	}
	function getIdSelections() {
		return $.map($table.bootstrapTable('getSelections'), function (row) {
			return row.id;
		});
	}
	function responseHandler(res) {
		if (res) {
			if (res.code == "ok") {
				return {
					"total": res.recodes.length,
					"rows": res.recodes
				};
			} else {
				var msg = res.code;
				if (res.msg && res.code.length)
					msg += ("   " + res.msg);
				alert(msg);
			}
		}
		return {
			"total": 0,
			"rows": []
		};
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
			'<a class="save" href="javascript:void(0)" title="Like">',
			'<i class="	glyphicon glyphicon-saved"></i>',
			'</a>',
			'<a class="remove" href="javascript:void(0)" title="Remove">',
			'<i class="glyphicon glyphicon-remove"></i>',
			'</a>'
		].join('');
	}
	window.operateEvents = {
		'click .save': function (e, value, row, idx) {
			if (row.id && row.id > 0) {
				ddns_update_recode(row.id, { name: row.name, value: row.value, ttl: row.ttl, dynamic: row.dynamic },
					function (rspData) {
						if (rspData.code === "ok") {
							alert("ok");
						} else {
							var msg = rspData.code;
							if (rspData.msg && rspData.msg.length)
								msg += ("   " + rspData.msg);
							alert(msg);
						}
					},
					function (a, b, c) { });
			} else {
				ddns_new_recode({ name: row.name, value: row.value, ttl: row.ttl },
					function (rspData) {
						if (rspData.code == "ok") {
							$table.bootstrapTable('updateRow', { index: idx, row: { id: rspData.id, key: rspData.key, dynamic: rspData.dynamic } });
						} else {
							var msg = rspData.code;
							if (rspData.msg && rspData.msg.length)
								msg += ("   " + rspData.msg);
							alert(msg);
						}
					},
					function (a, b, c) {

					}
				);
			}
		},
		'click .remove': function (e, value, row, index) {

			if (row.id === undefined || row.id === '') {
				if (row.randomId) {
					$table.bootstrapTable('remove', {
						field: 'randomId',
						values: [row.randomId]
					});
				}
			} else {
				ddns_delete_recode(row.id,
					function (data) {
						if (data.code === "ok") {
							$table.bootstrapTable('remove', {
								field: 'id',
								values: [row.id]
							});
						} else {
							var msg = data.code;
							if (data.msg && data.msg.length)
								msg += ("   " + data.msg);
							alert(msg);
						}
					},
					function (a, b, c) { }
				);
			}
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
	initTable();
	/*$(function () {
		var scripts = [
			location.search.substring(1) ||
			'/vendors/bootstrap-table-1.11.0/extensions/export/bootstrap-table-export.js',
			'/vendors/tableExport.jquery.plugin/tableExport.js',
			'/vendors/bootstrap-table-1.11.0/extensions/editable/bootstrap-table-editable.js',
			'/vendors/x-editable/bootstrap-editable.js'
		];

		load_series(scripts, initTable);
	});*/
}

function on_init_domians_view() {

}