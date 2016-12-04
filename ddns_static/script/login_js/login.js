
jQuery(document).ready(function () {

    /*
        Fullscreen background
    */
    $.backstretch([
        (AssetsHost ? AssetsHost : "") + "/assets/img/backgrounds/2.jpg",
        (AssetsHost ? AssetsHost : "") + "/assets/img/backgrounds/3.jpg",
        (AssetsHost ? AssetsHost : "") + "/assets/img/backgrounds/1.jpg"
    ], { duration: 3000, fade: 750 });
    /*
        Form validation
    */
    $('.login-form input[type="text"], .login-form input[type="password"], .login-form textarea').on('focus', function () {
        $(this).removeClass('input-error');
    });

    $('.login-form').on('submit', function (e) {
        e.preventDefault();

        var username = $('input[type="text"]');
        if (username.val() === '') {
            username.addClass('input-error');
            return;
        } else {
            username.removeClass('input-error');
        }

        var password = $('input[type="password"]');
        if (password.val() === '') {
            password.addClass('input-error');
            return;
        } else {
            password.removeClass('input-error');
        }

        var getTo = function () {
            var reg = new RegExp("(\\?|&)to=([^&]*)(&|$)");
            var r = window.location.search.match(reg);
            if (r !== null)
                return unescape(r[2]);
            return null;
        };

        $.ajax({
            type: 'POST',
            url: '/login',
            data: { "username": username.val(), "password": password.val() },
            contentType: 'application/x-www-form-urlencoded; charset=UTF-8',
            success: function (respData) {
                if (respData.code === "ok") {
                    var to = getTo();
                    if (to === null)
                        to = '/';
                    window.location.href = to;
                } else if (respData.code === "UserNameError") {
                    username.addClass('input-error');
                } else if (respData.code === "PasswordError") {
                    password.addClass('input-error');
                } else {
                    var msg = respData.code;
                    if (respData.msg)
                        msg += (" " + respData.msg);
                    alert(msg);
                }
            }
        });
    });
});
