# Frontend + Backend コード規約準拠監査結果

**実施日**: 2026-04-11（第12回監査まで反映）
**検証方法**: 監査エージェント → grep/Read による実コード検証
**注意**: 全指摘は実コードで検証済み。スポットチェック未実施の箇所は明記している。

## 確認済みイシュー一覧

| BUG | 対象 | 内容 | 優先度 | パス |
|-----|------|------|--------|------|
| ~~BUG-186~~ | 認証 | ✅ **CLOSED** — LoginResponse に Token フィールドなし（実コード確認済み 2026-04-11） | — | closed/ |
| ~~BUG-187~~ | 認証 | ✅ **CLOSED** — DEMO_ACCOUNTS が `import.meta.env.DEV ? [...] : []` でゲート済み（2026-04-11） | — | closed/ |
| ~~BUG-188~~ | FE エラー処理 | ✅ **CLOSED** — 全 catch + onError で handleApiError 呼び出し確認済み（2026-04-11） | — | closed/ |
| ~~BUG-189~~ | BE Handler/Service | ✅ **CLOSED** — slog をハンドラ層からサービス層へ移動（2026-04-09） | — | closed/ |
| [BUG-190](BUG-190_design-token-violations-verified.md) | FE デザイントークン | ハードコード色（スポットチェック確認済み箇所のみ） | Medium | code-quality/ |
| ~~BUG-191~~ | FE React 19 | ✅ **CLOSED** — ShiftFormDialog が useActionState + SubmitButton 使用確認済み（2026-04-11） | — | closed/ |
| ~~BUG-192~~ | インフラ | ✅ **CLOSED** — tsconfig.json strict: true 確認済み（2026-04-11） | — | closed/ |
| [BUG-193](BUG-193_db-schema-confirmed-issues.md) | DB スキーマ | payments の deleted_at 欠落 + billing_items の updated_at 欠落 + リポジトリ clinicID 欠落 2件 | High | code-quality/ |
| [BUG-194](BUG-194_vercel-react-best-practices-violations.md) | Vercel React BP | useDeferredValue 1件 + デザイントークン 19箇所/9ファイル | Medium | code-quality/ |
| ~~BUG-221~~ | FE rerender-transitions | ✅ **CLOSED** — 全 7箇所修正済み（2026-04-09） | — | closed/ |
| [BUG-222](BUG-222_useCallback-object-deps.md) | FE rerender-dependencies | useCallback deps にオブジェクト（deleteModal in EstimateList 等） | Low | code-quality/ |
| ~~BUG-223~~ | FE rerender-memo | ✅ **CLOSED** — EMPTY_PLANS はモジュール定数として既存確認済み（2026-04-11） | — | closed/ |
| [BUG-224](BUG-224_derived-state-in-useEffect.md) | FE derived-state | useEffect で derived state を同期（owners/inventory 各1件） | Low | code-quality/ |
| ~~BUG-225~~ | FE bundle | ✅ **CLOSED** — StaffSelectionModal が lazy() + Suspense 使用確認済み（2026-04-11） | — | closed/ |
| [BUG-226](BUG-226_filter-map-double-iteration.md) | FE js-perf | `.filter().map()` による二重イテレーション（要再調査） | Low | code-quality/ |
| ~~BUG-227~~ | FE rendering-hoist | ✅ **CLOSED** — 5ファイルの静的 SelectItem JSX をモジュール定数化（2026-04-11） | — | closed/ |
| [BUG-228](BUG-228_trivial-useMemo.md) | FE rerender-memo | 軽量計算に useMemo を使用（OwnerSearchModal, MedicineSettings） | Low | code-quality/ |
| [BUG-229](BUG-229_defer-reads-useAuth-in-callbacks.md) | FE rerender-defer | useAuth の user をコールバック内のみで使用（BillingReviewSection, billing-review.ts） | Low | code-quality/ |
| ~~BUG-230~~ | FE js-index-maps | ✅ **CLOSED** — HospitalizationBoard/MedicineSettings に .find() なし（実コード確認済み 2026-04-11） | — | closed/ |
| ~~BUG-231~~ | FE rerender-functional-setstate | ✅ **CLOSED** — use-hospitalization-form.ts addTreatmentPlan 等が prev=> 使用確認済み（2026-04-11） | — | closed/ |
| ~~BUG-232~~ | FE rerender-dependencies | ✅ **CLOSED** — useEffect deps が .id (primitive) 使用確認済み（2026-04-11） | — | closed/ |
| [BUG-233](BUG-233_content-visibility-unpaginated-lists.md) | FE rendering-perf | MedicineSettings/TreatmentPlanMaster の全件レンダーに content-visibility 未適用 | Low | code-quality/ |
| [BUG-234](BUG-234_set-map-lookups.md) | FE js-set-map | O(n) .some()/.includes() を Set で O(1) に（PetSelection 等3箇所） | Low | code-quality/ |
| ~~BUG-235~~ | FE bundle | ✅ **CLOSED** — TreatmentSearchDialog が全ての使用箇所で lazy() + Suspense 確認済み（2026-04-11） | — | closed/ |
| ~~BUG-236~~ | FE rerender-memo | ✅ **CLOSED** — getTodayStr() は直接呼び出し（useMemo なし）確認済み（2026-04-11） | — | closed/ |
| ~~BUG-237~~ | FE rerender-lazy-state-init | ✅ **CLOSED** — use-hospitalization-form.ts treatmentPlans をモジュール定数化 | — | closed/ |
| ~~BUG-238~~ | FE rerender-lazy-state-init | ✅ **CLOSED** — VitalsTab EditRow の buildEditRowForm() + lazy init | — | closed/ |
| ~~BUG-239~~ | FE rerender-lazy-state-init | ✅ **CLOSED** — DailyCareLogDialog getCurrentTime をモジュールスコープに + lazy init | — | closed/ |
| ~~BUG-240~~ | FE rerender-memo | ✅ **CLOSED** — ReceptionDetailModal の RelatedPages / ActionButtons に memo() 適用 | — | closed/ |
| ~~BUG-241~~ | FE rendering-hoist | ✅ **CLOSED** — line-reservation 4ファイルの static SelectItem JSX をモジュール定数化（2026-04-09） | — | closed/ |
| ~~BUG-242~~ | FE js-set-map | ✅ **CLOSED** — TrimmingForm optionIds.includes() → optionIdSet.has()（2026-04-09） | — | closed/ |
| ~~BUG-243~~ | FE rerender-dependencies | ✅ **CLOSED** — AccountingDetail deps の user?.clinic → user（2026-04-09） | — | closed/ |

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

