# playcheck

Google Play Store申請前に**あらゆるAndroidプロジェクト**のコンプライアンス問題を検出するCLIツール。ネイティブAndroid（Kotlin/Java）、Flutter、React Native、その他のフレームワークに対応。ポリシー違反、危険なパーミッション、セキュリティ問題、データ安全性の不備を開発早期に発見します。

> 💡 **[Greenlight](https://github.com/cpisciotta/Greenlight)に触発されました** - iOSで実現されているコンプライアンススキャン機能をAndroid開発者にも。

[English](README.md) | 日本語

## 特徴

- **Manifest検証** - SDKバージョンチェック、危険なパーミッション、エクスポートされたコンポーネント、平文通信
- **コードスキャン** - HTTP URL、SMS API使用、広告ID、弱い暗号化、サードパーティSDKのデータ収集を検出
- **データ安全性コンプライアンス** - プライバシーポリシー検出、アカウント削除要件、パーミッション開示チェック、ユーザー同意検証
- **31以上のポリシールール** - 危険なパーミッション、プライバシー、SDKコンプライアンス、アカウント管理、セキュリティなどをカバー
- **複数の出力形式** - カラーターミナル出力とCI/CD統合用のJSON
- **重大度フィルタリング** - 重大度レベル（critical、warning、info）で絞り込み可能

## インストール

### Go installから

```bash
go install github.com/kotaroyamazaki/playcheck/cmd/playcheck@latest
```

### ソースからビルド

```bash
git clone https://github.com/kotaroyamazaki/playcheck.git
cd playcheck
go build -o playcheck ./cmd/scanner
```

## 使い方

### 基本的なスキャン

```bash
playcheck scan /path/to/android/project
```

### 出力形式

```bash
# カラーターミナル出力（デフォルト）
playcheck scan ./my-app

# CI/CDパイプライン用のJSON出力
playcheck scan ./my-app --format json

# ファイルにレポートを書き出し
playcheck scan ./my-app --format json --output report.json
```

### 重大度フィルタリング

```bash
# CRITICALな問題のみ表示
playcheck scan ./my-app --severity critical

# WARNINGレベル以上を表示
playcheck scan ./my-app --severity warn
```

### CI/CD統合

```bash
# 終了コード: 0 = critical問題なし、1 = critical問題あり
playcheck scan ./my-app
if [ $? -eq 1 ]; then
  echo "Critical issues found!"
  exit 1
fi
```

## サポートされているルール

### 危険なパーミッション（10ルール）

| ルールID | 重大度 | 説明 |
|---------|--------|------|
| DP001 | CRITICAL | SMS パーミッション使用（READ_SMS、SEND_SMS、RECEIVE_SMS） |
| DP002 | CRITICAL | 通話ログパーミッション使用 |
| DP003 | CRITICAL | バックグラウンド位置情報パーミッション |
| DP004 | WARNING | カメラパーミッションが宣言されているがコードで使用されていない |
| DP005 | ERROR | 広範なストレージアクセス（MANAGE_EXTERNAL_STORAGE） |
| DP006 | WARNING | 正確なアラームパーミッション |
| DP007 | WARNING | 全パッケージクエリパーミッション |
| DP008 | WARNING | アクセシビリティサービス使用 |
| DP009 | WARNING | VPN サービス |
| DP010 | WARNING | フォアグラウンドサービスタイプ |

### プライバシー＆データ安全性（4ルール）

| ルールID | 重大度 | 説明 |
|---------|--------|------|
| PDS001 | CRITICAL | プライバシーポリシーの欠落 |
| PDS002 | ERROR | 開示なしのデータ収集 |
| PDS003 | ERROR | データ安全性セクションの不一致 |
| PDS004 | WARNING | データ削除メカニズムの欠落 |

### SDKコンプライアンス（4ルール）

| ルールID | 重大度 | 説明 |
|---------|--------|------|
| SDK001 | CRITICAL | 古いTarget SDKバージョン（< 35） |
| SDK002 | WARNING | 古いPlay Coreライブラリバージョン |
| SDK003 | ERROR | 広告同意要件違反 |
| SDK004 | WARNING | 非推奨API使用 |

### アカウント管理（2ルール）

| ルールID | 重大度 | 説明 |
|---------|--------|------|
| AD001 | CRITICAL | アカウント削除オプションの欠落 |
| AD002 | WARNING | ログインデータ開示要件 |

### Manifest検証（5ルール）

| ルールID | 重大度 | 説明 |
|---------|--------|------|
| MV001 | WARNING | アプリアイコンの欠落 |
| MV002 | ERROR | デバッグ可能なビルド |
| MV003 | WARNING | バージョンコードの欠落 |
| MV004 | WARNING | バックアップルールの欠落 |
| MV005 | WARNING | インテントフィルターの問題 |

### コンテンツポリシー（1ルール）

| ルールID | 重大度 | 説明 |
|---------|--------|------|
| MC001 | WARNING | コンテンツレーティング |

### 収益化（1ルール）

| ルールID | 重大度 | 説明 |
|---------|--------|------|
| MP002 | ERROR | Play Store以外の課金システム |

### セキュリティ（4ルール）

| ルールID | 重大度 | 説明 |
|---------|--------|------|
| MS001 | ERROR | 平文トラフィック許可 |
| MS002 | CRITICAL | ハードコードされたシークレット |
| MS003 | WARNING | エクスポートされたコンポーネント |
| MS004 | WARNING | WebView JavaScriptインターフェース |

### コードスキャン（14ルール）

| ルールID | 重大度 | 説明 |
|---------|--------|------|
| CS001 | ERROR | 暗号化されていないHTTP URL検出 |
| CS002 | INFO | プライバシーポリシーURL検出 |
| CS003 | WARNING | Firebase Analytics SDK使用検出 |
| CS004 | WARNING | AdMob SDK使用検出 |
| CS005 | WARNING | 広告ID使用検出 |
| CS006 | WARNING | アカウント作成パターン検出 |
| CS007 | INFO | アカウント削除パターン検出 |
| CS008 | CRITICAL | SMS API使用検出 |
| CS009 | WARNING | 位置情報API使用検出 |
| CS010 | WARNING | カメラAPI使用検出 |
| CS011 | ERROR | 弱い暗号化使用検出 |
| CS012 | WARNING | WebView JavaScript有効 |
| CS013 | WARNING | Facebook SDK使用検出 |
| CS014 | WARNING | サードパーティ追跡SDK検出 |

## 出力例

### ターミナル出力

```
Scanning...

=== Play Store Compliance Report ===
Project: /path/to/android/project
Duration: 123.45ms

CRITICAL (2)
  [CRITICAL] Target SDK 34 is below required version 35 (SDK001)
     Location: AndroidManifest.xml:5
     Fix: Update targetSdkVersion to 35

  [CRITICAL] SMS permission without disclosure (DP001)
     Location: AndroidManifest.xml:12
     Fix: Remove SMS permissions or submit declaration form

WARNING (5)
  [WARNING] Firebase Analytics SDK detected (CS003)
     Location: MainActivity.kt:23
     Fix: Disclose in Data Safety section

Summary:
  CRITICAL: 2
  WARN: 5
  INFO: 3

❌ Fix CRITICAL issues before submission.
```

### JSON出力

```json
{
  "timestamp": "2026-02-15T12:00:00Z",
  "project_path": "/path/to/android/project",
  "summary": {
    "total_checks": 3,
    "passed": 0,
    "failed": 3,
    "critical": 2,
    "warning": 5,
    "info": 3,
    "duration": "123.45ms"
  },
  "findings": [
    {
      "check_id": "SDK001",
      "severity": "CRITICAL",
      "title": "Target SDK version 34 below required 35",
      "description": "targetSdkVersion is 34 but Play Store requires >= 35",
      "location": "AndroidManifest.xml:5",
      "suggestion": "Update targetSdkVersion to 35 or higher"
    }
  ]
}
```

## プロジェクト構造

```
playcheck/
├── cmd/scanner/              # CLIエントリポイント
├── internal/
│   ├── cli/                  # Cobraコマンド実装
│   ├── preflight/            # コアオーケストレーション
│   ├── manifest/             # AndroidManifest.xml検証
│   ├── codescan/             # Kotlin/Javaコードスキャナー
│   ├── datasafety/           # データ安全性コンプライアンスチェッカー
│   └── policies/             # 埋め込みポリシーデータベース
├── pkg/utils/                # 共有ユーティリティ
└── testdata/                 # テストフィクスチャ
```

## 開発

### テスト実行

```bash
# すべてのテストを実行
go test ./...

# 特定のパッケージをテスト
go test ./internal/manifest
go test ./internal/codescan
go test ./internal/datasafety

# カバレッジ付きテスト
go test -cover ./...

# 統合テストを実行
go test -v .
```

### 新しいルール追加

1. ルールを `internal/policies/policies.json` に追加
2. 検出ロジックを適切なスキャナーに実装:
   - Manifestルール → `internal/manifest/validator.go`
   - コードパターン → `internal/codescan/rules.go`
   - データ安全性 → `internal/datasafety/checker.go`
3. ルールのテストを追加
4. README.mdとREADME.ja.mdのルールテーブルを更新

## Google Play Storeポリシー準拠

このツールは**2026年のGoogle Play Store要件**に基づいて実装されています：

- 新規アプリはTarget API level 35（Android 15）必須
- データ安全性セクションの開示義務
- アカウント作成機能があるアプリはアカウント削除機能必須
- 制限付きパーミッション（SMS、通話ログ）は宣言フォーム提出必須
- バックグラウンド位置情報は正当化必須

## 貢献

貢献を歓迎します！以下の方法で参加できます：

1. このリポジトリをフォーク
2. 機能ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを開く

### 貢献ガイドライン

- すべての新機能にテストを追加してください
- コードスタイルは `gofmt` と `go vet` に従ってください
- プルリクエストの説明は明確に記述してください
- 既存のテストがすべて通ることを確認してください

## ライセンス

MIT License - 詳細は[LICENSE](LICENSE)ファイルを参照してください。

## 謝辞

このプロジェクトは以下のオープンソースプロジェクトを使用しています：

- [Cobra](https://github.com/spf13/cobra) - 強力なCLIフレームワーク
- [Color](https://github.com/fatih/color) - ターミナルカラー出力
- [Progressbar](https://github.com/schollz/progressbar) - プログレスバー表示

## サポート

- 🐛 バグ報告: [GitHub Issues](https://github.com/yourusername/playcheck/issues)
- 💡 機能リクエスト: [GitHub Issues](https://github.com/yourusername/playcheck/issues)
- 📖 ドキュメント: [GitHub Wiki](https://github.com/yourusername/playcheck/wiki)

## 免責事項

このツールは開発者がGoogle Play Storeポリシーの潜在的な問題を早期に発見するのを支援しますが、すべての問題を検出することを保証するものではありません。Google Play Consoleへの申請前に、すべてのポリシーを確認し、必要な対応を行ってください。

---

**Built with ❤️ for Android developers**
