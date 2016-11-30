
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
	cssSrc = './script/css/**/*.css',
	htmlSrc = './script/tmpl/**/*.html',
	destJsDir = './js',
	destCssDir = './css',
	destHtml = './tmpl',
	manifestDir = './rev',
	manifestFile = './rev/**/*.json',
	vendorJsSrc = ['./vendors/bootstrap-3.3.0/js/bootstrap.min.js',
				 './vendors/bootstrap-table-1.11.0/bootstrap-table.min.js',
				 './vendors/meny/meny.js',
				 './vendors/bootstrap-table-1.11.0/extensions/export/bootstrap-table-export.js',
				 './vendors/tableExport.jquery.plugin/tableExport.js',
				 './vendors/bootstrap-table-1.11.0/extensions/editable/bootstrap-table-editable.js',
				 './vendors/x-editable/bootstrap-editable.js'],
	vendorCssSrc = ['./vendors/bootstrap-3.3.0/css/bootstrap.min.css',
					'./vendors/bootstrap-3.3.0/css/bootstrap-theme.min.css',
					'./vendors/bootstrap-table-1.11.0/bootstrap-table.css',
					'./vendors/bootstrap3-editable/css/bootstrap-editable.css'];

function logout (e) {
	console.log(e.toString());
}

////////////////////////////////////////////////////////////////////////////////////////////////////
//BASIC

//css
gulp.task('css_clean', function() {
	return del([destCssDir + '/main*.css']);
});

gulp.task('css_sass', function () {
	return gulp.src(cssSrc)
		.pipe(sass({outputStyle: 'compressed'}).on('error', sass.logError))		//- sass 编译,压缩
		.pipe(concat('main.min.css'))                      				//- 合并后的文件名
		.pipe(rev())                   											//- 文件名加MD5后缀
		.pipe(gulp.dest(destCssDir))                      						//- 输出文件本地
		.pipe(rev.manifest( './rev/css/css.json', {base: manifestDir, merge: true}))    //- 生成一个css.json
		.pipe(gulp.dest(manifestDir));                          				//- 将 css.json 保存到 rev 目录内
});

// js
gulp.task('js_clean', function() {
	return del([destJsDir + '/main*.js']);
});

gulp.task('js_uglify', function () {
	gulp.src(jsSrc)
		.pipe(jshint())       													//- 进行检查
		.pipe(jshint.reporter('default'))  										//- 对代码进行报错提示
		.pipe(concat('main.js'))												//- 合成main.js
		.pipe(gulp.dest(destJsDir))												//- 输出
		.pipe(uglify().on('error', logout))										//- 压缩
		.pipe(rename("main.min.js"))											//- 当作是改名吧
		.pipe(rev())                    										//- 文件名加MD5后缀
		.pipe(gulp.dest(destJsDir))                      						//- 输出文件本地
		.pipe(rev.manifest('./rev/js/js.json', {base: manifestDir, merge: true}))     //- 生成一个js.json
		.pipe(gulp.dest(manifestDir));                          				//- 将js.json 保存到 rev 目录内
});

//TODO: images

// vendor js

gulp.task('vendor_clean', function() {
	return del([destJsDir + '/vendors*.js']);
});

gulp.task('vendor_concat', function () {
	gulp.src(vendorJsSrc)
		.pipe(concat('vendors.js'))												//- 合成vendors.js
		.pipe(uglify().on('error', logout))										//- 压缩
		.pipe(rename("vendors.min.js"))											//- 当作是改名吧
		.pipe(rev())                    										//- 文件名加MD5后缀
		.pipe(gulp.dest(destJsDir))                      						//- 输出文件本地
		.pipe(rev.manifest('./rev/vendor/js.json', {base: manifestDir, merge: true}))     //- 生成一个vendor.json
		.pipe(gulp.dest(manifestDir));                          				//- 将vendor.json 保存到 rev 目录内
});

// vendor css
gulp.task('vendor_css_clean', function() {
	return del([destCssDir + '/vendors*.css']);
});

gulp.task('vendor_css', function () {
	return gulp.src(vendorCssSrc)
		.pipe(concat('vendor.min.css'))                      				//- 合并后的文件名
		.pipe(rev())                   											//- 文件名加MD5后缀
		.pipe(gulp.dest(destCssDir))                      						//- 输出文件本地
		.pipe(rev.manifest( './rev/vendor/css.json', {base: manifestDir, merge: true}))    //- 生成一个css.json
		.pipe(gulp.dest(manifestDir));                          				//- 将 css.json 保存到 rev 目录内
});

// html
gulp.task('html_clean', function() {
	return del([destHtml + '/app_layout.html']);
});

gulp.task('rev', function() {
	gulp.src([manifestFile, htmlSrc])											//- 读取 *.json 文件以及需要进行css名替换的文件
		.pipe(revCollector(/*{
            replaceReved: true,
        }*/))													//- 执行文件内css名的替换
		.pipe(gulp.dest(destHtml));												//- 替换后的文件输出的目录
});


////////////////////////////////////////////////////////////////////////////////////////////////////
//SUB TASK
gulp.task('js', function() {
	runSequence('js_clean', 'js_uglify', 'rev');
});

gulp.task('css', function() {
	runSequence('css_clean', 'css_sass', 'rev');
});

gulp.task('vendors', function() {
	runSequence('vendor_clean', 'vendor_css_clean', 'vendor_concat', 'vendor_css', 'rev');
});

gulp.task('html', function() {
	runSequence(/*'html_clean',*/ 'rev');
});

////////////////////////////////////////////////////////////////////////////////////////////////////
//WATCH
gulp.task('css_watch', function () {
	gulp.watch(cssSrc, ['css']);
});

gulp.task('js_watch', function () {
	gulp.watch(jsSrc, ['js']);
});

gulp.task('vendors_watch', function () {
	gulp.watch(vendorJsSrc, ['js']);
});

gulp.task('html_watch', function () {
	gulp.watch(htmlSrc, ['html']);
});

gulp.task('watch', ['js_watch', 'css_watch', 'vendors_watch', 'html_watch']);

////////////////////////////////////////////////////////////////////////////////////////////////////
//DEFAULT
gulp.task('default', function() {
	runSequence(['js', 'css', 'vendors'], 'html');
});
