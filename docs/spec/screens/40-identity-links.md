# 同一飼主・ペット連携 仕様書 (Identity Links)

## 概要

- **画面の目的**: 所属医院内の飼主・ペットを手動で identity group にリンク／unlink し、連携先ペットの最小治療履歴を参照する（Phase 1 workbench）。
- **URLパターン**: `/identity-links`（`operations-routes.tsx` に lazy 登録。サイドバー常設ナビは持たない）。
- **アクセス権限**:
  - `identity-links:view` — ページ表示。未所持時は `IdentityLinksPage` が `/` へ `Navigate`（`ResourceIdentityLinks` + `usePermission`）。
  - `identity-links:edit` — リンク／unlink 操作の表示。閲覧のみのときは status バナー「閲覧のみ（link/unlink 権限なし）」。
- **スコープ**: 所属医院内の手動リンクのみ。権限のない医院 ID はサーバ側で拒否（UI 説明文と一致）。

---

## 1. 画面構成

### 1.1 シェル

- `PageLayout` タイトル「同一飼主・ペット連携」。
- `resource={ResourceIdentityLinks}`。
- 内部実装は `IdentityLinksWorkbench`（同一ファイル内）。

### 1.2 飼主リンク

- 検索入力（placeholder「氏名・カナ・電話」）。`useDebouncedValue` 300ms 後に `searchOwnersForLink`。
- ヒット一覧をトグル選択（clinic_id + owner_id）。
- `canEdit` 時: 「飼主をリンク」（選択 2 件以上）、メンバーごとの unlink（直近作成の `ownerGroupId` 必須）。

### 1.3 ペットリンク

- 親の飼主 identity group（`ownerGroupId`）が必要である旨を表示。
- 検索入力（placeholder「ペット名・番号」）→ `searchPetsForLink`（300ms debounce）。
- `canEdit` 時: 「ペットをリンク」（ownerGroupId ありかつ選択 2 件以上）、メンバーごとの unlink（`petGroupId` 必須）。
- 選択ペットごとに「連携履歴」ボタン → `getLinkedTreatmentHistory`（view/edit どちらでも可）。結果は plain text /「（履歴なし）」。

### 1.4 空・エラー・pending

- 検索クエリ空: ヒット一覧をクリア（専用 empty コピーなし）。
- 失敗: `role="alert"` のインラインエラー。
- 閲覧のみ: `role="status"` バナー。
- ミューテーション／履歴取得中は `useTransition` の pending でボタン disable。

---

## 2. 主要フロー・制約

1. 飼主を 2 件以上選択 → 飼主グループ作成 → `ownerGroupId` 保持。
2. その group を親にしてペットを 2 件以上選択 → ペットグループ作成。
3. Phase 1 UI は group をメンバー lookup GET で再読込しない（作成直後の group id をセッション内 state で保持）。
4. クロス医院の無断リンクはサーバ拒否。fail-closed の権限デフォルト（新規医院テンプレートに identity-links を付けない）。

---

## 3. 臨床安全・アクセシビリティ

- 本画面は臨床記録の直接編集ではなく identity 運用。誤リンク防止のため **edit 必須** で create/unlink を隠す。
- 検索・履歴の失敗を空成功にしない（alert 表示）。
- セクションは `aria-label`（飼主リンク／ペットリンク）。status / alert ロールを使用。

---

## 4. 技術仕様

### 構成

| 要素 | 役割 |
|:---|:---|
| `IdentityLinksPage` | view ゲート + workbench 委譲 |
| `IdentityLinksWorkbench` | 検索・選択・link/unlink・履歴 |
| `usePermission` / `ResourceIdentityLinks` | view / edit |
| `useDebouncedValue` | 検索 300ms |
| `identity-links-api.ts` | HTTP クライアント |

### API連携（本画面が呼ぶもの）

| メソッド | エンドポイント | 用途 | 必須アクション |
|:---|:---|:---|:---|
| GET | `/api/v1/identity-links/owners/search` | 飼主検索 (`q`, `limit`) | `view` |
| GET | `/api/v1/identity-links/pets/search` | ペット検索 | `view` |
| POST | `/api/v1/identity-links/owner-groups` | 飼主グループ作成 | `edit` |
| DELETE | `/api/v1/identity-links/owner-groups/:groupId/members` | 飼主メンバー unlink | `edit` |
| POST | `/api/v1/identity-links/pet-groups` | ペットグループ作成（`owner_group_id` 必須） | `edit` |
| DELETE | `/api/v1/identity-links/pet-groups/:groupId/members` | ペットメンバー unlink | `edit` |
| GET | `/api/v1/identity-links/pets/:clinicId/:petId/treatment-history` | 連携治療履歴（`include_linked`, page, limit） | `view` |

BE には group GET / member POST 等の追加ルートがあるが、Phase 1 の `IdentityLinksPage` は未使用。

### 関連

- ルート: `frontend/src/app/routes/operations-routes.tsx`
- Feature: `frontend/src/features/identity-links/`
- 製品 leaf 数の正本は `route-inventory.test.tsx`（84 product pages）。本ファイルは **画面仕様インデックス上の 1 葉**であり、product leaf 数と「番号付き md ファイル数」は一致しない。

---
