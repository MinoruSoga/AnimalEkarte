# Architecture improvement backlog

| 項目 | 値 |
|------|-----|
| **作成** | 2026-08-07 |
| **範囲** | アーキテクチャのみ（UAT / credential / seed / 納品 ops は [`STATUS.md`](STATUS.md) / [`PO-todo.md`](PO-todo.md)） |
| **正本の前提** | [ADR-005](docs/architecture/adr/005-go-gin-backend-guidelines.md) · [ADR-006](docs/architecture/adr/006-backend-domain-package-boundaries.md) · [boundary map](docs/architecture/be9-2a-boundary-map.md) · [product-philosophy](docs/product-philosophy.md) |
| **現状** | BE9 domain-first **完了** · 旧 `internal/handler\|service\|repository` **削除済み** · agent 製品実装 open **0**（2026-08-06 時点） |

---

## 0. 前提（触らない土台）

既に正しい。作り直さない。

- [x] `internal/<domain>` vertical slice（domain/capability-first modular monolith）
- [x] 旧 layer package（`handler` / `service` / `repository`）の廃止
- [x] `cmd/api` 明示 composition root
- [x] write owner（例: `appointments` → `reservation`）と lint gate
- [x] consumer-side interface による import cycle 解消
- [x] clinic isolation / clinical safety を package 配置だけで証明しない（runtime test + invariant）

### やらないこと（アーキ改善に見せかけた退行）

- [ ] ~~Clean Architecture の folder 導入（`domain/` `application/` `infrastructure/`）~~ — **禁止**
- [ ] ~~`internal/handler|service|repository` の復活~~ — **禁止**
- [ ] ~~medicalrecord を entity 単位 18 package に機械分割~~ — **禁止**（ADR-006 却下済み粒度）
- [ ] ~~全 `internal/model` の一括 domain 分散~~ — **禁止**（漸進のみ可）
- [ ] ~~mock のためだけの interface 先創り~~ — **禁止**
- [ ] ~~behavior change / schema / API 破壊を「整理」に同梱~~ — **禁止**

---

## 1. 目標像

```text
同じ ADR-006 のまま:

  - 各 domain の「1 PR の自然な単位」が小さい
  - write owner と cross-domain tx 契約が全部説明できる
  - composition と model が再び god にならない
  - 許可依存グラフがテストで守られる
```

成功の測り方（アーキ）:

- 機能改修 PR の平均タッチファイル数が下がる（domain ごと）
- cross-domain 新規 path に「tx / fail-closed or best-effort+補償」が必ず書かれる
- composition / model への「なんでも追加」がレビューで弾ける
- import allowlist 違反が CI または scoped test で落ちる

---

## 2. バックログ（優先度順）

### ARCH-A1: `medicalrecord` 内凝集（最高）

| 項目 | 内容 |
|------|------|
| **ID** | ARCH-A1 |
| **優先度** | P0 |
| **状態** | **COMPLETE + landed on origin/main（2026-08-07）** · tip `bd7f9f677` |
| **対象** | `backend/internal/medicalrecord/**` · composition は原則触らない（S1） |
| **実測（計画）** | prod ~34k LOC / ~195 files · test ~88k · 最大: examination 841 / hospitalization 725 / treatment 678 等 |
| **方針** | B 削除 → C 同一 package 内 file 分割。**D（subpackage / Clean / 境界またぎ）不採用** |
| **計画ソース** | Grok Plan-only 実行結果（code-review-graph + grep/read + ADR-006） |

#### チェックリスト

- [x] **A1-0** グラフ・肥大 file・write path 調査（計画完了）
- [x] **A1-1** S1: BE9 残骸削除・thin path 収束 — `e7a2a53b4`
- [x] **A1-2a** S2: examination items 分割 — `fe3bed53f`
- [x] **A1-2b** S3: hospitalization discharge 分割 — `a4b830648`
- [x] **A1-2c** S4: treatment helpers 分割 — `c56ad84c3`
- [x] **A1-2d** S6: vital 分離 — `1e58c85d0`
- [x] **A1-2e** S5: medical_record appointment context — `bd7f9f677`
- [x] **A1-3** 間接 isolation: repository 述語は書き換えず service 分割のみ
- [x] **A1-4** appointment: reservation intent のみ維持（reservation diff empty + write-owner PASS）
- [x] **A1-5** composition 未変更（全スライス）
- [x] **A1-6** 1 スライス 1 commit · API/schema 不変 · scoped Docker test

#### 不変条件（全スライス）

