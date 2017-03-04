
(function (exports) {

	var basic_func = function (type, api, data, suCallBack, errCallBack, contentType, async) {
		$.ajax({
			type: type,
			url: api,
			data: data,
			dataType: "json",
			cache: false,
			async: (async === undefined || async) ? true : false,
			contentType: (!contentType || contentType.length === 0) ? 'application/x-www-form-urlencoded; charset=UTF-8' : contentType,
			xhrFields: {
				withCredentials: true
			},
			success: suCallBack,
			error: errCallBack
		});
	};

	var get_recodes = function (did, suCallback, failCallback) {
		basic_func("GET", "/domain/"+did+"/recodes", undefined, suCallback, failCallback);
	};

	var new_recode = function (did, data, suCallback, failCallback) {
		basic_func("POST", "/domain/"+did+"/recodes", data, suCallback, failCallback);
	};

	var delete_recode = function (id, suCallback, failCallback) {
		basic_func("DELETE", "/recode/" + id, undefined, suCallback, failCallback);
	};

	var update_recode = function (id, data, suCallback, failCallback) {
		basic_func("PATCH", "/recode/" + id, data, suCallback, failCallback);
	};

	var new_domain = function (domain, suCallback, failCallback) {
		basic_func("POST", "/domains", { "domain": domain }, suCallback, failCallback);
	};

	var delete_domain = function (did, suCallback, failCallback) {
		basic_func("DELETE", "/domain/" + did, undefined, suCallback, failCallback);
	};

	var update_domain = function (did, name, suCallback, failCallback) {
		basic_func("PATCH", "/domain/" + did, { "domain": name }, suCallback, failCallback);
	};

	var download_url = function(url, suCallback, failCallback){
		basic_func("POST", "/downloads", { "url": url }, suCallback, failCallback);
	};

	var get_downloads = function(suCallback, failCallback){
		basic_func("GET", "/downloads", undefined, suCallback, failCallback);
	};

	exports.ddns_get_recodes = get_recodes;
	exports.ddns_new_domain = new_domain;
	exports.ddns_delete_domain = delete_domain;
	exports.ddns_update_domain = update_domain;

	exports.ddns_new_recode = new_recode;
	exports.ddns_delete_recode = delete_recode;
	exports.ddns_update_recode = update_recode;


	//download api
	exports.d_download_url = download_url;
	exports.d_get_downloads = get_downloads;

})((typeof (exports) === "object" ? exports : window), jQuery);

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

	function rowStyle(row, index) {
		if (row.id === '') {
			return { classes: 'new-row' };
		}
		return '';
	}

	function on_init_recode_lists(did) {
		var $table = $('#tb_recodes'),
			$remove = $('#remove'),
			$add = $('#add'),
			selections = [],
			operateEvents = {
				'click .save': function (e, value, row, idx) {
					if (row.id && row.id > 0) {
						ddns_update_recode(row.id, { host: row.host, type: row.type, value: row.value, ttl: row.ttl },
							function (rspData) {
								if (rspData.code === "ok") {
									alert("ok");
								} else {
									var msg = rspData.code;
									if (rspData.msg && rspData.msg.length)
										msg += ("   " + rspData.msg);
									alert(msg);
								}
							});
					} else {
						ddns_new_recode(
							did,
							{ host: row.host, type: row.type, value: row.value, ttl: row.ttl },
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
						}, {
							field: 'host',
							title: 'Host Recode',
							//sortable: true,
							editable: true,
							align: 'center'
						}, {
							field: 'value',
							title: 'Recode Value',
							editable: true,
							align: 'center'
						}, {
							field: 'type',
							title: 'Recode Type',
							align: 'center',
							editable: {
								type: 'select',
								value: 1,
								source: [
									{ value: 1, text: 'A' },
									{ value: 2, text: 'AAAA' },
									{ value: 3, text: 'CNAME' }
								]
							},
						}, {
							field: 'ttl',
							title: 'TTL',
							//sortable: true,
							editable: true,
							align: 'center'
						}, {
							field: 'dynamic',
							title: 'Dynamic',
							//sortable: true,
							//editable: true,
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
				responseHandler: responseHandler,
				url: "/domain/" + did + "/recodes",
				rowStyle: rowStyle
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
						host: '@',
						key: '',
						value: '',
						type: 1,
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
		var $table = $('#tb_domains'),
			$remove = $('#remove'),
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
							//visible: false,
						}, {
							field: 'domain',
							title: 'Domain',
							editable: true,
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
				responseHandler: responseHandler,
				rowStyle: rowStyle
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
		})();
	}

	function on_init_downloads_view() {
		var $table = $('#dtable'),
			$dbtn = $('#dbtn'),
			$DInput = $("#d-input");

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
							visible: false,
						}, {
							field: 'name',
							title: 'Name',
							align: 'left',
							formatter: nameFormatter
						}, {
							field: 'size',
							title: 'Size',
							align: 'right',
							formatter: sizeFormatter,
						}, {
							field: 'progress',
							title: 'Progress',
							formatter: progressFormatter,
							align: 'center'
						}, {
							field: 'bytesPerSecond',
							title: 'Speed',
							formatter: speedFormatter,
							align: 'center'
						}, {
							field: 'operate',
							title: 'Operate',
							align: 'center',
							//events: operateEvents,
							formatter: operateFormatter
						}
					]
				],
				responseHandler: responseHandler,
				rowStyle: rowStyle
			});
			$dbtn.click(function () {
				if ($DInput.val().length === 0) {
					alert("please enter url...");
					return;
				}

				var url = $DInput.val();
				d_download_url(url, function (data) {
					if (data.code === "ok") {
						$table.bootstrapTable('refresh');
						$DInput.val("");
					} else {
						var msg = data.code;
						if (data.msg && data.msg.length)
							msg += ("   " + data.msg);
						alert(msg);
					}
				});
			});
			$(window).resize(function () {
				$table.bootstrapTable('resetView', {
					height: getTableHeight()
				});
			});
			function responseHandler(res) {
				if (res) {
					if (res.code == "ok") {
						return {
							"total": res.tasks.length,
							"rows": res.tasks
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
					'<a class="download" href="' + row.dest + '" target=_blank title="Like">',
					'<i class="glyphicon glyphicon-download-alt"></i>',
					'</a>',
					'<a class="remove" href="javascript:void(0)" title="Remove">',
					'<i class="glyphicon glyphicon-remove"></i>',
					'</a>'
				].join('');
			}
			function nameFormatter(value, row, index) {
				if (row.src && row.src.length)
					return '<span title="' + row.src + '">' + value + '</span>';
				return value;
			}
			function progressFormatter(value, row, index) {
				if (row.size && row.size > 0) {
					if (row.transferred !== undefined) {
						return precision(row.transferred / (row.size * 1.0) * 100) + "%";
					} else {
						return "0%";
					}
				}
			}
			function speedFormatter(value, row, index) {
				if (value < 1024) {
					return precision(value) + "Byte/s";
				} else if (value < 1024 * 1024) {
					return precision(value / 1024.0) + "Kb/s";
				} else if (value < 1024 * 1024 * 1024) {
					return precision(value / (1024.0 * 1024)) + "Mb/s";
				} else {
					return precision(value / (1024.0 * 1024 * 1024)) + "Gb/s";
				}
			}
			function sizeFormatter(value, row, index) {
				if (value < 1024) {
					return precision(value) + "Byte";
				} else if (value < 1024 * 1024) {
					return precision(value / 1024.0) + "K";
				} else if (value < 1024 * 1024 * 1024) {
					return precision(value / (1024.0 * 1024)) + "M";
				} else {
					return precision(value / (1024.0 * 1024 * 1024)) + "G";
				}
			}
		})();

		var refreshTable = function () {
			if (!document.getElementById('dtable')) return;
			d_get_downloads(function (data) {
				var tableData = $table.bootstrapTable('getData');
				var updateTable = function (task) {
					var i = 0;
					for (; i < tableData.length; ++i) {
						if (tableData[i].dest === task.dest) {
							if (tableData[i].transferred !== task.transferred || tableData[i].bytesPerSecond !== task.bytesPerSecond) {
								tableData[i].transferred = task.transferred;
								tableData[i].bytesPerSecond = task.bytesPerSecond;
								$table.bootstrapTable('updateRow', i, tableData[i]);
							}
							break;
						}
					}
					if (i === tableData.length) {
						$table.bootstrapTable('refresh');
						return false;
					}
				};
				if (data.code === "ok" && data.tasks && data.tasks.length) {
					$.each(data.tasks, function (index, task, array) {
						if (updateTable(task) === false) {
							return false;
						}
					});
				}
			});
			setTimeout(refreshTable, 500);
		};
		refreshTable();
	}

	exports.on_init_recode_lists = on_init_recode_lists;
	exports.on_init_domians_view = on_init_domians_view;
	exports.on_init_downloads_view = on_init_downloads_view;

})((typeof (exports) === "object" ? exports : window), jQuery);


