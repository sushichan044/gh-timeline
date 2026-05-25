# Changelog

## [1.1.0](https://github.com/sushichan044/gh-timeline/compare/v1.0.1...v1.1.0) (2026-05-25)


### Features

* **githuburl:** accept URLs with extra path segments after the issue/PR number ([#12](https://github.com/sushichan044/gh-timeline/issues/12)) ([5ed51dd](https://github.com/sushichan044/gh-timeline/commit/5ed51dd26c5eea5d0daac9b22134595313284e2e))

## [1.0.1](https://github.com/sushichan044/gh-timeline/compare/v1.0.0...v1.0.1) (2026-05-24)


### Performance Improvements

* **timeline:** parallelize pagination using skip offset ([#9](https://github.com/sushichan044/gh-timeline/issues/9)) ([6a9002d](https://github.com/sushichan044/gh-timeline/commit/6a9002dab68b28ce98ed71fc88b3af90c4a00b65))

## 1.0.0 (2026-05-23)


### Features

* more descriptive summary ([3b1559f](https://github.com/sushichan044/gh-timeline/commit/3b1559fbd437dea76c6c9ec9caa51db1b13bb07b))
* show both of old and new shas for force push events ([98c0a5b](https://github.com/sushichan044/gh-timeline/commit/98c0a5b0be34c3ecdd06e8552e145b9c04e3b9f0)), closes [#2](https://github.com/sushichan044/gh-timeline/issues/2)
* support all events ([8334e57](https://github.com/sushichan044/gh-timeline/commit/8334e57fd08ecc7d54e56af1171e1d5e25a0ee12))
* support specifying issue / PR urls ([4f2483d](https://github.com/sushichan044/gh-timeline/commit/4f2483d5cc357690a946f950cd02cd91aea67e63)), closes [#3](https://github.com/sushichan044/gh-timeline/issues/3)


### Performance Improvements

* omit issue / pr title for better token efficiency ([0796559](https://github.com/sushichan044/gh-timeline/commit/07965595f2ce4fa634b5d2f98c5e99e52d119619))
