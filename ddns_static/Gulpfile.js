
'use strict';

var gulp = require('gulp'),
    jshint = require('gulp-jshint'),
    concat = require('gulp-concat'),
    uglify = require('gulp-uglify'),
    rename = require('gulp-rename'),
    sass = require('gulp-sass'),
    rev = require('gulp-rev'),
    revCollector = require('gulp-rev-collector'),
    del = require('del'),
    runSequence = require('run-sequence');
//	revDel = require('rev-del');

var jsSrc = './script/js/**/*.js',
    cssSrc = './script/css/**/*.scss',
    htmlSrc = './script/tmpl/**/*.html',
    destJsDir = './assets/js',
    destCssDir = './assets/css',
    destHtml = './tmpl',
    manifestDir = './rev',
    manifestFile = './rev/**/*.json',
    vendorJsSrc = ['./assets/vendors/bootstrap-3.3.0/js/bootstrap.min.js',
        './assets/vendors/bootstrap-table-1.11.0/bootstrap-table.min.js',
        './assets/vendors/meny/meny.js',
        './assets/vendors/bootstrap-table-1.11.0/extensions/export/bootstrap-table-export.js',
        './assets/vendors/tableExport.jquery.plugin/tableExport.js',
        './assets/vendors/bootstrap-table-1.11.0/extensions/editable/bootstrap-table-editable.js',
        './assets/vendors/x-editable/bootstrap-editable.js'],
    vendorCssSrc = ['./assets/vendors/bootstrap-3.3.0/css/bootstrap.min.css',
        './assets/vendors/bootstrap-3.3.0/css/bootstrap-theme.min.css',
        './assets/vendors/bootstrap-table-1.11.0/bootstrap-table.css',
        './assets/vendors/bootstrap3-editable/css/bootstrap-editable.css'],
    lginCssSrc = './script/login_css/*.css',
	loginJsSrc = './script/login_js/*.js';

function logout(e) {
    console.log(e.toString());
}

////////////////////////////////////////////////////////////////////////////////////////////////////
//BASIC

function defCSS(name, src, clearFiles) {
    //login css
	var cleanName = name + '_clean';
    gulp.task(cleanName, function() {
        return del(clearFiles);
    });

	var sassName = name + '_sass';
    gulp.task(sassName, function() {
        return gulp.src(src)
            .pipe(sass({ outputStyle: 'compressed' }).on('error', sass.logError))		//- sass 编译,压缩
            .pipe(concat('login.min.css'))                      				//- 合并后的文件名
            .pipe(rev())                   											//- 文件名加MD5后缀
            .pipe(gulp.dest(destCssDir))                      						//- 输出文件本地
            .pipe(rev.manifest('./rev/css/login.json', { base: manifestDir, merge: true }))    //- 生成一个css.json
            .pipe(gulp.dest(manifestDir));                          				//- 将 css.json 保存到 rev 目录内
    });
}

//login css
gulp.task('login_css_clean', function() {
    return del([destCssDir + '/login*.css']);
});

gulp.task('login_css_sass', function() {
    return gulp.src(lginCssSrc)
        .pipe(sass({ outputStyle: 'compressed' }).on('error', sass.logError))		//- sass 编译,压缩
        .pipe(concat('login.min.css'))                      				//- 合并后的文件名
        .pipe(rev())                   											//- 文件名加MD5后缀
        .pipe(gulp.dest(destCssDir))                      						//- 输出文件本地
        .pipe(rev.manifest('./rev/css/login.json', { base: manifestDir, merge: true }))    //- 生成一个css.json
        .pipe(gulp.dest(manifestDir));                          				//- 将 css.json 保存到 rev 目录内
});

//css
gulp.task('css_clean', function() {
    return del([destCssDir + '/main*.css']);
});

gulp.task('css_sass', function() {
    return gulp.src(cssSrc)
        .pipe(sass({ outputStyle: 'compressed' }).on('error', sass.logError))		//- sass 编译,压缩
        .pipe(concat('main.min.css'))                      				//- 合并后的文件名
        .pipe(rev())                   											//- 文件名加MD5后缀
        .pipe(gulp.dest(destCssDir))                      						//- 输出文件本地
        .pipe(rev.manifest('./rev/css/css.json', { base: manifestDir, merge: true }))    //- 生成一个css.json
        .pipe(gulp.dest(manifestDir));                          				//- 将 css.json 保存到 rev 目录内
});

// js
gulp.task('js_clean', function() {
    return del([destJsDir + '/main*.js']);
});

gulp.task('js_uglify', function() {
    gulp.src(jsSrc)
        .pipe(jshint())       													//- 进行检查
        .pipe(jshint.reporter('default'))  										//- 对代码进行报错提示
        .pipe(concat('main.js'))												//- 合成main.js
        .pipe(gulp.dest(destJsDir))												//- 输出
        .pipe(uglify().on('error', logout))										//- 压缩
        .pipe(rename("main.min.js"))											//- 当作是改名吧
        .pipe(rev())                    										//- 文件名加MD5后缀
        .pipe(gulp.dest(destJsDir))                      						//- 输出文件本地
        .pipe(rev.manifest('./rev/js/js.json', { base: manifestDir, merge: true }))     //- 生成一个js.json
        .pipe(gulp.dest(manifestDir));                          				//- 将js.json 保存到 rev 目录内
});

