
$(function (exports) {

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
