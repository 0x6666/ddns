
(function (exports) {

	var basic_func = function (type, api, data, suCallBack, errCallBack, contentType, async) {
		$.ajax({
			type: type,
			url: api,
			data: data,
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

	exports.ddns_get_recodes = get_recodes;
	exports.ddns_new_domain = new_domain;
	exports.ddns_delete_domain = delete_domain;
	exports.ddns_update_domain = update_domain;

	exports.ddns_new_recode = new_recode;
	exports.ddns_delete_recode = delete_recode;
	exports.ddns_update_recode = update_recode;
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

	var str_to_type = function(str) {
		switch(str)
		{
			case 'A': return 1;
			case 'AAAA': return 2;
			case 'CNAME': return 3;
			default: return -1;
		}
	};

	var type_to_str = function(type) {
		switch(type)
		{
			case 1: return 'A';
			case 2: return 'AAAA';
			case 3: return 'CNAME';
			default: return "Unknow";
		}
	};

	exports.load_series = load_series;
	exports.str_to_type = str_to_type;
	exports.type_to_str = type_to_str;

})((typeof (exports) === "object" ? exports : window), jQuery);