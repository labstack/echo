var gulp = require('gulp');
var shell = require('gulp-shell');
var ghPages = require('gulp-gh-pages');

gulp.task('build', shell.task('mkdocs build --clean'))

gulp.task('deploy',['build'], function() {
    return gulp.src('site/**/*')
        .pipe(ghPages());
});
