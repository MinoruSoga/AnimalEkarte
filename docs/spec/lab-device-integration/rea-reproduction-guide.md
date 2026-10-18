# REA (Reverse Engineer Anything) 再検証・新機器解析ガイド

本ドキュメントでは、新規導入された院内検査機器のバイナリや、他医院（八王子等）の未解析資材を **REA** を使ってリバースエンジニアリングする際の手順とコマンド例を記録します。

---

## 1. 事前準備・動作確認

REA CLI が正しく機能しているか確認します。

```bash
# REA の利用可能プロバイダと機能の確認
rea capabilities

# REA バージョン確認
rea --version
```

---

## 2. REA コマンドによる直接解析フロー

### ステップ 1: バイナリ基本構造・ハッシュ・アーティファクト検査
対象バイナリの形式（ELF/Mach-O/PE）や依存関係を把握します。

```bash
rea inspect-artifact "/path/to/target.exe"
```

* **出力のポイント**:
  * `root_format`: `pe`, `macho`, `elf` などの判別
  * `next_probes`: REA が次に推奨する解析操作（例: `inspect_managed_artifact`, `open_binary`）

---

### ステップ 2: マネージド (.NET) / ネイティブ (Win32 / C / VB6) の分類
バイナリが C# / .NET なのか、C/C++ や VB6 等のネイティブコードなのかを判定します。

```bash
rea inspect-managed-artifact "/path/to/target.exe"
```

* **判定結果の読み方**:
  * `classification.status: not-managed` ➔ Win32 / ネイティブバイナリ（VB6, VC++, Delphi 等）
  * `classification.status: managed` ➔ .NET アセンブリ（CIL メタデータやクラス定義の逆コンパイルが可能）

---

### ステップ 3: 依存DLL・ランタイム・埋め込み文字列の抽出
ネイティブバイナリの場合、埋め込まれているSQL文、通信コマンド、API名、エラーメッセージなどを抽出します。

```bash
# 依存しているランタイムやAPIの確認
strings "/path/to/target.exe" | grep -iE 'msvbvm|msvcr|mfc|com|serial'

# 通信コマンド（STX/ETX/ACK/NAK/ボーレート等）のキーワード検索
strings "/path/to/target.exe" | grep -iE 'COM[0-9]|baud|9600|SELECT|INSERT'
```

---

## 3. AI エージェント（Antigravity / Claude Code）連携プロンプト例

REA が MCP サーバーとして接続されている場合、AI エージェントに自然言語で指示するだけで自動的に上記ツールを組み合わせて解析してくれます。

### プロンプト例 ①：通信電文のパース関数の復元
> *「`old_db/clinics/jouto/lab-drwan/C->ProgramFiles->DRワン/jyusin.exe` を REA で解析して、シリアルポートから受け取ったバイト列をどう分割・パースしているか、ロジックを Go言語の関数として再実装して」*

### プロンプト例 ②：画面コントロールとDBテーブルの対応調査
> *「この古いアプリの実行ファイルを REA で調査し、画面上の『前回繰越金』や『検査結果』の入力欄が、Access MDB のどのテーブル・カラムに書き込まれているか特定して」*

### プロンプト例 ③：新規機器の電文キャプチャからのパーサー作成
> *「新しい検査機器のシリアル通信ログ（またはパケットキャプチャファイル）がある。この電文構造を REA で解析して、AnimalEkarte の `lab_device_*.go` に準拠したパーサーを書いて」*

---

## 4. セキュリティと倫理上の遵守事項

1. **患者個人情報の完全保護**:
   * 実際の患者カルテ情報、飼主情報を含むファイル（MDBや個人情報入りログ）は、解析対象としてプロンプトやパブリックなリポジトリに展開しないこと。
2. **安全なローカル実行**:
   * REA は完全ローカルで動作するため、機密バイナリを外部クラウドや不要な第三者サーバーにアップロードしないこと。
