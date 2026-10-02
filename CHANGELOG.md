# Changelog

## [1.17.0](https://github.com/UniPro-tech/UniQUE/compare/v1.17.0...v1.17.0) (2026-10-02)


### ⚠ BREAKING CHANGES

* **deploy:** environment variables are deprecated.

### Features

* **deploy:** Migrate deployment to Helm and store application settings in database ([#137](https://github.com/UniPro-tech/UniQUE/issues/137)) ([e064463](https://github.com/UniPro-tech/UniQUE/commit/e064463a673943641a4bc807b94825ba9aeeafcf))
* Discord通知機能を追加 ([#134](https://github.com/UniPro-tech/UniQUE/issues/134)) ([e825dd2](https://github.com/UniPro-tech/UniQUE/commit/e825dd2de93734fbf124e8e6a3f8e699a93eef42))
* **frontend:** Discordアカウントを承認画面に表示 ([#133](https://github.com/UniPro-tech/UniQUE/issues/133)) ([0234cb9](https://github.com/UniPro-tech/UniQUE/commit/0234cb9ae4f7d176c1b6ffa26b82854968459966))
* **tests:** add unit tests for toCamelcase and toSnakecase functions ([#136](https://github.com/UniPro-tech/UniQUE/issues/136)) ([25a3911](https://github.com/UniPro-tech/UniQUE/commit/25a3911c8884714017fb5d5ff5d9c1ac88ab3433))
* ウェルカムメッセージの内容を更新し、リンクを追加 ([5c53ce2](https://github.com/UniPro-tech/UniQUE/commit/5c53ce2fcc4681b25612ab9856a7b846816f435b))
* フォームの状態管理を追加して認証部分に関わる処理を強化 ([#44](https://github.com/UniPro-tech/UniQUE/issues/44)) ([0a97775](https://github.com/UniPro-tech/UniQUE/commit/0a97775a9761919b06b43c7912a2ebbbad179936))
* メールの確認を促すようにする ([54051b0](https://github.com/UniPro-tech/UniQUE/commit/54051b0da83cab31c5d9d8642cc54ab5aad71e56))
* 利用規約を書き込めるようにする ([#45](https://github.com/UniPro-tech/UniQUE/issues/45)) ([9c60074](https://github.com/UniPro-tech/UniQUE/commit/9c60074400bb703640244c6e68ab34ff3fbf63df))
* 新しいマイグレーションファイルを追加し、`authorization_requests`テーブルの`state`カラムの型を変更 ([9c36ae4](https://github.com/UniPro-tech/UniQUE/commit/9c36ae4b74de281a86f804012361432ec7ff4c14))


### Bug Fixes

* add section name to httproute ([b16df0b](https://github.com/UniPro-tech/UniQUE/commit/b16df0bbb72f40a1a6fd70eb776ddd1bbb4ca36e))
* biomeschemaのバージョンを2.4.4から2.4.11に更新 ([6a6bef8](https://github.com/UniPro-tech/UniQUE/commit/6a6bef8a5b160ab902da023f9ab54b0064a7238c))
* **db:** DBのスキーマの不整合を修正 ([#110](https://github.com/UniPro-tech/UniQUE/issues/110)) ([33ae037](https://github.com/UniPro-tech/UniQUE/commit/33ae0374c9d28910addf6653f45b205dc17d8f38))
* dbスキーマのミスを修正 ([3ea9479](https://github.com/UniPro-tech/UniQUE/commit/3ea9479471fa723a9e175e5df4660cac5ddd6cff))
* **deps:** update material-ui monorepo to v9 (major) ([#22](https://github.com/UniPro-tech/UniQUE/issues/22)) ([136d711](https://github.com/UniPro-tech/UniQUE/commit/136d7118c03ee93a0f404c440693c6a479e402b1))
* **deps:** update module github.com/disgoorg/disgo to v0.19.6 ([#89](https://github.com/UniPro-tech/UniQUE/issues/89)) ([5177b7d](https://github.com/UniPro-tech/UniQUE/commit/5177b7da3a7c7ddbe60e931622aa0d3a4838ea98))
* **deps:** update module github.com/go-sql-driver/mysql to v1.10.0 ([#60](https://github.com/UniPro-tech/UniQUE/issues/60)) ([7147de9](https://github.com/UniPro-tech/UniQUE/commit/7147de9b3c68c79e7f9f4d86b5ded2cd1faf69e3))
* **deps:** update module github.com/go-sql-driver/mysql to v1.10.1 ([#125](https://github.com/UniPro-tech/UniQUE/issues/125)) ([4b0bceb](https://github.com/UniPro-tech/UniQUE/commit/4b0bcebf68ba8e91cd418516a51e0da759e5cbe3))
* **deps:** update module github.com/oklog/ulid to v2 ([#15](https://github.com/UniPro-tech/UniQUE/issues/15)) ([8f62c34](https://github.com/UniPro-tech/UniQUE/commit/8f62c345ddad3b43a027fa0df78f8649ec1b99d0))
* **deps:** update module github.com/oklog/ulid to v2 ([#76](https://github.com/UniPro-tech/UniQUE/issues/76)) ([3ca8cf9](https://github.com/UniPro-tech/UniQUE/commit/3ca8cf99b8209f5e6d7b8bd1546150b3b16d6ca5))
* **deps:** update module github.com/oklog/ulid/v2 to v2.1.2 ([#126](https://github.com/UniPro-tech/UniQUE/issues/126)) ([1c553ae](https://github.com/UniPro-tech/UniQUE/commit/1c553ae3178f9053db6d5c6c1b890270c43c56b2))
* **deps:** update module github.com/swaggo/files to v2 ([#16](https://github.com/UniPro-tech/UniQUE/issues/16)) ([acd7a45](https://github.com/UniPro-tech/UniQUE/commit/acd7a45579abff1a78e3657fa9e2e919f68b91dc))
* **deps:** update module golang.org/x/crypto to v0.53.0 ([#61](https://github.com/UniPro-tech/UniQUE/issues/61)) ([0529e59](https://github.com/UniPro-tech/UniQUE/commit/0529e597221863549f9461a622de9dd990f84a74))
* **deps:** update module golang.org/x/crypto to v0.54.0 ([#107](https://github.com/UniPro-tech/UniQUE/issues/107)) ([0f0887d](https://github.com/UniPro-tech/UniQUE/commit/0f0887df670dfb6313850fa150f72b671b9e5c74))
* **deps:** update module golang.org/x/crypto to v0.57.0 ([#129](https://github.com/UniPro-tech/UniQUE/issues/129)) ([ba07822](https://github.com/UniPro-tech/UniQUE/commit/ba0782213920fa09536b3c71013bca8a19e37080))
* **deps:** update module gorm.io/gen to v0.3.28 ([#86](https://github.com/UniPro-tech/UniQUE/issues/86)) ([13ce51d](https://github.com/UniPro-tech/UniQUE/commit/13ce51d5603622ebe05bc492c86f20ab8f2d0c92))
* **deps:** update module gorm.io/gen to v0.3.29 ([#127](https://github.com/UniPro-tech/UniQUE/issues/127)) ([a014c21](https://github.com/UniPro-tech/UniQUE/commit/a014c21bf6c6bcea2c87f9e2d82cdbcc150baea3))
* **deps:** update module gorm.io/gorm to v1.31.2 ([#97](https://github.com/UniPro-tech/UniQUE/issues/97)) ([03e596b](https://github.com/UniPro-tech/UniQUE/commit/03e596bdc663f7705fb5b9832d997c7aa990a076))
* DiscordAPIのバージョンを固定 ([5b9a545](https://github.com/UniPro-tech/UniQUE/commit/5b9a5457e657d3a05ed687e0c59dfff04e5014e0))
* ExternalIDはlistできるようにする ([d415704](https://github.com/UniPro-tech/UniQUE/commit/d4157043899d412d015964404b55a28e6cadbb62))
* handle nil expiration in refresh token grant validation ([baf495b](https://github.com/UniPro-tech/UniQUE/commit/baf495bd59840e9ba7ec191549bad9192befbea4))
* idを事前に確定させてバグを修正 ([898e554](https://github.com/UniPro-tech/UniQUE/commit/898e5545db8ff76131d0f5a48a5e645692a68cc6))
* namespace miss ([ef5be2d](https://github.com/UniPro-tech/UniQUE/commit/ef5be2d28b3d2a9438da3f0c48c321dbe8520adb))
* preserve pending email verification ([#135](https://github.com/UniPro-tech/UniQUE/issues/135)) ([38ceb8c](https://github.com/UniPro-tech/UniQUE/commit/38ceb8c34240e8167b38f8eda970c6d01d75b2d2))
* セキュリティヘッダーを追加 ([#46](https://github.com/UniPro-tech/UniQUE/issues/46)) ([e46972c](https://github.com/UniPro-tech/UniQUE/commit/e46972ca0199ebecf3da5560eb883c2563d0552f))
* ユーザーリストにEmailVerifiedフィールドを追加 ([#27](https://github.com/UniPro-tech/UniQUE/issues/27)) ([d77f3cc](https://github.com/UniPro-tech/UniQUE/commit/d77f3cca06e62c36177bd5ad4f6fe44af96015aa))
* 時間軸の補正 ([#63](https://github.com/UniPro-tech/UniQUE/issues/63)) ([84e4003](https://github.com/UniPro-tech/UniQUE/commit/84e40036cdaa5fa538fe1357544fd979ce1a547e))
* 権限に関するミスを修正 ([#80](https://github.com/UniPro-tech/UniQUE/issues/80)) ([330e611](https://github.com/UniPro-tech/UniQUE/commit/330e611b790c9255d07d72460c4f933ef1c30b4b))
* 管理者である場合は自分のパスワードでもcurrentなしでリセットできるように判定ロジックを修正 ([e1bd35a](https://github.com/UniPro-tech/UniQUE/commit/e1bd35af331c6d5e7c242673fd8306ef7d70234d))
* 認証コードの連結方法のミスを修正 ([#113](https://github.com/UniPro-tech/UniQUE/issues/113)) ([384bbed](https://github.com/UniPro-tech/UniQUE/commit/384bbeddd8e90a576b3e8237f763754a81e0b391))


### Miscellaneous Chores

* force initial release version ([f050592](https://github.com/UniPro-tech/UniQUE/commit/f0505927539bb2b699d406640246c68971ed0ad8))

## Changelog

All notable changes to UniQUE are documented in this file. The project follows
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Release Please updates this file from Conventional Commit messages when it
prepares a release pull request.
