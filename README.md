# commute2invoice

交通費精算Webアプリ

macOSのApple Containerを使う起動手順は、[macOS Containerでの起動](docs/macos-container.md)を参照してください。

## Docker Composeでの起動

Docker Desktopなど、Docker Composeを利用できる環境では次のコマンドで起動できます。

```sh
docker compose up --build --detach
```

ブラウザで <http://localhost:8080> を開いてください。SQLiteのデータベースは
ホストの `./data/commute2invoice.db` に保存されるため、コンテナを停止・再作成しても
保持されます。

```sh
docker compose ps
curl --fail http://localhost:8080/health
docker compose logs --follow app
docker compose down
```

データも削除する場合は、`docker compose down` の後に `./data` ディレクトリを削除してください。

## 開発時のローカル起動

Go 1.21以降があれば、次のコマンドでローカル起動できます。PDF生成には
`wkhtmltopdf` が必要なため、PDFを含む動作確認は上記のContainer環境で行ってください。

```sh
go run ./src/cmd
```