- clinic isolation（直接・親経由）— JOIN/EXISTS/subquery を「整理」名目で書き換えない
- appointment 通常カルテ — reservation owner のみ。medicalrecord は Backfill/Prepare intent のみ
- fail-closed — tx/validator 欠落を成功にしない
- required audit は同 tx
- 公開 HTTP contract 不変（status / error code / JSON / RBAC）

#### スライス計画（確定）

```text
S1 ──► S2
 │
 ├──► S3（S1 後）
 ├──► S4（S1 後）
 ├──► S6（S1 後）
 └──► S5（S1 後・高リスクは後ろ）
S2 ∥ S3 ∥ S4 ∥ S6 は file 衝突少なら並列可
```

| Slice | 目的 | リスク | 主なパス | 状態 |
|-------|------|--------|----------|------|
| **S1** | BE9 残骸削除: legacy `NewMedicalRecordService`（test helper）除去 + `goSafe` thin wrapper 削除 | **LOW** | `go_safe.go` 削除、`checkup_service.go`→`sharedkernel.GoSafe`、test を `WithTxAudit` へ、`medical_record_legacy_constructor_test.go` 削除 | **done** `e7a2a53b4` |
| **S2** | `examination_service.go`（841）純粋分割（items/replaceItemsTx 等） | MEDIUM | → `examination_items.go` | **done** `fe3bed53f` |
| **S3** | `hospitalization_service.go` から Discharge 分離 | **HIGH** | → `hospitalization_discharge.go` | **done** `a4b830648` |
| **S4** | `treatment_service.go`（678）fields / master FK 分離 | MEDIUM | `treatment_fields.go` / `treatment_master_fk.go` | **done** `c56ad84c3` |
| **S5** | `medical_record_crud` の appointment context 可視化 | CRITICAL〜HIGH | → `medical_record_appointment_context.go` | **done** `bd7f9f677` |
| **S6** | `vital_service` の validation/audit 分離 | MEDIUM | `vital_validation.go` / `vital_audit.go` | **done** `1e58c85d0` |

#### S1 詳細（実装仕様）

| 項目 | 内容 |
|------|------|
| 削除 | `go_safe.go`（`sharedkernel.GoSafe` への 1 行 delegate） |
| 削除 | `medical_record_legacy_constructor_test.go` の `NewMedicalRecordService` helper |
| 変更 | `checkup_service.go` の `goSafe(...)` → `sharedkernel.GoSafe(...)` |
| 変更 | medicalrecord 配下 test の `NewMedicalRecordService(` → `NewMedicalRecordServiceWithTxAudit(`（audit 引数は既存 production 署名に合わせ `nil` 等） |
| 触らない | examination/hospitalization/treatment 本体、`medical_record_lock.go`、routes、`cmd/api` composition（既に `WithTxAudit`） |
| 検証 | `docker compose exec backend go test ./internal/medicalrecord/ -count=1 -run 'TestMedicalRecordService_'` および影響 test 名を必要最小追加 |
| ロールバック | 単一 commit revert |
| 完了条件 | production/test から legacy ctor 消滅 · `go_safe.go` 削除 · 公開 API 無変更を PR 説明に明記 · focused test green |

**S1 完了証跡（2026-08-07）:**

- Commit: `e7a2a53b4` — `refactor(medicalrecord): remove BE9 thin delegates`
- 削除: `go_safe.go` · `medical_record_legacy_constructor_test.go`
- 変更: `checkup_service.go` + 9 `*_test.go` → `WithTxAudit(..., nil /*auditTx*/, ...)`
- 確認: `rg 'func goSafe|goSafe\(|NewMedicalRecordService\('` → 0 hits
- 検証: 共有 DB 依存の広い `TestMedicalRecordService_` は schema 欠落で不安定なため、モック系 + `TestCheckupService_` + auto-create の focused run で green
- composition / lock / API 未変更 · push/PR 未実施

#### 次ラウンド（今回 Out of scope）

- lab_import_repository 547 行分割
- medicine_service + inventory 交差
- 巨大 `_test.go` 分割（prod と混ぜない）
- validators.go 全 call site の sharedkernel 直参照化
- daily_record（S6 でも原則触らない）
- RegisterRoutes / checkup package import 肥大

#### 検証（例）

```bash
docker compose exec backend go test ./internal/medicalrecord/ -count=1 -run 'TestMedicalRecordService_'
# S2: -run 'TestExamination'
# S3: -run 'TestHospitalization'
# S4: -run 'TestTreatment'
# S5: medicalrecord + 必要なら reservation write-owner focused
# S6: -run 'TestVital'
```

