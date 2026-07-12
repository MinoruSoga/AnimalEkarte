# FE-refactor.md — 残バックログ（第 4 期以降）

- **更新日**: 2026-07-12（HEAD `a7e4b7ce` に対し全残件をコード実測で再検証済み）
- **完了済み**: FE4-1〜18 / FE5-1〜4 / FE6-1〜9 / FE7-1〜3 / FE8-1〜4（M3 health-card 含む）。実装の詳細・コミットは `git log --grep='FE4-\|FE5-\|FE6-\|FE7-\|FE8-'` が正本
- **クローズ・PO 決定の記録**: 本書には残さない。正本はメモリ `fe_backlog_decision_pack_20260711.md`（ラベル分岐 Q3-A 現状維持 / iso-date・design-tokens YAGNI / XSS 該当なし / tygo widen 上流解決 等）
- **本書の規約**: 行動可能な未対応タスクのみを記載する

---

## 残件

### 1. query-keys registry 全面採用 + clinic-id ルートキー化（長期）

- **現状**: registry（`frontend/src/lib/query-keys.ts`）は accountings / masters / me のみ。inline `queryKey` は **187 ファイル**に残存（2026-07-12 実測）。
- **リスク**: 単一タブ切替は `reload` 依存で現状 SAFE。将来 SPA 切替すると clinic を含まないキーからクロステナントキャッシュ漏れが起き得る。
- **方針**: registry 拡張とキー先頭への `clinicId` 付与を**同一設計判断**で別チケット化する。

### 2. medical-records `Treatment` の #201 未追随（長期）

- **現状**: `medical-records/types` の `Treatment` に `dose_*` が無い。
- **方針**: #201 FE UI 実装チケットの一部として対応（本リファクタ単独では着手しない）。

### 3. `LinkOwner` の cross-clinic ownerID 書き込み検証（セキュリティ） — ✅ DONE (`44e35b3b`)

- **由来**: FE8 Independent Review。read 側は `43927aeb` で `Preload("Owner", "clinic_id = ? …")` 済み。
- **対応**: `LineCustomerService.LinkOwner`（`line_customer_service.go`）に `ownerRepo.FindByID(ctx, clinicID, *ownerID)` 事前検証を追加（`shared_file_service.go` 先例と同型）。`ownerID != nil` かつ他クリニック/不存在なら `apperrors.Wrap` で NotFound を保持したまま返し、`UpdateOwnerLink` へは到達しない。`ownerID == nil`（unlink）は検証スキップのまま現行挙動維持。DI: `NewLineCustomerService` が `repository.OwnerRepository` を受け取るよう変更（`service.go` 配線更新）。
- **テスト**: `line_customer_service_test.go` に他クリニック/不存在 owner ケースを追加し `UpdateOwnerLink` 未呼出を assert（RED→GREEN）。`line_customer_repository_test.go` の write 側コメントを service ガード追加後の事実に更新（repo 単体の Preload 防御テストは多層防御として残置）。

### 4. 低優先

| 項目 | 対応内容 |
|---|---|
| M2 `avatar_url` 三方向 drift | BE `MeResponse` に無く FE/openapi は期待（常に null・実害小）。BE に足すか FE/openapi から消すかを決めて drift を解消する。 |
| `meClinicInfoSchema` の広い `.default()` | BE リネームを黙って既定値へ落とす逆失敗モード。`.default()` を必要最小限に絞る、または契約テストで固定する。 |
| `NextRecommendedVisitDate` | `liff_service_health_card.go:23` で宣言のみ・未代入 = health-card で常に null（捏造禁止の Assumption）。プロダクト仕様の確定を経て算出ロジックを実装する。 |
| `liff_response.go` pet_id | `fmt.Sprintf`（`liff_response.go:263`）→ `strconv.FormatUint` への表記揃え（非ブロッキング）。 |
