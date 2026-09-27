# 一時バグ／障害メモ（ユーザー依頼 2026-09-13）

> Task migrated to Plane `EMR-180`. This file remains supporting acceptance/evidence material; use Plane for current status.

> **注意**: root `bug.md` は 2026-09-08 に `todo.md` の [製品 FAIL](todo.md#product-bugs) へ統合廃止済み。  
> 本ファイルはユーザー明示依頼により **一旦** 再作成したメモ台帳。製品 FAIL 正本は引き続き `todo.md#product-bugs`。  
> OPEN 項目は2026-09-23にPlaneへ移行済み。現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)とPlaneを参照する。
> 2026-09-26: `bug-2.md`（全ページ CRUD UAT メモ・2026-09-13）を本ファイルへ統合し、元ファイルは削除。
> 2026-09-26 棚卸し: 全項目を Plane・マージ済みPR（#467-481, #498-509）・現行コードで再調査し、対応済み項目（Done / FIXED / SPEC-OK）は削除。残存は未解決 2 件のみ。

更新日: 2026-09-26

## 索引

| ID | status | area | severity | 種別 | 修正プラン |
|:---|:---|:---|:---|:---|:---|
| PO-PET-DECEASED-DATA-BACKFILL | Plane EMR-108 · **Blocked** | data / pet | Medium | **PO確認**（不整合データの修復方針） | 現在の課題・状態は `EMR-108`（詳細は移行記録）。修復方針の PO 決定待ち |
| NOTE2-SWEEP-COVERAGE | Plane EMR-200 · **Blocked** | test coverage | - | **調査**（route×operation 受入範囲の補完） | 現在の課題・状態は `EMR-200`（詳細は移行記録）。下記に UAT 証拠を保持 |

---

### PO-PET-DECEASED-DATA-BACKFILL: `status=deceased` かつ `deceased_at` null のペット大量不整合

- **実測**: pets（未削除）のうち  
  - `status=deceased` かつ `deceased_at` null: **5836**（城東だけでも **4699**）  
  - `status=deceased` かつ `deceased_at` あり: 6075  
  - `deceased_at` ありで status≠deceased: 0
- **問い（PO）**:
  1. 旧DB／handoff 由来の不整合を **`deceased_at` を埋めて死亡確定**するのか、**status を alive に戻す**のか
  2. 修復は STG のみか、本番 cutover 前の必須か
  3. `BUG-RES-DECEASED-STATUS-BYPASS` の恒久対策（status もガードに含めるか、SoT を deceased_at のみに揃えて FE も追従か）のどちらを採るか
- **注意**: credential／PHI を台帳に載せない。個体例は検証用 ID のみ
- **関連（解消済み）**: `BUG-RES-DECEASED-STATUS-BYPASS` の write ガードは `status=deceased OR deceased_at != null` に統一済み（`backend/internal/sharedkernel/pet_not_deceased.go`・PR マージ済み）。本項目は**データ修復方針のみ**が未決。

---

## NOTE2-SWEEP-COVERAGE — Plane `EMR-200`

> Current task details are in Plane. The earlier route and API findings remain as evidence.

### 進捗・ページ結果サマリ（2026-09-13 全ページ CRUD UAT）

- 証拠: `reports/uat-2026-09-13/all-pages/`
  - ルート巡回: `all-pages-20260913-033603.json`（**82** ページ、clinic_id=2）
  - CRUD深掘り: `crud-deep-*.json` / `crud-retry-*.json` / `crud-final-*.json` / `crud-verify.json`
- 対象医院: **城東センター病院 (2)**（デモ執行・メインは八王子だが X-Clinic-ID/localStorage で 2）
- **ルート巡回**: 82 / 展開可能ルートすべて **画面到達 PASS**（認証破綻なし）
- **コンソール（ノイズ除外）**: `/reservations` で DialogContent Description 欠落警告のみ（新規ダイアログオープン時）— BUG2-RES-DIALOG-A11Y として対応済み
- **HTTP≥400（ページ sweep）**: 実質なし（login の一時 `/me` は除外）
- 動的詳細で ID 未取得のためスキップ: `['/accounting/:id', '/hospitalization/:id', '/hospitalization/:id/edit', '/inventory/:id']`
- 方針: HydrateFallback は既知ノイズとして非起票

#### CRUD 成立したもの（城東 clinic_id=2・執行デモ）

- 職種 / ケージ / 保険 / 予約区分 / 薬品 / 物販 / 問診テンプレ / 権限グループ / 診断タイプ / 健診タイプ / キャンペーン / シフトテンプレ
- 飼主・ペット CUD、見積 CRUD、ワクチン CD、トリミング Create、在庫 CRUD

#### CRUD で権限・契約上ブロック（製品欠陥としない／要PO）

- **動物種 create**: 403（system admin 限定の既存仕様）
- **支払方法 create**: 403（view 可・create 不可 — BUG2-PAYMETHOD-CREATE-FORBIDDEN として SPEC-OK 確定済み）
- **健診 POST `/api/v1/checkups`**: ルートなし（作成はカルテ経由 `create-checkup-medical-record` → `/v1/medical-records`）

### NOTE2-SWEEP-COVERAGE 詳細

- API CRUD: 主要マスタ／飼主ペット／見積／ワクチン／トリミング／在庫まで確認
- カルテ新規は BUG2-MR-ENTERED-BY-CLINIC で当時ブロック（同件は修正済み。UI 新規も同経路）
- 入院は `cage_id` 必須など契約が厳しく、ケージ指定付きの追加確認は未完了（契約バリデーションとしては正常応答）
- 検査 create は examination-types 取得後も追加フィールド不足で 400 — 追加切り分け余地あり（本ラウンドでは製品断定せず）
