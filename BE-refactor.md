# BE-refactor.md — バックエンド リファクタリング計画書

- **第7期**: **完了**（BE7-0〜BE7-21 全22項目、2026-07-14 棚卸しで裏取り済み）。詳細は git 履歴（`565c8708`〜`a6cbdc70`、進捗同期 `5d5600f1`）を参照。
- **本書の役割**: 次期監査への引き継ぎのみ。新規の第7期作業はない。
- **別台帳**: 今期着手可能な BE 残 = `BE_todo.md` / 任意検証・次期送り = `BE-pending.md`
- **更新日**: 2026-07-15（#236 CLOSED 事実に同期）

---

## 次期監査への引き継ぎ・PO 判断待ち

- **[決定済み 2026-07-14] cron 未配線3本（FEAT-377/379）**: 「配線」「削除」の二択は誤りだったと調査で判明し、**現状維持＋配線条件の明文化**で確定。
  - Write API は noop でも `applyTagState` がローカルタグキャッシュを更新し、配線済み delivery trigger が実 LINE 配信を発火する → エンジニアリング判断で配線してはならない。
  - 3本は孤立死コードではなく FEAT-377/379 エコシステムの自動計算層。廃止するなら FE 設定・トリガー・マッピングを含む全体を畳む PO 決定が先。
  - **確定運用**: Lステップ Write 再有効化の手順に「3本を `cmd/api/main.go` に配線」を組み込む。**Sunset**: 2026-10 の棚卸し時点で Write 再開が未定なら FEAT-377/379 全体の廃止可否を PO に提起。
- **[設計判断] accounting service の repository 型パススルー**（`GetMonthlyUnpaidCarryover` 等4メソッド）: service DTO を挟むか現状維持か。
- **[要精査] `middleware/liff_auth.go:56,116`**: `LineCustomerRepository.FindOrCreateByLineUserID` を middleware から直接呼んでいる（層違反の疑い）。
- **[再検討] `checkStaffActive` の fail-open**（DB 一時障害時に認証を通す）: 文書化済みの意図的挙動だが、権限系レビュー時に再検討。
- **[次期] apicontract のフィールドレベルゲート**: BE7-17 クラスの api.yaml 乖離を機械検出する拡張。
- **[次期] god-function 走査**: `examination_service.go` ReplaceItems（101行）/ medicine・reservation・cash_register の80行台。
- **[次期] `lstep_csv_import_service.go` が自前 `s.db`（gorm 直 import）を持つ妥当性**。
- **[次期] `reservation_type_handler.go` の weekly/specific 相互依存バリデーションの service 移動**（LOW）。
- **[次期] `AuthService`/`TokenService`（BE7-20/21 で抽出済み）と `validators_auth.go` の統合、P8/P11 準拠の付与**。
- **[次期] `AuditService.Log` の interface 露出整理**。

---

## 第7期で確定した「やらない」判断（次期でも踏襲推奨）

- duration リテラル（`24*time.Hour` 等）の一括定数化・`INTERVAL '365 days'` SQL 共通化はしない。
- `internal/middleware/response.go` の `respondError` と handler `RespondError` の統一はしない（X-17 で意図的見送り済み）。
- 犬猫種別判定ヘルパ `isDog/isCatSpeciesName`（部分一致・マーケ）と `doseSpeciesAliases`（完全一致・投薬 fail-closed）は**契約が意図的に異なるため統合禁止**。
