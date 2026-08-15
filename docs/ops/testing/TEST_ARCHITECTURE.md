# テストアーキテクチャ (Test Architecture)

> **目的**: AnimalEkarte の検証層・正本・実行手段・記録先を一枚で定義する。  
> **読者**: QA・開発者・AI エージェント。  
> **タイミング**: 受入・回帰・テスト戦略変更の前。  
> **最新更新**: 2026-08-14

---

## 1. 設計原則

1. **正本は1つ**: 同じ検証を複数文書に重複定義しない。層ごとの正本表（§2）に従う。
2. **受入は scenarios/**: 納品前・大きなリリース前の「業務が通る」証明は `scenarios/` が正本。
3. **項目単位まで受入する**: フォーム受入（V シリーズ）はフォーム単位の C1〜C3 に加え、**各入力項目に F プロトコル**を適用する（[scenarios/FIELD-LEVEL-PROTOCOL.md](scenarios/FIELD-LEVEL-PROTOCOL.md)）。
4. **結果を正本に書かない**: シナリオ md は手順の正本。実行結果は `reports/uat-YYYY-MM-DD/` と root `todo.md` 受入バグ節のみ。
5. **自動化は回帰、受入はギャップ**: E2E は壊れやすい回帰の固定化。臨床安全・フォーム永続・LIFF 等のギャップは scenarios が担当。
6. **最終合否は人**: 製品 FAIL の起票は AI/実行者、破壊操作・実 LINE・staging 決裁は `todo-po.md`。

---

## 2. 層構造（正本マップ）

```text
┌─────────────────────────────────────────────────────────────────┐
│ L0  仕様正本                                                      │
│     docs/spec/  ·  product-philosophy · ADR · screen specs       │
└────────────────────────────┬────────────────────────────────────┘
                             │ 何が正しいか
┌────────────────────────────▼────────────────────────────────────┐
│ L1  単体 (Unit)                                                   │
│     BE: go test（domain package） / FE: Vitest + RTL              │
│     正本: 各 *test.go / *.test.ts · カバレッジ: coverage-policy    │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────┐
│ L2  API 統合 (Integration)                                        │
│     実 DB · 認可 · FK · clinic_id 隔離                             │
│     正本: CI Backend job · local make ci の inventory gate        │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────┐
│ L3  E2E 回帰 (Playwright)                                         │
│     画面遷移・スモーク・主要ハッピーパスの自動固定                   │
│     正本: frontend/e2e/ · E2E_TESTING_GUIDE.md                    │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────┐
│ L4  受入 (UAT)  ★ 納品前の合否層                                   │
│     業務フロー S01〜S13 + フォーム V01〜V05（項目単位 F プロトコル）│
│     正本: scenarios/ · 実行: browser-test / Playwright / 人手     │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────┐
│ L5  補完手動・監視                                                 │
│     SECTION_14（ドメイン重点）· canary / post-deploy 監視          │
└─────────────────────────────────────────────────────────────────┘
```

| 層 | 何を証明するか | 正本 | 実行者 |
|:--|:--|:--|:--|
| L1 | 関数・コンポーネントが仕様どおり | コード隣の test | CI / 開発 |
| L2 | HTTP+DB+認可が破綻しない | domain HTTP test · lint gate | CI / `make ci` |
| L3 | 主要画面が壊れない | `frontend/e2e/` | CI `e2e.yml` · `run-e2e.sh` |
| **L4** | **業務と全フォーム項目が通る** | **`scenarios/`** | **AI+ブラウザ / 人手** |
| L5 | 重点ドメインの探索・デプロイ後健全性 | SECTION_14 · canary | AI / 人手 |

詳細の歴史的「三層」記述は [INTEGRATION_TEST_PLAN.md](INTEGRATION_TEST_PLAN.md) を維持し、**受入層（L4）の定義は本書を優先**する。

---

## 3. 受入層（L4）の内訳

### 3.1 業務フロー（S シリーズ）

| ID 群 | 対象 | 深度 |
|:--|:--|:--|
| S01〜S13 | 臨床安全・LIFF・入院・会計境界・集計・identity-links 等 | シナリオ README の表 |

- 実行順制約: **S01 最初**、**S10 は S08 の後**、S13 は独立。
- E2E や SECTION_14 が覆う外来ハッピーパスの重複は最小限にし、**ギャップ担当**を維持する。

### 3.2 フォーム受入（V シリーズ）— 項目単位必須

| ID | ドメイン | フォーム数（棚卸） |
|:--|:--|:--|
| V01 | 臨床 | 18 |
| V02 | 会計・予約・受付・シフト・**在庫** | 12（旧 11 + inventory） |
| V03 | 飼主・ペット・スタッフ・権限・医院 | 7 |
| V04 | /settings マスタ | 30 |
| V05 | 認証・LINE/LIFF・Lステップ | 18 |
| **合計** | | **85**（在庫フォーム追加後） |

各フォームで実施する層:

| 記号 | 名前 | 内容 |
|:--|:--|:--|
| **C1** | 入力チェック | 必須空・形式違反・境界（フォーム単位の入口） |
| **C2** | 更新永続 | 保存 → 一覧/詳細 → F5 → 再オープン |
| **C3** | DB 整合 | FK 選択肢・一意制約・存在しない ID |
| **F\*** | **項目単位** | [FIELD-LEVEL-PROTOCOL.md](scenarios/FIELD-LEVEL-PROTOCOL.md) を**全入力項目**に適用 |
| **M** | 項目棚卸 | [FORM-FIELD-INVENTORY.md](scenarios/FORM-FIELD-INVENTORY.md) の行がカバー範囲 |

> **受入完了の定義（フォーム）**: 対象フォームの inventory 全項目について、適用可能な F チェックが PASS または N/A（理由付き）であり、C1〜C3 が PASS であること。代表 1 項目だけの C1 では完了とみなさない。

### 3.3 実行手段の優先順位

| 優先 | 手段 | 用途 |
|:--|:--|:--|
| 1 | **browser-test スキル + Chrome DevTools MCP**（`:9222`） | プロジェクト公式。scenarios / SECTION_14 |
| 2 | **Playwright 再現ラン**（`reports/uat-*/` 型スクリプト） | フル通し・再実施 |
| 3 | **Playwright MCP**（対話） | スコープ受入・項目単位の深掘り |
| 4 | 人手 | 実 LINE・破壊操作・最終合否 |

環境準備の正本: [UAT-ENV-SETUP.md](UAT-ENV-SETUP.md)。

---

## 4. 記録と不合格の扱い

| 成果物 | 置き場 | 書くこと |
|:--|:--|:--|
| 実行レポート | `reports/uat-YYYY-MM-DD/FINAL.md` | シナリオ別 PASS/PARTIAL/BLOCKED/FAIL 集計 |
| 機械可読結果 | 同ディレクトリ `results.json` | ステップ単位 |
| 証跡 | 同ディレクトリ `*.png` 等 | FAIL/PARTIAL の根拠 |
| **製品欠陥** | root `todo.md` §2 受入バグ | `### BUG-xxx` のみ Open を残す |
| 環境・仕様・決裁 | `todo-po.md` | 実 LINE・fixture・破壊 gate |
| シナリオ md | **編集禁止（結果を書かない）** | 手順・期待の正本のみ |

