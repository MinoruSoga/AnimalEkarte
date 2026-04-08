# Frontend + Backend コード規約準拠監査結果

**実施日**: 2026-04-09（第5回監査まで反映）
**検証方法**: 監査エージェント → grep/Read による実コード検証
**注意**: 全指摘は実コードで検証済み。スポットチェック未実施の箇所は明記している。

## 確認済みイシュー一覧

| BUG | 対象 | 内容 | 優先度 | パス |
|-----|------|------|--------|------|
| [BUG-186](../security/BUG-186_jwt-token-leaked-in-login-response.md) | 認証 | JWT がレスポンスボディに漏洩 | Critical | security/ |
| [BUG-187](../security/BUG-187_demo-credentials-no-env-check.md) | 認証 | デモ認証情報が環境チェックなし | High | security/ |
| [BUG-188](BUG-188_frontend-catch-onError-handleApiError-missing.md) | FE エラー処理 | catch 9箇所 + onError 17箇所で handleApiError 未呼び出し | High | code-quality/ |
| [BUG-189](BUG-189_backend-handler-service-violations.md) | BE Handler/Service | RefreshToken c.JSON 6箇所 + slog 11箇所 + 裸 return err 2箇所 | High | code-quality/ |
| [BUG-190](BUG-190_design-token-violations-verified.md) | FE デザイントークン | ハードコード色（スポットチェック確認済み箇所のみ） | Medium | code-quality/ |
| [BUG-191](BUG-191_shifts-useActionState-migration.md) | FE React 19 | ShiftFormDialog が useActionState 未使用 | Medium | code-quality/ |
| [BUG-192](../infra/BUG-192_tsconfig-strict-false.md) | インフラ | tsconfig.json strict: false | High | infra/ |
| [BUG-193](BUG-193_db-schema-confirmed-issues.md) | DB スキーマ | payments の deleted_at 欠落 + billing_items の updated_at 欠落 + リポジトリ clinicID 欠落 2件 | High | code-quality/ |
| [BUG-194](BUG-194_vercel-react-best-practices-violations.md) | Vercel React BP | useDeferredValue 1件 + デザイントークン 19箇所/9ファイル | Medium | code-quality/ |
| BUG-221 | FE rerender-transitions | ✅ **CLOSED** — 全 7箇所修正済み（2026-04-09） | — | closed/ |
| [BUG-222](BUG-222_useCallback-object-deps.md) | FE rerender-dependencies | useCallback deps にオブジェクト/配列（9箇所/4ドメイン） | Medium | code-quality/ |
| [BUG-223](BUG-223_memo-default-nonprimitive-prop.md) | FE rerender-memo | DailyRecordSection の `plans = []` が memo を無効化 | Low | code-quality/ |
| [BUG-224](BUG-224_derived-state-in-useEffect.md) | FE derived-state | useEffect で derived state を同期（owners/inventory 各1件） | Low | code-quality/ |
| [BUG-225](BUG-225_StaffSelectionModal-not-lazy-loaded.md) | FE bundle | StaffSelectionModal が lazy() でロードされていない | Low | code-quality/ |
| [BUG-226](BUG-226_filter-map-double-iteration.md) | FE js-perf | `.filter().map()` による二重イテレーション（4箇所） | Low | code-quality/ |
| [BUG-227](BUG-227_static-selectitem-not-hoisted.md) | FE rendering-hoist | 静的 SelectItem JSX がモジュール定数に未巻き上げ（4箇所） | Low | code-quality/ |
| [BUG-228](BUG-228_trivial-useMemo.md) | FE rerender-memo | 軽量計算に useMemo を使用（OwnerSearchModal, MedicineSettings） | Low | code-quality/ |
| [BUG-229](BUG-229_defer-reads-useAuth-in-callbacks.md) | FE rerender-defer | useAuth の user をコールバック内のみで使用（BillingReviewSection, billing-review.ts） | Low | code-quality/ |
| [BUG-230](BUG-230_repeated-find-without-map.md) | FE js-index-maps | レンダーループ内で O(n) .find() を繰り返し（HospitalizationBoard, MedicineSettings） | Medium | code-quality/ |
| [BUG-231](BUG-231_functional-setstate-hospitalization-form.md) | FE rerender-functional-setstate | use-hospitalization-form.ts の addTreatmentPlan 等 3箇所が prev=> 未使用 | Medium | code-quality/ |
| [BUG-232](BUG-232_useEffect-object-deps.md) | FE rerender-dependencies | useEffect deps にオブジェクト（hospitalization + estimate 各1件） | Medium | code-quality/ |
| [BUG-233](BUG-233_content-visibility-unpaginated-lists.md) | FE rendering-perf | MedicineSettings/TreatmentPlanMaster の全件レンダーに content-visibility 未適用 | Medium | code-quality/ |
| [BUG-234](BUG-234_set-map-lookups.md) | FE js-set-map | O(n) .some()/.includes() を Set で O(1) に（PetSelection 等3箇所） | Low | code-quality/ |
| [BUG-235](BUG-235_TreatmentSearchDialog-not-lazy-in-CarePlanDialog.md) | FE bundle | CarePlanDialog 内の TreatmentSearchDialog が static import かつ常時マウント | Low | code-quality/ |
| [BUG-236](BUG-236_trivial-useMemo-daily-records-tab.md) | FE rerender-memo | DailyRecordsTab の getTodayStr() を useMemo で不要メモ化 | Low | code-quality/ |

