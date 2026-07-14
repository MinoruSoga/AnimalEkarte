# BE-refactor.md — バックエンド リファクタリング計画書

- **第7期**: **完了**（BE7-0〜BE7-21 全22項目、2026-07-14 棚卸しで裏取り済み）。実行手順・完了条件・コミット指定は削除済み。詳細は git 履歴（`565c8708`〜`a6cbdc70`、進捗同期 `5d5600f1`）を参照。
- **本書の役割**: 次期監査への引き継ぎのみ。新規の第7期作業はない。
- **別台帳**: PERF・SEED残・既知バグ skip 台帳（#236 で修正予定） = `BE_todo.md` / 任意検証 = `BE-pending.md`

---

## 次期監査への引き継ぎ・PO 判断待ち

- **[決定済み 2026-07-14] cron 未配線3本（FEAT-377/379）**: 「配線」「削除」の二択は誤りだったと調査で判明し、**現状維持＋配線条件の明文化**で確定。
  - **配線禁止の根拠**: Write API は noop でも `applyTagState`（`lstep_tag_sync_api.go`）が**ローカルタグキャッシュを更新**し、配線済みの delivery trigger バッチ（毎時）が `FindOwnerIDsByTag(PREV_フィラリア未完了 等)` でそのキャッシュを条件に**実 LINE 配信を発火**する。つまり配線＝FEAT-379 配信トリガー3種の本番アクティベーションであり、Write 一時停止（2026-05 のマーケ施策停止）と矛盾する。エンジニアリング判断で配線してはならない。
  - **削除しない根拠**: 3本は孤立死コードではなく、全層現役のエコシステム（delivery trigger 配線済み・タグコードマッピング UI・健診予防閾値設定・タグキャッシュ基盤）の自動計算層。バッチだけ削除すると製品に機能が見えたまま計算だけ消える不整合になる。廃止するなら FE 設定・トリガー・マッピングを含む FEAT-377/379 全体を畳む PO 決定（マーケ施策の廃止）が先。
  - **確定した運用**: Lステップ Write 再有効化の手順に「3本を `cmd/api/main.go` に配線」を組み込む（memory `lstep-write-pause-20260515` に追記済み）。**Sunset 条項**: 2026-10 の棚卸し時点で Write 再開が未定のままなら、バッチ単体ではなく FEAT-377/379 エコシステム全体の廃止可否を PO に提起する。
- **[設計判断] accounting service の repository 型パススルー**（`GetMonthlyUnpaidCarryover` 等4メソッド）: service DTO を挟むか現状維持か。
- **[要精査] `middleware/liff_auth.go:56,116`**: `LineCustomerRepository.FindOrCreateByLineUserID` を middleware から直接呼んでいる（層違反の疑い。構造監査では lookup インターフェース経由とも見える — 両者の実態を精査）。
- **[再検討] `checkStaffActive` の fail-open**（DB 一時障害時に認証を通す）: 文書化済みの意図的挙動だが、過去の CRITICAL fail-open 修正クラスと同型。権限系レビュー時に再検討。
- **[次期] apicontract のフィールドレベルゲート**: BE7-17 クラスの api.yaml 乖離（ShiftEntry 型不一致等）を機械検出する拡張。再発防止として価値が高い。
- **[次期] god-function 走査**: `examination_service.go` ReplaceItems（101行）/ medicine・reservation・cash_register の80行台関数は第7期未精査。
- **[次期] `lstep_csv_import_service.go` が自前 `s.db`（gorm 直 import）を持つ妥当性**。
- **[次期] `reservation_type_handler.go` の weekly/specific 相互依存バリデーションの service 移動**（LOW）。
- **[次期] `AuthService`/`TokenService`（BE7-20/21 で抽出済み）と `validators_auth.go` の統合、P8/P11 準拠の付与**: 第7期は「移すだけ」に限定した。
- **[次期] `AuditService.Log` の interface 露出整理**: 同ファイル内 `LogXxx` 群から内部利用されており削除不可。露出範囲の設計は要精査。

---

## 第7期で確定した「やらない」判断（次期でも踏襲推奨）

- **#236 の既知バグ**（クロステナント Staff 削除 / ClinicSettings 列名不一致 / IsNotFound 恒偽）は Issue 側で管理。skip テストの解除も Issue クローズと同時に行う。
- duration リテラル（`24*time.Hour` 等）の一括定数化・`INTERVAL '365 days'` SQL 3箇所の共通化はしない — 「同一値」でも意味の同一性が保証できず、機械統合は誤りを生む。
- `internal/middleware/response.go` の `respondError` と handler `RespondError` の統一はしない（X-17 で意図的見送り済み）。
- 犬猫種別判定ヘルパ `isDog/isCatSpeciesName`（部分一致・マーケティングタグ用途）と `doseSpeciesAliases`（完全一致・投薬安全の fail-closed）は**契約が意図的に異なるため統合禁止**。
- **[記録] 旧監査数値の訂正**: 「死にモック30」は現存せず（全数走査で未使用0）。「モック6重複」の実態は5クラスタ15型（BE7-15 で解消済み）。