#### 完了条件（A1 全体）

- [x] S1–S6 が各 commit で着地（behavior-preserving file split）
- [x] examination / hospitalization discharge / treatment helpers / vital / appointment context が分離
- [x] composition 未変更
- [x] 公開 API / write owner / isolation 不変（focused test + reservation write-owner）

#### ARCH-A1 完了証跡（全コミット）

| Slice | Commit | Message |
|-------|--------|---------|
| S1 | `e7a2a53b4` | remove BE9 thin delegates |
| S2 | `fe3bed53f` | split examination items write path |
| S3 | `a4b830648` | split hospitalization discharge with billing path |
| S4 | `c56ad84c3` | split treatment field and master FK helpers |
| S6 | `1e58c85d0` | split vital validation and audit helpers |
| S5 | `bd7f9f677` | extract medical record appointment context helpers |

**USER 残作業**

- [x] push: `main` → `origin/main`（`54e876262..bd7f9f677`）
- [x] claim release:
  - `claim/ARCH-A1-S5` deleted（was `1e58c85d0`）
  - `claim/ARCH-A1-S6` deleted（was `c56ad84c3`）

---

### ARCH-A2: 共有 `internal/model` の所有明確化（高）

| 項目 | 内容 |
|------|------|
| **ID** | ARCH-A2 |
| **優先度** | P1 |
| **状態** | **done**（claim `claim/ARCH-A2` — user release after merge） |
| **対象** | `backend/internal/model/**` · 各 domain の DTO/command |
| **実測** | ~91 production files / ~4.5k LOC |
| **問題** | package 境界は domain だが型所有が中央に残り、見えない結合が増えやすい |
| **方針** | 一括移動しない。新規と触った箇所だけ owner を明確化 |
| **成果物** | `docs/architecture/model-write-owner-catalog.md` + TAP test |

#### チェックリスト

- [x] **A2-1** 主要 business fact について「GORM 型の write owner package」一覧を作る — catalog table
- [x] **A2-2** 新規型: owner domain に寄せるか、model に置くなら owner を明記 — Rules §1–2
- [x] **A2-3** domain 専用 command / DTO を model から分離し続ける — Rules §3
- [x] **A2-4** 「model に足しただけ・振る舞いが owner にない」新規をレビューで落とす — Rules §4 + PR checklist
- [x] **A2-5** 共有 ID / 列挙の複製はしない — Rules §5

#### 完了条件

- [x] 新規 fact の owner が PR 説明で必ず言える（catalog Rules + checklist）
- [x] model 一括再配置の Issue を起票しない（Rules §6）

---

### ARCH-A3: Cross-domain orchestration 契約の固定（高）

| 項目 | 内容 |
|------|------|
| **ID** | ARCH-A3 |
| **優先度** | P0–P1 |
| **状態** | done（A3-4 residual） |
| **対象** | reservation ↔ medicalrecord ↔ billing ↔ trimming ↔ lstep ↔ staff 等 |
| **問題** | intent API（良い）と best-effort 別 tx（契約として残存）が混在し、保証範囲が path ごとに違う |
| **既知例** | 予約確定→カルテ auto-create best-effort · キャンセル→draft カルテ cleanup best-effort · billing の ambient tx 参加 intent |
| **成果物** | `docs/architecture/cross-domain-orchestration-catalog.md` + `cross-domain-orchestration-catalog.test.mjs`（graph r8 COMPLETE `9d706c97-8fa3-4f41-854e-d753d474cafd`） |

#### チェックリスト

- [x] **A3-1** cross-domain 経路カタログを作る（1 表）  
  列: initiator / owner operation / transaction boundary / fail-closed vs best-effort / failure recovery / audit participation / test anchors
- [x] **A3-2** 新規 path ルールを明文化（catalog「New-path rules」）  
  - owner の typed intent のみ  
  - ambient tx 参加 or 明示 orchestration  
  - fail-closed か、best-effort なら補償・再試行・観測をセット  
  - silent 部分成功を増やさない
- [x] **A3-3** 既存 best-effort path をカタログ上で「意図的契約」としてラベル付け（無断で同 tx 化しない）  
  PATH-RES-MR-AUTOCREATE / PATH-RES-MR-CANCEL-CLEANUP
- [ ] **A3-4** automation / batch も同じ契約表に載せる（residual — 次バッチで lstep/batch 等を追記）
- [x] **A3-5** カタログの置き場: `docs/architecture/cross-domain-orchestration-catalog.md`（STATUS 長文なし）