// login js
gulp.task('loginjs_clean', function() {
    return del([destJsDir + '/login*.js']);
});

gulp.task('loginjs_uglify', function() {
    gulp.src(loginJsSrc)
        .pipe(jshint())       													//- 进行检查
        .pipe(jshint.reporter('default'))  										//- 对代码进行报错提示
        .pipe(concat('login.js'))												//- 合成main.js
        .pipe(gulp.dest(destJsDir))												//- 输出
        .pipe(uglify().on('error', logout))										//- 压缩
        .pipe(rename("login.min.js"))											//- 当作是改名吧
        .pipe(rev())                    										//- 文件名加MD5后缀
        .pipe(gulp.dest(destJsDir))                      						//- 输出文件本地
        .pipe(rev.manifest('./rev/js/login.json', { base: manifestDir, merge: true }))     //- 生成一个js.json
        .pipe(gulp.dest(manifestDir));                          				//- 将js.json 保存到 rev 目录内
});


// vendor js
gulp.task('vendorjs_clean', function() {
    return del([destJsDir + '/vendors*.js']);
});

gulp.task('vendorjs_concat', function() {
    gulp.src(vendorJsSrc)
        .pipe(concat('vendors.js'))												//- 合成vendors.js
        .pipe(uglify().on('error', logout))										//- 压缩
        .pipe(rename("vendors.min.js"))											//- 当作是改名吧
        .pipe(rev())                    										//- 文件名加MD5后缀
        .pipe(gulp.dest(destJsDir))                      						//- 输出文件本地
        .pipe(rev.manifest('./rev/vendor/js.json', { base: manifestDir, merge: true }))     //- 生成一个vendor.json
        .pipe(gulp.dest(manifestDir));                          				//- 将vendor.json 保存到 rev 目录内
});

// vendor css
gulp.task('vendorcss_clean', function() {
    return del([destCssDir + '/vendors*.css']);
});

gulp.task('vendorcss_css', function() {
    return gulp.src(vendorCssSrc)
        .pipe(concat('vendor.min.css'))                      				//- 合并后的文件名
        .pipe(rev())                   											//- 文件名加MD5后缀
        .pipe(gulp.dest(destCssDir))                      						//- 输出文件本地
        .pipe(rev.manifest('./rev/vendor/css.json', { base: manifestDir, merge: true }))    //- 生成一个css.json
        .pipe(gulp.dest(manifestDir));                          				//- 将 css.json 保存到 rev 目录内
});

// html
gulp.task('html_clean', function() {
    return del([destHtml + '/app_layout.html', destHtml + '/login.html']);
});

gulp.task('rev', function() {
    gulp.src([manifestFile, htmlSrc])											//- 读取 *.json 文件以及需要进行css名替换的文件
        .pipe(revCollector(/*{
            replaceReved: true,
        }*/))													//- 执行文件内css名的替换
        .pipe(gulp.dest(destHtml));												//- 替换后的文件输出的目录
});

/////////////////////////////////////////////////////////////////////////////
// task define
var basicTasks = [
    { name: 'js', src: jsSrc, seq: ['js_clean', 'js_uglify', 'rev'] },
    { name: 'css', src: cssSrc, seq: ['css_clean', 'css_sass', 'rev'] },
    { name: 'vendorjs', src: vendorJsSrc, seq: ['vendorjs_clean', 'vendorjs_concat', 'rev'] },
    { name: 'vendorcss', src: vendorCssSrc, seq: ['vendorcss_clean', 'vendorcss_css', 'rev'] },
    { name: 'login_css', src: lginCssSrc, seq: ['login_css_clean', 'login_css_sass', 'rev'] },
	{ name: 'loginjs', src: loginJsSrc, seq: ['loginjs_clean', 'loginjs_uglify', 'rev'] },
	// new tasks here
    { name: 'html', src: htmlSrc, seq: ['rev'] },
];

var names = new Array();
var wNames = new Array();
for (var i = 0; i < basicTasks.length; i++) {
    (function(bt) {
        //SUB TASK
        gulp.task(bt.name, function() {
            runSequence.apply(runSequence, bt.seq);
        });
        names.push(bt.name);

        //SUB WATCH
        var wn = bt.name + '_watch';
        gulp.task(wn, function() { gulp.watch(bt.src, [bt.name]); });
        wNames.push(wn);
    })(basicTasks[i]);
}

//WATCH ALL
gulp.task('watch', wNames);

//DEFAULT
var html = names.pop();
gulp.task('default', function() {
    runSequence(names, html);
});
