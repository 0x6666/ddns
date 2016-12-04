
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

var htmlSrc = './script/tmpl/**/*.html',
    destJsDir = './assets/js',
    destCssDir = './assets/css',
    destHtml = './tmpl',
    manifestDir = './rev';

function logout(e) {
    console.log(e.toString());
}

////////////////////////////////////////////////////////////////////////////////////////////////////
//BASIC

gulp.task('rev', function () {
    gulp.src([manifestDir + '/**/*.json', htmlSrc])											//- 读取 *.json 文件以及需要进行css名替换的文件
        .pipe(revCollector())													//- 执行文件内css名的替换
        .pipe(gulp.dest(destHtml));												//- 替换后的文件输出的目录
});

gulp.task('html_watch', function () { gulp.watch(htmlSrc, ['rev']); });

var watchTasks = ['html_watch'];
var subTasks = [];

function regCSS(name, src) {
    var cleanName = name + '_css_clean';
    gulp.task(cleanName, function () {
        return del([destCssDir + '/' + name + '*.css']);
    });

    var sassName = name + '_sass';
    gulp.task(sassName, function () {
        return gulp.src(src)
            .pipe(sass({ outputStyle: 'compressed' }).on('error', sass.logError))		//- sass 编译,压缩
            .pipe(concat(name + '.min.css'))                      				//- 合并后的文件名
            .pipe(rev())                   											//- 文件名加MD5后缀
            .pipe(gulp.dest(destCssDir))                      						//- 输出文件本地
            .pipe(rev.manifest('./rev/css/' + name + '.json', { base: manifestDir, merge: true }))    //- 生成一个css.json
            .pipe(gulp.dest(manifestDir));                          				//- 将 css.json 保存到 rev 目录内
    });

    var subName = name + '_css';
    gulp.task(subName, function () {
        runSequence(cleanName, sassName, 'rev');
    });
    subTasks.push(subName);

    var wn = name + '_css_watch';
    gulp.task(wn, function () { gulp.watch(src, [subName]); });
    watchTasks.push(wn);
}

function regJs(name, src, noJshint) {
    var cleanName = name + '_js_clean';
    gulp.task(cleanName, function () {
        return del([destJsDir + '/' + name + '*.js']);
    });

    var uglifyName = name + '_uglify';
    gulp.task(uglifyName, function () {
        var _gulp = gulp.src(src);
        if (!noJshint) {
            _gulp = _gulp.pipe(jshint())       													//- 进行检查
                .pipe(jshint.reporter('default'));  										//- 对代码进行报错提示
        }
        _gulp.pipe(concat(name + '.js'))												//- 合成main.js
            .pipe(gulp.dest(destJsDir))												//- 输出
            .pipe(uglify().on('error', logout))										//- 压缩
            .pipe(rename(name + ".min.js"))											//- 当作是改名吧
            .pipe(rev())                    										//- 文件名加MD5后缀
            .pipe(gulp.dest(destJsDir))                      						//- 输出文件本地
            .pipe(rev.manifest('./rev/js/' + name + '.json', { base: manifestDir, merge: true }))     //- 生成一个js.json
            .pipe(gulp.dest(manifestDir));                          				//- 将js.json 保存到 rev 目录内
    });

    var subName = name + '_js';
    gulp.task(subName, function () {
        runSequence(cleanName, uglifyName, 'rev');
    });
    subTasks.push(subName);

    var wn = name + '_js_watch';
    gulp.task(wn, function () { gulp.watch(src, [subName]); });
    watchTasks.push(wn);
}

////////////////////////////////////////////////////////////////
// REG
regCSS('main', './script/css/**/*.scss');
regCSS('login', './script/login_css/*.css');
regCSS('vendor', ['./assets/vendors/bootstrap-3.3.0/css/bootstrap.min.css',
    './assets/vendors/bootstrap-3.3.0/css/bootstrap-theme.min.css',
    './assets/vendors/bootstrap-table-1.11.0/bootstrap-table.css',
    './assets/vendors/bootstrap3-editable/css/bootstrap-editable.css']);
regJs('main', './script/js/**/*.js');
regJs('login', './script/login_js/*.js');
regJs('vendor', ['./assets/vendors/bootstrap-3.3.0/js/bootstrap.min.js',
    './assets/vendors/bootstrap-table-1.11.0/bootstrap-table.min.js',
    './assets/vendors/meny/meny.js',
    './assets/vendors/bootstrap-table-1.11.0/extensions/export/bootstrap-table-export.js',
    './assets/vendors/tableExport.jquery.plugin/tableExport.js',
    './assets/vendors/bootstrap-table-1.11.0/extensions/editable/bootstrap-table-editable.js',
    './assets/vendors/x-editable/bootstrap-editable.js'], true);
//new task here....


///////////////////////////////////////////////////////////////
// EXPORT

//watch all
gulp.task('watch', watchTasks);
//default
gulp.task('default', function () { runSequence(subTasks, 'rev'); });