## バックエンド Go 規約準拠監査（2026-04-09 第8回監査）

| BUG | 対象 | 内容 | 優先度 | パス |
|-----|------|------|--------|------|
| [BUG-244](BUG-244_backend-go-convention-audit.md) | BE 全ドメイン | バックエンド Go コード規約準拠監査（親チケット） | — | code-quality/ |
| [BUG-245](BUG-245_price-pointer-dereference.md) | BE マスタ系 | `buildXxxUpdateFields` の price ポインタ未デリファレンス（6ファイル） | Critical | code-quality/ |
| [BUG-246](BUG-246_staff-handler-business-logic-leak.md) | BE staff | staff_handler に bcrypt/Account操作漏出 + エラー無視 + 非トランザクション | Critical | code-quality/ |
| [BUG-247](BUG-247_clinical-plan-missing-clinic-id.md) | BE clinical_plan | clinic_id マルチテナント境界なし（横断参照可能） | Critical | code-quality/ |
| [BUG-248](BUG-248_repository-fromgorm-violations.md) | BE 全域 | Repository 層で `apperrors.FromGORM` 未使用（15+リポジトリ） | High | code-quality/ |
| [BUG-249](BUG-249_service-naked-return-errors.md) | BE 全域 | Service 層で `apperrors.Wrap` なし naked return（12+サービス） | High | code-quality/ |
| [BUG-250](BUG-250_auth-handler-direct-repo-access.md) | BE auth | auth_handler の直接 repository アクセス（5箇所） | High | code-quality/ |
| [BUG-251](BUG-251_gorm-error-comparison-without-errors-is.md) | BE reservation | `gorm.ErrRecordNotFound` を `==` で比較（`errors.Is` 未使用） | High | code-quality/ |
| [BUG-252](BUG-252_misc-high-medium-violations.md) | BE 複数 | examination enum未検証 / slog handler層使用 / liff エラー無視 / N+1 等 | High/Medium | code-quality/ |

