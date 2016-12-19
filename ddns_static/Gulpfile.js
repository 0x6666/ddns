
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
    gulp.src([manifestDir + '/**/*.json', htmlSrc])
        .pipe(revCollector())
        .pipe(gulp.dest(destHtml));
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
            .pipe(sass({ outputStyle: 'compressed' }).on('error', sass.logError))
            .pipe(concat(name + '.min.css'))
            .pipe(rev())
            .pipe(gulp.dest(destCssDir))
            .pipe(rev.manifest('./rev/css/' + name + '.json', { base: manifestDir, merge: true }))
            .pipe(gulp.dest(manifestDir));
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
            _gulp = _gulp.pipe(jshint())
                .pipe(jshint.reporter('default'));
        }
        _gulp.pipe(concat(name + '.js'))
            .pipe(gulp.dest(destJsDir))
            .pipe(uglify().on('error', logout))
            .pipe(rename(name + ".min.js"))
            .pipe(rev())
            .pipe(gulp.dest(destJsDir))
            .pipe(rev.manifest('./rev/js/' + name + '.json', { base: manifestDir, merge: true }))
            .pipe(gulp.dest(manifestDir));
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
