
(function (exports) {

	(function () {
		$('.dropdown-menu .logout').click(function (e) {
			e.preventDefault();
			$.ajax({
				type: 'POST',
				url: '/logout',
				cache: false,
				xhrFields: {
					withCredentials: true
				},
				success: function (rsp) {
					if (rsp.code === "ok" || rsp.code === "InvalidSession") {
						var href = window.location.href;
						if (href !== '/login') {
							href = '/login?to=' + encodeURIComponent(href);
						}
						window.location.href = href;
					} else {
						alert("logout failed," + rsp.code);
					}
				}
			});
		});
	})();

	function getTableHeight() {
		return $(window).height() - $('h1').outerHeight(true);
	}
	function totalTextFormatter(data) {
		return 'Total';
	}
	function totalNameFormatter(data) {
		return data.length;
	}

	function on_init_recode_lists(did) {
		$table = $('#tb_recodes');
		var $remove = $('#remove'),
			$add = $('#add'),
			selections = [],
			operateEvents = {
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
						ddns_new_recode(
							did,
							{ host: row.host, type: str_to_type(row.type), value: row.value, ttl: row.ttl },
							function (rspData) {
								if (rspData.code == "ok") {
									$table.bootstrapTable('updateRow', { index: idx, row: { id: rspData.id, key: rspData.key, dynamic: rspData.dynamic } });
								} else {
									var msg = rspData.code;
									if (rspData.msg && rspData.msg.length)
										msg += ("   " + rspData.msg);
									alert(msg);
								}
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
							}
						);
					}
				}
			};

		function formatType(value, row, index) {
			return type_to_str(value);
		}
		function formatDynamic(value, row, index) {
			return '<input class="dynamic" type="checkbox"' + (value === true ? 'checked="checked"' : '') + '"/>';
		}

		function initTable() {
			$table.bootstrapTable({
				height: getTableHeight(),
				columns: [
					[
						{
							title: 'ID',
							field: 'id',
							align: 'center',
							valign: 'middle',
							visible: false,
							//sortable: true,
							footerFormatter: totalTextFormatter
						}, {
							field: 'host',
							title: 'Host Recode',
							//sortable: true,
							editable: true,
							footerFormatter: totalNameFormatter,
							align: 'center'
						}, {
							field: 'value',
							title: 'Recode Value',
							editable: true,
							footerFormatter: totalNameFormatter,
							align: 'center'
						}, {
							field: 'type',
							title: 'Recode Type',
							editable: true,
							footerFormatter: totalNameFormatter,
							align: 'center',
							formatter: formatType
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
							//editable: true,
							footerFormatter: totalNameFormatter,
							align: 'center',
							formatter: formatDynamic,
							events: {
								'click .dynamic': function (e, value, row, index) {
									row.dynamic = !value;
								}
							}
						}, {
							field: 'key',
							title: 'Update Key',
							align: 'center',
							visible: false,
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
				responseHandler: responseHandler,
				url: "/domain/" + did + "/recodes"
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
					height: getTableHeight()
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
		function operateFormatter(value, row, index) {
			return [
				'<a class="save" href="javascript:void(0)" title="Save">',
				'<i class="	glyphicon glyphicon-saved"></i>',
				'</a>',
				'<a class="remove" href="javascript:void(0)" title="Remove">',
				'<i class="glyphicon glyphicon-remove"></i>',
				'</a>'
			].join('');
		}
		initTable();
	}

	function on_init_domians_view() {
		$table = $('#tb_domains');
		var $remove = $('#remove'),
			$add = $('#add'),
			selections = [],
			operateEvents = {
				'click .save': function (e, value, row, idx) {
					if (row.id && row.id > 0) {
						ddns_update_domain(row.id, row.domain,
							function (rspData) {
								if (rspData.code == "ok") {
									$table.bootstrapTable('updateRow', { index: idx, row: { id: rspData.id } });
								} else {
									var msg = rspData.code;
									if (rspData.msg && rspData.msg.length)
										msg += ("   " + rspData.msg);
									alert(msg);
								}
							});
					} else {
						ddns_new_domain(row.domain,
							function (rspData) {
								if (rspData.code == "ok") {
									$table.bootstrapTable('updateRow', { index: idx, row: { id: rspData.id } });
								} else {
									var msg = rspData.code;
									if (rspData.msg && rspData.msg.length)
										msg += ("   " + rspData.msg);
									alert(msg);
								}
							},
							function (a, b, c) { }
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
						ddns_delete_domain(row.id,
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
							}
						);
					}
				}
			};

		function operateFormatter() {
			return [
				'<a class="save" href="javascript:void(0)" title="Like">',
				'<i class="	glyphicon glyphicon-saved"></i>',
				'</a>',
				'<a class="remove" href="javascript:void(0)" title="Remove">',
				'<i class="glyphicon glyphicon-remove"></i>',
				'</a>'
			].join('');
		}
		function responseHandler(res) {
			if (res) {
				if (res.code == "ok") {
					return {
						"total": res.domains.length,
						"rows": res.domains
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
		function formatId(value, row, index) {
			return '<a href="/domain/' + value + '/recodes">' + value + '</a>';
		}

		(function () {
			$table.bootstrapTable({
				height: getTableHeight(),
				columns: [
					[
						{
							title: 'ID',
							field: 'id',
							align: 'center',
							valign: 'middle',
							formatter: formatId,
							//footerFormatter: totalTextFormatter
						}, {
							field: 'domain',
							title: 'Domain',
							editable: true,
							footerFormatter: totalNameFormatter,
							align: 'center'
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
		})();
	}

	exports.on_init_recode_lists = on_init_recode_lists;
	exports.on_init_domians_view = on_init_domians_view;

})((typeof (exports) === "object" ? exports : window), jQuery);