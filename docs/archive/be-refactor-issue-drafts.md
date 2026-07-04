# GitHub Issue 起票ドラフト（未作成・PO/architect承認後に起票）

> **本ファイルは下書きのみ**。`gh issue create` は本タスクで一度も実行していない。
> 起票する場合はPO/architectの承認を得たうえで、下記本文を元に手動または`gh issue create`で作成すること。
> 各ドラフトの根拠は `docs/be-refactor-followup-status.md` のPhase A該当節を参照。

---

## Draft 1: RLS full実効化設計（由来: Phase A-2a/A-2b/A-2c）

**Title**: `[設計] RLS full実効化 — 非ownerロール + FORCE + GUC配線設計`

**Body**:
```
## 背景
現状 `ekarte_user` は rolsuper=true rolbypassrls=true で接続しており、001_init.sqlのRLSは
ENABLEのみでFORCEなし。superuser接続はFORCE ROW LEVEL SECURITYすらbypassするため、
RLSは実質dormant（実効的なテナント境界はapp層のclinic_id述語が担っている）。

## 必要な変更
1. 非ownerアプリロールの新設（CREATE ROLE ... NOLOGIN + GRANT SELECT/INSERT/UPDATE/DELETE個別付与）
2. 全clinicテーブルへの `FORCE ROW LEVEL SECURITY` 適用
3. 全27 tx開始点（ambient起点1 + dbOrTx参加可能4 + ignore-ambient21 + service1）への
   `SET LOCAL app.current_clinic_ids` 配線
4. system_adminセッションのDBロール/GUC設計（bypass_rls=on にするか、
   resolveAllClinicIDs()の許可集合をapp.current_clinic_idsに設定するか）
5. batchバイナリ（cmd/migrate, cmd/lstep-migrate, cmd/seed-old-db, cmd/stage-import）の
   ロール設計（BYPASSRLS専用ロールを分離するか検討）

## リスク
all-or-nothing。FORCE適用を配線なしで行うとアプリ全クエリが遮断され機能停止する。
接続層・transactor・clinic-contextミドルウェアを跨ぐ高リスク改修。

## Acceptance Criteria
- [ ] 非superuserロールでの全経路動作確認（E2E含む）
- [ ] rls_effectiveness_test.go 相当のテストがCIで恒久実行される
- [ ] system_adminセッションの挙動が明示的にテストされる
- [ ] STG環境での段階的ロールアウト計画

## 参照
docs/be-refactor-followup-status.md の Phase A-2a/A-2b/A-2c 節
```

---

## Draft 2: batchエントリポイントのlocalガード強化（由来: Phase A-4・security-reviewer指摘）

**Title**: `[security] seed-old-db/stage-importのホスト名ガードをSSHトンネル耐性化 + confirmフラグ非対称解消`

**Body**:
```
## 背景
Phase A-4（batch/bypass経路監査）でsecurity-reviewerから2件のMEDIUM指摘。

## 問題
1. `cmd/seed-old-db`/`cmd/stage-import` のlocalガードはホスト名文字列完全一致
   （`db`/`localhost`/`127.0.0.1`）のみ。SSH/kubectl port-forward経由でSTG/本番を
   localhostへトンネリングする運用では無力化される。typo防止には十分だが、
   意図的/不注意な誤接続防止としては不十分。
2. `cmd/stage-import` には `--confirm-local-destroy` の2要素確認があるが、
   同じく破壊的な `cmd/seed-old-db` には対応する確認フラグが無く非対称。

## 提案
- 両ツールに対称的な確認フラグを追加（seed-old-dbにも--confirm相当を追加）
- ホスト名文字列一致だけでなく、実際の接続先解決（DNS逆引き等）や
  環境変数の追加検証を検討

## Acceptance Criteria
- [ ] seed-old-dbに確認フラグ追加 + テスト
- [ ] 両ツールのガードロジックが対称になっていることのテスト

## 参照
docs/be-refactor-followup-status.md の Phase A-4 節（security-reviewer指摘 MEDIUM 1,2）
```

---

## Draft 3: date-only response wire統一（由来: Phase A-1）

**Title**: `[FE/PO確認] date-only response統一の個別対応3系統`

**Body**:
```
## 背景
Phase A-1（date-only FE影響インベントリ）で、response side 22箇所のdate-only driftを
datetime→date-only へ統一する場合の影響調査を実施。18キー中14キーは既存transform層が
防御済みで低リスクだが、3系統は個別対応が必要。

## 個別対応が必要な3系統
1. **examination.date**（features/medical-records/api/get-record-examinations.ts:29）:
   `.slice(0,16).replace("T"," ")` で時刻まで意図的に表示（例: 09:30）。date-only化すると
   時刻表示が失われる臨床上の実害を伴う回帰。
2. **inventory.expiry_date / last_restocked**（features/inventory/api/inventory.ts:29,31、
   InventoryList.tsx:228）: 変換なしで生ISO文字列をそのままテーブル描画（既存表示バグ、
   date-only化はむしろ改善方向だが正しいフォーマッタ通過は別修正が必要）。
3. **estimate.valid_until**（EstimateForm.tsx:105 / EstimateDetail.tsx:108 /
   EstimateList.tsx:117,200）: 4箇所個別`.slice(0,10)`。移行は安全だが正規化ポイントが
   分散し保守性が低い。

## 調査ギャップ（未確定）
owner.last_visit / treatment CRUD直接画面のtreatment.date / checkups.ts:28のdateは、
決定論的grepの時間予算内では消費箇所を確定できなかった。着手前に追加調査が必要。

## Acceptance Criteria
- [ ] FE 3系統の個別対応完了（examination.date時刻表示の代替案含む）
- [ ] 調査ギャップ3箇所の消費箇所確定
- [ ] R3-3 gate（openapi-date-format-drift）のallowlist更新
- [ ] BE側datetime→date-only wire変更はFE対応完了後に実施
- [ ] PO承認（臨床表示影響のため必須）

## 参照
docs/be-refactor-followup-status.md の Phase A-1 節
```
