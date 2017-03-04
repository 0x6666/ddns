
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

	var download_url = function(url, suCallback, failCallback){
		basic_func("POST", "/downloads", { "url": url }, suCallback, failCallback);
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

})((typeof (exports) === "object" ? exports : window), jQuery);