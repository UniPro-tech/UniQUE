# UniQUE-Auth

## How to Run

### Manual

HTTP で動作をテストする場合、`.env`ファイルを作成し以下の内容を記述してください。

```
REQUIRE_TLS=FALSE
REQUIRE_CLIENT_AUTH=FALSE
```

RSA鍵を利用する場合は以下のコマンドを実行してください
```
ssh-keygen -t RSA -b 2048
openssl genpkey -algorithm RSA -out rsa_private.pem -pkeyopt rsa_keygen_bits:2048
```


### Docker

dockercompose で立ててください。

## How to Develop

1. Clone this Repo.
2. Run `uv sync`
3. Run `pre-commit install`