#### 完了条件

- [x] 主要 cross-domain path が一覧で追える（TAP 回帰付き）
- [x] 新規 path の設計レビューで表の列が埋まらないとマージしない運用になる（catalog new-path rules + 行追加必須）

---

### ARCH-A4: 次点 domain の局所整理（中・トリガー駆動）

| 項目 | 内容 |
|------|------|
| **ID** | ARCH-A4 |
| **優先度** | P2 |
| **状態** | open（着手条件付き） |
| **方針** | 大きいから割らない。変更痛みの実測が出てから |

#### A4-lstep（~16.5k prod LOC）

- [ ] 着手条件: LINE 再開（#259 等）が実際に始まる直前
- [ ] batch / tag sync / delivery の内部凝集を見直す（同一 package 内）
- [ ] 通常 app から変な cross write を増やさない

#### A4-billing（~14.5k prod LOC）

- [ ] 着手条件: estimate / accounting / billing_item の PR が毎回広域
- [ ] 巨大 file 分割候補: `billing_item_service.go` · `billing_item_repository.go` · `estimate_service.go` · accounting 系
- [ ] 締め後編集の audit 同 tx fail-closed を壊さない

#### A4-reservation（~13.3k prod LOC）

- [ ] write owner lint を維持（owner 外 generic update 禁止）
- [ ] owner **内** `map[string]any` update を、触る変更のたびに typed command へ寄せる（一括禁止）
- [ ] intent repository / reservation_service の肥大が痛みになったら file 分割

#### 完了条件

- [ ] 各 subdomain は「トリガー記録 → スライス 1 本」単位で閉じる
- [ ] 三つ同時リファクタをしない

---

### ARCH-A5: Composition root の再 god 化防止（中）

| 項目 | 内容 |
|------|------|
| **ID** | ARCH-A5 |
| **優先度** | P1 |
| **状態** | **done**（claim `claim/ARCH-A5` — user release after merge） |
| **対象** | `backend/cmd/api/**` |
| **問題** | medicalrecord 配線だけで ~595 行規模。第2の Services aggregator 化リスク |
| **成果物** | `docs/architecture/composition-root-conventions.md` · `cmd/api/composition_root_conventions_lint_test.go` |

#### チェックリスト

- [x] **A5-1** 新規依存は `composition_<domain>.go` + domain 側 constructor / `Dependencies` に閉じる — conventions + required file pin
- [x] **A5-2** main に field を増やし続けない — main.go domain wiring call 禁止 lint
- [x] **A5-3** lstep `Application`/`Dependencies` を medicalrecord へ強制せず評価 — conventions §A5-3（現状 composition_* 維持）
- [x] **A5-4** route composition smoke を監視ゲートとして固定 — conventions が `route_composition_smoke_test` を正本参照
- [x] **A5-5** consumer 0 の root facade / god `Services`·`Repositories` 型を lint で拒否

#### 完了条件

- [x] 新規 domain 機能追加時、composition diff が「配線だけ」で読める運用文書がある
- [x] 巨大 aggregator 型の復活を scoped test で落とせる

---

### ARCH-A6: 許可依存グラフの機械ガード（中）

| 項目 | 内容 |
|------|------|
| **ID** | ARCH-A6 |
| **優先度** | P1 |
| **状態** | **done**（claim `claim/ARCH-A6` — user release after merge） |
| **対象** | domain 間 production import · [boundary map §5](docs/architecture/be9-2a-boundary-map.md) · ADR-006 |
| **問題** | 文書上の許可依存と実装がドリフトしうる |
| **成果物** | `backend/internal/lintscan/domain_import_allowlist_lint_test.go` · boundary map §5.2 |

#### チェックリスト

- [x] **A6-1** production import の allowlist（誰が誰を import してよいか）を定義 — `domainImportAllowlist`
- [x] **A6-2** 専用 test で機械チェック — `TestDomainImportAllowlistLint`（allowlist acyclic + real tree + mutations）
- [x] **A6-3** 新規 edge は ADR / boundary map 更新を必須にする運用を書く — §5.2
- [x] **A6-4** code-review-graph の community / bridge を四半期境界監査に使う手順を 1 節で固定 — §5.2
- [x] **A6-5** consumer-side interface 経由であるべき結合が具象 import に戻っていないか重点監視 — §5.2 禁止例 + allowlist 非掲載

#### 完了条件

