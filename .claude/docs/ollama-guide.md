# Ollama使い方ガイド

InfoCenterSystemプロジェクトにおけるOllamaの導入・使用方法を説明します。

## 目次

- [はじめに](#はじめに)
- [インストールと初期設定](#インストールと初期設定)
- [モデル管理](#モデル管理)
- [基本的な使い方](#基本的な使い方)
- [Cipher MCPとの統合](#cipher-mcpとの統合)
- [トラブルシューティング](#トラブルシューティング)
- [パフォーマンス最適化](#パフォーマンス最適化)
- [参考資料](#参考資料)

---

## はじめに

### Ollamaとは

Ollamaは、ローカル環境でLLM（大規模言語モデル）と埋め込みモデルを実行するためのオープンソースツールです。

**主な特徴:**
- 🔒 **完全ローカル実行**: データがクラウドに送信されない
- 💰 **コスト削減**: APIコール料金が不要
- ⚡ **高速**: ローカル実行のため低レイテンシ
- 🔄 **モデル管理が簡単**: コマンド一つでモデルのダウンロード・切り替えが可能

### InfoCenterSystemでの役割

このプロジェクトでは、Ollamaを **Cipher MCP（Model Context Protocol）** と統合して、以下の機能を実現しています：

1. **会話履歴のセマンティック検索**
   - ローカル埋め込みモデル（nomic-embed-text）を使用
   - 過去の会話内容を意味的に検索
   - プライバシーを保ちながら高速な検索を実現

2. **開発知識の永続化**
   - プロジェクト固有の決定事項を記憶
   - コーディングパターンの保存
   - チーム全体で知識を共有

### このドキュメントで学べること

- Ollamaのインストールと起動方法
- 埋め込みモデルの管理
- Cipher MCPとの統合設定
- トラブルシューティングとパフォーマンス最適化

---

## インストールと初期設定

### 前提条件

- **OS**: macOS (Apple Silicon または Intel)
- **ツール**: Homebrew がインストール済み
- **メモリ**: 最低 8GB RAM（推奨: 16GB以上）
- **ディスク**: 最低 5GB の空き容量

### インストール手順

#### 1. Homebrewでのインストール

```bash
# Ollamaをインストール
brew install ollama
```

#### 2. インストールの確認

```bash
# バージョン確認
ollama --version
# 出力例: ollama version is 0.13.1

# インストール場所の確認
which ollama
# 出力例: /opt/homebrew/bin/ollama
```

### サービスの起動

Ollamaは2つの方法で起動できます。

#### 方法1: バックグラウンドサービスとして起動（推奨）

```bash
# サービスとして起動（自動起動有効）
brew services start ollama
# 出力: Successfully started `ollama` (label: homebrew.mxcl.ollama)

# サービスの状態確認
brew services list | grep ollama
```

**メリット:**
- マシン再起動後も自動で起動
- バックグラウンドで常駐
- 手動で起動する必要がない

#### 方法2: 手動で起動

```bash
# フォアグラウンドで起動
ollama serve

# またはバックグラウンドで起動
ollama serve &
```

### 動作確認

Ollamaが正常に起動しているか確認します。

```bash
# プロセスの確認
ps aux | grep ollama

# APIエンドポイントの確認
curl http://localhost:11434/api/version
# 出力例: {"version":"0.13.1"}
```

---

## モデル管理

### モデルの一覧表示

```bash
# ダウンロード済みモデルの一覧
ollama list
```

**出力例:**
```
NAME                    ID              SIZE    MODIFIED
nomic-embed-text:latest 970aa74c0a90    274 MB  5 minutes ago
```

### モデルのダウンロード

#### 推奨埋め込みモデル

InfoCenterSystemでは、以下の埋め込みモデルを推奨します：

| モデル名 | 次元数 | サイズ | 用途 | 推奨度 |
|---------|--------|--------|------|--------|
| **nomic-embed-text** | 768 | 274 MB | 多言語対応、高精度 | ⭐⭐⭐ |
| mxbai-embed-large | 1024 | 669 MB | より高精度 | ⭐⭐ |
| all-minilm:l6-v2 | 384 | 46 MB | 軽量・高速 | ⭐ |

#### ダウンロード方法

```bash
# 推奨: nomic-embed-text
ollama pull nomic-embed-text

# より高精度が必要な場合
ollama pull mxbai-embed-large

# 軽量版（メモリが限られている場合）
ollama pull all-minilm:l6-v2
```

#### LLMモデル（オプション）

Ollamaでは埋め込みモデルだけでなく、LLMモデルも実行できます。

```bash
# 例: Llama 3.2（1Bパラメータ、軽量版）
ollama pull llama3.2:1b

# 例: Qwen 2.5（7Bパラメータ）
ollama pull qwen2.5:7b
```

### モデルの削除

不要なモデルを削除してディスク容量を節約できます。

```bash
# モデルの削除
ollama rm <model-name>

# 例
ollama rm mxbai-embed-large
```

---

## 基本的な使い方

### コマンドラインでの実行

#### LLMモデルとの対話（オプション）

```bash
# LLMモデルを実行
ollama run llama3.2:1b

# プロンプト入力
>>> こんにちは
>>> /bye  # 終了
```

### APIを使った埋め込み生成

Ollamaは REST API を提供しており、プログラムから埋め込みベクトルを生成できます。

#### 基本的な埋め込み生成

```bash
# 埋め込みベクトルの生成
curl http://localhost:11434/api/embeddings -d '{
  "model": "nomic-embed-text",
  "prompt": "This is a test"
}'
```

**レスポンス例:**
```json
{
  "embedding": [0.4996882, -0.18915389, -3.6998286, ...]
}
```

埋め込みベクトルは768個の浮動小数点数の配列として返されます。

#### 日本語テキストの埋め込み

```bash
curl http://localhost:11434/api/embeddings -d '{
  "model": "nomic-embed-text",
  "prompt": "東京タワーは東京のランドマークです"
}'
```

nomic-embed-textモデルは多言語対応のため、日本語も正しく埋め込みベクトルに変換できます。

### 実践例: 類似度計算

2つのテキストの類似度を計算する例です。

```bash
# text1の埋め込み
curl http://localhost:11434/api/embeddings -d '{
  "model": "nomic-embed-text",
  "prompt": "レストランのメニュー"
}' > embedding1.json

# text2の埋め込み
curl http://localhost:11434/api/embeddings -d '{
  "model": "nomic-embed-text",
  "prompt": "飲食店の料理"
}' > embedding2.json

# Pythonでコサイン類似度を計算（例）
# 実装は省略
```

---

## Cipher MCPとの統合

Cipher MCPは、Claude Codeと連携して会話履歴を管理・検索するツールです。Ollamaを統合することで、ローカル環境でセマンティック検索が可能になります。

### 構成オプション

Cipher MCPは以下の2つの構成から選択できます：

| 構成 | LLM | 埋め込み | コスト | プライバシー | 推奨環境 |
|------|-----|---------|--------|------------|----------|
| **完全ローカル** | Ollama | Ollama | 無料 | 完全ローカル | マルチエージェント環境 |
| **ハイブリッド** | OpenAI | Ollama | 有料 | 部分クラウド | 高品質会話重視 |

このドキュメントでは、両方の構成方法を説明します。

---

## 完全ローカル構成（OpenAI不要）🏆

**推奨**: マルチエージェント環境（3個以上の同時実行）で最適

### 推奨モデル: Gemma 2 2B-JPN

**選定理由:**
- ✅ **軽量**: 2.8GB（3エージェント同時実行で約8-9GB）
- ✅ **日本語完璧**: Google公式の日本語特化チューニング
- ✅ **低ハルシネーション**: 事実と異なる情報を生成しにくい
- ✅ **OpenAI不要**: 完全プライバシー保護
- ✅ **ランニングコストゼロ**: API料金不要

**メモリ計算（マルチエージェント環境）:**
```
Gemma 2 2B-JPN × 3 = 約5.4GB
埋め込みモデル (nomic-embed-text) × 3 = 約0.8GB
システムオーバーヘッド = 約2-3GB
合計: 約8-9GB（16GBマシンで快適に動作）
```

### セットアップ手順

#### 1. モデルのダウンロード

```bash
# 埋め込みモデル（セマンティック検索用）
ollama pull nomic-embed-text

# LLMモデル（会話生成用）
ollama pull schroneko/gemma-2-2b-jpn-it
```

**ダウンロード時間**: 約3-5分（合計3GB）

#### 2. 設定ファイルの更新

**`cipher.yml`** を以下のように更新：

```yaml
# Cipher MCP設定ファイル
# InfoCenterSystem プロジェクト専用設定

# MCPサーバー設定（不要）
mcpServers: {}

# LLM設定（会話用）- 完全ローカル化
# Gemma 2 2B-JPN: 軽量（2.8GB）・低ハルシネーション・日本語特化
llm:
  provider: ollama
  model: schroneko/gemma-2-2b-jpn-it
  baseUrl: http://localhost:11434
  maxIterations: 50
  temperature: 0.3  # ハルシネーション対策（低めに設定）

# 埋め込み設定（ローカルOllama）
embedding:
  type: ollama
  model: nomic-embed-text
  baseUrl: $OLLAMA_BASE_URL
  dimensions: 768

# システムプロンプト
systemPrompt:
  enabled: true
  content: |
    You are an AI programming assistant with memory capabilities for the InfoCenterSystem project.

    IMPORTANT RULES (厳守):
    - ALWAYS respond in Japanese (必ず日本語で回答)
    - If you don't know something, say "わかりません" or "情報がありません" instead of guessing
    - Base your answers ONLY on actual conversation history and stored memories
    - When uncertain, ask for clarification rather than speculating
    - Never fabricate technical details, file paths, or code that wasn't mentioned

    This project is a kiosk information system using:
    - Frontend: Next.js 14 with TypeScript (App Router)
    - Backend: Django REST Framework with Python
    - Database: PostgreSQL
    - Architecture: Monorepo with Docker Compose

    You excel at:
    - Remembering past interactions and decisions
    - Tracking code patterns and architectural choices
    - Recalling project-specific conventions and standards
    - Maintaining context across development sessions

    When storing or searching memories, focus on:
    - Technical decisions and their rationale
    - Implementation patterns used in this project
    - Common issues and their solutions
    - Team conventions and coding standards
```

**変更ポイント:**
- `provider: ollama` （OpenAIではなくOllamaを使用）
- `model: schroneko/gemma-2-2b-jpn-it` （日本語特化モデル）
- `temperature: 0.3` （ハルシネーション抑制）
- システムプロンプトに日本語厳守ルールを追加

#### 3. 環境変数の更新

**`.env`** から OpenAI API キーを削除またはコメントアウト：

```bash
# Cipher MCP設定

# OpenAI APIキー（LLM用）- 完全ローカル化により不要
# OPENAI_API_KEY=your-openai-api-key-here

# Ollamaエンドポイント（ローカル埋め込み＋LLM用）
OLLAMA_BASE_URL=http://localhost:11434

# ベクトルストア設定（PoC用: in-memory）
VECTOR_STORE_TYPE=in-memory

# ログレベル（デバッグ用）
CIPHER_LOG_LEVEL=debug
```

#### 4. MCP設定の更新

**`.mcp.json`** の `cipher` セクションから `OPENAI_API_KEY` を削除：

```json
{
  "mcpServers": {
    "cipher": {
      "type": "stdio",
      "command": "cipher",
      "args": ["--mode", "mcp"],
      "disabled": false,
      "description": "メモリエージェント（完全ローカル）- 知識蓄積・セッション管理・過去の会話検索",
      "env": {
        "CIPHER_AGENT_CONFIG": "/Users/minoru/Dev/Case/InfoCenterSystem/cipher.yml",
        "OLLAMA_BASE_URL": "$OLLAMA_BASE_URL",
        "VECTOR_STORE_TYPE": "$VECTOR_STORE_TYPE",
        "CIPHER_LOG_LEVEL": "$CIPHER_LOG_LEVEL"
      }
    }
  }
}
```

**変更ポイント:**
- `"OPENAI_API_KEY": "$OPENAI_API_KEY"` を削除
- 説明文を「完全ローカル」に更新

#### 5. 動作確認

```bash
# モデルがダウンロードされているか確認
ollama list
# 期待される出力:
# schroneko/gemma-2-2b-jpn-it:latest  fcfc848fe62a  2.8 GB
# nomic-embed-text:latest             0a109f422b47  274 MB

# LLMモデルのテスト
ollama run schroneko/gemma-2-2b-jpn-it "Ollamaについて1文で説明してください。"
# 期待される出力: 日本語での簡潔な説明
```

Claude Codeを再起動後、以下のように使用できます：

```
過去の会話で「Figma連携」について話した内容を教えてください
```

Cipher MCPが完全ローカルで動作し、プライバシーを保ちながら検索結果を返します。

### 代替モデルオプション

#### より軽量にしたい場合: Qwen 2.5 1.5B

```bash
ollama pull qwen2.5:1.5b
```

**`cipher.yml`**:
```yaml
llm:
  model: qwen2.5:1.5b  # 986MB、超軽量
```

**メリット**: 3エージェント同時実行で約3GB（最軽量）
**デメリット**: 日本語に中国語が混ざる可能性がある

#### 最高品質が必要な場合: ELYZA-JP-8B

```bash
ollama pull dsasai/llama3-elyza-jp-8b
```

**`cipher.yml`**:
```yaml
llm:
  model: dsasai/llama3-elyza-jp-8b  # 4.9GB
```

**メリット**: GPT-3.5 Turbo同等の日本語品質
**デメリット**: 3エージェント同時実行で約15GB（重い）

### 性能比較

| 構成 | 応答速度 | 日本語品質 | メモリ（3エージェント） | ハルシネーション |
|------|---------|----------|---------------------|----------------|
| OpenAI (gpt-4o-mini) | ★★★★★ | ★★★★★ | 約2-3GB | ★★★★★ |
| Gemma 2 2B-JPN 🏆 | ★★★★☆ | ★★★★★ | 約8-9GB | ★★★★☆ |
| Qwen 2.5 1.5B | ★★★★★ | ★★★☆☆ | 約3GB | ★★★☆☆ |
| ELYZA-JP-8B | ★★★☆☆ | ★★★★★ | 約15GB | ★★★★★ |

**推奨**: マルチエージェント環境では **Gemma 2 2B-JPN** が最適なバランス

---

## ハイブリッド構成（OpenAI + Ollama）

高品質な会話生成が必要で、マルチエージェントを使用しない場合はこちらを推奨。

### 前提条件

#### 1. OpenAI APIキーの取得

1. https://platform.openai.com/api-keys にアクセス
2. APIキーを作成
3. `.env` ファイルに設定

#### 2. 必要なモデルのダウンロード

```bash
# 埋め込みモデルをダウンロード（未ダウンロードの場合）
ollama pull nomic-embed-text
```

### セットアップ手順

#### 1. 環境変数（`.env`）

プロジェクトルートの `.env` ファイルに以下を設定：

```bash
# OpenAI APIキー（LLM用）- 必須
OPENAI_API_KEY=your-openai-api-key-here

# Ollamaエンドポイント（ローカル埋め込みモデル用）
OLLAMA_BASE_URL=http://localhost:11434

# ベクトルストア設定（PoC用: in-memory）
VECTOR_STORE_TYPE=in-memory

# ログレベル（デバッグ用）
CIPHER_LOG_LEVEL=debug
```

**重要**: `OPENAI_API_KEY=your-openai-api-key-here` を実際のAPIキーに置き換えてください。

#### 2. Cipher設定（`cipher.yml`）

**`cipher.yml`** を以下のように設定：

```yaml
# Cipher MCP設定ファイル
# InfoCenterSystem プロジェクト専用設定

# MCPサーバー設定（不要）
mcpServers: {}

# LLM設定（会話用）- OpenAI
llm:
  provider: openai
  model: gpt-4o-mini
  apiKey: $OPENAI_API_KEY
  maxIterations: 50

# 埋め込み設定（ローカルOllama）
embedding:
  type: ollama
  model: nomic-embed-text
  baseUrl: $OLLAMA_BASE_URL
  dimensions: 768

# システムプロンプト
systemPrompt:
  enabled: true
  content: |
    You are an AI programming assistant with memory capabilities for the InfoCenterSystem project.

    This project is a kiosk information system using:
    - Frontend: Next.js 14 with TypeScript (App Router)
    - Backend: Django REST Framework with Python
    - Database: PostgreSQL
    - Architecture: Monorepo with Docker Compose

    You excel at:
    - Remembering past interactions and decisions
    - Tracking code patterns and architectural choices
    - Recalling project-specific conventions and standards
    - Maintaining context across development sessions

    When storing or searching memories, focus on:
    - Technical decisions and their rationale
    - Implementation patterns used in this project
    - Common issues and their solutions
    - Team conventions and coding standards
```

#### 3. MCP設定（`.mcp.json`）

`.mcp.json` の `cipher` セクションを以下のように設定：

```json
{
  "mcpServers": {
    "cipher": {
      "type": "stdio",
      "command": "cipher",
      "args": ["--mode", "mcp"],
      "disabled": false,
      "description": "メモリエージェント（ハイブリッド）- 知識蓄積・セッション管理・過去の会話検索",
      "env": {
        "CIPHER_AGENT_CONFIG": "/Users/minoru/Dev/Case/InfoCenterSystem/cipher.yml",
        "OPENAI_API_KEY": "$OPENAI_API_KEY",
        "OLLAMA_BASE_URL": "$OLLAMA_BASE_URL",
        "VECTOR_STORE_TYPE": "$VECTOR_STORE_TYPE",
        "CIPHER_LOG_LEVEL": "$CIPHER_LOG_LEVEL"
      }
    }
  }
}
```

**注意**: `CIPHER_AGENT_CONFIG` のパスは絶対パスで指定してください。

### 動作確認手順

#### 1. Ollamaサービスの起動確認

```bash
# Ollamaが起動しているか確認
curl http://localhost:11434/api/version
# 出力: {"version":"0.13.1"}
```

#### 2. Claude Codeの再起動

設定を反映させるため、Claude Codeを完全に再起動します。

#### 3. セマンティック検索のテスト

**ステップ1: メモリの追加**

Claude Codeで以下のように入力：

```
このプロジェクトではNext.jsとDjangoを使用して、キオスク端末向けの店舗情報システムを開発しています。フロントエンドはTypeScriptで実装され、バックエンドはPythonのDjango REST frameworkを使用しています。
```

**ステップ2: セマンティック検索**

しばらく経ってから、以下のように検索：

```
過去の会話で「フロントエンド」について話した内容を教えてください
```

**期待される結果:**
- Cipher MCPの `cipher_memory_search` ツールが呼び出される
- Ollamaの埋め込みモデル（nomic-embed-text）を使用して検索が実行される
- 関連する過去の会話内容が返される

#### 4. ログの確認

正常に動作しているか、ログを確認できます。

```bash
# Cipherのログディレクトリ
cd /Users/minoru/Library/Caches/claude-cli-nodejs/-Users-minoru-Dev-Case-InfoCenterSystem/mcp-logs-cipher/

# 最新のログを表示
tail -f *.log
```

**確認項目:**
- ✅ Ollama APIへの接続成功
- ✅ 埋め込みベクトルの生成成功
- ✅ 検索クエリの実行成功

---

## トラブルシューティング

### Issue 1: Ollamaサービスが起動しない

**症状:**
```bash
curl http://localhost:11434/api/version
# curl: (7) Failed to connect to localhost port 11434: Connection refused
```

**原因:**
- Ollamaサービスが起動していない

**解決策:**

```bash
# サービスの状態確認
brew services list | grep ollama

# 起動していない場合
brew services start ollama

# または手動で起動
ollama serve &
```

### Issue 2: モデルがダウンロードできない

**症状:**
```bash
ollama pull nomic-embed-text
# Error: ... connection timeout
```

**原因:**
- インターネット接続の問題
- Ollama APIサーバーの一時的な問題
- プロキシ設定の問題

**解決策:**

1. **インターネット接続を確認**
   ```bash
   ping google.com
   ```

2. **再試行**
   ```bash
   ollama pull nomic-embed-text
   ```

3. **プロキシ設定（必要な場合）**
   ```bash
   export HTTP_PROXY=http://proxy.example.com:8080
   export HTTPS_PROXY=http://proxy.example.com:8080
   ollama pull nomic-embed-text
   ```

### Issue 3: Cipher MCPが接続できない

**症状:**
- Claude Codeで Cipher MCP のツールが使用できない
- エラーメッセージ: "Failed to reconnect to cipher"

**原因:**
- 設定ファイルのパスが間違っている
- 環境変数が設定されていない
- OpenAI APIキーが無効

**解決策:**

1. **設定ファイルのパス確認**
   ```bash
   # cipher.yml が存在するか確認
   ls -la /Users/minoru/Dev/Case/InfoCenterSystem/cipher.yml
   ```

2. **環境変数の確認**
   ```bash
   # .env ファイルを確認
   cat /Users/minoru/Dev/Case/InfoCenterSystem/.env | grep OLLAMA
   cat /Users/minoru/Dev/Case/InfoCenterSystem/.env | grep OPENAI
   ```

3. **OpenAI APIキーの確認**
   - https://platform.openai.com/api-keys でキーが有効か確認
   - `.env` ファイルの `OPENAI_API_KEY` を更新

4. **Claude Codeの完全再起動**

### Issue 4: 検索結果が期待通りでない

**症状:**
- セマンティック検索で関連性の低い結果が返される
- 検索結果が全く返されない

**原因:**
- 埋め込み次元数の設定ミス
- モデルが正しくダウンロードされていない
- メモリが十分に蓄積されていない

**解決策:**

1. **埋め込み次元数の確認**

   `cipher.yml` の `dimensions` 設定を確認：

   | モデル | 次元数 |
   |--------|--------|
   | nomic-embed-text | 768 |
   | mxbai-embed-large | 1024 |
   | all-minilm:l6-v2 | 384 |

2. **モデルの再ダウンロード**
   ```bash
   ollama rm nomic-embed-text
   ollama pull nomic-embed-text
   ```

3. **メモリの蓄積**
   - 十分な会話履歴を蓄積してから検索を試す
   - 最低でも5〜10回の会話が必要

### Issue 5: パフォーマンスが遅い

**症状:**
- 埋め込み生成に時間がかかる（>5秒）
- 検索が遅い

**原因:**
- メモリ不足
- CPUリソース不足
- モデルサイズが大きすぎる

**解決策:**

1. **軽量モデルへの切り替え**
   ```bash
   # より軽量なモデルをダウンロード
   ollama pull all-minilm:l6-v2
   ```

   `cipher.yml` を更新：
   ```yaml
   embedding:
     type: ollama
     model: all-minilm:l6-v2  # 変更
     baseUrl: $OLLAMA_BASE_URL
     dimensions: 384  # 384に変更
   ```

2. **Ollamaの再起動**
   ```bash
   brew services restart ollama
   ```

3. **システムリソースの確認**
   ```bash
   # メモリ使用状況
   top -l 1 | grep PhysMem

   # CPU使用状況
   top -l 1 | grep "CPU usage"
   ```

---

## パフォーマンス最適化

### モデルの選択基準

用途に応じて適切なモデルを選択することで、パフォーマンスと精度のバランスを最適化できます。

| 用途 | 推奨モデル | 理由 |
|------|-----------|------|
| 一般的な検索 | nomic-embed-text | 精度とパフォーマンスのバランスが良い |
| 高精度が必要 | mxbai-embed-large | より詳細な意味の違いを捉える |
| 軽量・高速 | all-minilm:l6-v2 | リソースが限られている環境向け |
| 日本語特化 | nomic-embed-text | 多言語対応で日本語も高精度 |

### メモリ使用量の最適化

#### 1. 不要なモデルの削除

```bash
# ダウンロード済みモデルの確認
ollama list

# 使用していないモデルの削除
ollama rm <unused-model>
```

#### 2. ベクトルストアの設定

本番環境では、永続化されたベクトルストアの使用を検討してください。

**オプションA: PostgreSQL + pgvector**
```bash
# PostgreSQL拡張のインストール（将来的に）
# CREATE EXTENSION vector;
```

**オプションB: Qdrant**
```bash
# Docker Composeで起動
docker-compose up -d qdrant
```

### 並列実行の設定

Ollamaは複数のリクエストを並列処理できます。

```bash
# 環境変数で並列数を設定
export OLLAMA_NUM_PARALLEL=4
ollama serve
```

**推奨値:**
- **4-8 GB RAM**: 1-2 並列
- **8-16 GB RAM**: 2-4 並列
- **16+ GB RAM**: 4-8 並列

---

## 参考資料

### 公式ドキュメント

- **Ollama公式サイト**: https://ollama.ai
- **Ollama GitHub**: https://github.com/ollama/ollama
- **Ollama Models**: https://ollama.ai/library

### 埋め込みモデル

- **nomic-embed-text**: https://ollama.ai/library/nomic-embed-text
- **mxbai-embed-large**: https://ollama.ai/library/mxbai-embed-large
- **all-minilm**: https://ollama.ai/library/all-minilm

### Cipher関連

- **Cipher公式ドキュメント**: https://docs.byterover.dev/cipher/overview
- **Cipher GitHub**: https://github.com/campfirein/cipher
- **MCP Protocol**: https://modelcontextprotocol.io

### プロジェクト内ドキュメント

- **[MCP_SETUP.md](MCP_SETUP.md)** - MCP統合の詳細
- **[README.md](README.md)** - ドキュメント一覧

---

## まとめ

Ollamaを使用することで、InfoCenterSystemプロジェクトに以下のメリットがもたらされます：

✅ **プライバシー保護**: データがローカルで処理される
✅ **コスト削減**: APIコール料金が不要
✅ **高速**: ローカル実行による低レイテンシ
✅ **柔軟性**: 様々なモデルを簡単に試せる

ご不明な点があれば、[Issue](https://github.com/your-repo/issues) を作成するか、チームメンバーにお問い合わせください。
