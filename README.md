# mboxview

軽量な mbox ビューアのサーバ実装（Go）です。複数の mbox ファイルを読み込み、フロントエンド（`static/app.js`）が呼ぶシンプルな HTTP API を提供します。

主な特徴:

- ローカルに置いた mbox ファイル群を読み込み一覧表示
- ファイル名は IMAP-UTF7 でエンコードされている想定でデコードして表示
- メールの件名/本文はヘッダ等の charset を解析して UTF-8 に変換して返却
- プロジェクトで管理されている静的フロントエンド（`static`）が含まれており、必要に応じてソースを編集してコミットできます。ローカルでそのまますぐにブラウザで確認可能です。

## 必要要件

- Go 1.20+（go.mod の go バージョンに合わせてください）

## インストール / ビルド

```powershell
# 依存を取得（オプション）
go mod download

# ビルド
go build ./...
```

## 使い方（起動）

```powershell
# mbox ファイルが置かれたディレクトリを --path で指定して起動
go run main.go --path /path/to/mboxdir
```

デフォルトで :8080 をリッスンします。ブラウザで http://localhost:8080/ を開くと `static/index.html` が表示されます。

## API（エンドポイント）

- GET /api/mailboxes
	- 説明: 作業ディレクトリにある mbox ファイル名の一覧を返します（サーバ内部で IMAP-UTF7 -> UTF-8 にデコード）。
	- レスポンス: JSON 配列

- GET /api/mailboxes/{mailboxName}/emails
	- 説明: 指定 mailbox のメール一覧（id, from, date, subject）を返します。`{mailboxName}` は UTF-8 表示名をそのまま指定します。
	- レスポンス: JSON 配列（Email オブジェクト）

- GET /api/mailboxes/{mailboxName}/emails/{emailId}
	- 説明: 指定メールの本文と添付情報を返します。
	- レスポンス: JSON（body, bodyType, attachments）

サンプル（curl）:

```bash
curl -s http://localhost:8080/api/mailboxes | jq .
curl -s http://localhost:8080/api/mailboxes/INBOX/emails | jq .
curl -s http://localhost:8080/api/mailboxes/INBOX/emails/0 | jq .
```

## 静的ファイル

フロントエンドの静的アセットは `/static/`（リポジトリ内の `static/` ディレクトリ）で管理・提供します。主要ファイルは `static/index.html`, `static/app.js`, `static/style.css` です。

注意点:
- これらのファイルはバージョン管理されています。変更した場合はコミットしてプッシュしてください。
- サーバは `.css` と `.js` の MIME タイプを明示的に設定しますが、CI/CD やリバースプロキシが `Content-Type` を上書きする場合は配信側の設定も確認してください。
- ブラウザのキャッシュやプロキシのキャッシュが古いアセットを返すことがあるので、デプロイ後はキャッシュクリアやバージョニングを検討してください。

## ライセンス

MIT License
