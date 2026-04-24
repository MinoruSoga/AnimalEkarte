# BUG-377: 医院マスタ新規作成 — Frontend/Backend 権限モデル不整合 + 403 サイレント失敗

**作成日**: 2026-04-15
**Status**: OPEN
**Priority**: HIGH (Security: 権限モデル破綻 + UX: サイレント失敗)
**Affects**: `features/hospital-settings`, `backend/internal/handler/clinic_handler.go`, `frontend/src/lib/handle-api-error.ts`
**発見経緯**: BUG-374 以降のマスタ系ブラウザテスト (Section 14 / `/settings/clinic`)

---

## 概要

`/settings/clinic` の「+ 新規登録」で医院作成を試みると HTTP 403 が返るが、UI は無反応（サイドパネルは開いたまま、トーストなし）。

原因は Frontend と Backend の権限モデル不整合：

- **Frontend**: `canCreate = usePermission(ResourceHospitalSettings).canCreate` で 新規登録ボタンを表示 (`ClinicMasterSettings.tsx:146, 303-309`)
- **Backend**:
  - ルーティング層: `RequirePermission(ResourceHospitalSettings, "create")` は PASS (`handler.go:228`)
  - ハンドラ層: 追加で `if !isSystemAdmin { 403 }` でブロック (`clinic_handler.go:165-173`)
- **Result**: `hospital-settings.can_create=true` だが `is_system_admin=false` のユーザーはボタンを押せるが作成できない。
- さらに `handle-api-error.ts:23-26` が 403 を「UI 側で制御済み」として**トースト抑制**するため、ユーザーは失敗に気づかない。

---

## 再現手順

1. `admin@example.com` (安田 希恵, `is_system_admin=false`, 「執行」権限グループ所属) でログイン
2. `/settings/clinic` に遷移
3. 「+ 新規登録」ボタンが表示される（本来は見えるべきでない）
4. クリック → サイドパネル表示
5. 院名「テスト医院」を入力して「保存」
6. **実際**: `POST /api/v1/clinics` が **403** を返す、サイドパネルは閉じない、トーストなし
7. **期待**:
   - (A) ボタン自体が非表示、または
   - (B) 403 時に「権限がありません」トースト表示 + サイドパネル閉じ

---

## 影響範囲

### CREATE / DELETE に同じパターン

`clinic_handler.go:165-173` (`CreateClinic`) と `clinic_handler.go:202-210` (`DeleteClinic`) に同一の `is_system_admin` ガードが存在。DELETE も同じ症状を持つ可能性（未検証）。

### UPDATE は別ロジック

`clinic_handler.go:115-161` (`UpdateClinic`) は `is_system_admin=false` でも自クリニックのみ更新可能 (line 124-133)。CREATE/DELETE とは挙動が異なる。

### 403 サイレント抑制の波及

`handle-api-error.ts:23-26` は **全 feature の catch ブロック**に波及するため、「権限がないのに UI が許可している」バグが他マスタにもあれば同様にサイレント失敗する。

---

## 根本原因分析

1. **権限モデルが resource x action を超える粒度で壊れている**
   - `permission_group_rules` 表は `(resource, can_create, can_edit, can_delete, can_view)` まで。
   - Backend はさらに「`is_system_admin=true` でないと叩けない」という**隠れたゲート**を持っている。
   - Frontend はこの隠れゲートを知らないため `canCreate=true` で GO と判断する。

2. **403 トースト抑制のリスクモデル破綻**
   - 設計前提: "UI 側で完全にゲートされている → 403 は来るはずがない → 万一来てもバグなので抑制"
   - 実態: UI 側ゲートが不完全なケースで、ユーザーは何のフィードバックも得ずに失敗する。

---

## 修正方針 (優先順位付き)

### Option A (推奨): Backend の二重ガード廃止 — CRITICAL セキュリティ審査対象

`RequirePermission` ミドルウェアに集約し、`is_system_admin` チェックを handler から除去する。
かわりに「医院作成」は特別権限として以下いずれか：

- (A-1) `ResourceClinicCreation` などの独立 resource を追加、`permission_group_rules` で管理し、`is_system_admin` スタッフにのみ付与する seed を書く。
- (A-2) `ResourceHospitalSettings` の `can_create=true` を付与するのは system admin の permission group のみに限定（seed/マスタ編集 UI で制御）。

どちらも「Frontend の `canCreate` だけで UI 制御が完結する」状態を作る。**権限モデルを一次ソースにする。**

### Option B (最小修正): Frontend が `is_system_admin` を追加チェック

`ClinicMasterSettings.tsx` の `canCreate` を

```tsx
const { canCreate: permCreate } = usePermission(ResourceHospitalSettings);
const { user } = useAuth(); // isSystemAdmin を取得
const canCreate = permCreate && user.isSystemAdmin;
```

に変更し、DELETE ボタンも同様に絞る。

**デメリット**: 隠れゲートを Frontend にも複製するだけで、将来別のリソースで同じ不整合が起きる。

### Option C (UX 救済 — B と独立): 403 サイレント抑制を撤回または緩和

`handle-api-error.ts:23-26` を以下に改修：

```ts
} else if (status === 403) {
  toast.error(serverMessage ?? `${context}の権限がありません。`);
  return;
}
```

「UI ゲートで制御済みならトーストは冗長」は正論だが、**制御漏れに対する安全網がない**のが問題。トースト 1 行分の冗長性より、サイレント失敗のリスクの方が大きい。

---

## 推奨実装

- **Option A-2 + Option C** の組合せ:
  1. `permission_groups` seed を見直し、「執行」グループから `hospital-settings.can_create` / `can_delete` を外す（system_admin 限定グループにのみ残す）
  2. Backend `clinic_handler.go` の `is_system_admin` 明示チェック削除（`RequirePermission` ミドルウェアに一本化）
  3. `handle-api-error.ts` の 403 抑制を撤回し、サーバーメッセージを表示

---

## 確認事項

- [ ] admin@noavet.jp (`is_system_admin=true`) では 新規登録・削除が正常動作することを確認 (clinic Create/Delete 両方)
- [ ] 「執行」グループの他マスタ（hospital-settings 以外）でも CREATE/DELETE が正常に動作することを確認（副作用なし）
- [ ] 他リソースに同種の `is_system_admin` 二重ガードが潜んでいないか grep 調査 (`extractIsSystemAdmin` を handler 直接使用している箇所)

---

## 既知の関連箇所 grep 結果

```
backend/internal/handler/clinic_handler.go:166  CreateClinic の is_system_admin ガード
backend/internal/handler/clinic_handler.go:203  DeleteClinic の is_system_admin ガード
backend/internal/handler/clinic_handler.go:91-103  GetClinic の自クリニック制限
backend/internal/handler/clinic_handler.go:124-133  UpdateClinic の自クリニック制限
```

`extractIsSystemAdmin` を使っている他ハンドラがあれば同様のパターンがある可能性。
