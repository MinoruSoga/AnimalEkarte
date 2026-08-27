# AnimalEkarte バグ台帳（着手可能のみ · 2026-08-27 解析確定）

- 解析: コード根拠に基づき、判断待ちを **推奨仕様で ACTIONABLE 化**
- 対象 root: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte`
- main 参考 tip: `87f442956`（対応済みは本文から削除済み）
- 本番デプロイは別判断

---

## 対応状況

| 区分 | 件数 |
|------|------|
| **ACTIONABLE（実装可）** | **4** |
| DONE / 削除済み | 旧 BUG-001〜013, 015〜030 等（本文に載せない） |

---

### BUG-014（中）検査取込モーダルが取込不可の確定済み検査を選択候補に出す

- **状態**: **ACTIONABLE**（製品欠陥として確定）
- **根拠（コード）**:
  - `frontend/src/features/medical-records/components/ExaminationImportDialog.tsx`
  - `availableExams` は `medicalRecordId` 紐付けのみでフィルタし、**確定/編集不可ステータスを除外していない**（L38–42）
  - 取込は `updateExamination({ medical_record_id })` のため、BE が確定済みを 409 拒否すると UI 上は選べて失敗する
- **環境要因ではない**: フィルタ欠落は実装上の欠陥。並行検証で「全部確定済み」だったことは再現条件を強めただけ
- **修正方針**:
  1. 取込可能な検査だけを選択可能にする（確定済み・編集不可は候補から除外、または disabled + 理由表示）
  2. 可能なら一覧 API/クエリ側でも取込可能のみに絞る
  3. 回帰テスト: 確定済みが選べない / 未確定・未紐付けは選べる / 取込成功パス
- **受け入れ**:
  - [ ] 確定済み検査を選んで 409 になる操作経路が UI から消える（または明示 disabled）
  - [ ] 取込可能な検査は従来どおり取込できる
  - [ ] テスト追加・PASS
- **想定パス**: `frontend/src/features/medical-records/components/ExaminationImportDialog.*`（必要なら examinations hooks/API）

---

### BUG-031（高）スタッフの自己無効化（および最終システム管理者の無効化）をサーバーが拒否しない

- **状態**: **ACTIONABLE**（仕様を安全側に確定）
- **旧称**: 「要ユーザー判断: 管理者セルフ無効化」
- **根拠（コード）**:
  - `DELETE` は自己削除を拒否（`自分自身を削除することはできません`）
  - `Update` の `is_active=false` には **自己無効化ガードも最終管理者ガードもない**（`staff_service_core.go` Update）
- **確定仕様（推奨を採用）**:
  1. **自分自身**を `is_active=false` にできない
  2. **有効なシステム管理者が自分を含めて1人だけのとき**、その管理者を無効化できない
  3. 他スタッフの無効化・複数管理者いる場合の他者無効化は従来どおり可
  4. エラーは `InvalidInput` または `Conflict` で明示メッセージ（日本語）
- **受け入れ**:
  - [ ] 自己 `is_active=false` → 拒否・DB 不変
  - [ ] 唯一の有効 sysadmin の無効化 → 拒否
  - [ ] 他者無効化（条件充足時）→ 成功
  - [ ] 単体/統合テスト
- **想定パス**: `backend/internal/staff/**`（必要なら FE 無効化 UI のエラー表示）

---

### BUG-032（高）誤作成したレジ締めを業務上取り消す正規手段がない（id=8/9 型の復旧）

- **状態**: **ACTIONABLE**
- **背景**: 負の `actual_cash` 受理などコード欠陥は **修正済み**。しかし誤って作られた締め行は **app から reopen/void できず** 当日帯が `is_already_closed` のまま残る
- **根拠**: `cash_register_*` に運用者向け reopen/void の正規 API が無い（Create 後は Find 可能だが reopen 不可のテスト固定あり）
- **修正方針（製品）**:
  1. 権限付き **締め取消（void/reopen）** API を追加（監査ログ: 誰が・なぜ・元 id）
  2. 取消後は同一 clinic/date/period の preview が未締めとして正規締め可能
  3. 二重締め・権限なし・存在しない id は fail-closed
  4. 可能なら管理 UI に取消導線（最低限 API + テストで可）
- **データ**: 既存 id=8/9 は本 BUG の受け入れ検証用に残してよい。取消 API で消せることを DoD に含めてよい（環境に行が無い場合はテスト fixture で代替）
- **受け入れ**:
  - [ ] 権限者のみ void/reopen 可能
  - [ ] 取消後に同一 period を再度 Close できる
  - [ ] 監査フィールドまたはログが残る
  - [ ] テスト PASS
- **想定パス**: `backend/internal/billing/cash_register_*`（+ 必要なら FE レジ締め画面）

---

### BUG-010（低）ローカル LIFF 実トークン検証ができない（検証経路の不足）

- **状態**: **ACTIONABLE**（環境制約を「検証経路の整備」タスクに変換）
- **事実**: ローカル既定は `VITE_LIFF_MOCK=true` 等で実 LINE トークン検証不可（製品ロジックというより検証ハーネス）
- **修正方針（どちらか一方で可 · 推奨は A）**:
  - **A（推奨）**: 実トークン検証を **staging 手順 + 秘密情報は env のみ** として `docs` に固定し、ローカルは mock E2E を公式の受け入れ経路と明記。CI は mock のみ
  - **B**: オプションで real-token ジョブを staging secret 付きに追加（失敗時は skip せず Needs Human）
- **受け入れ**:
  - [ ] 「ローカル mock で何を保証し、staging で何を保証するか」が docs に1ページで書かれている
  - [ ] mock 経路の自動テストまたは手順が実行可能
  - [ ] real-token をローカル必須にしない
- **想定パス**: `docs/ops/testing/**` または LIFF 関連 README / CI 設定

---

## 実装時の注意

- claim / 専用 worktree 推奨。`bug.md` の Done 遷移や本番デプロイは人間
- 秘密・本番 DB 直操作を agent に任せる場合は、**BUG-032 の正規 API** を優先（生 SQL は人間）
- agent-loop 利用時は本ファイルを ledger にし、上から BUG-014 → 031 → 032 → 010 の順を推奨

---

## 解析メモ（短）

| 旧状態 | 確定 |
|--------|------|
| BUG-014 製品 vs 環境 | **製品**（フィルタ欠落） |
| 管理者セルフ無効化 仕様未決 | **安全側仕様で実装**（自己不可・最終 sysadmin 不可） |
| 締め id=8/9 データ判断 | **取消 API 実装**で復旧可能にする（データだけの放置をやめる） |
| BUG-010 環境だから不可 | **検証経路ドキュメント/ハーネス**として着手可能 |