- [x] 許可外 import が scoped test で落ちる（`go test ./internal/lintscan/ -run DomainImportAllowlist`）
- [x] 新規 domain edge のレビュー観点がチェックリスト化されている（§5.2）

---

### ARCH-A7: Frontend feature 境界と BE 対応（中〜低）

| 項目 | 内容 |
|------|------|
| **ID** | ARCH-A7 |
| **優先度** | P2 |
| **状態** | open |
| **対象** | `frontend/src/features/**` · `components` · `hooks` · `lib` · `shared-liff` |
| **問題** | feature 分割は良いが、FE↔BE domain / RBAC 対応が暗黙。共有層肥大で境界が薄まる |

#### チェックリスト

- [ ] **A7-1** 主要フローの FE feature ↔ BE domain / RBAC resource 対応表（薄い 1 枚）
- [ ] **A7-2** 新規 UI は必ず `features/` 配下
- [ ] **A7-3** 共有昇格ルール: 消費者 2 以上 + 理由。`components`/`hooks`/`lib` の安易な肥大を止める
- [ ] **A7-4** `shared-liff` と `line-reservation` の責務境界を維持
- [ ] **A7-5** Feature Indexing / `index.ts` 公開面の崩れを見つけたらその feature だけ直す（一括再編しない）

#### 完了条件

- [ ] 新規画面の置き場で迷わない
- [ ] FE 全体の Clean/layer 再編をしない

---

### ARCH-A8: 例外 package を増やさない（低・規律）

| 項目 | 内容 |
|------|------|
| **ID** | ARCH-A8 |
| **優先度** | P2 |
| **状態** | open（継続規律） |
| **対象** | `csvimport` · `identitylink` · `httpapi` · `sharedkernel` · `persistence` · `audit` 等 |

#### チェックリスト

- [ ] **A8-1** `csvimport` は 21 表 cutover 専用のまま。通常 app から汎用 write API 化しない
- [ ] **A8-2** 新規「便利な cross-domain write 例外」を作らない。作るなら ADR 級
- [ ] **A8-3** `identitylink` は owner/pet を Go import しない設計を維持
- [ ] **A8-4** `common` / `util` 的 bucket package を新設しない
- [ ] **A8-5** 横断能力は実 consumer が 2+ になってから命名抽出（先回り抽出禁止）

#### 完了条件

- [ ] 例外 package の増加が ADR なしで起きていない

---

## 3. 実施順（アーキのみ）

| 順 | ID | 内容 |
|----|-----|------|
| 1 | ARCH-A1 | medicalrecord 内凝集（削除 → file 分割） |
| 2 | ARCH-A3 | cross-domain 契約カタログ |
| 3 | ARCH-A6 | 許可依存の機械ガード |
| 4 | ARCH-A2 | model 所有の漸進明確化 |
| 5 | ARCH-A5 | composition 規律・Dependencies 横展開 |
| 6 | ARCH-A4 / A7 | 痛みが出た domain / FE のみ |
| 継続 | ARCH-A8 | 例外を増やさない |

---

## 4. 作業ルール（このバックログ用）

1. **Plan-only → 小さな PR**。大きな「アーキ刷新」PR を作らない  
2. **product-philosophy**: 追加より削除・簡素化。存在すべきでない最適化をしない  
3. **検証**: Docker スコープ限定。フル `go test ./...` / フル lint / DB reset は agent 自動実行禁止  
4. **migration**: agent は適用しない。必要になったスライスは別ゲート  
5. **並行作業**: claim プロトコル / worktree 隔離（[`git-worktree-safety`](.claude/rules/git-worktree-safety.md)）  
6. **正本**: 製品 residual・BUG・Issue は [`STATUS.md`](STATUS.md)。**本ファイルはアーキ改善専用**  
7. 完了したら該当チェックを `[x]` にし、証跡（commit / PR / 短いメモ）を 1 行足す  

---

## 5. 参考コマンド（調査）

```bash
# グラフ
code-review-graph status
code-review-graph update
code-review-graph detect-changes --brief

# package 規模（prod）
# find backend/internal/<pkg> -name '*.go' ! -name '*_test.go' | xargs wc -l

# 許可依存の下調べ例
# rg -l 'AnimalEkarte/backend/internal/medicalrecord' backend --glob '*.go' -g '!*_test.go'
```

---

## 6. 一行まとめ

> アーキの次の仕事は流派の乗り換えではない。  
> domain-first を保ったまま、巨大 module・共有 model・境界オーケストレーション・依存グラフを、  
> **変更単位と write owner** の軸で締める。