## 誤報として撤回した指摘

| 当初の指摘 | 撤回理由 |
|-----------|---------|
| Deep Import 25箇所+（auth/hooks/use-permission） | **grep で 0 件。** 全ファイルが正しくバレル経由 |
| router.tsx Deep Import 6箇所 | **grep で 0 件。** 正しくバレル経由 |
| chart.tsx CSS インジェクション | **color 値はアプリ内定数。** ユーザー入力ではない |
| pet_service.go ビルドエラー | **正しいシグネチャで呼び出し済み。** コンパイルエラーなし |
| catch ブロック handleApiError 欠落 15箇所 | **実際は 9箇所。** 大幅に過大評価（catch の 95% は正しく呼んでいた） |
| リポジトリ 50+メソッドで clinicID 欠落 | **FK チェーン間接分離を「欠落」と誤認。** 確認済みは 2件のみ |

## 対応除外

| 指摘 | 理由 |
|------|------|
| /uploads 認証なし | `TASK-S3-IMAGE-UPLOAD` で S3 移行対応中 |

## 修正優先順位（BUG-221〜223 追加後）

1. **BUG-186** (Critical): LoginResponse から Token フィールド削除 — 5分
2. **BUG-187** (High): デモアカウントを `import.meta.env.DEV` でゲート — 5分
3. **BUG-188** (High): handleApiError 一括適用 26箇所 — 1時間
4. **BUG-189** (High): RefreshToken の RespondError 統一 + slog 移動 — 2時間
5. **BUG-192** (High): tsconfig strict 段階的有効化 — 半日
6. **BUG-193** (High): billing_items updated_at 追加 — 30分
7. **BUG-190** (Medium): デザイントークン置換 — 要個別確認
8. **BUG-191** (Medium): ShiftFormDialog useActionState 移行 — 1時間
9. ~~**BUG-221**~~ ✅ CLOSED（全 7箇所修正済み）
10. **BUG-222** (Medium): useCallback deps を primitive に — 2時間（9箇所）
11. **BUG-223** (Low): DailyRecordSection EMPTY_PLANS 定数化 — 5分
12. **BUG-224** (Low): derived state を useEffect で同期 → レンダー中に直接導出 — 30分
13. **BUG-225** (Low): StaffSelectionModal を lazy() に変更 — 10分
14. **BUG-226** (Low): filter().map() → flatMap/reduce に統一 — 30分（4箇所）
15. **BUG-227** (Low): 静的 SelectItem をモジュール定数に巻き上げ — 30分（4ファイル）
16. **BUG-228** (Low): trivial useMemo 削除 — 10分（2箇所）
17. **BUG-229** (Low): useAuth → ref パターンに変更 — 20分（2ファイル）
18. **BUG-230** (Medium): .find() → Map に変更（HospitalizationBoard 優先） — 30分
19. **BUG-231** (Medium): functional setState — use-hospitalization-form.ts 3箇所 — 20分
20. **BUG-232** (Medium): useEffect deps オブジェクト → id (primitive) に変更 — 20分
21. **BUG-233** (Medium): content-visibility: auto 適用（MedicineSettings, TreatmentPlanMaster） — 30分
22. **BUG-234** (Low): .some()/.includes() → Set.has() に変更（3箇所） — 20分
23. **BUG-235** (Low): TreatmentSearchDialog を lazy + 条件レンダーに変更 — 15分
24. **BUG-236** (Low): DailyRecordsTab の trivial useMemo 削除 — 1分
