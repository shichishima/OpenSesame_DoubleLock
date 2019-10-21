# OpenSesame_DoubleLock
AWS IoT エンタープライズボタンを使ってスマートロック「SESAME(セサミ)」のダブルロック(1ドア2ロック)を同時に施錠/解錠するためのLambda関数。 IoTボタンの シングルクリック・ダブルクリック・長押し それぞれの動作を 同時施錠・同時解錠 に自由に設定できる。 Go言語の go func を用いることで、上錠・下錠のシリアル動作ではなく、APIの並列呼び出しによる同時施錠・同時解錠を実現している。

なお、スマホのウィジェット等から操作する場合(iOSのショートカット機能から操作する場合など)は、このLambda関数を起動するために、前段に AWS API Gateway を用いたAPIを用意する必要がある。

# 事前準備

- 設定ずみの2組のセサミ(または セサミmini) と Wi-Fiアダプター
  - セサミAPIキー
  - 2つのデバイスの セサミID
    - 取得方法： [APIキーを取得する & セサミIDを確認方法](https://ameblo.jp/candyhouse-inc/entry-12416936959.html)
  - AWSアカウント


# 設定方法
- バイナリファイルをダウンロード： `OpenSesame_DoubleLock.zip`
- Lambdaにアップロード
  - 一から作成
  - 関数名＝ `OpenSesame_DoubleLock`
  - ランタイム＝Go 1.x
  - 関数の作成
- (次の画面で) コードエントリタイプ＝.zipファイルをアップロード
  - ハンドラ＝ `OpenSesame_DoubleLock`
- 環境変数をセット
  - `DEVICE1` : 錠(1)のデバイスID
  - `DEVICE2` : 錠(2)のデバイスID
  - `APIKEY` : セサミAPIキー
  - `SINGLE` : シングルクリック時の動作 (施錠＝`lock` / 解錠＝`unlock` / 同期＝`sync` / 何もしない＝空)
  - `DOUBLE` : ダブルクリック時の動作 (`lock` / `unlock` / `sync` / 空)
  - `LONG` : 長押し時の動作 (`lock` / `unlock` / `sync` / 空)
  - `SYNC` : `sync` 固定値をセットしておく。(IoTボタン以外からlambdaを起動した時に利用可能な引数。CloudWatchによる定期同期に利用)
- 保存
- トリガーを追加：(参考) [AWS Lambdaの初回起動が遅い問題を解決する魔法の設定](https://qiita.com/yukitter/items/af77beb4c77dae1d7a1f)
  - CloudWatch Events
  - 新規ルールの作成
  - ルール名＝ `every_5min`
  - ルールの説明＝ コンテナ維持のため5分おきに空実行
  - ルールタイプ＝ スケジュール式
  - スケジュール式＝ `rate(5 minutes)`
  - トリガの有効化 にチェック
  - 追加

ここまでの設定で、IoTボタンの操作でダブルロックの同時施錠/同時解錠ができるようになった。


# (任意) CloudWatch の設定
必要に応じ、本lambda関数の「設定」「＋トリガーを追加」から以下を設定する。 (目的：セサミのlock/unlock状態をサーバ側と同期する)

- ＋トリガーを追加
  - CloudWatch Events
  - 新規ルールの作成
  - ルール名＝ `sync_05_and_17_everyday`
  - ルールの説明＝ セサミを毎日5時と17時に同期する
  - Cron式： `cron(0 8,20 * * ? *)` (17時と5時のUTC表記)
  - トリガの有効化 にチェック
  - 追加
- `sync_05_and_17_everyday` をクリックして CloudWatch Events 画面に移動
  - アクション - 編集
  - 入力の設定 : 定数(JSONテキスト)
```
{
  "deviceEvent": {
    "buttonClicked": {
      "clickType": "SYNC"
    }
  }
}
```
- 設定の詳細
- ルールの更新


# (任意) API Gateway
IoTボタンではなく、スマホのウィジェットから施錠/解錠を行う場合(iOSのショートカット機能から操作する場合)には 次のような AWS API Gateway を設定する。

## APIの作成
- プロトコル＝REST
- 新しいAPI
- API名＝ `OpenSesameAPI`
- エンドポイントタイプ＝リージョン
- 「APIの作成」ボタン押下

## プルダウン「アクション」から「リソースの作成」
- リソース名＝ `lock`
- 「リソースの作成」ボタン押下

## 「`/`」を選択し、プルダウン「アクション」から「リソースの作成」
- リソース名＝ `unlock`
- 「リソースの作成」ボタン押下

## 「`/lock`」を選択し、プルダウン「アクション」から「メソッドの作成」
- 「GET」でチェック
- 統合タイプ＝Lambda関数
- Lambdaリージョン＝適切なものを選ぶ (ap-northeast-1等)
- Lambda関数＝`OpenSesame_DoubleLock`
- デフォルトタイムアウトの使用＝ON
- 「保存」押下
- 「Lambda 関数に権限を追加する」ダイアログで「OK」
- 「メソッドリクエスト」リンク押下
  - 「APIキーの必要性」を true に変更
- 「統合リクエスト」リンク押下
  - マッピングテンプレート
  - リクエスト本文のパススルー＝なし
  - マッピングテンプレートの追加
  - `application/json`
```
{
  "deviceEvent": {
    "buttonClicked": {
      "clickType": "DOUBLE"   (lockに対応するボタン操作を設定する)
    }
  }
}
```
- 保存
- テスト
  - 施錠動作することを確認

## 「`/unlock`」を選択し、プルダウン「アクション」から「メソッドの作成」
- 「GET」でチェック
- 統合タイプ＝Lambda関数
- Lambdaリージョン＝適切なものを選ぶ (ap-northeast-1等)
- Lambda関数＝`OpenSesame_DoubleLock`
- デフォルトタイムアウトの使用＝ON
- 「保存」押下
- 「Lambda 関数に権限を追加する」ダイアログで「OK」
- 「メソッドリクエスト」リンク押下
  - 「APIキーの必要性」を true に変更
- 「統合リクエスト」リンク押下
  - マッピングテンプレート
  - リクエスト本文のパススルー＝なし
  - マッピングテンプレートの追加
  - `application/json`
```
{
  "deviceEvent": {
    "buttonClicked": {
      "clickType": "LONG"   (unlockに対応するボタン操作を設定する)
    }
  }
}
```
- 保存
- テスト
  - 解錠動作することを確認

## 使用量プラン
- 「作成」
  - 名前＝「1秒1回」
  - スロットリング - レート = 1
  - スロットリング - バースト = 1
  - クォータを有効にする＝オフ (クォータ無効)
  - 「次へ」
- 関連付けられたAPIステージ
  - APIステージの追加
  - OpenSesameAPI - prod
  - 保存(チェック)
-  「次へ」

## APIキー
- 「アクション」から「APIキーの作成」
- 名前を設定 (利用者名など)
- 自動生成
- 「保存」 必要な人数分作成する


# 参考

## 自前でのビルド方法
- バイナリビルド
```
$ GOOS=linux GOARCH=amd64 go build -o OpenSesame_DoubleLock OpenSesame_DoubleLock.go
```
- zip圧縮
```
$ zip OpenSesame_DoubleLock.zip OpenSesame_DoubleLock
```

## テストデータ

### シングルクリック用のテストデータ
イベント名： `SINGLE`
```
{
  "deviceEvent": {
    "buttonClicked": {
      "clickType": "SINGLE"
    }
  }
}
```

### ダブルクリック用のテストデータ
イベント名： `DOUBLE`
```
{
  "deviceEvent": {
    "buttonClicked": {
      "clickType": "DOUBLE"
    }
  }
}
```

### ボタン長押し用のテストデータ
イベント名： `LONG`
```
{
  "deviceEvent": {
    "buttonClicked": {
      "clickType": "LONG"
    }
  }
}
```