判定:

| 判定 | 意味 |
|:--|:--|
| PASS | 期待どおり |
| PARTIAL | 中核は通るが一部未確認 |
| BLOCKED | 環境・仕様・fixture 不足で実行不能（製品バグではない） |
| FAIL | 製品が期待を満たさない → `todo.md` 起票 |

---

## 5. 環境プロファイル

| プロファイル | DB seed | LIFF | 用途 |
|:--|:--|:--|:--|
| **local UAT** | `003_demo` | `LIFF_MOCK=true` / `VITE_LIFF_MOCK=true` | 通常の scenarios フル |
| **STG UAT** | `004_staging` | mock **禁止**（実 LINE） | 納品前最終・実機連携 |
| **CI E2E** | CI 用 | モック/限定 | L3 回帰のみ |

認証: `E2E_LOGIN_EMAIL` / `E2E_LOGIN_PASSWORD`（`.env.local`）。値を repo・chat・シナリオに書かない。ロールは admin / doctor / staff 等の**役割名**で指定する。

---

## 6. 変更時のスコープ指針

| 変更領域 | 最低限の L4 |
|:--|:--|
| 臨床・カルテ | S02,S03,S06 + V01（変更タブの**全項目 F**） |
| 会計・締め | S07–S09,S11 + V02 該当フォーム全項目 |
| 飼主・ペット | S01 + V03 該当フォーム全項目 |
| マスタ | V04 該当行の全項目 |
| LINE/LIFF | S04,S12 + V05 該当（local は mock） |
| 在庫 | V02 inventory 全項目 |
| 横断・権限 | S13 + 権限グループ項目 |

PR ではスコープ受入。**納品前は S 全本 + V 全本（項目単位）**。

---

## 7. 関連文書

| 文書 | 役割 |
|:--|:--|
| [scenarios/README.md](scenarios/README.md) | 受入シナリオ索引 |
| [scenarios/FIELD-LEVEL-PROTOCOL.md](scenarios/FIELD-LEVEL-PROTOCOL.md) | 項目単位チェック定義 |
| [scenarios/FORM-FIELD-INVENTORY.md](scenarios/FORM-FIELD-INVENTORY.md) | フォーム×項目の棚卸し |
| [UAT-ENV-SETUP.md](UAT-ENV-SETUP.md) | 受入実行環境の準備 |
| [E2E_TESTING_GUIDE.md](E2E_TESTING_GUIDE.md) | L3 Playwright |
| [SECTION_14_MANUAL_TEST_GUIDE.md](SECTION_14_MANUAL_TEST_GUIDE.md) | L5 ドメイン重点 |
| [INTEGRATION_TEST_PLAN.md](INTEGRATION_TEST_PLAN.md) | L1–L3・負荷の計画 |
| [../coverage-policy.md](../coverage-policy.md) | カバレッジ ratchet |