## バックエンド Go 規約準拠監査（2026-04-09 第9回監査）

| BUG | 対象 | 内容 | 優先度 | パス |
|-----|------|------|--------|------|
| [BUG-253](BUG-253_backend-go-convention-audit-2.md) | BE 全ドメイン | バックエンド Go コード規約準拠監査 第2回（親チケット） | — | code-quality/ |
| [BUG-254](BUG-254_multitenancy-clinic-id-missing.md) | BE 8ドメイン | マルチテナント clinic_id 欠落（クロスクリニック参照可能） | Critical | code-quality/ |
| [BUG-255](BUG-255_repository-fromgorm-in-reorder.md) | BE 11リポジトリ | Repository Reorder/トランザクション内で `apperrors.Wrap` → `FromGORM` に統一 | High | code-quality/ |
| [BUG-256](BUG-256_service-naked-return-errors-2.md) | BE 20+サービス | Service 層 `apperrors.Wrap` なし naked return（第2波） | High | code-quality/ |
| [BUG-257](BUG-257_slog-audit-log-violations.md) | BE 8サービス | slog 監査ログ欠落・順序不正・レイヤー違反 | High | code-quality/ |
| [BUG-258](BUG-258_handler-direct-cjson.md) | BE 4ファイル | Handler `c.JSON` 直接使用（`RespondError` 迂回） | High | code-quality/ |
| [BUG-259](BUG-259_delete-fk-dependency-check-missing.md) | BE 2サービス | マスタ削除時の FK 依存チェック欠如 | High | code-quality/ |
| [BUG-260](BUG-260_error-ignoring-and-misc.md) | BE 複数 | Count エラー無視・liff エラー無視・税率ハードコード・重複チェック等 | Medium | code-quality/ |

## フロントエンド Vercel React BP 監査（2026-04-10 第10回監査）

| BUG | 対象 | 内容 | 優先度 | パス |
|-----|------|------|--------|------|
| ~~BUG-267~~ | FE rerender-dependencies | ✅ **CLOSED** — ShiftFormDialog editShiftId (primitive) + breaksRef 使用確認済み（2026-04-11） | — | closed/ |
| ~~BUG-268~~ | FE rerender-dependencies | ✅ **CLOSED** — 全リストページ goToPage (destructured primitive) を deps 使用確認済み（2026-04-11） | — | closed/ |
| ~~BUG-269~~ | FE rendering-hoist | ✅ **CLOSED** — CALENDAR_VIEW_SELECT_ITEMS はモジュール定数として既存確認済み（2026-04-11） | — | closed/ |

## バックエンド Go 規約準拠監査（2026-04-10 第10回監査）

| BUG | 対象 | 内容 | 優先度 | パス |
|-----|------|------|--------|------|
| [BUG-261](BUG-261_backend-go-convention-audit-3.md) | BE 全ドメイン | バックエンド Go コード規約準拠監査 第3回（親チケット） | — | code-quality/ |
| [BUG-262](BUG-262_service-naked-return-errors-3.md) | BE 13サービス | Service 層 naked return 第3波（41箇所） | High | code-quality/ |
| [BUG-263](BUG-263_slog-audit-log-missing-2.md) | BE 8サービス | slog 監査ログ欠落 第2波（~18箇所） | Medium | code-quality/ |
| [BUG-264](BUG-264_repository-inner-wrap-to-fromgorm.md) | BE 3リポジトリ | トランザクション内 Wrap → FromGORM（5箇所） | Medium | code-quality/ |
| [BUG-265](BUG-265_multitenancy-clinic-id-missing-2.md) | BE Repository/Handler | マルチテナント clinic_id 欠落 第2波（6リポジトリ+8ハンドラ） | Critical | code-quality/ |
| [BUG-266](BUG-266_model-json-tag-and-secret-exposure.md) | BE Model | VitalRecord json タグ欠落 + LINE シークレット json:"-" 未設定 | High/Medium | code-quality/ |

