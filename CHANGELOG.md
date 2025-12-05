## [2.2.1](https://github.com/parlorhub/api-core/compare/v2.2.0...v2.2.1) (2025-12-05)

### Bug Fixes

* merge sql ([8cd1b1b](https://github.com/parlorhub/api-core/commit/8cd1b1b6ed487e7da567b3a1bfe720cd10cf1bf2))
* merge sql migrations ([bab35bb](https://github.com/parlorhub/api-core/commit/bab35bb0d357e3a34b230ec4347c574df9ee3f66))
* merge sql migrations ([5f8d065](https://github.com/parlorhub/api-core/commit/5f8d065b929dec8b7e18fad2237e58c934d7c4a6))

## [2.2.0](https://github.com/parlorhub/api-core/compare/v2.1.0...v2.2.0) (2025-12-04)

### Features

* add onboarding ([33442a1](https://github.com/parlorhub/api-core/commit/33442a15927e3976d1dbc071ab4c66ceac385c8f))
* add sso, rbac and ci artifact release ([f1b47d9](https://github.com/parlorhub/api-core/commit/f1b47d93956f01f89566274cc1b91111bdc32486))
* added onboarding ([3fa56f7](https://github.com/parlorhub/api-core/commit/3fa56f78c85a99285651632f46d4a274e7969f6b))

## [2.1.0](https://github.com/parlorhub/api-core/compare/v2.0.0...v2.1.0) (2025-12-03)

### Features

* update nginx port mapping to use environment variable ([9edac7f](https://github.com/parlorhub/api-core/commit/9edac7f897eec611ebbad9491dd263489f62b8f1))

## [2.0.0](https://github.com/parlorhub/api-core/compare/v1.4.1...v2.0.0) (2025-12-03)

### ⚠ BREAKING CHANGES

* update user role column

### Features

* added logger and implemented it ([4a3e174](https://github.com/parlorhub/api-core/commit/4a3e174011069b6b696eb54ae5762706dd8b1213))
* added sso ([e26cb18](https://github.com/parlorhub/api-core/commit/e26cb185c82c059282f29c9f2f04c6c4ac122987))
* added sso and rbac ([5a1335a](https://github.com/parlorhub/api-core/commit/5a1335a87c37c7370ddaea995a354e59b7f7ee84))
* added sso, env.local logic and improved usage of env vars ([0d328fa](https://github.com/parlorhub/api-core/commit/0d328fa997f9f63df8ad35cfed3ee6dc0e0009ae))
* added sso, env.local logic and improved usage of env vars ([61b669a](https://github.com/parlorhub/api-core/commit/61b669a6fcc2dde4baeb91077b205f1f432a7bc5))
* created user basic rbac ([cfd06ef](https://github.com/parlorhub/api-core/commit/cfd06ef9e9111428ac65f060a7f1dda0ae836a1f))
* rbac ([f0fc9fc](https://github.com/parlorhub/api-core/commit/f0fc9fcf4c2e5300828109d53bea7db3bc9dc5f9))

### Bug Fixes

* fixed conflicts at routes.go ([7fbbae8](https://github.com/parlorhub/api-core/commit/7fbbae862844bb20f0fd30e1bafddde1cbbabb80))
* fixed refresh_token url ([0b5bdff](https://github.com/parlorhub/api-core/commit/0b5bdff47b99526a22f533be0b24df40d1256a46))
* fixed refresh_token url ([b7bf6b8](https://github.com/parlorhub/api-core/commit/b7bf6b871e600b427ea692352625e6bf1fc1bdc0))
* nginx not starting up ([0072e67](https://github.com/parlorhub/api-core/commit/0072e6724faccb2040e824388ad36b5dbdbcb6f8))
* rbac middleware and contact form handler ([37d80e5](https://github.com/parlorhub/api-core/commit/37d80e53cb348a17932962f0a049239867168e6c))
* rbac middleware and contact form handler ([c210d52](https://github.com/parlorhub/api-core/commit/c210d52497e6c4e43be9f5485778d77dc38dca83))
* update user role column ([fd66594](https://github.com/parlorhub/api-core/commit/fd66594ddff21f5c90fb7479e3e280131d619b02))
* updated sso roles and created services, controller and middleware ([4290e4f](https://github.com/parlorhub/api-core/commit/4290e4f9d9de64fdfd7d7924b56d2b7caabc4bea))
* updated sso roles and created services, controller and middleware ([fa230ab](https://github.com/parlorhub/api-core/commit/fa230ab72e570d1af9f8158e3b2ec510336b0673))

## [1.4.1](https://github.com/parlorhub/api-core/compare/v1.4.0...v1.4.1) (2025-11-29)

### Bug Fixes

* docker-compose and migrations down fixed ([ade35e0](https://github.com/parlorhub/api-core/commit/ade35e0dddc01d75d5245ee8fbb8ff2674628eeb))

## [1.4.0](https://github.com/parlorhub/api-core/compare/v1.3.0...v1.4.0) (2025-11-28)

### Features

*  add migration to add uuidv7 as the default id for all tables ([876e499](https://github.com/parlorhub/api-core/commit/876e499968ccca5b5d5fddeaae0b482d94cd9307))
*  add migration to add uuidv7 as the default id for all tables ([4a15f5d](https://github.com/parlorhub/api-core/commit/4a15f5d0ac2ff1341066e7e18795afdd87bcb508))

## [1.3.0](https://github.com/parlorhub/api-core/compare/v1.2.0...v1.3.0) (2025-11-28)

### Features

* add routes for contact form ([327c5f0](https://github.com/parlorhub/api-core/commit/327c5f0407be266ca727b97c7ede5ff57c1607ea))

## [1.2.0](https://github.com/parlorhub/api-core/compare/v1.1.0...v1.2.0) (2025-11-28)

### Features

* implement contact form functionality including database schema ([10b578f](https://github.com/parlorhub/api-core/commit/10b578fe46287feb644ef4ed615e894e1dc910cd))
* implement contact form functionality including database schema, API endpoints, and handlers. ([e27a463](https://github.com/parlorhub/api-core/commit/e27a46312fa7adf0c96713d1a7c4076375fd28de))

## [1.1.0](https://github.com/parlorhub/api-core/compare/v1.0.0...v1.1.0) (2025-11-26)

### Features

* implement email handling with template management and MailTrap integration ([e6be596](https://github.com/parlorhub/api-core/commit/e6be5964d1d131fc8791b8e68586940fdd32bd49))

### Bug Fixes

* fixed error at testhealth on database_test.go ([14e2500](https://github.com/parlorhub/api-core/commit/14e25003d48618ceff07e08af89b3865c8c1d9d3))

## 1.0.0 (2025-11-26)

### ⚠ BREAKING CHANGES

* added migrations, nginx, sqlc and created test folder
* updated references

### Features

* added authentication and session layers ([0927b5e](https://github.com/parlorhub/api-core/commit/0927b5e46a62034a0561009b87aac60e70d427da))
* added migrations, nginx, sqlc and created test folder ([7ab98ae](https://github.com/parlorhub/api-core/commit/7ab98ae99e40d1a36c7738d590dae5dddc083995))
* added redis and updated config files location ([23b3bcf](https://github.com/parlorhub/api-core/commit/23b3bcf88e1f56e692a0c3db8ec4c1c8d183266f))

### Bug Fixes

* fixed error at testhealth on database_test.go ([2caf894](https://github.com/parlorhub/api-core/commit/2caf8940fb0bcea27fe24cbca3e37df2de16e4ce))
* updated references ([ab407a7](https://github.com/parlorhub/api-core/commit/ab407a734bf06dfa52ac91714d506dfa4b5d16cb))
