
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