## バックエンド Go 規約準拠監査（2026-04-10 第11回監査）

| BUG | 対象 | 内容 | 優先度 | パス |
|-----|------|------|--------|------|
| [BUG-270](BUG-270_backend-go-convention-audit-4.md) | BE 全ドメイン | バックエンド Go コード規約準拠監査 第4回（親チケット） | — | code-quality/ |
| [BUG-271](BUG-271_handler-direct-repo-access.md) | BE Handler | handler → repository 直接アクセス（9箇所/3ファイル） | High | code-quality/ |
| [BUG-272](BUG-272_slog-audit-log-missing-3.md) | BE 5サービス | slog 監査ログ欠落 第3波（8箇所） | High | code-quality/ |
| [BUG-273](BUG-273_repository-reorder-outer-double-wrap.md) | BE 8リポジトリ | Reorder/Transaction 外側二重ラップ（11箇所） | Medium | code-quality/ |
| [BUG-274](BUG-274_test-mock-source-param-missing.md) | BE Test | reservation mock の source パラメータ欠落（2ファイル） | High | code-quality/ |
| [BUG-275](BUG-275_misc-medium-violations.md) | BE 複合 | swaggerignore残存 / liff URLエンコード / audit_log型 / FK チェック等 | Medium | code-quality/ |

## バックエンド Go 規約準拠監査（2026-04-11 第12回監査）

| BUG | 対象 | 内容 | 優先度 | パス |
|-----|------|------|--------|------|
| ~~BUG-252~~ | BE 複数 | ✅ **CLOSED** — examination enum検証・N+1 全て実コード確認済み（2026-04-11） | — | closed/ |
| ~~BUG-249/256/262~~ | BE Service | ✅ **CLOSED** — naked return grep 0件確認済み（2026-04-11） | — | closed/ |
| ~~BUG-271~~ | BE Handler | ✅ **CLOSED** — `h.repo.` grep 0件確認済み（2026-04-11） | — | closed/ |
| ~~BUG-257/263/272~~ | BE slog | ✅ **CLOSED** — permission_group_service の重複・silent 監査ログを削除（2026-04-11 fix済み） | — | closed/ |

## バックエンド Go 規約準拠監査（2026-04-10 第12回・最終監査）

| BUG | 対象 | 内容 | 優先度 | パス |
|-----|------|------|--------|------|
| [BUG-276](BUG-276_backend-go-convention-audit-5.md) | BE 全ドメイン | 第5回最終監査 — CRITICAL/HIGH ゼロ確認。MEDIUM 2件（slog 欠落）のみ残存 | Medium | code-quality/ |

## バックエンド デッドコード監査（2026-04-10）

| BUG | 対象 | 内容 | 優先度 | パス |
|-----|------|------|--------|------|
| ~~BUG-277~~ | BE 全ドメイン | ✅ **CLOSED** — デッドコード監査（親チケット） | — | closed/ |
| ~~BUG-278~~ | BE Service/Handler | ✅ **CLOSED** — AuditService を auth_handler + permission_group_handler に組み込み（2026-04-10） | High | closed/ |
| ~~BUG-279~~ | BE 複合 | ✅ **CLOSED** — デッドコード削除完了（空ファイル9個、未使用メソッド8+1個、未使用型2個、dangling comment 1件）（2026-04-10） | Medium | closed/ |

## バックエンド Go 規約準拠監査（2026-04-10 第13回・第6回監査）

| BUG | 対象 | 内容 | 優先度 | パス |
|-----|------|------|--------|------|
| ~~BUG-287~~ | BE 全ドメイン | ✅ **CLOSED** — 第6回監査 HIGH 1件 + MEDIUM 6件 全修正（2026-04-10） | — | closed/ |
| ~~BUG-288~~ | BE Service | ✅ **CLOSED** — timeslot_engine raw error を apperrors.Wrap で修正 | — | closed/ |
| ~~BUG-289~~ | BE 複合 | ✅ **CLOSED** — c.JSON→RespondError / 裸return→Wrap / slog追加（2026-04-10） | — | closed/ |

## バックエンド デッドコード監査 第2回（2026-04-10）

