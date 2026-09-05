# macOS Containerでの起動

macOSではDocker Desktopや`docker compose`を必要とせず、Apple Containerの
`container` CLIでアプリを実行できます。

## 前提条件

- macOS Container CLIがインストール済みであること
- `container system start` を実行できること

## イメージのビルドと起動

```sh
container system start
container build --tag commute2invoice:latest .
container volume create commute2invoice-data
container run --detach \
  --name commute2invoice \
  --publish 8080:8080 \
  --mount type=volume,source=commute2invoice-data,target=/data \
  commute2invoice:latest
```

ブラウザで <http://localhost:8080> を開くとアプリを利用できます。SQLiteの
データベースは名前付きボリューム `commute2invoice-data` 内の
`/data/commute2invoice.db` に保存されるため、コンテナを作り直しても保持されます。

## 状態確認と停止

```sh
curl --fail http://localhost:8080/health
container logs commute2invoice
container stop commute2invoice
container delete commute2invoice
```

データも含めて削除する場合だけ、コンテナ削除後に次を実行してください。

```sh
container volume delete commute2invoice-data
```