(function (exports) {

	var meny = Meny.create({
		menuElement: document.querySelector('.meny'),
		contentsElement: document.querySelector('.contents'),
		position: Meny.getQuery().p || 'left',
		height: 200,
		width: 260,
		threshold: 30,
		mouse: true,
		touch: true,
		overlap: 0
	});

	exports.meny = meny;

	// Embed an iframe if a URL is passed in
	/*if (Meny.getQuery().u && Meny.getQuery().u.match(/^http/gi)) {
		var contents = document.querySelector('.contents');
		contents.style.padding = '0px';
		contents.innerHTML = '<div class="cover"></div><iframe src="' + Meny.getQuery().u + '" style="width: 100%; height: 100%; border: 0; position: absolute;"></iframe>';
	}*/

	$(".menu-item").click(function (e) {
		//e.preventDefault();
		var href = $(this).attr('href');
		if (href.match(/^http/gi)) {
			return true;
		}

		$.ajax({
			type: "GET",
			url: href,
			cache: false,
			dataType: "html",
			xhrFields: {
				withCredentials: true
			},
			headers: { 'DDNS-View': "true" },
			success: function (data, b, c) {
				if (c.status == 401) {
					window.location.href = "/login";
					return;
				} else if (c.status == 200) {
					$("#view-container").html(data);
					var state = { title: '', url: href };
					history.pushState(state, '', href);
					//meny.close();
					return;
				}
				window.location.href = href;
			},
			error: function () {
				window.location.href = href;
			}
		});
		e.preventDefault();
	});
})(typeof (exports) === "object" ? exports : window);


(function (exports) {

	var get_script = function (url, callback) {
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
	};

	var load_series = function (arr, callback) {
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

	var precision = function (f){
		return f.toFixed(2) * 100 / 100.0;
	};

	exports.load_series = load_series;
	exports.precision = precision;

})((typeof (exports) === "object" ? exports : window), jQuery);