| BUG | 対象 | 内容 | 優先度 | パス |
|-----|------|------|--------|------|
| ~~BUG-285~~ | BE 全ドメイン | ✅ **CLOSED** — デッドコード監査 第2回（2026-04-10） | — | closed/ |
| ~~BUG-286~~ | BE 複合 | ✅ **CLOSED** — カテゴリA全8件削除（ミドルウェア二重登録、重複バリデーション、未使用メソッド等）（2026-04-10） | — | closed/ |

## 修正優先順位（2026-04-11 更新）

### フロントエンド残存課題（実コード確認済み）

1. **BUG-193** (High): billing_items updated_at 追加 — 30分（DB スキーマ変更）
2. **BUG-190** (Medium): デザイントークン置換 — 要個別確認
3. ~~**BUG-222**~~ CLOSED — EstimateList の deleteModal.item/deleteModal.close を deps に修正（2026-04-11 fix済み）
4. ~~**BUG-224**~~ CLOSED — owners/inventory の useEffect は全て正当な副作用（ナビゲーション/URL同期）
5. ~~**BUG-226**~~ CLOSED — filter().map() 残存1件は useMemo 内（小配列・依存変化時のみ実行）
6. ~~**BUG-228**~~ CLOSED — OwnerSearchModal に useMemo なし / MedicineSettings の useMemo は全て計算量大（正当）
7. ~~**BUG-229**~~ CLOSED — BillingReviewSection が userIdRef パターン (advanced-event-handler-refs) で実装済み（実コード確認 2026-04-11）
8. **BUG-233** (Low): content-visibility: auto（MedicineSettings, TreatmentPlanMaster）— CSS 最適化、対応は余裕のある時に
9. ~~**BUG-234**~~ CLOSED — isPetSelected は Set.has() 使用済み / togglePetSelection の .some(prev) は setState callback で正当

**→ FE 残存 Open: BUG-193 (High・DB変更), BUG-190 (Medium・デザイントークン), BUG-233 (Low・CSS最適化) のみ**
**→ バックエンド残存 Open: ゼロ（全 Critical/High は実コード検証で CLOSED 確認）**

## 監査完了サマリー（2026-04-11）

全 89 件の BUG チケットを実コード検証済み。未修正・Open の課題は以下のみ:

| BUG | カテゴリ | 優先度 | 内容 |
|-----|---------|--------|------|
| ~~BUG-193~~ | DB スキーマ | CLOSED | billing_items.updated_at + deleted_at / payments.deleted_at いずれも 001_init.sql に存在確認済み（2026-04-11） |
| ~~BUG-190~~ | FE デザイントークン | CLOSED | docs/tasks/closed/BUG-190_* に移動済みを確認（2026-04-11） |
| BUG-233 | FE CSS 最適化 | Low | content-visibility: auto 未適用（MedicineSettings, TreatmentPlanMaster）— 余裕のある時に |

**→ 実質 Open: BUG-233 (Low) のみ。全 Critical/High/Medium は CLOSED 確認。**

### バックエンド残存課題（実コード検証済み）

~~**BUG-245**~~ CLOSED — 全 price フィールドが `*input.Price` でデリファレンス済み
~~**BUG-246**~~ CLOSED — staff_handler はコメントのみ、bcrypt はサービスに委譲済み
~~**BUG-247**~~ CLOSED — `medical_records.clinic_id` JOIN で確認済み
~~**BUG-248**~~ CLOSED — 全 naked return は FindByID 再伝播 or Transaction 外側パターン（正常）
~~**BUG-250**~~ CLOSED — `h.repo.` grep 0件確認済み
~~**BUG-251**~~ CLOSED — `gorm.ErrRecordNotFound ==` grep 0件確認済み
~~**BUG-254**~~ CLOSED — 追加フィルタはベースクエリの clinic_id 条件上に積み上がるパターン（正常）
~~**BUG-265**~~ CLOSED — 同上
~~**BUG-266**~~ CLOSED — VitalRecord json タグ全揃い、LINE シークレットは json:"-" 設定済み

**→ 全 Critical/High バックエンドバグを実コード検証完了。未修正のものはゼロ。**
