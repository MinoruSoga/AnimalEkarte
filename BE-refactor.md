# BE-refactor — backend 規約適合監査と改善計画

> 更新: 2026-07-28（round 3: round 1 の 105 所見へ敵対的反証）／round 2 是正: false positive 除去／round 2 初回: 2026-07-28 `72fec9d1d`／初版: 2026-07-27
>
> **測定基準リビジョン (round 2)**: `c4ce786e0a52968a4fcdeaaffe07a409501ca80b`（`git rev-parse HEAD`）
> **working tree dirty (backend/)**: `billing/accounting_repository_reports_close.go` ほか billing 4、`lintscan/grandchild_parent_clinic_correlation_lint_test.go` 1（並行セッション編集中）。判定は on-disk 実コードを優先。
> **round 2 範囲**: CRITICAL/HIGH 35 件再実測 / 未監査面スイープ（lstep 未引用 89 + 未引用 package 11 + テスト品質 + N+1/unbounded + X-01〜X-10 畳み込み）/ MEDIUM+LOW 70 件位置照合。
>
> **本書の位置づけ**: `backend/` 全域の規約適合監査で検出した改善項目の正本。`FE-refactor.md` の backend 版であり、readiness view（`3-session-agent.html`）とは役割が異なる。
>
> **本書に書かないもの**:
> - **規約テキストそのもの**（2026-07-26 `ddc609a79` で旧 BE-refactor.md の 4 規定を `.claude/refs/go-gin-backend-review.md` §6 へ移設して退役済み。ルールを本書へ書き戻すと移設が無言で巻き戻る）。本書は規約を **引用するだけ** で、正本は常に規約側にある。
> - **既に起票済みの課題の詳細**（`3-session-agent.html#ledger` / `BE-pending.md` / `phase2.html` が正本。該当項目は「区分: 既知」として台帳 ID を示す）。
>
> **writer 規則**: 本書の writer は監査を実行するセッション単独とする。レーン A/B の実行セッションは本書を読むのみで書かない（4 セッション並行のため）。
>
> **2026-07-28 追補（着手フェーズの writer 規則）**: 着手プランの実行セッションは例外として、**自分が担当するユニットの `- Status:` 行 1 行だけ**を書き換えてよい。それ以外（所見本文・判定行・ユニット定義・表）は読むのみ。書き換え前に本ファイルを再読し、commit は本ファイルと自ユニットの所有パスを path 指定した形で行うこと。**commit 後は速やかに push する** — ローカルにしかない commit は、並行セッションが origin へ強制的に合わせ直す操作を打った際に消える（2026-07-28 に実害あり。reflog から復旧した）。
>
> **着手手順**: 本書の項目は台帳タスクではない。着手すると決めたものを `3-session-agent.html#ledger` へ `<section class="task">` として起票し、本書の該当項目へ起票先 ID を追記する。**着手前に evidence の再実測を必須とする**（下記「監査の限界」5 を参照）。

---

## 監査方法とカバレッジ

- **実施**: 2026-07-27・14 ドメイン単位に分割した並列監査 → 単位ごとの敵対的検証 → 集約（31 エージェント・read-only）。
- **判定**: 全所見に規約正本の `file:line` と逐語引用を必須とし、検証フェーズで「引用された規約がその行に実在し、その状況に適用されるか」「evidence の file:line が記述どおりか」「既に対処済みでないか」を再実測した。**18 件が却下**され、残った **105 件**を本書に全件列挙する。
- **未実施の検証**: read-only 制約により `go test` / `golangci-lint` / `gofmt` / DB 接続を一切実行していない。したがって並行性シナリオ（TOCTOU・ロック挙動）は **コード構造と PostgreSQL のセマンティクスからの導出であり、実測トレースではない**。性能系（N+1・クエリプラン）は実測が取れないため原則として発行していない。
- **読了率**: 12/14 単位は担当ファイルを **全文読了**（sample・skip ゼロ）。例外は lstep 2 単位で、`lstep [a-l]` は 124 ファイル中 79 ファイルを全文読了し、残 45 ファイルは 10 種のアンチパターン grep を全数走査で代替した。**lstep の未読領域に残存所見がある可能性を明示しておく**。
- **並行 WIP の扱い**: 監査中にレーンが `internal/lintscan` の inventory 登録等を清算した（`afd8404a4` 他）。検証時点で clean になったファイルは「新規（着手可能）」へ格上げ済み。`pet/repository.go` 系のみ WIP のまま残る。

---

## サマリー

| severity | 件数 | 内訳の代表 |
|---|---|---|
| CRITICAL | 2（round1 検出）／ open 1 after round3 | 締め後会計編集ゲートの迂回（BIL-01・open・round3 REFRAMED 分割）／ stage-import DELETE（CMD-01・**WITHDRAWN** round3・BUG-430 退役） |
| HIGH | 33 | commit 済み write の 5xx 反転、fail-open、監査欠落、TOCTOU、入力境界検証 |
| MEDIUM | 56 | 非原子 write、raw error 露出、copy-paste drift、契約非対称 |
| LOW | 14 | 重複ログ、命名 stutter、doc コメント |
| **計** | **105** | うち既知起票済み 6・WIP-adjacent 5 |

**round 3（2026-07-28）反証後の現況**: round 1 の 105 件に敵対的却下パスを通した結果、**WITHDRAWN 8 / DOWNGRADED 8 / REFRAMED 17 / UPHELD 72**。上表の件数は round 1 検出時点のものであり、**現在の着手可否は各所見の `- round3-review(2026-07-28):` 行が正本**である。severity 別索引の `WITHDRAWN(round3)` 印が付いた 8 件は着手対象外。判定の全件一覧と理由は「## round 3（2026-07-28）— round 1 の 105 所見へ敵対的反証」節を参照。

**着手の運用（2026-07-28 決定）**: 本書の所見は台帳へ一括起票しない。**本書を正本として直接着手する。** 起票済みは反証を通過した確定 7 件のみ（BUG-463〜466 / TASK-467〜469）。着手前に evidence を再実測する規則は継続する（round 3 で 8 件が LINE-DRIFT・2 件が ALREADY-FIXED と判明している）。

**規約側の疑義**: 32 論点（実装を直す前に確定すべきもの。末尾に全件掲載）

**最も重い 3 点**:

1. **[BIL-01] 締め後の会計編集ゲートが `billing-items` 経路で完全に迂回される**（CRITICAL）。`POST/PATCH /billing-items` は billing の status もレジ締め状態も検査せず、監査も残さずに billing 総額を書き換える。**coordinator 独立実測済み**。
2. **[X-01 パターン] commit 済みの成功を tx 外の再取得エラーで 5xx へ反転させる実装が 7 ドメイン 15 箇所以上**（HIGH）。臨床側には「失敗」と表示されるが DB は更新済みで、再送すると 409 か二重適用になる。
3. **[X-04 パターン] error 握り潰しによる fail-open が lstep バッチ群に集中**（HIGH）。DB エラーで `return 0, nil` を返すため、クリニック全体のタグ同期が全滅しても失敗件数に計上されず監査にも残らない。（注: LSA-02/LSB-01 の opt-out 欠落は **X-04 ではなく** business-evidence fail-open。error 握り潰しではない）。

---

## 横断パターン（同型の所見をまとめて直すための索引）

個別に 105 回直すのではなく、パターン単位で**同一リポジトリ内の正しい先例へ寄せる**のが正しい対応である。各パターンの構成員 ID は下記の全件列挙で参照する。

| パターン | 規約 | 構成員 ID | 推奨アプローチ |
|---|---|---|---|
| **X-01** commit 済み write を tx 外の再取得 error で失敗へ反転 | `CODING_RULES.md:78` | MRA-02, MRC-01, MRD-02, POC-02, RSV-03, TRM-01, LSB-03 | 再取得を `WithTx` 内へ移す。先例 = `vital_service.go:294-298` / `owner/repository.go:288` / `pet/repository.go:417` / `trimming_service.go:330` |
| **X-02** 入力境界検証の欠落（列挙値・範囲・長さ） | `guidelines:151` | AUS-03, LSA-13, LSA-14, LSB-06, TRM-05, MRA-03, MRD-04, RSV-04, MRC-08, POC-13, POC-17, MRB-06, TRM-03（兼 X-07） | enum は model 定数から `oneof=` を導出し、追随を test で固定。長さは `001_init.sql` の列定義から導出。MDL-04 は LSA-13 と同一 trigger_type のため除外（正本 LSA-13）・round3 **WITHDRAWN**（重複）。POC-14 は X-09 へ移動 |
| **X-03** destructive / irreversible 操作の監査・recovery 欠落 | `invariants.md:31` | AUS-01, LSA-05, MRA-01, MRB-05, POC-07, TRM-02, TRM-07 | 既存の fail-closed 監査（`audit.LogEntryTx`）へ寄せる。5 領域に確立済みパターンあり・新テーブル不要 |
| **X-04** error 握り潰しによる fail-open | `error-handling.md:9` | LSA-03, LSA-10, LSA-11, LSA-12, LSA-16, LSB-04, MRC-04, MDL-02, INF-06, CMD-03, CMD-07, RSV-09, LSB-02 | 「握り潰す」と「意図的 best-effort」を分離。後者は `CODING_RULES.md:36` が要求する補償・再試行・監査・部分失敗 contract を実際に持たせる。**LSA-02/LSB-01 は X-04 ではない**（business evidence 欠落） |
| **X-05** 検証と write が同一 tx にない（TOCTOU） | `CODING_RULES.md:38` | POC-03, MRB-02, MRB-08, MRC-03, MRC-07, POC-05, RSV-02, RSV-07, POC-06, LSA-15 | 検証と write を単一 `WithTx` へ収め、判定根拠の行を `FOR UPDATE`/`FOR SHARE` で固定。先例 = `reservation_intent_repository.go:591-625` |
| **X-06** business graph を構成する write の非原子化 | `CLAUDE.md:33` | BIL-03, LSA-06, MRC-05, RSV-06, MRC-12（既知） | `Transactor` を注入し repository を `persistence.DBOrTx` 経由へ揃える |
| **X-07** request body サイズ上限の非対称・middleware による無効化 | `guidelines:180`（round3 LINE_DRIFT 是正・旧:179） | INF-02, POC-12, TRM-03（兼 X-02） | `protected` グループ全体へ body size middleware を 1 本入れて統一する（handler 個別対応は非対称を再生産する）。TRM-03 の string max 面は X-02 |
| **X-08** 外部 API の raw error を応答・DB へ露出 | `CLAUDE.md:34` | LSA-08, LSA-09 | 応答は安定した分類コードのみ。生エラーは slog に限定。LSB-05 は LSA-08 と同一疎通 raw error のため除外（正本 LSA-08）・round3 **WITHDRAWN**（重複） |
| **X-09** copy-paste drift（複製と乖離） | `coding-style.md:26` | MDL-01, POC-11, POC-16, MRC-14, POC-14 | `sharedkernel` へ 1 本化。**税計算の乖離（MDL-01）は金額に直結するため先行**。POC-14 は空 PATCH ガードの兄弟欠落 |
| **X-10** 同一 error の重複ログ / 二重レスポンス | `CODING_RULES.md:67`, `error-handling.md:29` | POC-15, TRM-06, AUS-09 | 文脈の揃う境界 1 箇所へ集約 |

---

## 全所見一覧（105 件・severity 順）

| ID | severity | 対象ドメイン | 概要 | 横断 | 区分 |
|---|---|---|---|---|---|
| BIL-01 | CRITICAL | billing / inventory | billing-items の POST/PATCH がレジ締め後の会計編集ゲート（po… | — | 既知 → BUG-463 |
| CMD-01 | CRITICAL | cmd | stage-import の破壊的 DELETE が clinic_id で制約されてい… | — | 既知・**WITHDRAWN(round3)** |
| AUS-01 | HIGH | auth / staff | スタッフの所属医院全置換（destructive replace）に監査エントリが一切残… | X-03 | 新規 |
| BIL-03 | HIGH | billing / inventory | campaignService.Update が本体更新と対象（カテゴリ/商品）差し替え… | X-06 | 新規 |
| INF-01 | HIGH | infra 横断 | 未分類のPostgreSQLサーバエラーが全て HTTP 400「入力値が正しくありませ… | — | 新規・**WITHDRAWN(round3)** |
| INF-02 | HIGH | infra 横断 | SanitizeNullBytes の透過 Reader が http.MaxBytes… | X-07 | 新規 |
| LSA-01 | HIGH | lstep | lstep_base_url が無検証で保存され、復号済み Lステップ API キーが任… | — | 新規 |
| LSA-02 | HIGH | lstep | 自動配信トリガーの除外判定が owner.LstepOptOut を読まず、オプトアウト… | — | 新規 |
| LSA-03 | HIGH | lstep | visit-dormant バッチが DB エラーを握り潰して (0, nil) を返し… | X-04 | 新規 |
| LSA-04 | HIGH | lstep | lstep タグ設定3テーブルが clinic_id を持たない全院共有行なのに、cli… | — | 新規 |
| LSA-05 | HIGH | lstep | 飼い主 opt-out / opt-in / 削除クリーンアップが監査ログを一切残さず、… | X-03 | 新規 |
| LSA-06 | HIGH | lstep | PATCH /lstep-settings が clinic_integrations … | X-06 | 新規 |
| LSA-07 | HIGH | lstep | LINE push 成功ログに LINE User ID（飼い主識別子）をそのまま出力し… | — | 新規 |
| LSB-01 | HIGH | lstep | 自動配信トリガーが owner.LstepOptOut を一切参照せず、オプトアウト済み… | — | 新規 |
| LSB-02 | HIGH | lstep | removeAllTagsFromLstep がリモート解除失敗・client nil … | — | 新規 |
| LSB-03 | HIGH | lstep | HandlePetDeath がcommit済みのペット死亡記録をcommit後のrea… | X-01 | 新規 |
| LSB-04 | HIGH | lstep | credential 復号失敗を空文字へ置換して握り潰す3箇所 — Lステップ連携がサイ… | X-04 | 新規 |
| MRA-01 | HIGH | medicalrecord | care_plan_items の物理削除に監査もrecovery経路も存在しない（権限… | X-03 | 新規 |
| MRA-02 | HIGH | medicalrecord | commit済みの care plan item write を後段 re-fetch … | X-01 | 新規 |
| MRB-02 | HIGH | medicalrecord | hospitalizationRepository.LockByIDForUpdate … | X-05 | WIP |
| MRB-03 | HIGH | medicalrecord | hospitalizationRepository.FindByID が clinic … | — | 既知 |
| MRC-01 | HIGH | medicalrecord | 処方更新: commit済みの成功を tx 外の再取得 error で失敗応答へ反転させ… | X-01 | 新規 |
| MRC-02 | HIGH | medicalrecord | 薬剤削除の連携在庫カスケードが FK ではなく可変の name をキーにし、affect… | — | 新規 |
| MRC-04 | HIGH | medicalrecord | カルテ作成の主訴・治療方針・診断がサイレントに消失し、API は 201 を返す | X-04 | 新規 |
| MRC-05 | HIGH | medicalrecord | lab import: exams と exam_results が非原子的に書かれ、補… | X-06 | 新規 |
| MRD-01 | HIGH | medicalrecord | 治療項目の並び順一括更新が affected rows を確認せず、かつ施錠した親カルテ… | — | 新規 |
| MRD-02 | HIGH | medicalrecord | commit 済みの更新を、後段の応答用再取得エラーで失敗応答へ反転させる | X-01 | 新規 |
| MRD-03 | HIGH | medicalrecord | treatment plan の write が親（カルテ/入院）所属を検証せず、res… | — | 新規 |
| MRD-04 | HIGH | medicalrecord | treatment plan の金額系入力が binding tag でも servic… | X-02 | 新規 |
| POC-02 | HIGH | pet / owner / clinic | write→reload が transaction 外で行われ、commit 済み成功… | X-01 | 新規 |
| POC-03 | HIGH | pet / owner / clinic | pet 更新の owner_id / insurance_id 再検証が write t… | X-05 | WIP |
| RSV-02 | HIGH | reservation | 管理画面予約作成だけが AcquireBookingLock を取得せず、空き枠へのファ… | X-05 | 新規 |
| TRM-01 | HIGH | trimming / manualarticle / csvimport | manualarticle Upsert がcommit後に再取得し、commit済みの… | X-01 | 新規 |
| TRM-03 | HIGH | trimming / manualarticle / csvimport | trimming/manualarticle の自由入力文字列に長さ上限が一切なく、re… | X-07 | 新規 |
| TRM-04 | HIGH | trimming / manualarticle / csvimport | 既存detailを持つappointmentへのPOST /trimmingsが、何も書… | — | 新規 |
| AUS-03 | MEDIUM | auth / staff | staff_type が application 層で一切検証されず DB enum だ… | X-02 | 新規 |
| AUS-04 | MEDIUM | auth / staff | ShiftTemplateRepository.CountUsageByShiftTem… | — | 新規 |
| AUS-05 | MEDIUM | auth / staff | toShiftResponse が *model.Staff を nil ガードなしで参… | — | 新規 |
| AUS-06 | MEDIUM | auth / staff | auth/http_session.go が 820 行で上限 800 行を超過 | — | 新規・**WITHDRAWN(round3)** |
| BIL-02 | MEDIUM | billing / inventory | accountingService の会計取消・クレジット訂正監査が「監査depende… | — | 新規 |
| CMD-02 | MEDIUM | cmd | 全院バッチを起動する /_internal/scheduled-jobs が未認証の r… | — | 新規 |
| CMD-03 | MEDIUM | cmd | coverage-ratchet の失敗が二重に無音化され gate が構造的に機能しない | X-04 | 新規 |
| CMD-04 | MEDIUM | cmd | pet 死亡/復活の cross-domain write が composition … | — | WIP |
| CMD-05 | MEDIUM | cmd | /uploads の StaticFS が無条件・未認証で登録され非 release m… | — | 新規 |
| CMD-06 | MEDIUM | cmd | csv-import-failure-rehearsal が error chain を… | — | 新規 |
| CMD-07 | MEDIUM | cmd | lstep-migrate の進捗台帳書込失敗が warn ログのみで呼び出し元へ返らない | X-04 | 新規 |
| INF-03 | MEDIUM | infra 横断 | Lステップ API の URL パスに line_user_id を無エスケープで埋め込… | — | 新規 |
| INF-04 | MEDIUM | infra 横断 | apperrors.FromGORM が pgx ドライバエラーを err.Error(… | — | 新規 |
| LSA-08 | MEDIUM | lstep | 疎通確認 API が外部 HTTP エラーを err.Error() のまま JSON … | X-08 | 新規 |
| LSA-09 | MEDIUM | lstep | LINE 送信失敗時に外部 API の生エラー文字列を 502 応答と送信履歴 API … | X-08 | 新規 |
| LSA-10 | MEDIUM | lstep | GetOwnerTags のキャッシュフォールバックが DB エラーを握り潰し、200 … | X-04 | 新規 |
| LSA-11 | MEDIUM | lstep | removeStaleTagsByPrefixes がタグキャッシュ読み取り失敗を「AP… | X-04 | 新規 |
| LSA-12 | MEDIUM | lstep | 配信トリガーログの excluded/failed への status 更新失敗が no… | X-04 | 新規 |
| LSA-13 | MEDIUM | lstep | trigger_type が列挙値検証されず、存在しないトリガー種別の優先順位が 200… | X-02 | 新規 |
| LSA-14 | MEDIUM | lstep | 自動管理プレフィックスの category が列挙値検証されず、C2 以外の綴りで登録す… | X-02 | 新規 |
| LSA-15 | MEDIUM | lstep | 配信トリガーの二重発火防止が check-then-Create のみで、DB 側に一意… | — | 新規 |
| LSB-06 | MEDIUM | lstep | shared-files アップロードの purpose が列挙値・長さとも未検証で、超… | X-02 | 新規 |
| MDL-01 | MEDIUM | model / lintscan / testdb | EstimateItem.CalculateTaxAmount が DiscountAm… | X-09 | 新規 |
| MDL-02 | MEDIUM | model / lintscan / testdb | testdb のテスト間データ分離 TRUNCATE が error を全件破棄し fa… | X-04 | 新規・**WITHDRAWN(round3)** |
| MDL-04 | MEDIUM | model / lintscan / testdb | 配信トリガー優先順位の trigger_type が列挙値検証を受けず、任意文字列が永続… | X-02 | 新規・**WITHDRAWN(round3)** |
| MDL-05 | MEDIUM | model / lintscan / testdb | CASCADE lint が完全一致リテラル検索のため、表記ゆれした ON DELETE… | — | 新規 |
| MRA-03 | MEDIUM | medicalrecord | consultation master の tax_rate に範囲検証が無い（同pac… | X-02 | 新規 |
| MRB-05 | MEDIUM | medicalrecord | 入院の削除および PATCH による退院化が監査ログを残さない（同 service は監… | X-03 | 新規 |
| MRB-07 | MEDIUM | medicalrecord | 検査種別フィールドの transaction 依存欠落を 400 InvalidInpu… | — | WIP |
| MRB-08 | MEDIUM | medicalrecord | 検査種別の request 由来 parent_id 検証が永続化と同じ transac… | X-05 | WIP |
| MRC-03 | MEDIUM | medicalrecord | #201 投与量パラメータ: per_weight 医療安全ガードが write tra… | X-05 | 新規 |
| MRC-08 | MEDIUM | medicalrecord | lab import の検査項目 DTO が無検証で、異常判定を決める基準値をクライアン… | X-02 | 新規 |
| MRC-09 | MEDIUM | medicalrecord | 診療画像の JSON 作成経路が MIME allowlist と形式検証を一切通らない… | — | 新規 |
| MRC-12 | MEDIUM | medicalrecord | 問診 upsert が Conflict 応答時に FirstOrCreate で作った… | X-06 | 既知 |
| MRC-14 | MEDIUM | medicalrecord | 診断FK の clinic 所有権検証が ClinicalPlanService から複… | X-09 | 新規 |
| POC-01 | MEDIUM | pet / owner / clinic | 休診日ミューテーションが2つのroute群で二重登録され、必要権限が分岐している | — | 新規 |
| POC-05 | MEDIUM | pet / owner / clinic | 特別期間の重複禁止が application 検証のみで、DB 制約も transact… | X-05 | 新規 |
| POC-06 | MEDIUM | pet / owner / clinic | 飼主 phone の一意性が application 検証のみで、email と異なり … | — | WIP |
| POC-07 | MEDIUM | pet / owner / clinic | 全クリニック共有マスタ animal_species の更新・削除に監査記録が一切残らない | X-03 | 新規 |
| POC-08 | MEDIUM | pet / owner / clinic | clinic スコープ付き飼主更新 route 5本が OpenAPI 未宣言で、:cl… | — | 新規 |
| POC-10 | MEDIUM | pet / owner / clinic | 締め時刻の既定値 3 種が clinic_settings_repository に 6… | — | 新規 |
| POC-11 | MEDIUM | pet / owner / clinic | ペット列挙値バリデータ4関数が owner / pet で完全複製され、既に構造ドリフト… | X-09 | 新規 |
| POC-12 | MEDIUM | pet / owner / clinic | pet / owner / clinic の JSON エンドポイントに request… | X-07 | 新規 |
| POC-13 | MEDIUM | pet / owner / clinic | pet / owner の自由記述フィールドに長さ検証が無い（同一 struct 内の他… | X-02 | 新規 |
| POC-14 | MEDIUM | pet / owner / clinic | 慢性疾患 PATCH に「更新対象フィールド0件」のガードが無い（兄弟実装3箇所には存在… | X-02 | 新規 |
| POC-15 | MEDIUM | pet / owner / clinic | 同一 error を同一関数内で2回 ErrorContext ログしている | X-10 | 新規 |
| POC-17 | MEDIUM | pet / owner / clinic | clinic / company の email・電話番号・郵便番号が未検証（owner… | X-02 | 新規 |
| RSV-03 | MEDIUM | reservation | write成功後の再取得がtransaction外にあり、read失敗がcommit済み… | X-01 | 新規 |
| RSV-04 | MEDIUM | reservation | LINE予約設定の28フィールドrequest DTOにbinding tagが1件も無… | X-02 | 新規 |
| RSV-06 | MEDIUM | reservation | 予約キャンセルが「status更新」と「soft delete」の2 writeに分かれ… | X-06 | 新規 |
| RSV-07 | MEDIUM | reservation | 依存チェック→削除がtransaction外で行われ、予約側がFOR SHAREで守って… | X-05 | 新規 |
| RSV-09 | MEDIUM | reservation | LIFF予約のスタッフ自動割当で ToDateTime / delegateStaff … | X-04 | 新規 |
| TRM-02 | MEDIUM | trimming / manualarticle / csvimport | マニュアル削除は物理削除で編集履歴がCASCADE消滅し、監査はbest-effortか… | X-03 | 新規 |
| TRM-05 | MEDIUM | trimming / manualarticle / csvimport | bw_unit のみ境界で列挙値検証が無く、DB enum 到達まで不正値が進む | X-02 | 新規 |
| TRM-06 | MEDIUM | trimming / manualarticle / csvimport | trimming detail 生成失敗が同一call chainで二重にERRORログ… | X-10 | 新規 |
| TRM-07 | MEDIUM | trimming / manualarticle / csvimport | csvimport.Import は35表を無条件DELETEするexported AP… | X-03 | 新規 |
| AUS-09 | LOW | auth / staff | HasPermission / CalculateEffectivePermission… | X-10 | 新規 |
| INF-06 | LOW | infra 横断 | audit.MarshalJSON が marshal 失敗を無記録で握り潰し、監査ログ… | X-04 | 新規 |
| LSA-16 | LOW | lstep | CSV インポートで json.Marshal のエラーを `_` で明示破棄している3箇所 | X-04 | 新規・**WITHDRAWN(round3)** |
| LSB-05 | LOW | lstep | 疎通確認APIが外部HTTPトランスポートのraw error文字列をそのままrespo… | X-08 | 新規・**WITHDRAWN(round3)** |
| MDL-06 | LOW | model / lintscan / testdb | model.Payment に ClinicID が無く、同一 business fac… | — | 既知 |
| MRA-04 | LOW | medicalrecord | 実在しない package 名を主語にする package comment が担当 7 … | — | 新規 |
| MRB-06 | LOW | medicalrecord | 入院の end_date >= start_date は DB CHECK のみで ap… | X-02 | 新規 |
| MRC-07 | LOW | medicalrecord | マスタ削除の使用中ガードが write と同一 transaction・同一ロック下にな… | X-05 | 新規 |
| POC-09 | LOW | pet / owner / clinic | clinic_settings / clinic_holiday の書き込みで Scop… | — | 新規 |
| POC-16 | LOW | pet / owner / clinic | parseHHMM が billing の同名 helper の自己申告済み複製として残… | X-09 | 新規 |
| RSV-08 | LOW | reservation | shift_entry_breaks の孫read が親 shift_entries の… | — | 新規 |
| TRM-08 | LOW | trimming / manualarticle / csvimport | [既知] Options preload の clinic predicate が末尾 … | — | 既知 |
| TRM-09 | LOW | trimming / manualarticle / csvimport | [WIP-adjacent] trimming の3 master repository… | — | WIP |
| TRM-10 | LOW | trimming / manualarticle / csvimport | package 名と export 名の stutter | — | 新規・**WITHDRAWN(round3)** |

---

## ドメイン別 詳細（全 105 件）

各項目は `ID: タイトル — severity` ／ 区分（新規・既知・WIP-adjacent）／ 規約の逐語引用 ／ 対象 `file:line` ／ 内容 ／ 修正方針 ／ 検証時の補正、で構成する。**「検証時の補正」は敵対的検証エージェントが原本コードと規約を再実測して訂正した内容であり、元の所見記述より優先する**。

### billing / inventory（会計・在庫）（3件）

#### BIL-01: billing-items の POST/PATCH がレジ締め後の会計編集ゲート（post-close権限・理由・fail-closed監査）を完全に迂回して billing 総額を書き換える — CRITICAL
- 区分: **既知** → BUG-463
- 規約: `.claude/refs/backend-application-invariants.md:37` 「fail-closedと定めたclinical/financial監査はbusiness writeと同じtransactionへ参加させ、監査dependency欠落または監査write失敗時はbusiness writeもrollbackする。締め後の会計編集はこの対象とする。」
- 対象: `backend/internal/billing/routes.go:111`、`backend/internal/billing/routes.go:112`、`backend/internal/billing/billing_item_service.go:404`、`backend/internal/billing/billing_item_service.go:282`、`backend/internal/billing/billing_item_repository.go:146`、`backend/internal/billing/billing_item_repository.go:280`、`backend/internal/billing/billing_item_service.go:471`、`backend/internal/billing/billing_item_repository.go:563`、`backend/internal/billing/accounting_handler.go:173`、`backend/internal/billing/accounting_service_core.go:234`、`backend/internal/billing/accounting_service_core.go:274`、`backend/internal/billing/cash_register_service.go:359`、`backend/internal/billing/cash_register_close_repository.go:33`
- 内容: `PATCH /billing-items/:id` と `POST /billing-items` は `accounting:edit` / `accounting:create` だけで通り、billing の status も レジ締め状態も検査しない。UpdateItem(billing_item_service.go:404-454) には status 検査が一切なく、CreateItem(:282-402) の完了/取消拒否は `input.VaccinationID != nil` の枝でしか到達しない（billing_item_repository.go:280-281）。非ワクチン経路の `ValidateCreateReferences` は billing を FOR UPDATE でロックしながら `medical_record_id, owner_id, pet_id` しか SELECT せず status を読まない（:146-153）。どちらの経路も末尾で `recalculateTotals` → `UpdateBillingTotals` を呼び、billings.subtotal/tax_total/total_amount を status 述語なしで上書きする（:563-571）。一方 DeleteItem は同一 billing をロックして completed/cancelled を 409 で拒否しており（billing_item_service.go:471-472）、この不変条件が意図されたものであることを自ら証明している。`IsDateClosed` の呼び出しは accounting_handler.go:173,232 の2箇所のみで、billing-items 経路からは一度も呼ばれない。結果として、accounting 側 PATCH では必須の `accounting-post-close-edit:edit` 権限・`post_close_reason`・同一tx fail-closed 監査（accounting_service_core.go:234-239, 274-296）が、明細経由なら全て回避される。既知台帳との弁別: BUG-440 は DeleteItem のワクチン provenance 解放 actor 監査、phase2.html:216 は accounting_handler.go:161 の権限チェックの「置き場所」であり、いずれも billing-items 経路に post-close ゲートが無いことを扱っていない。
- 修正: CreateItem / UpdateItem を DeleteItem と同型にし、tx 内で `billingRepo.LockAndFindByID` により completed/cancelled を 409 拒否した上で、締め済み期間（IsDateClosed）の明細変更は accounting PATCH と同じ post-close 権限・理由・`logPostCloseEdit` 相当の同一tx fail-closed 監査を必須にする。
- 検証時の補正: invariants:37 は監査participation規則であるため、本所見のうち本規約で成立するのは「締め後会計編集に該当する billing-items 経路が監査を一切持たない（fail-closed以前に監査が不在）」の部分に限る。所見が同梱する「UpdateItem/CreateItem が completed/cancelled を拒否しない（DeleteItem との非対称）」は実測上いずれも事実だが、これは state guard の欠落であり invariants:37 ではなく backend/CLAUDE.md:33（business graph 原子性）/ go-gin-backend-review.md:66（scope外対象を成功扱いしない）を根拠とすべき別命題である。是正時は post-close 監査（invariants:37 準拠の同一tx fail-closed）と status guard（DeleteItem:471-474 と同型）を別根拠の2件として扱うこと。
- **coordinator 独立実測済み（2026-07-27）**: `UpdateItem`（`billing_item_service.go:404-425`）は `FindByID` → 単価/数量検証 → `WithTx` → `repo.Update` で `LockAndFindByID` も status 検査も持たない。`DeleteItem`（`:456-475`）は `LockAndFindByID` で行を固定し completed/cancelled を `WrapConflict` で拒否する。`routes.go:112` の PATCH は `accounting:edit` のみを要求する。**本項はエージェント報告ではなく直接確認に基づく**（他 104 件は未実測 — 「監査の限界」5 を参照）。

- 再実測(2026-07-28): **LINE-DRIFT** — UpdateBillingTotals は billing_item_repository.go:568-576（旧:563）。IsDateClosed は cash_register_service.go:407（旧:359）。CreateItem/UpdateItem に status/post-close ゲート無し・DeleteItem のみ completed/cancelled 拒否は継続。dirty: billing_item_repository.go 作業ツリー変更中だが status 欠落構造は残存。
- round3-review(2026-07-28): **REFRAMED** — 内容は成立するが invariants:37 単独の CRITICAL 一本化は過剰。A) 締め後総額書換の post-close 権限・理由・同一tx監査欠落（CRITICAL・invariants:37 + cash-register #115）と B) Create/Update の completed/cancelled ガード欠落（DeleteItem 非対称・別 HIGH state-guard）に分割。`billing_item_service.go:404-429` UpdateItem に status/IsDateClosed 無し・DeleteItem のみ拒否を現 HEAD で確認。
#### BIL-03: campaignService.Update が本体更新と対象（カテゴリ/商品）差し替えを別々のtransactionで実行し、部分成功で割引マッチング対象が不整合になる — HIGH
- 区分: 新規 ／ 横断パターン: X-06
- 規約: `backend/CLAUDE.md:33` 「1つのbusiness graphを構成する複数rowのwriteは同じtransactionで原子的に扱い、commit済みの成功を後段の再取得errorで失敗応答へ反転させない。」
- 対象: `backend/internal/billing/campaign_service.go:286`、`backend/internal/billing/campaign_service.go:292`、`backend/internal/billing/campaign_service.go:177`、`backend/internal/billing/campaign_repository.go:76`、`backend/internal/billing/campaign_repository.go:93`、`backend/internal/billing/campaign_service.go:299`
- 内容: `campaignService` は Transactor を一切保持しない（campaign_service.go:177-180）。そのため Update は `s.repo.Update`（campaign_repository.go:76-80 が `r.db.WithContext(ctx)` で独立コミット）と `s.repo.ReplaceTargets`（campaign_repository.go:93 の `persistence.DBOrTx(ctx, r.db)` は ambient tx 不在のため素の db にフォールバックし、自前 tx を開く）を逐次実行する。ReplaceTargets が失敗すると、割引率・期間・有効フラグだけが更新済みで対象カテゴリ/商品は旧集合のまま残り、`FindApplicableForItem`（campaign_repository.go:145-161）を通る自動割引が誤った対象集合に対して新しい割引額を適用する。さらに `:299-303` は両writeがcommit済みになった後に再取得し、その read error を失敗応答へ写像しており、`backend/CODING_RULES.md:78`「write後の再取得が失敗し得る場合はcommit前の同じtransaction内で行うか、commit済みの成功を後段read errorで失敗へ反転させないcontractにする」にも反する（同一経路のため本件に内包して扱う）。
- 修正: `campaignService` に Transactor を注入し、Update/ReplaceTargets/最終再取得を単一 `WithTx` 内へ収める（billingItemService.UpdateItem と同型）。repository 側は既に `DBOrTx` なので ambient tx へ自動参加する。
- 検証時の補正: evidence の行番号を1行精密化: 実際の write 呼び出し行は campaign_service.go:287（`if _, err := s.repo.Update(ctx, clinicID, id, fields)`）と :293（`if err := s.repo.ReplaceTargets(ctx, id, cats, itemIDs)`）であり、所見が挙げた :286 / :292 はそれぞれ直前の `if len(fields) > 0 {` / `if hasTargets {` ガード行。repository 側の該当も :76-80 ではなく関数定義 :75 起点の :76-80 チェーンで正しい。

- 再実測(2026-07-28): **CONFIRMED** — campaignService に Transactor 無し（:177-180）。Update :287 と ReplaceTargets :293 が独立 commit。後段 FindByID :299 は X-01 併発。
- round3-review(2026-07-28): **UPHELD** — campaignService に Transactor 無し（:177-180）。Update:287 と ReplaceTargets:293 が独立 commit。後段 FindByID:299 は X-01 併発。`backend/CLAUDE.md:33` 適用。
#### BIL-02: accountingService の会計取消・クレジット訂正監査が「監査dependency欠落」を成功扱いにする（宣言済みの fail-closed 契約と逆） — MEDIUM
- 区分: 新規 ／ 検証で severity 引き下げ
- 規約: `.claude/refs/backend-application-invariants.md:37` 「fail-closedと定めたclinical/financial監査はbusiness writeと同じtransactionへ参加させ、監査dependency欠落または監査write失敗時はbusiness writeもrollbackする。締め後の会計編集はこの対象とする。」
- 対象: `backend/internal/billing/accounting_service_correction.go:178`、`backend/internal/billing/accounting_service_correction.go:179`、`backend/internal/billing/accounting_service_reports.go:93`、`backend/internal/billing/accounting_service_core.go:275`、`backend/internal/billing/accounting_service.go:146`
- 内容: 規約は監査dependency欠落時も business write を rollback せよと定め、`NewAccountingService` の doc コメント（accounting_service.go:146-149）も「3経路とも fail-closed 化済み」と宣言している。しかし3経路のうち2経路が dependency 欠落を成功扱いにする: `logCreditCorrection` は `if s.auditTx == nil { return nil }`（accounting_service_correction.go:178-180）で監査ゼロのまま確定済み会計のカード金額訂正を commit させ、`Cancel` は `if s.auditTx != nil` で囲うため（accounting_service_reports.go:93-106）nil なら status=cancelled だけが監査なしで確定する。正しい形は同 package の `logPostCloseEdit` にあり、こちらは `return apperrors.WrapInternalServerError(...)` で拒否する（accounting_service_core.go:275-277）。現行の composition root は値型 `billingAuditTxBridge` を渡すため interface が nil になることは production では起きず、実害は潜在（契約の穴）に留まる — ただし規約が要求する挙動と実装が逆であり、別配線・別 composition では静かに監査が消える。
- 修正: `logCreditCorrection` の nil 分岐を `logPostCloseEdit` と同じ `apperrors.WrapInternalServerError` 返却に置換し、`Cancel` の `if s.auditTx != nil` 条件を外して nil 時にエラー返却する（3経路の dependency 欠落挙動を1つのヘルパーへ寄せる）。
- round3-review(2026-07-28): **UPHELD** — logCreditCorrection/Cancel が auditTx nil を成功扱い。invariants:37。production は bridge 注入で潜在。
### infra 横断（httpapi / middleware / apperrors / audit ほか）（5件）

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
#### INF-01: 未分類のPostgreSQLサーバエラーが全て HTTP 400「入力値が正しくありません」に落ちる（5xxが計上されずサイレント障害化） — HIGH
- 区分: 新規
- 規約: `.claude/rules/go-gin-backend-guidelines.md:166` 「- 既知の application error を安定した HTTP status/code に変換し、未知の error は汎用的な 500 response にする。」
- 対象: `backend/internal/httpapi/response.go:87`、`backend/internal/httpapi/response.go:88`、`backend/internal/httpapi/response.go:89`、`backend/internal/httpapi/response_pg.go:10`、`backend/internal/httpapi/response_pg.go:28`、`backend/internal/httpapi/response_pg.go:44`、`backend/internal/apperrors/errors.go:47`、`backend/internal/apperrors/errors.go:195`、`backend/internal/middleware/logging.go:32`、`backend/internal/middleware/logging.go:34`
- 内容: ResolveErrorResponse は `case isPgError(err)` を default(500) の直前に置き、pgconn.PgError がエラーチェーンに含まれれば無条件に 400 を返す（response.go:87-89）。classifyPgError が明示的に扱うのは 23503/23505/22003/22P02/23514 の5コードだけで、それ以外は switch を抜けて response_pg.go:44 の汎用文言になる。apperrors.FromGORM の最終行 `Wrap(err, "database error")`(errors.go:195) は `%w`(errors.go:47) で PgError を保存するため、repository 側で分類されなかった DB エラーも同じ経路に合流する。結果、40001(serialization_failure)・40P01(deadlock_detected)・55P03(lock_not_available)・57P01(admin_shutdown)・53300(too_many_connections)・42P01(undefined_table) といったサーバ側障害が、クライアントには「入力値が正しくありません」の 400 として返る。
- 修正: classifyPgError を「クライアント起因と確定できるコードの allowlist」に限定し、isPgError 分岐を『allowlist に一致した場合のみ 400』へ変更する。一致しない PgError は default 分岐へ落として 500 とし、既存の未分類 error と同じく詳細非露出の汎用 500 にする。

- 再実測(2026-07-28): **ALREADY-FIXED** — httpapi/response.go:90-107 が未知 PgError を 500 に落とし、classifyPgError は allowlist のみ known=true（response_pg.go:34-54）。BUG-2026-07-27-01 コメント付き。
- round3-review(2026-07-28): **WITHDRAWN** — ALREADY-FIXED。`httpapi/response.go:90-107` が未知 PgError を 500 に落とし、`classifyPgError` は allowlist のみ known=true（BUG-2026-07-27-01）。欠陥パスは HEAD に存在しない。
#### INF-02: SanitizeNullBytes の透過 Reader が http.MaxBytesReader の body 上限を無効化し、制御バイトのみの巨大 body が無制限に読まれる — HIGH
- 区分: 新規 ／ 横断パターン: X-07
- 規約: `.claude/rules/go-gin-backend-guidelines.md:180` 「- rate limit、request/body/upload size、content type、file path を制限する。」

- 対象: `backend/internal/middleware/sanitize_null_bytes.go:64`、`backend/internal/middleware/sanitize_null_bytes.go:65`、`backend/internal/middleware/sanitize_null_bytes.go:84`、`backend/internal/middleware/sanitize_null_bytes.go:95`、`backend/internal/middleware/sanitize_null_bytes.go:108`、`backend/internal/auth/http_binding.go:30`、`backend/internal/staff/http_binding.go:30`、`backend/internal/billing/billing_confirmation_handler.go:135`、`backend/internal/lstep/lstep_csv_import_handler.go:75`
- 内容: middleware は POST/PATCH/PUT の非バイナリ body を `sanitizedBodyReader` で包み(:64)、同時に `c.Request.ContentLength = -1` を設定する(:65)。sanitizedBodyReader.Read は除去後のバイト数 writeIndex を返し、全バイトが除去対象なら writeIndex==0 のまま source を読み直すループに入る(:85-98)。下流ハンドラが後から重ねる `http.MaxBytesReader` は「ラップした Reader が返した n」でのみ残量を減算するため、除去されたバイトは一切カウントされず上限に達しない。ContentLength も -1 にされているため事前の長さ判定も効かない。
- 修正: sanitizedBodyReader に呼び出し側から上限バイト数（読み捨て分を含む消費バイト総量）を渡し、超過時に error を返して打ち切る。あるいは SanitizeNullBytes をハンドラ側 MaxBytesReader より内側（下流）に適用する順序へ変更し、除去前の生バイト数が MaxBytesReader を通過するようにする。ContentLength の -1 上書きも、上限判定を持つ層より前で行わないようにする。
- 検証時の補正: 影響を受けるのは JSON binder 3経路（backend/internal/auth/http_binding.go:26,30 / backend/internal/staff/http_binding.go:26,30 / backend/internal/billing/billing_confirmation_handler.go:131,135）に限る。evidence に挙げられた backend/internal/lstep/lstep_csv_import_handler.go:75 は multipart/form-data であり、sanitize_null_bytes.go:52-55 の isBinaryContentType 早期 return によって middleware を通過しない（ContentLength は保持され MaxBytesReader は正常動作する）。したがって同行は本欠陥の evidence から外すべきである。

- 再実測(2026-07-28): **CONFIRMED** — sanitize_null_bytes.go:64-65 ContentLength=-1、:84-97 が sanitization 後バイトのみを計上。MaxBytesReader 無効化は継続。
- round3-review(2026-07-28): **UPHELD** — sanitize_null_bytes.go:64-65 ContentLength=-1、:84-98 が sanitization 後 n のみ計上。MaxBytesReader 無効化継続。規約は guidelines:180（body size）が正（:179 は LINE_DRIFT）。
#### INF-03: Lステップ API の URL パスに line_user_id を無エスケープで埋め込んでおり、リクエスト由来文字列でパス/クエリを改変できる — MEDIUM
- 区分: 新規
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「- 外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/infra/lstep/user.go:33`、`backend/internal/infra/lstep/tag.go:48`、`backend/internal/infra/lstep/client.go:81`、`backend/internal/owner/http_request.go:269`、`backend/internal/owner/http_owner.go:149`、`backend/internal/owner/service_line.go:13`、`backend/internal/owner/service_line.go:28`
- 内容: httpLstepClient は `fmt.Sprintf("/contacts/%s", lineUserID)`(user.go:33) / `fmt.Sprintf("/contacts/%s/tags", lineUserID)`(tag.go:48) でパスを組み立て、`c.baseURL+path`(client.go:81) をそのまま URL にする。url.PathEscape は使われていない。入口の `patchOwnerLineUserIDRequest.LineUserID`(owner/http_request.go:269) には binding tag が無く、LinkLineUserID(owner/service_line.go:13-28) も形式検証をしないため、owners:edit 権限のあるスタッフが任意文字列を保存できる。scheme://host は baseURL 側で確定しているためホスト差し替えは起きないが、dot-segment traversal と `?`/`#` 注入により、クリニックの API キーを載せた外部リクエストの宛先パス・クエリを操作できる。
- 修正: user.go:33 / tag.go:48 の埋め込みを `url.PathEscape(lineUserID)` にする（infra 側の fail-safe）。併せて owner/http_request.go:269 に LINE user ID の形式 binding（例 `omitempty,max=64` + 英数字制限）を追加して入口でも閉じる。
- 検証時の補正: 現時点で実際に HTTP を発行する経路は tag.go:48（GetUserTags、呼び出し元 = backend/internal/lstep/lstep_tag_service.go:170 で `*owner.LineUserID` を渡す）のみである。user.go:33 の GetUser は infra/lstep/client.go:39 の interface 宣言以外に非 test の呼び出し元が無く、AddTag(tag.go:18)/RemoveTag(tag.go:31)/SetProperty(user.go:65) は「L-step write operations are paused by policy」の無通信スタブであるため現状では到達しない。fail-safe としての PathEscape 適用対象は 2 箇所のままでよいが、今日の実害面は tag.go:48 に限られる。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — line_user_id PathEscape なし。guidelines:151。実害は tag.go:48。
#### INF-04: apperrors.FromGORM が pgx ドライバエラーを err.Error() の部分文字列一致で分類している — MEDIUM
- 区分: 新規
- 規約: `.claude/refs/error-handling.md:18` 「- error の種類は `errors.Is` / `errors.As` で判定し、message 文字列比較をしない。」
- 対象: `backend/internal/apperrors/errors.go:174`、`backend/internal/apperrors/errors.go:175`、`backend/internal/apperrors/errors.go:176`、`backend/internal/apperrors/errors.go:177`、`backend/internal/apperrors/errors.go:178`
- 内容: FromGORM は `strings.Contains(errMsg, "unable to encode")` 等 3 本の message 文字列一致で「数値が範囲外です」(400) を返す。規約が禁じる message 文字列比較そのものであり、pgx 側の文言変更で分類が黙って外れる（その場合 IN-A-01 の経路に落ちて別文言の 400 になり、テストがなければ検知されない）。逆に無関係なエラー文言が偶然一致すると誤って 400 として扱われる。なお pgx の encode 失敗は型付き sentinel を公開していないため、この逸脱は現状回避不能である（rule_defects 1 参照）。
- 修正: pgx が公開する型（例 *pgconn.PgError 以外の driver エラー型）を errors.As で判定できるか upstream を確認し、可能なら型判定へ置換する。不可能な場合は、この 3 本を「規約 error-handling.md:18 の明示的例外」として ADR または application invariant に根拠・適用範囲・検知テストを記録し（guidelines:233 の要求形式）、文言変更を捕捉する回帰テストを付ける。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **DOWNGRADED** → severity **LOW** — message 文字列比較は error-handling.md:18 違反だがメンテ脆弱性中心。
#### INF-06: audit.MarshalJSON が marshal 失敗を無記録で握り潰し、監査ログの old/new/metadata が欠落しても観測できない — LOW
- 区分: 新規 ／ 横断パターン: X-04
- 規約: `~/.claude/rules/ecc/common/coding-style.md:49` 「- Never silently swallow errors」
- 対象: `backend/internal/audit/repository.go:54`、`backend/internal/audit/repository.go:58`、`backend/internal/audit/repository.go:59`、`backend/internal/audit/repository.go:60`、`backend/internal/audit/service.go:128`、`backend/internal/audit/service.go:129`、`backend/internal/audit/service.go:130`
- 内容: MarshalJSON は json.Marshal 失敗時に nil を返すだけで、log 出力も metrics も残さない(repository.go:58-61)。buildLog は OldValue/NewValue/Metadata の3フィールド全てをこの関数経由で埋める(service.go:128-130)ため、marshal に失敗した監査は「値が無かった監査」と外形上区別できないレコードとして commit される。error-handling.md:9 の `明示的に回復する` に該当する設計判断ではあるが、回復したこと自体が観測不能な点が問題である。根拠区分は project quality policy（Go/Gin 公式要件ではない）。
- 修正: MarshalJSON 内で marshal 失敗時に slog.Warn を1行出す（値そのものは出さず、フィールド名と型名のみ）。監査本体を止めない現行契約は維持しつつ、欠落が事後に追跡可能な状態にする。
- round3-review(2026-07-28): **REFRAMED** — MarshalJSON の omit は意図的 best-effort。残るのは observability のみ。project quality policy。
### lstep（LINE / Lステップ連携）（22件）

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
#### LSA-01: lstep_base_url が無検証で保存され、復号済み Lステップ API キーが任意ホストへ Bearer 送信される（保管secretの持ち出し + SSRF） — HIGH
- 区分: 新規
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/lstep/lstep_settings_request.go:6`、`backend/internal/lstep/lstep_settings_update.go:17`、`backend/internal/lstep/lstep_settings_update.go:25`、`backend/internal/lstep/lstep_settings_connection.go:36`、`backend/internal/lstep/lstep_settings_connection.go:64`、`backend/internal/lstep/lstep_settings_connection.go:69`、`backend/internal/lstep/lstep_settings_connection.go:70`、`backend/internal/lstep/lstep_settings_credentials.go:43`、`backend/internal/lstep/lstep_tag_service.go:130`
- 内容: `lstep_base_url` は request DTO に binding tag が無く（:6）、service 側でも scheme/host の検証や allowlist 無しに `ClinicIntegration` へ Upsert される。`testLstepAPI` はその値をそのまま連結して `Authorization: Bearer <復号済み lstep API key>` を送る。GET 応答では同キーが `crypto.MaskValue` でマスクされる設計であるにもかかわらず、この経路で平文のまま外部へ出る。
- 修正: `lstep_base_url` を `url.Parse` で検証し、https スキーム固定 + 許可ホスト（既定 `lstep.DefaultBaseURL` のホスト）の allowlist を境界で強制する。到達先が allowlist 外なら 400 で拒否し、`testLstepAPI` / `lstep.NewClient` は検証済み URL のみ受け取る。併せて `http.DefaultClient` を timeout 付き client へ置換する。
- 検証時の補正: evidence の lstep_settings_update.go:17 は `{model.IntegrationKeyLstepAPIKey, input.LstepAPIKey},` であり、base URL の pair は :18。無検証 Upsert の実体は :25-:43 のループ。指摘の substance には影響しない off-by-one。

- 再実測(2026-07-28): **CONFIRMED** — lstep_base_url 無検証保存 + testLstepAPI が任意 baseURL へ Bearer 送信は継続。
- round3-review(2026-07-28): **UPHELD** — lstep_base_url 無検証保存 + 復号 API キーを任意 host へ。guidelines:151 + secret 境界。
#### LSA-02: 自動配信トリガーの除外判定が owner.LstepOptOut を読まず、オプトアウト済み飼い主へ配信タグが付与され得る — HIGH
- 区分: 新規
- 規約: `backend/CODING_RULES.md:42` 「自動status transitionは、対象条件をwrite時にcompare-and-setで再評価する。臨床記録など遷移を否定するbusiness evidenceも同じ判定へ含め、resource単位の監査が必須なら状態変更と同じtransactionでfail-closedにする。同じevidenceを逆向きに変更する競合workflowがある場合は、両者を同じresource-scoped serialization機構へ参加させ、各writeのcommitまで順序を保持する。」
- 対象: `backend/internal/lstep/lstep_delivery_trigger_state.go:11`、`backend/internal/lstep/lstep_delivery_trigger_state.go:12`、`backend/internal/lstep/lstep_delivery_trigger_state.go:15`、`backend/internal/lstep/lstep_delivery_trigger_state.go:23`、`backend/internal/lstep/lstep_delivery_trigger_batch.go:94`、`backend/internal/lstep/lstep_tag_sync_pet_exclusion.go:47`、`backend/internal/lstep/line_send_service.go:147`、`backend/internal/lstep/checkup_sync_service_create.go:127`、`backend/internal/lstep/lstep_tag_sync_service.go:300`
- 内容: `checkExclusion` は `owner.DeliveryExcluded` / `LineUserID` / `EXCL_配信停止` タグの3点しか見ず、配信を否定する第一級の business evidence である `owner.LstepOptOut` を write 時に再評価しない。opt-out の反映は `SyncExclusionTags` が付ける EXCL タグ経由の間接依存で、そのタグ付与自体が best-effort（LINE API 失敗時は付かない）かつバッチ非同期である。同じ判定を line_send / checkup_sync / tag_sync の3経路はいずれも `LstepOptOut` 直読で行っており、配信トリガーだけが例外になっている。
- 修正: `checkExclusion` の先頭に `if owner.LstepOptOut { return true, "lstep_opt_out", nil }` を追加し、EXCL タグ経由の間接判定を多層防御に降格する。除外理由 enum に `lstep_opt_out` を追加して delivery-monitor の内訳集計にも現れるようにする。
- 検証時の補正: detail の「opt-out の反映は SyncExclusionTags が付ける EXCL タグ経由の間接依存」は正しいが不足。HandleOwnerOptOut は SyncExclusionTags を呼ばず、逆に lstep_lifecycle_service.go:325 の DeleteAllByOwner で EXCL_配信停止 の cache 行ごと削除するため、間接依存は「best-effort で不確実」ではなく「opt-out 経路では確定的に無効化される」。なお規約適合は CODING_RULES.md:42 の中間節（遷移を否定する business evidence を同じ判定へ含める）に依拠しており、.claude/refs/backend-application-invariants.md:36「必須依存が欠ける場合はwrite前にfail-closed」も同等以上に妥当な根拠。

- 再実測(2026-07-28): **CONFIRMED** — checkExclusion が LstepOptOut を未参照。LSB-01 と同一 root cause（修正は LSB-01 に一本化可）。
- round3-review(2026-07-28): **REFRAMED** — 実害は成立するが CODING_RULES.md:42（status transition CAS）適用は拡張解釈。正本は LSB-01（business-evidence fail-open）。修正は LSB-01 に畳む。
#### LSA-03: visit-dormant バッチが DB エラーを握り潰して (0, nil) を返し、durable scheduler と audit_logs の双方から失敗が消える — HIGH
- 区分: 新規 ／ 横断パターン: X-04
- 規約: `.claude/refs/error-handling.md:9` 「error を無視しない。処理できる境界まで返すか、明示的に回復する。」
- 対象: `backend/internal/lstep/lstep_batch_segmentation.go:22`、`backend/internal/lstep/lstep_batch_segmentation.go:23`、`backend/internal/lstep/lstep_batch_segmentation.go:24`、`backend/internal/lstep/lstep_batch_segmentation.go:25`、`backend/internal/lstep/lstep_batch_service.go:276`、`backend/internal/lstep/lstep_batch_service.go:284`、`backend/internal/lstep/lstep_batch_service.go:296`
- 内容: `syncVisitDormantForClinic` は `FindDormantOwnerEntries` の失敗時に slog のみ出して `return 0, nil` する（コメントも「静かにスキップする（挙動保存）」と明記）。呼び出し元 `runBatchAllClinicsWithResult` は `count>0 || len(errs)>0` でしか audit を書かず、`result.add(count, len(errs))` にも 0 が積まれるため、当該クリニックの VISIT_* タグ同期が全滅しても BatchRunResult.Failed は増えず audit_logs にも行が残らない。
- 修正: `return 0, []error{apperrors.Wrap(findErr, "failed to find visit dormant entries")}` へ変更し、同型の「ログのみで握り潰す」分岐が他の perClinic 実装に無いことを併せて確認する。

- 再実測(2026-07-28): **LINE-DRIFT** — return 0,nil は lstep_batch_segmentation.go:22-25 で継続。result.add 呼び出しは lstep_batch_service.go:277（旧:276）。defect 自体は未修正。
- round3-review(2026-07-28): **UPHELD** — visit-dormant が DB エラーで (0,nil)。error-handling.md:9 / X-04。
#### LSA-04: lstep タグ設定3テーブルが clinic_id を持たない全院共有行なのに、clinic-scoped RBAC ルートから任意院が作成・削除できる — HIGH
- 区分: 新規
- 規約: `backend/migrations/CLAUDE.md:14` 「**clinic_id スコープ**: 新テーブルにクリニック間分離が必要な場合は `clinic_id NOT NULL` を付ける」
- 対象: `backend/migrations/001_init.sql:619`、`backend/migrations/001_init.sql:635`、`backend/migrations/001_init.sql:649`、`backend/internal/lstep/lstep_tag_config_repository.go:37`、`backend/internal/lstep/lstep_tag_config_repository.go:51`、`backend/internal/lstep/lstep_tag_config_repository.go:63`、`backend/internal/lstep/lstep_tag_config_repository.go:77`、`backend/internal/lstep/lstep_tag_config_repository.go:89`、`backend/internal/lstep/lstep_tag_config_repository.go:103`、`backend/internal/lstep/routes.go:153`、`backend/internal/lstep/routes.go:156`、`backend/internal/lstep/routes.go:160`、`backend/internal/lstep/routes.go:164`
- 内容: `lstep_auto_managed_prefixes` / `lstep_condition_tag_mappings` / `lstep_send_purpose_tag_prefixes` は DDL に clinic_id が無い（001_init.sql:619/635/649）。一方 route は `ResourceHospitalSettings` という院単位 permission で GET/POST/DELETE を公開しており、repository 側の Find/Create/Delete はいずれも clinic 述語を持たない（`Delete(&model.LstepAutoManagedPrefix{}, id)` は id のみ）。結果、A院の設定管理者が prefix を追加すると B院でもその prefix が「自動管理タグ」となり手動付与が拒否され、A院が削除すると B院のペット死亡時タグ掃除（Category=="C2"）が壊れる。description 欄も他院から読める。
- 修正: 設計意図を先に確定する。全院共通マスタが正なら route を院管理者から外して platform-admin 専権にする。院ごとに持たせるのが正なら `clinic_id NOT NULL` + `(clinic_id, prefix)` UNIQUE の incremental migration を追加し、repository 全メソッドへ clinic 述語を入れる。
- 検証時の補正: rule_citation の backend/migrations/CLAUDE.md:14「**clinic_id スコープ**: 新テーブルにクリニック間分離が必要な場合は `clinic_id NOT NULL` を付ける」は逐語で実在するが、対象は既存 DDL（001_init.sql、mig-010 由来）であり「新テーブル」条項の適用は弱い。主根拠は .claude/refs/backend-application-invariants.md:11「clinic-scoped data のすべての read/write/delete は、認証済み `clinic_id` で制約する。」および routes.go:154-164 の認可境界（全院共有マスタを院単位 RBAC で mutate 可能にしている点）に置き換えるべき。また evidence の 001_init.sql:619/635/649 は各テーブルの先頭列行で、CREATE TABLE 自体は 618/634/648。category を B/C1/C2/C3 と定義する :630 は逐語一致。

- 再実測(2026-07-28): **CONFIRMED** — 3 テーブル無 clinic_id + clinic RBAC 経由 Create/Delete 継続。
- round3-review(2026-07-28): **REFRAMED** — global master 無 clinic_id は設計判断（boundary-map）。欠陥は hospital-settings RBAC で全院共有 master を mutate できる認可境界不一致。
#### LSA-05: 飼い主 opt-out / opt-in / 削除クリーンアップが監査ログを一切残さず、全タグキャッシュを無記録で破棄する — HIGH
- 区分: 新規 ／ 横断パターン: X-03
- 規約: `.claude/refs/backend-application-invariants.md:31` 「destructive または irreversible な操作には、権限、対象 scope、監査、recovery 方針を持たせる。」
- 対象: `backend/internal/lstep/lstep_lifecycle_service.go:239`、`backend/internal/lstep/lstep_lifecycle_service.go:248`、`backend/internal/lstep/lstep_lifecycle_service.go:258`、`backend/internal/lstep/lstep_lifecycle_service.go:266`、`backend/internal/lstep/lstep_lifecycle_service.go:286`、`backend/internal/lstep/lstep_lifecycle_service.go:325`、`backend/internal/lstep/lstep_lifecycle_handler.go:61`、`backend/internal/lstep/lstep_lifecycle_handler.go:89`、`backend/internal/lstep/lstep_lifecycle_deps.go:54`
- 内容: 同一 service 内の `HandlePetDeath` / `HandlePetRevival` は同一 tx の fail-closed 監査を持つのに、`HandleOwnerOptOut` / `HandleOwnerOptIn` / `HandleOwnerDeletion` は `lifecycleOperationAuditor` を一度も呼ばない。handler 側も actorID を取り出さないため誰が配信同意状態を変えたか復元不能。さらに opt-out 経路は `removeAllTagsFromLstep` 内で `DeleteAllByOwner` により当該飼い主のタグキャッシュ行を全削除する（不可逆）。
- 修正: handler で `httpapi.OptionalStaffID(c)` を取得し service へ渡す。`HandleOwnerOptOut` / `HandleOwnerOptIn` / `HandleOwnerDeletion` に `LogLstepOperation`（最低限 best-effort、可能なら opt-out 書込と同一 tx の `LogEntryTx`）を追加し、tag cache 全削除の前後を監査対象にする。
- 検証時の補正: detail の「同一 service 内の HandlePetDeath / HandlePetRevival は同一 tx の fail-closed 監査を持つ」は誤り。:174-177 / :230-233 はいずれもコメント「監査ログ（best-effort）」付きで、失敗時は slog.WarnContext のみで継続する best-effort 監査であり、同一 tx でも fail-closed でもない。対比は「fail-closed 監査あり vs なし」ではなく「best-effort 監査あり vs 監査呼び出しゼロ」。

- 再実測(2026-07-28): **CONFIRMED** — opt-out/in/deletion に audit 無し。pet death 側は fail-closed 監査あり（検証時補正の best-effort 記述は陳腐化）。
- round3-review(2026-07-28): **UPHELD** — opt-out/in/delete cleanup に監査ゼロ。invariants:31。
#### LSA-06: PATCH /lstep-settings が clinic_integrations 6行 + lstep_settings + clinic_settings を非トランザクションで逐次 Upsert し、部分適用で 500 を返す — HIGH
- 区分: 新規 ／ 横断パターン: X-06
- 規約: `backend/CLAUDE.md:33` 「1つのbusiness graphを構成する複数rowのwriteは同じtransactionで原子的に扱い、commit済みの成功を後段の再取得errorで失敗応答へ反転させない。」
- 対象: `backend/internal/lstep/lstep_settings_service.go:278`、`backend/internal/lstep/lstep_settings_service.go:279`、`backend/internal/lstep/lstep_settings_service.go:282`、`backend/internal/lstep/lstep_settings_service.go:287`、`backend/internal/lstep/lstep_settings_update.go:25`、`backend/internal/lstep/lstep_settings_update.go:39`、`backend/internal/lstep/lstep_settings_update.go:51`
- 内容: `UpdateSettings` は `updateIntegrationCredentials`（key ごとに独立 Upsert を最大6回）→ `updateSyncEnabled` → `updateClinicSyncConfig`（さらに4種の独立 Update）を ambient transaction 無しで順に呼ぶ。途中で失敗すると API は 500 を返すが、先行した API キー / channel secret / LIFF ID は既にコミット済みで残る。認証情報の組が不整合な状態（新 API キー + 旧 channel secret）でも LINE Webhook 署名検証と Lステップ同期が走る。
- 修正: `UpdateSettings` 全体を `Transactor.WithTx` で包み、配下の repository を `persistence.DBOrTx` 経由に揃えて全 write を同一 tx に参加させる。
- 検証時の補正: detail の不整合例「新 API キー + 旧 channel secret」は不正確。Lステップ API キーと LINE channel secret は別サービス向けの独立値であり組を成さない。実際に結合しているのは LINE の access token（IntegrationKeyLineChannelAccessToken）と channel secret（IntegrationKeyLineChannelSecret）の対で、前者のみ commit された場合に push は新トークン・Webhook 署名検証は旧 secret という不整合が生じる。また updateIntegrationCredentials は lstep_settings_update.go:26-28 で空文字を skip するため、実際に多重 row が書かれるのは複数フィールドを同時送信した場合に限られる。

- 再実測(2026-07-28): **CONFIRMED** — lstep_settings_service.go:278-289 の逐次 Upsert 非原子は継続。
- round3-review(2026-07-28): **UPHELD** — PATCH settings 多 row が非 tx。CLAUDE.md:33。
#### LSA-07: LINE push 成功ログに LINE User ID（飼い主識別子）をそのまま出力している — HIGH
- 区分: 新規
- 規約: `backend/CODING_RULES.md:68` 「log は構造化し、secret、token、credential、owner/pet/staff/medical data を含めない。」
- 対象: `backend/internal/lstep/line_messaging_service.go:81`、`backend/internal/lstep/line_messaging_service.go:47`
- 内容: `slog.InfoContext(ctx, "LINE push sent", "to", lineUserID)` は飼い主に一意対応する外部識別子を平文でログへ落とす。同 package の他経路（line_send_service / tag_sync / lifecycle）は一貫して `owner_id` のみをログしており、ここだけが例外。
- 修正: `"to", lineUserID` を削除し、呼び出し元から渡す `owner_id`（無い場合は識別子のハッシュ）へ置換する。

- 再実測(2026-07-28): **CONFIRMED** — line_messaging_service.go:81 が to=lineUserID を Info ログ。
- round3-review(2026-07-28): **DOWNGRADED** → severity **MEDIUM** — LINE user id の Info ログは CODING_RULES.md:68 違反だが HIGH の秘密流出級ではない。
#### LSB-01: 自動配信トリガーが owner.LstepOptOut を一切参照せず、オプトアウト済み飼主へLINE配信が発火する（fail-open） — HIGH
- 区分: 新規
- 規約: `backend/migrations/001_init.sql:316` 「true = すべてのタグ付与をスキップ」+ `backend/CLAUDE.md:35` OpenAPI/migration contract（LstepOptOut は配信除外 business evidence）

- 対象: `backend/internal/lstep/lstep_delivery_trigger_state.go:11`、`backend/internal/lstep/lstep_delivery_trigger_state.go:12`、`backend/internal/lstep/lstep_delivery_trigger_state.go:15`、`backend/internal/lstep/lstep_delivery_trigger_state.go:23`、`backend/internal/lstep/lstep_delivery_trigger_batch.go:94`、`backend/internal/lstep/lstep_delivery_trigger_batch.go:113`、`backend/internal/lstep/lstep_lifecycle_service.go:248`、`backend/internal/owner/repository.go:515`、`backend/internal/owner/repository.go:516`、`backend/internal/owner/repository.go:517`、`backend/internal/owner/repository.go:518`、`backend/migrations/001_init.sql:316`、`backend/internal/lstep/lstep_tag_sync_pet_exclusion.go:47`、`backend/internal/lstep/lstep_tag_sync_pet_exclusion.go:48`、`backend/internal/lstep/lstep_tag_sync_service.go:300`、`backend/internal/lstep/line_send_service.go:147`、`backend/internal/pet/repository.go:529`、`backend/internal/medicalrecord/vaccination_repository.go:225`
- 内容: FEAT-383 自動配信バッチ（10:00 JST）の唯一の除外ゲート checkExclusion は owner.DeliveryExcluded・LineUserID空・EXCL_配信停止タグの3条件のみを見ており、owner.LstepOptOut を参照しない。一方 opt-out API（POST /owners/:id/lstep-opt-out、PATCH /owners/:id/lstep/opt-out）が呼ぶ RecordLstepOptOut は lstep_opt_out / _at / _reason の3列のみ更新し delivery_excluded を立てない。DDL の列コメント自体が「true = すべてのタグ付与をスキップ」と宣言しているのに、配信トリガーは applyTagAndLog でタグを付与する。同packageの他の全配信経路（lstep_tag_sync_pet_exclusion.go:47、lstep_tag_sync_service.go:300、line_send_service.go:147、checkup_sync系）は LstepOptOut を検査しており、この1経路だけが判定式のドリフトで欠落している。
- 修正: checkExclusion の先頭に `if owner.LstepOptOut { return true, "lstep_opt_out", nil }` を追加する。恒久対策として lstep_tag_sync_pet_exclusion.go:47 の述語を共有ヘルパへ抽出し、両経路が同一の除外判定を参照するようにして再ドリフトを封じる。併せて opt-out 済み owner が配信対象にならないことの回帰testを lstep_delivery_trigger_service_test.go へ追加する（現状 LstepOptOut への言及ゼロ）。
- 検証時の補正: 規約接地を差し替えるべき。backend/CLAUDE.md:25 は「停止手段を設けること」を要求する規定であり、本バッチには停止手段自体は存在する（owner.DeliveryExcluded / EXCL_配信停止タグ / IsSyncEnabled）。したがって :25 は適合の弱い引用である。より正確な接地は (a) migration宣言contract: backend/migrations/001_init.sql:316 および :352 の逐語「Lステップ配信オプトアウトフラグ。true = すべてのタグ付与をスキップ。」に対し applyTagAndLog がタグを付与している点、(b) backend/CLAUDE.md:35 逐語「- OpenAPI、migration、security ADR との contract を維持する。」— migration が宣言した列contractの不履行。さらに finding本文より強い機序を実測で追加する: HandleOwnerOptOut(lstep_lifecycle_service.go:258) が removeAllTagsFromLstep → tagCacheRepo.DeleteAllByOwner(:325) を呼ぶため、opt-out実行そのものが checkExclusion:23 の唯一の代替ゲートである EXCL_配信停止タグのcache行を消去する。すなわち opt-out した飼主は「LstepOptOut未参照」かつ「EXCL タグcacheも空」となり、配信除外ゲートを二重に素通りする。

- 再実測(2026-07-28): **CONFIRMED** — checkExclusion 未参照 + RecordLstepOptOut が delivery_excluded を立てない + EXCL cache 全消去の複合 fail-open 継続。LSA-02 の canonical。
- round3-review(2026-07-28): **UPHELD** — delivery trigger が LstepOptOut 未参照。business-evidence fail-open（X-04 ではない）。migration 契約と整合。
#### LSB-02: removeAllTagsFromLstep がリモート解除失敗・client nil でもローカルtag cacheを全消去し、再照合の根拠を不可逆に失う — HIGH
- 区分: 新規
- 規約: `.claude/refs/backend-application-invariants.md:35` 「- cross-domain writeはtransaction owner、全参加write、rollback範囲を明示し、部分成功でbusiness factを不整合にしない。意図的なsaga/best-effort処理は、補償、再試行、監査、部分失敗contractを持たせる。」
- 対象: `backend/internal/lstep/lstep_lifecycle_service.go:317`、`backend/internal/lstep/lstep_lifecycle_service.go:318`、`backend/internal/lstep/lstep_lifecycle_service.go:319`、`backend/internal/lstep/lstep_lifecycle_service.go:320`、`backend/internal/lstep/lstep_lifecycle_service.go:321`、`backend/internal/lstep/lstep_lifecycle_service.go:325`、`backend/internal/lstep/lstep_lifecycle_service.go:258`、`backend/internal/lstep/lstep_lifecycle_service.go:259`、`backend/internal/lstep/lstep_lifecycle_service.go:297`、`backend/internal/lstep/lstep_lifecycle_service.go:148`
- 内容: 個々の client.RemoveTag 失敗は :320 でログのみに握り潰され、その後 :325 の DeleteAllByOwner が無条件にローカルcacheを全削除する。また buildClient は同期無効・APIキー未設定時に nil を返し（:78-87）、その場合リモート解除は1件も実行されないのに :325 は同じく実行される。結果としてLステップ側にタグが残ったままローカルの保持記録だけが消え、どのタグを解除すべきだったかを復元できず補償・再試行が不可能になる。best-effort を意図した設計であることは :257 のコメントで明示されているが、規約が best-effort に要求する補償・再試行・監査・部分失敗contractのいずれも備えていない。呼び出し元は opt-out(:258)・飼主削除(:297)・全ペット死亡(:148) の3経路。
- 修正: RemoveTag の失敗を収集し、成功したタグだけを DeleteTag で個別削除する（現行 removePetDerivedTagsFromLstep :346-351 が既に採っている形）。client==nil の場合は cache を消さずに未解除として残す。残存分は再試行可能な状態として保持し、失敗件数を監査メタデータへ記録する。
- 検証時の補正: detail の「どのタグを解除すべきだったかを復元できず補償・再試行が不可能になる」「不可逆に失う」は過大。internal/infra/lstep/client.go:35 に `GetUserTags(ctx, lineUserID) ([]string, error)` が存在し、リモート側のタグ集合は列挙可能である。したがって欠陥は「復元不能」ではなく「規約が best-effort に要求する補償・再試行・監査・部分失敗contractがコード上ゼロである」点に限定して記述すべき。修正案の実装コストが低い（GetUserTags で再照合可能）ことも併記されるべきで、これは所見を弱めるものではなく規約違反の事実は変わらない。

- 再実測(2026-07-28): **CONFIRMED** — removeAllTagsFromLstep :317-327 が remote 失敗でも DeleteAllByOwner 無条件実行。
- round3-review(2026-07-28): **UPHELD** — remote 解除失敗でも local cache 全削除。invariants:35 best-effort contract 欠落。
#### LSB-03: HandlePetDeath がcommit済みのペット死亡記録をcommit後のread errorで失敗応答へ反転させる — HIGH
- 区分: 新規 ／ 横断パターン: X-01
- 規約: `backend/CODING_RULES.md:78` 「- write後の再取得が失敗し得る場合はcommit前の同じtransaction内で行うか、commit済みの成功を後段read errorで失敗へ反転させないcontractにする。」
- 対象: `backend/internal/lstep/lstep_lifecycle_service.go:113`、`backend/internal/lstep/lstep_lifecycle_service.go:130`、`backend/internal/lstep/lstep_lifecycle_service.go:137`、`backend/internal/lstep/lstep_lifecycle_service.go:138`、`backend/internal/lstep/lstep_lifecycle_service.go:140`、`backend/internal/lstep/lstep_lifecycle_service.go:141`、`backend/internal/lstep/lstep_lifecycle_handler.go:33`、`backend/internal/lstep/lstep_lifecycle_handler.go:34`、`backend/internal/lstep/lstep_lifecycle_handler.go:37`、`backend/internal/lstep/lstep_lifecycle_service.go:100`、`backend/internal/lstep/lstep_lifecycle_service.go:101`
- 内容: status/deceased_at 更新と監査は :113-132 の transaction で commit 済みになる。その直後 :137 の FindLivingByOwner が失敗すると :141 で error を返し、handler :33-35 が 204 ではなく error 応答へ写像する。:140 のコメントは「死亡記録の巻き戻しは行わない」と巻き戻さないことを明示しており、まさに規約が禁じる「commit済みの成功を後段read errorで失敗へ反転」に該当する。臨床上の帰結として、獣医師はエラー表示を見て再実行するが、2回目は :100-101 の `pet.Status == deceased` 判定で 409 Conflict「死亡記録は既に登録されています」となり、実際には成功しているのに失敗と矛盾表示の連続になる。
- 修正: transaction commit 後の FindLivingByOwner 以降（タグ再同期・cleanup）は既に best-effort として扱われている他の副作用（:156-177）と同様に、error をログのみに留めて nil を返す。あるいは commit 前に同一 tx 内で生存ペットを取得しておく。

- 再実測(2026-07-28): **CONFIRMED** — HandlePetDeath WithTx commit 後 FindLivingByOwner 失敗で応答反転。
- round3-review(2026-07-28): **UPHELD** — HandlePetDeath の commit 後 FindLiving 失敗で 5xx。CODING_RULES.md:78 / X-01。
#### LSB-04: credential 復号失敗を空文字へ置換して握り潰す3箇所 — Lステップ連携がサイレントに停止し、設定画面は未設定と表示する — HIGH
- 区分: 新規 ／ 横断パターン: X-04
- 規約: `.claude/refs/error-handling.md:9` 「- error を無視しない。処理できる境界まで返すか、明示的に回復する。」
- 対象: `backend/internal/lstep/lstep_settings_credentials.go:36`、`backend/internal/lstep/lstep_settings_credentials.go:37`、`backend/internal/lstep/lstep_settings_credentials.go:38`、`backend/internal/lstep/lstep_settings_credentials.go:39`、`backend/internal/lstep/lstep_settings_credentials.go:47`、`backend/internal/lstep/lstep_settings_service.go:176`、`backend/internal/lstep/lstep_settings_service.go:177`、`backend/internal/lstep/lstep_settings_service.go:178`、`backend/internal/lstep/lstep_settings_service.go:179`、`backend/internal/lstep/lstep_settings_connection.go:25`、`backend/internal/lstep/lstep_settings_connection.go:27`、`backend/internal/lstep/lstep_settings_connection.go:28`、`backend/internal/lstep/lstep_settings_connection.go:41`、`backend/internal/lstep/lstep_lifecycle_service.go:85`、`backend/internal/lstep/lstep_lifecycle_service.go:86`
- 内容: 同一の握り潰しパターンが3箇所に複製されている。decErr は log されるだけで caller へ返らず、val は "" に置換され関数は nil error を返す。帰結は経路ごとに異なりいずれもサイレント: (a) GetRawCredentials 経由では apiKey が "" になり buildClient が :85-87 で nil,nil を返すため「同期無効」と区別できないまま全Lステップ配信が静かに停止する。(b) GetSettings 経由では復号不能な値が空として response に載るため、実際にはDBに値があるのに管理画面が未設定と表示し、管理者の上書き再入力を誘発する。(c) TestConnection 経由では lstepKey が "" になり :41 の疎通probe自体がスキップされ、LstepOK=false・LstepError="" という原因不明の失敗表示になる。暗号鍵ローテーション事故がどの経路でも検知不能になる。
- 修正: 復号失敗は握り潰さず caller へ返す（少なくとも GetRawCredentials と GetSettings は error を伝播する）。運用継続が必要なら「復号不能」を空文字と区別できる明示的な状態として表現し、TestConnection は probe skip ではなく復号失敗を LstepError に区別可能な固定文言で示す。3箇所の重複した kvMap 構築ループは共通ヘルパへ集約する。

- 再実測(2026-07-28): **CONFIRMED** — GetRawCredentials/GetSettings/TestConnection の3箇所で decrypt 失敗→空文字+nil error。
- round3-review(2026-07-28): **UPHELD** — credential 復号失敗を空文字置換。error-handling.md:9 / X-04。
#### LSA-08: 疎通確認 API が外部 HTTP エラーを err.Error() のまま JSON へ返し、到達先 URL・解決 IP・DNS エラーを露出する — MEDIUM
- 区分: 新規 ／ 横断パターン: X-08
- 規約: `backend/CLAUDE.md:34` 「error response と log に secret、credential、個人情報、内部詳細を出さない。」
- 対象: `backend/internal/lstep/lstep_settings_connection.go:44`、`backend/internal/lstep/lstep_settings_connection.go:55`、`backend/internal/lstep/lstep_settings_connection.go:72`、`backend/internal/lstep/lstep_settings_connection.go:91`、`backend/internal/lstep/lstep_settings_response.go:52`、`backend/internal/lstep/lstep_settings_response.go:54`、`backend/internal/lstep/lstep_settings_response.go:99`、`backend/internal/lstep/lstep_settings_handler.go:93`
- 内容: `fmt.Errorf("connection failed: %w", err)` の err は `*url.Error` であり、リクエスト URL と `dial tcp <ip>:<port>: ...` を含む。これが `LstepError` / `LineError` として `lstep_error` / `line_error` JSON フィールドに素通しされる。LS-A-01 の base URL 無検証と組み合わさると、内部ネットワークの到達性オラクルとして機能する。
- 修正: `LstepError` / `LineError` には安定した分類コード（`unauthorized` / `unreachable` / `timeout` など）だけを入れ、詳細は server-side ログのみに残す。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — 疎通確認が raw err.Error を JSON へ。CLAUDE.md:34。
#### LSA-09: LINE 送信失敗時に外部 API の生エラー文字列を 502 応答と送信履歴 API の双方へ露出している — MEDIUM
- 区分: 新規 ／ 横断パターン: X-08
- 規約: `backend/CLAUDE.md:34` 「error response と log に secret、credential、個人情報、内部詳細を出さない。」
- 対象: `backend/internal/lstep/line_send_service.go:169`、`backend/internal/lstep/line_send_service.go:170`、`backend/internal/lstep/line_send_service.go:180`、`backend/internal/lstep/line_send_service.go:195`、`backend/internal/lstep/line_send_response.go:19`、`backend/internal/lstep/line_send_response.go:33`、`backend/internal/lstep/line_send_handler.go:80`
- 内容: `sendErr.Error()` が (a) `WrapBadGateway(fmt.Sprintf("LINE送信に失敗しました: %s", ...))` として即時応答に、(b) `LineSendLog.ErrorMessage` として DB に永続化され `GET /owners/:id/line/send-logs` の `error_message` で再露出される。LINE Messaging API のエラー本文には endpoint・request id・宛先関連の内部詳細が含まれ得る。
- 修正: 応答は固定文言 + 安定コードに丸め、`ErrorMessage` へは正規化した分類値だけを保存する。生エラーは slog へ限定する。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — LINE 送信失敗の raw error を 502 と履歴へ。CLAUDE.md:34。
#### LSA-10: GetOwnerTags のキャッシュフォールバックが DB エラーを握り潰し、200 + tags:[] を返して「タグ無し」と区別できなくする — MEDIUM
- 区分: 新規 ／ 横断パターン: X-04
- 規約: `.claude/refs/error-handling.md:9` 「error を無視しない。処理できる境界まで返すか、明示的に回復する。」
- 対象: `backend/internal/lstep/lstep_tag_service.go:158`、`backend/internal/lstep/lstep_tag_service.go:159`、`backend/internal/lstep/lstep_tag_service.go:160`、`backend/internal/lstep/lstep_tag_service.go:161`、`backend/internal/lstep/lstep_tag_service.go:162`、`backend/internal/lstep/lstep_tag_handler.go:33`
- 内容: Lステップ API 未設定時のフォールバック経路で `tagCacheRepo.FindByOwner` が失敗すると、slog を出して `return result, nil`（Tags は空スライスのまま）となる。同 method の API 経路は失敗時に 500 を返しており、フォールバック経路だけが fail-open。UI 上「この飼い主はタグ0件」と誤表示され、配信対象判断の誤りに直結する。
- 修正: `return nil, apperrors.Wrap(cacheErr, "failed to load lstep tag cache")` に変更し、API 経路と同じ失敗契約へ揃える。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — GetOwnerTags キャッシュフォールバックが DB エラー握り潰し。error-handling.md:9。
#### LSA-11: removeStaleTagsByPrefixes がタグキャッシュ読み取り失敗を「API失敗なし」として返し、同一カテゴリ1タグ保持ルールが黙って破れる — MEDIUM
- 区分: 新規 ／ 横断パターン: X-04
- 規約: `.claude/refs/error-handling.md:9` 「error を無視しない。処理できる境界まで返すか、明示的に回復する。」
- 対象: `backend/internal/lstep/lstep_tag_sync_api.go:15`、`backend/internal/lstep/lstep_tag_sync_api.go:16`、`backend/internal/lstep/lstep_tag_sync_api.go:17`、`backend/internal/lstep/lstep_tag_sync_api.go:18`、`backend/internal/lstep/lstep_tag_sync_api.go:19`、`backend/internal/lstep/lstep_tag_sync_api.go:20`
- 内容: 戻り値 `apiFailed bool` は `false` = 正常を意味するが、cache 読み取り失敗時も `false` を返す。呼び出し元は古いタグを1件も解除しないまま新タグを付与するため、`next_visit_*` / `checkup_done_*` 等が複数世代同時に残り、Lステップ側のセグメント配信が誤対象へ届く。エラーは呼び出し元にも BatchRunResult にも伝わらない。
- 修正: シグネチャを `(apiFailed bool, err error)` に拡張して cache 読み取り失敗を呼び出し元へ返すか、少なくとも `apiFailed=true` を返して後続の付与をスキップさせる。
- 検証時の補正: detail が挙げる影響タグ名 `next_visit_*` / `checkup_done_*` は実際の呼び出し元の prefix と一致しない。実測の対象 prefix は lstep_tag_sync_vaccine.go:49/:123 の `vaccine_dog_` / `vaccine_cat_` / `vaccine_rabies_` と lstep_tag_sync_visit.go:50 の `ltv_amount_` / `visit_count_annual_` / `first_visit_` / `last_visit_` の計7種。加えて未指摘の副作用がある — 3呼び出し元とも末尾で `if !apiFailed { s.notifyAPISuccess(...) }` を実行するため（vaccine.go:56-58 / :130-132、visit.go 同型）、cache 読み取り失敗時には「API 復旧」通知タグまで誤って付与される。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — removeStaleTagsByPrefixes が cache 失敗を非 API 失敗扱い。error-handling.md:9。
#### LSA-12: 配信トリガーログの excluded/failed への status 更新失敗が non-fatal 扱いで、ログが scheduled のまま残り監視集計が実態と乖離する — MEDIUM
- 区分: 新規 ／ 横断パターン: X-04
- 規約: `.claude/refs/error-handling.md:9` 「error を無視しない。処理できる境界まで返すか、明示的に回復する。」
- 対象: `backend/internal/lstep/lstep_delivery_trigger_batch.go:106`、`backend/internal/lstep/lstep_delivery_trigger_batch.go:107`、`backend/internal/lstep/lstep_delivery_trigger_batch.go:108`、`backend/internal/lstep/lstep_delivery_trigger_client.go:17`、`backend/internal/lstep/lstep_delivery_trigger_client.go:18`、`backend/internal/lstep/lstep_delivery_trigger_client.go:19`、`backend/internal/lstep/lstep_delivery_monitor_service.go:93`、`backend/internal/lstep/lstep_delivery_monitor_service.go:96`
- 内容: `recordTrigger` で status=scheduled の行を作った後、excluded 化（batch.go:107）と failed 化（client.go:18）はいずれも失敗時に WarnContext のみで継続する。結果、実際には送っていない／送信に失敗した配信が delivery-monitor 上では Scheduled として数えられ、Excluded/Failed の内訳がゼロのまま運用者に見える。`ExistsTodayByOwnerAndType` は status を見ないため翌日以降の再送も抑止される。
- 修正: status 更新失敗を呼び出し元へ返して perClinic の errs に積み、BatchRunResult.Failed と audit metadata に反映する。少なくとも failed 化の失敗は fatal とする。
- 検証時の補正: detail の2点を訂正。(1)「ExistsTodayByOwnerAndType は status を見ないため翌日以降の再送も抑止される」は誤り。lstep_delivery_trigger_log_repository.go:74-85 は dayStart..dayEnd（`scheduled_at >= ? AND scheduled_at < ?`）で当日のみに限定するため、翌日の再送は抑止されない。抑止されるのは同日中の再試行だけ。(2)「送信に失敗した配信が…Failed の内訳がゼロのまま」は監視画面についてのみ正しい。applyTagAndLog は lstep_delivery_trigger_client.go:21 で `apperrors.Wrap(err, "failed to add lstep tag")` を返し、batch.go:113-115 → runBatch:36-38 経由で errs に積まれるため、BatchRunResult.Failed と audit metadata には反映される。欠落するのは delivery-monitor の status 別内訳のみ。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **DOWNGRADED** → severity **LOW** — status 更新失敗は観測歪み。タグ失敗自体は batch Failed に載る。翌日抑止主張は過大。
#### LSA-13: trigger_type が列挙値検証されず、存在しないトリガー種別の優先順位が 200 で保存されて永久に無視される — MEDIUM
- 区分: 新規 ／ 横断パターン: X-02
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/lstep/lstep_trigger_priority_request.go:5`、`backend/internal/lstep/lstep_trigger_priority_handler.go:62`、`backend/internal/lstep/lstep_trigger_priority_service.go:77`、`backend/internal/lstep/lstep_trigger_priority_service.go:78`、`backend/internal/lstep/lstep_trigger_priority_service.go:87`、`backend/internal/lstep/lstep_trigger_priority_service.go:55`
- 内容: `updateTriggerPriorityItemRequest.TriggerType` は `binding:"required"` のみで `oneof` が無く、service 側も `priority < 1` しか検査せずに `UpsertBatch` する。一方 read 側 `GetByClinicID` は `model.AllTriggerTypes()` を基準に補完するため、綴り違いで登録された行は一覧に現れず `GetPriorityFor` からも参照されない。運用者は 200 応答で設定できたと誤認するが、Q23 の配信抑制は既定値のまま動く。
- 修正: `binding:"required,oneof=..."` もしくは service 冒頭で `model.AllTriggerTypes()` に対する membership 検査を行い、未知の trigger_type は 400 で拒否する。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — trigger_type 列挙未検証。guidelines:151。
#### LSA-14: 自動管理プレフィックスの category が列挙値検証されず、C2 以外の綴りで登録するとペット死亡時のタグ掃除が静かに効かなくなる — MEDIUM
- 区分: 新規 ／ 横断パターン: X-02
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/lstep/lstep_tag_config_request.go:5`、`backend/internal/lstep/lstep_tag_config_handler.go:30`、`backend/internal/lstep/lstep_tag_config_service.go:71`、`backend/internal/lstep/lstep_tag_config_service.go:77`、`backend/internal/lstep/lstep_lifecycle_service.go:371`、`backend/internal/lstep/lstep_lifecycle_service.go:372`、`backend/migrations/001_init.sql:630`
- 内容: DDL コメントは category を「B / C1 / C2 / C3」と定義し（001_init.sql:630）、`loadPetDerivedPrefixes` は `p.Category == "C2"` の完全一致でペット由来タグを選ぶ。しかし request DTO は `binding:"required"` のみ、service も無検証で Create する。`"c2"` や `"C2 "` を登録すると 201 が返るが、死亡ペット由来の vaccine_/checkup_done_ タグは解除されずフォールバック値だけが使われる。
- 修正: `binding:"required,oneof=B C1 C2 C3"` を付与し、既存行に対しては migration での正規化または起動時の drift 検出 gate を追加する。
- 検証時の補正: detail の帰結記述が不正確。`"c2"` 等で登録された行が無視されても、既存の正しい C2 行があれば loadPetDerivedPrefixes は DB 値を返すためフォールバックには落ちず、`vaccine_` / `checkup_done_` の掃除は従来どおり機能する。実際の失敗は「新規追加したプレフィックスが category 綴り違いのため死亡ペット時の掃除対象から静かに漏れる」ことであり、既存タグの掃除が壊れるわけではない。フォールバック（lstep_lifecycle_service.go:376-380、petDerivedPrefixFallback）が使われるのは C2 行が1件も無い場合に限られる。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — category 列挙未検証で C2 cleanup から黙って外れる。guidelines:151。
#### LSA-15: 配信トリガーの二重発火防止が check-then-Create のみで、DB 側に一意制約が無い — MEDIUM
- 区分: 新規
- 規約: `.claude/refs/backend-application-invariants.md:30` 「foreign key、unique constraint、transaction を application check の代替ではなく、追加の防御として使う。」
- 対象: `backend/internal/lstep/lstep_delivery_trigger_state.go:31`、`backend/internal/lstep/lstep_delivery_trigger_state.go:32`、`backend/internal/lstep/lstep_delivery_trigger_state.go:40`、`backend/internal/lstep/lstep_delivery_trigger_state.go:48`、`backend/internal/lstep/lstep_delivery_trigger_batch.go:61`、`backend/internal/lstep/lstep_delivery_trigger_batch.go:100`、`backend/migrations/001_init.sql:552`、`backend/migrations/001_init.sql:553`
- 内容: `alreadyFiredToday`（COUNT）と `recordTrigger`（Create）の間はトランザクションでも lock でも守られておらず、`lstep_delivery_trigger_log` には (clinic_id, owner_id, trigger_type, 日付) の UNIQUE が無く非一意 index のみ（001_init.sql:552-553）。durable scheduler の再実行や、cron と event 駆動 `TriggerFirstVisitWelcome` / `TriggerCheckupFollowUp` の同時実行で同日二重配信が成立し得る。
- 修正: `(clinic_id, owner_id, trigger_type, date(scheduled_at))` の部分一意 index を incremental migration で追加し、Create の一意制約違反を「既発火」として吸収する形へ変更する。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — 二重発火防止が check-then-create のみ・UNIQUE なし。invariants:30。
#### LSB-06: shared-files アップロードの purpose が列挙値・長さとも未検証で、超長値がDB errorとして500になる — MEDIUM
- 区分: 新規 ／ 横断パターン: X-02
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「- 外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/lstep/shared_file_request.go:15`、`backend/internal/lstep/shared_file_request.go:16`、`backend/internal/lstep/shared_file_request.go:17`、`backend/internal/lstep/shared_file_request.go:43`、`backend/internal/lstep/shared_file_request.go:44`、`backend/internal/lstep/shared_file_request.go:45`、`backend/internal/lstep/shared_file_handler.go:62`、`backend/internal/lstep/shared_file_handler.go:63`、`backend/internal/model/shared_file.go:36`、`backend/internal/model/shared_file.go:37`、`backend/internal/model/shared_file.go:38`、`backend/migrations/001_init.sql:316`
- 内容: uploadSharedFileRequest には binding tag が一切無い。purpose は model 側で inspection_result / vaccine_cert / other の3値だけが定義された列挙相当だが、:44-45 は空文字を other に補うのみで、任意文字列をそのまま通過させる。DDLの purpose 列は varchar(50) のため、50文字超の入力は境界で 400 を返さずDB層まで到達して 500 になる。50文字以内の任意値は無検証で永続化され、集計・画面表示の前提を壊す。owner_id は service 側で clinic 所属検証があるため対象外。
- 修正: `Purpose string \`form:"purpose" binding:"omitempty,oneof=inspection_result vaccine_cert other"\`` を付与する。列挙値は model の3定数から導出し、定数追加時に binding tag が追随することを確認するtestを添える。
- 検証時の補正: evidence の `backend/migrations/001_init.sql:316` は誤り。同行は owners テーブルの `lstep_opt_out boolean NOT NULL DEFAULT false` であり shared_files とは無関係（LS-B-01 の evidence と取り違えている）。正しい行は `backend/migrations/001_init.sql:1260` 逐語「    purpose     varchar(50)  NOT NULL,」（CREATE TABLE shared_files は :1251 から）。varchar(50) NOT NULL・CHECK制約なしという detail の主張自体は :1260 で実測確認済みで成立するため、行番号のみ :316 → :1260 へ訂正して台帳へ記録すること。あわせて cross_domain:true も誤り: 本件は internal/lstep 単一package内のrequest DTO binding tag欠落であり cross-domain ではない（false が正）。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — purpose 未検証。guidelines:151。
#### LSA-16: CSV インポートで json.Marshal のエラーを `_` で明示破棄している3箇所 — LOW
- 区分: 新規 ／ 横断パターン: X-04
- 規約: `.claude/refs/error-handling.md:9` 「error を無視しない。処理できる境界まで返すか、明示的に回復する。」
- 対象: `backend/internal/lstep/lstep_csv_import_service.go:176`、`backend/internal/lstep/lstep_csv_import_service.go:241`、`backend/internal/lstep/lstep_csv_import_service.go:245`
- 内容: `errLog, _ := json.Marshal(result.errors.entries)` と snapshot の tags/scenarios 変換2箇所で error を破棄している。失敗時は空/不正な JSON が `error_log` / `tags` / `scenarios` カラムへ入り、原因が追えない。
- 修正: エラーを受けて処理を中断するか、少なくとも slog + 明示的な回復値（空配列）への分岐を書き、`_` による無言破棄をやめる。
- 検証時の補正: detail の「失敗時は空/不正な JSON が `error_log` / `tags` / `scenarios` カラムへ入り、原因が追えない」という failure scenario は到達不能。json.Marshal がエラーを返すのは unsupported type / 循環参照 / NaN・Inf の場合に限られるところ、:176 の引数は csvImportErrorEntry（int + string フィールドのみ）のスライス、:241/:245 の引数は splitMultiValue が返す []string であり、いずれも構造上 marshal 失敗し得ない。したがって本件は実害のある欠陥ではなく、「破棄理由を注記していない」という package 内慣行との一貫性のみを根拠とする LOW 指摘として扱うべき。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **WITHDRAWN** — json.Marshal 失敗は当該引数型では構造的に到達不能。
#### LSB-05: 疎通確認APIが外部HTTPトランスポートのraw error文字列をそのままresponseへ載せる — LOW
- 区分: 新規 ／ 横断パターン: X-08 ／ 検証で severity 引き下げ
- 規約: `backend/CLAUDE.md:34` 「- error response と log に secret、credential、個人情報、内部詳細を出さない。」
- 対象: `backend/internal/lstep/lstep_settings_connection.go:42`、`backend/internal/lstep/lstep_settings_connection.go:44`、`backend/internal/lstep/lstep_settings_connection.go:53`、`backend/internal/lstep/lstep_settings_connection.go:55`、`backend/internal/lstep/lstep_settings_connection.go:70`、`backend/internal/lstep/lstep_settings_connection.go:72`、`backend/internal/lstep/lstep_settings_connection.go:88`、`backend/internal/lstep/lstep_settings_connection.go:90`、`backend/internal/lstep/lstep_settings_response.go:52`、`backend/internal/lstep/lstep_settings_response.go:54`、`backend/internal/lstep/lstep_settings_response.go:99`、`backend/internal/lstep/lstep_settings_response.go:101`
- 内容: testLstepAPI / testLineAPI は :72 :90 で http.DefaultClient.Do の error を `connection failed: %w` で包んで返す。Do が返すのは *url.Error であり、その Error() は要求URL全体と解決先アドレスを含む（例: `Get "https://internal-host:8443/api/v1/tags": dial tcp 10.0.0.5:8443: connect: connection refused`）。この文字列が :44 :55 で LstepError / LineError に代入され、:52 :54 の json tag 経由でAPI応答へそのまま出力される。内部ホスト名・ポート・解決IPという内部詳細が応答に露出する。
- 修正: 応答へ載せる文字列は「接続失敗」「認証失敗」等の安定した分類コードに正規化し、raw error は slog 側にだけ記録する。:76-78 :94-96 の認証失敗判定は既に分類済みなので、トランスポート層 error も同様に分類する。
- 検証時の補正: detail に以下2点を追記して過大表現を是正すべき: ①当該endpointは routes.go:103,110 で `ResourceHospitalSettings:view` 権限にゲートされた管理者向け診断APIであり、未認証・一般利用者には到達しない。②露出するbaseURLは lstep_settings_service.go:116 の LstepBaseURL として同一actorに既に平文で返却されているため、raw error による新規露出は解決IPと低レベルnetwork error文字列に限られる。「内部ホスト名・ポート・解決IPという内部詳細が応答に露出する」という記述は、管理者が内部URLを設定した場合の仮定であり、既定構成（lstep.DefaultBaseURL = https://api.lstep.jp / line.APIHost）では外部公開ホストの解決IPに留まる。
- round3-review(2026-07-28): **WITHDRAWN** — LSA-08 と同一疎通 raw error 面の重複（X-08 も除外済み）。
### medicalrecord（診療記録）（24件）

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
#### MRA-01: care_plan_items の物理削除に監査もrecovery経路も存在しない（権限・scopeのみ充足） — HIGH
- 区分: 新規 ／ 横断パターン: X-03 ／ 検証で severity 引き下げ
- 規約: `.claude/refs/backend-application-invariants.md:31` 「destructive または irreversible な操作には、権限、対象 scope、監査、recovery 方針を持たせる。」
- 対象: `backend/internal/medicalrecord/care_plan_item_repository.go:95`、`backend/internal/medicalrecord/care_plan_item_repository.go:97`、`backend/internal/medicalrecord/care_plan_item_service.go:252`、`backend/internal/medicalrecord/care_plan_item_service.go:261`、`backend/migrations/001_init.sql:1757`、`backend/internal/model/audit_log.go:102`、`backend/internal/medicalrecord/routes.go:335`、`backend/internal/medicalrecord/hospitalization_service.go:490`、`backend/internal/medicalrecord/checkup_field_result_service.go:199`
- 内容: 規約が要求する4要素のうち権限（routes.go:335 の perm(ResourceHospitalization,"delete")）と対象scope（repository.go:96 の hospitalizations サブクエリによるclinic限定）は充足するが、監査とrecovery方針が両方とも無い。care_plan_items のDDL（001_init.sql:1757-1779）に deleted_at 列が無く、repository.go:97 は .Unscoped().Delete() で行を物理削除する。model.AuditResource の定数群（audit_log.go:102-115）に care_plan_item 相当の値が無く、care_plan_item_*.go のいずれにも audit 呼び出しが無いため、監査経路はコードベースのどこにも存在しない。【失敗シナリオ】入院中の患者に対し DELETE /api/v1/hospitalizations/{id}/care-plan-items/{itemId} を実行すると、投薬・処置指示行（type/name/medicine_id/procedure_id/unit_price を含む）が物理削除され、deleted_at が無いためテーブル上に痕跡が残らず、audit_logs 側にも care_plan_item の resource 種別が無いため記録されない。結果として「誰がいつどの投薬指示を削除したか」が復元も追跡も不可能になる。さらに care_plan_items は退院処理で billing_items へ変換されるため（hospitalization_service.go:490-537、grep実測）、退院前に削除された請求対象の臨床指示が無記録で消える。【位置づけ】同一packageは hard-delete を監査必須と明示的に扱っており（checkup_field_result_service.go:199 逐語「checkup_field_results は hard-delete のため old_value が唯一の耐久記録であり」）、同一sliceのマスタ（cage/consultation/chief_complaint/checkup_type）は persistence.DeleteScopedByID による soft delete で復元可能である。本件はその中の例外であり、house style ではない。【反論の先回り】care_plan_items.hospitalization_id は ON DELETE CASCADE（001_init.sql:1759）のため「従属行だから独立した耐久性は不要」と読む余地があるが、当該DELETEは独立に認可され独立に呼び出せるendpointであり（routes.go:335）、行は請求へ変換される。従属行のCASCADE設計は、請求対象となる臨床指示行の無監査な単独削除を正当化しない。
- 修正: carePlanItemService.Delete（および Update）を Transactor.WithTx の中へ移し、削除前スナップショットと actorID を新設 AuditResourceCarePlanItem として同一tx内に fail-closed で記録する（checkup_field_result_service.go:204-241 の logReplaceDeletionTx と同型）。代替として care_plan_items に deleted_at TIMESTAMPTZ を追加して soft delete 化し、復元経路を用意する。

- 再実測(2026-07-28): **CONFIRMED** — care_plan_item hard Unscoped delete + audit resource 不在。
- round3-review(2026-07-28): **UPHELD** — care_plan_items Unscoped hard-delete・audit/recovery 欠落は invariants:31 適用。権限/scope のみでは不十分。
#### MRA-02: commit済みの care plan item write を後段 re-fetch の error で 5xx へ反転させている — HIGH
- 区分: 新規 ／ 横断パターン: X-01
- 規約: `backend/CODING_RULES.md:78` 「write後の再取得が失敗し得る場合はcommit前の同じtransaction内で行うか、commit済みの成功を後段read errorで失敗へ反転させないcontractにする。」
- 対象: `backend/internal/medicalrecord/care_plan_item_service.go:184`、`backend/internal/medicalrecord/care_plan_item_service.go:194`、`backend/internal/medicalrecord/care_plan_item_service.go:234`、`backend/internal/medicalrecord/care_plan_item_service.go:244`、`backend/internal/medicalrecord/care_plan_item_repository.go:67`、`backend/internal/medicalrecord/care_plan_item_repository.go:54`、`backend/internal/medicalrecord/checkup_service.go:218`、`backend/internal/medicalrecord/checkup_service.go:222`
- 内容: carePlanItemService は Transactor を保持せず、repository の Create（repository.go:67）・Update（:78）・FindByID（:54）はいずれも r.db.WithContext(ctx) を直接使い ambient tx に参加しない（同一fileで persistence.DBOrTx を使うのは FindByHospitalizationID:39 だけ）。したがって write は独立にcommitされ、後段の再取得は別文となる。【失敗シナリオ】POST /api/v1/hospitalizations/{id}/care-plan-items で :184 の repo.Create が成功しcommitされた直後、:194 の repo.FindByID が一時的なDB障害（接続断・statement timeout 等）で失敗すると service は error を返し handler は 5xx を返す。クライアントは作成失敗と判断して再送するため、同じケアプラン項目（投薬指示等）が二重登録される。PATCH 経路（:234 の Update がcommitされた後 :244 の FindByID が失敗）も同型で、更新は適用済みなのに失敗応答になる。【位置づけ】同一packageの checkup_service.go:218-226 は Create → FindByID を同一 transactor.WithTx クロージャ内で実行しており、規約が求める形は既にこのpackage内に存在する。なお clinical_plan_service.go:208-220 も同型の形を持つが、その原因である clinicalPlanService の Transactor 欠如は phase2.html:195 で deferred 済みのため本findingのscopeから外し、care_plan_item のみを対象とする。
- 修正: carePlanItemService に Transactor を注入し、Create/Update の「書込＋再取得」を単一 WithTx クロージャに包む（checkup_service.go:189-227 と同型）。repository 側の Create/Update/FindByID も r.db.WithContext から persistence.DBOrTx(ctx, r.db) へ揃えて ambient tx へ参加させる。

- 再実測(2026-07-28): **CONFIRMED** — Create/Update 後 tx 外 FindByID で 5xx 反転。Transactor 無し。
- round3-review(2026-07-28): **UPHELD** — Create/Update 後 FindByID が tx 外（care_plan_item_service.go:184-248）。CODING_RULES.md:78 / X-01 適用。
#### MRB-02: hospitalizationRepository.LockByIDForUpdate が ambient transaction 不在を fail-closed にせず、ロックが無効なまま成功を返す — HIGH
- 区分: WIP-adjacent（清算後に再検証） ／ 横断パターン: X-05
- 規約: `.claude/refs/backend-application-invariants.md:38` 「`FOR UPDATE`、`FOR SHARE`、transaction-scoped advisory lockに依存するowner operationはambient transaction不在をfail-closedにする。request由来のclinic-scoped FKは同じtransactionで最終検証し、並行するmaster変更がinvariantを壊す場合は参照行をcommitまで固定する。」
- 対象: `backend/internal/medicalrecord/hospitalization_repository.go:98`、`backend/internal/medicalrecord/hospitalization_repository.go:102`、`backend/internal/medicalrecord/hospitalization_repository.go:106`、`backend/internal/medicalrecord/examination_repository.go:101`、`backend/internal/medicalrecord/examination_repository.go:102`、`backend/internal/medicalrecord/hospitalization_service.go:296`、`backend/internal/medicalrecord/hospitalization_service.go:340`、`backend/internal/medicalrecord/hospitalization_service.go:451`、`backend/internal/medicalrecord/hospitalization_service.go:463`
- 内容: LockByIDForUpdate は clause.Locking{Strength: "UPDATE"} を発行するが persistence.TxFromContext(ctx) == nil の検査がなく、:98 の doc comment で「r.db が tx にバインドされていないとロックは SELECT 終了と同時に解放され直列化できない」と危険性を散文で認めるにとどまる一方、同 package の examinationRepository.LockByIDForUpdate は examination_repository.go:102-104 で `if persistence.TxFromContext(ctx) == nil { return nil, apperrors.WrapInternalServerError(...) }` とコードで強制している。現在の全呼び出し元は WithTx 内のため live な悪用経路はないが、tx 外から呼ばれると PostgreSQL は autocommit の文終了時に行ロックを解放するため、呼び出し側は「ロック済み」と信じたまま直列化されていない行スナップショットを受け取りエラーは一切返らない（fail-open）。DischargeWithBilling では locked 行の OwnerID/PetID が :463 の再検証と :496-504 の billing 生成の根拠なので、検証と会計行生成の間に owner/pet が変わると古い owner/pet に紐いた会計行が黙って作られる；増幅要因として :451 だけが Create(:296)/Update(:340) と違い s.transactor の nil ガードを持たない。WIP-adjacent: 本規約の機械gateである backend/internal/lintscan/dbortx_inventory_lint_test.go が現在 dirty（MM）のため清算後に再検証すること。本所見は inventory への登録有無ではなくコード側の fail-closed ガード不在を対象とする。
- 修正: examination_repository.go:102-104 と同じ ambient transaction ガードを hospitalization_repository.go:102 の先頭に追加し、併せて DischargeWithBilling に Create/Update と同じ transactor nil ガードを置く。
- 検証時の補正: (1) known_or_new は WIP-adjacent ではなく「新規」。検証時点の git status --porcelain 実測で backend/internal/lintscan/dbortx_inventory_lint_test.go は clean（afd8404a4 等で清算済み・保護リスト7件もいずれも dirty でない）であり、清算後再検証の deferral は不要で本所見は現時点で actionable。(2) detail の「DischargeWithBilling では検証と会計行生成の間に owner/pet が変わると古い owner/pet に紐いた会計行が黙って作られる」は現行コードでは発生しない（:451 で transactor.WithTx 内、:454 の lock は有効）。潜在影響の例示として留め、実害として記載しない。(3) :451 の s.transactor nil ガード欠落は fail-open ではなく nil interface の panic であり、:296/:340 との一貫性欠如として記述するのが正確。

- 再実測(2026-07-28): **LINE-DRIFT** — LockByIDForUpdate 本体は hospitalization_repository.go:113-122（旧:98/102/106 近傍）。TxFromContext 検査は依然無し。examination 側は fail-closed。
- round3-review(2026-07-28): **DOWNGRADED** → severity **MEDIUM** — LockByIDForUpdate が ambient-tx ガード欠落は実在（examination 対比）だが production 呼び出しは常に WithTx 内。潜在 API 契約穴。
#### MRB-03: hospitalizationRepository.FindByID が clinic 所有の子テーブルを clinic 述語も親相関もなく Preload している — HIGH
- 区分: **既知** → BUG-437
- 規約: `.claude/refs/backend-application-invariants.md:15` 「bulk query、join、preload、count、export、background job にも同じ scope を適用する。」
- 対象: `backend/internal/medicalrecord/hospitalization_repository.go:86`、`backend/internal/medicalrecord/hospitalization_repository.go:87`、`backend/internal/medicalrecord/daily_record_repository.go:42`、`backend/internal/medicalrecord/daily_record_repository.go:45`、`backend/internal/lintscan/preload_clinic_scope_lint_test.go:255`、`backend/internal/lintscan/grandchild_parent_clinic_correlation_lint_test.go:234`
- 内容: `Preload("DailyRecords")`(:86) と `Preload("TreatmentPlans", "deleted_at IS NULL")`(:87) はいずれも clinic_id 列を持つモデル（model.DailyRecord:136 / model.TreatmentPlan:111）を hospitalization_id だけで引くため、生成SQLに clinic 述語も親 hospitalizations との clinic 相関も入らない（同 package の dailyRecordRepository は同じ子テーブルを JOIN ... AND hospitalizations.clinic_id = daily_records.clinic_id(:42) と Preload("VitalRecords", "clinic_id = ?", clinicID)(:45) で明示 scope する準拠形を持つ）。両 gate とも盲点で、preload gate は :255 の `!isMaster` 分岐で master 以外の association を素通りさせ、grandchild 相関 gate は :234 の read terminal 判定（Find/First/Take/Scan/Count）に Preload を含まない。失敗シナリオ: clinic A の入院 42 に対し FK が破損した daily_records 行（hospitalization_id=42 だが clinic_id=B）や treatment_plans 行が存在すると、GET /hospitalizations/42 は clinic A の認証で clinic B の日次記録日付・治療計画内容をレスポンスに載せる。これは BUG-437 が「`Preload("Items")` を機械的に守れない再発防止の穴」として保持する論点の追加evidenceであり（SEC-SWEEP-02 が列挙した残5面に hospitalization_repository.go は含まれていない）、新規起票ではなく BUG-437 への面追加として扱うべき。
- 修正: `Preload("DailyRecords", "clinic_id = ?", clinicID)` / `Preload("TreatmentPlans", "clinic_id = ? AND deleted_at IS NULL", clinicID)` へ変更し、恒久策として BUG-437 の read 側 registry を master 限定から clinic_id 列を持つ全 association へ広げるか、grandchild 相関 gate の検出面に Preload を追加する。
- 検証時の補正: reason に BUG-437（read側 clinic-scope registry が master 限定という再発防止の穴＝gate側の所有者）と SEC-SWEEP-02（残5面リストの所有者・hospitalization_repository.go は未収載）の両IDを併記して面追加として扱う。新規entry起票はしない。

- 再実測(2026-07-28): **LINE-DRIFT** — Preload CarePlanItems/DailyRecords は :97-98（旧:86-87）。grandchild lint read-terminal は :400-410（旧:234 引用）。clinic 述語欠落は継続。
- round3-review(2026-07-28): **REFRAMED** — 未 scope Preload は isolation ギャップだが BUG-437 / SEC-SWEEP-02 既知面。新規 HIGH ではなく既知ポインタ。区分: 既知。
#### MRC-01: 処方更新: commit済みの成功を tx 外の再取得 error で失敗応答へ反転させている — HIGH
- 区分: 新規 ／ 横断パターン: X-01
- 規約: `backend/CODING_RULES.md:78` 「write後の再取得が失敗し得る場合はcommit前の同じtransaction内で行うか、commit済みの成功を後段read errorで失敗へ反転させないcontractにする。」
- 対象: `backend/internal/medicalrecord/prescription_service.go:129`、`backend/internal/medicalrecord/prescription_service.go:146`、`backend/internal/medicalrecord/prescription_service.go:153`、`backend/internal/medicalrecord/prescription_service.go:156`、`backend/internal/medicalrecord/prescription_repository.go:79`
- 内容: prescriptionService.Update は :129 の WithTx 内で lockDraftMedicalRecord + repo.Update を終え :146 で commit するが、:153 で `s.repo.FindByID(ctx, ...)` を tx 外の ctx で実行し、失敗時に :156 で error を返す。prescriptionRepository.FindByID (:79) は `r.db.WithContext(ctx)` なので必ず別接続の新規読み取りになる。UPDATE が commit された直後にコネクションプール枯渇や瞬断で SELECT が失敗すると handler (prescription_handler.go:105) は 5xx を返し、臨床側は「処方期間の変更が失敗した」と認識して再送するが DB 上は既に適用済みで、再送が二重適用または誤った上書きになる。
- 修正: PrescriptionRepository.Update が更新後行を返す形にするか、再取得を WithTx クロージャ内へ移して commit 前に済ませる。tx 外の再取得を残す場合は read 失敗時に更新済みの値を返す contract にする。
- 検証時の補正: 失敗時の帰結の記述を訂正する。buildPrescriptionUpdate は指定 field を map で置換する冪等更新のため、同一 payload の再送は「二重適用または誤った上書き」にはならない。実測される実害は2点: ①commit 済みの成功が 5xx として臨床側へ返る（規約 :78 が明示的に禁じた状態そのもの） ②:156 で early return するため :158 の `s.syncPrescriptionTag(ctx, clinicID, updated.OwnerID)` が実行されず、処方変更に対する LINE タグ同期だけが恒久的に欠落する（DB は更新済みのため再送でも差分が出ず自己修復しない）。

- 再実測(2026-07-28): **CONFIRMED** — prescription Update: WithTx 後 post-commit FindByID。
- round3-review(2026-07-28): **UPHELD** — prescription Update 後 tx 外 FindByID → 5xx 反転。CODING_RULES.md:78。
#### MRC-02: 薬剤削除の連携在庫カスケードが FK ではなく可変の name をキーにし、affected rows も監査も持たない — HIGH
- 区分: 新規
- 規約: `.claude/refs/go-gin-backend-review.md:66` 「update/deleteのaffected rowsを確認し、存在しない対象やscope外対象を成功扱いしていないか。」
- 対象: `backend/internal/medicalrecord/medicine_service.go:312`、`backend/internal/medicalrecord/medicine_service.go:429`、`backend/internal/medicalrecord/medicine_service.go:463`、`backend/internal/medicalrecord/medicine_service.go:500`、`backend/internal/inventory/repository.go:150`、`backend/internal/inventory/repository.go:163`
- 内容: medicineService.Create は :312-320 で medicine.Name を名前に持つ inventory_items 行を自動生成するが、その id を medicines.inventory_id (001_init.sql:747) へ書き戻さない。そのため Delete (:500) と rename 同期 (:429) は id ではなく (clinic_id, name, category=medicine) で対象を選び、inventory 側実装 (inventory/repository.go:150-159, 163-172) は RowsAffected を一切検査せず 0 件でも nil を返す。スタッフが在庫マスタ画面で自動生成された在庫「アモキシシリン」を「アモキシシリン(院内)」へ改名した後にその薬剤を削除すると、DeleteByNameAndMedicineCategory は 0 行削除で成功を返し、薬剤だけが消えて在庫行と在庫数が孤児として残る — error も監査記録も残らず検知経路がない。逆に category=medicine の在庫を手動作成して薬剤名と一致させた場合は、その薬剤の削除で当該在庫行が巻き添えで論理削除される（inventory_items に name の一意制約はない）。加えて薬剤削除経路 (:463-514) には audit_logs 書き込みが一切なく、同ファイルの per_weight 有効化 (:326-330) や medicine_dose_param_service.go:181 の削除が fail-closed で監査される方針と不整合である。
- 修正: Create 時に生成した inventory item の id を medicines.inventory_id へ保存し、削除/rename を id 指定 + RowsAffected==1 assert に変える。0 件時は error を返す。併せて薬剤削除に AuditTxLogger.LogEntryTx を同一 tx で追加する。
- 検証時の補正: 補足2点。①「複数薬剤の同名衝突」経路は成立しない — 001_init.sql:2433 に `CREATE UNIQUE INDEX idx_medicines_clinic_name ON medicines(clinic_id, name) WHERE deleted_at IS NULL` が存在するため、孤児化の再現経路は finding 記載どおり「自動生成在庫を改名してから薬剤を削除する」か「category=medicine の在庫を手動作成して薬剤名と一致させる」の2つに限られる。②audit_logs 欠落の指摘（:463-514 に監査書込なし）は実測どおりだが、根拠は本 finding が引く review.md:66 ではなく backend-application-invariants.md:31（destructive操作に監査を持たせる）である。affected rows の指摘とは根拠行が異なる点を明示すべき。

- 再実測(2026-07-28): **CONFIRMED** — name キー inventory cascade + RowsAffected 未検査 + 監査無し。
- round3-review(2026-07-28): **UPHELD** — name キー inventory カスケード・RowsAffected 未確認・inventory_id 未使用は整合性欠陥。
#### MRC-04: カルテ作成の主訴・治療方針・診断がサイレントに消失し、API は 201 を返す — HIGH
- 区分: 新規 ／ 横断パターン: X-04
- 規約: `.claude/refs/backend-application-invariants.md:35` 「意図的なsaga/best-effort処理は、補償、再試行、監査、部分失敗contractを持たせる。」
- 対象: `backend/internal/medicalrecord/medical_record_subrecords.go:18`、`backend/internal/medicalrecord/medical_record_subrecords.go:38`、`backend/internal/medicalrecord/medical_record_subrecords.go:64`、`backend/internal/medicalrecord/medical_record_subrecords.go:90`、`backend/internal/medicalrecord/medical_record_service.go:69`、`backend/internal/medicalrecord/medical_record_handler.go:117`、`backend/internal/medicalrecord/medical_record_auto_create.go:198`
- 内容: CreateSubRecords は戻り値を持たない (medical_record_service.go:69) ため、inquiry の upsert 失敗 (:38-42)、chief_complaint_type 所有権検証失敗によるスキップ (:18-23)、診断FK検証失敗による clinical_plan 更新スキップ (:64-69)、clinical_plan 更新失敗 (:90-94) がすべて slog.Warn だけで終わる。呼び出し元 medical_record_handler.go:117 は戻り値を受けず直後の :119 で 201 Created を返す。規約が best-effort に要求する補償・再試行・監査のいずれも存在せず、部分失敗 contract もレスポンスに現れない。POST /api/v1/medical-records に chief_complaint と assessment を含めて送り inquiries への INSERT が一時的な DB エラーで失敗すると、201 が返って作成されたカルテの主訴欄は空のままになり、獣医師は主訴が保存されたと認識してカルテを確定させ、以後 addendum でしか訂正できない臨床記録が主訴欠落のまま残る。同一パッケージの auditReservationDraftCleanupFailure (medical_record_auto_create.go:198-235) は同種の best-effort 失敗を専用 audit action で必ず記録しており、本 project の標準はそちらである。
- 修正: 同期 HTTP 経路 (medical_record_handler.go:117) では CreateSubRecords に error を返させ、失敗時は 201 を返さないか部分失敗フィールドを応答に含める。best-effort を維持する AutoCreateFromReservation 経路には auditReservationDraftCleanupFailure と同型の失敗 audit action を追加する。

- 再実測(2026-07-28): **CONFIRMED** — CreateSubRecords 失敗 slog.Warn のみ、handler は 201。
- round3-review(2026-07-28): **UPHELD** — subrecord 失敗を Warn のみで 201。意図的 best-effort でも invariants:35 の補償・再試行・監査・部分失敗 contract 欠落。
#### MRC-05: lab import: exams と exam_results が非原子的に書かれ、補償削除の失敗が握り潰されて再試行不能な孤児 exam を作る — HIGH
- 区分: 新規 ／ 横断パターン: X-06
- 規約: `backend/CLAUDE.md:33` 「1つのbusiness graphを構成する複数rowのwriteは同じtransactionで原子的に扱い、commit済みの成功を後段の再取得errorで失敗応答へ反転させない。」
- 対象: `backend/internal/medicalrecord/lab_import_examination_service.go:219`、`backend/internal/medicalrecord/lab_import_examination_service.go:242`、`backend/internal/medicalrecord/lab_import_examination_service.go:249`、`backend/internal/medicalrecord/lab_import_examination_service.go:254`、`backend/internal/medicalrecord/lab_import_repository.go:202`
- 内容: persistExam は exam 本体を :219 で作成し exam_results を :242 で別 statement として保存する。両者は 1 つの検査結果という business graph を構成するが transaction で括られていない。labImportExaminationService は Transactor を注入されておらず、代わりに :254 の補償削除に依存するが、その delErr は :255-260 でログ出力されるだけで呼び出し元へ返らない。DB 障害中に commit した exam に対し ReplaceItemsByExamID が失敗し、続く補償 Delete も同じ障害で失敗すると、結果なしの exam 行が残る。以後の再取り込みでは LabImportDuplicateCheckerDB.IsDuplicate (lab_import_repository.go:202-217) が deleted_at IS NULL の当該行を検出して true を返すため、行は failed ではなく duplicate として黙ってスキップされ、検査結果は永久に取り込まれずジョブは正常終了に見える。コメント :249-253 自身がこの状態を避けるべきものとして明示している。
- 修正: labImportExaminationService に Transactor を注入し、examRepo.Create と ReplaceItemsByExamID を単一の WithTx に収める。原子化により :254-261 の補償削除経路は不要になり、握り潰される error 自体が消える。
- 検証時の補正: exploitability の限定を追記する（severity 論拠ではなく現況の但し書き）。①補償削除が成功する通常の失敗では primary error が :262 で返り当該行は failed として集計されるため、恒久的な silent skip は「ReplaceItems 失敗」と「補償 Delete 失敗」の複合障害を要する。②commit 経路は lab_result_import_service.go:117-118 で `batch.SourceType != model.LabImportSourceTypeFixture` を拒否しており、現時点では fixture ソース限定（Phase 1 凍結・phase2.html:120 で FE 着手禁止）。ただし POST /api/v1/lab-imports 自体は routes.go:378 で perm(ResourceLabImport,"create") 付きの live route である。

- 再実測(2026-07-28): **CONFIRMED** — exam Create と ReplaceItems 非原子 + 補償 Delete 失敗 swallow。
- round3-review(2026-07-28): **UPHELD** — lab import exams と results が非原子。CLAUDE.md:33。
#### MRD-01: 治療項目の並び順一括更新が affected rows を確認せず、かつ施錠した親カルテに束縛されていない — HIGH
- 区分: 新規
- 規約: `.claude/refs/go-gin-backend-review.md:66` 「- update/deleteのaffected rowsを確認し、存在しない対象やscope外対象を成功扱いしていないか。」
- 対象: `backend/internal/medicalrecord/treatment_repository.go:257`、`backend/internal/medicalrecord/treatment_repository.go:260`、`backend/internal/medicalrecord/treatment_repository.go:261`、`backend/internal/medicalrecord/treatment_service.go:528`、`backend/internal/medicalrecord/treatment_service.go:534`、`backend/internal/medicalrecord/treatment_repository_test.go:429`、`backend/internal/medicalrecord/treatment_repository_test.go:433`
- 内容: 欠陥は2つある。(a) `BulkUpdateSortOrder` は `Update("sort_order", ...)` の `result.RowsAffected` を一切見ず、`result.Error` だけを確認して nil を返す(treatment_repository.go:260-265)。存在しない ID・別 clinic の ID を渡しても 204 No Content が返る。これは同 package 内の `Update`(:236)・`Delete`(:251) が `RowsAffected == 0` を NotFound へ写像しているのと非対称であり、`treatment_repository_test.go:429` の subtest 名「wrong clinic_id silently skips the update without error」と `require.NoError`(:433) がこの silent success を*テストで固定化*している。(b) より重い方: repo の WHERE は `treatments.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ?)` であり、URL で指定され service が施錠・draft 検証した `medicalRecordID` には束縛されていない(treatment_repository.go:261)。service 側 `BulkUpdateSortOrder` は `lockDraftMedicalRecord(txCtx, ..., medicalRecordID, ...)` で URL のカルテだけを draft 判定する(treatment_service.go:528-532)ため、同一 clinic 内の*別*カルテ（確定済みを含む）に属する treatment ID を body に混ぜれば確定済みカルテの並び順が書き換わる。同 service の Update(:363)・Delete(:485) は `existing.MedicalRecordID != medicalRecordID` を明示検証しており、bulk 経路だけがこの検証を欠く。
- 修正: repo の per-item WHERE に `treatments.medical_record_id = ?`（service が施錠した medicalRecordID）を追加して `TreatmentSortUpdate` に MedicalRecordID を持たせ、各 `Update` の `RowsAffected == 0` を `apperrors.WrapNotFound` にする。treatment_repository_test.go:429 の subtest は「別 clinic は NotFound」を期待する形へ反転させる。
- 検証時の補正: (a)(b) は別欠陥。(a)=repository.go:263-265 が result.Error のみ確認し RowsAffected を見ない。(b)=:261 の WHERE が service 施錠済み medicalRecordID に束縛されていない。fix は両方必須で、(a) のみの修正では (b) の同一 clinic 別カルテ越境は残る。

- 再実測(2026-07-28): **CONFIRMED** — BulkUpdateSortOrder が RowsAffected 未確認 + medical_record_id 非束縛。
- round3-review(2026-07-28): **UPHELD** — BulkUpdateSortOrder が RowsAffected 未確認・medical_record_id 未拘束。review.md:66 適用。
#### MRD-02: commit 済みの更新を、後段の応答用再取得エラーで失敗応答へ反転させる — HIGH
- 区分: 新規 ／ 横断パターン: X-01
- 規約: `backend/CODING_RULES.md:78` 「- write後の再取得が失敗し得る場合はcommit前の同じtransaction内で行うか、commit済みの成功を後段read errorで失敗へ反転させないcontractにする。」
- 対象: `backend/internal/medicalrecord/treatment_service.go:462`、`backend/internal/medicalrecord/treatment_service.go:472`、`backend/internal/medicalrecord/treatment_service.go:475`、`backend/internal/medicalrecord/treatment_plan_service.go:141`、`backend/internal/medicalrecord/treatment_plan_service.go:164`
- 内容: `treatmentService.Update` は `WithTx` の閉包が :462 で閉じて commit された*後*に、応答生成用の `FindByID` を :472 で実行し、失敗時に error を返す(:475)。DB の一時障害・接続断でこの read が落ちると、更新は commit 済みなのに client には 5xx が返り「失敗した」と誤認させる（client の再送は同一更新の二重適用になる）。`treatmentPlanService` も Create(:141-145) / Update(:164-168) で同型だが、こちらはそもそも transaction を持たない。同 package の兄弟実装は正しい側を選んでおり — `vitalService.Update` は再取得を tx 内(vital_service.go:294-298)で行い、`vital_repository.go:52-53` に「FindByID participates in an ambient transaction so Update can complete its response re-fetch before commit and roll back when that re-fetch fails」と設計意図が明記されている。`vaccinationService.Create` も同様に tx 内(vaccination_service.go:175-182)。つまり package 内で確立済みのパターンが treatment / treatment-plan の 3 経路だけ踏襲されていない。
- 修正: 再取得を `WithTx` 閉包の内側へ移す（vital_service.go:294-298 と同形）。treatment_plan 側は現在 transaction を持たないため、Transactor を注入して write+再取得を単一 tx に収めるか、再取得を廃して write 時点の in-memory 値から応答を構築する。
- 検証時の補正: 「client の再送は同一更新の二重適用になる」が literal に成立するのは treatment_plan_service.go:135 の Create 経路（重複行生成）。treatment_service.go の PATCH 経路での再送影響は同一 fields の再適用と dose 逸脱監査行の重複であり、在庫の二重減算ではない。

- 再実測(2026-07-28): **CONFIRMED** — treatment/treatment_plan の post-commit re-fetch 反転。
- round3-review(2026-07-28): **UPHELD** — treatment/plan の post-commit re-fetch → 失敗反転。CODING_RULES.md:78。
#### MRD-03: treatment plan の write が親（カルテ/入院）所属を検証せず、resource 単位の権限境界を跨げる — HIGH
- 区分: 新規
- 規約: `.claude/refs/backend-application-invariants.md:22` 「- authentication、role/permission authorization、resource ownership をそれぞれ検証する。」
- 対象: `backend/internal/medicalrecord/treatment_plan_handler.go:181`、`backend/internal/medicalrecord/treatment_plan_handler.go:197`、`backend/internal/medicalrecord/treatment_plan_handler.go:216`、`backend/internal/medicalrecord/treatment_plan_handler.go:223`、`backend/internal/medicalrecord/treatment_plan_service.go:149`、`backend/internal/medicalrecord/treatment_plan_service.go:172`、`backend/internal/medicalrecord/treatment_plan_repository.go:99`、`backend/internal/medicalrecord/treatment_plan_repository.go:113`
- 内容: handler は親 resource の所属を検証する（`verifyMedicalRecordOwnership` at :181 / :216、入院側は `hospitalization.GetByID` at :241 / :277）が、その直後に呼ぶ `service.Update(ctx, clinicID, planID, ...)`(:197) と `service.Delete(ctx, clinicID, planID)`(:223) は planID を clinicID だけで解決する。service(:149, :172) も repo(:99, :113) も `treatment_plans.clinic_id + id` のみで WHERE を組み、`medical_record_id` / `hospitalization_id` が URL の親と一致するかを一度も見ない。結果、親の検証が装飾になっている。カルテ配下ルートと入院配下ルートは別 resource の RBAC（前者 ResourceMedicalRecords、後者 ResourceHospitalization）で守られているため、`medical-records:edit` は持つが `hospitalization:edit` を持たない staff が、自 clinic の任意カルテ ID と入院所属 plan の ID を組み合わせて入院プランを更新・削除できる。cross-tenant ではない（clinic scope 自体は保たれる）が、resource 単位の認可境界は越える。CODING_RULES.md:82「client が送信した clinic/owner/pet/staff ID を認可根拠にせず、関連 resource の ownership を server-side で確認する。」も同時に満たしていない。加えて treatment_plan には audit_logs 経路が存在しない（internal/model/audit_log.go に treatment_plan の action/resource なし、audit middleware も未配線）ため、この越境 write は追跡されない。
- 修正: service の Update/Delete signature に親 ID（medicalRecordID または hospitalizationID）を追加し、`FindByID` で得た plan の `MedicalRecordID` / `HospitalizationID` が渡された親と一致することを write 前に検証して不一致は NotFound を返す。repo 側 WHERE にも同条件を足して RowsAffected で二重に閉じる。

- 再実測(2026-07-28): **CONFIRMED** — Update/Delete が planID+clinic のみで親 MR/入院所属を未検証。
- round3-review(2026-07-28): **UPHELD** — treatment plan write が parent ownership 未検証。invariants:22。
#### MRD-04: treatment plan の金額系入力が binding tag でも service でも検証されず、client 提示の小計をそのまま採用する — HIGH
- 区分: 新規 ／ 横断パターン: X-02
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「- 外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/medicalrecord/treatment_plan_request.go:7`、`backend/internal/medicalrecord/treatment_plan_request.go:9`、`backend/internal/medicalrecord/treatment_plan_request.go:10`、`backend/internal/medicalrecord/treatment_plan_request.go:11`、`backend/internal/medicalrecord/treatment_plan_service.go:112`、`backend/internal/medicalrecord/treatment_plan_service.go:118`、`backend/internal/medicalrecord/treatment_plan_service.go:154`
- 内容: `createTreatmentPlanRequest` は `TreatmentContent` にだけ `binding:"required"` を付け、`UnitPrice`(:7) `Quantity`(:8) `DiscountRate`(:9) `DiscountAmount`(:10) `Subtotal`(:11) には範囲制約が一切ない。service 側も `validateNonNegativePrice` / `validateDiscountRate` / quantity>0 を通さず(treatment_plan_service.go:111-134、update は :154 の buildTreatmentPlanUpdate へ直行)、`subtotal := input.Subtotal` で client 提示の小計を無条件に採用し、0 のときだけ `subtotal = int64(float64(UnitPrice)*qty*(1-DiscountRate/100)) - DiscountAmount` を計算する(:112-119)。`discount_rate=1000` を送れば小計は大きな負値になり、`subtotal=-999999` を直接送ればそのまま永続化される。同 package の treatment 明細は同じ 3 種を必ず検証しており(treatment_service.go:224-232、Update は :393-403)、金額系フィールドで扱いが割れている。sharedkernel の検証関数はすでに存在する(validators_accounting.go:11-17)ため追加コストはない。
- 修正: `treatmentPlanService.Create` / `Update` で `validateNonNegativePrice(&UnitPrice)`・`validateDiscountRate(DiscountRate)`・quantity>0 を通し、`Subtotal` は client 値を捨ててサーバ側で常に再計算する。request struct にも `binding:"omitempty,min=0"` / `binding:"omitempty,min=0,max=100"` を付けて境界でも弾く。
- 検証時の補正: detail の「sharedkernel の検証関数はすでに存在する(validators_accounting.go:11-17)」は所在が不正確。実体は backend/internal/medicalrecord/validators_accounting.go:11-17 で、sharedkernel.ValidateNonNegativePrice / ValidateDiscountRate への同 package delegate である（:3 の import と :5-6 のコメントに明記）。sharedkernel 直下の validators_accounting.go は存在しない。

- 再実測(2026-07-28): **CONFIRMED** — 金額 field に min/max 無く client subtotal を採用し得る。
- round3-review(2026-07-28): **UPHELD** — 金額系 binding 無し・クライアント subtotal 信頼。guidelines:151。
#### MRA-03: consultation master の tax_rate に範囲検証が無い（同packageのpeer masterは min=0,max=1 を課す） — MEDIUM
- 区分: 新規 ／ 横断パターン: X-02 ／ 検証で severity 引き下げ
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/medicalrecord/consultation_request.go:13`、`backend/internal/medicalrecord/consultation_request.go:42`、`backend/internal/medicalrecord/consultation_service.go:52`、`backend/internal/medicalrecord/consultation_service.go:133`、`backend/internal/medicalrecord/consultation_response.go:39`、`backend/internal/medicalrecord/medicine_request.go:16`、`backend/internal/medicalrecord/procedure_request.go:13`、`backend/internal/inventory/merchandise_item_request.go:19`、`backend/migrations/001_init.sql:897`
- 内容: consultation_request.go:13 / :42 の TaxRate *float64 には binding tag が一切無い。同じ struct の TimeCondition / TaxType には oneof が付いているのに、数値の範囲だけが抜けている。値は consultation_service.go:52-54（buildConsultationUpdate）および :133-136（Create）で無検証のまま model へ渡る。【失敗シナリオ（実測確定分のみ）】PATCH /v1/masters/consultations/{id} に {"tax_rate": 99} を送ると 400 にならず永続化され、consultationResponse.tax_rate（consultation_response.go:39）として 99 が echo される。{"tax_rate": -1} も同様に通る。同じ値を medicine / procedure / merchandise_item の同種 endpoint へ送ると binding:"omitempty,min=0,max=1"（medicine_request.go:16、procedure_request.go:13、merchandise_item_request.go:19）により 400 になるため、同一UI上の税率入力がマスタ種別で通ったり弾かれたりする。【防御層】consultations.tax_rate は numeric NOT NULL DEFAULT 0.10 で CHECK 制約が無く（001_init.sql:897）、DB側にも第二の防御線が存在しない。【到達範囲の限界（正直な申告）】consultation.TaxRate の server 側消費先は grep 実測で consultation_response.go:39 のみであり、billing_items は request 由来の自前 tax_rate を持つため、server 側金額計算まで届く経路は本監査では確認できていない。確定しているのは「税率マスタとして無検証で永続化・返却される」と「同種endpoint間で入力契約が割れている」の2点。scope外だが hospitalization_plan_request.go:13,40 も同じ欠落を持つ。
- 修正: createConsultationRequest.TaxRate（consultation_request.go:13）と updateConsultationRequest.TaxRate（:42）に binding:"omitempty,min=0,max=1" を付け、medicine_request.go:16 と同一契約に揃える。Price *int64 についても min=0 の付与を検討する。
- 検証時の補正: 【失敗シナリオを実測で確定した範囲に限定して再記述】PATCH /v1/masters/consultations/{id} または POST に {"tax_rate": 99} / {"tax_rate": -1} を送ると、createConsultationRequest.TaxRate（consultation_request.go:13）・updateConsultationRequest.TaxRate（:42）に binding tag が一切無いため 400 にならず、consultation_service.go:53（fields[colConsultationTaxRate] = *input.TaxRate）および :133-136（Create の taxRate = *input.TaxRate）経由で無検証のまま永続化され、consultation_response.go:39 の tax_rate として echo される。DB 側にも第二防御線は無い（001_init.sql:897 は `tax_rate numeric NOT NULL DEFAULT 0.10` で CHECK 制約なし）。同一 payload を medicine / procedure / merchandise_item の同種 endpoint へ送ると binding:"omitempty,min=0,max=1"（medicine_request.go:16 / procedure_request.go:13 / inventory/merchandise_item_request.go:19）により 400 になるため、同種マスタ間で入力契約が割れている。【元 detail から撤回する主張】「同一UI上の税率入力がマスタ種別で通ったり弾かれたりする」は実測で支持されない。frontend/src/features/master 配下で tax rate 入力を持つのは merchandise / medicine / hospitalization の side panel のみで（MerchandiseSidePanel.tsx:118, MedicineSidePanelSections.tsx:121, HospitalizationSidePanel.tsx:151）、consultation の tax rate 入力 UI は存在しない（features/master 全体を grep して consultation × taxRate のヒット 0 件）。したがって本欠陥は直接 API 呼び出しでのみ到達可能である。【到達範囲】server 側消費先は consultation_service.go:53（write）と consultation_response.go:39（read）のみで、金額計算経路は無い（grep 実測）。frontend 側でも master 由来 tax_rate を会計明細へ引き継ぐのは merchandise のみ（get-merchandise-items.ts:22）で、consultation からの伝播経路は無い。確定しているのは「税率マスタとして無検証で永続化・返却される」と「同種 endpoint 間で入力契約が割れている」の2点のみ。scope 外だが hospitalization_plan_request.go:13,40 も同じ欠落を持つ（実測確認済み）。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — consultation tax_rate 範囲未検証（medicine は min/max あり）。guidelines:151。
#### MRB-05: 入院の削除および PATCH による退院化が監査ログを残さない（同 service は監査依存を保持し退院会計でのみ使用） — MEDIUM
- 区分: 新規 ／ 横断パターン: X-03
- 規約: `.claude/refs/backend-application-invariants.md:31` 「destructive または irreversible な操作には、権限、対象 scope、監査、recovery 方針を持たせる。」
- 対象: `backend/internal/medicalrecord/hospitalization_service.go:385`、`backend/internal/medicalrecord/hospitalization_service.go:419`、`backend/internal/medicalrecord/hospitalization_service.go:169`、`backend/internal/medicalrecord/hospitalization_service.go:545`、`backend/internal/medicalrecord/hospitalization_request.go:143`、`backend/internal/medicalrecord/hospitalization_service.go:107`、`backend/internal/medicalrecord/hospitalization_service.go:371`、`backend/internal/medicalrecord/examination_service.go:577`、`backend/internal/medicalrecord/examination_service.go:131`
- 内容: 4要素のうち権限（perm middleware）・対象 scope（ClinicScope）・recovery（gorm.DeletedAt の soft delete）は揃うが、監査だけが欠落している。hospitalizationService は auditTx フィールド(:169)を持ち DischargeWithBilling(:545)では LogEntryTx を fail-closed に呼ぶのに Delete(:385-429)では一度も呼ばない。さらに PATCH /hospitalizations/:id は status に "discharged" を許容し（hospitalization_request.go:143 の oneof → hospitalization_service.go:107-109 で fields["status"] へ）plain Updates(:371) で退院状態にできるため、UpdateIfNotDischarged の CAS と :545 の監査の両方を迂回する退院経路が存在する。失敗シナリオ: DELETE /hospitalizations/42 または PATCH /hospitalizations/42 {"status":"discharged"} を実行すると、入院が消える／退院済みになるが audit_logs に行が一切残らず、誰がいつ実施したかを事後に特定できない。DischargeWithBilling 経由だけが監査される非対称のため、監査ログを見ても退院実績の全体像が得られない。examinationService.Delete(:577-621) も auditTx(:131) を保持しながら削除監査を書かない同型。
- 修正: MR-B-01 で導入する削除 tx の中で auditTx.LogEntryTx を fail-closed に呼び、old_value に owner/pet/status/期間を残す。PATCH 経由の discharged 遷移も監査対象にするか、UpdateHospitalization では discharged を受け付けず discharge 専用エンドポイントへ寄せる。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — 入院削除/PATCH 退院に監査なし。invariants:31。
#### MRB-07: 検査種別フィールドの transaction 依存欠落を 400 InvalidInput として返している — MEDIUM
- 区分: WIP-adjacent（清算後に再検証）
- 規約: `.claude/rules/go-gin-backend-guidelines.md:167` 「- 既知の application error を安定した HTTP status/code に変換し、未知の error は汎用的な 500 response にする。」

- 対象: `backend/internal/medicalrecord/exam_type_field.go:293`、`backend/internal/medicalrecord/exam_type_field.go:295`、`backend/internal/medicalrecord/exam_type_service.go:90`、`backend/internal/medicalrecord/examination_service.go:182`、`backend/internal/medicalrecord/daily_record_service.go:125`
- 内容: examTypeService.withTx は s.transactor == nil のとき apperrors.WrapInvalidInput("transaction dependency is required") を返すが、これは composition の配線不備というサーバ側の内部状態でありクライアント入力に起因しないため、400 への写像は不適切であり、同 package の同種ガード（examination_service.go:182 / daily_record_service.go:125）がいずれも WrapInternalServerError を使うのと写像が割れている。NewExamTypeService(repo, transactors ...Transactor)（exam_type_service.go:90-96）が可変長で transactor 省略を許すためこの分岐は到達可能なままで、transactor を省いた構築で POST /v1/masters/examination-types/:id/fields を叩くとサーバ配線不備にもかかわらず 400 が返り、監視側も 4xx 扱いのため設定不備が 5xx アラートに乗らない（本番 composition は d.Transactor を渡しているため現時点は client-reachable ではない）。WIP-adjacent: exam_type_field.go は未追跡、exam_type_service.go は dirty（#249 U4 進行中）のため清算後に再検証すること。
- 修正: apperrors.WrapInternalServerError へ変更し、併せて NewExamTypeService の可変長 Transactor を必須引数化して配線漏れをコンパイル時に排除する。
- 検証時の補正: known_or_new は WIP-adjacent ではなく「新規」。検証時点の git status --porcelain 実測で exam_type_field.go / exam_type_service.go はいずれも clean（afd8404a4 feat(#249): add examination field master API で清算済み）であり、両ファイルは並行writer保護リストにも含まれない。清算後再検証の deferral は不要で本所見は現時点で actionable。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — transactor nil を InvalidInput 400。正は 500。規約は guidelines:167。
#### MRB-08: 検査種別の request 由来 parent_id 検証が永続化と同じ transaction の外で行われている — MEDIUM
- 区分: WIP-adjacent（清算後に再検証） ／ 横断パターン: X-05
- 規約: `.claude/refs/backend-application-invariants.md:38` 「`FOR UPDATE`、`FOR SHARE`、transaction-scoped advisory lockに依存するowner operationはambient transaction不在をfail-closedにする。request由来のclinic-scoped FKは同じtransactionで最終検証し、並行するmaster変更がinvariantを壊す場合は参照行をcommitまで固定する。」
- 対象: `backend/internal/medicalrecord/exam_type_service.go:118`、`backend/internal/medicalrecord/exam_type_service.go:131`、`backend/internal/medicalrecord/exam_type_service.go:148`、`backend/internal/medicalrecord/exam_type_service.go:158`、`backend/internal/medicalrecord/exam_type_service.go:87`、`backend/internal/medicalrecord/exam_type_repository.go:57`、`backend/internal/medicalrecord/exam_type_repository.go:60`、`backend/internal/medicalrecord/exam_type_field.go:293`
- 内容: examTypeService.Create は validateParentOwnership(:118) の直後に repo.Create(:131) を、Update は validateParentOwnership(:148) の直後に repo.Update(:158) を、いずれも transaction の外で実行する。同 service は transactor(:87) を保持し field 系のみ withTx（exam_type_field.go:293）を使っており、examTypeRepository.FindByID は ambient tx がある場合にのみ FOR SHARE を掛ける実装(:57-61)を既に備えているため、規約が要求する受け皿が未使用のまま残っている（同 package の hospitalizationService.Create:299-317 と examinationService.Create:200-222 は同じ検証を WithTx 内で行う準拠形）。失敗シナリオ: POST /v1/masters/examination-types {"parent_id":9} の検証が通った直後に別リクエストが検査種別 9 を削除・commit すると、削除済み親を指す exam_type が作られる。WIP-adjacent: exam_type_service.go は dirty（#249 U4 進行中）のため清算後に再検証すること。
- 修正: Create/Update/Delete も s.withTx で包み、validateParentOwnership と repo.Create/Update/Delete を同一 tx 内に置く（これにより repository の FOR SHARE 分岐が実際に発火する）。
- 検証時の補正: known_or_new は WIP-adjacent ではなく「新規」。検証時点の git status --porcelain 実測で exam_type_service.go は clean（afd8404a4 で清算済み）であり、清算後再検証の deferral は不要で本所見は現時点で actionable。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — parent_id 検証が write と同一 tx 外。invariants:38。
#### MRC-03: #201 投与量パラメータ: per_weight 医療安全ガードが write transaction の外で評価され親 medicine 行が固定されない — MEDIUM
- 区分: 新規 ／ 横断パターン: X-05 ／ 検証で severity 引き下げ
- 規約: `backend/CODING_RULES.md:38` 「request由来のclinic-scoped FKは永続化と同じtransactionで再検証し、並行master変更で判定が無効になる場合は対象行をcommitまで共有ロックする。」
- 対象: `backend/internal/medicalrecord/medicine_dose_param_service.go:100`、`backend/internal/medicalrecord/medicine_dose_param_service.go:105`、`backend/internal/medicalrecord/medicine_dose_param_service.go:109`、`backend/internal/medicalrecord/medicine_dose_param_service.go:121`、`backend/internal/medicalrecord/medicine_dose_param_service.go:148`、`backend/internal/medicalrecord/medicine_service.go:394`、`backend/internal/medicalrecord/medicine_service.go:421`
- 内容: medicineDoseParamService.Upsert は :100 で親 medicine を取得し、:105 で calculation_type==per_weight を、:109 で ValidateMedicineDoseConfig により medicine_unit/strength 整合を検証する。いずれも判定根拠が medicines の可変列であるのに、transaction は :121 / :148 の WithTx で初めて開始され、medicine 行に FOR SHARE 等のロックは掛からない。スタッフA が dose-params の PUT で :109 を通過した直後にスタッフB が PUT /masters/medicines/:id で calculation_type を per_weight→none にする（または clear_strength=true で含量を消す）と、A の tx が後から commit され、per_weight でない薬剤に mg/kg の dose param 行が紐づいた状態が確定する — 医療安全ガード①②が防ぐと宣言している状態そのもので、以後の自動計算が不整合な設定を参照する。medicineService.Create/Update の親FK・在庫FK検証 (:256-261 / :394-399) も WithTx (:306 / :421) の外にあるが、そちらは存在・所有のみを見る弱い変種である。
- 修正: Upsert/Delete の WithTx を先に開き、その txCtx で medRepo.FindByID を再実行して calculation_type / medicine_unit / strength を再検証する。medicineRepository に clause.Locking{Strength: "SHARE"} 付きの参照メソッドを追加し、判定に使った medicine 行を commit まで固定する（examination の master FK TOCTOU 封鎖 24929e83d と同型）。
- 検証時の補正: 「以後の自動計算が不整合な設定を参照する」を削除する。実測される残余は「calculation_type=none の薬剤に mg/kg の dose_params 行が残留する」データ整合性の瑕疵のみで、臨床計算経路は dose_calc.go:170 / treatment_dose_save.go:33 で live medicine 行に対し fail-closed であるため誤投与量は生じない。またこの残留状態は競合なしでも（dose param 設定後に PUT /masters/medicines/:id で calculation_type を none へ変更するだけで）到達可能であり、TOCTOU はその狭い部分集合にすぎない。是正の実質は「Upsert/Delete の tx 内再検証＋SHARE ロック」だけでなく「medicineService.Update 側で dose param 存在時の per_weight 解除を扱う契約」も要する。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — per_weight ガードが write tx 外。設定整合性欠陥（投与量誤計算の直接根拠ではない）。
#### MRC-08: lab import の検査項目 DTO が無検証で、異常判定を決める基準値をクライアントが自由に指定できる — MEDIUM
- 区分: 新規 ／ 横断パターン: X-02
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/medicalrecord/lab_import_request.go:38`、`backend/internal/medicalrecord/lab_import_request.go:61`、`backend/internal/medicalrecord/lab_import_request.go:125`、`backend/internal/medicalrecord/lab_import_examination_service.go:74`、`backend/internal/medicalrecord/lab_import_examination_service.go:79`
- 内容: labExamItemReq (:61-69) と labImportResultRowReq (:38-48) は binding tag を1つも持たない。Name / InspectionValue / Unit / ReferenceValue は長さ上限なし、RefMin / RefMax は範囲検証も大小関係検証もない。一方 lab_import_examination_service.go:74 のコメントは「status / is_abnormal はサービス層で計算し、呼び出し元から受け付けない（信頼境界保護）」と信頼境界を宣言しているが、その計算 (:79 computeExamResultStatus) の唯一の判定入力である RefMin/RefMax はこの無検証 DTO から toExamInputs (:125-136) 経由でそのまま渡るため、宣言された信頼境界が実際には成立していない。POST /api/v1/lab-imports で items[].ref_min=1000, ref_max=1001, inspection_value="5.0" を送ると正常値が IsAbnormal=true として保存され、逆に極端に広い範囲を送れば真の異常値が正常として保存される。取り込まれた異常フラグは以後カルテと飼主レポートの判断材料になる。name に数万文字を送ると varchar 制約違反が DB まで到達して 500 になる。
- 修正: labExamItemReq に binding tag を付与する（Name/InspectionValue/Unit/ReferenceValue に max、RefMin/RefMax に omitempty と RefMin<=RefMax の相互検証）。RefMin>RefMax の行は 400 で拒否する。
- 検証時の補正: DB 側の帰結を訂正する。exam_results.name / inspection_value / unit / reference_value は 001_init.sql:1696-1701 でいずれも `text` 型（PostgreSQL の text に長さ上限は無い）であるため「name に数万文字を送ると varchar 制約違反が DB まで到達して 500 になる」は誤り。無制限入力はストレージ肥大の問題に留まる。一方 ref_min / ref_max は :1702-1703 が `decimal(10,4)` であり、桁溢れする数値を送った場合は numeric overflow が DB まで到達して 500 になる。また本経路は routes.go:378 の perm(ResourceLabImport,"create") 付き認証 route であり、外部検査機器の基準値をペイロードで運ぶ import の性質上「基準値を呼び出し元が指定する」こと自体は設計意図に含まれうる。確実に欠けているのは max 長・RefMin<=RefMax の相互検証・数値範囲検証という境界検証そのものである。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **REFRAMED** — DTO 無検証は guidelines:151 で成立。ただし「信頼境界敗北」ではなく境界 validation 不足（lab 参照値は import 設計上 client 由来）。
#### MRC-09: 診療画像の JSON 作成経路が MIME allowlist と形式検証を一切通らない（upload 経路とのガード非対称） — MEDIUM
- 区分: 新規
- 規約: `backend/CODING_RULES.md:51` 「body/query/URI/header を型付き input に bind し、型・形式・範囲・長さを境界で検証する。」
- 対象: `backend/internal/medicalrecord/medical_record_image_request.go:24`、`backend/internal/medicalrecord/medical_record_image_request.go:33`、`backend/internal/medicalrecord/medical_record_image_request.go:37`、`backend/internal/medicalrecord/medical_record_image_request.go:111`、`backend/internal/medicalrecord/medical_record_image_service.go:110`、`backend/internal/medicalrecord/medical_record_image_handler.go:90`
- 内容: 同一リソースに 2 つの作成経路がある。upload 経路 (medical_record_image_handler.go:129-192) は :102 のサイズ上限と :111-120 の allowedMedicalRecordImageMIMETypes 照合を通すが、JSON 作成経路 (同 :71-97 → createMedicalRecordImageRequest) は ImageURL に `binding:"required"` があるだけで形式・長さ検証がなく、MimeType (:37)・FileName (:35)・FileSize (:36) は完全に自由で allowedMedicalRecordImageMIMETypes (:24-29) を参照しない。medicalRecordImageService.Create (:110-123) は validateMedicalImageType で image_type enum を見るのみで、URL と MIME はそのまま永続化され応答にも返る。POST /api/v1/medical-records/:id/images に {"image_url":"https://attacker.example/track.svg","mime_type":"text/html","file_size":-1} を送ると全項目が検証されずに保存され、以後そのカルテを開くすべての職員のブラウザが外部ホストへリクエストを送る。upload 経路で拒否される content type が JSON 経路では通るため、同一リソースに対する保証が経路依存になる。
- 修正: createMedicalRecordImageRequest の MimeType を allowedMedicalRecordImageMIMETypes に対する検証にかける（validate() と同じ判定を共有関数化）。ImageURL に絶対 URL 形式と max 長の binding を、FileSize に medicalRecordImageMaxUploadSize を上限とする範囲検証を追加する。
- 検証時の補正: 攻撃者モデルを明示すべき。routes.go:256 の JSON 作成経路は perm(model.ResourceMedicalRecords, "create") を要求し、handler の :80 verifyOwnership が clinic 所有も検証するため、投入できるのは同一 clinic の認証済みスタッフに限られる（未認証・越境の攻撃面ではない）。したがって帰結は外部ホストへの参照埋め込み（トラッキング／情報持ち出し）とカルテ表示の破損であり、権限昇格や越境ではない。確実な欠陥は、同一リソースの2経路で MIME allowlist・サイズ上限・URL 形式検証の適用が非対称であること、および FileSize に負値を含む任意値が通ることである。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — JSON 画像作成が MIME/size 検証なし（upload 経路はあり）。CODING_RULES.md:51。
#### MRC-12: 問診 upsert が Conflict 応答時に FirstOrCreate で作った空行を残す（既知ガードの残余） — MEDIUM
- 区分: **既知** → phase2.html:195 ／ 横断パターン: X-06
- 規約: `backend/CLAUDE.md:33` 「1つのbusiness graphを構成する複数rowのwriteは同じtransactionで原子的に扱い、commit済みの成功を後段の再取得errorで失敗応答へ反転させない。」
- 対象: `backend/internal/medicalrecord/inquiry_repository.go:42`、`backend/internal/medicalrecord/inquiry_repository.go:50`、`backend/internal/medicalrecord/inquiry_repository.go:64`、`backend/internal/medicalrecord/inquiry_repository.go:86`、`backend/internal/medicalrecord/inquiry_repository.go:94`
- 内容: SaveByMedicalRecordID は事前 status 確認 (:50-58)、FirstOrCreate による行確保 (:62-66)、status='draft' 述語付き Updates (:86-90) の 3 statement を transaction なしで実行する。:42-46 のコメント自身が「inquiryService は Transactor を持たない設計上の制約」を認めており、これは phase2.html:195「clinical_plan/inquiry の finalized ガード正規パターン統一（Transactor＋LockByIDForUpdate）」として起票済みの論点である。追加 evidence として、PATCH /medical-records/:id/inquiries と別セッションのカルテ確定が競合し :50 では draft だったカルテが :86 到達時に finalized になった場合、UPDATE は 0 行で Conflict 409 を返すが :64 の FirstOrCreate が挿入した全項目空の inquiries 行はロールバックされず確定済みカルテに紐づいたまま残り、以後のカルテ表示に空の問診が現れる点を記録する。
- 修正: phase2.html:195 の正規パターン（Transactor + LockByIDForUpdate）へ統一する際、FirstOrCreate と Updates を同一 tx に収めて Conflict 時に挿入行がロールバックされるようにする。cross_tenant_master_fk_write_test.go 側のコンストラクタ制約解消が前提。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **REFRAMED** — 既知 phase2.html:195 / X-06。新規 MEDIUM ではなく既知ポインタ。
#### MRC-14: 診断FK の clinic 所有権検証が ClinicalPlanService から複製されている（自己申告済み） — MEDIUM
- 区分: 新規 ／ 横断パターン: X-09
- 規約: `~/.claude/rules/ecc/common/coding-style.md:26` 「- Avoid copy-paste implementation drift」
- 対象: `backend/internal/medicalrecord/medical_record_subrecords.go:98`、`backend/internal/medicalrecord/medical_record_subrecords.go:101`、`backend/internal/medicalrecord/medical_record_subrecords.go:106`、`backend/internal/medicalrecord/medical_record_subrecords.go:115`
- 内容: validateCreateSubRecordDiagnosisFKs のコメント :98-100 が「clinicalPlanService.validateDiagnosisFKs と同型の clinic-scoped 所有権検証を CreateSubRecords の best-effort パスに複製したもの（最小差分のため ClinicalPlanService は注入しない）」と複製であることを明記している。複製対象が cross-tenant 所有権検証というセキュリティ判定であるため、将来 clinicalPlanService 側に検証が追加されても :101-121 の複製は追随せず、POST /medical-records 経由の診断設定だけが弱い検証のまま通る。複製である事実がコメントにしかないため正本側の変更時に追随漏れが検出されない。根拠区分は project quality policy であり、Go/Gin 公式要件としての指摘ではない（層構造・ファイルサイズを論拠にしていない）。
- 修正: 診断FK 所有権検証を sharedkernel またはパッケージ内の単一関数へ集約し、clinicalPlanService と CreateSubRecords の両方がそれを呼ぶ。MR-C-04 の対応で CreateSubRecords に ClinicalPlanService を注入するなら同時に解消できる。
- 検証時の補正: drift が既に発生している実測を追記できる（推測ではない）。clinicalPlanService は :189 の validateDiagnosisFKs に加えて clinical_plan_service.go:106 の validateDiagnosisTypeNameConsistency（第2診断スロットの type↔name 整合、AUD-007）を実行するが、CreateSubRecords 側の複製はこれを呼ばないため、POST /medical-records 経由の診断設定は既に弱い検証で通っている。一方で所有権判定そのものは両者とも共通ヘルパ validateOwnedMasterFK に集約済みであるため、drift 面は「どの field を・どの repo で・どの周辺検証と併せて検査するか」の配線に限られる。category は naming ではなく duplication/maintainability が適切。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — 診断 FK clinic 所有権検証の copy-paste drift（ClinicalPlan 側のみ name consistency）。coding-style:26。
#### MRA-04: 実在しない package 名を主語にする package comment が担当 7 ファイルに残存 — LOW
- 区分: 新規
- 規約: `.claude/refs/go-language.md:13` 「package comment と exported identifier の comment は GoDoc で意味が通る文にする。」
- 対象: `backend/internal/medicalrecord/cage_handler.go:1`、`backend/internal/medicalrecord/cage_request.go:1`、`backend/internal/medicalrecord/cage_repository.go:1`、`backend/internal/medicalrecord/cage_service.go:1`、`backend/internal/medicalrecord/consultation_handler.go:1`、`backend/internal/medicalrecord/consultation_repository.go:1`、`backend/internal/medicalrecord/consultation_service.go:1`、`backend/internal/medicalrecord/pagination.go:1`
- 内容: go/doc は package 宣言の直前にある comment を package comment として扱う。担当範囲の 7 ファイルはいずれも `package medicalrecord` の直前に、medicalrecord ではない package 名を主語とする文を置いている（cage_handler.go:1 「Package handler provides HTTP handler implementations for Cage entity.」、cage_repository.go:1 「Package cage owns cages master data access.」、cage_service.go:1 「Package service provides business logic implementations for Cage entity.」、cage_request.go:1 と consultation の 3 本も同型）。【失敗シナリオ】`go doc github.com/animal-ekarte/backend/internal/medicalrecord` や pkgsite で package documentation を参照すると、package medicalrecord の説明文として「Package handler provides HTTP handler implementations for Cage entity.」のような、このツリーに存在しない package を名指しした文が表示候補になる。読者は medicalrecord を cage 専用の handler package だと誤解する。pagination.go:1 には正しい「Package medicalrecord owns the medicalrecord domain's HTTP, application logic, and ...」がすでにあり、同一 package に矛盾する package comment が 8 本並立している。【対象の限定】本指摘は comment 文の内容が実在しない package を名指ししている点（ドキュメントの正確性）のみを対象とし、file 配置・ディレクトリ構成・層構造の是非は一切含まない。
- 修正: cage_handler.go:1 / cage_request.go:1 / cage_repository.go:1 / cage_service.go:1 / consultation_handler.go:1 / consultation_repository.go:1 / consultation_service.go:1 の 7 本を削除するか、package 宣言から 1 行空けた通常の file 説明 comment へ格下げし、pagination.go:1 を唯一の package comment として残す。
- 検証時の補正: 並立数の実測値を訂正する。所見は「同一 package に矛盾する package comment が 8 本並立している」と記すが、実測では internal/medicalrecord 配下の非 test file で `// Package ` 始まりの先頭行を持つ file は 19 本あり、うち正しく `// Package medicalrecord` で始まるのは pagination.go の 1 本のみ、残り 18 本が実在しない package 名を主語にしている。担当範囲 [a-c] 外の 11 本は cage/consultation と同型で、daily_record_repository.go（Package dailyrecord）・hospitalization_plan_{handler,repository,service}.go（Package handler / repository / service）・medicine_{repository,handler,service}.go・medicine_dose_param_repository.go・procedure_{handler,service,repository}.go である。所見の 7 file の指摘と proposed_fix は担当範囲として正しく、範囲外 11 本は別 unit で同様に処理されるべき同型残存として記録する（本訂正は所見の severity を変えない）。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **REFRAMED** — package comment の誤 package 名は GoDoc ノイズのみ。runtime 影響ゼロ。
#### MRB-06: 入院の end_date >= start_date は DB CHECK のみで application validation がなく、クライアント入力起因の 500 になる — LOW
- 区分: 新規 ／ 横断パターン: X-02 ／ 検証で severity 引き下げ
- 規約: `backend/CODING_RULES.md:79` 「schema constraint と application validation の両方を使う。」
- 対象: `backend/migrations/001_init.sql:1351`、`backend/internal/medicalrecord/hospitalization_request.go:83`、`backend/internal/medicalrecord/hospitalization_request.go:84`、`backend/internal/medicalrecord/hospitalization_request.go:141`、`backend/internal/medicalrecord/hospitalization_request.go:142`、`backend/internal/medicalrecord/hospitalization_service.go:285`、`backend/internal/medicalrecord/hospitalization_service.go:479`、`backend/internal/apperrors/errors.go:183`、`backend/internal/medicalrecord/dose_validators.go:126`
- 内容: schema 側には chk_hospitalizations_dates CHECK (end_date >= start_date)（001_init.sql:1351）があるが、application 側の検証は request にも service にも存在しない（createHospitalizationRequest は binding:"required" のみ :83-84、updateHospitalizationRequest は制約なし :141-142、Create は :285-286 でそのまま model へ写す。DischargeWithBilling も :479 で end_date を入力 DischargeDate に置き換える際に start_date と比較しない）。apperrors.FromGORM の pgErr switch（errors.go:183-192）は 23503/23505/22003/22P02 のみを扱い check_violation(23514) を含まないため、CHECK 違反は最終行の Wrap(err, "database error") へ落ち 500 になる。失敗シナリオ: POST /hospitalizations に start_date=2026-08-01 / end_date=2026-07-01 を送ると binding も service 検証も通過して INSERT に到達し、23514 が 500 Internal Server Error として返るためクライアントは入力のどこが悪いか判別できない（PATCH の日付単独更新、および start_date より前の discharge_date を渡した discharge-with-billing では退院 tx の途中で同じ 500 になる）。同 package の dose_validators.go:126 は min<=max の順序検証を application 側で二重化しており、基準は package 内に既にある。
- 修正: createHospitalizationRequest / updateHospitalizationRequest の toServiceInput と DischargeWithBilling の適用箇所で end_date >= start_date（更新時は既存値とマージした最終値）を検証し、違反時に apperrors.WrapInvalidInput を返す。
- 検証時の補正: title/failure_scenario から 500 の記述を全削除する。正しい挙動: POST /hospitalizations に start_date > end_date を送ると binding も service 検証も通過して INSERT に到達し 23514 が発生するが、httpapi/response.go:87-89 の isPgError 分岐（response_pg.go:10-13）で 400 Bad Request へ写像され、response_pg.go の classifyPgError の case "23514" により checkConstraintMessages に chk_hospitalizations_dates が未登録のため汎用文言「入力値が制約条件を満たしていません」が返る。残余欠陥は「どのフィールドが不正か特定できない汎用メッセージ」であり、是正はapplication validation追加（dose_validators.go:126 の min<=max 検証が同package内の先例）または checkConstraintMessages への制約名登録のいずれかで足りる。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — end>=start は DB CHECK のみ。23514 は 400（500 ではない）。CODING_RULES.md:79。
#### MRC-07: マスタ削除の使用中ガードが write と同一 transaction・同一ロック下になく TOCTOU で貫通する — LOW
- 区分: 新規 ／ 横断パターン: X-05 ／ 検証で severity 引き下げ
- 規約: `backend/CODING_RULES.md:38` 「request 由来 clinic-scoped FK は永続化と同じ transaction で再検証…」+ 使用中 COUNT と soft-delete の TOCTOU

- 対象: `backend/internal/medicalrecord/procedure_service.go:213`、`backend/internal/medicalrecord/procedure_service.go:224`、`backend/internal/medicalrecord/procedure_service.go:232`、`backend/internal/medicalrecord/medicine_service.go:471`、`backend/internal/medicalrecord/medicine_service.go:483`、`backend/internal/medicalrecord/medicine_service.go:495`、`backend/internal/medicalrecord/procedure_repository.go:46`、`backend/internal/medicalrecord/procedure_repository.go:64`
- 内容: procedureService.Delete は存在確認 (:213)、子処置数 (:216)、使用中数 (:224)、削除 (:232) を 4 つの独立 statement として実行し transaction も行ロックも持たない。medicineService.Delete も同型で CountChildren (:471) / CountUsage (:483) が WithTx (:495) の外にある。削除対象の scope は「参照が 0 件である」という直前の read だけに依存し commit まで固定されていない。スタッフA が DELETE /v1/masters/procedures/:id を実行し :224 が 0 を返した直後にスタッフB が同じ procedure_id で治療記録を作成すると（treatmentService の master FK 検証は論理削除前の行を読むので通過する）、A の :232 の削除が後から成立し、有効な treatments 行が論理削除済み procedure を参照した状態が残って「使用中のため削除できません」ガードの不変条件が壊れる。なお procedureRepository は :39,:46,:50,:57,:64,:71,:78,:95 がすべて `r.db.WithContext(ctx)` を使い persistence.DBOrTx を経由しないため、単に Delete を Transactor.WithTx で包んでも ambient transaction に参加せず修正にならない（medicine_repository.go:100,111 や prescription_repository.go:108 は DBOrTx 済みで非対称）。
- 修正: Delete を Transactor.WithTx で括り、対象マスタ行を FOR UPDATE でロックしてから使用中カウントを再取得する。前提として procedureRepository の全メソッドを persistence.DBOrTx(ctx, r.db) へ揃える。
- 検証時の補正: 引用根拠を invariants:31 から backend/CODING_RULES.md:38 へ差し替えるべき。invariants:31 の4要件（権限・対象scope・監査・recovery）のうち権限（routes.go:288 の perm(ResourceMasterMedical,"delete")）・対象scope（DeleteScopedByID の clinic_id+id）・recovery（soft delete）は実在し、欠けているのは「使用中カウントと削除が同一 transaction・同一ロック下にない」点に限られる。帰結も「有効な treatments 行が論理削除済み procedure を参照する」データ整合性の瑕疵であり、データ喪失や越境は生じない。
- round3-review(2026-07-28): **UPHELD** — 使用中ガードが write と同一 tx 外。TOCTOU 整合性として維持。
### reservation（予約）（7件）

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
#### RSV-02: 管理画面予約作成だけが AcquireBookingLock を取得せず、空き枠へのファントム二重予約とAB-BAデッドロックを許す — HIGH
- 区分: 新規 ／ 横断パターン: X-05
- 規約: `backend/CODING_RULES.md:42` 「同じ evidence を逆向きに変更する競合 workflow は同じ resource-scoped serialization へ参加させ…」+ `reservation_repository.go:164-176` AcquireBookingLock 不変条件

- 対象: `backend/internal/reservation/appointment_admin_service.go:128`、`backend/internal/reservation/appointment_admin_service.go:140`、`backend/internal/reservation/appointment_admin_service.go:143`、`backend/internal/reservation/reservation_repository.go:164`、`backend/internal/reservation/reservation_repository.go:170`、`backend/internal/reservation/reservation_service.go:257`、`backend/internal/reservation/reservation_service.go:408`、`backend/internal/reservation/reservation_validators.go:128`
- 内容: reservation_repository.go:164-176 は「CountConflicts/CountByTypeAndStartTime の SELECT FOR UPDATE は既存行0件の空き枠では何もロックしないためファントムで両方成功しうる」「appointments 行ロック取得前に必ず本メソッドを先頭で呼ぶこと」と不変条件を宣言し、遵守する呼び出し元を3箇所（Create/updateWithConflictCheck/ValidateAndCreate）と列挙している。しかし appointment_admin_service.Create は同じ WithTx 内で CheckSlotConflict/CheckReservationTypeCapacity（FOR UPDATE 依存）を実行しながら AcquireBookingLock を呼ばない4番目の経路である。POST /clinics/:id/reservations の同時2件で空き枠に定員超過の予約が両方成功し、さらに他3経路と行ロック取得順が逆になるためAB-BAデッドロックの成立条件も満たす。
- 修正: appointment_admin_service.Create の WithTx 先頭で `s.resRepo.AcquireBookingLock(ctx, clinicID)` を呼ぶ。併せて reservation_repository.go:174-176 の呼び出し元列挙を4件へ更新し、AST/ランタイムgateで「FOR UPDATE系conflict checkの前に advisory lock」を機械検査する。
- 検証時の補正: 規約引用のみ差し替えを要する。所見が引いた backend/CODING_RULES.md:38 逐語「`FOR UPDATE`、`FOR SHARE`、`pg_advisory_xact_lock`を正しさの根拠にするoperationはambient transaction不在を拒否する。」は本欠陥に適用されない — admin Create は :128 で s.tx.WithTx を開いており ambient transaction は存在するため、この条項の義務は満たしている。欠けているのは tx ではなく advisory lock である。正しい根拠は backend/CODING_RULES.md:42 末尾 逐語「同じevidenceを逆向きに変更する競合workflowがある場合は、両者を同じresource-scoped serialization機構へ参加させ、各writeのcommitまで順序を保持する。」— 同一 clinic の枠占有という同じ evidence を書き換える4つの予約作成/更新 workflow のうち3つ（reservation_service.Create/updateWithConflictCheck、reservation_validators.ValidateAndCreate）は clinic 単位 pg_advisory_xact_lock という resource-scoped serialization 機構へ参加しているのに、appointment_admin_service.Create だけが参加していない。加えて reservation_repository.go:164-176 のin-code不変条件宣言自体が live rule text として直接の根拠になる。

- 再実測(2026-07-28): **CONFIRMED** — appointment_admin_service Create に AcquireBookingLock 無し。
- round3-review(2026-07-28): **UPHELD** — admin Create が AcquireBookingLock 未取得。CODING_RULES.md:38 第一文は適用外だが :42（競合 workflow 直列化）と in-code 不変条件が適用。
#### RSV-03: write成功後の再取得がtransaction外にあり、read失敗がcommit済み成功を500へ反転させる（予約コース作成では重複作成を誘発） — MEDIUM
- 区分: 新規 ／ 横断パターン: X-01 ／ 検証で severity 引き下げ
- 規約: `backend/CODING_RULES.md:78` 「write後の再取得が失敗し得る場合はcommit前の同じtransaction内で行うか、commit済みの成功を後段read errorで失敗へ反転させないcontractにする。」
- 対象: `backend/internal/reservation/line_reservation_setting_service.go:198`、`backend/internal/reservation/line_reservation_setting_service.go:202`、`backend/internal/reservation/reservation_type_liff_service.go:156`、`backend/internal/reservation/reservation_type_liff_service.go:163`、`backend/internal/reservation/reservation_repository.go:346`、`backend/internal/reservation/reservation_repository.go:349`、`backend/internal/reservation/reservation_service.go:534`
- 内容: line_reservation_setting_service.Save は repo.Save（commit済み）の後に repo.FindByID を別文で発行し、その err を 500 へ写像する。reservation_type_liff_service.Create も同形で、Create成功→FindByID失敗時に500が返るためクライアントが再送すると予約コースが重複作成される（一意制約なし）。reservation_repository.update も UpdateScopedByID→FindByID を連ねており、ambient tx が無い経路（reservationService.Update の default 分岐:534、UpdateReservationRoute、tryAttachReservationOwnerPet）では同じ反転が起きる。同package内の reservation_staff_service.Update:194-204 は同じ再取得を WithTx 内で行っており、正しい形は既に確立されている。
- 修正: write と再取得を同一 WithTx に入れる（reservation_staff_service.Update と同形）。あるいは write の戻り値から応答を組み立て、再取得を廃止する。
- 検証時の補正: detail から「クライアントが再送すると予約コースが重複作成される（一意制約なし）」を削除する。reservation_types には 001_init.sql:2437 に `CREATE UNIQUE INDEX idx_reservation_types_clinic_name ON reservation_types(clinic_id, name) WHERE deleted_at IS NULL` があり、同名再送は 409 AlreadyExists となって重複行は作られない。指摘の実害は「Save/Create は commit 済みなのに後段 FindByID の失敗で 500 が返り、クライアントが成功を知れない」という CODING_RULES.md:78 の contract 違反に限定される。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — write 後 re-fetch が tx 外。CODING_RULES.md:78。
#### RSV-04: LINE予約設定の28フィールドrequest DTOにbinding tagが1件も無く、負の booking_window_max_days が make(cap<0) でプロセスをpanicさせる — MEDIUM
- 区分: 新規 ／ 横断パターン: X-02 ／ 検証で severity 引き下げ
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/reservation/line_reservation_setting_request.go:7`、`backend/internal/reservation/line_reservation_setting_request.go:21`、`backend/internal/reservation/line_reservation_setting_request.go:8`、`backend/internal/reservation/line_reservation_setting_request.go:27`、`backend/internal/reservation/line_reservation_setting_request.go:25`、`backend/internal/reservation/line_reservation_setting_service.go:182`、`backend/internal/reservation/liff_service_availability.go:52`、`backend/internal/reservation/available_dates.go:115`
- 内容: `upsertLineReservationSettingRequest` は binding tag が0件（同package内の reservation_request.go:141-148 は oneof を付けている）。`booking_window_max_days` は範囲検証なしで永続化され、liff_service_availability.go:52 経由で available_dates.go:115 の `make([]AvailableDateResult, 0, input.Settings.BookingWindowMaxDays)` に到達する。負値なら Go runtime の makeslice cap out of range で GET /api/liff/:clinicId/available-dates が確実にpanic、巨大値なら即OOM。加えて `status`（reservation_validators.go:109 の稼働ゲート）・`time_slot_mode`・`no_staff_mode` に oneof が無く、typo が無言で全LINE予約を停止させる。`notification_email` も形式検証が無く、送信時に infra/smtp の ValidateEnvelopeAddress で初めて失敗して通知だけが黙って落ちる。
- 修正: `booking_window_min_days`/`max_days`/`calendar_months`/`time_slot_interval_minutes` に `binding:"min=0,max=…"`、`status`/`time_slot_mode`/`no_staff_mode` に `oneof=`、`notification_email` に `omitempty,email`、自由文フィールドに `max=` を付与する。併せて available_dates.go:115 の cap を防御的に clamp する。
- 検証時の補正: title の「プロセスをpanicさせる」を「当該clinicのavailable-dates応答を確実に500へ落とす」に訂正する。backend/cmd/api/main.go:198 で `router.Use(gin.Recovery())` が配線されているため、available_dates.go:115 の makeslice panic は gin の recovery middleware が request 単位で回収し、プロセスは継続する。巨大値による OOM 主張も同様に「1 request 内の大容量確保」であり、プロセス即死は保証されない。指摘の本体である『upsertLineReservationSettingRequest の28フィールドに binding tag が0件で、列挙値・数値範囲・email形式が境界で検証されない』（guidelines:151 違反）はそのまま有効。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — LINE 予約設定 DTO に binding 無し。guidelines:151。
#### RSV-06: 予約キャンセルが「status更新」と「soft delete」の2 writeに分かれ、後段失敗でcancelled状態のまま予約管理に残る — MEDIUM
- 区分: 新規 ／ 横断パターン: X-06 ／ 検証で severity 引き下げ
- 規約: `.claude/refs/backend-application-invariants.md:36` 「appointment、trimming detail、option等で1つのbusiness graphを構成するwriteは同じtransactionで全体を成功またはrollbackさせる。」
- 対象: `backend/internal/reservation/reservation_service.go:499`、`backend/internal/reservation/reservation_service.go:521`、`backend/internal/reservation/reservation_service.go:534`、`backend/internal/reservation/reservation_service.go:545`、`backend/internal/reservation/reservation_service.go:546`
- 内容: reservationService.Update は status=cancelled を含む更新をまず適用（needsConflictCheck / needsLinkValidation 分岐では自前 WithTx が既にcommit済み、default 分岐ではtransactionすら無い）、その後 :545-550 で `s.repo.Delete` を transaction 外の第2 write として実行する。Delete が失敗すると status だけ cancelled で commit された行が deleted_at NULL のまま残り、コメント（:542-544）が意図する「予約管理から消す」business fact が成立しないまま 500 が返る。
- 修正: cancel 経路全体を単一 WithTx に入れ、status 更新と soft delete を同一transactionでcommit/rollbackする。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — cancel が status 更新と soft delete に分裂。invariants:36。
#### RSV-07: 依存チェック→削除がtransaction外で行われ、予約側がFOR SHAREで守っている参照整合をmaster削除側が直列化しない — MEDIUM
- 区分: 新規 ／ 横断パターン: X-05
- 規約: `backend/CODING_RULES.md:42` 「同じevidenceを逆向きに変更する競合workflowがある場合は、両者を同じresource-scoped serialization機構へ参加させ、各writeのcommitまで順序を保持する。」
- 対象: `backend/internal/reservation/reservation_type_liff_service.go:203`、`backend/internal/reservation/reservation_type_liff_service.go:211`、`backend/internal/reservation/reservation_type_liff_service.go:219`、`backend/internal/reservation/reservation_type_repository.go:67`、`backend/internal/reservation/reservation_service.go:588`、`backend/internal/reservation/reservation_service.go:596`、`backend/internal/reservation/appointment_admin_service.go:182`、`backend/internal/reservation/appointment_admin_service.go:185`
- 内容: 予約作成側は reservation_type_repository.FindByID:67-69 が ambient tx 内で FOR SHARE を取り、参照masterをcommitまで固定する。一方 reservationTypeLiffService.Delete は FindByID→CountChildren→ExistsByReservationTypeID→Delete をすべてtransaction外・ロックなしで実行するため、存在チェック通過後にcommitされた新規予約を後追いで参照先ごとsoft deleteできる（soft deleteはFK制約に触れないので:81のフォールバックも効かない）。同型は reservationService.Delete（CountMedicalRecordsByReservationID→Delete がtransaction外）と reservationAdminService.Delete（SoftDelete→DeleteDraftFromReservation が非原子）にもある。
- 修正: 依存チェックと削除を単一 WithTx に入れ、対象master行を FOR UPDATE で固定してから依存countを再評価する（reservation_intent_repository.DeleteForTrimming:591-625 が同packageの正しい先例）。
- 検証時の補正: detail の第3例「reservationAdminService.Delete（SoftDelete→DeleteDraftFromReservation が非原子）」は削除すること。これは欠陥ではなく意図的な設計である：medical_record_auto_create.go:159-161 が逐語で「予約キャンセルは既に確定済みのため、request cancellation や呼出元の ambient tx に cleanup/audit を巻き込まない。一方で同期 best-effort 処理の上限は明示的に制限する。」と宣言し、:161 で `persistence.DetachTx(ctx)` により呼出元 tx から明示的に切り離した上で timeout を掛け、失敗時は :191-195 で構造化log + audit_logs 記録を行っている。同一tx化はこの宣言済み契約の回帰になる。確認対象は reservationTypeLiffService.Delete（:200,:203,:211,:219）と reservationService.Delete（:585,:588,:596）の2件に限定する。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — 依存チェック→削除が tx 外。CODING_RULES.md:42 は拡張気味だが TOCTOU は実在。
#### RSV-09: LIFF予約のスタッフ自動割当で ToDateTime / delegateStaff の error が log すら残さず破棄される — MEDIUM
- 区分: 新規 ／ 横断パターン: X-04
- 規約: `.claude/refs/error-handling.md:9` 「error を無視しない。処理できる境界まで返すか、明示的に回復する。」
- 対象: `backend/internal/reservation/liff_service_reservations.go:26`、`backend/internal/reservation/liff_service_reservations.go:28`、`backend/internal/reservation/liff_service_reservations.go:29`
- 内容: `if err == nil` の入れ子で2つの error が黙って捨てられる。ToDateTime が失敗すれば自動割当自体がスキップされ、delegateStaff が DB error を返しても StaffID=0 のまま予約が確定して IsStaffDelegated=true になる。同ファイル内の他の best-effort 呼び出し（:44 / :62 / :88 / :140）はいずれも WarnContext を出しており、本箇所だけが観測経路を持たない。
- 修正: 両 error を `slog.WarnContext(ctx, "...(best-effort)", "error", err)` で記録する（同ファイルの既存 best-effort と同形）。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — ToDateTime/delegateStaff エラー握り潰し。error-handling.md:9。
#### RSV-08: shift_entry_breaks の孫read が親 shift_entries の clinic 相関を持たず、grandchild lint の登録対象にも入っていない — LOW
- 区分: 新規 ／ 検証で severity 引き下げ
- 規約: `.claude/refs/backend-application-invariants.md:15` 「bulk query、join、preload、count、export、background job にも同じ scope を適用する。」
- 対象: `backend/internal/reservation/reservation_schedule_repository.go:100`、`backend/internal/reservation/reservation_schedule_repository.go:112`、`backend/internal/lintscan/grandchild_parent_clinic_correlation_lint_test.go:32`、`backend/internal/lintscan/grandchild_parent_clinic_correlation_lint_test.go:119`
- 内容: `FindAllBreaksByEntryIDs` / `FindAllBreaksByEntryID` は `Where("shift_entry_id IN ?")` のみで、親 shift_entries.clinic_id との相関 EXISTS を持たない。現在の呼び出し元は clinic-scoped read から entryID を得ているため直接到達は無いが、これは SEC-SWEEP-02 が対象としている汚染FK経由の孫read露出と完全に同型である。grandchild_parent_clinic_correlation_lint_test.go の registry（:32-121）に shift_entry_breaks→shift_entries の target は存在せず、静的gateの死角になっている。
- 修正: 両クエリに `EXISTS (SELECT 1 FROM shift_entries se WHERE se.id = shift_entry_breaks.shift_entry_id AND se.clinic_id = ?)` を追加し、clinicID を引数に取るシグネチャへ変更する。併せて grandchild lint registry に shift_entry_breaks→shift_entries を登録する。
- 検証時の補正: known_or_new を "新規" から SEC-SWEEP-02 への追加evidence（既知entry内の未掃引面）へ変更し、detail に到達性の実測結果を明記する：production 呼出元3箇所（reservation_schedule_service.go:61、liff_service_availability_slots.go:52・:116）の entryID はすべて `Scopes(persistence.ClinicScope(clinicID))` 付きの親read（reservation_schedule_repository.go:63-68 / :83-88）由来であり、外部入力から entryID が直接到達する経路は無い。FindAllBreaksByEntryID（単数形）は production 呼出元0件。よって現時点で悪用可能な欠陥ではなく、SEC-SWEEP-02 の掃引対象面および grandchild lint registry の登録漏れとして扱う。なお backend/internal/lintscan は他セッションが編集中（WIP-adjacent）のため、registry 未登録の判定は清算後に再検証する。
- round3-review(2026-07-28): **REFRAMED** — 既知 SEC-SWEEP-02 孫 correlation 面。独立 NEW ではない。
### auth / staff（認証・スタッフ）（6件）

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
#### AUS-01: スタッフの所属医院全置換（destructive replace）に監査エントリが一切残らない — HIGH
- 区分: 新規 ／ 横断パターン: X-03
- 規約: `.claude/refs/backend-application-invariants.md:31` 「destructive または irreversible な操作には、権限、対象 scope、監査、recovery 方針を持たせる。」
- 対象: `backend/internal/staff/staff_clinic_assignment_service.go:213`、`backend/internal/staff/staff_clinic_assignment_service.go:263`、`backend/internal/staff/staff_clinic_assignment_service.go:273`、`backend/internal/staff/staff_clinic_assignment_service.go:278`、`backend/internal/staff/handler.go:203`、`backend/internal/staff/staff_handler.go:271`、`backend/internal/staff/staff_service_permissions.go:62`
- 内容: SetClinicAssignments は tx 内で assignmentRepo.Delete(ctx, staffID) により既存所属を全削除し、RestoreOrCreate で新集合を再作成、さらに staffs.clinic_id を書き換える（:263-281）。これはスタッフが到達できる clinic 集合＝テナント境界そのものを変更する destructive operation だが、監査書き込みが1件も無い。規約が要求する4要素のうち「権限」（perm(master-staff, edit)）「対象 scope」（authorizeExistingClinicAssignments / AuthorizedClinicIDs）「recovery」（soft delete + RestoreOrCreate）は満たすが「監査」だけが欠落している。同一 package の姉妹経路である SetPermissionGroupIDs は同一 tx 内で permissionAudit.LogEntryTx を fail-closed で書いており（staff_service_permissions.go:62-67・BUG-442 修正 1f9108e6e）、パスワード置換も staff_service_core.go:248 で同様。所属置換だけが取り残されている。
- 修正: SetPermissionGroupIDs と同型に揃える。attachPermissionAssignmentAudit 相当で actor/target/IP/UA を渡し、tx 内で old/new clinic_ids スナップショットを PermissionAssignmentAuditTxLogger（または新設の assignment audit port）へ fail-closed 書き込みする。

- 再実測(2026-07-28): **CONFIRMED** — staff_clinic_assignment_service.go:263-281 の Delete+RestoreOrCreate 連鎖に audit.LogEntry 無し。permission 側 fail-closed 監査との非対称は継続。
- round3-review(2026-07-28): **UPHELD** — staff clinic 全置換に監査ゼロ。invariants:31。permission 置換は fail-closed 監査あり。
#### AUS-03: staff_type が application 層で一切検証されず DB enum だけが防壁になっている — MEDIUM
- 区分: 新規 ／ 横断パターン: X-02
- 規約: `backend/CODING_RULES.md:79` 「schema constraint と application validation の両方を使う。」
- 対象: `backend/internal/staff/staff_request.go:17`、`backend/internal/staff/staff_request.go:70`、`backend/internal/staff/staff_service_builders.go:33`、`backend/internal/staff/staff_service_core.go:87`、`backend/internal/staff/staff_service_account.go:58`、`backend/migrations/001_init.sql:277`、`backend/internal/staff/shift_request.go:77`
- 内容: createStaffRequest.StaffType は binding:"max=16"、updateStaffRequest.StaffType は binding:"omitempty,max=16" のみで oneof 指定が無く、service 側（staff_service_core.go:87-90 / staff_service_account.go:58-61 / staff_service_builders.go:33-35）にも列挙値検証が無い。model.StaffType の有効値は doctor/nurse/trimmer/resource の4つで、staffs.staff_type は PostgreSQL enum 型（001_init.sql:277）。同一 package の shift_type は binding の oneof（shift_request.go:77）と service の validateShiftType の二重検証を持っており、staff_type だけが欠けている。**追加検証済み**: 不正値は PG 22P02 → apperrors.FromGORM（apperrors/errors.go:190）→ WrapInvalidInput → 400 に落ちるため、500 にはならずデータ整合も破れない。違反しているのは「schema constraint と application validation の両方」という要件そのもの（application validation が不在）で、実害は DB 往復後に返る非フィールド特定の汎用 400 メッセージ「入力値の形式が正しくありません」に留まる。列型が将来 text に緩和された場合は無検証の任意文字列が永続化する。
- 修正: createStaffRequest / updateStaffRequest の staff_type に binding:"omitempty,oneof=doctor nurse trimmer resource" を付け、shift_type と同様に service 側にも validateStaffType を置いて Create/CreateWithAccount/Update の3経路から呼ぶ。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — staff_type が app 層未検証。CODING_RULES.md:79。
#### AUS-04: ShiftTemplateRepository.CountUsageByShiftTemplateID は consumer ゼロの死んだ method で、testも恒真 — MEDIUM
- 区分: 新規
- 規約: `.claude/rules/go-gin-backend-guidelines.md:89` 「interface は一般に利用側で定義し、利用側が本当に必要とする最小メソッド集合にする。」
- 対象: `backend/internal/staff/shift_template_repository.go:28`、`backend/internal/staff/shift_template_repository.go:150`、`backend/internal/staff/shift_template_repository.go:154`、`backend/internal/staff/shift_template_service.go:257`、`backend/internal/staff/shift_template_repository_integration_test.go:264`
- 内容: CountUsageByShiftTemplateID は interface に露出しているが production consumer が0件（grep 全数: 宣言・実装・test のみ。shiftTemplateService.Delete は依存チェックせず lock→Delete のみ）。実装は `_ = ctx; _ = clinicID; _ = id; return 0, nil` で全引数を捨てて定数を返す。さらに integration test（:264-296）は3ケースとも `assert.Equal(int64(0), count)` のみで、定数実装に対して恒真であり回帰検出能力がゼロ。同 test file の header コメント（:6「FindAll / FindByID / CountUsageByShiftTemplateID は clinic_id でテナント隔離される」/ :9「deleted_at IS NULL のテンプレートに紐づく breaks のみ数える」）は実装に存在しない挙動を記述しており、doc/test と実装が乖離している。参考: backend/CODING_RULES.md:40「compatibility facadeは薄いdelegate/type aliasだけを許可し、business ruleやpersistence実装を複製しない。consumer移行後の削除条件を持たせる。」も削除条件の欠如として該当する。
- 修正: interface と実装から CountUsageByShiftTemplateID を削除し、対応する mock（shift_template_service_test.go:97）と恒真 test（:264-296）および誤った header コメント（:6, :9）を同時に除去する。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — CountUsage スタブ + 恒真テストは死んだ API。YAGNI/dead-code として維持。
#### AUS-05: toShiftResponse が *model.Staff を nil ガードなしで参照する（repository は意図的に nil にする） — MEDIUM
- 区分: 新規
- 規約: `~/.claude/rules/ecc/common/code-review.md:98` 「- Missing error handling - handle explicitly」
- 対象: `backend/internal/staff/shift_response.go:63`、`backend/internal/staff/shift_entry_repository.go:20`、`backend/internal/staff/shift_entry_repository.go:22`、`backend/internal/staff/shift_entry_repository.go:72`、`backend/internal/staff/shift_entry_repository.go:101`、`backend/internal/model/staff.go:71`、`backend/internal/staff/staff_response.go:86`
- 内容: model.ShiftEntry.Staff は *Staff（model/staff.go:71）。shift_response.go:63 は `if s.Staff.ID != 0` と無条件に参照するが、shift_entry_repository.go:72/:101 の Preload は staffAssignedToClinicCond（:22 = `deleted_at IS NULL AND EXISTS(該当clinicのactive assignment)`）で絞っており、:20-21 のコメント自身が「keeps the shift entry visible while hiding a Staff association that does not have an active assignment to the requested clinic」と、association を意図的に埋めない設計であることを明記している。条件不一致時 GORM は pointer を nil のまま残すため nil pointer dereference になる。同 package の toStaffSummary（staff_response.go:85-88）は nil ガードを持っており、対称性が崩れている。根拠区分は project quality policy であり Go/Gin 公式要件としては指摘しない（.claude/refs/go-gin-backend-review.md:91 準拠）。
- 修正: shift_response.go:63 を `if s.Staff != nil && s.Staff.ID != 0` に変更する（toStaffSummary と同じガード形）。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — toShiftResponse が *Staff nil ガードなし。Preload で nil になり得る。
#### AUS-06: auth/http_session.go が 820 行で上限 800 行を超過 — MEDIUM
- 区分: 新規
- 規約: `~/.claude/rules/ecc/common/coding-style.md:39` 「- 200-400 lines typical, 800 max」
- 対象: `backend/internal/auth/http_session.go:1`、`backend/internal/auth/http_session.go:820`
- 内容: wc -l 実測 820 行。担当2 package 58 file 中、上限超過はこの1本のみ（次点 permission_group_service.go 771 行は範囲内）。内容も Login / Logout / RefreshToken / GetMe の HTTP handler、cookie 発行・消去、clinic 解決、logout 監査 identity 解決、login 失敗監査が同居しており凝集度でも分割余地がある。**根拠区分 = project quality policy。Go/Gin 公式要件としては指摘しない**（.claude/rules/go-gin-backend-guidelines.md:227「- package/file/directory の固定サイズ」を公式要件から除外、:231 も同様）。severity は ~/.claude/rules/ecc/common/code-review.md:57 の Maintainability concern に接地させたが、この閾値自体の適用可否には規約側の未解決の矛盾がある（下記 rule_defects の1件目を参照）。
- 修正: cookie 発行・消去（IssueAuthCookies / issueAuthCookies / ClearCookie）を http_cookies.go へ、logout 監査 identity 解決（logoutAuditIdentity / parseRefreshTokenRevocations / auditLogoutBestEffort / logLogoutAudit）を http_logout_audit.go へ縦切りで分離する。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **WITHDRAWN** — 800 行上限は ECC soft style。Go/Gin guidelines はファイルサイズを公式要件としない（:227）。強制 ADR なし。
#### AUS-09: HasPermission / CalculateEffectivePermissions が既にレスポンスを書いた Extract* の後にもう一度書く経路がある — LOW
- 区分: 新規 ／ 横断パターン: X-10 ／ 検証で severity 引き下げ
- 規約: `.claude/refs/error-handling.md:29` 「response を書いた後に別の error response を重ねない。」
- 対象: `backend/internal/auth/http_permission.go:20`、`backend/internal/auth/http_permission.go:28`、`backend/internal/auth/http_permission.go:32`、`backend/internal/auth/http_permission.go:66`、`backend/internal/auth/http_response.go:210`、`backend/internal/httpapi/context.go:37`、`backend/internal/httpapi/response.go:15`
- 内容: httpapi.ExtractIsSystemAdmin / ExtractStaffID / ExtractClinicID は失敗時に RespondError で即レスポンスを書いて false を返す（httpapi/context.go:37,42,47 と :198,203）。RespondError は c.JSON のみで Abort も行わない（httpapi/response.go:15-18）。HasPermission はこれら3つを呼び（http_permission.go:20,28,32）、false 時に単に return false するため、呼び出し元 RequirePermission が更に RespondError(Forbidden)+Abort を実行する（:66-69）= 同一実行パスで2回書き込む。RequirePermissionAny（:79-88）は要件数だけ HasPermission を呼ぶため最大 N+1 回。CalculateEffectivePermissions（http_response.go:210-218）も ExtractClinicID 失敗時に空 permission を返すだけで、呼び出し元 GetMe が 200 JSON を重ねる。**到達性**: 前提状態は「認証済み扱いなのに gin context の user_id/clinic_id/is_system_admin が未設定または型不一致」。全 caller（composition_runtime.go:499-574 経由で medicalrecord 等へ method value として注入されるものを含む）は Authenticate middleware 配下の protected group に登録されており、HEAD では API のみで到達する経路を構成できなかった。防御的分岐の欠陥として報告する。
- 修正: HasPermission 内では RespondError を書かない Extract 系（OptionalStaffID 相当の非書き込み版）を使い、レスポンス書き込みは RequirePermission / RequirePermissionAny の1箇所に集約する。
- round3-review(2026-07-28): **REFRAMED** — 二重 response パターンは防御的だが Authenticate 後は実質到達不能。
### pet / owner / clinic（ペット・飼主・医院）（16件）

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
#### POC-02: write→reload が transaction 外で行われ、commit 済み成功が後段 read error で 5xx に反転する（4箇所） — HIGH
- 区分: 新規 ／ 横断パターン: X-01
- 規約: `backend/CODING_RULES.md:78` 「write後の再取得が失敗し得る場合はcommit前の同じtransaction内で行うか、commit済みの成功を後段read errorで失敗へ反転させないcontractにする。」
- 対象: `backend/internal/clinic/clinic_service.go:403`、`backend/internal/clinic/clinic_service.go:408`、`backend/internal/clinic/company_service.go:112`、`backend/internal/clinic/company_service.go:116`、`backend/internal/clinic/closing_special_period_repository.go:66`、`backend/internal/clinic/closing_special_period_repository.go:70`、`backend/internal/pet/chronic_condition_service.go:117`、`backend/internal/pet/chronic_condition_service.go:122`、`backend/internal/owner/repository.go:288`、`backend/internal/pet/repository.go:417`
- 内容: clinic 更新・company 更新・特別期間更新・慢性疾患更新のいずれも、UPDATE を先に commit してから別クエリで再取得しており、transactor は使っていない。再取得側が失敗すると（接続断・statement timeout 等）handler は 500 を返すが DB 上の更新は確定済みで、client は「失敗した」と誤認して再送する。同一 repo 内の owner.UpdateAndFind（owner/repository.go:288 の Transaction 内で reload）と pet.UpdateAndFind（pet/repository.go:417 の withPetUpdateTransaction 内で reload）は正しい形を既に実装しており、上記4箇所だけが逸脱している。
- 修正: owner/pet と同じく update+reload を1つの Transaction / Transactor.WithTx へ収める。clinic 側は ports.go:11 の Transactor が既にあるので clinicService.UpdateClinic・companyService.Update から利用し、closing_special_period_repository は persistence.DBOrTx へ寄せる。

- 再実測(2026-07-28): **CONFIRMED** — clinic/company/special-period/chronic の write→reload が tx 外。
- round3-review(2026-07-28): **UPHELD** — clinic/company/special-period/chronic の write→reload が tx 外。CODING_RULES.md:78。
#### POC-03: pet 更新の owner_id / insurance_id 再検証が write transaction の外にあり、write 時に再検証されない — HIGH
- 区分: WIP-adjacent（清算後に再検証） ／ 横断パターン: X-05
- 規約: `backend/CODING_RULES.md:38` 「`FOR UPDATE`、`FOR SHARE`、`pg_advisory_xact_lock`を正しさの根拠にするoperationはambient transaction不在を拒否する。request由来のclinic-scoped FKは永続化と同じtransactionで再検証し、並行master変更で判定が無効になる場合は対象行をcommitまで共有ロックする。」
- 対象: `backend/internal/pet/service.go:353`、`backend/internal/pet/service.go:374`、`backend/internal/pet/service.go:397`、`backend/internal/pet/repository.go:417`、`backend/internal/pet/repository.go:419`、`backend/internal/pet/owner_registration.go:180`、`backend/internal/pet/owner_registration.go:183`
- 内容: petService.Update は request 由来の owner_id（service.go:353）と insurance_id（:374）を transaction 外の FindByID で検証してから UpdateAndFind を呼ぶ。UpdateAndFind（repository.go:417-424）は pets 行だけを FOR UPDATE し、owners / insurances は tx 内で再検証も共有ロックもしない。検証と write の間に対象 owner または insurance が soft delete されると、pets.owner_id / insurance_id が削除済み行を指したまま commit される。同 package の create 経路（owner_registration.go:180 lockOwnerForRegistration / :183 lockOwnerRegistrationMasters）は同一 tx 内で owner を FOR UPDATE・master を FOR SHARE しており、update 経路だけが非対称に欠けている。
- 修正: UpdateAndFind の transaction 内で、fields に owner_id / insurance_id が含まれる場合に lockOwnerForRegistration / lockOwnerRegistrationMasters 相当の clinic 相関付き SHARE ロック検証を実行する。

- 再実測(2026-07-28): **LINE-DRIFT** — owner/insurance 事前検証 :352-376、UpdateAndFind :416-428（旧:417）。write tx 内再検証無しは継続。
- round3-review(2026-07-28): **UPHELD** — owner_id/insurance_id 再検証が write tx 外。CODING_RULES.md:38 第二文適用。BUG-415 Status 省略とは無関係。
#### POC-01: 休診日ミューテーションが2つのroute群で二重登録され、必要権限が分岐している — MEDIUM
- 区分: 新規 ／ 検証で severity 引き下げ
- 規約: `.claude/refs/backend-application-invariants.md:22` 「authentication、role/permission authorization、resource ownership をそれぞれ検証する。」
- 対象: `backend/internal/clinic/clinic_holiday_handler.go:90`、`backend/internal/clinic/clinic_holiday_handler.go:93`、`backend/internal/clinic/clinic_holiday_handler.go:94`、`backend/internal/clinic/closing_settings_handler.go:141`、`backend/internal/clinic/closing_settings_handler.go:143`、`backend/internal/clinic/closing_settings_handler.go:144`、`backend/cmd/api/composition_runtime.go:535`、`backend/cmd/api/composition_runtime.go:537`、`backend/internal/clinic/clinic_service.go:217`、`backend/internal/clinic/clinic_service.go:234`
- 内容: 同一の h.SetClinicHoliday / h.DeleteClinicHoliday が POST /clinic-holidays（ResourceShifts）と POST /closing-settings/holidays（ResourceClosingSettings）の2経路に登録され、composition_runtime.go:535 と :537 で両方 wire されている（GET は双方 view 要求のため対象外。分岐するのは POST/DELETE）。既定権限表では Shifts は一般グループに create=true（clinic_service.go:217）、ClosingSettings は執行・一般とも create=false / delete=false（:234）。結果、休診日設定の実効権限は `shifts:create OR closing-settings:create` となり、権限グループ設定でも監査でも表現できない。同時に /closing-settings/holidays の POST/DELETE は新規クリニックの全非 system_admin スタッフに 403（auth/http_permission.go:60 の deny-by-default）となり、経路として成立していない。
- 修正: 休診日の write は1経路に統一し、closing_settings_handler.go:143-144 の重複登録を撤去する（または逆に clinic_holiday_handler.go:93-94 を撤去）。残す側の resource/action を defaultPermissionRuleTable と突合し、routes_snapshot_test に (path, resource, action) の一意性 assertion を追加する。
- 検証時の補正: 二重登録は未検出のdriftではない。backend/docs/api.yaml:12119-12124 が /closing-settings/holidays を宣言し逐語コメント「既存 /clinic-holidays と同一ハンドラに委譲（clinic_holiday_handler.go）」で意図を明示、backend/internal/clinic/routes_test.go:40-42,51-53 が両系統6本を route snapshot として固定済み。権限の union も広がらない（closing-settings:create が既定 false のため実効は shifts:create 単独と等価）。実在する欠陥は security ではなく機能不整合: defaultPermissionRuleTable（clinic_service.go:234）が ClosingSettings を exec/gen とも create=false/delete=false と定め、backend/migrations/seeds/003_demo/permission_group_rules.csv に closing-settings 行が0件（grep -c 実測=0）。一方 frontend/src/features/closing-settings/api/holidays.ts:26,31 は POST/DELETE /v1/closing-settings/holidays を実呼び出しするため、既存・新規いずれのクリニックでも非 system_admin スタッフは締め設定画面の休診日登録/削除で 403 になる。category は security ではなく api-contract（既定権限表と route 要求権限の不一致）。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **REFRAMED** — 二重 route は意図的。問題は default 権限と FE が閉じる path の契約不一致であり、権限 OR 昇格ではない。
#### POC-05: 特別期間の重複禁止が application 検証のみで、DB 制約も transaction も無い — MEDIUM
- 区分: 新規 ／ 横断パターン: X-05
- 規約: `backend/CODING_RULES.md:79` 「schema constraint と application validation の両方を使う。」
- 対象: `backend/internal/clinic/closing_settings_service.go:183`、`backend/internal/clinic/closing_settings_service.go:203`、`backend/internal/clinic/closing_settings_service.go:256`、`backend/internal/clinic/closing_special_period_repository.go:59`、`backend/internal/clinic/closing_special_period_repository.go:94`、`backend/migrations/001_init.sql:2115`、`backend/migrations/001_init.sql:2116`
- 内容: CreateSpecialPeriod は CheckOverlap（service:183）の後に Create（:203）を呼ぶが両者は同一 transaction ではない。closingSpecialPeriodRepository は全メソッドが r.db.WithContext(ctx) を直接使い（:20/:32/:44/:59/:80/:94）、Update も r.db を渡す（:66）ため persistence.DBOrTx を一切経由せず、呼び出し側が transaction で包んでも参加できない。DDL 側は chk_closing_special_periods_date_range と chk_closing_special_periods_time_order（001_init.sql:2115-2116）のみで、(clinic_id, 期間) の排他制約は無い。同時 POST 2件が双方 CheckOverlap を通過して重複期間が確定し、ResolveSchedule（service:300 FindByDate の First）が非決定的に一方を返して締め時刻が不定になる。
- 修正: closingSpecialPeriodRepository を persistence.DBOrTx ベースへ揃えたうえで CheckOverlap→Create/Update を1 transaction に収め、併せて btree_gist の EXCLUDE 制約（clinic_id WITH =, daterange(start_date, end_date, '[]') WITH &&）を新規 incremental migration で追加する。
- 検証時の補正: 「ResolveSchedule（service:300 FindByDate の First）が非決定的に一方を返して締め時刻が不定になる」は不正確。closing_special_period_repository.go:42-56 の FindByDate は GORM の First を使っており、GORM First は主キー昇順の ORDER BY を自動付与するため、重複時は id 最小行が決定的に返る。正しい failure は『重複期間が確定した後、id の大きい側の設定が恒久的に無視され、UI 上は2件見えるのに適用されるのは常に1件だけになる』である。TOCTOU による重複作成そのものは指摘どおり成立する。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — 特別期間重複が app のみ・DB/tx なし。CODING_RULES.md:79。
#### POC-06: 飼主 phone の一意性が application 検証のみで、email と異なり DB 制約が無い — MEDIUM
- 区分: WIP-adjacent（清算後に再検証）
- 規約: `backend/CODING_RULES.md:79` 「schema constraint と application validation の両方を使う。」
- 対象: `backend/internal/owner/service_core.go:163`、`backend/internal/owner/service_core.go:167`、`backend/internal/owner/service_core.go:172`、`backend/internal/owner/repository.go:178`、`backend/internal/owner/repository.go:207`、`backend/migrations/001_init.sql:2175`
- 内容: ensureOwnerPhoneUnique は FindByPhone の結果で重複を弾くだけで、CreateWithPets / Update の transaction 外にある。owners の一意インデックスは uk_owners_clinic_email（001_init.sql:2175）のみで phone には無いため、同時 POST で同一 phone の飼主が2件確定しうる。その状態になると FindByPhone（repository.go:178 の First）は任意の1件を返し、FindByNameAndPhone（:207 の `len(owners) != 1` 判定）は nil を返して LINE 自動紐付けが黙って不成立になる。git show HEAD:backend/internal/owner/service_core.go で確認したところ本ロジックは in-flight 差分の前後で同一であり、WIP-adjacent タグは情報提供目的である。
- 修正: email と同形の部分一意インデックス（clinic_id, phone）WHERE deleted_at IS NULL AND phone <> '' を追加し、repository 側で persistence.IsUniqueConstraintErr を phone にも写像する。
- 検証時の補正: 「FindByPhone（repository.go:178 の First）は任意の1件を返し」は不正確。GORM の First は主キー昇順を自動付与するため id 最小行が決定的に返る。実害の中核は指摘どおり FindByNameAndPhone（repository.go:207 の len(owners) != 1 判定）が nil を返して LINE 自動紐付けが黙って不成立になる点。なお First の行は :178 ではなく :180（:178 は関数宣言行）。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — phone 一意が app のみ。CODING_RULES.md:79。
#### POC-07: 全クリニック共有マスタ animal_species の更新・削除に監査記録が一切残らない — MEDIUM
- 区分: 新規 ／ 横断パターン: X-03
- 規約: `.claude/refs/backend-application-invariants.md:31` 「destructive または irreversible な操作には、権限、対象 scope、監査、recovery 方針を持たせる。」
- 対象: `backend/internal/pet/animal_species_handler.go:56`、`backend/internal/pet/animal_species_handler.go:77`、`backend/internal/pet/animal_species_service.go:107`、`backend/internal/pet/animal_species_service.go:132`、`backend/internal/pet/animal_species_repository.go:58`、`backend/internal/pet/animal_species_repository.go:72`、`backend/internal/pet/routes.go:43`、`backend/internal/pet/routes.go:44`、`backend/internal/pet/ports.go:57`
- 内容: animal_species は clinic_id を持たないシステム共通マスタ（handler.go:90 のコメントで明示、model/animal_species.go に ClinicID 無し）だが、更新・削除は各クリニックの master-animal-species 権限だけで実行でき（routes.go:43-44）、実行者・変更前後値を残す監査書き込みが無い。pet/owner/clinic の3パッケージで audit へ書き込むのは pet_owner_service.go:148 のみである（ports.go:57 の PetOwnerAuditLogger が唯一の監査 port）。Delete は pets.animal_species_id の ON DELETE RESTRICT（001_init.sql:1102）と CountUsageByAnimalSpeciesID で守られるが、Update には使用中ガードが無く、あるクリニックが is_active=false やリネームを行うと全クリニックの登録フォームから種別が消え、誰がいつ行ったかを事後に特定できない。
- 修正: animal_species の Create/Update/Delete/Reorder に audit.Entry（ClinicID=実行元、OldValue/NewValue）の書き込みを business write と同じ経路で追加する。
- 検証時の補正: 補強事実として、model.AnimalSpecies（backend/internal/model/animal_species.go 全17行）には DeletedAt フィールドが無いため repository の Delete は soft delete ではなく物理削除である。したがって invariants:31 が求める4要素のうち『監査』だけでなく『recovery 方針』も欠けている。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — animal_species 更新削除に監査なし。invariants:31。
#### POC-08: clinic スコープ付き飼主更新 route 5本が OpenAPI 未宣言で、:clinic_id も handler から読まれない — MEDIUM
- 区分: 新規
- 規約: `backend/CODING_RULES.md:53` 「OpenAPI contract と route、request、response、status code を同期する。」
- 対象: `backend/internal/owner/http_routes.go:24`、`backend/internal/owner/http_routes.go:25`、`backend/internal/owner/http_routes.go:29`、`backend/internal/owner/http_owner.go:135`、`backend/internal/owner/http_owner.go:159`、`backend/internal/owner/http_owner.go:234`
- 内容: protected.Group("/clinics/:clinic_id/owners") 配下に line-user-id / delivery-exclusion / delivery-caution / transfer-status / line-id-confirm の5 PATCH が登録されているが、対応する handler は httpapi.ExtractClinicID(c)（http_owner.go:135 等）で認証済み clinic を使うため :clinic_id を一度も読まない。git show HEAD:backend/docs/api.yaml で確認したところ、この5パスは OpenAPI に宣言されていない（宣言があるのは /clinics/{clinic_id}/owners/aggregations と .../lstep/friend-attributes のみ）。認可上は fail-closed（別 clinic の owner id を指定すると UpdateScopedByID の RowsAffected 0 で 404）だが、client からは「clinic を指定できる」ように見えて実際は無視されるため、#86 拠点横断 UI が選択中クリニックを URL に載せた場合に原因不明の 404 になる。
- 修正: 5本を撤去して /owners/:id 系に一本化するか、残すなら CreateOwner（http_owner.go:71-76）と同じく :clinic_id を httpapi.AuthorizeClinicIDs で検証して実際のスコープに使い、いずれの場合も api.yaml と routes_snapshot_test を同期させる。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — clinic スコープ owner route 5 本が OpenAPI 未宣言。CODING_RULES.md:53。
#### POC-10: 締め時刻の既定値 3 種が clinic_settings_repository に 6 重ハードコードされ、DDL DEFAULT と二重管理になっている — MEDIUM
- 区分: 新規
- 規約: `~/.claude/rules/ecc/common/coding-style.md:75` 「Use named constants for meaningful thresholds, delays, and limits.」
- 対象: `backend/internal/clinic/clinic_settings_repository.go:27`、`backend/internal/clinic/clinic_settings_repository.go:60`、`backend/internal/clinic/clinic_settings_repository.go:81`、`backend/internal/clinic/clinic_settings_repository.go:111`、`backend/internal/clinic/clinic_settings_repository.go:142`、`backend/internal/clinic/clinic_settings_repository.go:190`、`backend/internal/clinic/closing_settings_service.go:48`、`backend/internal/clinic/closing_settings_service.go:49`
- 内容: "14:00" / "18:30" / "17:30" が FindByClinicID のフォールバックと5つの UpdateXxx メソッドに計6回リテラルで書かれている。設定行が未作成のクリニックに対して UpdateCPMVersion / UpdateDormantThresholds / UpdateCPMV*Thresholds / UpdateHealthPreventionThresholds のいずれかが先に走ると、これらのリテラルがそのクリニックの実際の締め時刻として INSERT される。実測では現行 DDL DEFAULT（001_init.sql の closing_am_pm_boundary '14:00' / closing_weekday_end '18:30' / closing_sunday_end '17:30'）と一致しているため現時点の不整合は無いが、DDL 側だけを変更すると6箇所の Go リテラルが黙って正本になる。第4の値 ClosingAmStart は closing_settings_service.go:48-49 で defaultClosingAmStart 定数化され「migration 011 の DB default と一致させる」と規約化されており、残り3値だけがこの規律から外れている。根拠区分は project quality policy（Go/Gin 公式要件ではない）。
- 修正: defaultClosingAmStart と同じ形で defaultClosingAmPmBoundary / defaultClosingWeekdayEnd / defaultClosingSundayEnd を定義して6箇所を置換し、DDL DEFAULT との一致を検証する lintscan テストを追加する。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — 締め時刻既定値 6 箇所ハードコード。coding-style:75。
#### POC-11: ペット列挙値バリデータ4関数が owner / pet で完全複製され、既に構造ドリフトしている — MEDIUM
- 区分: 新規 ／ 横断パターン: X-09
- 規約: `~/.claude/rules/ecc/common/coding-style.md:26` 「Avoid copy-paste implementation drift」
- 対象: `backend/internal/owner/validators.go:108`、`backend/internal/owner/validators.go:117`、`backend/internal/owner/validators.go:126`、`backend/internal/owner/validators.go:136`、`backend/internal/pet/validators.go:14`、`backend/internal/pet/validators.go:27`、`backend/internal/pet/validators.go:40`、`backend/internal/pet/validators.go:54`
- 内容: validatePetGender / validatePetStatus / validatePetAcquisitionType / validatePetDangerLevel が owner と pet の両パッケージに同名・同一列挙集合で存在する。両者は同一の判定を意図しているにもかかわらず既に構造が分岐しており（owner 版は switch の case に "" を含める形、pet 版は先頭 early return する形）、model 側に列挙値が追加された際に片方だけ更新されると、POST /pets は通るが POST /owners のネストペットは 400 になる（またはその逆）という経路依存の受理差が生じる。なお phase2.html:234 が統合を禁じている isDog/isCatSpeciesName と doseSpeciesAliases は「契約が意図的に異なる」ケースであり、本件はその逆で、同一契約であるべき複製が既にドリフトしている点で性質が異なる。根拠区分は project quality policy。
- 修正: 4関数を sharedkernel（既に ValidateRequiredName 等を持つ）へ1本化し、owner/pet の両方から参照する。
- 検証時の補正: 「既に構造ドリフトしている」は構造レベルの記述としては正しいが、現時点の挙動は両者等価である（owner 版は switch の case に "" を含め、pet 版は先頭 early return するが、いずれも空文字を許容し同一列挙集合を受理する）。したがって現在の受理差は存在せず、リスクは将来の片側更新による分岐である。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **DOWNGRADED** → severity **LOW** — validator 複製は将来 drift リスク。現状 accept set は同等。
#### POC-12: pet / owner / clinic の JSON エンドポイントに request body サイズ上限が無い — MEDIUM
- 区分: 新規 ／ 横断パターン: X-07
- 規約: `.claude/rules/go-gin-backend-guidelines.md:180` 「- rate limit、request/body/upload size、content type、file path を制限する。」

- 対象: `backend/internal/pet/pet_handler.go:183`、`backend/internal/owner/http_owner.go:62`、`backend/internal/clinic/clinic_handler.go:109`、`backend/internal/auth/http_binding.go:30`、`backend/internal/staff/http_binding.go:30`、`backend/internal/billing/billing_confirmation_handler.go:135`、`backend/cmd/api/main.go:203`
- 内容: auth・staff・billing・lstep・medicalrecord・scheduler の6パッケージは http.MaxBytesReader で JSON body を明示的に上限化しているのに対し、pet / owner / clinic の全 ShouldBindJSON 呼び出しには上限が無い。cmd/api/main.go:198-203 のグローバル middleware にも body size 制限は含まれていない（Recovery / SecurityHeaders / RequestID / CORS / RequestLogging / SanitizeNullBytes のみ）。認証済みスタッフ1名が POST /pets や POST /owners に巨大 body を送るだけでプロセスメモリを消費でき、text 型カラム（remarks 等）はそのまま永続化される。
- 修正: staff/http_binding.go と同型の bindJSON helper を pet / owner / clinic に置くか、protected グループ全体へ body size middleware を1本入れて全パッケージの扱いを統一する。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — pet/owner/clinic JSON に MaxBytesReader なし。規約は guidelines:180。
#### POC-13: pet / owner の自由記述フィールドに長さ検証が無い（同一 struct 内の他フィールドには存在する） — MEDIUM
- 区分: 新規 ／ 横断パターン: X-02
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/pet/pet_request.go:106`、`backend/internal/pet/pet_request.go:110`、`backend/internal/pet/pet_request.go:111`、`backend/internal/pet/pet_request.go:119`、`backend/internal/pet/pet_request.go:120`、`backend/internal/pet/pet_request.go:121`、`backend/internal/pet/pet_request.go:123`、`backend/internal/pet/pet_request.go:112`、`backend/internal/pet/pet_request.go:113`、`backend/internal/owner/http_request.go:167`、`backend/internal/owner/http_request.go:169`、`backend/internal/owner/http_request.go:171`、`backend/internal/owner/http_request.go:179`
- 内容: createPetRequest の NameKana / Breed / Color / Food / Environment / Phone / Remarks、createOwnerRequest の OwnerNameKana / Company / Address1 / Remarks 等に binding の max= が無く、service 側にも長さ検証が無い。同じ struct の BloodType（pet_request.go:112 max=32）と MicrochipNumber（:113 max=64）には上限があり、Name は sharedkernel.ValidateRequiredName（sharedkernel/validators.go:35）で MasterNameMaxLength まで、AcquisitionType / DangerLevel は service の列挙値検証で守られているため、境界検証の意図自体は存在する。上限のあるフィールドと無いフィールドが同一 struct に混在している状態であり、PO-12（body size 上限不在）と組み合わさると任意長の入力が DB の text カラムへ到達する。
- 修正: 自由記述フィールドに binding:"omitempty,max=N" を付与する。列ごとの上限値は 001_init.sql の型定義（varchar(N) を持つ列はその値、text 列は運用上の妥当値）から導出し、pet と owner で同一フィールドは同一値に揃える。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — 自由記述に max なし。guidelines:151。
#### POC-14: 慢性疾患 PATCH に「更新対象フィールド0件」のガードが無い（兄弟実装3箇所には存在する） — MEDIUM
- 区分: 新規 ／ 横断パターン: X-02
- 規約: `~/.claude/rules/ecc/common/coding-style.md:26` 「Avoid copy-paste implementation drift」（兄弟 Update の empty-fields ガード非対称）

- 対象: `backend/internal/pet/chronic_condition_service.go:116`、`backend/internal/pet/chronic_condition_service.go:117`、`backend/internal/pet/chronic_condition_repository.go:78`、`backend/internal/pet/service.go:380`、`backend/internal/owner/service_core.go:125`、`backend/internal/clinic/company_service.go:109`
- 内容: chronicConditionService.Update は buildChronicConditionUpdateFields の戻り値を長さ検査せずに repo.Update へ渡す。同一 repo 内の petService.Update（service.go:380-382）、ownerService.Update（service_core.go:125-127）、companyService.Update（company_service.go:109-111）はいずれも len(fields)==0 を WrapInvalidInput で弾いており、慢性疾患だけが逸脱している。全フィールド未指定の PATCH（例 `{}`）は境界で拒否されず repository まで到達し、RowsAffected==0 分岐（chronic_condition_repository.go:78-80）で WrapNotFound となって「存在しない」旨の 404 が返る見込みである（GORM が空 assignment set で SQL を発行しない挙動に依存するため、この観測結果は read-only 環境では実行検証していない）。
- 修正: chronic_condition_service.go:116 の直後に len(fields)==0 → apperrors.WrapInvalidInput(sharedkernel.ErrMsgAtLeastOneField) を追加し、兄弟実装と同じ contract に揃える。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — 空 PATCH ガード欠落は実在。peer contract（X-09）が正。
#### POC-15: 同一 error を同一関数内で2回 ErrorContext ログしている — MEDIUM
- 区分: 新規 ／ 横断パターン: X-10
- 規約: `backend/CODING_RULES.md:67` 「同じ error を複数箇所で重複ログしない。十分な request 文脈を持つ境界で1回記録する。」
- 対象: `backend/internal/clinic/closing_settings_service.go:271`、`backend/internal/clinic/closing_settings_service.go:275`
- 内容: UpdateSpecialPeriod の失敗分岐で、slog.ErrorContext(ctx, "failed to update closing special period", ...slog.Any("error", err)) の直後に slog.ErrorContext(ctx, "failed to update special period", "error", err) を連続で呼んでおり、同じ err が2レコード出力される。後者は clinic_id/id を持たず情報量も少ない。
- 修正: closing_settings_service.go:275 の重複行を削除し、clinic_id と id を含む :271-274 の1回に統合する。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — 同一 error 二重 ErrorContext。CODING_RULES.md:67。
#### POC-17: clinic / company の email・電話番号・郵便番号が未検証（owner の同名フィールドは検証済み） — MEDIUM
- 区分: 新規 ／ 横断パターン: X-02
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/clinic/clinic_request.go:16`、`backend/internal/clinic/clinic_request.go:18`、`backend/internal/clinic/clinic_request.go:22`、`backend/internal/clinic/clinic_request.go:50`、`backend/internal/clinic/company_request.go:5`、`backend/internal/clinic/company_request.go:7`、`backend/internal/clinic/company_request.go:9`、`backend/internal/owner/validators_contact.go:19`、`backend/internal/owner/validators_contact.go:31`、`backend/internal/owner/validators_contact.go:43`
- 内容: CreateClinicRequest / UpdateClinicRequest / UpdateCompanyRequest の Email・PhoneNumber・FaxNumber・PostalCode には binding tag が無く、clinicService.UpdateClinic / companyService.Update にも形式検証が無い（BuildClinicUpdate が検証するのは税率と section key のみ）。同じリポジトリの owner は validators_contact.go:19/31/43 で email・電話番号・郵便番号を正規表現検証しており、同一種のフィールドに対する境界検証の有無が domain ごとに割れている。clinic / company の値は会計帳票のヘッダ（accounting_document_show_clinic_header 経路）に出力されるため、不正形式がそのまま帳票へ載る。
- 修正: owner/validators_contact.go の3関数を sharedkernel へ移して clinic / company の Update 入力でも呼ぶ（PO-11 の統合と同じ移設先に揃える）。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — clinic/company email/phone 未検証。guidelines:151。
#### POC-09: clinic_settings / clinic_holiday の書き込みで Scopes(ClinicScope) が INSERT に適用されず、テナントガードが不活性 — LOW
- 区分: 新規 ／ 検証で severity 引き下げ
- 規約: `.claude/refs/backend-application-invariants.md:11` 「clinic-scoped data のすべての read/write/delete は、認証済み `clinic_id` で制約する。」
- 対象: `backend/internal/clinic/clinic_settings_repository.go:39`、`backend/internal/clinic/clinic_settings_repository.go:66`、`backend/internal/clinic/clinic_settings_repository.go:90`、`backend/internal/clinic/clinic_settings_repository.go:120`、`backend/internal/clinic/clinic_settings_repository.go:160`、`backend/internal/clinic/clinic_settings_repository.go:197`、`backend/internal/clinic/clinic_holiday_repository.go:38`、`backend/internal/persistence/scope.go:16`、`backend/internal/clinic/ports.go:71`
- 内容: これら7箇所は Scopes(persistence.ClinicScope(clinicID)) を付けたうえで Create(...) + clause.OnConflict を実行する。ClinicScope は db.Where("clinic_id = ?")（scope.go:16-19）を追加するだけで、GORM の INSERT ... ON CONFLICT 文には WHERE 句が出ないため、この tenant guard は一度も実行されない。実際に書き込み先を決めるのは引数 clinicID ではなく struct の ClinicID フィールドであり、Save(ctx, clinicID, s) は s.ClinicID == clinicID を assert しない。現行の唯一の呼び出し元（closing_settings_service.go:143 で FindByClinicID から取得した current を渡す）では両者が一致するため実害は発火しない。ただし ClinicSettingsRepository は ports.go:69-71 で「LSTEP consumer と共有する port」として export されており、将来の呼び出し元が不一致な struct を渡した場合に見た目上のガードは何も止めない。
- 修正: Save / Update* の先頭で s.ClinicID を引数 clinicID で上書きするか不一致を InternalServerError で拒否し、効果の無い Scopes(ClinicScope(...)) は削除して誤解を除く。
- 検証時の補正: clinic-isolation の不変条件違反は成立しない。書き込み先を決める struct の ClinicID は、UpdateCPMVersion(:58)・UpdateDormantThresholds(:78)・UpdateCPMV2Thresholds(:108)・UpdateCPMV1Thresholds(:139)・UpdateHealthPreventionThresholds(:187) のいずれも引数 clinicID をそのまま代入しており、clinic_holiday_service.go:38-42 も ClinicID: clinicID を設定、Save の唯一の呼び出し元も FindByClinicID(clinicID) の戻り値を渡す。よって全経路で認証済み clinic_id により制約されており invariants:11 は充足している。正確な指摘内容は『INSERT ... ON CONFLICT に WHERE は出力されないため Scopes(persistence.ClinicScope(clinicID)) が7箇所で no-op であり、テナントガードが効いているように誤読させる死んだコードになっている』であり、分類は clinic-isolation ではなく可読性/誤解防止（LOW）。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **REFRAMED** — ClinicScope on INSERT は no-op だが ClinicID は arg から設定され isolation は成立。dead scope。
#### POC-16: parseHHMM が billing の同名 helper の自己申告済み複製として残り、削除条件が無い — LOW
- 区分: 新規 ／ 横断パターン: X-09 ／ 検証で severity 引き下げ
- 規約: `~/.claude/rules/ecc/common/coding-style.md:26` 「Avoid copy-paste implementation drift」

- 対象: `backend/internal/clinic/closing_settings_service.go:383`、`backend/internal/clinic/closing_settings_service.go:384`、`backend/internal/clinic/closing_settings_service.go:385`
- 内容: 関数直上のコメントが「billing/cash_register_service.go の同名helperの複製。clinic domain 移行時に片側へ統合」と複製であることを自認しているが、移行完了の判定条件も削除期限も設定されていない。締め時刻の "HH:MM" / "HH:MM:SS" 正規化は締め集計の判定に直結する business rule であり、片側だけが修正された場合（例: 秒付き入力の扱い変更）に billing 側の締め判定と clinic 側の設定検証が食い違う。
- 修正: sharedkernel へ1本化して billing / clinic 双方から参照するか、統合できないなら「どちらを正本とし、いつ削除するか」を条件付きでコメントに明記して台帳へ起票する。
- 検証時の補正: 引用規約が対象を取り違えている。backend/CODING_RULES.md:40 の主語は「compatibility facade」であり、parseHHMM は facade でも delegate でもなく package 内 private の純粋パース関数である。また『締め時刻の HH:MM 正規化は business rule』という性格付けも過大で、実体は文字列長判定と Sscanf だけの15行（clinic/closing_settings_service.go:385-399 と billing/cash_register_service.go:462-476 はバイト等価）。指摘の実質は移行期の重複helper が削除条件を持たない点であり、適用すべき根拠は ~/.claude/rules/ecc/common/coding-style.md:26「Avoid copy-paste implementation drift」（根拠区分 project quality policy）、severity は LOW が妥当。
- round3-review(2026-07-28): **REFRAMED** — parseHHMM 複製。CODING_RULES.md:40（facade）は適用外。coding-style:26 の DRY メモ。
### trimming / manualarticle / csvimport（10件）

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
#### TRM-01: manualarticle Upsert がcommit後に再取得し、commit済みの成功を失敗応答へ反転させる（監査も欠落） — HIGH
- 区分: 新規 ／ 横断パターン: X-01
- 規約: `backend/CODING_RULES.md:78` 「write後の再取得が失敗し得る場合はcommit前の同じtransaction内で行うか、commit済みの成功を後段read errorで失敗へ反転させないcontractにする。」
- 対象: `backend/internal/manualarticle/repository.go:63`、`backend/internal/manualarticle/repository.go:109`、`backend/internal/manualarticle/repository.go:114`、`backend/internal/manualarticle/service.go:102`、`backend/internal/manualarticle/handler.go:96`、`backend/internal/manualarticle/handler.go:104`
- 内容: repository.go:63 で開いた Transaction は :109 で閉じ、最新行の再取得 FindByCategoryAndSlug は :114 の transaction 外で実行される。この read が失敗すると service.go:102 が error を wrap し handler.go:96 が RespondError を返すが、manual_articles の UPDATE/INSERT と manual_article_versions の履歴行は既に commit 済みである。同一 package の trimming は同じ再取得を tx 内（trimming_service.go:330/471/562/736）で行っており、契約が揃っていない。
- 修正: Transaction closure 内で最新行を読み、その値を戻り値として closure 外へ引き渡す（trimming_service.go:330 と同じ形）。
- 検証時の補正: detail の「同一 package の trimming は同じ再取得を tx 内で行っており」は不正確。trimming は別 package(internal/trimming)であり、正しくは「同一 audit unit の internal/trimming が trimming_service.go:330 / :471 / :562 / :736 で同じ再取得を tx 内 closure で行い、結果を closure 外の変数へ引き渡している」。また title の「（監査も欠落）」は過大。監査自体は handler.go:103-118 に存在する。正確には「post-commit read が失敗すると service→handler が error 応答へ分岐するため、commit 済みの write が監査 LogEntry を一度も書かないまま終わる」。

- 再実測(2026-07-28): **CONFIRMED** — manualarticle Upsert Transaction 後 FindByCategoryAndSlug。
- round3-review(2026-07-28): **UPHELD** — manualarticle Upsert が commit 後 re-fetch。CODING_RULES.md:78。監査欠落は過大（成功時は best-effort 監査あり）。
#### TRM-03: trimming/manualarticle の自由入力文字列に長さ上限が一切なく、request body上限も無い — HIGH
- 区分: 新規 ／ 横断パターン: X-07
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/trimming/trimming_request.go:70`、`backend/internal/trimming/trimming_request.go:76`、`backend/internal/trimming/trimming_request.go:78`、`backend/internal/manualarticle/request.go:10`、`backend/cmd/api/main.go:198`、`backend/internal/staff/staff_request.go:20`、`backend/migrations/001_init.sql:2853`
- 内容: 実測で本2 packageの string field に max= は0件（slice長のOptionIDs max=50 のみ）。一方 repo 全体では binding の max= が98件あり、staff_request.go:20 の reservation_comment は max=2000 と、直接比較できる自由入力欄が bound されている。cmd/api/main.go:198-203 の全体 middleware に body size 制限は無く、JSON binding 境界で http.MaxBytesReader を張る先例は auth/http_binding.go:30・staff/http_binding.go:30・billing/billing_confirmation_handler.go:135 の3件存在する（他6件はupload/webhook用）。格納先は 001_init.sql:1362-1368 / :2853 いずれも無制限 text。manualarticle は Upsert のたびに repository.go:96-106 で全文スナップショットを manual_article_versions へ追記するため増幅が効く。
- 修正: remarks/style_request/used_shampoo/used_ribbon/style_image/completed_image/title/section/body_markdown に binding の max= を付与し、両 handler の JSON bind 前に auth/staff と同形の http.MaxBytesReader を適用する。

- 再実測(2026-07-28): **CONFIRMED** — 自由文字列 max= 無し + body size middleware 無し。
- round3-review(2026-07-28): **UPHELD** — free-text 長さ上限なし・global body size なし。guidelines:151。
#### TRM-04: 既存detailを持つappointmentへのPOST /trimmingsが、何も書かずに201 Createdを返し送信値と監査を捨てる — HIGH
- 区分: 新規
- 規約: `backend/CODING_RULES.md:53` 「OpenAPI contract と route、request、response、status code を同期する。」
- 対象: `backend/internal/trimming/trimming_service.go:470`、`backend/internal/trimming/trimming_service.go:471`、`backend/internal/trimming/trimming_service.go:566`、`backend/internal/trimming/trimming_handler.go:97`、`backend/docs/api.yaml:12621`
- 内容: createDetailForExistingAppointment は trimming_service.go:470 で既存 detail を検出すると :471 で現在値を読んで return し、以降の Create／SetOptions／:566 の logTrimmingAuditTx をすべて飛ばす。request が載せた course_id・style_request・bw・bt・option_ids・remarks は黙って破棄されるが、handler は trimming_handler.go:97-98 で Location header 付き 201 を返す。api.yaml:12621-12622 は同 route の 201 を「登録成功」と定義しており、書き込みゼロの経路に成功 status が割り当てられている。監査も残らないため事後追跡もできない。
- 修正: 既存 detail 検出時は 409 Conflict（または PATCH への誘導）を返して api.yaml に追記するか、送信値で更新したうえで監査を記録する。いずれにせよ「何も書かずに201」は解消する。
- 検証時の補正: proposed_fix の「409 Conflict を返して api.yaml に追記する」のうち追記部分は不要。api.yaml:12640-12641 に POST /trimmings の '409' が既に `$ref: '#/components/responses/Conflict'` として宣言済みであり、契約側の枠は存在してコード側が使っていないだけである（この事実は所見をむしろ補強する）。

- 再実測(2026-07-28): **LINE-DRIFT** — existing detail early return :470-472。handler は常に 201。api.yaml 201 は :12826 近傍（旧:12621）。
- round3-review(2026-07-28): **UPHELD** — 既存 detail でも 201 Created。CODING_RULES.md:53 / OpenAPI 409 と不一致。
#### TRM-02: マニュアル削除は物理削除で編集履歴がCASCADE消滅し、監査はbest-effortかつ条件付きでスキップされる — MEDIUM
- 区分: 新規 ／ 横断パターン: X-03 ／ 検証で severity 引き下げ
- 規約: `.claude/refs/backend-application-invariants.md:31` 「destructive または irreversible な操作には、権限、対象 scope、監査、recovery 方針を持たせる。」
- 対象: `backend/internal/manualarticle/repository.go:117`、`backend/internal/manualarticle/repository.go:120`、`backend/migrations/001_init.sql:2873`、`backend/internal/manualarticle/handler.go:152`、`backend/internal/manualarticle/handler.go:163`
- 内容: model.ManualArticle に gorm.DeletedAt が無く repository.go:120 の Delete は物理削除である。001_init.sql:2873 の manual_article_versions.article_id は ON DELETE CASCADE のため、削除で全編集履歴が同時に消え recovery 経路が残らない。監査は handler.go:152 の ExtractStaffID が false のとき丸ごとスキップされ、成功時も :163 で LogEntry の error を slog.Warn するだけで 204 を返す。4要件のうち権限のみが成立し、監査と recovery 方針が欠ける。
- 修正: soft delete 化して履歴を保持するか、削除を business write と同一 transaction の fail-closed 監査（trimming の logTrimmingAuditTx 相当）に参加させ、actor 不明時は削除自体を拒否する。
- 検証時の補正: 残存する欠陥は次の2点に限定される。①001_init.sql:2873 の ON DELETE CASCADE により manual_article_versions の全編集履歴が消え、監査 old_value には記事の最終状態しか残らないため版履歴だけは復元不能（GET /manual/articles/:category/:slug/versions が提供する価値が失われる）。②handler.go:163-166 の監査は best-effort で、LogEntry が失敗しても slog.Warn のみで :169 が 204 を返すため、物理削除が成立したのに復元用 old_value が一切残らない状態になり得る。権限・対象 scope・記事本体の recovery 方針（audit old_value と bundled MD への回帰）は成立しているので「権限のみが成立」という記述は撤回する。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — マニュアル物理削除 + history CASCADE + 監査 best-effort。invariants:31。
#### TRM-05: bw_unit のみ境界で列挙値検証が無く、DB enum 到達まで不正値が進む — MEDIUM
- 区分: 新規 ／ 横断パターン: X-02
- 規約: `backend/CODING_RULES.md:79` 「schema constraint と application validation の両方を使う。」
- 対象: `backend/internal/trimming/trimming_request.go:72`、`backend/internal/trimming/trimming_request.go:128`、`backend/internal/trimming/trimming_service.go:224`、`backend/internal/trimming/trimming_service.go:707`、`backend/migrations/001_init.sql:1364`
- 内容: trimming_request.go:72／:128 の bw_unit は binding tag を持たず :112／:161 で model.BodyWeightUnit へ素キャストされ、service も :224-227／:707-709 で検証しない。同一ファイルの status・reservation_route・target_size はすべて oneof を持つため、enum で唯一の欠落である。001_init.sql:1364 の body_weight_unit enum が最終防波堤なので値の破損は起きないが、application validation が無いぶん Create では AcquireBookingLock 取得（trimming_service.go:270）と CreateForTrimming による appointment 行 INSERT（:288）を経た後にはじめて 22P02 由来の汎用 400「入力値の形式が正しくありません」で rollback する。
- 修正: bw_unit に binding:"omitempty,oneof=Kg g" を付与し、他 enum と同じく bind 時点で 400 を返す。
- 検証時の補正: detail の「同一ファイルの status・reservation_route・target_size はすべて oneof を持つ」は不正確。target_size は同一ファイルではなく trimming_course_request.go:8 と :32 にある（`binding:"omitempty,oneof=small medium large cat"`）。trimming_request.go 内で正確なのは「status(:66,:123) と reservation_route(:67) は oneof を持ち、bw_unit(:72,:128) だけが欠落している」。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — bw_unit 境界未検証。CODING_RULES.md:79。
#### TRM-06: trimming detail 生成失敗が同一call chainで二重にERRORログされる — MEDIUM
- 区分: 新規 ／ 横断パターン: X-10
- 規約: `.claude/refs/error-handling.md:12` 「同じ failure を複数層で重複ログしない。必要な文脈が揃う境界で1回記録する。」
- 対象: `backend/internal/trimming/trimming_service.go:554`、`backend/internal/trimming/trimming_service.go:577`、`backend/internal/trimming/trimming_service.go:345`、`backend/internal/trimming/trimming_service.go:751`
- 内容: trimmingDetail.Create 失敗時、:554 が tx 内で「failed to create trimming detail」を ERROR 出力し、同じ error を wrap して返した先の :577 が再び「failed to create trimming detail for existing appointment」を ERROR 出力する。1 failure に対し同一 package 内で 2 レコードが出る。付随して、:345 と :751 は Create／Update transaction 全体（予約作成・detail 作成・master 検証・監査を含む）の失敗を一律「failed to set options trimming」として記録しており、option 設定と無関係な失敗でも同じ文言になるため、重複ログの片方すら失敗箇所を指さない。
- 修正: tx 内 closure の ERROR ログを削って error 返却のみとし、記録は文脈の揃う外側1箇所に集約する。あわせて :345／:751 の message を各 operation 名に合わせる。
- 検証時の補正: detail 後半の「:345 と :751 は…一律『failed to set options trimming』として記録しており…重複ログの片方すら失敗箇所を指さない」という付随主張は、引用した error-handling.md:12（重複ログ禁止）ではカバーされず、log message の正確性を要求する規約原文を提示できない。規約引用を伴わない指摘は提出しない方針に従い、この一文は所見から削除する（:345 と :751 に当該文言が実在すること自体は実測で真）。所見は :554/:577 の二重 ERROR ログのみに限定する。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — detail 生成失敗の二重 ERROR。error-handling.md:12。
#### TRM-07: csvimport.Import は35表を無条件DELETEするexported APIだが、対象DBの同定手段を持たない — MEDIUM
- 区分: 新規 ／ 横断パターン: X-03
- 規約: `.claude/refs/backend-application-invariants.md:31` 「destructive または irreversible な操作には、権限、対象 scope、監査、recovery 方針を持たせる。」
- 対象: `backend/internal/csvimport/import.go:30`、`backend/internal/csvimport/import.go:69`、`backend/internal/csvimport/import.go:80`、`backend/internal/csvimport/cutover_import.go:174`、`backend/internal/csvimport/failure_rehearsal.go:42`、`backend/cmd/seed-export/main.go:115`
- 内容: deleteDemoGraph は import.go:80 で owners／pets／medical_records／billings／payments／appointments を含む35表へ述語なしの DELETE FROM を発行する。対象 scope 述語・監査・recovery 方針のいずれも無く、「disposable target database」であることは import.go:28-29 の doc comment が期待を書いているだけで実行時の同定手段が無い。同 package の姉妹 API は逆で、cutover_import.go:174-176 は既存行を一切削除しない契約を明示し、failure_rehearsal.go:42 は TargetDatabaseIdentitySHA256 による caller 側の証明を要求する。現行の唯一の呼び出し元 cmd/seed-export/main.go:115 は自プロセスが作成・削除する seed_export_tmp を渡すため実害は無く、これは潜在的な API 契約の欠落である。BUG-430（cmd/stage-import の医院非限定 DELETE）と同種だが対象 file／tool が異なるため新規として挙げる。
- 修正: Import に対象 DB 同定（current_database() 照合または caller 提供の identity digest）を required 引数として課し、不一致なら fail-closed にする。
- 検証時の補正: 表数が不正確。import.go:72-79 の削除リストを全数計上すると 35 表ではなく 37 表（:73=6, :74=7, :75=6, :76=5, :77=4, :78=9）。それ以外の記述は実測どおり。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — csvimport.Import の unconditional DELETE が exported API。invariants:31。
#### TRM-08: [既知] Options preload の clinic predicate が末尾 trimming_options にしか掛からない — LOW
- 区分: **既知** → SEC-SWEEP-02 ／ 検証で severity 引き下げ
- 規約: `.claude/refs/backend-application-invariants.md:39` 「nested `Preload`のpredicateは末尾associationだけに適用される。clinic-ownedの中間associationも独立したclinic predicateでscopeし、破損FKから他院の詳細・個人情報を復元しない。」
- 対象: `backend/internal/trimming/trimming_repository.go:30`、`backend/internal/trimming/trimming_repository.go:35`
- 内容: FindByAppointmentID は detail 本体を ClinicScope で、Course／Options を clinic_id predicate 付き Preload で絞るが、Options は appointment_trimming_options 経由の many2many であり、この中間 junction table は clinic_id 列を持たない（001_init.sql:1379-1386）。predicate は末尾の trimming_options にのみ適用されるため、junction 行そのものは親 appointment の clinic と相関されない。SEC-SWEEP-02 が残5面のひとつとして trimming_repository.go:30 を明示的に保持しており、DEC-23 で是正方針が裁定済みである。追加evidenceとしてのみ記録する。
- 修正: SEC-SWEEP-02 の既定方針（孫 read への親 clinic 相関の必須化）に従う。本 unit で独立に修正しない。
- 検証時の補正: 本項は SEC-SWEEP-02（3-session-agent.html:254、DEC-23 裁定済み）への補足evidenceであり、独立した所見として再起票してはならない。新規に確定した事実は1点のみ: 中間 junction table appointment_trimming_options（001_init.sql:1380-1387）に clinic_id 列が存在しない。一方 trimming_repository.go:35 の末尾述語は trimming_options 側に clinic predicate を掛けるため、本 finding の範囲では他院 option の復元経路は実証できていない。是正方針は SEC-SWEEP-02 の既定（孫 read への親 clinic 相関の必須化）に従い、本 unit で独立に修正しない。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **REFRAMED** — 既知 SEC-SWEEP-02 / DEC-23。parent JOIN は部分修正済み。
#### TRM-09: [WIP-adjacent] trimming の3 master repository で ambient transaction 参加可否が食い違う — LOW
- 区分: WIP-adjacent（清算後に再検証）
- 規約: `backend/CODING_RULES.md:77` 「transaction の開始、commit、rollback、resource ownership を明確にする。」
- 対象: `backend/internal/trimming/trimming_course_repository.go:55`、`backend/internal/trimming/trimming_course_repository.go:62`、`backend/internal/trimming/trimming_option_repository.go:55`、`backend/internal/trimming/trimming_option_repository.go:62`、`backend/internal/trimming/trimming_course_type_repository.go:62`、`backend/internal/trimming/trimming_course_type_repository.go:69`、`backend/internal/persistence/scope.go:74`
- 内容: 構造が同型の3 repository のうち、course は Create/Update とも persistence.DBOrTx を通すが、option と course_type は r.db を直接渡す。persistence.UpdateScopedByID は scope.go:74 で db.WithContext(ctx) をそのまま使い内部で DBOrTx を呼ばないため（ReorderByClinicID:128 とは異なる）、r.db 版は ambient transaction に参加しない。現時点で該当 method を WithTx で包む呼び出し元は存在せず実害は無いが、interface 上は同じ契約に見えるため、将来 course と同様に tx 化した際に write が黙って tx 外へ逃げる。本件は他セッションが編集中の backend/internal/lintscan/dbortx_inventory_lint_test.go と同一論点であり WIP-adjacent として扱う。清算後に再検証すること。
- 修正: 清算後、3 repository の tx 参加可否を DBOrTx へ統一する（または参加しない設計を doc comment で明示する）。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **REFRAMED** — WIP-adjacent / dbortx inventory。現呼び出しに ambient tx 無し。
#### TRM-10: package 名と export 名の stutter — LOW
- 区分: 新規
- 規約: `.claude/rules/go-gin-backend-guidelines.md:50` 「export 名に package 名を繰り返さない。例: `http.HTTPServer` のような stutter を避ける。」
- 対象: `backend/internal/trimming/trimming_service.go:96`、`backend/internal/trimming/trimming_course_service.go:76`、`backend/internal/manualarticle/service.go:26`、`backend/internal/manualarticle/response.go:11`
- 内容: trimming.TrimmingService / trimming.TrimmingCourseService / manualarticle.ManualArticleService / manualarticle.ManualArticleResponse のように、呼び出し側で package 名が二重になる。BE9 の domain package 化で package 名が付いた後も、旧 layer package 時代の型名がそのまま残っていることによる。
- 修正: 外部 consumer の移行コストと釣り合う範囲で trimming.Service / manualarticle.Service 等へ改名する。API wire 形状には影響しない。
- round3-review(2026-07-28): **WITHDRAWN** — export stutter は style preference。欠陥扱いは過剰。
### model / lintscan / testdb（モデル・静的lint）（5件）

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
#### MDL-01: EstimateItem.CalculateTaxAmount が DiscountAmount を課税ベースから引かず、BillingItem の #85 実装と乖離している — MEDIUM
- 区分: 新規 ／ 横断パターン: X-09
- 規約: `/Users/minoru/.claude/rules/ecc/common/coding-style.md:26` 「- Avoid copy-paste implementation drift」
- 対象: `backend/internal/model/estimate.go:82`、`backend/internal/model/estimate.go:83`、`backend/internal/model/accounting.go:151`、`backend/internal/model/accounting.go:152`、`backend/internal/billing/estimate_response.go:52`、`backend/internal/billing/estimate_response.go:53`、`backend/internal/csvimport/cutover_contract.go:140`、`backend/internal/csvimport/cutover_import.go:143`
- 内容: model/accounting.go:152 の BillingItem は #85 に従い `subtotal := max(UnitPrice*Quantity - DiscountAmount, 0)` を課税ベースにするが、同型の model/estimate.go:83 の EstimateItem は `subtotal := UnitPrice*Quantity` のままで DiscountAmount を無視する。EstimateItem は DiscountRate/DiscountAmount 両フィールドを持ち（estimate.go:61-62）、DDL の estimate_items にも discount_amount 列が存在する（001_init.sql:1816）。同じ税計算がさらに internal/billing/billing_service.go:15 にも第3の実装として存在し、そこでも割引が反映されない。
- 修正: model/estimate.go:83 を accounting.go:152 と同一式（max(単価×数量−割引額, 0)）へ揃え、estimate_response.go:52 の subtotal も割引後へ合わせる。恒久策として BillingItem/EstimateItem/billing_service.go の3実装を単一の課税ベース算出関数へ統合し、model 側 method はそれを呼ぶだけにする。
- 検証時の補正: detail の2点を訂正する。(1)「第3の実装が internal/billing/billing_service.go:15 に存在しそこでも割引が反映されない」は誤導。同関数のシグネチャは CalculateTaxAmount(unitPrice, quantity, taxType, taxRate) で割引引数を持たず、production 呼び出し元はゼロ（実測: 参照は billing_service_test.go:79 のみ）。実質の重複は model 側2メソッドの2実装であり、billing_service.go:15 は未使用の第3コピーとして「削除候補」と書くのが正確。(2) 現時点の実害は潜在。estimate_items への production write path は存在せず（EstimateItem の Create は csvimport 経路以外に無い）、seed 003_demo/estimate_items.csv は discount_amount>0 の行が0件。よって現在の応答値は誤っていない。将来 csvimport cutover（cutover_contract.go の estimate_items spec に discount_amount を含む）や明細API追加で顕在化する latent drift である。また proposed_fix には前提が必要: 見積の subtotal/tax_total はクライアント供給値をそのまま保存する契約（estimate_request.go:52-56 → estimate_service.go:263-267 で再計算しない）ため、明細側だけを割引後へ揃えると既存行で「明細合計 ≠ 保存済み見積合計」が生じる。API contract 上 behavior-preserving ではないので、totals の扱いと同時に決める必要がある。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — EstimateItem 税が Discount 無視（BillingItem は考慮）。潜在・現状 discount write 無し。coding-style:26。
#### MDL-02: testdb のテスト間データ分離 TRUNCATE が error を全件破棄し fail-open している — MEDIUM
- 区分: 新規 ／ 横断パターン: X-04
- 規約: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.claude/refs/error-handling.md:9` 「- error を無視しない。処理できる境界まで返すか、明示的に回復する。」
- 対象: `backend/internal/testdb/testdb.go:141`、`backend/internal/testdb/testdb.go:142`、`backend/internal/testdb/testdb.go:143`、`backend/internal/testdb/testdb.go:144`、`backend/internal/testdb/testdb.go:145`、`backend/internal/testdb/testdb.go:232`、`backend/internal/testdb/testdb.go:236`、`backend/internal/testdb/testdb.go:496`
- 内容: SetupTestDB（:141-145）と SetupIsolatedTestDB（:232-236）は `db.Exec("TRUNCATE TABLE ... CASCADE")` を戻り値を受けずに10箇所で呼び、error を完全に破棄する。testdb.go:32 が「TRUNCATE のみ SetupTestDB 内で呼び出し毎に実行し、テスト間データ分離を維持する」と明記する通り、この5行が全 real-DB integration test の分離保証そのものである。SetupIsolatedTestDB は同一物理DB（ekarte_db_test）上で checkup_field クラスタの DROP+CREATE を許容する設計（:205-217）なので、テーブル不在・ロック競合時に TRUNCATE が失敗しても誰も気付かない。:496 の main DB 接続失敗も stderr 警告のみで続行する。
- 修正: 5本の TRUNCATE を1つのヘルパへ集約し `require.NoError(t, db.Exec(...).Error)` で fail-closed にする（SetupTestDB/SetupIsolatedTestDB とも *testing.T を持っているため追加引数不要）。:496 は警告のままにするなら「test DB 接続で必ず失敗する」ことをコメントで契約化する。
- 検証時の補正: evidence の testdb.go:496 は fail-open の例として不成立なので detail から外すべき。実測: mainDSN(:491) と testDSN(:516) は dbname 以外（host/port/user/password）が同一。したがって :494 の main DB open が失敗する状況では :520 の test DB open も同じ理由で失敗し、:521-522 が wrap した error を返して fail-closed になる。唯一 :496 の警告続行が効くのは「ekarte_db は不在だが ekarte_db_test は存在する」ケースで、その場合 CREATE DATABASE をスキップして進むのは無害。よって本所見の成立範囲は :141-145 / :232-236 の10箇所の TRUNCATE error 破棄のみ。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **WITHDRAWN** — testdb TRUNCATE の error 破棄は test harness 品質。product MEDIUM として不成立。
#### MDL-04: 配信トリガー優先順位の trigger_type が列挙値検証を受けず、任意文字列が永続化される — MEDIUM
- 区分: 新規 ／ 横断パターン: X-02
- 規約: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.claude/rules/go-gin-backend-guidelines.md:151` 「- 外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/lstep/lstep_trigger_priority_request.go:5`、`backend/internal/lstep/lstep_trigger_priority_service.go:78`、`backend/internal/lstep/lstep_trigger_priority_service.go:79`、`backend/internal/lstep/lstep_trigger_priority_service.go:83`、`backend/internal/model/lstep_trigger_priority.go:9`、`backend/internal/model/lstep_delivery_trigger_log.go:6`、`backend/internal/model/lstep_delivery_trigger_log.go:41`、`backend/migrations/001_init.sql:521`
- 内容: request DTO は `TriggerType string binding:"required"` のみ（lstep_trigger_priority_request.go:5）で `oneof=` が無く、service も `if it.Priority < 1` しか検証しない（:78-79）。DDL 側も `trigger_type VARCHAR(64) NOT NULL` で CHECK も FK も無い（001_init.sql:521）ため、DB も fail-closed にならない。model 側の型もこれを一切支えていない: model.TriggerType は `type TriggerType = string` という型エイリアス（lstep_delivery_trigger_log.go:6）でコンパイル時保護がゼロ、model.LstepTriggerPriority.TriggerType は素の string（lstep_trigger_priority.go:9）、正の列挙は model.AllTriggerTypes()（lstep_delivery_trigger_log.go:41-59）に存在するがどの境界でも強制されていない。失敗シナリオ: PATCH /api/v1/lstep/trigger-priorities に `{"items":[{"trigger_type":"vaccine_deadline_60","priority":1}]}`（正しくは vaccine_deadline_60d）を送ると 200 OK で行が永続化されるが、GET は AllTriggerTypes() を基準に構築するため画面に現れず、GetPriorityFor も実トリガー名でしか引かないので抑止順序に一切反映されない。運用者は設定したつもりでサイレントに無効。
- 修正: lstep_trigger_priority_request.go:5 に model.AllTriggerTypes() 由来の membership 検証（binding oneof か service 側での明示チェック）を入れ、未知値を 400 で拒否する。併せて model 側で `type TriggerType string`（defined type）へ格上げし、AllTriggerTypes() を唯一の正本にする。
- 検証時の補正: proposed_fix の後段（`type TriggerType string` への defined type 格上げ）は本 fix と同一単位で実施できない点を明記すべき。model.TriggerType は tygo 生成対象で frontend/src/types/generated/models.ts へ波及し、`make codegen` は .claude/CLAUDE.md の自動実行禁止コマンドかつ台帳上 USER 専権のため、codegen ゲート付きの別タスクになる。本所見で今すぐ閉じられるのは境界での membership 検証（request DTO の binding oneof、または service UpdatePriorities 内で model.AllTriggerTypes() 集合に対する照合を行い未知値を 400 で拒否）までである。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **WITHDRAWN** — LSA-13 と同一 trigger_type 欠陥の重複（X-02 も MDL-04 除外済み）。
#### MDL-05: CASCADE lint が完全一致リテラル検索のため、表記ゆれした ON DELETE CASCADE を見逃す（偽陰性） — MEDIUM
- 区分: 新規
- 規約: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/migrations/CLAUDE.md:18` 「**禁止（絶対）**: `owners` / `pets` / `medical_records` 等の PHI・業務データが親となるCASCADEで、削除により診療履歴・会計・バイタル等が連鎖消去されうる設計は禁止。」
- 対象: `backend/internal/lintscan/migration_cascade_lint_test.go:62`、`backend/internal/lintscan/migration_cascade_lint_test.go:63`、`backend/internal/lintscan/migration_cascade_lint_test.go:64`、`backend/internal/lintscan/migration_cascade_lint_test.go:74`、`backend/internal/lintscan/migration_cascade_lint_test.go:120`
- 内容: countCascadeOccurrences は `strings.Count(sql, "ON DELETE CASCADE")`（:62-64）だけで判定する。PostgreSQL は `on delete cascade`（小文字）、`ON DELETE  CASCADE`（空白2つ）、`ON DELETE\n    CASCADE`（改行分割）をすべて等価に受理するが、いずれもこのカウンタでは 0 になる。:74 の分岐は「新規ファイルで count > 0」のときだけ違反にするため、表記ゆれした PHI 親テーブルへの CASCADE を含む新規 migration は gate を無条件で通過する。既知として台帳にあるのは「コメント内の字句も数える」偽陽性の罠であり、本件はその逆方向（偽陰性）で未記録。
- 修正: countCascadeOccurrences を `regexp.MustCompile(`(?is)ON\s+DELETE\s+CASCADE`)` の FindAllStringIndex 件数へ置換する。既存 allowlist 値（001_init.sql: 54）は正規化後の実測で再ピンし、TestReconcileMigrationCascade_Analyzer に小文字・複数空白・改行分割の fixture を追加する。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — CASCADE lint が完全一致のみで false-negative。lint 品質欠陥。
#### MDL-06: model.Payment に ClinicID が無く、同一 business fact の sibling である PaymentSplit だけが clinic_id を持つ — LOW
- 区分: **既知** → TASK-445 ／ 検証で severity 引き下げ
- 規約: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.claude/refs/backend-application-invariants.md:11` 「- clinic-scoped data のすべての read/write/delete は、認証済み `clinic_id` で制約する。」
- 対象: `backend/internal/model/accounting.go:163`、`backend/internal/model/accounting.go:191`、`backend/internal/model/accounting.go:196`、`backend/internal/model/accounting.go:198`、`backend/migrations/001_init.sql:1946`、`backend/migrations/001_init.sql:1970`
- 内容: model.Payment（accounting.go:163-191）は clinic_id フィールドを持たず、DDL の payments（001_init.sql:1946-1966）にも clinic_id 列が無い。一方、同じ会計の支払内訳を表す PaymentSplit（accounting.go:196-212 / DDL :1970-1990）は `ClinicID uint64 gorm:"not null"` + FK + 複合 index を持つ。結果として payments に対する全 read/write は billings 経由の相関でしか clinic 制約を表現できず、model 型の上では clinic 述語を書く手段が存在しない。既に TASK-445（DEC-28）が「payments.clinic_id」を対象として起票済み。
- 修正: TASK-445 の宣言的複合FK掃引で payments.clinic_id を追加し、同時に model.Payment へ `ClinicID uint64 gorm:"not null"` を追加して PaymentSplit と表現を揃える。それまでは payments を触る repository が必ず billings への clinic 相関 join を持つことを既存 grandchild lint の対象へ含めるか確認する。
- round3-review(2026-07-28): **REFRAMED** — 既知 TASK-445 / DEC-28。新規ではなく既知ポインタ。
### cmd（コマンド）（7件）

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
#### CMD-01: stage-import の破壊的 DELETE が clinic_id で制約されていない（既知 BUG-430） — CRITICAL
- 区分: **既知** → BUG-430
- 規約: `.claude/refs/backend-application-invariants.md:11` 「- clinic-scoped data のすべての read/write/delete は、認証済み `clinic_id` で制約する。」
- 対象: `backend/cmd/stage-import/tables.go:76`、`backend/cmd/stage-import/tables.go:79`、`backend/cmd/stage-import/tables.go:103`、`backend/cmd/stage-import/tables.go:118`、`backend/cmd/stage-import/tables.go:66`、`backend/cmd/stage-import/apply.go:176`、`backend/cmd/stage-import/apply.go:187`
- 内容: deleteScope / cascadeChildScope / petScopedExists / billingsOldDBScope が生成する DELETE 述語は owner_id >= 300000 という id レンジ由来の provenance 軸のみで、clinic_id 述語を一切含まない。apply.go:176-203 はこの述語で owners/pets/medical_records/billings/billing_items/exams/exam_results と 4 つの CASCADE 子テーブルを削除する。
- 修正: BUG-430 の裁定どおり cmd/stage-import 一式（Makefile 5 target・docker-compose.stage-import.yml・scripts/verify-stage-import.sh 含む）を退役させる。退役までの暫定措置を採る場合は resolveTargetRefs が解決した clinicID を全 deleteScope 述語へ AND 結合する。

- 再実測(2026-07-28): **ALREADY-FIXED** — backend/cmd/stage-import ディレクトリが HEAD から消失（ls backend/cmd に stage-import 無し）。BUG-430 退役済み。破壊的 unscoped DELETE の live path は存在しない。
- round3-review(2026-07-28): **WITHDRAWN** — ALREADY-FIXED。`backend/cmd/stage-import` は HEAD から消失（BUG-430 退役）。破壊的 unscoped DELETE の live path 無し。
#### CMD-02: 全院バッチを起動する /_internal/scheduled-jobs が未認証の root engine に登録されている — MEDIUM
- 区分: 新規 ／ 検証で severity 引き下げ
- 規約: `.claude/rules/go-gin-backend-guidelines.md:134` 「- public route と authenticated/authorized route の境界を route 登録時に明示する。」
- 対象: `backend/cmd/api/base_routes.go:26`、`backend/cmd/api/base_routes.go:14`、`backend/cmd/api/main.go:198`、`backend/cmd/api/main.go:203`、`backend/cmd/api/batch_scheduler.go:43`、`backend/cmd/api/batch_scheduler.go:59`、`backend/internal/scheduler/handler.go:118`、`backend/internal/scheduler/handler.go:122`、`backend/internal/scheduler/handler.go:220`、`docker-compose.yml:40`
- 内容: registerBaseRoutes は scheduler route を *gin.Engine へ直接登録する。main.go:198-204 が engine に載せる middleware は Recovery/SecurityHeaders/RequestID/CORS/RequestLogging/SanitizeNullBytes のみで認証は無い。handler.run(:122-201) も credential を検証せず、RunRequest.validate(:220-245) が要求する scheduler 名・run_id 形式・fence_token>0 は全て repo 内の公開定数から導出できる。防御は worker/index.ts:236 の edge path filter 一枚だけで、docker-compose.yml:40 は 8080 を直接公開しているため Worker を経由しない環境では誰でも RunDeliveryTriggerBatchAllClinicsAt（全院への LINE 配信）を起動できる。
- 修正: scheduler route を共有 secret（例 X-Scheduler-Token の定数時間比較）を要求する専用 RouterGroup 配下へ移し、Worker 側 containerFetch にも同ヘッダを付与する。最低限、base_routes.go に「edge でのみ保護される特権 route」であることを明示し、非 release mode では登録しないか loopback 限定にする。
- 検証時の補正: 「防御は worker/index.ts:236 の edge path filter 一枚だけ」は事実だが、その filter は encoding/traversal bypass まで封じた文書化済み・単体test済みの設計上の境界（scheduler-ops.ts:83-119 / scheduler-ops.test.ts:578-596）であり、production/staging に非-Worker ingress は存在しない（infra/ 実測）。したがって failure scenario は「ローカル docker-compose で 8080 を直接叩ける開発環境」に限定して記述すべき。残る指摘は「特権 route であることが route 登録側に表明されていない／Go 側に二重防御が無い」に絞られる。行番号の微差: handler.go の :118/:122/:220 は実測 :119-121/:124/:221、docker-compose.yml:40 は :41（ports mapping）。いずれも指示された構造の内側に着地しており evidence 実在性は満たす。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **DOWNGRADED** → severity **LOW** — Go router は未認証だが production は Worker edge 隔離前提。defense-in-depth/文書化ギャップ。
#### CMD-03: coverage-ratchet の失敗が二重に無音化され gate が構造的に機能しない — MEDIUM
- 区分: 新規 ／ 横断パターン: X-04 ／ 検証で severity 引き下げ
- 規約: `.claude/refs/error-handling.md:9` 「- error を無視しない。処理できる境界まで返すか、明示的に回復する。」
- 対象: `backend/cmd/coverage-ratchet/main.go:91`、`backend/cmd/coverage-ratchet/main.go:93`、`backend/cmd/coverage-ratchet/main.go:105`、`backend/cmd/coverage-ratchet/main.go:58`、`backend/cmd/coverage-ratchet/main.go:62`、`backend/cmd/coverage-ratchet/main.go:118`、`.github/workflows/ci.yml:200`
- 内容: readBaseline は os.ReadFile の err を破棄して 0 を返し（:92-95）、解析不能行も無視して 0 を返す（:96-105）。0 は evaluateRatchet の warn-only 分岐（:118-124）へ落ち、ファイルが存在するのに読めない場合でも "baseline not recorded" という事実と異なる文言を印字して exit 0 になる。さらに唯一の呼び出し元 ci.yml:200 は出力を tee へパイプするが、当該 workflow には defaults: も shell: も無く GitHub Actions 既定の `bash -e {0}`（pipefail 無効）で実行されるため、os.Exit(1)(:62-64) 自体が tee の終了コードに飲まれる。
- 修正: readBaseline を (float64, error) にして「欠落＝未記録」と「読取/解析失敗」を分離し、後者は os.Exit(2) にする。あわせて ci.yml の当該 step に shell: bash（pipefail 有効）を付けるか tee を削って exit code を露出させる。
- 検証時の補正: ①「ファイルが存在するのに読めない場合でも事実と異なる文言で exit 0」は正しいが、この 0 返却は main.go:89-90 のコメントで宣言された仕様であり、「無音化」ではなく粒度の粗い明示的回復として記述すべき。②gate を構造的に無効化している決定的要因は ci.yml:201 の tee による exit code 吸収であり、これは error-handling.md:9 の適用対象外。同一欠陥は frontend ratchet（ci.yml:297 `node scripts/coverage-ratchet.mjs ... | tee -a`）にも存在するため、是正は backend cmd 単体ではなく workflow 横断で行う必要がある。③evidence の ci.yml:200 は実測 :201（:200 は `if [ -f coverage-summary.txt ]; then`）。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **DOWNGRADED** → severity **LOW** — coverage-ratchet は CI tooling。runtime safety ではない。workflow exit-code 契約。
#### CMD-04: pet 死亡/復活の cross-domain write が composition root で map[string]any 経由に降格している — MEDIUM
- 区分: WIP-adjacent（清算後に再検証）
- 規約: `backend/CODING_RULES.md:35` 「- cross-domain呼び出しはbusiness intentを表すconsumer側の最小interfaceと型安全なDIを基本とする。owner外へ`map[string]any`等の任意field更新APIを公開せず、複数domainにまたがるwriteはownerとtransaction境界を明示する。」
- 対象: `backend/cmd/api/lstep_adapters.go:110`、`backend/cmd/api/lstep_adapters.go:112`、`backend/cmd/api/lstep_adapters.go:118`、`backend/cmd/api/composition_runtime.go:164`、`backend/internal/lstep/lstep_lifecycle_deps.go:30`、`backend/internal/pet/repository.go:68`、`backend/internal/pet/repository.go:75`、`backend/internal/pet/repository.go:309`、`backend/internal/pet/repository.go:322`、`backend/internal/pet/repository.go:387`、`backend/internal/pet/repository.go:481`
- 内容: pet.CompleteRepository(:75-78) は既に LifecycleWriter 経由で RecordDeath/ClearDeath(:68-69,:481-515) を型安全に公開しており lstep.PetLifecycleWriter を構造的に充足するのに、petLifecycleWriterAdapter(:110-122) は汎用の Update(ctx, clinicID, petID, map[string]any) を通す。pet 側はこれを支えるため legacyLifecycleTransition(:322-345) という第二の CAS 実装を保持しており、同じ遷移規則と conflict message が 2 箇所に複製されている。lstep_lifecycle_deps.go:30-31 と pet/repository.go:73-74 はいずれも「composition が cut over するまで」という条件付き carve-out と明記しているが、cut over は行われていない。
- 修正: petLifecycleWriterAdapter を削除し composition_runtime.go:164 で repositories.ownerPet.Pet を PetLifecycle へ直接渡す。consumer が無くなった時点で pet/repository.go の legacyLifecycleTransition / updateLegacyLifecycleFieldsWithDB を撤去する。internal/pet/repository.go は並行 writer 保護対象のため清算後に再検証する。
- 検証時の補正: evidence の :387 は updatePetFieldsWithDB であり、本文が指す updateLegacyLifecycleFieldsWithDB は :347。また map[string]any 汎用 API が渡っているのは composition root（cmd/api）に対してのみで、lstep 本体には型付き PetLifecycleWriter しか渡っていない。したがって規約 :35 のうち抵触するのは「owner外へ map[string]any 等の任意field更新APIを公開せず」の中核ではなく、前段の「型安全なDIを基本とする」からの逸脱＋業務ルール複製（DRY）として記述するのが正確。backend/internal/pet/repository.go は並行 writer 保護対象のため**清算後に再検証**すること。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **REFRAMED** — typed LifecycleWriter 自体は意図的 narrowing。欠陥は cutover 未完の dual map path 債務。
#### CMD-05: /uploads の StaticFS が無条件・未認証で登録され非 release mode では PHI を配信する — MEDIUM
- 区分: 新規
- 規約: `.claude/rules/go-gin-backend-guidelines.md:134` 「- public route と authenticated/authorized route の境界を route 登録時に明示する。」
- 対象: `backend/cmd/api/base_routes.go:25`、`backend/cmd/api/base_routes.go:10`、`backend/cmd/api/main.go:151`、`backend/cmd/api/main.go:121`、`backend/internal/medicalrecord/medical_record_image_request.go:153`、`backend/internal/config/config.go:207`
- 内容: base_routes.go:25 は認証 middleware を持たない engine に /app/uploads を丸ごと mount する。STORAGE_TYPE が s3 以外のとき main.go:151 の LocalUploader がカルテ画像を同じ木の medical-records/<medicalRecordID>/... へ書くため、認証も clinic scope も経ずに PHI が取得できる。config.go:207-208 が release mode で s3 を強制するので production では実害が無く、鍵は 16byte crypto random（medical_record_image_request.go:145-151、lstep 共有ファイルも shared_file_service.go:215 で同様）だが、capability URL は認証の代替にはならず（退職・院移動・画像削除後も URL は失効しない）、route 登録側にも public である旨の表明が無い。
- 修正: StaticFS の登録を cfg.StorageType != "s3" のときだけに限定し、公開 route であることを base_routes.go に明記する。恒久策としては署名付き URL または認証済み配信 handler 経由に寄せる。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **DOWNGRADED** → severity **LOW** — release は STORAGE_TYPE=s3 強制。local StaticFS は dev 露出。prod PHI StaticFS ではない。
#### CMD-06: csv-import-failure-rehearsal が error chain を破棄し原因を復元不能にしている — MEDIUM
- 区分: 新規
- 規約: `.claude/refs/error-handling.md:17` 「- 文脈を追加する場合は `fmt.Errorf("...: %w", err)` を使う。」
- 対象: `backend/cmd/csv-import-failure-rehearsal/main.go:146`、`backend/cmd/csv-import-failure-rehearsal/main.go:166`、`backend/cmd/csv-import-failure-rehearsal/main.go:170`、`backend/cmd/csv-import-failure-rehearsal/main.go:210`、`backend/cmd/csv-import-failure-rehearsal/main.go:305`、`backend/cmd/csv-import/main.go:143`
- 内容: 上記 5 箇所は err を受けながら固定文字列の fmt.Errorf を返し、%w も %v も付けずに元の error を捨てる。同じ csvimport 基盤を使う cmd/csv-import/main.go:143 以降は一貫して %w で wrap しており、rehearsal だけが逆になっている。rehearsal は G4 rollback の production 適格性を証明する道具であるため、DB 到達不能なのか TLS 設定不正なのか receipt 書込失敗なのかが stderr の 1 行から判別できず、失敗時の切り分けが不能になる。
- 修正: 各所を fmt.Errorf("...: %w", err) へ揃える。receipt に機微値が乗らないことは既に validateDisposableTarget(:276-293) で担保済みなので、chain 保持による情報漏えい増分は無い。
- 検証時の補正: ①「rehearsal だけが逆になっている」は不正確。同ファイル :176 は `fmt.Errorf("synthetic G4 rehearsal failed: %w", err)` で domain error を正しく wrap している。実際のパターンは「domain error は wrap、infra error（timezone / openTarget / Ping / encode / pgxpool.ParseConfig）は chain 破棄」であり、この記述の方が欠陥として正確かつ強い。②「DB 到達不能なのか TLS 設定不正なのか receipt 書込失敗なのかが stderr の 1 行から判別できず」は誤り。固定文字列は段階ごとに異なる（"open disposable target database" / "ping disposable target database" / "write aggregate execution receipt"）ため**どの段階で落ちたかは判別可能**であり、失われるのは根本原因（driver error・OS error・DNS/認証失敗の別）のみ。③proposed_fix の「receipt に機微値が乗らないことは validateDisposableTarget(:276-293) で担保済み」は非論理。同関数は host=="db"・sslmode=="disable"・DB名の disposable パターン・確認文字列を検証するだけで error chain の内容には一切関与しない。%w 化の妥当性自体はこの誤った根拠に依存しない。

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — csv-import-failure-rehearsal が %w 欠落。error-handling.md:17。
#### CMD-07: lstep-migrate の進捗台帳書込失敗が warn ログのみで呼び出し元へ返らない — MEDIUM
- 区分: 新規 ／ 横断パターン: X-04
- 規約: `.claude/refs/error-handling.md:9` 「- error を無視しない。処理できる境界まで返すか、明示的に回復する。」
- 対象: `backend/cmd/lstep-migrate/migrator.go:292`、`backend/cmd/lstep-migrate/migrator.go:296`、`backend/cmd/lstep-migrate/migrator.go:277`、`backend/cmd/lstep-migrate/migrator.go:177`、`backend/cmd/lstep-migrate/migrator.go:257`、`backend/cmd/lstep-migrate/main.go:37`
- 内容: updateProgress は戻り値を持たず、FirstOrCreate の err を logger.Warn するだけで握り潰す(:292-297)。呼び出し元 processOwner(:177,:257) は成功として続行し、ProgressRecord も CSV レポートも「success」のまま出力される。結果として lstep_migration_progress は実際の同期結果と乖離し、--resume-from による再開判断の根拠が壊れる。
- 修正: updateProgress を error 返却へ変え、processOwner が ProgressRecord.Status を partial/failed へ落として ErrorMessage に台帳書込失敗を含めるか、Run 側で集約して非ゼロ終了させる。
- 検証時の補正: 「--resume-from による再開判断の根拠が壊れる」は間接的で誤読を招く。実測では --resume-from（main.go:37 / :126 → migrator.go:27）は filterOwners（:160）で `o.ID < m.cfg.ResumeFrom` の owner ID 比較を行うだけで lstep_migration_progress を一切読まない。壊れるのは flag の動作ではなく、**台帳テーブルを見て再開点を決める運用者の判断**である。また「ProgressRecord も CSV レポートも success のまま出力される」は正しいが、CSV はタグ同期の実結果を正しく反映しており誤りではない。実害は「CSV は success、DB 側 lstep_migration_progress の行は pending/旧値のまま」という2系統の乖離であり、その乖離を検知する手段が Warn ログしか無いこと。

---

- 再実測(2026-07-28): **位置照合OK** — 対象 `file:line` は現 HEAD で全件現存（機械照合）。内容の再審査は未実施（Tier 3 契約）。
- round3-review(2026-07-28): **UPHELD** — lstep-migrate 進捗台帳失敗が warn のみ。error-handling.md:9。
## 既知として扱う項目（本書では詳細を再記述しない）

上記一覧のうち「区分: 既知」の項目は、正本が別文書にある。本書の記述は**面の追加証跡**であり、新規起票してはならない。round 1 起票分 6 件に加え、2026-07-28 に反証を通過した 4 件を台帳へ起票済み（下表の後半）。

| 所見 ID | 正本 | 本書が追加した事実 |
|---|---|---|
| CMD-01 | [`#BUG-430`](3-session-agent.html#BUG-430)（DEC-24 で削除退役を裁定済み） | `tables.go:76,79,103,118` の 4 scope 生成関数すべてが clinic 述語を欠く |
| MRB-03 | [`#BUG-437`](3-session-agent.html#BUG-437) ／ [`#SEC-SWEEP-02`](3-session-agent.html#SEC-SWEEP-02) | `hospitalization_repository.go:86,87` は SEC-SWEEP-02 の残 5 面リストに未収載の**面追加** |
| RSV-08 | [`#SEC-SWEEP-02`](3-session-agent.html#SEC-SWEEP-02) | `shift_entry_breaks` が grandchild lint の registry にも未登録 |
| TRM-08 | [`#SEC-SWEEP-02`](3-session-agent.html#SEC-SWEEP-02) | 中間 junction `appointment_trimming_options`（`001_init.sql:1380-1387`）に clinic_id 列がない |
| MDL-06 | [`#TASK-445`](3-session-agent.html#TASK-445)（DEC-28 で案 b を裁定済み） | model 側（`model/accounting.go:163`）の ClinicID 追加も同 unit に含める必要がある |
| MRC-12 | `phase2.html:195`（Transactor + LockByIDForUpdate の正規パターン統一） | `inquiry_repository.go:42-46` のコメントが Transactor 不在を設計上の制約として自認している |
| BIL-01 | [`#BUG-463`](3-session-agent.html#BUG-463)（2026-07-28 起票・CRITICAL） | 105 件中唯一の coordinator 独立実測。round3 で REFRAMED（締め後監査と status ガードの2命題へ分割） |
| G2C-04 | [`#BUG-464`](3-session-agent.html#BUG-464)（2026-07-28 起票） | round2 の敵対的却下パスを UPHELD で通過した 3 件のひとつ |
| G2P-02 | [`#BUG-465`](3-session-agent.html#BUG-465)（2026-07-28 起票） | 同上。X-01（commit 済み write の tx 外 re-fetch）の構成員 |
| G2P-03 | [`#BUG-466`](3-session-agent.html#BUG-466)（2026-07-28 起票） | 同上。`001_init.sql` に quantity の CHECK 制約が無いことを追加確認 |

---

## 規約側の疑義（32 論点）

**規約が曖昧・自己矛盾・実装と乖離している箇所**。実装を直す前にこちらを確定しないと、修正が振動する。特に 4（fail-closed 監査の指定台帳がない）と 7（金額・税率の範囲 contract がない）は、CRITICAL / HIGH 項目の適用範囲そのものを左右する。

1. **`.claude/CLAUDE.md:7`** — 「Type Safety First: Prohibit `any` in both Go and TypeScript」は無条件禁止として書かれているが、本unitの41ファイル中13ファイル・計39箇所で `map[string]any` が GORM Updates の列名→値マップとして使われている（hospitalization_service.go:90, examination_service.go:69, diagnosis_service.go:65,83, hospitalization_plan_service.go:25, exam_type_service.go:21, および各 repository の Update(fields map[string]any) シグネチャ）。.claude/refs/go-language.md:47「`any` や reflection は型で表現できない境界に限定する。」は境界での使用を許容しており、両者の強度が食い違う（台帳の疑義-3）。結果、この規約は systematically violated か systematically waived のどちらかだが、waive を記録した ADR が repo 内に存在しないため、レビュー時に指摘すべきか否かを規約から決定できない。
   - 提案: .claude/CLAUDE.md:7 に「ORM の部分更新マップ等、型で表現できない persistence 境界を除く」旨の例外句を入れて go-language.md:47 と整合させるか、逆に許容境界を列挙した ADR を起こして参照させる。どちらも採らないなら、`map[string]any` を型付き update struct へ寄せる移行計画を記録し、それまでは指摘対象外と明記する。
   - 影響所見: `CMD-04`。着手ブロック: **いいえ**（着手後でも可）。

2. **`.claude/refs/backend-application-invariants.md:11 + backend/migrations/CLAUDE.md:14`** — invariants:11 は「clinic-scoped data のすべての read/write/delete は、認証済み clinic_id で制約する」と要求し、migrations/CLAUDE.md:14 は「新テーブルにクリニック間分離が必要な場合は clinic_id NOT NULL を付ける」と述べるが、どちらも『必要な場合』の判定基準を定義していない。実装は同じ深さの子テーブルで二分している: clinic_id を非正規化保持する側（payment_splits / treatment_plans / daily_records / care_logs / medicine_dose_params / checkup_field_results）と、親 join で相関する側（payments / care_plan_items / staff_notes / estimate_items / treatments / clinical_plans / inquiries / exam_results / medical_record_images）。レビュー時に、ある model が clinic_id を持たないことを欠陥として起票すべきか、意図的な相関設計として不問にすべきかが規約から導けない（本監査でも Payment を既知 TASK-445 として扱う一方、care_plan_items/staff_notes を不問にする判断根拠が規約に無かった）。
   - 提案: backend-application-invariants.md に判定基準を1条追加する。例:『親が単一 clinic に属することが FK で保証され、かつ当該テーブルへの全 read/write が親への clinic 相関を伴う lint（grandchild_parent_clinic_correlation_lint_test.go）の対象に登録されている場合に限り、clinic_id の非正規化を省略してよい。それ以外は clinic_id NOT NULL + 複合FK を必須とする』。併せて migrations/CLAUDE.md:14 からこの条文へポインタを張る。
   - 影響所見: `MDL-06`。着手ブロック: **はい**（`DEC-29`）。

3. **`.claude/refs/backend-application-invariants.md:11 / backend/migrations/CLAUDE.md:14`** — 「clinic-scoped data」と「クリニック間分離が必要な場合」のいずれにも判定基準が無い。clinic_id 列を持たないが院単位 RBAC ルート（ResourceHospitalSettings 等）から mutate されるテーブル（lstep_auto_managed_prefixes 等）が clinic-scoped に該当するかを規約から導けず、LS-A-04 の severity と是正方向が規約だけでは決まらない。
   - 提案: invariants に「clinic-scoped data の定義: clinic 単位 permission で mutate 可能な行、または clinic 固有の業務判断に影響する行」を明示し、それに該当するなら (a) clinic_id を持つ (b) platform-admin 専権にする のいずれかを取れ、という選択肢付きの判定基準を書く。
   - 影響所見: `LSA-04`。着手ブロック: **はい**（`DEC-30`）。

4. **`.claude/refs/backend-application-invariants.md:15`** — 「preload にも同じ scope を適用する」と無条件に書かれているが、実装側の機械gate backend/internal/lintscan/preload_clinic_scope_lint_test.go:129-146 は `staffExemptAssoc`（Doctor / CreatedByStaff 等8association）を「leak は staff NAME only, low severity」として恒久免除している。この免除は規約正本にも ADR にも記載がなく、go-gin-backend-guidelines.md:233 が要求する「ADR または application invariant として根拠・適用範囲・検証方法を記録する」を満たしていない。結果として監査者は appointment_admin_repository.go:44 の `Preload("Doctor", "deleted_at IS NULL")` を違反と読むか免除と読むかを規約から決定できない（reservation_repository.go:17-21 は同じ association に対し assignment-EXISTS の厳格述語を課しており、同一package内でも扱いが割れている）。
   - 提案: staff系 association の clinic-scope 免除を invariants 側へ明記し、免除の範囲（staffs は staff_clinic_assignments が正本のため単純 clinic_id scope が誤りになる旨）と、reservation_repository.go が採る assignment-EXISTS 述語を推奨形として記録する。
   - 影響所見: なし。着手ブロック: **いいえ**（着手後でも可）。

5. **`.claude/refs/backend-application-invariants.md:31`** — 「destructive または irreversible な操作には、権限、対象 scope、監査、recovery 方針を持たせる」の「対象 scope」が、clinic_id を持たないグローバル共有マスタ（animal_species 等）に対して何を意味するかを定義していない。個々のクリニックの権限グループで全クリニック共有行を破壊・改名できる構造が規約違反か否かを、この文からは判定できない。同ファイル :11-16 の tenant isolation 節はグローバルマスタを明示的に対象外にも対象にもしていない。
   - 提案: グローバル共有マスタを1カテゴリとして節を起こし、(a) 誰が write owner か (b) 変更に監査を必須とするか (c) 使用中チェックの範囲は全テナントか、を明記する。少なくとも「clinic_id を持たないマスタへの write は必ず audit_logs へ記録する」を追加すれば PO-07 の判定は機械化できる。
   - 影響所見: `POC-07`。着手ブロック: **はい**（`DEC-31`）。

6. **`.claude/refs/backend-application-invariants.md:32`** — 「`appointments`とそのlifecycleは`reservation`、`staffs`と`shift_entries`は`staff`がwrite ownerであり、BE9-2E-0で収束済みの境界を`appointment_write_owner_lint_test.go`の自動gateで維持する」と書かれているが、当該gateは appointments しか検出面に持たない（CODING_RULES.md:102 の検出面列挙もすべて appointment 前提）。実装は reservation_staff_repository.go:46-50 の staffsWriter / reservation_schedule_repository.go:40-48 の shiftWriter という手作りの delegate で staffs / shift_entries の write owner を守っており、機械gateは存在しない。規約が実際より強い保証を主張している。
   - 提案: 「appointments は自動gateで維持、staffs/shift_entries は consumer-side writer port による設計上の分離のみ（gate未整備）」と現状を書き分けるか、write-owner gate を staffs / shift_entries にも拡張して記述を実態に合わせる。
   - 影響所見: なし。着手ブロック: **いいえ**（着手後でも可）。

7. **`.claude/refs/backend-application-invariants.md:35`** — 「意図的なsaga/best-effort処理は、補償、再試行、監査、部分失敗contractを持たせる。」は best-effort 部分が失敗したときに同期 HTTP 応答が何を返してよいかを規定していない。「部分失敗contract」がレスポンス上の表現を含むのか、コード内コメントで宣言すれば足りるのかが決定不能である。実装は両極に割れている: medical_record_subrecords.go は主訴消失時も medical_record_handler.go:119 で無条件 201 を返し監査も残さない一方、medical_record_auto_create.go:198-235 は同種の best-effort 失敗に専用 audit action を必ず記録する。lab_result_import_service.go:184-189 は終端遷移失敗をログのみにしてジョブを非終端のまま 201 を返す。
   - 提案: 「宣言した best-effort 構成要素が失敗した場合、同期 HTTP 応答で無条件の成功を返してはならない。応答に部分失敗を示す field を含めるか、失敗を専用 audit action として記録するかのいずれかを必須とする」を明文で追加する。auditReservationDraftCleanupFailure を参照実装として名指しし、best-effort 経路の一覧を invariant 側で管理する。
   - 影響所見: `MRC-04`。着手ブロック: **はい**（`DEC-32`）。

8. **`.claude/refs/backend-application-invariants.md:35（意図的 best-effort の要件）`** — 「意図的なsaga/best-effort処理は、補償、再試行、監査、部分失敗contractを持たせる」は4要素を列挙するが、これらが全て必須なのか、いずれか1つで足りるのかを規定していない。lstep には slog によるログのみの best-effort 経路が多数あり（lstep_lifecycle_service.go:149,157,160,163,259,298 等）、ログが「監査」に該当するかも不明。判定が査読者の裁量に委ねられ、LS-B-02 のような不可逆な記録喪失と、単なる再試行可能なタグ同期失敗が同じ扱いになる。
   - 提案: 4要素の必須/選択を明示し、最低ラインとして「後から失敗を検出し再実行できる持続的な記録（slog 以外）」を必須と定める。加えて『best-effort の副作用が、再試行に必要な状態自体を破壊してはならない』という不可逆性の禁止条項を追加する。
   - 影響所見: `LSB-02`。着手ブロック: **いいえ**（着手後でも可）。

9. **`.claude/refs/backend-application-invariants.md:37`** — 「締め後の会計編集はこの対象とする」の『会計編集』が定義されていない。billings 行の直接 UPDATE だけを指すのか、billing_items の変更（結果として billings.subtotal/tax_total/total_amount を書き換える）も含むのかが規約から決定できない。実装はこの曖昧さのまま分岐しており、accounting PATCH には post-close ゲートがあるのに billing-items PATCH/POST には無い（BL-01）。cash_register_closes は締め時点のスナップショットを永続化する（backend/internal/billing/cash_register_service.go:359-377、backend/migrations/001_init.sql:3659「既存の締め記録（cash_register_closes のスナップショット）は再計算しない」）一方、月次・日次レポートは status=completed 行から都度再集計する（backend/internal/billing/accounting_repository_reports_monthly.go:146, accounting_repository_reports_daily.go:29）ため、対象範囲の曖昧さは実際に「締めスナップショットと再集計値の乖離」という具体的損害へ直結する。
   - 提案: 『会計編集』を「billings 行または billings の金額列を導出する子行（billing_items / payments / payment_splits / billing_refunds）に対する write」と列挙で定義し、対象テーブル一覧を invariant 本文に明記する。合わせて lintscan に「これらのテーブルへ write する service は IsDateClosed 判定と post-close 監査を経ているか」の inventory gate を新設できる形にする。
   - 影響所見: `BIL-01`。着手ブロック: **はい**（`DEC-33`）。

10. **`.claude/refs/error-handling.md:18`** — 「error の種類は errors.Is / errors.As で判定し、message 文字列比較をしない」に例外規定が無いが、pgx ドライバの encode 失敗は型付き error も sentinel も公開しておらず、errors.Is/As では判定できない。backend/internal/apperrors/errors.go:172-179 はこの制約により BUG-138 対応として意図的に文字列一致を使っている。規約どおりに書くと分類自体が不可能になるため、現行実装は「規約違反」ではなく「規約が到達不能」の状態にある。
   - 提案: 「upstream が型/sentinel を公開しない外部ドライバ境界に限り文字列一致を許容する。その場合は ADR または application invariant に対象・根拠・文言変更を検知する回帰テストを記録する」旨の例外条項を追記する（guidelines:233 が求める記録形式に合わせる）。
   - 影響所見: `INF-04`。着手ブロック: **はい**（`DEC-34`）。

11. **`.claude/refs/error-handling.md:9 と backend/CODING_RULES.md:36`** — error-handling.md:9 は error の無視を無条件に禁じるが、CODING_RULES.md:36 は「best-effortを選ぶ場合は部分成功contract、再試行、補償、監査を明示する」として best-effort を許容する。lstep には slog.Warn + continue の best-effort が数十箇所あり、そのどれが :9 違反でどれが :36 の許容範囲かを分ける基準が無い（実際 tag cache upsert / audit warn / status 更新 warn が同じ書式で混在している）。
   - 提案: 「best-effort と宣言してよい条件」を列挙する（例: 失敗しても business fact が不整合にならない、次回同期で必ず収束する、失敗が集計値に載る、の3点すべてを満たす場合のみ）。満たさない箇所は :9 の対象として扱う、と明記する。
   - 影響所見: `LSA-03`, `LSA-10`, `LSA-11`, `LSA-12`, `LSB-02`, `LSB-04`, `MRC-04`, `RSV-09`, `INF-06`, `CMD-03`, `CMD-07`, `G2B-01`。着手ブロック: **はい**（`DEC-35`）。

12. **`.claude/refs/go-gin-backend-review.md:68`** — 「soft-deleteやhistory semanticsをschema/ADRに合わせ、暗黙条件に依存していないか。」は GORM の gorm.DeletedAt を持つ model に対して字義どおりには充足不能である。GORM は Query/Update/Delete に deleted_at IS NULL を常に暗黙適用するため、明示条件を書いても暗黙条件は消えない。本 unit 内でも扱いが割れている: staff_clinic_assignment_repository.go:40-41 と :71 は「GORM SoftDelete スコープにより deleted_at IS NULL フィルタは自動適用される。明示的な条件追加は不要。」と暗黙依存を明文で肯定する一方、staff_repository.go:243 / :155 / occupation_repository.go:98 は同じ状況で明示的に deleted_at IS NULL を書いている。同一 package 内で規約解釈が2通り並存している。
   - 提案: :68 に判定基準を1文追加する。例:「ORM が暗黙適用する soft-delete scope に依存する場合は、依存していることを当該 method の doc comment に明記すれば充足とみなす。raw SQL・JOIN・EXISTS サブクエリなど ORM scope が届かない箇所は明示必須。」これで staff_clinic_assignment_repository.go の形が適合、EXISTS 内の未記述が違反、と機械的に切り分けられる。
   - 影響所見: なし。着手ブロック: **いいえ**（着手後でも可）。

13. **`.claude/refs/go-gin-backend-review.md:69`** — 「riskのあるquery/transaction/isolationを実DBのintegration testで確認しているか」の『実DB』が、(a) 本物の PostgreSQL インスタンスであること、(b) 本番と同一の schema（migrations/*.sql 適用結果）であること、のどちらを指すのか規定していない。本プロジェクトの実DBハーネス internal/testdb は (a) は満たすが (b) は満たさない — schema は GORM AutoMigrate（testdb.go:99, :446）と手書き DDL リテラル2箇所（SharedTestSchemaEnumTypes :269-334 / EnsureClinicSettingsTable :156-195）から構築され、migrations/001_init.sql は一切適用されない。(a) の解釈なら現行ハーネスは適合、(b) の解釈なら構造上永久に不適合となり、同じコードに対する監査結論が読み手によって反転する。3-session-agent.html:169 はこれを「既知の落とし穴・是正task未起票」として記録済みで、規約側が解釈を確定していないことが未起票の一因になっている。
   - 提案: go-gin-backend-review.md:69（または backend/CLAUDE.md の Required safety 節）に、どちらの解釈が支配するかを明記する。(a) を採るなら『schema 生成経路が本番 migration と異なる場合は、両者の差分を機械的に突合する gate を持つこと』を条件として併記し、ENUM parity gate と同型の突合を clinic_settings 等の手書き DDL にも義務付ける。(b) を採るなら是正 task を起票し、それまでの猶予条件を明示する。
   - 影響所見: `MDL-05`, `MRA-01`, `POC-06`, `G2P-03`。着手ブロック: **いいえ**（着手後でも可）。

14. **`.claude/rules/go-gin-backend-guidelines.md:134`** — 「public route と authenticated/authorized route の境界を route 登録時に明示する」は Go の route 登録面に閉じた要求だが、本 system の /_internal 境界は Go 側ではなく backend/worker/index.ts:236-238（Cloudflare Worker が 404 で遮断）で強制されている。edge で閉じられた route を Go 側で「明示」する手段が規約に定義されていないため、設計として正しい構成が規約上は非適合に読める。逆に「Worker が守るから Go 側は不要」という運用も、規約からは正当化も禁止もできない。
   - 提案: 「境界の強制が application 外（edge/WAF/private network）にある場合は、route 登録直上に強制主体のファイルパスを明記し、当該経路を列挙した inventory と、迂回時の挙動を検証する test を持たせる」条項を追加する。あるいは『edge 依存を認めず Go 側でも必ず fail-closed にする』と明記して曖昧さを消す。
   - 影響所見: `CMD-02`。着手ブロック: **はい**（`DEC-36`）。

15. **`.claude/rules/go-gin-backend-guidelines.md:151`** — 規約本文（「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」）は妥当だが、台帳が付した検査法「string で `max=` 指定が無いものを抽出」がこの package では運用不能な偽陽性量を生む。担当範囲だけでも vaccination_request.go:67-72 の Supplemental / Lot1 / Lot2 / Lot3 / Lot4 / Remarks、vital_request.go:18 の Notes、treatment_request.go:17-19 の Content / Memo / AdminRoute、treatment_plan_request.go:5 の Memo が全て無条件ヒットする。これらは自由記述の臨床テキストで、DB 側も length 制約を持たないため、max= の欠如自体は欠陥ではない。この検査法をそのまま適用すると、真に危険な列挙値の未検証（本 unit の MR-D-05）が偽陽性の中に埋もれる。
   - 提案: 検査法から一律の string max= 走査を外し、(a) DB 側に length 制約または enum 型がある列に bind される field、(b) 外部システムへ転送される field に限定する。列挙値は独立の検査軸として切り出し「DB enum 型に bind されるのに binding oneof も service validator も無い field」を機械判定条件にする。
   - 影響所見: `AUS-03`, `LSA-13`, `LSA-14`, `LSB-06`, `MRA-03`, `MRB-06`, `MRC-08`, `MRD-04`, `POC-13`, `POC-17`, `RSV-04`, `TRM-03`, `TRM-05`, `G2A-05`, `G2C-02`, `G2P-03`。着手ブロック: **いいえ**（着手後でも可）。

16. **`.claude/rules/go-gin-backend-guidelines.md:166`** — PostgreSQL エラーコード → HTTP status のマッピングを『どの層が所有するか』を定める規約が無く、実装が二重管理になっている。backend/internal/apperrors/errors.go:183-192 は 23503/23505/22003/22P02 の4コードを sentinel へ写像し、backend/internal/httpapi/response_pg.go:28-42 は 23514 を加えた5コードを HTTP メッセージへ写像したうえで default を 400 に落とす。両者は集合が食い違い（FromGORM に 23514 が無い／classifyPgError に 23505→409 の概念が無い）、同じ制約違反が経路により 400 と 409 に分かれ得る。RULE-063(guidelines:166) と RULE-054(CODING_RULES.md:53) はいずれも status の安定性を要求するが、所有層を指定していない。
   - 提案: 「DB ドライバ固有コードの分類は単一の層（例: apperrors）が所有し、HTTP 境界はその結果のみを写像する。境界側に第二の分類表を置かない」を規約へ追加し、既存2表の統合を移行タスクとして起票する。
   - 影響所見: `INF-04`, `MRB-06`, `TRM-05`, `G2P-03`。着手ブロック: **いいえ**（着手後でも可）。

17. **`.claude/rules/go-gin-backend-guidelines.md:179`** — 「request/body/upload size ... を制限する」の適用単位（グローバル middleware か handler ごとか）が未定義であり、実装は6パッケージが handler ごとの http.MaxBytesReader、3パッケージ（pet/owner/clinic）が無制限という分裂状態にある。どちらの形も条文を満たすと読めるため、未適用パッケージの存在を機械的に欠陥と判定できない。
   - 提案: 「JSON body の上限はグローバル middleware で一律に設定し、より小さい上限が必要な endpoint だけ handler で上書きする」のように既定の適用単位を規定する。そのうえで lintscan に「ShouldBind* を呼ぶ handler が上限 middleware 配下にあること」の gate を置ける。
   - 影響所見: `INF-02`, `POC-12`, `TRM-03`。着手ブロック: **はい**（`DEC-37`）。

18. **`.claude/rules/go-gin-backend-guidelines.md:194`** — §11 は表題どおり「Production server lifecycle」だけを対象にしており、:196-201 の 6 則は全て HTTP server 前提である。しかし backend/cmd 配下には破壊的 DB 権限を持つ 7 本の one-shot CLI（migrate / stage-import / csv-import / csv-import-failure-rehearsal / seed-export / lstep-migrate / coverage-ratchet）が存在し、これらの終了コード契約・signal 処理・resource close・error chain 方針・確認フラグ設計を規定する条文が台帳のどこにも無い。結果として同種 binary 間の方針差（cmd/migrate/main.go:73 は DB を close するが cmd/lstep-migrate/main.go:76-80 は close しない、cmd/csv-import は %w で wrap するが cmd/csv-import-failure-rehearsal は chain を捨てる）が規約違反として判定できず、CM-06 のような drift が野放しになる。
   - 提案: 「operational CLI / batch binary」の節を新設し、(a) 失敗は必ず非ゼロ終了かつ呼び出し元がそれを観測できること、(b) error は %w で chain 保持、(c) 破壊的操作は明示フラグ + host guard、(d) 保有 resource の close 順序、の 4 則を Go/Gin 公式とは別区分（project decision）として明記する。
   - 影響所見: `CMD-03`, `CMD-06`, `CMD-07`, `TRM-07`。着手ブロック: **いいえ**（着手後でも可）。

19. **`.claude/rules/go-gin-backend-guidelines.md:227 / :233 vs ~/.claude/rules/ecc/common/coding-style.md:39`** — guidelines:227 は「- package/file/directory の固定サイズ」を Go/Gin 公式要件から明示除外し、:233 は「これらが必要なら、公式由来の規約と混ぜず、ADR または application invariant として根拠・適用範囲・検証方法を記録する。」と条件を付す。一方 ecc/common/coding-style.md:39「- 200-400 lines typical, 800 max」は無条件に課され、.claude/CLAUDE.md:129 はこれを「正本」と呼ぶ。しかし :233 が要求する ADR / invariant 記録は repo 内に存在しない（backend-application-invariants.md 全46行にサイズ規定なし）。本監査では MR-A-06（http_session.go 820行）がこれに直撃し、違反として発行してよいか、発行する場合の severity を規約から導出できなかった。
   - 提案: backend 側で明示的に決める。(A) 800行閾値を backend にも適用するなら ADR を1本起こし適用範囲（production .go のみか test も含むか）と検証方法（lintscan gate の有無）を書く。(B) 適用しないなら .claude/CLAUDE.md:129 の「正本」記述に「サイズ/immutability/coverage 閾値は backend には適用しない」旨の除外を明記する。どちらでも監査側の判断は決定可能になる。
   - 影響所見: なし。着手ブロック: **いいえ**（着手後でも可）。

20. **`.claude/rules/go-gin-backend-guidelines.md:92`** — 「export は最小限にし、公開識別子には GoDoc で読めるコメントを付ける。」は識別子の GoDoc のみを扱い、package 宣言コメントの正しさを規定していない。このため backend/internal/trimming/trimming_service.go:1 の「// Package service provides business logic implementations for Trimming entity.」が `package trimming` に付いている（同 package 内の trimming_course_repository.go:1 には正しい package doc が別途あり、1 package に2つの package コメントが存在し片方が誤り）状態を、規約違反として指摘する根拠が無い。BE9 の layer package → domain package 移設で同型の残存が他 package にも生じうる。
   - 提案: 「package コメントは 1 package につき1つとし、`Package <実際のpackage名>` で始める」を命名節に追記する（go vet / revive の package-comments で機械検出可能）。
   - 影響所見: `MRA-04`。着手ブロック: **いいえ**（着手後でも可）。

21. **`backend/CLAUDE.md:25`** — 「自動化には停止、失敗通知、監査、手動fallback、idempotencyまたは明示的retry policyを設ける」の「自動化」の外延が未定義であり、本unitの ①予約確定/キャンセル通知の fire-and-forget goroutine（appointment_notification_service.go:106,157）②予約確定に伴うカルテ自動作成（reservation_handler.go:164）③LINE顧客のowner自動紐付け（liff_service_reservations.go:182）が対象かどうかを規約から決定できない。いずれも停止手段・失敗通知・監査を持たずlogのみで完結している。
   - 提案: 「自動化」を「人間の明示操作を起点としない処理、および起点はあるが結果が呼び出し元へ返らない副作用（fire-and-forget / best-effort）」と定義し、後者に最低限「失敗の構造化log＋観測可能なメトリクスまたはaudit記録」を必須とする、と範囲を明文化する。
   - 影響所見: `MRC-04`, `POC-06`, `RSV-09`。着手ブロック: **いいえ**（着手後でも可）。

22. **`backend/CLAUDE.md:25 / backend/CODING_RULES.md:41`** — 「自動化には停止、失敗通知、監査、手動fallback、idempotencyまたは明示的retry policyを設ける」の「失敗通知」に受理条件が無い。lstep バッチ群は slog.ErrorContext のみで満たしていると読めるが、DEC-25 で「CI失敗通知は実装せず運用でカバー」と裁定された経緯から、slog が十分なのか人間に届く経路が必須なのかがレビュー時に判定不能。LS-A-03 も「失敗通知が無い」ではなく「error を握り潰した」でしか指摘できなかった。
   - 提案: 「失敗通知」の最低受理条件を明記する（例: 失敗が audit_logs に行として残り、かつ BatchRunResult.Failed に計上されること。人間への push 通知は必須としない）。満たさない実装を検出する gate の所在も併記する。
   - 影響所見: `LSA-03`, `LSA-11`, `LSA-12`, `RSV-09`, `CMD-03`, `CMD-07`, `G2B-01`。着手ブロック: **いいえ**（着手後でも可）。

23. **`backend/CLAUDE.md:25（RULE-037 自動化要件）`** — 「停止」の粒度が定義されていない。lstep 配信バッチには clinic 単位の停止スイッチ（IsSyncEnabled）が存在し規約を満たしているように見えるが、飼主単位の停止（lstep_opt_out）は data model と DDL コメント（001_init.sql:316「true = すべてのタグ付与をスキップ」）で宣言されているにもかかわらず、規約が要求する「停止」に含まれるかが読み取れない。結果として LS-B-01 のような per-subject 停止の取りこぼしが規約適合レビューをすり抜ける。
   - 提案: 「停止」を clinic 単位（システム停止）と data subject 単位（個別オプトアウト）に分け、後者については「停止フラグを持つ列は、その列の contract を消費する全経路を列挙し gate を持つこと」を要求する。lstep_opt_out / delivery_excluded のように意味が重なる二列については、どちらがどの経路の正本かを ADR で確定する（現状 owner ドメインの service_delivery.go:31-35 は両方を同時に立てるが、lstep 側の RecordLstepOptOut は片方しか立てない非対称がある）。
   - 影響所見: `LSA-02`, `LSB-01`。着手ブロック: **はい**（`DEC-38`）。

24. **`backend/CLAUDE.md:44`** — 「P1–P18 は廃止された project 固有 checklist であり、レビュー基準に使わない。」と宣言している一方、production code のコメントには退役番号が live 参照として残っている（backend/internal/manualarticle/repository.go:19「P4 例外」、service.go:114「P1: Delete 前の存在確認」、handler.go:179「SEC-602: P5 RequirePermission(view) 付与」）。番号の定義が退役済みのため、これらのコメントは読み手が根拠を辿れない。規約とコードの乖離である。
   - 提案: 退役番号の掃引対象を doc だけでなく production code コメントにも広げ、各コメントを現行正本（例: P4 → backend-application-invariants.md:11 の clinic-scope 例外）への参照へ置換する。
   - 影響所見: なし。着手ブロック: **いいえ**（着手後でも可）。

25. **`backend/CLAUDE.md:44 と .claude/refs/go-gin-backend-review.md:89`** — 「P1–P18 を判定基準に使わない」という規約は、レビュー側の判定基準としての使用を禁じるだけで、production code 内のコメントが退役番号を規範的根拠（MANDATORY 等）として引用し続けている状態を欠陥と見なすか否かを定めていない。さらに実コードには CPM 閾値の見出しラベルとしての「P1」「P2」「P9」（lstep_settings_service.go:30,35,49 / lstep_tag_sync_service.go:108）が同じ字句で共存しており、機械的な grep 検査では両者を判別できない。RULE-133 の検査法が `--include='*.md'` に限定されている点も .go コメントの扱いを未定義のまま残している。
   - 提案: 退役番号の扱いを (a) 検査対象は live doc と production code コメントの双方、(b) 例外として業務ドメイン由来の P 記号（CPM フェーズ等）は対象外、と明文化する。判別可能にするため、退役 checklist 参照は `P<n>:` 形式、ドメインラベルは別記法（例: `CPM-P<n>`）へ正規化する規約を追加する。
   - 影響所見: なし。着手ブロック: **いいえ**（着手後でも可）。

26. **`backend/CODING_RULES.md:35 / backend/CODING_RULES.md:40`** — CODING_RULES.md:40 は逐語「compatibility facadeは薄いdelegate/type aliasだけを許可し、business ruleやpersistence実装を複製しない。consumer移行後の削除条件を持たせる。」と定めるが、削除条件の記述形式・所有者・失効検知手段を規定していない。実装側の carve-out（internal/lstep/lstep_lifecycle_deps.go:30-31「generic field maps stay at the composition adapter until the owner/pet domains migrate in BE9-2E」/ internal/pet/repository.go:73-74「until central composition cuts over to the typed lifecycle capability」）は条件をコメントで宣言しているだけで、BE9 移行が 2026-07-24 に code complete となった後も誰も失効を検知できず CM-04 が残置された。
   - 提案: 削除条件を「コメント」ではなく機械検査可能な形（内部 package の deprecated 宣言 + internal/lintscan の allowlist entry + 撤去 issue ID）で持たせることを CODING_RULES.md:40 に追記し、allowlist に載った facade が consumer 0 になったら gate が赤くなる規定にする。
   - 影響所見: `CMD-04`。着手ブロック: **いいえ**（着手後でも可）。

27. **`backend/CODING_RULES.md:38`** — ロック要件が『FOR UPDATE / FOR SHARE / pg_advisory_xact_lock を正しさの根拠にする operation』と『request由来のclinic-scoped FK の再検証』の2方向しか規定しておらず、逆方向すなわち『依存件数を数えてから親を削除する』read-then-write を直列化せよという要件がどこにも無い。このため internal/inventory/inventory_service.go:176-186 と internal/inventory/merchandise_item_service.go:214-226 が、依存件数 0 の判定と soft delete を transaction にも行ロックにも入れずに実行していても、規約上どの条項にも接地できない（同一 review unit 内の internal/billing/estimate_service.go:403-425 は tx + CountItems + DeleteIfNotLocked の原子述語で正しく実装しており、実装側には既に正解パターンが存在する）。
   - 提案: :38 に「依存存在チェックを根拠に親行を削除・無効化する operation は、チェックと write を同一 transaction に入れ、削除述語自体へ依存不在条件を含める（条件付き原子削除）」の1文を追加し、正解実装例として estimate_service.go の DeleteIfNotLocked を参照させる。
   - 影響所見: なし。着手ブロック: **いいえ**（着手後でも可）。

28. **`backend/CODING_RULES.md:77`** — repository method が ambient transaction に参加すべき条件を定めた規約が存在しない。`backend/CODING_RULES.md:38` と `.claude/refs/backend-application-invariants.md:38` は FOR UPDATE / FOR SHARE / advisory lock に依存する operation しか扱わず、通常の Update/Create は対象外である。結果として persistence.UpdateScopedByID の呼び出し28箇所が `r.db` 版と `persistence.DBOrTx(ctx, r.db)` 版に分裂し（例: trimming_course_repository.go:62 は DBOrTx、trimming_option_repository.go:62 と trimming_course_type_repository.go:69 は r.db）、どちらが正しいか規約から判定できない。
   - 提案: 「clinic-scoped write を行う repository method は既定で ambient transaction に参加する（DBOrTx を通す）。参加させない場合は doc comment で理由を明示する」を規約化し、lintscan の DBOrTx inventory の判定基準として参照させる。
   - 影響所見: `TRM-09`。着手ブロック: **はい**（`DEC-39`）。

29. **`backend/CODING_RULES.md:78`** — 「write後の再取得が失敗し得る場合はcommit前の同じtransaction内で行うか」は、その責務を repository と service のどちらが負うかを定めていない。実装は owner/pet が repository（UpdateAndFind 内の Transaction）、clinic が service（Transactor.WithTx）と分かれており、どちらを正本とすべきかを規約から導けないため、PO-02 の修正方針が実装者ごとに割れる。
   - 提案: 「update+reload を1つの操作として提供する repository method を write owner 側に置く」等の配置基準を1文追加するか、少なくとも「同一 business graph の write と reload の transaction 境界は write owner package が所有する」と明記する。
   - 影響所見: `BIL-03`, `MRA-02`, `MRC-01`, `MRD-02`, `POC-02`, `RSV-03`, `TRM-01`, `LSB-03`, `G2A-01`, `G2B-02`, `G2P-02`。着手ブロック: **いいえ**（着手後でも可）。

30. **`backend/CODING_RULES.md:79`** — 「schema constraint と application validation の両方を使う。」は、DB 側が既に制約を持つ場合に application validation の欠落を違反とみなすのか、defense-in-depth として許容するのかを定めていない。本 unit の bw_unit（PostgreSQL enum body_weight_unit が実効的に拒否するが binding tag が無い）はこの空白に落ちるため、TR-A-05 の severity が規約から一意に導けない。
   - 提案: 「DB 制約のみで application validation が無い場合も違反とする（理由: 境界で 4xx を返せず、深部での失敗に退化するため）」のように、片側のみの場合の扱いを明記する。
   - 影響所見: `MRB-06`, `TRM-05`, `G2P-03`。着手ブロック: **いいえ**（着手後でも可）。

31. **`backend/migrations/CLAUDE.md:15`** — 規約本文は「業務データは `deleted_at TIMESTAMPTZ` を追加する」と現在形の設計不変条件として書かれているが、台帳側の検査法は「新規 `CREATE TABLE` を migration diff から抽出」という diff 前提で定義されている。2026-07-27 の migration 完全統合（0efcca770 / edbb162f7、backend/migrations/CLAUDE.md:44 ・ :51 に記録）で直下 DDL が 001_init.sql 単一ファイルになった結果、新規テーブルを含む diff が存在せず、既存 schema に対して本ルールを機械適用する経路が消えた。実例として care_plan_items（001_init.sql:1757-1779）は業務データでありながら deleted_at を持たないが（MR-A-01）、diff 起点の検査法では検出不能のため本ルールを根拠にできなかった。同型の diff 依存が RULE-091（schema constraint と application validation の併用）・RULE-121（clinic_id NOT NULL）・RULE-124（clinic_id 複合 index）にもある。
   - 提案: これらの検査法を「migration diff に対する検査」から「001_init.sql の現行 CREATE TABLE 全数に対する不変条件」へ書き換える。統合後は差分が存在しないため、既に起票済みの TASK-447（CREATE TABLE 実数 ⇔ ERD 宣言値の照合 gate）と同じ走査基盤の上で、業務データテーブルの deleted_at 有無・clinic_id NOT NULL・clinic_id 先頭複合 index を現行 schema 全数に対して判定する形へ移す。適用除外（中間テーブル・lookup・意図的 hard-delete）は allowlist として明示登録する。
   - 影響所見: `LSA-04`, `MRA-01`, `MDL-06`, `POC-06`, `G2P-03`。着手ブロック: **いいえ**（着手後でも可）。

32. **`~/.claude/rules/ecc/common/coding-style.md:64 (Constants: `UPPER_SNAKE_CASE`)`** — 言語スコープ宣言のないグローバル規約が Go の慣用（MixedCaps）と正面から衝突する。本 unit の const は全て camelCase / PascalCase（例 auth/http_types.go:15 AccessTokenCookieName、auth/token_service.go:21 accessTokenTTL、staff/staff_service_builders.go:4 colStaffName、staff/http_binding.go:17 staffJSONBodyMaxBytes）。字義適用すると担当2 package の全定数が違反になるが、.claude/refs/naming-conventions.md:7-16 の Go 命名節には定数命名規定が無く、意図した結果とは考えられない。台帳側でも疑義-2 として RULE 化が見送られており、実装との乖離が固定化している。
   - 提案: coding-style.md の Naming Conventions 節冒頭に適用言語を明記する（当該節は camelCase / PascalCase / use prefix hooks を並置しており TypeScript/React 前提と読める）。例:「本節は TypeScript/JavaScript に適用する。Go の命名は .claude/refs/naming-conventions.md および Go 公式（MixedCaps）に従う。」
   - 影響所見: なし。着手ブロック: **いいえ**（着手後でも可）。

---

## 監査の限界（次回の改善点・着手前に必読）

### round 1（2026-07-27）自己申告 — 履歴

1. **lstep の 45 ファイルが全文未読**（grep 全数走査で代替）。次回は lstep を 3 分割する。
2. **性能系（N+1・unbounded query）を実質的に監査できていない**。
3. **並行性シナリオは全て机上導出**。
4. **テストコードを監査対象から除外**。
5. **coordinator による独立実測は 105 件中 1 件（BIL-01）のみ**。HIGH 以上は着手前再実測必須。
6. **統合フェーズのエージェントが API stall で失敗**。X-01〜X-10 の畳み込みは敵対的検証未了。
7. **監査パイプライン自体の ID 衝突事故**（AU 単位が MR-A-* を誤発行）。
8. **severity は着手優先度そのものではない**。

### round 2（2026-07-28）で閉じた盲点 / 残る限界

| # | round 1 盲点 | round 2 結果 |
|---|---|---|
| 1 | lstep 未読 45（実測未引用 89） | 未引用 89 を 3 分割全文読了。新規 G2A/G2B/G2C 系を追加。 |
| 2 | 性能系未監査 | 静的走査で G2F-01〜11 を発行（N+1 / unbounded）。クエリプラン実測は未実施。 |
| 3 | 並行性は机上 | 変更なし（read-only 制約）。X-05 は構造再確認のみ。 |
| 4 | テスト品質未監査 | 系統 grep + サンプリング。AUS-04 未修正確認 + G2T-02 twin。恒真 `assert.True(t,true)` は 1 件のみ。 |
| 5 | HIGH 以上未実測 | CRITICAL+HIGH 35 件を実装者再実測（CONFIRMED/LINE-DRIFT/ALREADY-FIXED）。 |
| 6 | X 畳み込み未検証 | 敵対的再検査実施。needs revision（重複・over-fold・missed fold-in）。詳細は下記。 |
| 7 | ID 衝突 | round 2 新規は `G2*` prefix のみ。`uniq -d` 空を検証。 |
| 8 | severity≠優先度 | 継承。 |

**round 2 でも残る限界**: (a) 並行性の実 DB 再現なし (b) EXPLAIN なし (c) MEDIUM/LOW の内容再審査は位置照合のみ (d) working tree dirty 領域は並行編集中のため、着手時に再々実測を推奨。

---

## round 2（2026-07-28）抜け漏れ監査結果

### 測定基準

| 項目 | 値 |
|---|---|
| HEAD | `c4ce786e0a52968a4fcdeaaffe07a409501ca80b` |
| 初版監査日 | 2026-07-27 |
| round 2 実施日 | 2026-07-28 |
| orchestration | parallel subagent fan-out（T1×3 / T2-lstep×3 / T2-pkg / T2-test / T2-perf / T2-fold）→ main writer 単独追記 → REVIEW |
| dirty backend | billing 4 files + lintscan 1 test（baseline 相対で追加 dirty 無しを確認してから commit） |

### Tier 1: CRITICAL/HIGH 35 件 再実測サマリー

| 分類 | 件数 | ID |
|---|---:|---|
| CONFIRMED | 27 | AUS-01, BIL-03, INF-02, LSA-01, LSA-02, LSA-04, LSA-05, LSA-06, LSA-07, LSB-01, LSB-02, LSB-03, LSB-04, MRA-01, MRA-02, MRC-01, MRC-02, MRC-04, MRC-05, MRD-01, MRD-02, MRD-03, MRD-04, POC-02, RSV-02, TRM-01, TRM-03 |
| LINE-DRIFT | 6 | BIL-01, LSA-03, MRB-02, MRB-03, POC-03, TRM-04 |
| ALREADY-FIXED | 2 | **CMD-01**, **INF-01** |
| REJECTED | 0 | — |

**内訳検算**: 27 + 6 + 2 + 0 = 35

**ALREADY-FIXED 詳細**:

| ID | 根拠 |
|---|---|
| CMD-01 | `backend/cmd/stage-import` が消失。stage-import 由来の clinic 非限定 DELETE は live path に存在しない。 |
| INF-01 | `ResolveErrorResponse` が未知 PgError を HTTP 500 に分類（allowlist のみ 400）。 |

**重複メモ**: LSA-02 ≡ LSB-01（同一 checkExclusion gap）。修正作業は LSB-01 に一本化してよい（両方 CONFIRMED のまま残す）。

### Tier 3: MEDIUM 56 + LOW 14 位置照合

| 結果 | 件数 |
|---|---:|
| 位置照合OK | **70** |
| 位置DRIFT | 0 |
| 位置消失 | 0 |

全 70 ID の cited `file:line` は現 HEAD でファイル存在かつ行番号がファイル長内。内容再審査は Tier 3 契約外。

### Tier 2 coverage ledger（positive evidence 必須）

| 面 | 調べた数 / 総数 | 走査方法 | 新規所見数 | clean 判定根拠 |
|---|---|---|---:|---|
| lstep 未引用 batch A | 31 / 31 | 全文読了（uncited non-test alpha 先頭 1/3） | 7 (G2A) | clean: composition*/doc/DTO/response/handler 配線、checkup_sync request 境界検証、line_link tx+audit、analytics year_month 検証。defect は line_customer/aggregation/preview のみ |
| lstep 未引用 batch B | 31 / 31 | 全文読了（uncited non-test middle 1/3） | 5 (G2B) | clean: batch_noshow WithTx+CAS+audit、CSV stream preflight、delivery monitor handler/repo clinic join、health 付与 path fail-closed。defect は remove 非対称/suppression/monitor enum 等 |
| lstep 未引用 batch C | 30 / 30 | 全文読了（uncited non-test 末尾 1/3） | 8 (G2C) | clean: LTV SyncLTVTopPercent fail-closed batch cache、trigger priority Upsert clinic 強制、tag summary page bounds、formula sanitize。defect は LIKE/export/remove-after-fail 等 |
| lstep 合計 | **89 / 89** | 3 分割並列全文 | 20 | 暗黙 truncation なし |
| apicontract | 1/1 | 全文 | 0 | OpenAPI drift test package のみ |
| authjwt | 1/1 | 全文 | 0 | claims DTO のみ |
| dbconn | 2/2 | 全文 | 0 | pool/DSN、secret 非ログ |
| logger | 1/1 | 全文 | 0 | slog wrapper |
| seedbundle | 1/1 | 全文 | 0 | manifest schema |
| sharedkernel | 10/10 | 全文 | 0 | tax は Validate のみ・金額計算ドリフト無し（MDL-01 は model） |
| textsearch | 1/1 | 全文 | 0 | LIKE escape 実装本体 |
| timeutil | 2/2 | 全文 | 0 | format/weekday |
| infra | 18/18 | 全文 | 0 new | 既存 INF-03 のみ（新規無し） |
| persistence | 5/5 | 全文 | 0 | DBOrTx/WithTx 健全 |
| inventory | 12/12 | 全文 | 3 (G2P) | clinic scope は一貫、削除 TOCTOU 等を新規発行 |
| テスト品質 | corpus grep + ~35 精読 | tautology/isolation/mock/Skip 系統 | 3 実質 (G2T; AUS-04 は既知) | isolation スイートは概ね本物。silent Skip は high-risk 無し |
| N+1/unbounded | 10 domain + 74 FindAll シグネチャ | grep + 精読 | 11 (G2F) | lexical n1_lint は clean。interprocedural/batch が主戦場 |
| X-01〜X-10 畳み込み | 全構成員 | 所見本文突合 | 0 新規 ID | 畳み込み品質 **needs revision**（下記） |

### 横断パターン畳み込みの妥当性（監査の限界 6 への回答）

**総合判定: needs revision**（パターン定義自体は有用。構成員の重複・over-fold・missed fold-in あり）。

| 優先修正 | 内容 |
|---|---|
| Dedup | LSA-13 ↔ MDL-04（同一 trigger_type）。LSA-08 ↔ LSB-05（同一疎通 raw error） |
| Move | POC-14 は X-02 ではなく **X-09**（空 PATCH ガードの兄弟欠落） |
| Split | TRM-03 は X-07（body size）+ X-02（string max）の両面 |
| Add to X-05 | POC-06（phone 一意性 check-then-write）、LSA-15（trigger check-then-Create） |
| Add to X-04 | LSB-02（RemoveTag 失敗 swallow） |
| Narrative | サマリー L43 が LSA-02/LSB-01 を X-04 扱いしているのは誤り（error swallow ではなく business evidence 欠落） |

### 新規所見一覧（round 2）

既存 14 prefix と衝突しない `G2*` 体系。区分はすべて **新規**（既知台帳へのポインタのみのものは「区分: 既知」）。

#### G2A-01: line_customer LinkOwner が commit 後 re-fetch 失敗で成功を反転する — HIGH
- 区分: 新規 ／ 横断: X-01
- 規約: `backend/CODING_RULES.md:78` 「write後の再取得が失敗し得る場合はcommit前の同じtransaction内で行うか、commit済みの成功を後段read errorで失敗へ反転させないcontractにする。」
- 対象: `backend/internal/lstep/line_customer_service.go:46`、`backend/internal/lstep/line_customer_service.go:58`
- 内容: `UpdateOwnerLink` 成功後に tx 外 `FindByID` が失敗すると error 応答になるが DB はリンク済み。
- 修正: WithTx 内 reload、または更新成功後は構築 DTO を返す。

- round2-review(2026-07-28): **DOWNGRADED** (HIGH→LOW) — 反証: write 後 Find 失敗は稀でリンク状態は正しい・再送は冪等（`line_customer_service.go:46-58`）。CODING_RULES:78 は適用可だが HIGH 過大。
#### G2A-02: line_customer の owner FK 検証が write と同一 tx にない — MEDIUM
- 区分: 新規 ／ 横断: X-05
- 規約: `backend/CODING_RULES.md:38` 「request由来のclinic-scoped FKは永続化と同じtransactionで再検証し、並行master変更で判定が無効になる場合は対象行をcommitまで共有ロックする。」
- 対象: `backend/internal/lstep/line_customer_service.go:36`、`backend/internal/lstep/line_customer_service.go:48`
- 内容: owner 所属確認と UpdateOwnerLink が別 transaction。
- 修正: 同一 WithTx 内で owner FOR SHARE → UpdateOwnerLink。

- round2-review(2026-07-28): **WITHDRAWN** — 反証: CODING_RULES:38 は「並行master変更で判定が無効になる場合」の条件付き。owner の clinic 所属は実質固定で pre-write clinic-scoped Find 済み（`line_customer_service.go:40-48`）。tenant hole として過適用。
#### G2A-03: LINE顧客↔飼主リンク変更に監査が無い — MEDIUM
- 区分: 新規 ／ 横断: X-03
- 規約: `.claude/refs/backend-application-invariants.md:31` 「destructive または irreversible な操作には、権限、対象 scope、監査、recovery 方針を持たせる。」
- 対象: `backend/internal/lstep/line_customer_service.go:35`、`backend/internal/lstep/line_customer_service.go:59`
- 内容: 患者識別境界の変更なのに audit 呼び出しが無い。
- 修正: write と同一 tx で fail-closed 監査。

- round2-review(2026-07-28): **WITHDRAWN** — 反証: invariants:31 は destructive/irreversible 対象。Link/unlink は可逆（`ownerID==nil` で解除）で slog.Info 済み。fail-closed 監査必須の clinical/financial ではない。
#### G2A-04: UpdateAdditionalFields が RowsAffected 0 を成功扱いする — MEDIUM
- 区分: 新規
- 規約: `.claude/refs/error-handling.md:9` 「error を無視しない。処理できる境界まで返すか、明示的に回復する。」
- 対象: `backend/internal/lstep/line_customer_repository.go:86`、`backend/internal/lstep/line_customer_repository.go:94`
- 内容: clinic 不一致でも nil 成功。同 file の UpdateOwnerLink は RowsAffected 検査あり。
- 修正: RowsAffected==0 → NotFound。

- round2-review(2026-07-28): **WITHDRAWN** — 反証: result.Error は処理済み。RowsAffected 0 は意図的 best-effort（`line_customer_repository_test.go:229`「0件更新でもエラーにはならない」）。error-handling.md:9 の error 無視ではない。
#### G2A-05: aggregation クエリの enum/日付が境界未検証 — MEDIUM
- 区分: 新規 ／ 横断: X-02
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/lstep/aggregation_request.go:79`、`backend/internal/lstep/aggregation_request.go:90`
- 内容: from/to・period_preset・amount_basis 等が未検証のまま 200 誤集計または 500 になり得る。
- 修正: allowlist + 日付 parse を 400 化。

- round2-review(2026-07-28): **DOWNGRADED** (MEDIUM→LOW) — 反証: 未知 enum は safe default/空結果（`ltv_repository.go:280,378-401`）。SQL injection なし。from/to の 4xx 化のみ残課題。
#### G2A-06: checkup sync プレビューがタグキャッシュ DB エラーを空 tags に置換する — MEDIUM
- 区分: 新規 ／ 横断: X-04
- 規約: `.claude/refs/error-handling.md:9` 「error を無視しない。処理できる境界まで返すか、明示的に回復する。」
- 対象: `backend/internal/lstep/checkup_sync_service_preview.go:49`、`backend/internal/lstep/checkup_sync_service_preview.go:54`
- 内容: FindByOwners 失敗を空 map に置換し 200 継続。「タグ無し」と障害が区別不能。
- 修正: fail-closed または `tags_degraded` 明示 flag。

- round2-review(2026-07-28): **WITHDRAWN** — 反証: `checkup_sync_service_preview.go:49-54` が G7-2 固定の non-fatal と slog.Error を明示。preview 専用の意図的 degraded であり error 無視ではない。
#### G2A-07: line send の purpose が列挙未検証 — LOW
- 区分: 新規 ／ 横断: X-02
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/lstep/line_send_request.go:15`、`backend/internal/lstep/line_send_request.go:37`
- 内容: 未知 purpose は silent no-op（タグ非付与）。
- 修正: allowlist + 400。

- round2-review(2026-07-28): **WITHDRAWN** — 反証: purpose は動的 master（`lstep_send_purpose_tag_prefixes`）照合。未一致はタグ非付与の正しい contract（`line_send_service.go:116-129`）。固定 oneof は設定破壊。
#### G2B-01: 健診・予防タグの Remove 失敗を nil で握り潰し成功件数に入れる — HIGH
- 区分: 新規 ／ 横断: X-04
- 規約: `.claude/refs/error-handling.md:9` 「error を無視しない。処理できる境界まで返すか、明示的に回復する。」
- 対象: `backend/internal/lstep/lstep_health_tag_sync_prevention.go:110`、`backend/internal/lstep/lstep_health_tag_sync_prevention.go:112`、`backend/internal/lstep/lstep_health_tag_sync_food.go:61`、`backend/internal/lstep/lstep_health_tag_sync_vaccine.go:69`、`backend/internal/lstep/lstep_health_tag_sync_checkup.go:81`、`backend/internal/lstep/lstep_health_tag_sync_batch.go:82`
- 内容: desired=true の付与失敗は return err、desired=false の解除失敗は apiFailed に落として return nil。Lステップ上に PREV_* 残タグ → 配信トリガー誤発火し得る。バッチ成功件数にも入る。
- 修正: 解除失敗も err 伝播し BatchRunResult.Failed / audit に計上。

- round2-review(2026-07-28): **DOWNGRADED** (HIGH→MEDIUM) — 反証: Remove 失敗は notifyAPIFailure 経由で計上・テスト固定（`lstep_health_tag_sync_prevention_test.go:85-94`）。外部 API best-effort の意図的非対称。残リスクは PREV_* 残留→配信候補で MEDIUM 相当。
#### G2B-02: lstep_settings Upsert が commit 後 Find で成功を反転し得る — MEDIUM
- 区分: 新規 ／ 横断: X-01
- 規約: `backend/CODING_RULES.md:78` 「write後の再取得が失敗し得る場合はcommit前の同じtransaction内で行うか、commit済みの成功を後段read errorで失敗へ反転させないcontractにする。」
- 対象: `backend/internal/lstep/lstep_sync_settings_repository.go:40`、`backend/internal/lstep/lstep_sync_settings_repository.go:51`
- 内容: Upsert 成功後の別 query Find 失敗で 5xx。flag は永続化済み。
- 修正: 同一 tx 内 reload または構築済み struct 返却。

- round2-review(2026-07-28): **DOWNGRADED** (MEDIUM→LOW) — 反証: Upsert 済み flags は durable。post-Find は hydration のみ（`lstep_sync_settings_repository.go:40-51`）。CODING_RULES:78 は成立するが運用リスクは LOW。
#### G2B-03: 優先度 demote がログ上書きのみで既付与タグを取り消さない — MEDIUM
- 区分: 新規
- 規約: `backend/CODING_RULES.md:36` 「best-effortを選ぶ場合は部分成功contract、再試行、補償、監査を明示する。」（旧 CLAUDE.md:25 は主題不一致のため差し替え）
- 対象: `backend/internal/lstep/lstep_delivery_trigger_suppression.go:55`、`backend/internal/lstep/lstep_delivery_trigger_suppression.go:68`、`backend/internal/lstep/lstep_batch_delivery.go:28`
- 内容: suppressed_by_priority=true にしても Lステップ上の低優先タグは残る。固定実行順と priority 設定の不一致で多重配信。
- 修正: demote 時 Remove+cache 削除、または実行順を priority 昇順固定、または多重タグ許容を contract 明示。

- round2-review(2026-07-28): **REFRAMED** — 反証: demote はログ `suppressed_by_priority` 仕様（`docs/spec/screens/34-lstep-delivery-monitor.md:48`）。CLAUDE.md:25 の自動化停止リスト違反ではなく、固定実行順 vs priority 排他の製品設計ギャップ。
#### G2B-04: delivery-monitor の trigger_type/status が列挙未検証 — MEDIUM
- 区分: 新規 ／ 横断: X-02
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/lstep/lstep_delivery_monitor_request.go:31`、`backend/internal/lstep/lstep_delivery_monitor_request.go:52`
- 内容: typo で 200+空集計。
- 修正: AllTriggerTypes / status allowlist。

- round2-review(2026-07-28): **WITHDRAWN** — 反証: 任意 query filter の typo→空結果は REST 標準。injection/auth なし。列挙必須は write 経路の話。
#### G2B-05: lifecycle request の reason に長さ上限が無い — LOW
- 区分: 新規 ／ 横断: X-02
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/lstep/lstep_lifecycle_request.go:28`、`backend/internal/lstep/lstep_lifecycle_request.go:34`
- 内容: reason が無 max。DB は text のため破壊はしにくいが境界未制限。
- 修正: binding max（例 100–500）。

- round2-review(2026-07-28): **WITHDRAWN** — 反証: reason は optional text（`001_init.sql:318`）。破壊リスク実質なし。スタイル一貫性のみ。
#### G2C-01: tag summary の owner 名 LIKE が EscapeLike 無し — MEDIUM
- 区分: 新規
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/lstep/lstep_tag_cache_repository.go:178`、`backend/internal/lstep/lstep_tag_cache_repository.go:180`
- 内容: `%`/`_` が wildcard 化。clinic scope 内だが意図より広い一覧。owner/pet は EscapeLike 使用。
- 修正: textsearch.EscapeLike + ESCAPE。

- round2-review(2026-07-28): **DOWNGRADED** (MEDIUM→LOW) — 反証: parameterized LIKE で injection ではない。clinic 内検索意味論の EscapeLike 不整合。guidelines:151 適用は PARTIAL。
#### G2C-02: tag code mapping の code_type/species/age が境界未検証 — MEDIUM
- 区分: 新規 ／ 横断: X-02
- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 対象: `backend/internal/lstep/lstep_tag_code_mapping_request.go:3`、`backend/internal/lstep/lstep_tag_code_mapping_service.go:69`
- 内容: required 以外の enum/range/length 無し。
- 修正: oneof/min/max/dive。

- round2-review(2026-07-28): **DOWNGRADED** (MEDIUM→LOW) — 反証: 未知 code_type は automation no-op。staff 設定 CRUD。境界検証ギャップは残るが MEDIUM 過大。
#### G2C-03: care/pet/visit タグ同期が remove 失敗後も add して nil 成功 — MEDIUM
- 区分: 新規 ／ 横断: X-04
- 規約: `.claude/refs/error-handling.md:9` 「error を無視しない。処理できる境界まで返すか、明示的に回復する。」／ `backend/CLAUDE.md:25` 「自動化には停止、失敗通知、監査、手動fallback、idempotencyまたは明示的retry policyを設ける。」
- 対象: `backend/internal/lstep/lstep_tag_sync_care.go:61`、`backend/internal/lstep/lstep_tag_sync_pet.go:59`、`backend/internal/lstep/lstep_tag_sync_visit_cpm.go:82`、`backend/internal/lstep/lstep_tag_sync_visit_dormant.go:53`
- 内容: remove fail → apiFailed=true continue → 新規 add → return nil。旧+新タグ併存。LTV 同期は fail-closed 先例あり。
- 修正: remove 失敗時は return err、または add 前に補償。

- round2-review(2026-07-28): **WITHDRAWN** — 反証: remove fail は notifyAPIFailure + apiFailed、成功カウンタ非リセット。テストが「apiFailed のまま nil」を固定。外部 LSTEP の意図的 best-effort。error-handling.md:9 を「batch 必須 fail-closed」へ過適用。
#### G2C-04: tag owners CSV が 5000 件で無信号 truncate し、stream 後に JSON error し得る — MEDIUM
- 区分: **既知** → BUG-464
- 規約: `.claude/rules/go-gin-backend-guidelines.md:154` 「response は一度だけ書き、return して handler を終了する。」／ `.claude/refs/error-handling.md:29` 「response を書いた後に別の error response を重ねない。」（旧 guidelines:151「集合サイズ」は入力検証条項のため差し替え）
- 対象: `backend/internal/lstep/lstep_tag_summary_service.go:131`、`backend/internal/lstep/lstep_tag_summary_handler.go:43`
- 内容: total 破棄で truncate 無信号。headers 後の RespondError で body 混在。
- 修正: total>5000 は fail-closed または paginate。stream 後は JSON error しない。

- round2-review(2026-07-28): **UPHELD** — 反証失敗: headers 後の RespondError は `guidelines:154` / `error-handling.md:29` 違反がコード上明確（`lstep_tag_summary_handler.go:43-59`）。truncate 無信号は副次。


#### G2C-05: owner tags API が line_user_id 全文を返し masked フィールドが未配線 — LOW
- 区分: 新規
- 規約: `.claude/refs/backend-application-invariants.md:24` 「response、export、event、audit log は最小限の field だけを含める。」
- 対象: `backend/internal/lstep/lstep_tag_response.go:10`、`backend/internal/lstep/lstep_tag_summary_response.go:23`
- 内容: 全文 line_user_id 露出。masked は DTO のみで未セット。
- 修正: mask または削除。

- round2-review(2026-07-28): **WITHDRAWN** — 反証: staff owners:view の owner 詳細に line_user_id は運用上意図的。masked 未配線は別 DTO の死んだ optional field。
#### G2C-06: SharedFileRepository.Create が clinic 引数を取らない — LOW
- 区分: 新規
- 規約: `.claude/refs/backend-application-invariants.md:11` 「clinic-scoped data のすべての read/write/delete は、認証済み clinic_id で制約する」
- 対象: `backend/internal/lstep/shared_file_repository.go:31`、`backend/internal/lstep/shared_file_repository.go:35`
- 内容: Create は f.ClinicID 信頼のみ。service は正しい clinic をセットするが defense-in-depth 欠落。
- 修正: clinicID 引数 + assert。

- round2-review(2026-07-28): **WITHDRAWN** — 反証: 唯一の production writer `sharedFileService.Upload` が認証 clinicID を常に設定（`shared_file_service.go:94-104`）。repo 署名は defense-in-depth のみ。
#### G2C-07: tag code mapping replace に監査が無い — LOW
- 区分: 新規 ／ 横断: X-03
- 規約: `.claude/refs/backend-application-invariants.md:31` 「destructive または irreversible な操作には、権限、対象 scope、監査、recovery 方針を持たせる。」
- 対象: `backend/internal/lstep/lstep_tag_code_mapping_service.go:69`、`backend/internal/lstep/lstep_tag_code_mapping_service.go:109`
- 内容: automation 設定の soft-delete+insert に audit 無し。
- 修正: 成功後に actor+tag_name を監査。

- round2-review(2026-07-28): **WITHDRAWN** — 反証: soft-delete+insert は再設定可能で irreversible ではない。invariants:31 の destructive 監査必須に該当しない。
#### G2C-08: tag sync が nil/空 config で破壊的クリーンアップまたはクリーンアップ不能 — LOW
- 区分: 新規
- 規約: `.claude/refs/error-handling.md:9` 「error を無視しない。処理できる境界まで返すか、明示的に回復する。」
- 対象: `backend/internal/lstep/lstep_tag_sync_care_chronic.go:41`、`backend/internal/lstep/lstep_tag_sync_pet_basic.go:37`
- 内容: nil config → 全 chronic strip、空 prefix → stale 非削除。composition では通常注入される。
- 修正: 必須依存欠落は fail-closed。

- round2-review(2026-07-28): **WITHDRAWN** — 反証: composition が `repos.tagConfig` を常時注入（`composition_services.go:33-37`）。nil 分岐は test 耐性。error 無視ではない。
#### G2P-01: inventory/merchandise の使用中カウント→削除が tx 外 TOCTOU — LOW
- 区分: 新規 ／ 横断: X-05
- 規約: `backend/CODING_RULES.md:38` 「`FOR UPDATE`、`FOR SHARE`、`pg_advisory_xact_lock`を正しさの根拠にするoperationはambient transaction不在を拒否する。…」の運用解釈（依存チェック後 delete の直列化。規約疑義 #27 も参照）
- 対象: `backend/internal/inventory/inventory_service.go:172`、`backend/internal/inventory/merchandise_item_service.go:209`、`backend/internal/inventory/merchandise_item_repository.go:68`
- 内容: count→delete が独立 statement。merchandise repo は DBOrTx 非参加。
- 修正: WithTx + FOR UPDATE + 条件付き delete。merchandise を DBOrTx 化。

- round2-review(2026-07-28): **WITHDRAWN** — 反証: CODING_RULES:38 字面は FOR UPDATE/SHARE/advisory を正しさ根拠にする operation の ambient 拒否。count→delete は lock 非依存。「運用解釈」は規約拡張のため cite 失格。TOCTOU 残存は New Work として分離。
#### G2P-02: inventory/merchandise Update 後 re-fetch が tx 外 — MEDIUM
- 区分: **既知** → BUG-465 ／ 横断: X-01
- 規約: `backend/CODING_RULES.md:78` 「write後の再取得が失敗し得る場合はcommit前の同じtransaction内で行うか、commit済みの成功を後段read errorで失敗へ反転させないcontractにする。」
- 対象: `backend/internal/inventory/repository.go:117`、`backend/internal/inventory/merchandise_item_repository.go:55`
- 内容: UpdateScopedByID 後の別 Find が 5xx 反転し得る。
- 修正: 同一 tx 内 UpdateAndFind。

- round2-review(2026-07-28): **UPHELD** — 反証: 稀有だが CODING_RULES:78 字面どおり（`inventory/repository.go:117-121` Update 後別 Find）。成功の 5xx 反転 contract 違反。


#### G2P-03: inventory quantity が非負未検証で DecreaseStock が負在庫を作れる — MEDIUM
- 区分: **既知** → BUG-466 ／ 横断: X-02
- 規約: `backend/CODING_RULES.md:79` 「schema constraint と application validation の両方を使う。」
- 対象: `backend/internal/inventory/inventory_request.go:38`、`backend/internal/inventory/repository.go:133`、`backend/migrations/001_init.sql:668`
- 内容: quantity に min=0 / CHECK 無し。DecreaseStock は quantity-N 無条件。
- 修正: binding min=0、WHERE quantity>=? 条件付き減算、CHECK。

- round2-review(2026-07-28): **UPHELD** — 反証: DecreaseStock の無条件 `quantity - N` と min/CHECK 欠落は CODING_RULES:79 に直撃。負在庫は status で隠れるだけで数値が壊れる。


#### G2T-02: inquiry_template CountUsage が常に 0 を返す AUS-04 同型 — MEDIUM
- 区分: **決裁済み**（PO 2026-05-25: inquiry_answers 当面実装しない）／ 旧「新規」は撤回
- 規約: `backend/internal/medicalrecord/inquiry_template_repository.go:71` 「PO判断（2026-05-25）: inquiry_answers は当面実装しない。」（旧 error-handling.md:9 および出典なし interface 規則は本事象に不適用のため差し替え）
- 対象: `backend/internal/medicalrecord/inquiry_template_repository.go:70`、`backend/internal/medicalrecord/inquiry_template_repository_test.go:196`、`backend/internal/medicalrecord/inquiry_template_service.go:143`
- 内容: CountUsage が常に 0。Delete の使用中ガードは本番到達不能。repo test は count==0 の恒真。
- 修正: （決裁優先）stub は inquiry_answers 実装 PR 内で COUNT に置換。それまでは実装/除去を強制しない。恒真 test は任意 cleanup。

- round2-review(2026-07-28): **WITHDRAWN**（区分: **決裁済み**） — `inquiry_template_repository.go:70-74`「PO判断（2026-05-25）: inquiry_answers は当面実装しない。」常時 0 は意図的スタブ。error-handling.md:9 は error 無視ではない。恒真 test はドキュメント的 residual のみ。**修正: 実装するか除去** は決裁と矛盾するため無効。
#### G2T-03: daily_record_vital_tx_atomicity_test の assert.True(t, true) — LOW
- 区分: 新規
- 規約: `.claude/refs/error-handling.md:9` 「error を無視しない。処理できる境界まで返すか、明示的に回復する。」— **本事象（恒真 assert）には不適用**と判定し所見 WITHDRAWN。代替条項なし（製品規則外の test noise）。
- 対象: `backend/internal/medicalrecord/daily_record_vital_tx_atomicity_test.go:54`
- 内容: 有意味 assert の後に恒真 1 行。
- 修正: 行削除。

- round2-review(2026-07-28): **WITHDRAWN** — `assert.True(t, true)` は有意味 assert 後のノイズ1行（`daily_record_vital_tx_atomicity_test.go:48-54`）。error-handling.md:9 の「test 側対偶」は創作ルールで MUST_FIX。製品リスクなし。
#### G2F-01: 配信トリガー batch が owner ごと多段 query（N+1） — HIGH
- 区分: 新規
- 規約: `.claude/refs/go-gin-backend-review.md:67` 「N+1、unbounded query、missing indexを実測とquery planに基づいて改善しているか（推測のみで最適化・放置していないか）。」— **本所見は EXPLAIN 未実施の構造導出**であり、:67 の「実測に基づく改善」完了判定には使えない。改善優先度確定には要実測。
- 対象: `backend/internal/lstep/lstep_delivery_trigger_batch.go:34`、`backend/internal/lstep/lstep_delivery_trigger_state.go:11`、`backend/internal/lstep/lstep_tag_cache_repository.go:233`
- 内容: owner ごとに FindByID / ExistsToday / tag cache 等。候補 ID 集合も Limit 無し。
- 修正: IN 一括取得 + cursor chunk。

- round2-review(2026-07-28): **REFRAMED**（構造 N+1 は維持・HIGH） — 反証: `runBatch`→`processSingleOwner` 内で owner ごと Find/Exists/tag（`lstep_delivery_trigger_batch.go:34-55` 等）。測定なしで :67 を充足した扱いは不可。
#### G2F-02: health prevention batch が owner ごと無制限 history Find — HIGH
- 区分: 新規
- 規約: `.claude/refs/go-gin-backend-review.md:67` 「N+1、unbounded query、missing indexを実測とquery planに基づいて改善しているか（推測のみで最適化・放置していないか）。」— **本所見は EXPLAIN 未実施の構造導出**であり、:67 の「実測に基づく改善」完了判定には使えない。改善優先度確定には要実測。
- 対象: `backend/internal/lstep/lstep_health_tag_sync_batch.go:56`、`backend/internal/lstep/lstep_health_tag_sync_checkup.go:43`、`backend/internal/lstep/lstep_health_tag_sync_vaccine.go:26`
- 内容: page は cursor でも child Find が owner 全件無 Limit。
- 修正: page 単位 bulk SQL。

- round2-review(2026-07-28): **REFRAMED**（構造 multi-query 維持・HIGH） — 反証: outer cursor はあるが child `FindByOwnerID`/`FindByOwner` は無 LIMIT（checkup/vaccine）。:67 実測要件は未充足。
#### G2F-03: LTV 同期が clinic 全 owner を unbounded materialize — HIGH
- 区分: 新規
- 規約: `.claude/refs/go-gin-backend-review.md:67` 「N+1、unbounded query、missing indexを実測とquery planに基づいて改善しているか（推測のみで最適化・放置していないか）。」— **本所見は EXPLAIN 未実施の構造導出**であり、:67 の「実測に基づく改善」完了判定には使えない。改善優先度確定には要実測。
- 対象: `backend/internal/lstep/lstep_tag_sync_visit_ltv.go:20`、`backend/internal/billing/accounting_repository_ltv.go:67`、`backend/internal/owner/repository.go:381`
- 内容: コメントで全件ロード自認。FindAllWithLineUserID も Limit 無し。
- 修正: cursor / DB 側 top-N。

- round2-review(2026-07-28): **DOWNGRADED** (HIGH→MEDIUM) + **REFRAMED** — 反証: tag cache は FindByOwners で N+1 解消済み。残は `FindAllWithLineUserID` 等の unbounded materialize（`owner/repository.go:381`）。N+1 表記は不正確。
#### G2F-04: visit-dormant タグ同期が非 cursor dormant list を使用 — HIGH
- 区分: 新規
- 規約: `.claude/refs/go-gin-backend-review.md:67` 「N+1、unbounded query、missing indexを実測とquery planに基づいて改善しているか（推測のみで最適化・放置していないか）。」— **本所見は EXPLAIN 未実施の構造導出**であり、:67 の「実測に基づく改善」完了判定には使えない。改善優先度確定には要実測。
- 対象: `backend/internal/lstep/lstep_batch_segmentation.go:20`、`backend/internal/medicalrecord/medical_record_owner_visit_repository.go:192`
- 内容: 検出側は Cursor 済みだがタグ同期は旧 unbounded API。
- 修正: Cursor ループへ置換。

- round2-review(2026-07-28): **REFRAMED**（構造 unbounded 維持・HIGH） — 反証: 検出は Cursor、タグ同期のみ旧 API（`lstep_batch_segmentation.go:22`）。:67 は要実測注記必須。
#### G2F-05: LINE customers list が Preload 付き unbounded — MEDIUM
- 区分: 新規
- 規約: `.claude/refs/go-gin-backend-review.md:67` 「N+1、unbounded query、missing indexを実測とquery planに基づいて改善しているか（推測のみで最適化・放置していないか）。」— **本所見は EXPLAIN 未実施の構造導出**であり、:67 の「実測に基づく改善」完了判定には使えない。改善優先度確定には要実測。
- 対象: `backend/internal/lstep/line_customer_repository.go:30`、`backend/internal/lstep/line_customer_handler.go:30`
- 内容: ページネーション無し全件 + Owner Preload。
- 修正: ParsePaginationWithMax。

- round2-review(2026-07-28): **REFRAMED** — 反証: 単一 Find+Preload で N+1 ではない（`line_customer_repository.go:30-39`）。ページネーション欠落の API 契約問題。
#### G2F-06: shared files list/cleanup が unbounded + per-row delete — MEDIUM
- 区分: 新規
- 規約: `.claude/refs/go-gin-backend-review.md:67` 「N+1、unbounded query、missing indexを実測とquery planに基づいて改善しているか（推測のみで最適化・放置していないか）。」— **本所見は EXPLAIN 未実施の構造導出**であり、:67 の「実測に基づく改善」完了判定には使えない。改善優先度確定には要実測。
- 対象: `backend/internal/lstep/shared_file_repository.go:50`、`backend/internal/lstep/shared_file_service.go:162`
- 内容: FindAll/FindExpired 無 Limit。cleanup 1 件ずつ。
- 修正: list limit、cleanup batch。

- round2-review(2026-07-28): **REFRAMED** — 反証: FindAll 無 Limit は成立。per-row delete は storage+DB 対の意図的逐次処理でもある。N+1 SELECT ではない。
#### G2F-07: 管理画面予約月次/顧客履歴が hard cap 無し — MEDIUM
- 区分: 新規
- 規約: `.claude/refs/go-gin-backend-review.md:67` 「N+1、unbounded query、missing indexを実測とquery planに基づいて改善しているか（推測のみで最適化・放置していないか）。」— **本所見は EXPLAIN 未実施の構造導出**であり、:67 の「実測に基づく改善」完了判定には使えない。改善優先度確定には要実測。
- 対象: `backend/internal/reservation/appointment_admin_repository.go:37`、`backend/internal/reservation/appointment_admin_repository.go:106`
- 内容: 月次は日付範囲のみ、顧客履歴は生涯件数。
- 修正: 履歴 page、月次 safety cap。

- round2-review(2026-07-28): **DOWNGRADED** (MEDIUM→LOW) — 反証: 月次は calendar bound（`FindAllByMonth`）。生涯履歴のみ soft unbounded。:67「unbounded」は月次に過大。
#### G2F-08: no-show 候補 unbounded + per-row WithTx — MEDIUM
- 区分: 新規
- 規約: `.claude/refs/go-gin-backend-review.md:67` 「N+1、unbounded query、missing indexを実測とquery planに基づいて改善しているか（推測のみで最適化・放置していないか）。」— **本所見は EXPLAIN 未実施の構造導出**であり、:67 の「実測に基づく改善」完了判定には使えない。改善優先度確定には要実測。
- 対象: `backend/internal/reservation/reservation_repository.go:689`、`backend/internal/lstep/lstep_batch_noshow.go:88`
- 内容: 候補 Limit 無し。各候補で tx+audit。
- 修正: LIMIT/cursor + batch size。

- round2-review(2026-07-28): **REFRAMED** — 反証: per-row WithTx は audit fail-closed の正しさ要件。N+1 欠陥ではない。候補 LIMIT 欠落のみ残す。
#### G2F-09: レジ締め詳細が completed billings 全行 Scan — MEDIUM
- 区分: 新規
- 規約: `.claude/refs/go-gin-backend-review.md:67` 「N+1、unbounded query、missing indexを実測とquery planに基づいて改善しているか（推測のみで最適化・放置していないか）。」— **本所見は EXPLAIN 未実施の構造導出**であり、:67 の「実測に基づく改善」完了判定には使えない。改善優先度確定には要実測。
- 対象: `backend/internal/billing/accounting_repository_reports_close.go:135`、`backend/internal/billing/accounting_repository_reports_close.go:183`
- 内容: 日次 detail dump が unbounded（busy day リスク）。※当該 file は dirty 並行編集中。
- 修正: optional/paginated detail。

- round2-review(2026-07-28): **DOWNGRADED** (MEDIUM→LOW) — 反証: 締め期間で bound された単一 Raw SQL。ループは in-memory map のみ（`accounting_repository_reports_close.go:140-186`）。lifetime unbounded ではない。
#### G2F-10: shift_entry list が year_month 未指定で全件 — MEDIUM
- 区分: 新規
- 規約: `.claude/refs/go-gin-backend-review.md:67` 「N+1、unbounded query、missing indexを実測とquery planに基づいて改善しているか（推測のみで最適化・放置していないか）。」— **本所見は EXPLAIN 未実施の構造導出**であり、:67 の「実測に基づく改善」完了判定には使えない。改善優先度確定には要実測。
- 対象: `backend/internal/staff/shift_entry_repository.go:70`、`backend/internal/staff/shift_entry_service.go:114`
- 内容: year_month 省略で全 shift_entries + Preload。
- 修正: required または当月 default。

- round2-review(2026-07-28): **REFRAMED** — 反証: year_month 空で date filter 省略は構造的（`shift_entry_repository.go:70-87`）。:67 実測なし。API 契約（必須 or 当月 default）問題。
#### G2F-11: 一部 master FindAll に MaxMasterListRows 未適用 — LOW
- 区分: 新規
- 規約: `.claude/refs/go-gin-backend-review.md:67` 「N+1、unbounded query、missing indexを実測とquery planに基づいて改善しているか（推測のみで最適化・放置していないか）。」— **本所見は EXPLAIN 未実施の構造導出**であり、:67 の「実測に基づく改善」完了判定には使えない。改善優先度確定には要実測。
- 対象: `backend/internal/medicalrecord/consultation_repository.go:37`、`backend/internal/medicalrecord/exam_type_repository.go:46`、`backend/internal/inventory/merchandise_item_repository.go:32`
- 内容: vaccine/procedure のみ cap。他 master は safety Limit 不整合。
- 修正: master list 一律 Limit。

### 既知として新規に再記述しない項目

| 観察 | 扱い |
|---|---|
| G2T-01 / AUS-04 未修正 | **区分: 既知** — AUS-04 を参照。round 2 は未修正を CONFIRMED 相当で確認したのみ。 |
| INF-03 | infra 全文再読で追加欠陥なし。既存 INF-03 のまま。 |
| MRC-02 が inventory を使用 | inventory 側に新規 G2P を切り出し。name cascade 本体は MRC-02。 |

### round 2 サマリー数

| 項目 | 数 |
|---|---:|
| 既存 105 の再実測注記 | 105（HIGH 以上 4 分類、MEDIUM/LOW 位置照合） |
| ALREADY-FIXED | 2 |
| 新規所見 | **40**（G2A 7 + G2B 5 + G2C 8 + G2P 3 + G2T 2 + G2F 11 + 調整）※G2T-01 は既知扱いのため未採番新規から除外 |
| 新規 CRITICAL | 0 |
| 新規 HIGH | 9（G2A-01, G2B-01, G2F-01..04 ほか） |

- round2-review(2026-07-28): **DOWNGRADED** (LOW→nit) — 反証: master は小 cardinality。MaxMasterListRows 不整合は一貫性の話で measured unbounded ではない。

---

## round 2 是正（2026-07-28）— false positive 除去・規約適用性・畳み込み

### 測定基準
- HEAD at write: 実行セッションの `git rev-parse HEAD`（Phase 0 スナップショット相対）
- 対象: round 2 新規 36 所見（G2*）+ 横断パターン表 + 決裁コメント突合
- 方針: **却下バイアス**。迷ったら UPHELD ではなく WITHDRAWN/DOWNGRADED/REFRAMED

### 36 所見 判定サマリー

| 判定 | 件数 | IDs |
|---|---:|---|
| UPHELD | 3 | G2C-04, G2P-02, G2P-03 |
| WITHDRAWN | 15 | G2A-02, G2A-03, G2A-04, G2A-06, G2A-07, G2B-04, G2B-05, G2C-03, G2C-05, G2C-06, G2C-07, G2C-08, G2P-01, G2T-02, G2T-03 |
| DOWNGRADED | 10 | G2A-01, G2A-05, G2B-01, G2B-02, G2C-01, G2C-02, G2F-03, G2F-07, G2F-09, G2F-11 |
| REFRAMED | 8 | G2B-03, G2F-01, G2F-02, G2F-04, G2F-05, G2F-06, G2F-08, G2F-10 |

※ G2F 11 件はすべて `:67` の「実測と query plan」を**未充足**と判定し、構造導出・要実測を明示（REFRAMED または DOWNGRADED）。測定済み :67 充足として UPHELD した G2F は **0**。

### 規約引用の適用性（母数 36 / 実施 36）

| 結果 | 件数 | 処理 |
|---|---:|---|
| 適用可のまま | 多数 | 変更なし |
| 差し替え | 数件 | G2B-03→CODING_RULES:36; G2C-04→guidelines:154 + error-handling:29; G2T-02/03 撤回; G2F-* に要実測 caveat |
| 出典なしルール除去 | 3 | G2T-02「interface 最小化」; G2T-03「test 側対偶」; G2C-04「Gin運用」無 path |

`grep '^- 規約:' BE-refactor.md | grep -vc ':[0-9]'` 目標: 0（撤回表記の括弧書きを除く）。

### error-handling.md:9 過剰適用
- **差し替え/撤回**: G2A-04（RowsAffected→非適用）, G2T-02, G2T-03, G2C-08 系は依存欠落, G2C-03 は意図的 best-effort で WITHDRAWN
- **正当利用として残る例**: G2A-06 は WITHDRAWN（意図的 recovery）, G2B-01 は DOWNGRADED（best-effort と両立）

### 決裁済み設計判断の走査
- コマンド: `rg -n 'PO判断|ADR-|当面実装しない|意図的に' backend/internal --glob '*.go' --glob '!*_test.go'`
- ヒット: **52**
- 所見 `対象:` と同一ファイル±30〜40行で重なった件数: **9 近傍ヒット**（実質的に決裁が欠陥主張を覆すのは **G2T-02 のみ**）
- **G2T-02**: 区分 **決裁済み** + WITHDRAWN（production 欠陥 framing 取消）
- 他近傍（MRA-01, MRA-04, RSV-08, AUS-05, MDL-01/02/06 等）: **決裁と無関係**（ADR package 境界や別フィールドの PO）

### 横断パターン表の更新（fold 実反映）
- X-02: 除去 MDL-04, POC-14; 追加 TRM-03（兼 X-07）
- X-04: 追加 LSB-02; 注記 LSA-02/LSB-01 は非 X-04
- X-05: 追加 POC-06, LSA-15
- X-07: TRM-03 兼 X-02
- X-08: 除去 LSB-05（正本 LSA-08）
- X-09: 追加 POC-14
- サマリー L43 の X-04 誤配置文言を修正

### G2F 11 件の根拠条項処理
| 処理 | IDs |
|---|---|
| REFRAMED（構造維持+要実測） | G2F-01,02,04,05,06,08,10 |
| DOWNGRADED | G2F-03 (HIGH→MED), G2F-07 (→LOW), G2F-09 (→LOW), G2F-11 (→nit) |
| :67 を測定済み根拠として UPHELD | **0** |

### New Work Surfaced（scope 外・文書に新規 G2 として追加しない）
1. G2P-01 の count→delete TOCTOU は工学的に妥当だが CODING_RULES:38 字面外。規約疑義 #27 の成文化後に再起票可。
2. G2C-03 / G2B-01 系 best-effort を CODING_RULES:36 の「補償必須」に照らした製品契約の明文化。
3. AUS-04（shift_template CountUsage）は PO ではない compatibility stub — 既存 AUS-04 のまま。

---

## round 3（2026-07-28）— round 1 の 105 所見へ敵対的反証

### 測定基準
- **HEAD**: `2a76890801a44e3e99bcc8f6362c8cbf741e73ab`（Phase 0 `git rev-parse HEAD`）
- **working tree (backend/)**: clean at Phase 0（`git status --porcelain -- backend/` empty）
- **対象**: round 1 由来 105 件のみ（G2* 36 件は `552550403` 反証済み・再判定しない）
- **方法**: 並列 subagent 10 本（R1-crit / R1-mr / R1-lstep / R1-poc-rsv / R1-med-a/b/c / R1-low / R1-rule / R1-decision）→ 主エージェント join → 単独 writer

### 判定サマリー（母数 105 / 実施 105）

| 判定 | 全体 | Tier A (CRIT+HIGH 35) | Tier B (MED+LOW 70) |
|---|---:|---:|---:|
| UPHELD | 72 | 27 | 45 |
| WITHDRAWN | 8 | 2 | 6 |
| DOWNGRADED | 8 | 2 | 6 |
| REFRAMED | 17 | 4 | 13 |

### WITHDRAWN（8）
| ID | 一行理由 |
|---|---|
| AUS-06 | 800 行上限は ECC soft style。Go/Gin guidelines はファイルサイズを公式要件としない（:227）。強制 ADR なし。 |
| CMD-01 | ALREADY-FIXED。`backend/cmd/stage-import` は HEAD から消失（BUG-430 退役）。破壊的 unscoped DELETE の live path 無し。 |
| INF-01 | ALREADY-FIXED。`httpapi/response.go:90-107` が未知 PgError を 500 に落とし、`classifyPgError` は allowlist のみ known=true（BUG-2026-07-27-01）。欠陥パスは HEAD  |
| LSA-16 | json.Marshal 失敗は当該引数型では構造的に到達不能。 |
| LSB-05 | LSA-08 と同一疎通 raw error 面の重複（X-08 も除外済み）。 |
| MDL-02 | testdb TRUNCATE の error 破棄は test harness 品質。product MEDIUM として不成立。 |
| MDL-04 | LSA-13 と同一 trigger_type 欠陥の重複（X-02 も MDL-04 除外済み）。 |
| TRM-10 | export stutter は style preference。欠陥扱いは過剰。 |

### DOWNGRADED（8）
| ID | 新 severity | 一行理由 |
|---|---|---|
| CMD-02 | LOW | Go router は未認証だが production は Worker edge 隔離前提。defense-in-depth/文書化ギャップ。 |
| CMD-03 | LOW | coverage-ratchet は CI tooling。runtime safety ではない。workflow exit-code 契約。 |
| CMD-05 | LOW | release は STORAGE_TYPE=s3 強制。local StaticFS は dev 露出。prod PHI StaticFS ではない。 |
| INF-04 | LOW | message 文字列比較は error-handling.md:18 違反だがメンテ脆弱性中心。 |
| LSA-07 | MEDIUM | LINE user id の Info ログは CODING_RULES.md:68 違反だが HIGH の秘密流出級ではない。 |
| LSA-12 | LOW | status 更新失敗は観測歪み。タグ失敗自体は batch Failed に載る。翌日抑止主張は過大。 |
| MRB-02 | MEDIUM | LockByIDForUpdate が ambient-tx ガード欠落は実在（examination 対比）だが production 呼び出しは常に WithTx 内。潜在 API 契約穴。 |
| POC-11 | LOW | validator 複製は将来 drift リスク。現状 accept set は同等。 |

### REFRAMED（17）
| ID | 一行理由 |
|---|---|
| AUS-09 | 二重 response パターンは防御的だが Authenticate 後は実質到達不能。 |
| BIL-01 | 内容は成立するが invariants:37 単独の CRITICAL 一本化は過剰。A) 締め後総額書換の post-close 権限・理由・同一tx監査欠落（CRITICAL・invariants:37 + cash-register #115）と B) Create/Upd |
| CMD-04 | typed LifecycleWriter 自体は意図的 narrowing。欠陥は cutover 未完の dual map path 債務。 |
| INF-06 | MarshalJSON の omit は意図的 best-effort。残るのは observability のみ。project quality policy。 |
| LSA-02 | 実害は成立するが CODING_RULES.md:42（status transition CAS）適用は拡張解釈。正本は LSB-01（business-evidence fail-open）。修正は LSB-01 に畳む。 |
| LSA-04 | global master 無 clinic_id は設計判断（boundary-map）。欠陥は hospital-settings RBAC で全院共有 master を mutate できる認可境界不一致。 |
| MDL-06 | 既知 TASK-445 / DEC-28。新規ではなく既知ポインタ。 |
| MRA-04 | package comment の誤 package 名は GoDoc ノイズのみ。runtime 影響ゼロ。 |
| MRB-03 | 未 scope Preload は isolation ギャップだが BUG-437 / SEC-SWEEP-02 既知面。新規 HIGH ではなく既知ポインタ。区分: 既知。 |
| MRC-08 | DTO 無検証は guidelines:151 で成立。ただし「信頼境界敗北」ではなく境界 validation 不足（lab 参照値は import 設計上 client 由来）。 |
| MRC-12 | 既知 phase2.html:195 / X-06。新規 MEDIUM ではなく既知ポインタ。 |
| POC-01 | 二重 route は意図的。問題は default 権限と FE が閉じる path の契約不一致であり、権限 OR 昇格ではない。 |
| POC-09 | ClinicScope on INSERT は no-op だが ClinicID は arg から設定され isolation は成立。dead scope。 |
| POC-16 | parseHHMM 複製。CODING_RULES.md:40（facade）は適用外。coding-style:26 の DRY メモ。 |
| RSV-08 | 既知 SEC-SWEEP-02 孫 correlation 面。独立 NEW ではない。 |
| TRM-08 | 既知 SEC-SWEEP-02 / DEC-23。parent JOIN は部分修正済み。 |
| TRM-09 | WIP-adjacent / dbortx inventory。現呼び出しに ambient tx 無し。 |

### 規約引用の適用性（母数 105 / 実施 105）
- APPLIES 81 / PARTIAL 15 / DOES_NOT_APPLY 9 / MISSING_CITE 0
- **LINE_DRIFT 是正**: `go-gin-backend-guidelines.md:179`→`:180`（INF-02, POC-12）; `:166`→`:167`（MRB-07; INF-01 は WITHDRAWN）
- **適用差し替え**: LSB-01（CLAUDE.md:25 → migration opt-out 契約）; RSV-02（CODING_RULES:38 → :42）; POC-14（guidelines:151 → coding-style:26）; POC-16（CODING_RULES:40 → coding-style:26）; MRC-07（invariants:31 → CODING_RULES:38）

### 決裁済み設計判断の走査
- grep non-test ヒット: **42**（`PO判断|当面実装しない|意図的に|BUG-415|BUG-430|intentionally|Intentional`）
- 所見 対象 との file 重なり: **11** findings / **9** decision files
- **決裁を欠陥扱いして WITHDRAWN にした件数: 0**
- 特記: inquiry_answers PO判断 / BUG-415 Status 省略 / inventory SD-4 / Method PO判断B は所見と衝突せず維持
- CMD-04 のみ REFRAMED（意図的 temporary dual path の cutover 債務）

### サマリー表・横断パターン表への反映
- CRITICAL open: CMD-01 **WITHDRAWN** → open CRITICAL は BIL-01（REFRAMED 分割後も post-close 迂回は open）の **1**
- HIGH: INF-01 WITHDRAWN; MRB-02/LSA-07 DOWNGRADED; LSA-02→LSB-01 畳み込み; MRB-03 既知 REFRAME
- MEDIUM: AUS-06/MDL-02/MDL-04 WITHDRAWN; INF-04/LSA-12/POC-11/CMD-02/03/05 DOWNGRADED
- LOW: LSA-16/LSB-05/TRM-10 WITHDRAWN; 複数 REFRAME を既知/WIP ポインタ化
- X-08 から LSB-05 除去（WITHDRAWN 重複）; X-02 から MDL-04 除去（WITHDRAWN 重複 LSA-13）

### New Work Surfaced（scope 外・新規所見として追加しない）
- billing-items の status guard と post-close audit は別命題（BIL-01 REFRAME 分割）
- L-step write ops が policy stub の間、LSA-11/LSA-15 の外部影響はミュート（制御フロー欠陥は残る）
- AUS-06 800 行ルールを backend に ADR で強制するかの規約側議論

## 着手プラン

- 基準 HEAD: `09ea4716455fc013b8ad99c5d7e8f1383a92cf40`
- 対象導出: 総所見 141、見出しスコープの `WITHDRAWN` 23、live 118。既知は減算しない。
- 導出コマンド: `grep -c '^#### ' BE-refactor.md` → `141`
- 撤回コマンド: `awk '/^#### /{id=$2; sub(/:$/,"",id)} /^- round[23]-review/ && /WITHDRAWN/{print id}' BE-refactor.md | sort -u | wc -l` → `23`
- 既知の扱い: 本文 marker 9件、表のみの `RSV-08` を含む。`CMD-01` は既知かつ WITHDRAWN のため live から1回だけ除外し、live 中の既知は9件。
- 実装 unit 数: 84。別に finding を所有しない shared schema barrier 1 を置き、`backend/migrations/001_init.sql` の単一 writer を固定する。全実装 unit と barrier は所有パス8本以内。
- 統合検証: live 118/118、撤回混入0、所見重複0、所有パス重複0、missing path 0、所有パス382/382、最大8本。
- 着手時注意: 承認済み再開 baseline には `grok-handoff-plan.md` と `3-session-agent.html` の別セッション変更が存在する。各 unit は自身の所有パスを着手直前に再実測し、他者の変更を revert しない。

### Contract summary

- Live findings: 118.
- Finding-owning units: 84.
- Shared-owner barrier units: 1.
- Concrete owned paths: 382.
- Maximum unit size: 8 concrete existing files.
- `backend/` is planning evidence only; no repository file was edited by this probe.
- MRB-08 correction: `U-X05-MR-EXAMTYPE` owns five exact files and includes the mandatory DBOrTx lintscan gate.
- LSA-02 is mapped once and folded into the same implementation unit as canonical LSB-01; both live IDs remain independently accounted.

### 決裁結果（2026-07-28・USER 裁定）

`DEC-29`〜`DEC-39` の 11 論点は **全件「推奨どおり（選択肢 A）」で裁定済み**。裁定日 2026-07-28、裁定者 = USER（「推奨があるのであれば推奨通りでいい」）。**これらを依存に持つユニットの blocked 状態は解除される。**

| 決裁 ID | 裁定 | 影響所見 | 採用した方針（推奨 A の要旨） |
|---|---|---|---|
| DEC-29 | A | `MDL-06` | 二重保持を増やさず、機械 gate を省略条件にして tenant isolation を fail-closed にする |
| DEC-30 | A | `LSA-04` | 現行の hospital-settings 操作モデルに整合させ、ある clinic の操作が他 clinic の自動化を変える経路を除去する |
| DEC-31 | A | `POC-07` | clinic-scoped 権限で global fact を変更できる境界をなくす |
| DEC-32 | A | `MRC-04` | 主訴・診断等の臨床記録は caller が保存成功を誤認しない原子 contract を優先する |
| DEC-33 | A | `BIL-01` | 締め snapshot と日次・月次再集計の不一致を、同じ post-close 権限・理由・監査で防止する |
| DEC-34 | A | `INF-04` | 既存の client-correctable 分類を保ちつつ例外面を限定し、upstream 文言 drift を test で検知する |
| DEC-35 | A | `LSA-03` `LSA-10` `LSA-11` `LSA-12` `LSB-02` `LSB-04` `MRC-04` `RSV-09` `INF-06` ほか | silent success と再試行不能を機械的に除外する。単なる slog を成功根拠にしない |
| DEC-36 | A | `CMD-02` | 全 clinic batch のような高影響 route は deployment topology に依存しない defense-in-depth とする |
| DEC-37 | A | `INF-02` `POC-12` `TRM-03` | 未設定 handler と middleware 順序 drift を一つの gate で防ぐ |
| DEC-38 | A | `LSA-02` `LSB-01` | opt-out の business evidence を外部タグ同期の成否に依存させず fail-closed に評価する |
| DEC-39 | A | `TRM-09` | 将来 caller が `WithTx` 化されたとき write が黙って transaction 外へ逃げる contract drift を防ぐ |

**着手時の扱い**: 各ユニットの `決裁:` 欄が上記 ID を指している場合、選択肢 A の方針で実装する。A の具体的な選択肢本文は `q&a.html` の当該パケットを参照（本書へ全文転記しない — 二重管理を避ける）。

### セッション振り分け（2026-07-28・依存成分ベースへ再設計）

全 84 ユニットを **4 レーン**へ配分する。1 レーン = 1 セッション。

**分割軸 = ユニット依存グラフの連結成分**。依存エッジは 23 本しかなく、連結成分は 62 個（うち孤立 50 個・最大成分 5 ユニット）。依存で結ばれたユニットを同一レーンへ閉じ込めたため、**レーンをまたぐ先行依存は 0 件**。各レーンは他レーンの完了を待たずに独立して最後まで進める。

**所有パスの排他性**: 全ユニットの所有パスは相互に非重複であり、レーンをまたぐファイル書き込み衝突も起きない。

| レーン | ユニット数 | 主な担当パッケージ |
|---|---:|---|
| LANE-1 | 24 | `internal/lstep`(14), `internal/clinic`(3), `internal/billing`(2), `internal/reservation`(2), `internal/owner`(2) |
| LANE-2 | 19 | `internal/pet`(3), `cmd/api`(3), `internal/reservation`(3), `internal/trimming`(2), `cmd/csv`(1) |
| LANE-3 | 18 | `internal/medicalrecord`(15), `internal/lintscan`(2), `internal/inventory`(1) |
| LANE-4 | 23 | `internal/lstep`(12), `internal/billing`(5), `internal/staff`(3), `internal/owner`(2), `internal/auth`(1) |

#### Wave 0 — 全レーン共通の前提（`U-SCHEMA-BARRIER`）

- **状態の正本**: `### Unit-by-unit start order` の `##### U-SCHEMA-BARRIER` ブロック内の `- Status:` 行。**Wave 0 節に状態を二重に持たない**（2026-07-28 に本節へ重複行を新設して不整合を起こしたため撤去）。他レーンは依存ユニット着手前に当該ブロックの `- Status:` を確認する。確認コマンド: `grep -A1 '^- Size: S (1/8 files)' BE-refactor.md`、または `grep -n 'U-SCHEMA-BARRIER' BE-refactor.md` で unit ブロックを開く。
- **前提ブロッカー（2026-07-28 検出）**: barrier の SchemaDrift 検証は `Payment.clinic_id` を要求するが live DB に未適用（migration 005）。**`make migrate` はユーザー専権であり、エージェントは自動適用しない。** ユーザーが実行するまで barrier は着手不能であり、依存 16 ユニットも進まない。
- **lintscan allowlist の所有権例外**: `backend/internal/lintscan/dbortx_inventory_lint_test.go` はプラン上 `SOLO-36` の所有だが、`DBOrTx` の新規参加者を作る全ユニットが同ファイルの allowlist 登録を必要とする。**allowlist 行の追記に限り、どのレーンも同ファイルへ書いてよい**（登録行の追加のみ。テスト本体の改変は SOLO-36 の担当）。追記は path-scoped commit に含め、Deliverables に記録すること。

`U-SCHEMA-BARRIER` は共有 schema を所有する barrier unit であり、**どのレーンにも属さない**。**16 ユニット**（4 レーンに分散）がこれを先行依存に持つ。

- **実行者: LANE-1**。LANE-1 は自レーンのユニットに着手する前に、まずこれを完了させる。
- **他の 3 レーンは、`U-SCHEMA-BARRIER` を先行に持たないユニットから先に着手してよい。** 依存を持つユニットに到達したら、本書の `- Status:` が `完了` になっていることを確認してから進む。
- 二重実行を避けるため、LANE-1 以外はこの unit を実行しない。

#### レーン別ユニット一覧（この順に着手する。レーン内の順序は依存を満たす）

**LANE-1** (24 units)

`BE-X06-BIL-CAMPAIGN-01`, `BE-X06-LSTEP-SETTINGS-01`, `BE-X06-RSV-CANCEL-01`, `BE-X08-LSTEP-CONNECTION-01`, `BE-X08-LSTEP-SEND-01`, `BE-X09-CLOSING-01`, `U-X02-PET-OWNER-FREETEXT`, `SOLO-08`, `U-X02-LSTEP-TAG-CONFIG`, `SOLO-09`, `U-LSTEP-OPTOUT`, `SOLO-12`, `U-X04-LSTEP-HEALTH-REMOVE`, `SOLO-14`, `U-X04-LSTEP-BATCH`, `U-X05-OWNER-PHONE`, `SOLO-15`, `U-X01X05-RESERVATION`, `SOLO-19`, `SOLO-32`, `SOLO-34`, `U-X01X03X04-LSTEP-LIFECYCLE`, `U-X01X05-CLINIC`, `U-X04X05-LSTEP-DELIVERY`

**LANE-2** (19 units)

`BE-X07-BODY-01`, `BE-X09-PET-PATCH-01`, `SOLO-04`, `U-X05-PET-UPDATE`, `SOLO-05`, `SOLO-06`, `SOLO-18`, `SOLO-22`, `SOLO-36`, `U-TRIMMING-SERVICE`, `U-X01X03-MANUALARTICLE`, `U-X02-CLINIC-CONTACT`, `U-X02-RESERVATION-SETTINGS`, `U-X03-CSVIMPORT-GUARD`, `U-X03-PET-SPECIES-AUDIT`, `U-X04-AUDIT-MARSHAL`, `U-X04-COVERAGE-RATCHET`, `U-X04-LSTEP-MIGRATE`, `U-X04-RESERVATION-AUTODELEGATE`

**LANE-3** (18 units)

`BE-X06-MEDICAL-ATOMIC-01`, `BE-X09-MEDICAL-DIAGNOSIS-01`, `U-X05-MR-EXAMTYPE`, `U-X01X02-INVENTORY`, `SOLO-21`, `SOLO-23`, `SOLO-25`, `SOLO-26`, `SOLO-29`, `SOLO-30`, `U-MR-TREATMENT-PLAN`, `U-X01-MR-PRESCRIPTION`, `U-X01X03-MR-CARE`, `U-X02-MR-CONSULTATION`, `U-X02-MR-LAB-IMPORT`, `U-X02X03X05-MR-HOSPITALIZATION`, `U-X04-MR-SUBRECORDS`, `U-X05-MR-MEDICINE-MASTERS`

**LANE-4** (23 units)

`BE-X09-ESTIMATE-TAX-01`, `BE-X09-PET-ENUMS-01`, `BE-X10-AUTH-RESPONSE-01`, `SOLO-01`, `SOLO-02`, `SOLO-03`, `SOLO-10`, `SOLO-11`, `SOLO-13`, `SOLO-16`, `SOLO-17`, `SOLO-20`, `SOLO-24`, `SOLO-33`, `U-X01-LSTEP-LINE-CUSTOMER`, `U-X02-LSTEP-AGGREGATION`, `U-X02-LSTEP-SHARED-FILE`, `U-X02-LSTEP-TAG-MAPPING`, `U-X02-LSTEP-TRIGGER-PRIORITY`, `U-X02-STAFF-TYPE`, `U-X03-STAFF-ASSIGNMENT-AUDIT`, `U-X04-LSTEP-OWNER-TAGS`, `U-X04-LSTEP-STALE-TAGS`

#### レーンをまたぐ先行依存

なし（84 ユニット間）。**4 レーンは互いの完了を待たずに独立して進行できる。**

ただし上記 Wave 0 の `U-SCHEMA-BARRIER` のみ全レーン共通の前提であり、これに依存する 16 ユニットは LANE-1 による barrier 完了後に着手する。依存を持たないユニットは barrier を待たずに着手してよい。

### Execution waves and parallel groups

| Wave | Parallel group | Units | Reason |
|---:|---|---|---|
| 0 | 0.1 | SOLO-04, SOLO-11, SOLO-16, SOLO-18, SOLO-26, SOLO-32, SOLO-33, U-LSTEP-OPTOUT, U-MR-TREATMENT-PLAN, U-SCHEMA-BARRIER, U-X01-LSTEP-LINE-CUSTOMER, U-X01-MR-PRESCRIPTION, U-X02-LSTEP-TAG-MAPPING, U-X02-LSTEP-TRIGGER-PRIORITY, U-X02-PET-OWNER-FREETEXT, U-X02-RESERVATION-SETTINGS, U-X03-PET-SPECIES-AUDIT, U-X03-STAFF-ASSIGNMENT-AUDIT, U-X05-MR-EXAMTYPE, U-X05-PET-UPDATE | Decision-gated roots and the sole schema owner; execute a unit only after its named packet is decided. |
| 1 | 1.1 | BE-X09-ESTIMATE-TAX-01, BE-X06-BIL-CAMPAIGN-01, BE-X06-LSTEP-SETTINGS-01, BE-X06-MEDICAL-ATOMIC-01, BE-X06-RSV-CANCEL-01, BE-X07-BODY-01, SOLO-24, U-X01X02-INVENTORY, U-X01X03-MANUALARTICLE, U-X01X03-MR-CARE, U-X01X05-CLINIC, U-X02-LSTEP-SHARED-FILE, U-X02-LSTEP-TAG-CONFIG, U-X02-MR-CONSULTATION, U-X02-STAFF-TYPE, U-X02X03X05-MR-HOSPITALIZATION, U-X03-CSVIMPORT-GUARD, U-X04-RESERVATION-AUTODELEGATE, U-X04X05-LSTEP-DELIVERY, U-X05-OWNER-PHONE | First dependency-free atomicity, isolation, audit, and contract units. |
| 2 | 2.1 | BE-X08-LSTEP-CONNECTION-01, BE-X08-LSTEP-SEND-01, BE-X09-PET-ENUMS-01, BE-X10-AUTH-RESPONSE-01, SOLO-09, U-TRIMMING-SERVICE, U-X01X05-RESERVATION, U-X02-CLINIC-CONTACT, U-X02-LSTEP-AGGREGATION, U-X02-MR-LAB-IMPORT, U-X05-MR-MEDICINE-MASTERS | Second layer after shared schema/repository owners. |
| 3 | 3.1 | BE-X09-CLOSING-01, BE-X09-MEDICAL-DIAGNOSIS-01, BE-X09-PET-PATCH-01, U-X01X03X04-LSTEP-LIFECYCLE, U-X04-AUDIT-MARSHAL, U-X04-COVERAGE-RATCHET, U-X04-LSTEP-BATCH, U-X04-LSTEP-HEALTH-REMOVE, U-X04-LSTEP-MIGRATE, U-X04-LSTEP-OWNER-TAGS, U-X04-LSTEP-STALE-TAGS | Third layer after public-boundary and shared-service owners. |
| 4 | 4.1 | SOLO-01, SOLO-02, SOLO-03, SOLO-05, SOLO-06, SOLO-08, SOLO-10, SOLO-12, SOLO-13, SOLO-14, SOLO-15, SOLO-17, SOLO-19, SOLO-20, SOLO-21, SOLO-22, SOLO-23, SOLO-25, SOLO-29, SOLO-30, SOLO-34, SOLO-36, U-X04-MR-SUBRECORDS | Fourth layer for dependent behavior/performance continuations. |

A unit with a decision ID remains blocked until that packet is decided. Within each displayed group, the ownership ledger is path-disjoint; explicit dependencies still take precedence over group membership.

### Unit-by-unit start order

| 順序 | 種別 | unit / barrier | 先行ユニット | 決裁 | この順である理由 |
|---:|---|---|---|---|---|
| 1 | 共有所有 barrier | `U-SCHEMA-BARRIER` | なし | なし | 所見 `none (shared-owner barrier; no live-ID ownership)`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 2 | 実装 unit | `BE-X09-ESTIMATE-TAX-01` | `U-SCHEMA-BARRIER` | なし | 所見 `MDL-01`。共有schema barrier完了後に会計total/tax contractを固定し、後続のfinancial integrity変更を安定させるため。 |
| 3 | 実装 unit | `SOLO-04` | なし | `DEC-36` | 所見 `CMD-02`, `CMD-05`。決裁境界または共有schema所有を先に固定するため。 |
| 4 | 実装 unit | `SOLO-11` | なし | なし | 所見 `G2B-03`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 5 | 実装 unit | `SOLO-16` | なし | なし | 所見 `G2F-05`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 6 | 実装 unit | `SOLO-18` | なし | なし | 所見 `G2F-07`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 7 | 実装 unit | `SOLO-26` | なし | なし | 所見 `RSV-08`, `TRM-08`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 8 | 実装 unit | `SOLO-32` | なし | なし | 所見 `POC-01`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 9 | 実装 unit | `SOLO-33` | なし | なし | 所見 `POC-08`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 10 | 実装 unit | `U-LSTEP-OPTOUT` | なし | `DEC-38` | 所見 `LSA-02`, `LSB-01`。決裁境界または共有schema所有を先に固定するため。 |
| 11 | 実装 unit | `U-MR-TREATMENT-PLAN` | なし | なし | 所見 `MRD-02`, `MRD-03`, `MRD-04`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 12 | 実装 unit | `U-X01-LSTEP-LINE-CUSTOMER` | なし | なし | 所見 `G2A-01`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 13 | 実装 unit | `U-X01-MR-PRESCRIPTION` | なし | なし | 所見 `MRC-01`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 14 | 実装 unit | `U-X02-LSTEP-TAG-MAPPING` | なし | なし | 所見 `G2C-02`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 15 | 実装 unit | `U-X02-LSTEP-TRIGGER-PRIORITY` | なし | なし | 所見 `LSA-13`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 16 | 実装 unit | `U-X02-PET-OWNER-FREETEXT` | なし | なし | 所見 `POC-13`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 17 | 実装 unit | `U-X02-RESERVATION-SETTINGS` | なし | なし | 所見 `RSV-04`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 18 | 実装 unit | `U-X03-PET-SPECIES-AUDIT` | なし | `DEC-31` | 所見 `POC-07`。決裁境界または共有schema所有を先に固定するため。 |
| 19 | 実装 unit | `U-X03-STAFF-ASSIGNMENT-AUDIT` | なし | なし | 所見 `AUS-01`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 20 | 実装 unit | `U-X05-MR-EXAMTYPE` | なし | なし | 所見 `MRB-07`, `MRB-08`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 21 | 実装 unit | `U-X05-PET-UPDATE` | なし | なし | 所見 `POC-03`。外部依存のないroot unitとして、後続の業務境界を早期に固定するため。 |
| 22 | 実装 unit | `BE-X06-BIL-CAMPAIGN-01` | なし | なし | 所見 `BIL-03`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 23 | 実装 unit | `BE-X06-LSTEP-SETTINGS-01` | なし | なし | 所見 `LSA-06`, `G2B-02`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 24 | 実装 unit | `BE-X06-MEDICAL-ATOMIC-01` | なし | なし | 所見 `MRC-05`, `MRC-12`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 25 | 実装 unit | `BE-X06-RSV-CANCEL-01` | なし | なし | 所見 `RSV-06`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 26 | 実装 unit | `BE-X07-BODY-01` | `U-SCHEMA-BARRIER` | `DEC-37` | 所見 `INF-02`, `POC-12`, `TRM-03`, `TRM-05`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 27 | 実装 unit | `SOLO-24` | `U-SCHEMA-BARRIER` | `DEC-29` | 所見 `MDL-06`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 28 | 実装 unit | `U-X01X02-INVENTORY` | `U-SCHEMA-BARRIER` | なし | 所見 `G2P-02`, `G2P-03`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 29 | 実装 unit | `U-X01X03-MANUALARTICLE` | `U-SCHEMA-BARRIER` | なし | 所見 `TRM-01`, `TRM-02`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 30 | 実装 unit | `U-X01X03-MR-CARE` | `U-SCHEMA-BARRIER` | なし | 所見 `MRA-01`, `MRA-02`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 31 | 実装 unit | `U-X01X05-CLINIC` | `SOLO-32`, `U-SCHEMA-BARRIER` | なし | 所見 `POC-02`, `POC-05`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 32 | 実装 unit | `U-X02-LSTEP-SHARED-FILE` | `U-SCHEMA-BARRIER` | なし | 所見 `LSB-06`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 33 | 実装 unit | `U-X02-LSTEP-TAG-CONFIG` | `U-SCHEMA-BARRIER` | なし | 所見 `LSA-14`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 34 | 実装 unit | `U-X02-MR-CONSULTATION` | `U-SCHEMA-BARRIER` | なし | 所見 `MRA-03`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 35 | 実装 unit | `U-X02-STAFF-TYPE` | `U-SCHEMA-BARRIER` | なし | 所見 `AUS-03`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 36 | 実装 unit | `U-X02X03X05-MR-HOSPITALIZATION` | `U-SCHEMA-BARRIER` | なし | 所見 `MRB-02`, `MRB-03`, `MRB-05`, `MRB-06`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 37 | 実装 unit | `U-X03-CSVIMPORT-GUARD` | `U-SCHEMA-BARRIER` | なし | 所見 `TRM-07`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 38 | 実装 unit | `U-X04-RESERVATION-AUTODELEGATE` | なし | `DEC-35` | 所見 `RSV-09`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 39 | 実装 unit | `U-X04X05-LSTEP-DELIVERY` | `U-LSTEP-OPTOUT`, `U-SCHEMA-BARRIER` | `DEC-35` | 所見 `LSA-12`, `LSA-15`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 40 | 実装 unit | `U-X05-OWNER-PHONE` | `U-SCHEMA-BARRIER` | なし | 所見 `POC-06`。atomicity・tenant isolation・監査境界を先に確立するため。 |
| 41 | 実装 unit | `BE-X08-LSTEP-CONNECTION-01` | `BE-X06-LSTEP-SETTINGS-01` | なし | 所見 `LSA-01`, `LSA-08`。shared schema/repository owner確定後に公開contractを収束させるため。 |
| 42 | 実装 unit | `BE-X08-LSTEP-SEND-01` | なし | なし | 所見 `LSA-09`。shared schema/repository owner確定後に公開contractを収束させるため。 |
| 43 | 実装 unit | `BE-X09-PET-ENUMS-01` | なし | なし | 所見 `POC-11`。shared schema/repository owner確定後に公開contractを収束させるため。 |
| 44 | 実装 unit | `BE-X10-AUTH-RESPONSE-01` | なし | なし | 所見 `AUS-09`。shared schema/repository owner確定後に公開contractを収束させるため。 |
| 45 | 実装 unit | `SOLO-09` | `U-X02-LSTEP-TAG-CONFIG` | `DEC-30` | 所見 `LSA-04`。shared schema/repository owner確定後に公開contractを収束させるため。 |
| 46 | 実装 unit | `U-TRIMMING-SERVICE` | `BE-X07-BODY-01` | なし | 所見 `TRM-04`, `TRM-06`。shared schema/repository owner確定後に公開contractを収束させるため。 |
| 47 | 実装 unit | `U-X01X05-RESERVATION` | `BE-X06-RSV-CANCEL-01`, `U-SCHEMA-BARRIER` | なし | 所見 `RSV-02`, `RSV-03`, `RSV-07`。shared schema/repository owner確定後に公開contractを収束させるため。 |
| 48 | 実装 unit | `U-X02-CLINIC-CONTACT` | なし | なし | 所見 `POC-17`。shared schema/repository owner確定後に公開contractを収束させるため。 |
| 49 | 実装 unit | `U-X02-LSTEP-AGGREGATION` | なし | なし | 所見 `G2A-05`。shared schema/repository owner確定後に公開contractを収束させるため。 |
| 50 | 実装 unit | `U-X02-MR-LAB-IMPORT` | `BE-X06-MEDICAL-ATOMIC-01` | なし | 所見 `MRC-08`。shared schema/repository owner確定後に公開contractを収束させるため。 |
| 51 | 実装 unit | `U-X05-MR-MEDICINE-MASTERS` | `U-X01X02-INVENTORY` | なし | 所見 `MRC-02`, `MRC-03`, `MRC-07`。shared schema/repository owner確定後に公開contractを収束させるため。 |
| 52 | 実装 unit | `BE-X09-CLOSING-01` | なし | なし | 所見 `POC-15`, `POC-16`。public boundary/shared service確定後にerror contractを統一するため。 |
| 53 | 実装 unit | `BE-X09-MEDICAL-DIAGNOSIS-01` | なし | なし | 所見 `MRC-14`。public boundary/shared service確定後にerror contractを統一するため。 |
| 54 | 実装 unit | `BE-X09-PET-PATCH-01` | なし | なし | 所見 `POC-14`。public boundary/shared service確定後にerror contractを統一するため。 |
| 55 | 実装 unit | `U-X01X03X04-LSTEP-LIFECYCLE` | `BE-X08-LSTEP-CONNECTION-01`, `BE-X06-LSTEP-SETTINGS-01` | `DEC-35` | 所見 `LSA-05`, `LSB-02`, `LSB-03`, `LSB-04`。public boundary/shared service確定後にerror contractを統一するため。 |
| 56 | 実装 unit | `U-X04-AUDIT-MARSHAL` | なし | `DEC-35` | 所見 `INF-06`。public boundary/shared service確定後にerror contractを統一するため。 |
| 57 | 実装 unit | `U-X04-COVERAGE-RATCHET` | なし | `DEC-35` | 所見 `CMD-03`。public boundary/shared service確定後にerror contractを統一するため。 |
| 58 | 実装 unit | `U-X04-LSTEP-BATCH` | なし | `DEC-35` | 所見 `LSA-03`。public boundary/shared service確定後にerror contractを統一するため。 |
| 59 | 実装 unit | `U-X04-LSTEP-HEALTH-REMOVE` | なし | `DEC-35` | 所見 `G2B-01`。public boundary/shared service確定後にerror contractを統一するため。 |
| 60 | 実装 unit | `U-X04-LSTEP-MIGRATE` | なし | `DEC-35` | 所見 `CMD-07`。public boundary/shared service確定後にerror contractを統一するため。 |
| 61 | 実装 unit | `U-X04-LSTEP-OWNER-TAGS` | なし | `DEC-35` | 所見 `LSA-10`。public boundary/shared service確定後にerror contractを統一するため。 |
| 62 | 実装 unit | `U-X04-LSTEP-STALE-TAGS` | なし | `DEC-35` | 所見 `LSA-11`。public boundary/shared service確定後にerror contractを統一するため。 |
| 63 | 実装 unit | `SOLO-01` | なし | なし | 所見 `AUS-04`, `AUS-05`, `G2F-10`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 64 | 実装 unit | `SOLO-02` | なし | `DEC-33` | 所見 `BIL-01`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 65 | 実装 unit | `SOLO-03` | なし | なし | 所見 `BIL-02`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 66 | 実装 unit | `SOLO-05` | `U-X05-PET-UPDATE`, `BE-X07-BODY-01` | なし | 所見 `CMD-04`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 67 | 実装 unit | `SOLO-06` | なし | なし | 所見 `CMD-06`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 68 | 実装 unit | `SOLO-08` | `U-X02-PET-OWNER-FREETEXT` | なし | 所見 `INF-03`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 69 | 実装 unit | `SOLO-10` | なし | なし | 所見 `LSA-07`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 70 | 実装 unit | `SOLO-12` | `U-LSTEP-OPTOUT` | なし | 所見 `G2C-01`, `G2F-01`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 71 | 実装 unit | `SOLO-13` | なし | なし | 所見 `G2C-04`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 72 | 実装 unit | `SOLO-14` | `U-X04-LSTEP-HEALTH-REMOVE` | なし | 所見 `G2F-02`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 73 | 実装 unit | `SOLO-15` | `U-X04-LSTEP-BATCH`, `U-X05-OWNER-PHONE` | なし | 所見 `G2F-03`, `G2F-04`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 74 | 実装 unit | `SOLO-17` | なし | なし | 所見 `G2F-06`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 75 | 実装 unit | `SOLO-19` | `U-X01X05-RESERVATION` | なし | 所見 `G2F-08`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 76 | 実装 unit | `SOLO-20` | なし | なし | 所見 `G2F-09`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 77 | 実装 unit | `SOLO-21` | `U-X05-MR-EXAMTYPE`, `U-X01X02-INVENTORY` | なし | 所見 `G2F-11`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 78 | 実装 unit | `SOLO-22` | なし | `DEC-34` | 所見 `INF-04`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 79 | 実装 unit | `SOLO-23` | なし | なし | 所見 `MDL-05`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 80 | 実装 unit | `SOLO-25` | なし | なし | 所見 `MRA-04`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 81 | 実装 unit | `SOLO-29` | なし | なし | 所見 `MRC-09`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 82 | 実装 unit | `SOLO-30` | なし | なし | 所見 `MRD-01`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 83 | 実装 unit | `SOLO-34` | `BE-X06-LSTEP-SETTINGS-01`, `BE-X09-CLOSING-01` | なし | 所見 `POC-09`, `POC-10`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 84 | 実装 unit | `SOLO-36` | なし | `DEC-39` | 所見 `TRM-09`。先行境界に依存する低結合の継続作業として最後に行うため。 |
| 85 | 実装 unit | `U-X04-MR-SUBRECORDS` | `BE-X09-MEDICAL-DIAGNOSIS-01` | `DEC-32`, `DEC-35` | 所見 `MRC-04`。先行境界に依存する低結合の継続作業として最後に行うため。 |

### Implementation-unit ledger

##### BE-X09-ESTIMATE-TAX-01

- 含む所見 ID: MDL-01
- 所有パス (8):
  - `backend/internal/model/estimate.go`
  - `backend/internal/model/estimate_test.go`
  - `backend/internal/billing/estimate_response.go`
  - `backend/internal/billing/estimate_response_test.go`
  - `backend/internal/billing/estimate_service.go`
  - `backend/internal/billing/estimate_service_test.go`
  - `backend/internal/billing/estimate_request.go`
  - `backend/internal/billing/estimate_request_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-SCHEMA-BARRIER。
- 再実測: `git grep -n -F -e '/Users/minoru/.claude/rules/ecc/common/coding-style.md:26' -e 'subtotal := max(UnitPrice*Quantity - DiscountAmount, 0)' -e 'subtotal := UnitPrice*Quantity' -e 'Users' -e 'Avoid' -e 'BillingItem' HEAD -- backend/internal/model/estimate.go backend/internal/model/estimate_test.go backend/internal/billing/estimate_response.go backend/internal/billing/estimate_response_test.go backend/internal/billing/estimate_service.go backend/internal/billing/estimate_service_test.go backend/internal/billing/estimate_request.go backend/internal/billing/estimate_request_test.go`
- 検証: `docker compose exec backend go test ./internal/model/... ./internal/billing/... -run 'Estimate|CalculateTaxAmount'`
- 既知台帳: none
- Size: L (8/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: f9c51cef52269e6bc8f0ad4ddbd516fb475c48c5

##### SOLO-04

- 含む所見 ID: CMD-02, CMD-05
- 所有パス (7):
  - `backend/cmd/api/base_routes.go`
  - `backend/cmd/api/main.go`
  - `backend/cmd/api/batch_scheduler.go`
  - `backend/internal/scheduler/handler.go`
  - `backend/internal/config/config.go`
  - `backend/cmd/api/batch_scheduler_test.go`
  - `backend/cmd/api/route_composition_smoke_test.go`
- 依存 / 決裁: 決裁=DEC-36。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:134' -e 'docker-compose.yml:40' -e 'handler.run' -e 'RunRequest.validate' -e 'Engine' -e 'Recovery' HEAD -- backend/cmd/api/base_routes.go backend/cmd/api/main.go backend/cmd/api/batch_scheduler.go backend/internal/scheduler/handler.go backend/internal/config/config.go backend/cmd/api/batch_scheduler_test.go backend/cmd/api/route_composition_smoke_test.go`
- 検証: `docker compose exec backend go test ./cmd/api/ ./internal/scheduler/ -run "Scheduler|Route|Upload"`
- 既知台帳: none
- Size: L (7/8 files)
- Status: 完了 ｜ 担当レーン: LANE-2 ｜ 完了 commit: c03c792f1594cf698d8e3be9702a82fcf6c6515c

##### SOLO-11

- 含む所見 ID: G2B-03
- 所有パス (4):
  - `backend/internal/lstep/lstep_delivery_trigger_suppression.go`
  - `backend/internal/lstep/lstep_batch_delivery.go`
  - `backend/internal/lstep/lstep_delivery_trigger_suppression_test.go`
  - `backend/internal/lstep/lstep_delivery_trigger_service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e 'CODING_RULES' -e 'CLAUDE' -e 'Remove' -e 'lstep_delivery_trigger_suppression' -e 'lstep_batch_delivery' -e 'suppressed_by_priority' HEAD -- backend/internal/lstep/lstep_delivery_trigger_suppression.go backend/internal/lstep/lstep_batch_delivery.go backend/internal/lstep/lstep_delivery_trigger_suppression_test.go backend/internal/lstep/lstep_delivery_trigger_service_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/ -run "Suppression|Priority|DeliveryTrigger"`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: e6e92b39fb12015c02537bf7be119256b9bfd254

##### SOLO-16

- 含む所見 ID: G2F-05
- 所有パス (4):
  - `backend/internal/lstep/line_customer_repository.go`
  - `backend/internal/lstep/line_customer_handler.go`
  - `backend/internal/lstep/line_customer_repository_test.go`
  - `backend/internal/lstep/line_customer_handler_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/go-gin-backend-review.md:67' -e 'EXPLAIN' -e 'Owner' -e 'Preload' -e 'ParsePaginationWithMax' -e 'line_customer_repository' HEAD -- backend/internal/lstep/line_customer_repository.go backend/internal/lstep/line_customer_handler.go backend/internal/lstep/line_customer_repository_test.go backend/internal/lstep/line_customer_handler_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/ -run "LineCustomer"`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: 8551773b24b824b61da61852e93d5b61ede567e3

##### SOLO-18

- 含む所見 ID: G2F-07
- 所有パス (2):
  - `backend/internal/reservation/appointment_admin_repository.go`
  - `backend/internal/reservation/appointment_admin_repository_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/go-gin-backend-review.md:67' -e 'EXPLAIN' -e 'appointment_admin_repository' HEAD -- backend/internal/reservation/appointment_admin_repository.go backend/internal/reservation/appointment_admin_repository_test.go`
- 検証: `docker compose exec backend go test ./internal/reservation/ -run "AppointmentAdminRepository"`
- 既知台帳: none
- Size: S (2/8 files)
- Status: 完了 ｜ 担当レーン: LANE-2 ｜ 完了 commit: 0a52c5509e9916447e7dc020df11e585ec062f0d

##### SOLO-26

- 含む所見 ID: RSV-08, TRM-08
- 所有パス (6):
  - `backend/internal/reservation/reservation_schedule_repository.go`
  - `backend/internal/trimming/trimming_repository.go`
  - `backend/internal/lintscan/preload_clinic_scope_lint_test.go`
  - `backend/internal/lintscan/grandchild_parent_clinic_correlation_lint_test.go`
  - `backend/internal/reservation/reservation_schedule_repository_test.go`
  - `backend/internal/trimming/trimming_repository_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/backend-application-invariants.md:15' -e 'FindAllBreaksByEntryIDs' -e 'FindAllBreaksByEntryID' -e 'Where("shift_entry_id IN ?")' -e 'Where' -e 'EXISTS' HEAD -- backend/internal/reservation/reservation_schedule_repository.go backend/internal/trimming/trimming_repository.go backend/internal/lintscan/preload_clinic_scope_lint_test.go backend/internal/lintscan/grandchild_parent_clinic_correlation_lint_test.go backend/internal/reservation/reservation_schedule_repository_test.go backend/internal/trimming/trimming_repository_test.go`
- 検証: `docker compose exec backend go test ./internal/medicalrecord/ ./internal/reservation/ ./internal/trimming/ ./internal/lintscan/ -run "Hospitalization|Schedule|Trimming|Preload|Grandchild"`
- 既知台帳: RSV-08=SEC-SWEEP-02 (table-only known pointer); TRM-08=SEC-SWEEP-02 / DEC-23
- Size: M (6/8 files)
- Status: 完了 ｜ 担当レーン: LANE-3 ｜ 完了 commit: 0c44f49fdfe5764206e456b30372a70e651006ff

##### SOLO-32

- 含む所見 ID: POC-01
- 所有パス (5):
  - `backend/internal/clinic/clinic_holiday_handler.go`
  - `backend/internal/clinic/closing_settings_handler.go`
  - `backend/internal/clinic/clinic_service.go`
  - `backend/internal/clinic/clinic_holiday_handler_test.go`
  - `backend/internal/clinic/closing_settings_handler_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/backend-application-invariants.md:22' -e 'shifts:create OR closing-settings:create' -e 'SetClinicHoliday' -e 'DeleteClinicHoliday' -e 'POST' -e 'ResourceShifts' HEAD -- backend/internal/clinic/clinic_holiday_handler.go backend/internal/clinic/closing_settings_handler.go backend/internal/clinic/clinic_service.go backend/internal/clinic/clinic_holiday_handler_test.go backend/internal/clinic/closing_settings_handler_test.go`
- 検証: `docker compose exec backend go test ./internal/clinic/ -run "Holiday|ClosingSettings"`
- 既知台帳: none
- Size: M (5/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### SOLO-33

- 含む所見 ID: POC-08
- 所有パス (4):
  - `backend/internal/owner/http_routes.go`
  - `backend/internal/owner/http_owner.go`
  - `backend/internal/owner/http_routes_test.go`
  - `backend/docs/api.yaml`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e 'protected.Group' -e 'httpapi.ExtractClinicID' -e '/clinics/:clinic_id/owners' -e 'CODING_RULES' -e 'Group' -e 'PATCH' HEAD -- backend/internal/owner/http_routes.go backend/internal/owner/http_owner.go backend/internal/owner/http_routes_test.go backend/docs/api.yaml`
- 検証: `docker compose exec backend go test ./internal/owner/ ./internal/apicontract/ -run "Routes|OpenAPI"`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: 5f29b793ab2b0826bad80f50700cf9a3c2794d1b

##### U-LSTEP-OPTOUT

- 含む所見 ID: LSA-02, LSB-01
- 所有パス (6):
  - `backend/internal/lstep/lstep_delivery_trigger_state.go`
  - `backend/internal/lstep/lstep_delivery_trigger_batch.go`
  - `backend/internal/lstep/lstep_delivery_trigger_state_test.go`
  - `backend/internal/lstep/lstep_delivery_trigger_batch_test.go`
  - `backend/internal/lstep/lstep_tag_sync_pet_exclusion.go`
  - `backend/internal/lstep/lstep_tag_sync_pet_exclusion_test.go`
- 依存 / 決裁: 決裁=DEC-38。先行 unit=なし。
- 再実測: `git grep -n -F -e 'checkExclusion' -e 'owner.DeliveryExcluded' -e 'LineUserID' -e 'EXCL_配信停止' -e 'owner.LstepOptOut' -e 'SyncExclusionTags' HEAD -- backend/internal/lstep/lstep_delivery_trigger_state.go backend/internal/lstep/lstep_delivery_trigger_batch.go backend/internal/lstep/lstep_delivery_trigger_state_test.go backend/internal/lstep/lstep_delivery_trigger_batch_test.go backend/internal/lstep/lstep_tag_sync_pet_exclusion.go backend/internal/lstep/lstep_tag_sync_pet_exclusion_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/... -run 'DeliveryTrigger|PetExclusion|OptOut'`
- 既知台帳: none
- Size: M (6/8 files)
- Status: 完了 ｜ 担当レーン: LANE-1 ｜ 完了 commit: 34b3d72ec744287843c0ce0cf7fe6bc9d2e69877

##### U-MR-TREATMENT-PLAN

- 含む所見 ID: MRD-02, MRD-03, MRD-04
- 所有パス (8):
  - `backend/internal/medicalrecord/treatment_plan_request.go`
  - `backend/internal/medicalrecord/treatment_plan_service.go`
  - `backend/internal/medicalrecord/treatment_plan_repository.go`
  - `backend/internal/medicalrecord/treatment_plan_handler.go`
  - `backend/internal/medicalrecord/treatment_plan_request_test.go`
  - `backend/internal/medicalrecord/treatment_plan_service_test.go`
  - `backend/internal/medicalrecord/treatment_plan_repository_test.go`
  - `backend/internal/medicalrecord/treatment_plan_handler_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e 'treatmentService.Update' -e 'WithTx' -e 'FindByID' -e 'treatmentPlanService' -e 'vitalService.Update' -e 'vital_repository.go:52-53' HEAD -- backend/internal/medicalrecord/treatment_plan_request.go backend/internal/medicalrecord/treatment_plan_service.go backend/internal/medicalrecord/treatment_plan_repository.go backend/internal/medicalrecord/treatment_plan_handler.go backend/internal/medicalrecord/treatment_plan_request_test.go backend/internal/medicalrecord/treatment_plan_service_test.go backend/internal/medicalrecord/treatment_plan_repository_test.go backend/internal/medicalrecord/treatment_plan_handler_test.go`
- 検証: `docker compose exec backend go test ./internal/medicalrecord/... -run 'TreatmentPlan'`
- 既知台帳: none
- Size: L (8/8 files)
- Status: 完了 ｜ 担当レーン: LANE-3 ｜ 完了 commit: 9f3390d08d8e4378b62422a71da09962c03ba7f6

##### 共有所有 barrier U-SCHEMA-BARRIER

- 含む所見 ID: none (shared-owner barrier; no live-ID ownership)
- 所有パス (1):
  - `backend/migrations/001_init.sql`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- Barrier再実測: `git grep -n -E 'CHECK|FOREIGN KEY|UNIQUE|clinic_id|CONSTRAINT' HEAD -- backend/migrations/001_init.sql`
- Barrier検証: `docker compose exec backend go test ./internal/model/... -run 'SchemaDrift' && docker compose exec backend go test ./internal/lintscan/... -run 'Migration|DBOrTx'`
- 既知台帳: TASK-445 / DEC-28 supports MDL-06; BUG-466 supports G2P-03
- Size: S (1/8 files)
- Status: 完了 ｜ 担当レーン: LANE-1 ｜ 完了 commit: b1785b5bb53f5259cbb936364508ea1ec7fc32e7

##### U-X01-LSTEP-LINE-CUSTOMER

- 含む所見 ID: G2A-01
- 所有パス (2):
  - `backend/internal/lstep/line_customer_service.go`
  - `backend/internal/lstep/line_customer_service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e 'UpdateOwnerLink' -e 'FindByID' -e 'CODING_RULES' -e 'WithTx' -e 'line_customer_service' HEAD -- backend/internal/lstep/line_customer_service.go backend/internal/lstep/line_customer_service_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/... -run 'LineCustomer'`
- 既知台帳: none
- Size: S (2/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: 5a1cef17488690719b3b36c7fa9eaca055e07826

##### U-X01-MR-PRESCRIPTION

- 含む所見 ID: MRC-01
- 所有パス (4):
  - `backend/internal/medicalrecord/prescription_repository.go`
  - `backend/internal/medicalrecord/prescription_service.go`
  - `backend/internal/medicalrecord/prescription_repository_test.go`
  - `backend/internal/medicalrecord/prescription_service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e 's.repo.FindByID(ctx, ...)' -e 'r.db.WithContext(ctx)' -e 'repo.FindByID' -e 'prescriptionRepository.FindByID' -e 'db.WithContext' -e 'handler' HEAD -- backend/internal/medicalrecord/prescription_repository.go backend/internal/medicalrecord/prescription_service.go backend/internal/medicalrecord/prescription_repository_test.go backend/internal/medicalrecord/prescription_service_test.go`
- 検証: `docker compose exec backend go test ./internal/medicalrecord/... -run 'Prescription'`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 完了 ｜ 担当レーン: LANE-3 ｜ 完了 commit: b434d77aa263c7d8a71bf504e802b5ef53a2069b

##### U-X02-LSTEP-TAG-MAPPING

- 含む所見 ID: G2C-02
- 所有パス (3):
  - `backend/internal/lstep/lstep_tag_code_mapping_request.go`
  - `backend/internal/lstep/lstep_tag_code_mapping_service.go`
  - `backend/internal/lstep/lstep_tag_code_mapping_service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:151' -e 'lstep_tag_code_mapping_request' -e 'lstep_tag_code_mapping_service' HEAD -- backend/internal/lstep/lstep_tag_code_mapping_request.go backend/internal/lstep/lstep_tag_code_mapping_service.go backend/internal/lstep/lstep_tag_code_mapping_service_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/... -run 'TagCodeMapping'`
- 既知台帳: none
- Size: M (3/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: 404880803c9ff9ecee568b3ad851156a7f61775e

##### U-X02-LSTEP-TRIGGER-PRIORITY

- 含む所見 ID: LSA-13
- 所有パス (5):
  - `backend/internal/lstep/lstep_trigger_priority_handler.go`
  - `backend/internal/lstep/lstep_trigger_priority_request.go`
  - `backend/internal/lstep/lstep_trigger_priority_service.go`
  - `backend/internal/lstep/lstep_trigger_priority_handler_test.go`
  - `backend/internal/lstep/lstep_trigger_priority_service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:151' -e 'updateTriggerPriorityItemRequest.TriggerType' -e 'binding:"required"' -e 'oneof' -e 'priority < 1' -e 'UpsertBatch' HEAD -- backend/internal/lstep/lstep_trigger_priority_handler.go backend/internal/lstep/lstep_trigger_priority_request.go backend/internal/lstep/lstep_trigger_priority_service.go backend/internal/lstep/lstep_trigger_priority_handler_test.go backend/internal/lstep/lstep_trigger_priority_service_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/... -run 'TriggerPriority'`
- 既知台帳: none
- Size: M (5/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: f224c289af6137744234896baf850df38de4a761

##### U-X02-PET-OWNER-FREETEXT

- 含む所見 ID: POC-13
- 所有パス (4):
  - `backend/internal/owner/http_request.go`
  - `backend/internal/pet/pet_request.go`
  - `backend/internal/owner/http_request_test.go`
  - `backend/internal/pet/pet_request_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:151' -e 'varchar' -e 'NameKana' -e 'Breed' -e 'Color' -e 'Food' HEAD -- backend/internal/owner/http_request.go backend/internal/pet/pet_request.go backend/internal/owner/http_request_test.go backend/internal/pet/pet_request_test.go`
- 検証: `docker compose exec backend go test ./internal/owner/... ./internal/pet/... -run 'Request|Validation'`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 完了 ｜ 担当レーン: LANE-1 ｜ 完了 commit: 1b83757246f9831071e27ddef1465bc273647fb1

##### U-X02-RESERVATION-SETTINGS

- 含む所見 ID: RSV-04
- 所有パス (5):
  - `backend/internal/reservation/available_dates.go`
  - `backend/internal/reservation/liff_service_availability.go`
  - `backend/internal/reservation/line_reservation_setting_request.go`
  - `backend/internal/reservation/available_dates_test.go`
  - `backend/internal/reservation/line_reservation_setting_request_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:151' -e 'upsertLineReservationSettingRequest' -e 'booking_window_max_days' -e 'make([]AvailableDateResult, 0, input.Settings.BookingWindowMaxDays)' -e 'status' -e 'time_slot_mode' HEAD -- backend/internal/reservation/available_dates.go backend/internal/reservation/liff_service_availability.go backend/internal/reservation/line_reservation_setting_request.go backend/internal/reservation/available_dates_test.go backend/internal/reservation/line_reservation_setting_request_test.go`
- 検証: `docker compose exec backend go test ./internal/reservation/... -run 'LineReservationSettingRequest|AvailableDates'`
- 既知台帳: none
- Size: M (5/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### U-X03-PET-SPECIES-AUDIT

- 含む所見 ID: POC-07
- 所有パス (7):
  - `backend/internal/pet/animal_species_handler.go`
  - `backend/internal/pet/animal_species_repository.go`
  - `backend/internal/pet/animal_species_service.go`
  - `backend/internal/pet/ports.go`
  - `backend/internal/pet/animal_species_handler_test.go`
  - `backend/internal/pet/animal_species_repository_test.go`
  - `backend/internal/pet/animal_species_service_test.go`
- 依存 / 決裁: 決裁=DEC-31。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/backend-application-invariants.md:31' -e 'ClinicID' -e 'PetOwnerAuditLogger' -e 'DELETE' -e 'RESTRICT' -e 'CountUsageByAnimalSpeciesID' HEAD -- backend/internal/pet/animal_species_handler.go backend/internal/pet/animal_species_repository.go backend/internal/pet/animal_species_service.go backend/internal/pet/ports.go backend/internal/pet/animal_species_handler_test.go backend/internal/pet/animal_species_repository_test.go backend/internal/pet/animal_species_service_test.go`
- 検証: `docker compose exec backend go test ./internal/pet/... -run 'AnimalSpecies.*Audit|AnimalSpecies'`
- 既知台帳: none
- Size: L (7/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### U-X03-STAFF-ASSIGNMENT-AUDIT

- 含む所見 ID: AUS-01
- 所有パス (6):
  - `backend/internal/staff/staff_clinic_assignment_service.go`
  - `backend/internal/staff/handler.go`
  - `backend/internal/staff/staff_handler.go`
  - `backend/internal/staff/staff_service_permissions.go`
  - `backend/internal/staff/staff_clinic_assignment_service_test.go`
  - `backend/internal/staff/staff_service_permissions_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/backend-application-invariants.md:31' -e 'assignmentRepo.Delete' -e 'perm' -e 'SetClinicAssignments' -e 'RestoreOrCreate' -e 'AuthorizedClinicIDs' HEAD -- backend/internal/staff/staff_clinic_assignment_service.go backend/internal/staff/handler.go backend/internal/staff/staff_handler.go backend/internal/staff/staff_service_permissions.go backend/internal/staff/staff_clinic_assignment_service_test.go backend/internal/staff/staff_service_permissions_test.go`
- 検証: `docker compose exec backend go test ./internal/staff/... -run 'ClinicAssignment|PermissionAssignment|Audit'`
- 既知台帳: none
- Size: M (6/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: 0ea639b16c64300591531a6c0ed6d29de5b7180f

##### U-X05-MR-EXAMTYPE

- 含む所見 ID: MRB-07, MRB-08
- 所有パス (5):
  - `backend/internal/medicalrecord/exam_type_field.go`
  - `backend/internal/medicalrecord/exam_type_service.go`
  - `backend/internal/medicalrecord/exam_type_repository.go`
  - `backend/internal/medicalrecord/exam_type_service_test.go`
  - `backend/internal/medicalrecord/exam_type_repository_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:167' -e 'apperrors.WrapInvalidInput' -e 'NewExamTypeService' -e 'WrapInvalidInput' -e 'WrapInternalServerError' -e 'Transactor' HEAD -- backend/internal/medicalrecord/exam_type_field.go backend/internal/medicalrecord/exam_type_service.go backend/internal/medicalrecord/exam_type_repository.go backend/internal/medicalrecord/exam_type_service_test.go backend/internal/medicalrecord/exam_type_repository_test.go`
- 検証: `docker compose exec backend go test ./internal/medicalrecord/... -run 'ExamType' && docker compose exec backend go test ./internal/lintscan/ -run DBOrTx`
- 既知台帳: none
- Size: M (5/8 files)
- Status: 完了 ｜ 担当レーン: LANE-3 ｜ 完了 commit: 15e3c9184131c56351f7f885323543760fc87f0f

##### U-X05-PET-UPDATE

- 含む所見 ID: POC-03
- 所有パス (6):
  - `backend/internal/pet/owner_registration.go`
  - `backend/internal/pet/repository.go`
  - `backend/internal/pet/service.go`
  - `backend/internal/pet/owner_registration_test.go`
  - `backend/internal/pet/repository_test.go`
  - `backend/internal/pet/service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e 'FOR UPDATE' -e 'FOR SHARE' -e 'pg_advisory_xact_lock' -e 'CODING_RULES' -e 'UPDATE' -e 'SHARE' HEAD -- backend/internal/pet/owner_registration.go backend/internal/pet/repository.go backend/internal/pet/service.go backend/internal/pet/owner_registration_test.go backend/internal/pet/repository_test.go backend/internal/pet/service_test.go`
- 検証: `docker compose exec backend go test ./internal/pet/... -run 'Update.*Owner|Update.*Insurance|UpdateAndFind'`
- 既知台帳: none
- Size: M (6/8 files)
- Status: 完了 ｜ 担当レーン: LANE-2 ｜ 完了 commit: e7bc7aceb90f0876f1231e1c6eb66b4186c66d56

##### BE-X06-BIL-CAMPAIGN-01

- 含む所見 ID: BIL-03
- 所有パス (5):
  - `backend/internal/billing/campaign_service.go`
  - `backend/internal/billing/campaign_repository.go`
  - `backend/internal/billing/campaign_service_test.go`
  - `backend/internal/billing/campaign_cross_tenant_master_fk_write_test.go`
  - `backend/cmd/api/composition_billing_services.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e 'campaignService' -e 's.repo.Update' -e 'r.db.WithContext(ctx)' -e 's.repo.ReplaceTargets' -e 'persistence.DBOrTx(ctx, r.db)' -e 'FindApplicableForItem' HEAD -- backend/internal/billing/campaign_service.go backend/internal/billing/campaign_repository.go backend/internal/billing/campaign_service_test.go backend/internal/billing/campaign_cross_tenant_master_fk_write_test.go backend/cmd/api/composition_billing_services.go`
- 検証: `docker compose exec backend go test ./internal/billing/... ./cmd/api/... -run 'Campaign|BillingComposition' && docker compose exec backend go test ./internal/lintscan/ -run DBOrTx`
- 既知台帳: none
- Size: M (5/8 files)
- Status: 完了 ｜ 担当レーン: LANE-1 ｜ 完了 commit: 652d3d2073e3aaecb0bd9de0e5ea94cddb38c826

##### BE-X06-LSTEP-SETTINGS-01

- 含む所見 ID: LSA-06, G2B-02
- 所有パス (8):
  - `backend/internal/lstep/lstep_settings_service.go`
  - `backend/internal/lstep/lstep_settings_update.go`
  - `backend/internal/lstep/lstep_settings_repository.go`
  - `backend/internal/lstep/lstep_sync_settings_repository.go`
  - `backend/internal/clinic/clinic_settings_repository.go`
  - `backend/internal/lstep/composition_services.go`
  - `backend/internal/lstep/lstep_settings_service_test.go`
  - `backend/internal/lstep/lstep_settings_update_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e 'UpdateSettings' -e 'updateIntegrationCredentials' -e 'updateSyncEnabled' -e 'updateClinicSyncConfig' -e 'Transactor.WithTx' -e 'persistence.DBOrTx' HEAD -- backend/internal/lstep/lstep_settings_service.go backend/internal/lstep/lstep_settings_update.go backend/internal/lstep/lstep_settings_repository.go backend/internal/lstep/lstep_sync_settings_repository.go backend/internal/clinic/clinic_settings_repository.go backend/internal/lstep/composition_services.go backend/internal/lstep/lstep_settings_service_test.go backend/internal/lstep/lstep_settings_update_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/... ./internal/clinic/... -run 'LstepSettings|ClinicSettings' && docker compose exec backend go test ./internal/lintscan/ -run DBOrTx`
- 既知台帳: none
- Size: L (8/8 files)
- Status: 完了 ｜ 担当レーン: LANE-1 ｜ 完了 commit: 33f4ee2ee0d70add01fee30fd138ab0bc397c4bb

##### BE-X06-MEDICAL-ATOMIC-01

- 含む所見 ID: MRC-05, MRC-12
- 所有パス (8):
  - `backend/cmd/api/composition_medicalrecord_services.go`
  - `backend/internal/medicalrecord/lab_import_examination_service.go`
  - `backend/internal/medicalrecord/lab_import_examination_service_test.go`
  - `backend/internal/medicalrecord/inquiry_repository.go`
  - `backend/internal/medicalrecord/inquiry_repository_test.go`
  - `backend/internal/medicalrecord/inquiry_service.go`
  - `backend/internal/medicalrecord/inquiry_service_test.go`
  - `backend/internal/medicalrecord/cross_tenant_master_fk_write_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e 'LabImportDuplicateCheckerDB.IsDuplicate' -e 'CLAUDE' -e 'Transactor' -e 'ReplaceItemsByExamID' -e 'LabImportDuplicateCheckerDB' -e 'IsDuplicate' HEAD -- backend/cmd/api/composition_medicalrecord_services.go backend/internal/medicalrecord/lab_import_examination_service.go backend/internal/medicalrecord/lab_import_examination_service_test.go backend/internal/medicalrecord/inquiry_repository.go backend/internal/medicalrecord/inquiry_repository_test.go backend/internal/medicalrecord/inquiry_service.go backend/internal/medicalrecord/inquiry_service_test.go backend/internal/medicalrecord/cross_tenant_master_fk_write_test.go`
- 検証: `docker compose exec backend go test ./internal/medicalrecord/... ./cmd/api/... -run 'LabImportExamination|Inquiry|CrossTenantMasterFK' && docker compose exec backend go test ./internal/lintscan/ -run DBOrTx`
- 既知台帳: MRC-12=phase2.html:195
- Size: L (8/8 files)
- Status: 完了 ｜ 担当レーン: LANE-3 ｜ 完了 commit: 807ec2c9103cd8d72a88a4f51f4e8b837c532ae2

##### BE-X06-RSV-CANCEL-01

- 含む所見 ID: RSV-06
- 所有パス (2):
  - `backend/internal/reservation/reservation_service.go`
  - `backend/internal/reservation/appointment_service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/backend-application-invariants.md:36' -e 's.repo.Delete' -e 'WithTx' -e 'NULL' -e 'reservation_service' HEAD -- backend/internal/reservation/reservation_service.go backend/internal/reservation/appointment_service_test.go`
- 検証: `docker compose exec backend go test ./internal/reservation/... -run 'ReservationService.*Update|Cancel' && docker compose exec backend go test ./internal/lintscan/ -run DBOrTx`
- 既知台帳: none
- Size: S (2/8 files)
- Status: 完了 ｜ 担当レーン: LANE-1 ｜ 完了 commit: 970dd987c859646a3965cb6732667cb572ae475b

##### BE-X07-BODY-01

- 含む所見 ID: INF-02, POC-12, TRM-03, TRM-05
- 所有パス (8):
  - `backend/cmd/api/composition_runtime.go`
  - `backend/cmd/api/composition_core_test.go`
  - `backend/internal/middleware/sanitize_null_bytes.go`
  - `backend/internal/middleware/sanitize_null_bytes_test.go`
  - `backend/internal/trimming/trimming_request.go`
  - `backend/internal/trimming/trimming_request_test.go`
  - `backend/internal/manualarticle/request.go`
  - `backend/internal/manualarticle/request_test.go`
- 依存 / 決裁: 決裁=DEC-37。先行 unit=U-SCHEMA-BARRIER。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:180' -e 'sanitizedBodyReader' -e 'c.Request.ContentLength = -1' -e 'http.MaxBytesReader' -e 'POST' -e 'PATCH' HEAD -- backend/cmd/api/composition_runtime.go backend/cmd/api/composition_core_test.go backend/internal/middleware/sanitize_null_bytes.go backend/internal/middleware/sanitize_null_bytes_test.go backend/internal/trimming/trimming_request.go backend/internal/trimming/trimming_request_test.go backend/internal/manualarticle/request.go backend/internal/manualarticle/request_test.go`
- 検証: `docker compose exec backend go test ./internal/middleware/... ./internal/trimming/... ./internal/manualarticle/... ./cmd/api/... -run 'Body|Sanitize|Request|Composition'`
- 既知台帳: none
- Size: L (8/8 files)
- Status: 完了 ｜ 担当レーン: LANE-2 ｜ 完了 commit: dee5b4c73bb4bc8f2c2abe0da3e68bd819451c77

##### SOLO-24

- 含む所見 ID: MDL-06
- 所有パス (3):
  - `backend/internal/model/accounting.go`
  - `backend/internal/billing/accounting_repository.go`
  - `backend/internal/billing/accounting_repository_tx_atomicity_test.go`
- 依存 / 決裁: 決裁=DEC-29。先行 unit=U-SCHEMA-BARRIER。
- 再実測: `git grep -n -F -e 'ClinicID uint64 gorm:"not null"' -e 'Users' -e 'Case' -e 'AnimalHospital' -e 'AnimalEkarte' -e 'Payment' HEAD -- backend/internal/model/accounting.go backend/internal/billing/accounting_repository.go backend/internal/billing/accounting_repository_tx_atomicity_test.go`
- 検証: `docker compose exec backend go test ./internal/billing/ -run "Payment|TxAtomicity"`
- 既知台帳: MDL-06=TASK-445 / DEC-28
- Size: M (3/8 files)
- Status: 対象消失（ALREADY-FIXED: model.Payment.ClinicID + migration 005 + SavePayment lockBillingClinic + PersistsClinicID test） ｜ 担当レーン: LANE-4 ｜ 完了 commit: —

##### U-X01X02-INVENTORY

- 含む所見 ID: G2P-02, G2P-03
- 所有パス (6):
  - `backend/internal/inventory/inventory_request.go`
  - `backend/internal/inventory/repository.go`
  - `backend/internal/inventory/merchandise_item_repository.go`
  - `backend/internal/inventory/inventory_request_test.go`
  - `backend/internal/inventory/medicine_inventory_tx_atomicity_test.go`
  - `backend/internal/inventory/merchandise_item_repository_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-SCHEMA-BARRIER。
- 再実測: `git grep -n -F -e 'CODING_RULES' -e 'UpdateScopedByID' -e 'UpdateAndFind' -e 'merchandise_item_repository' -e 'CHECK' -e 'DecreaseStock' HEAD -- backend/internal/inventory/inventory_request.go backend/internal/inventory/repository.go backend/internal/inventory/merchandise_item_repository.go backend/internal/inventory/inventory_request_test.go backend/internal/inventory/medicine_inventory_tx_atomicity_test.go backend/internal/inventory/merchandise_item_repository_test.go`
- 検証: `docker compose exec backend go test ./internal/inventory/... -run 'Update|Decrease|Quantity' && docker compose exec backend go test ./internal/lintscan/ -run DBOrTx`
- 既知台帳: G2P-02=BUG-465; G2P-03=BUG-466
- Size: M (6/8 files)
- Status: 対象消失（ALREADY-FIXED: G2P-02/G2P-03 は BUG-465/BUG-466 として HEAD で Update 同一tx再取得・binding min=0・DecreaseStock quantity>=? 済み） ｜ 担当レーン: LANE-3 ｜ 完了 commit: —

##### U-X01X03-MANUALARTICLE

- 含む所見 ID: TRM-01, TRM-02
- 所有パス (6):
  - `backend/internal/manualarticle/handler.go`
  - `backend/internal/manualarticle/repository.go`
  - `backend/internal/manualarticle/service.go`
  - `backend/internal/manualarticle/handler_test.go`
  - `backend/internal/manualarticle/repository_test.go`
  - `backend/internal/manualarticle/service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-SCHEMA-BARRIER。
- 再実測: `git grep -n -F -e 'CODING_RULES' -e 'Transaction' -e 'FindByCategoryAndSlug' -e 'RespondError' -e 'UPDATE' -e 'INSERT' HEAD -- backend/internal/manualarticle/handler.go backend/internal/manualarticle/repository.go backend/internal/manualarticle/service.go backend/internal/manualarticle/handler_test.go backend/internal/manualarticle/repository_test.go backend/internal/manualarticle/service_test.go`
- 検証: `docker compose exec backend go test ./internal/manualarticle/...`
- 既知台帳: none
- Size: M (6/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### U-X01X03-MR-CARE

- 含む所見 ID: MRA-01, MRA-02
- 所有パス (6):
  - `backend/internal/medicalrecord/care_plan_item_repository.go`
  - `backend/internal/medicalrecord/care_plan_item_service.go`
  - `backend/internal/medicalrecord/care_plan_item_repository_test.go`
  - `backend/internal/medicalrecord/care_plan_item_service_test.go`
  - `backend/internal/model/audit_log.go`
  - `backend/internal/model/audit_log_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-SCHEMA-BARRIER。
- 再実測: `git grep -n -F -e '.claude/refs/backend-application-invariants.md:31' -e 'perm' -e 'Unscoped' -e 'ResourceHospitalization' -e 'AuditResource' -e 'DELETE' HEAD -- backend/internal/medicalrecord/care_plan_item_repository.go backend/internal/medicalrecord/care_plan_item_service.go backend/internal/medicalrecord/care_plan_item_repository_test.go backend/internal/medicalrecord/care_plan_item_service_test.go backend/internal/model/audit_log.go backend/internal/model/audit_log_test.go`
- 検証: `docker compose exec backend go test ./internal/medicalrecord/... -run 'CarePlanItem' && docker compose exec backend go test ./internal/lintscan/ -run DBOrTx`
- 既知台帳: none
- Size: M (6/8 files)
- Status: 完了 ｜ 担当レーン: LANE-3 ｜ 完了 commit: ab1395886b1a5b0695a2667537817cb6d075bb76

##### U-X01X05-CLINIC

- 含む所見 ID: POC-02, POC-05
- 所有パス (4):
  - `backend/internal/clinic/company_service.go`
  - `backend/internal/clinic/closing_special_period_repository.go`
  - `backend/internal/clinic/clinic_service_test.go`
  - `backend/internal/clinic/closing_special_period_repository_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=SOLO-32, U-SCHEMA-BARRIER。
- 再実測: `git grep -n -F -e 'CODING_RULES' -e 'UPDATE' -e 'UpdateAndFind' -e 'Transaction' -e 'Transactor' -e 'WithTx' HEAD -- backend/internal/clinic/company_service.go backend/internal/clinic/closing_special_period_repository.go backend/internal/clinic/clinic_service_test.go backend/internal/clinic/closing_special_period_repository_test.go`
- 検証: `docker compose exec backend go test ./internal/clinic/... ./internal/pet/... -run 'Clinic|Company|SpecialPeriod|Chronic' && docker compose exec backend go test ./internal/lintscan/ -run DBOrTx`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### U-X02-LSTEP-SHARED-FILE

- 含む所見 ID: LSB-06
- 所有パス (5):
  - `backend/internal/lstep/shared_file_handler.go`
  - `backend/internal/lstep/shared_file_request.go`
  - `backend/internal/model/shared_file.go`
  - `backend/internal/lstep/shared_file_handler_test.go`
  - `backend/internal/lstep/shared_file_request_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-SCHEMA-BARRIER。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:151' -e 'Purpose string \' -e 'varchar' -e 'Purpose' -e 'shared_file_request' -e 'shared_file_handler' HEAD -- backend/internal/lstep/shared_file_handler.go backend/internal/lstep/shared_file_request.go backend/internal/model/shared_file.go backend/internal/lstep/shared_file_handler_test.go backend/internal/lstep/shared_file_request_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/... -run 'SharedFile'`
- 既知台帳: none
- Size: M (5/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: 6d6785c83167ff80c92f8257026c8678597d02f2

##### U-X02-LSTEP-TAG-CONFIG

- 含む所見 ID: LSA-14
- 所有パス (6):
  - `backend/internal/lstep/lstep_tag_config_handler.go`
  - `backend/internal/lstep/lstep_tag_config_request.go`
  - `backend/internal/lstep/lstep_tag_config_service.go`
  - `backend/internal/lstep/lstep_tag_config_handler_test.go`
  - `backend/internal/lstep/lstep_tag_config_request_test.go`
  - `backend/internal/lstep/lstep_tag_config_service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-SCHEMA-BARRIER。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:151' -e 'loadPetDerivedPrefixes' -e 'p.Category == "C2"' -e 'binding:"required"' -e '"c2"' -e '"C2 "' HEAD -- backend/internal/lstep/lstep_tag_config_handler.go backend/internal/lstep/lstep_tag_config_request.go backend/internal/lstep/lstep_tag_config_service.go backend/internal/lstep/lstep_tag_config_handler_test.go backend/internal/lstep/lstep_tag_config_request_test.go backend/internal/lstep/lstep_tag_config_service_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/... -run 'TagConfig|Lifecycle'`
- 既知台帳: none
- Size: M (6/8 files)
- Status: 完了 ｜ 担当レーン: LANE-1 ｜ 完了 commit: ca66820b3784d7419eab2e28c53e54c0d5a1b1de

##### U-X02-MR-CONSULTATION

- 含む所見 ID: MRA-03
- 所有パス (2):
  - `backend/internal/medicalrecord/consultation_request.go`
  - `backend/internal/medicalrecord/consultation_request_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-SCHEMA-BARRIER。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:151' -e 'TaxRate' -e 'TimeCondition' -e 'TaxType' -e 'PATCH' -e 'NULL' HEAD -- backend/internal/medicalrecord/consultation_request.go backend/internal/medicalrecord/consultation_request_test.go`
- 検証: `docker compose exec backend go test ./internal/medicalrecord/... -run 'ConsultationRequest'`
- 既知台帳: none
- Size: S (2/8 files)
- Status: 完了 ｜ 担当レーン: LANE-3 ｜ 完了 commit: f2804a58cdcf694af98e0378f16ea7aaa736ff1c

##### U-X02-STAFF-TYPE

- 含む所見 ID: AUS-03
- 所有パス (8):
  - `backend/internal/staff/staff_request.go`
  - `backend/internal/staff/staff_service_builders.go`
  - `backend/internal/staff/staff_service_core.go`
  - `backend/internal/staff/staff_service_account.go`
  - `backend/internal/staff/staff_request_test.go`
  - `backend/internal/staff/staff_service_builders_test.go`
  - `backend/internal/staff/staff_service_core_test.go`
  - `backend/internal/staff/staff_service_account_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-SCHEMA-BARRIER。
- 再実測: `git grep -n -F -e 'CODING_RULES' -e 'StaffType' -e 'PostgreSQL' -e 'FromGORM' -e 'WrapInvalidInput' -e 'CreateWithAccount' HEAD -- backend/internal/staff/staff_request.go backend/internal/staff/staff_service_builders.go backend/internal/staff/staff_service_core.go backend/internal/staff/staff_service_account.go backend/internal/staff/staff_request_test.go backend/internal/staff/staff_service_builders_test.go backend/internal/staff/staff_service_core_test.go backend/internal/staff/staff_service_account_test.go`
- 検証: `docker compose exec backend go test ./internal/staff/... -run 'StaffType|Create|Update'`
- 既知台帳: none
- Size: L (8/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: ff973b3818594fcd870bba4cd6cdc78c9c916f8b

##### U-X02X03X05-MR-HOSPITALIZATION

- 含む所見 ID: MRB-02, MRB-03, MRB-05, MRB-06
- 所有パス (6):
  - `backend/internal/medicalrecord/hospitalization_repository.go`
  - `backend/internal/medicalrecord/hospitalization_request.go`
  - `backend/internal/medicalrecord/hospitalization_service.go`
  - `backend/internal/medicalrecord/hospitalization_repository_test.go`
  - `backend/internal/medicalrecord/hospitalization_request_test.go`
  - `backend/internal/medicalrecord/hospitalization_service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-SCHEMA-BARRIER。
- 再実測: `git grep -n -F -e '.claude/refs/backend-application-invariants.md:38' -e 'FOR UPDATE' -e 'FOR SHARE' -e 'persistence.TxFromContext' -e 'apperrors.WrapInternalServerError' -e 'UPDATE' HEAD -- backend/internal/medicalrecord/hospitalization_repository.go backend/internal/medicalrecord/hospitalization_request.go backend/internal/medicalrecord/hospitalization_service.go backend/internal/medicalrecord/hospitalization_repository_test.go backend/internal/medicalrecord/hospitalization_request_test.go backend/internal/medicalrecord/hospitalization_service_test.go`
- 検証: `docker compose exec backend go test ./internal/medicalrecord/... -run 'Hospitalization'`
- 既知台帳: MRB-03=BUG-437 / SEC-SWEEP-02
- Size: M (6/8 files)
- Status: 完了 ｜ 担当レーン: LANE-3 ｜ 完了 commit: 3cea46b55421b3809f9be4b1250bb17ccd2db310

##### U-X03-CSVIMPORT-GUARD

- 含む所見 ID: TRM-07
- 所有パス (7):
  - `backend/internal/csvimport/import.go`
  - `backend/internal/csvimport/cutover_import.go`
  - `backend/internal/csvimport/failure_rehearsal.go`
  - `backend/cmd/seed-export/main.go`
  - `backend/internal/csvimport/import_test.go`
  - `backend/internal/csvimport/cutover_import_test.go`
  - `backend/internal/csvimport/failure_rehearsal_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-SCHEMA-BARRIER。
- 再実測: `git grep -n -F -e '.claude/refs/backend-application-invariants.md:31' -e 'current_database' -e 'DELETE' -e 'FROM' -e 'TargetDatabaseIdentitySHA256' -e 'Import' HEAD -- backend/internal/csvimport/import.go backend/internal/csvimport/cutover_import.go backend/internal/csvimport/failure_rehearsal.go backend/cmd/seed-export/main.go backend/internal/csvimport/import_test.go backend/internal/csvimport/cutover_import_test.go backend/internal/csvimport/failure_rehearsal_test.go`
- 検証: `docker compose exec backend go test ./internal/csvimport/... ./cmd/seed-export/...`
- 既知台帳: none
- Size: L (7/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### U-X04-RESERVATION-AUTODELEGATE

- 含む所見 ID: RSV-09
- 所有パス (2):
  - `backend/internal/reservation/liff_service_reservations.go`
  - `backend/internal/reservation/liff_service_reservations_test.go`
- 依存 / 決裁: 決裁=DEC-35。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/error-handling.md:9' -e 'if err == nil' -e 'slog.WarnContext(ctx, "...(best-effort)", "error", err)' -e 'slog.WarnContext' -e 'ToDateTime' -e 'StaffID' HEAD -- backend/internal/reservation/liff_service_reservations.go backend/internal/reservation/liff_service_reservations_test.go`
- 検証: `docker compose exec backend go test ./internal/reservation/... -run 'Liff.*Reservation|Delegate'`
- 既知台帳: none
- Size: S (2/8 files)
- Status: 完了 ｜ 担当レーン: LANE-2 ｜ 完了 commit: 7cd03ea9039b67361cdfb0af7222c7d2f9678a75

##### U-X04X05-LSTEP-DELIVERY

- 含む所見 ID: LSA-12, LSA-15
- 所有パス (4):
  - `backend/internal/lstep/lstep_delivery_trigger_client.go`
  - `backend/internal/lstep/lstep_delivery_monitor_service.go`
  - `backend/internal/lstep/lstep_delivery_trigger_client_test.go`
  - `backend/internal/lstep/lstep_delivery_monitor_service_test.go`
- 依存 / 決裁: 決裁=DEC-35。先行 unit=U-LSTEP-OPTOUT, U-SCHEMA-BARRIER。
- 再実測: `git grep -n -F -e '.claude/refs/error-handling.md:9' -e 'recordTrigger' -e 'ExistsTodayByOwnerAndType' -e 'WarnContext' -e 'Scheduled' -e 'Excluded' HEAD -- backend/internal/lstep/lstep_delivery_trigger_client.go backend/internal/lstep/lstep_delivery_monitor_service.go backend/internal/lstep/lstep_delivery_trigger_client_test.go backend/internal/lstep/lstep_delivery_monitor_service_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/... -run 'DeliveryTrigger|DeliveryMonitor'`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### U-X05-OWNER-PHONE

- 含む所見 ID: POC-06
- 所有パス (4):
  - `backend/internal/owner/repository.go`
  - `backend/internal/owner/service_core.go`
  - `backend/internal/owner/repository_test.go`
  - `backend/internal/owner/service_core_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-SCHEMA-BARRIER。
- 再実測: `git grep -n -F -e 'len(owners) != 1' -e 'CODING_RULES' -e 'FindByPhone' -e 'CreateWithPets' -e 'POST' -e 'First' HEAD -- backend/internal/owner/repository.go backend/internal/owner/service_core.go backend/internal/owner/repository_test.go backend/internal/owner/service_core_test.go`
- 検証: `docker compose exec backend go test ./internal/owner/... -run 'Phone|Unique'`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### BE-X08-LSTEP-CONNECTION-01

- 含む所見 ID: LSA-01, LSA-08
- 所有パス (6):
  - `backend/internal/lstep/lstep_settings_request.go`
  - `backend/internal/lstep/lstep_settings_connection.go`
  - `backend/internal/lstep/lstep_settings_credentials.go`
  - `backend/internal/lstep/lstep_settings_connection_test.go`
  - `backend/internal/lstep/lstep_settings_response.go`
  - `backend/internal/lstep/lstep_settings_handler_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=BE-X06-LSTEP-SETTINGS-01。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:151' -e 'lstep_base_url' -e 'ClinicIntegration' -e 'testLstepAPI' -e 'Authorization: Bearer <復号済み lstep API key>' -e 'crypto.MaskValue' HEAD -- backend/internal/lstep/lstep_settings_request.go backend/internal/lstep/lstep_settings_connection.go backend/internal/lstep/lstep_settings_credentials.go backend/internal/lstep/lstep_settings_connection_test.go backend/internal/lstep/lstep_settings_response.go backend/internal/lstep/lstep_settings_handler_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/... -run 'LstepSettings|Connection'`
- 既知台帳: none
- Size: M (6/8 files)
- Status: 完了 ｜ 担当レーン: LANE-1 ｜ 完了 commit: d26f86b75f6ca0a02615db83b6b035c2623ffcfc

##### BE-X08-LSTEP-SEND-01

- 含む所見 ID: LSA-09
- 所有パス (5):
  - `backend/internal/lstep/line_send_service.go`
  - `backend/internal/lstep/line_send_response.go`
  - `backend/internal/lstep/line_send_handler.go`
  - `backend/internal/lstep/line_send_service_test.go`
  - `backend/internal/lstep/line_send_handler_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e 'sendErr.Error' -e 'WrapBadGateway(fmt.Sprintf("LINE送信に失敗しました: %s", ...))' -e 'LineSendLog.ErrorMessage' -e 'GET /owners/:id/line/send-logs' -e 'error_message' -e 'ErrorMessage' HEAD -- backend/internal/lstep/line_send_service.go backend/internal/lstep/line_send_response.go backend/internal/lstep/line_send_handler.go backend/internal/lstep/line_send_service_test.go backend/internal/lstep/line_send_handler_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/... -run 'LineSend'`
- 既知台帳: none
- Size: M (5/8 files)
- Status: 完了 ｜ 担当レーン: LANE-1 ｜ 完了 commit: d822c85b722e839e7ebecdb8b9c13b66e21646b8

##### BE-X09-PET-ENUMS-01

- 含む所見 ID: POC-11
- 所有パス (6):
  - `backend/internal/owner/validators.go`
  - `backend/internal/owner/validators_test.go`
  - `backend/internal/pet/validators.go`
  - `backend/internal/pet/validators_test.go`
  - `backend/internal/sharedkernel/enum_validators.go`
  - `backend/internal/sharedkernel/sharedkernel_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '~/.claude/rules/ecc/common/coding-style.md:26' -e 'Avoid' -e 'POST' -e 'ValidateRequiredName' HEAD -- backend/internal/owner/validators.go backend/internal/owner/validators_test.go backend/internal/pet/validators.go backend/internal/pet/validators_test.go backend/internal/sharedkernel/enum_validators.go backend/internal/sharedkernel/sharedkernel_test.go`
- 検証: `docker compose exec backend go test ./internal/owner/... ./internal/pet/... ./internal/sharedkernel/... -run 'Valid|Enum|Pet'`
- 既知台帳: none
- Size: M (6/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: b60138112f5b22e63cecb1a3ae76152d67873acc

##### BE-X10-AUTH-RESPONSE-01

- 含む所見 ID: AUS-09
- 所有パス (6):
  - `backend/internal/auth/http_permission.go`
  - `backend/internal/auth/http_permission_test.go`
  - `backend/internal/auth/http_response.go`
  - `backend/internal/auth/http_session_handlers_test.go`
  - `backend/internal/httpapi/context.go`
  - `backend/internal/httpapi/context_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/error-handling.md:29' -e 'RespondError' -e 'ExtractIsSystemAdmin' -e 'ExtractStaffID' -e 'ExtractClinicID' -e 'Abort' HEAD -- backend/internal/auth/http_permission.go backend/internal/auth/http_permission_test.go backend/internal/auth/http_response.go backend/internal/auth/http_session_handlers_test.go backend/internal/httpapi/context.go backend/internal/httpapi/context_test.go`
- 検証: `docker compose exec backend go test ./internal/auth/... ./internal/httpapi/... -run 'HasPermission|RequirePermission|CalculateEffectivePermissions|ExtractContext'`
- 既知台帳: none
- Size: M (6/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: f3758050b1edb01424d381d3ad4118cfd930cefa

##### SOLO-09

- 含む所見 ID: LSA-04
- 所有パス (3):
  - `backend/internal/lstep/lstep_tag_config_repository.go`
  - `backend/internal/lstep/routes.go`
  - `backend/internal/lstep/lstep_tag_config_repository_test.go`
- 依存 / 決裁: 決裁=DEC-30。先行 unit=U-X02-LSTEP-TAG-CONFIG。
- 再実測: `git grep -n -F -e 'clinic_id NOT NULL' -e 'lstep_auto_managed_prefixes' -e 'lstep_condition_tag_mappings' -e 'lstep_send_purpose_tag_prefixes' -e 'ResourceHospitalSettings' -e 'Delete(&model.LstepAutoManagedPrefix{}, id)' HEAD -- backend/internal/lstep/lstep_tag_config_repository.go backend/internal/lstep/routes.go backend/internal/lstep/lstep_tag_config_repository_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/ -run "TagConfig|AutoManaged|PurposeTag"`
- 既知台帳: none
- Size: M (3/8 files)
- Status: 完了 ｜ 担当レーン: LANE-1 ｜ 完了 commit: 7240c535ed194f7a6c665fb4c0ffa3299599fadd

##### U-TRIMMING-SERVICE

- 含む所見 ID: TRM-04, TRM-06
- 所有パス (4):
  - `backend/internal/trimming/trimming_service.go`
  - `backend/internal/trimming/trimming_service_test.go`
  - `backend/internal/trimming/trimming_handler.go`
  - `backend/internal/trimming/trimming_handler_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=BE-X07-BODY-01。
- 再実測: `git grep -n -F -e 'CODING_RULES' -e 'SetOptions' -e 'Location' -e 'Conflict' -e 'PATCH' -e 'trimming_service' HEAD -- backend/internal/trimming/trimming_service.go backend/internal/trimming/trimming_service_test.go backend/internal/trimming/trimming_handler.go backend/internal/trimming/trimming_handler_test.go`
- 検証: `docker compose exec backend go test ./internal/trimming/... -run 'ExistingAppointment|Create|Conflict|Detail|Log'`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 完了 ｜ 担当レーン: LANE-2 ｜ 完了 commit: eee2c2f6b9a670b5aa6201db9b52c131a6ffe2c3

##### U-X01X05-RESERVATION

- 含む所見 ID: RSV-02, RSV-03, RSV-07
- 所有パス (7):
  - `backend/internal/reservation/appointment_admin_service.go`
  - `backend/internal/reservation/line_reservation_setting_service.go`
  - `backend/internal/reservation/reservation_type_liff_service.go`
  - `backend/internal/reservation/reservation_type_repository.go`
  - `backend/internal/reservation/reservation_repository.go`
  - `backend/internal/reservation/appointment_admin_service_test.go`
  - `backend/internal/reservation/reservation_type_liff_service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=BE-X06-RSV-CANCEL-01, U-SCHEMA-BARRIER。
- 再実測: `git grep -n -F -e 'reservation_repository.go:164-176' -e 's.resRepo.AcquireBookingLock(ctx, clinicID)' -e 'resRepo.AcquireBookingLock' -e 'CODING_RULES' -e 'AcquireBookingLock' -e 'CountConflicts' HEAD -- backend/internal/reservation/appointment_admin_service.go backend/internal/reservation/line_reservation_setting_service.go backend/internal/reservation/reservation_type_liff_service.go backend/internal/reservation/reservation_type_repository.go backend/internal/reservation/reservation_repository.go backend/internal/reservation/appointment_admin_service_test.go backend/internal/reservation/reservation_type_liff_service_test.go`
- 検証: `docker compose exec backend go test ./internal/reservation/... -run 'Admin|LineReservationSetting|ReservationType|Delete' && docker compose exec backend go test ./internal/lintscan/ -run DBOrTx`
- 既知台帳: none
- Size: L (7/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### U-X02-CLINIC-CONTACT

- 含む所見 ID: POC-17
- 所有パス (5):
  - `backend/internal/clinic/clinic_request.go`
  - `backend/internal/clinic/company_request.go`
  - `backend/internal/owner/validators_contact.go`
  - `backend/internal/clinic/clinic_request_test.go`
  - `backend/internal/clinic/company_request_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:151' -e 'CreateClinicRequest' -e 'UpdateClinicRequest' -e 'UpdateCompanyRequest' -e 'Email' -e 'PhoneNumber' HEAD -- backend/internal/clinic/clinic_request.go backend/internal/clinic/company_request.go backend/internal/owner/validators_contact.go backend/internal/clinic/clinic_request_test.go backend/internal/clinic/company_request_test.go`
- 検証: `docker compose exec backend go test ./internal/clinic/... ./internal/owner/... -run 'Request|Email|Phone|Postal'`
- 既知台帳: none
- Size: M (5/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### U-X02-LSTEP-AGGREGATION

- 含む所見 ID: G2A-05
- 所有パス (2):
  - `backend/internal/lstep/aggregation_request.go`
  - `backend/internal/lstep/aggregation_request_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:151' -e 'aggregation_request' -e 'period_preset' -e 'amount_basis' HEAD -- backend/internal/lstep/aggregation_request.go backend/internal/lstep/aggregation_request_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/... -run 'AggregationRequest'`
- 既知台帳: none
- Size: S (2/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: 293d516da5e7270124c8e283a11599efcf40211b

##### U-X02-MR-LAB-IMPORT

- 含む所見 ID: MRC-08
- 所有パス (1):
  - `backend/internal/medicalrecord/lab_import_request.go`
- 依存 / 決裁: 決裁=なし。先行 unit=BE-X06-MEDICAL-ATOMIC-01。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:151' -e 'labExamItemReq' -e 'labImportResultRowReq' -e 'toExamInputs' -e 'Name' -e 'InspectionValue' HEAD -- backend/internal/medicalrecord/lab_import_request.go`
- 検証: `docker compose exec backend go test ./internal/medicalrecord/... -run 'LabImport'`
- 既知台帳: none
- Size: S (1/8 files)
- Status: 完了 ｜ 担当レーン: LANE-3 ｜ 完了 commit: c7cd4c568edc9915ea56559d3fb7b2e2d72d6eed

##### U-X05-MR-MEDICINE-MASTERS

- 含む所見 ID: MRC-02, MRC-03, MRC-07
- 所有パス (8):
  - `backend/internal/medicalrecord/medicine_dose_param_service.go`
  - `backend/internal/medicalrecord/medicine_service.go`
  - `backend/internal/medicalrecord/procedure_repository.go`
  - `backend/internal/medicalrecord/procedure_service.go`
  - `backend/internal/medicalrecord/medicine_dose_param_service_test.go`
  - `backend/internal/medicalrecord/medicine_service_test.go`
  - `backend/internal/medicalrecord/procedure_repository_test.go`
  - `backend/internal/medicalrecord/procedure_service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-X01X02-INVENTORY。
- 再実測: `git grep -n -F -e '.claude/refs/go-gin-backend-review.md:66' -e 'medicines.inventory_id' -e 'Name' -e 'RowsAffected' -e 'DeleteByNameAndMedicineCategory' -e 'AuditTxLogger' HEAD -- backend/internal/medicalrecord/medicine_dose_param_service.go backend/internal/medicalrecord/medicine_service.go backend/internal/medicalrecord/procedure_repository.go backend/internal/medicalrecord/procedure_service.go backend/internal/medicalrecord/medicine_dose_param_service_test.go backend/internal/medicalrecord/medicine_service_test.go backend/internal/medicalrecord/procedure_repository_test.go backend/internal/medicalrecord/procedure_service_test.go`
- 検証: `docker compose exec backend go test ./internal/medicalrecord/... -run 'MedicineDoseParam|Medicine.*Delete|Procedure.*Delete' && docker compose exec backend go test ./internal/lintscan/ -run DBOrTx`
- 既知台帳: none
- Size: L (8/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### BE-X09-CLOSING-01

- 含む所見 ID: POC-15, POC-16
- 所有パス (5):
  - `backend/internal/clinic/closing_settings_service.go`
  - `backend/internal/clinic/closing_settings_service_test.go`
  - `backend/internal/billing/cash_register_service.go`
  - `backend/internal/billing/cash_register_service_test.go`
  - `backend/internal/sharedkernel/shift_times.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e 'slog.ErrorContext' -e 'slog.Any' -e 'CODING_RULES' -e 'UpdateSpecialPeriod' -e 'ErrorContext' -e 'closing_settings_service' HEAD -- backend/internal/clinic/closing_settings_service.go backend/internal/clinic/closing_settings_service_test.go backend/internal/billing/cash_register_service.go backend/internal/billing/cash_register_service_test.go backend/internal/sharedkernel/shift_times.go`
- 検証: `docker compose exec backend go test ./internal/clinic/... ./internal/billing/... ./internal/sharedkernel/... -run 'Closing|CashRegister|HHMM'`
- 既知台帳: none
- Size: M (5/8 files)
- Status: 完了 ｜ 担当レーン: LANE-1 ｜ 完了 commit: 04199d6e18c097c82d4269330eef9f5aad649c9a

##### BE-X09-MEDICAL-DIAGNOSIS-01

- 含む所見 ID: MRC-14
- 所有パス (4):
  - `backend/internal/medicalrecord/medical_record_subrecords.go`
  - `backend/internal/medicalrecord/medical_record_subrecords_test.go`
  - `backend/internal/medicalrecord/clinical_plan_service.go`
  - `backend/internal/medicalrecord/clinical_plan_service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '~/.claude/rules/ecc/common/coding-style.md:26' -e 'Avoid' -e 'CreateSubRecords' -e 'ClinicalPlanService' -e 'POST' -e 'medical_record_subrecords' HEAD -- backend/internal/medicalrecord/medical_record_subrecords.go backend/internal/medicalrecord/medical_record_subrecords_test.go backend/internal/medicalrecord/clinical_plan_service.go backend/internal/medicalrecord/clinical_plan_service_test.go`
- 検証: `docker compose exec backend go test ./internal/medicalrecord/... -run 'CreateSubRecords.*Diagnosis|ClinicalPlan.*Diagnosis'`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 完了 ｜ 担当レーン: LANE-3 ｜ 完了 commit: 33553d4abff37ca6e831280b1cb6854e9c4d81db

##### BE-X09-PET-PATCH-01

- 含む所見 ID: POC-14
- 所有パス (2):
  - `backend/internal/pet/chronic_condition_service.go`
  - `backend/internal/pet/chronic_condition_service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '~/.claude/rules/ecc/common/coding-style.md:26' -e 'apperrors.WrapInvalidInput' -e 'Avoid' -e 'WrapInvalidInput' -e 'PATCH' -e 'RowsAffected' HEAD -- backend/internal/pet/chronic_condition_service.go backend/internal/pet/chronic_condition_service_test.go`
- 検証: `docker compose exec backend go test ./internal/pet/... -run 'ChronicCondition.*Update'`
- 既知台帳: none
- Size: S (2/8 files)
- Status: 完了 ｜ 担当レーン: LANE-2 ｜ 完了 commit: 409b1bc194889ea11b718986c25dc12f9cc2aa82

##### U-X01X03X04-LSTEP-LIFECYCLE

- 含む所見 ID: LSA-05, LSB-02, LSB-03, LSB-04
- 所有パス (4):
  - `backend/internal/lstep/lstep_lifecycle_handler.go`
  - `backend/internal/lstep/lstep_lifecycle_service.go`
  - `backend/internal/lstep/lstep_lifecycle_service_test.go`
  - `backend/internal/lstep/lstep_settings_credentials_test.go`
- 依存 / 決裁: 決裁=DEC-35。先行 unit=BE-X08-LSTEP-CONNECTION-01, BE-X06-LSTEP-SETTINGS-01。
- 再実測: `git grep -n -F -e '.claude/refs/backend-application-invariants.md:31' -e 'HandlePetDeath' -e 'HandlePetRevival' -e 'HandleOwnerOptOut' -e 'HandleOwnerOptIn' -e 'HandleOwnerDeletion' HEAD -- backend/internal/lstep/lstep_lifecycle_handler.go backend/internal/lstep/lstep_lifecycle_service.go backend/internal/lstep/lstep_lifecycle_service_test.go backend/internal/lstep/lstep_settings_credentials_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/... -run 'Lifecycle|Settings|Credentials'`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### U-X04-AUDIT-MARSHAL

- 含む所見 ID: INF-06
- 所有パス (4):
  - `backend/internal/audit/repository.go`
  - `backend/internal/audit/service.go`
  - `backend/internal/audit/repository_test.go`
  - `backend/internal/audit/service_test.go`
- 依存 / 決裁: 決裁=DEC-35。先行 unit=なし。
- 再実測: `git grep -n -F -e '~/.claude/rules/ecc/common/coding-style.md:49' -e '明示的に回復する' -e 'Never' -e 'MarshalJSON' -e 'Marshal' -e 'OldValue' HEAD -- backend/internal/audit/repository.go backend/internal/audit/service.go backend/internal/audit/repository_test.go backend/internal/audit/service_test.go`
- 検証: `docker compose exec backend go test ./internal/audit/... -run 'Marshal|BuildLog'`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 完了 ｜ 担当レーン: LANE-2 ｜ 完了 commit: dfa4013baaa5e6963df43e80c8d13529e81c70b5

##### U-X04-COVERAGE-RATCHET

- 含む所見 ID: CMD-03
- 所有パス (3):
  - `backend/cmd/coverage-ratchet/main.go`
  - `backend/cmd/coverage-ratchet/main_test.go`
  - `.github/workflows/ci.yml`
- 依存 / 決裁: 決裁=DEC-35。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/error-handling.md:9' -e '.github/workflows/ci.yml:200' -e 'bash -e {0}' -e 'os.Exit' -e 'ReadFile' -e 'GitHub' HEAD -- backend/cmd/coverage-ratchet/main.go backend/cmd/coverage-ratchet/main_test.go .github/workflows/ci.yml`
- 検証: `docker compose exec backend go test ./cmd/coverage-ratchet/...`
- 既知台帳: none
- Size: M (3/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### U-X04-LSTEP-BATCH

- 含む所見 ID: LSA-03
- 所有パス (4):
  - `backend/internal/lstep/lstep_batch_segmentation.go`
  - `backend/internal/lstep/lstep_batch_service.go`
  - `backend/internal/lstep/lstep_batch_segmentation_test.go`
  - `backend/internal/lstep/lstep_batch_service_test.go`
- 依存 / 決裁: 決裁=DEC-35。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/error-handling.md:9' -e 'syncVisitDormantForClinic' -e 'FindDormantOwnerEntries' -e 'return 0, nil' -e 'runBatchAllClinicsWithResult' -e 'count>0 || len(errs)>0' HEAD -- backend/internal/lstep/lstep_batch_segmentation.go backend/internal/lstep/lstep_batch_service.go backend/internal/lstep/lstep_batch_segmentation_test.go backend/internal/lstep/lstep_batch_service_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/... -run 'VisitDormant|Batch'`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### U-X04-LSTEP-HEALTH-REMOVE

- 含む所見 ID: G2B-01
- 所有パス (5):
  - `backend/internal/lstep/lstep_health_tag_sync_batch.go`
  - `backend/internal/lstep/lstep_health_tag_sync_food.go`
  - `backend/internal/lstep/lstep_health_tag_sync_prevention.go`
  - `backend/internal/lstep/lstep_health_tag_sync_batch_test.go`
  - `backend/internal/lstep/lstep_health_tag_sync_prevention_test.go`
- 依存 / 決裁: 決裁=DEC-35。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/error-handling.md:9' -e 'PREV_' -e 'BatchRunResult' -e 'Failed' -e 'lstep_health_tag_sync_prevention' -e 'lstep_health_tag_sync_food' HEAD -- backend/internal/lstep/lstep_health_tag_sync_batch.go backend/internal/lstep/lstep_health_tag_sync_food.go backend/internal/lstep/lstep_health_tag_sync_prevention.go backend/internal/lstep/lstep_health_tag_sync_batch_test.go backend/internal/lstep/lstep_health_tag_sync_prevention_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/... -run 'Health.*Tag|Prevention|Vaccine'`
- 既知台帳: none
- Size: M (5/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### U-X04-LSTEP-MIGRATE

- 含む所見 ID: CMD-07
- 所有パス (2):
  - `backend/cmd/lstep-migrate/main.go`
  - `backend/cmd/lstep-migrate/migrator.go`
- 依存 / 決裁: 決裁=DEC-35。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/error-handling.md:9' -e 'processOwner' -e 'FirstOrCreate' -e 'Warn' -e 'ProgressRecord' -e 'Status' HEAD -- backend/cmd/lstep-migrate/main.go backend/cmd/lstep-migrate/migrator.go`
- 検証: `docker compose exec backend go test ./cmd/lstep-migrate/...`
- 既知台帳: none
- Size: S (2/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### U-X04-LSTEP-OWNER-TAGS

- 含む所見 ID: LSA-10
- 所有パス (4):
  - `backend/internal/lstep/lstep_tag_handler.go`
  - `backend/internal/lstep/lstep_tag_service.go`
  - `backend/internal/lstep/lstep_tag_handler_test.go`
  - `backend/internal/lstep/lstep_tag_service_test.go`
- 依存 / 決裁: 決裁=DEC-35。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/error-handling.md:9' -e 'tagCacheRepo.FindByOwner' -e 'return result, nil' -e 'return nil, apperrors.Wrap(cacheErr, "failed to load lstep tag cache")' -e 'apperrors.Wrap' -e 'FindByOwner' HEAD -- backend/internal/lstep/lstep_tag_handler.go backend/internal/lstep/lstep_tag_service.go backend/internal/lstep/lstep_tag_handler_test.go backend/internal/lstep/lstep_tag_service_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/... -run 'GetOwnerTags'`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: 6b67a81c2e9fd5fa8bb6c7b8445dd881ed38beb7

##### U-X04-LSTEP-STALE-TAGS

- 含む所見 ID: LSA-11
- 所有パス (2):
  - `backend/internal/lstep/lstep_tag_sync_api.go`
  - `backend/internal/lstep/lstep_tag_sync_api_test.go`
- 依存 / 決裁: 決裁=DEC-35。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/error-handling.md:9' -e 'apiFailed bool' -e 'false' -e 'next_visit_*' -e 'checkup_done_*' -e '(apiFailed bool, err error)' HEAD -- backend/internal/lstep/lstep_tag_sync_api.go backend/internal/lstep/lstep_tag_sync_api_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/... -run 'RemoveStaleTags'`
- 既知台帳: none
- Size: S (2/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: ba1d03452777c70edf2a0458248ab96ee0877a79

##### SOLO-01

- 含む所見 ID: AUS-04, AUS-05, G2F-10
- 所有パス (8):
  - `backend/internal/staff/shift_template_repository.go`
  - `backend/internal/staff/shift_template_service.go`
  - `backend/internal/staff/shift_template_repository_integration_test.go`
  - `backend/internal/staff/shift_template_service_test.go`
  - `backend/internal/staff/shift_response.go`
  - `backend/internal/staff/shift_response_test.go`
  - `backend/internal/staff/shift_entry_repository.go`
  - `backend/internal/staff/shift_entry_service.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:89' -e '_ = ctx; _ = clinicID; _ = id; return 0, nil' -e 'assert.Equal(int64(0), count)' -e 'assert.Equal' -e 'int64' -e 'CountUsageByShiftTemplateID' HEAD -- backend/internal/staff/shift_template_repository.go backend/internal/staff/shift_template_service.go backend/internal/staff/shift_template_repository_integration_test.go backend/internal/staff/shift_template_service_test.go backend/internal/staff/shift_response.go backend/internal/staff/shift_response_test.go backend/internal/staff/shift_entry_repository.go backend/internal/staff/shift_entry_service.go`
- 検証: `docker compose exec backend go test ./internal/staff/ -run "ShiftTemplate|ShiftResponse|ShiftEntry"`
- 既知台帳: none
- Size: L (8/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: 38c41ae3173176680003e5ab85f682cd04cb784a

##### SOLO-02

- 含む所見 ID: BIL-01
- 所有パス (7):
  - `backend/internal/billing/routes.go`
  - `backend/internal/billing/billing_item_service.go`
  - `backend/internal/billing/billing_item_repository.go`
  - `backend/internal/billing/accounting_service_core.go`
  - `backend/internal/billing/billing_item_service_test.go`
  - `backend/internal/billing/billing_item_repository_tx_atomicity_test.go`
  - `backend/internal/billing/billing_item_handler_test.go`
- 依存 / 決裁: 決裁=DEC-33。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/backend-application-invariants.md:37' -e 'PATCH /billing-items/:id' -e 'POST /billing-items' -e 'accounting:edit' -e 'accounting:create' -e 'input.VaccinationID != nil' HEAD -- backend/internal/billing/routes.go backend/internal/billing/billing_item_service.go backend/internal/billing/billing_item_repository.go backend/internal/billing/accounting_service_core.go backend/internal/billing/billing_item_service_test.go backend/internal/billing/billing_item_repository_tx_atomicity_test.go backend/internal/billing/billing_item_handler_test.go`
- 検証: `docker compose exec backend go test ./internal/billing/ -run "BillingItem|PostClose"`
- 既知台帳: BIL-01=BUG-463
- Size: L (7/8 files)
- Status: 対象消失（BIL-01/BUG-463 は post-close・status ガード実装済み） ｜ 担当レーン: LANE-4 ｜ 完了 commit: —

##### SOLO-03

- 含む所見 ID: BIL-02
- 所有パス (4):
  - `backend/internal/billing/accounting_service_correction.go`
  - `backend/internal/billing/accounting_service_reports.go`
  - `backend/internal/billing/accounting_service.go`
  - `backend/internal/billing/accounting_service_correction_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/backend-application-invariants.md:37' -e 'NewAccountingService' -e 'logCreditCorrection' -e 'if s.auditTx == nil { return nil }' -e 'Cancel' -e 'if s.auditTx != nil' HEAD -- backend/internal/billing/accounting_service_correction.go backend/internal/billing/accounting_service_reports.go backend/internal/billing/accounting_service.go backend/internal/billing/accounting_service_correction_test.go`
- 検証: `docker compose exec backend go test ./internal/billing/ -run "CreditCorrection|Cancel"`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: ddc3250b99b4b948c8fd66ea689036d06506d9e0

##### SOLO-05

- 含む所見 ID: CMD-04
- 所有パス (2):
  - `backend/cmd/api/lstep_adapters.go`
  - `backend/internal/lstep/lstep_lifecycle_deps.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-X05-PET-UPDATE, BE-X07-BODY-01。
- 再実測: `git grep -n -F -e 'map[string]any' -e 'pet.CompleteRepository' -e 'ClearDeath' -e 'petLifecycleWriterAdapter' -e 'legacyLifecycleTransition' -e 'CODING_RULES' HEAD -- backend/cmd/api/lstep_adapters.go backend/internal/lstep/lstep_lifecycle_deps.go`
- 検証: `docker compose exec backend go test ./cmd/api/ ./internal/pet/ ./internal/lstep/ -run "Lifecycle|PetDeath"`
- 既知台帳: none
- Size: S (2/8 files)
- Status: 完了 ｜ 担当レーン: LANE-2 ｜ 完了 commit: b2a4af75a08b045adb4d5a76645f589329df829c

##### SOLO-06

- 含む所見 ID: CMD-06
- 所有パス (2):
  - `backend/cmd/csv-import-failure-rehearsal/main.go`
  - `backend/cmd/csv-import-failure-rehearsal/main_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/error-handling.md:17' -e 'fmt.Errorf("...: %w", err)' -e 'fmt.Errorf' -e 'validateDisposableTarget' -e 'Errorf' HEAD -- backend/cmd/csv-import-failure-rehearsal/main.go backend/cmd/csv-import-failure-rehearsal/main_test.go`
- 検証: `docker compose exec backend go test ./cmd/csv-import-failure-rehearsal/`
- 既知台帳: none
- Size: S (2/8 files)
- Status: 完了 ｜ 担当レーン: LANE-2 ｜ 完了 commit: e3a704c60a27d77055bbdb830c48a5c4153566c4

##### SOLO-08

- 含む所見 ID: INF-03
- 所有パス (4):
  - `backend/internal/infra/lstep/user.go`
  - `backend/internal/infra/lstep/tag.go`
  - `backend/internal/infra/lstep/client.go`
  - `backend/internal/owner/service_line.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-X02-PET-OWNER-FREETEXT。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:151' -e 'fmt.Sprintf("/contacts/%s", lineUserID)' -e 'fmt.Sprintf("/contacts/%s/tags", lineUserID)' -e 'c.baseURL+path' -e 'patchOwnerLineUserIDRequest.LineUserID' -e 'url.PathEscape(lineUserID)' HEAD -- backend/internal/infra/lstep/user.go backend/internal/infra/lstep/tag.go backend/internal/infra/lstep/client.go backend/internal/owner/service_line.go`
- 検証: `docker compose exec backend go test ./internal/infra/lstep/ ./internal/owner/ -run "Line|Lstep"`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 完了 ｜ 担当レーン: LANE-1 ｜ 完了 commit: 407fd11ea2a1d843e7ede17c621c4a217116b748

##### SOLO-10

- 含む所見 ID: LSA-07
- 所有パス (2):
  - `backend/internal/lstep/line_messaging_service.go`
  - `backend/internal/lstep/line_messaging_service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e 'slog.InfoContext(ctx, "LINE push sent", "to", lineUserID)' -e 'owner_id' -e '"to", lineUserID' -e 'slog.InfoContext' -e 'CODING_RULES' -e 'InfoContext' HEAD -- backend/internal/lstep/line_messaging_service.go backend/internal/lstep/line_messaging_service_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/ -run "LineMessaging"`
- 既知台帳: none
- Size: S (2/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: e42b295554dd258815743ed3eb35c99ccf4f8745

##### SOLO-12

- 含む所見 ID: G2C-01, G2F-01
- 所有パス (1):
  - `backend/internal/lstep/lstep_tag_cache_repository.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-LSTEP-OPTOUT。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:151' -e 'EscapeLike' -e 'ESCAPE' -e 'lstep_tag_cache_repository' -e '.claude/refs/go-gin-backend-review.md:67' -e 'EXPLAIN' HEAD -- backend/internal/lstep/lstep_tag_cache_repository.go`
- 検証: `docker compose exec backend go test ./internal/lstep/ -run "TagCache|DeliveryTriggerBatch|DeliveryTriggerState"`
- 既知台帳: none
- Size: S (1/8 files)
- Status: 完了 ｜ 担当レーン: LANE-1 ｜ 完了 commit: 837272bed58c08c32b60d37b53e6e70c2402f1bf

##### SOLO-13

- 含む所見 ID: G2C-04
- 所有パス (4):
  - `backend/internal/lstep/lstep_tag_summary_service.go`
  - `backend/internal/lstep/lstep_tag_summary_handler.go`
  - `backend/internal/lstep/lstep_tag_summary_service_test.go`
  - `backend/internal/lstep/lstep_tag_summary_handler_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/rules/go-gin-backend-guidelines.md:154' -e '.claude/refs/error-handling.md:29' -e 'RespondError' -e 'lstep_tag_summary_service' -e 'lstep_tag_summary_handler' HEAD -- backend/internal/lstep/lstep_tag_summary_service.go backend/internal/lstep/lstep_tag_summary_handler.go backend/internal/lstep/lstep_tag_summary_service_test.go backend/internal/lstep/lstep_tag_summary_handler_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/ -run "TagSummary"`
- 既知台帳: G2C-04=BUG-464
- Size: M (4/8 files)
- Status: 対象消失（BUG-464 実装済み: fail-closed 5000 行上限 + stream 後 RespondError 抑止） ｜ 担当レーン: LANE-4 ｜ 完了 commit: —

##### SOLO-14

- 含む所見 ID: G2F-02
- 所有パス (4):
  - `backend/internal/lstep/lstep_health_tag_sync_checkup.go`
  - `backend/internal/lstep/lstep_health_tag_sync_vaccine.go`
  - `backend/internal/lstep/lstep_health_tag_sync_checkup_test.go`
  - `backend/internal/lstep/lstep_health_tag_sync_vaccine_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-X04-LSTEP-HEALTH-REMOVE。
- 再実測: `git grep -n -F -e '.claude/refs/go-gin-backend-review.md:67' -e 'EXPLAIN' -e 'Limit' -e 'lstep_health_tag_sync_batch' -e 'lstep_health_tag_sync_checkup' -e 'lstep_health_tag_sync_vaccine' HEAD -- backend/internal/lstep/lstep_health_tag_sync_checkup.go backend/internal/lstep/lstep_health_tag_sync_vaccine.go backend/internal/lstep/lstep_health_tag_sync_checkup_test.go backend/internal/lstep/lstep_health_tag_sync_vaccine_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/ -run "HealthTagSync"`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### SOLO-15

- 含む所見 ID: G2F-03, G2F-04
- 所有パス (4):
  - `backend/internal/lstep/lstep_tag_sync_visit_ltv.go`
  - `backend/internal/lstep/lstep_tag_sync_visit_ltv_test.go`
  - `backend/internal/billing/accounting_repository_ltv.go`
  - `backend/internal/medicalrecord/medical_record_owner_visit_repository.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-X04-LSTEP-BATCH, U-X05-OWNER-PHONE。
- 再実測: `git grep -n -F -e '.claude/refs/go-gin-backend-review.md:67' -e 'EXPLAIN' -e 'FindAllWithLineUserID' -e 'Limit' -e 'lstep_tag_sync_visit_ltv' -e 'accounting_repository_ltv' HEAD -- backend/internal/lstep/lstep_tag_sync_visit_ltv.go backend/internal/lstep/lstep_tag_sync_visit_ltv_test.go backend/internal/billing/accounting_repository_ltv.go backend/internal/medicalrecord/medical_record_owner_visit_repository.go`
- 検証: `docker compose exec backend go test ./internal/lstep/ ./internal/owner/ ./internal/medicalrecord/ ./internal/billing/ -run "LTV|Dormant|Segmentation"`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### SOLO-17

- 含む所見 ID: G2F-06
- 所有パス (4):
  - `backend/internal/lstep/shared_file_repository.go`
  - `backend/internal/lstep/shared_file_service.go`
  - `backend/internal/lstep/shared_file_repository_test.go`
  - `backend/internal/lstep/shared_file_service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/go-gin-backend-review.md:67' -e 'EXPLAIN' -e 'FindAll' -e 'FindExpired' -e 'Limit' -e 'shared_file_repository' HEAD -- backend/internal/lstep/shared_file_repository.go backend/internal/lstep/shared_file_service.go backend/internal/lstep/shared_file_repository_test.go backend/internal/lstep/shared_file_service_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/ -run "SharedFile"`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 完了 ｜ 担当レーン: LANE-4 ｜ 完了 commit: a496d3904e383c42872a773fd967f9b586a1fd90

##### SOLO-19

- 含む所見 ID: G2F-08
- 所有パス (3):
  - `backend/internal/lstep/lstep_batch_noshow.go`
  - `backend/internal/lstep/lstep_batch_noshow_test.go`
  - `backend/internal/reservation/reservation_repository_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-X01X05-RESERVATION。
- 再実測: `git grep -n -F -e '.claude/refs/go-gin-backend-review.md:67' -e 'EXPLAIN' -e 'Limit' -e 'LIMIT' -e 'reservation_repository' -e 'lstep_batch_noshow' HEAD -- backend/internal/lstep/lstep_batch_noshow.go backend/internal/lstep/lstep_batch_noshow_test.go backend/internal/reservation/reservation_repository_test.go`
- 検証: `docker compose exec backend go test ./internal/lstep/ ./internal/reservation/ -run "NoShow"`
- 既知台帳: none
- Size: M (3/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### SOLO-20

- 含む所見 ID: G2F-09
- 所有パス (2):
  - `backend/internal/billing/accounting_repository_reports_close.go`
  - `backend/internal/billing/accounting_repository_reports_close_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/go-gin-backend-review.md:67' -e 'EXPLAIN' -e 'accounting_repository_reports_close' HEAD -- backend/internal/billing/accounting_repository_reports_close.go backend/internal/billing/accounting_repository_reports_close_test.go`
- 検証: `docker compose exec backend go test ./internal/billing/ -run "Close"`
- 既知台帳: none
- Size: S (2/8 files)
- Status: 対象消失（G2F-09 DOWNGRADED: PeriodStart/End 束縛の締め詳細 dump。lifetime unbounded 非該当・pagination 非必須） ｜ 担当レーン: LANE-4 ｜ 完了 commit: —

##### SOLO-21

- 含む所見 ID: G2F-11
- 所有パス (2):
  - `backend/internal/medicalrecord/consultation_repository.go`
  - `backend/internal/medicalrecord/consultation_repository_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=U-X05-MR-EXAMTYPE, U-X01X02-INVENTORY。
- 再実測: `git grep -n -F -e 'ConsultationRepository' -e 'consultationRepositoryImpl' -e 'NewConsultationRepository' -e 'FindAll' -e 'FindByID' -e 'Reorder' HEAD -- backend/internal/medicalrecord/consultation_repository.go backend/internal/medicalrecord/consultation_repository_test.go`
- 検証: `docker compose exec backend go test ./internal/medicalrecord/ ./internal/inventory/ -run "Consultation|ExamType|Merchandise"`
- 既知台帳: none
- Size: S (2/8 files)
- Status: 完了 ｜ 担当レーン: LANE-3 ｜ 完了 commit: 2b49489ac9e7db5ceb200200bf6c4a09eda73a19

##### SOLO-22

- 含む所見 ID: INF-04
- 所有パス (2):
  - `backend/internal/apperrors/errors.go`
  - `backend/internal/apperrors/errors_test.go`
- 依存 / 決裁: 決裁=DEC-34。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/error-handling.md:18' -e 'errors.Is' -e 'errors.As' -e 'strings.Contains(errMsg, "unable to encode")' -e 'strings.Contains' -e 'FromGORM' HEAD -- backend/internal/apperrors/errors.go backend/internal/apperrors/errors_test.go`
- 検証: `docker compose exec backend go test ./internal/apperrors/ -run "FromGORM"`
- 既知台帳: none
- Size: S (2/8 files)
- Status: 完了 ｜ 担当レーン: LANE-2 ｜ 完了 commit: 09bf7e5dfc4d1431b5cc9c1458c31a7666790156

##### SOLO-23

- 含む所見 ID: MDL-05
- 所有パス (1):
  - `backend/internal/lintscan/migration_cascade_lint_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e 'owners' -e 'pets' -e 'medical_records' -e 'strings.Count(sql, "ON DELETE CASCADE")' -e 'on delete cascade' -e 'ON DELETE  CASCADE' HEAD -- backend/internal/lintscan/migration_cascade_lint_test.go`
- 検証: `docker compose exec backend go test ./internal/lintscan/ -run "MigrationCascade"`
- 既知台帳: none
- Size: S (1/8 files)
- Status: 完了 ｜ 担当レーン: LANE-3 ｜ 完了 commit: 42aff952cd6b631aabe73af9a2f0e46656f76b34

##### SOLO-25

- 含む所見 ID: MRA-04
- 所有パス (6):
  - `backend/internal/medicalrecord/cage_handler.go`
  - `backend/internal/medicalrecord/cage_request.go`
  - `backend/internal/medicalrecord/cage_repository.go`
  - `backend/internal/medicalrecord/cage_service.go`
  - `backend/internal/medicalrecord/consultation_handler.go`
  - `backend/internal/medicalrecord/consultation_service.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/go-language.md:13' -e 'package medicalrecord' -e 'GoDoc' -e 'Package' -e 'Cage' -e 'cage_handler' HEAD -- backend/internal/medicalrecord/cage_handler.go backend/internal/medicalrecord/cage_request.go backend/internal/medicalrecord/cage_repository.go backend/internal/medicalrecord/cage_service.go backend/internal/medicalrecord/consultation_handler.go backend/internal/medicalrecord/consultation_service.go`
- 検証: `docker compose exec backend go test ./internal/medicalrecord/ -run "Cage|Consultation"`
- 既知台帳: none
- Size: M (6/8 files)
- Status: 完了 ｜ 担当レーン: LANE-3 ｜ 完了 commit: 73bb810da4b3a14f4b43ba309371128caee497b9

##### SOLO-29

- 含む所見 ID: MRC-09
- 所有パス (6):
  - `backend/internal/medicalrecord/medical_record_image_request.go`
  - `backend/internal/medicalrecord/medical_record_image_service.go`
  - `backend/internal/medicalrecord/medical_record_image_handler.go`
  - `backend/internal/medicalrecord/medical_record_image_request_test.go`
  - `backend/internal/medicalrecord/medical_record_image_service_test.go`
  - `backend/internal/medicalrecord/medical_record_image_handler_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e 'binding:"required"' -e 'MimeType' -e 'FileName' -e 'FileSize' -e 'allowedMedicalRecordImageMIMETypes' -e 'medicalRecordImageService.Create' HEAD -- backend/internal/medicalrecord/medical_record_image_request.go backend/internal/medicalrecord/medical_record_image_service.go backend/internal/medicalrecord/medical_record_image_handler.go backend/internal/medicalrecord/medical_record_image_request_test.go backend/internal/medicalrecord/medical_record_image_service_test.go backend/internal/medicalrecord/medical_record_image_handler_test.go`
- 検証: `docker compose exec backend go test ./internal/medicalrecord/ -run "MedicalRecordImage"`
- 既知台帳: none
- Size: M (6/8 files)
- Status: 完了 ｜ 担当レーン: LANE-3 ｜ 完了 commit: bb3d34582bf5e4b6b15ae4628c522d4a9b91c39a

##### SOLO-30

- 含む所見 ID: MRD-01
- 所有パス (4):
  - `backend/internal/medicalrecord/treatment_repository.go`
  - `backend/internal/medicalrecord/treatment_service.go`
  - `backend/internal/medicalrecord/treatment_repository_test.go`
  - `backend/internal/medicalrecord/treatment_service_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=なし。
- 再実測: `git grep -n -F -e '.claude/refs/go-gin-backend-review.md:66' -e 'BulkUpdateSortOrder' -e 'Update("sort_order", ...)' -e 'result.RowsAffected' -e 'result.Error' -e 'RowsAffected == 0' HEAD -- backend/internal/medicalrecord/treatment_repository.go backend/internal/medicalrecord/treatment_service.go backend/internal/medicalrecord/treatment_repository_test.go backend/internal/medicalrecord/treatment_service_test.go`
- 検証: `docker compose exec backend go test ./internal/medicalrecord/ -run "Treatment.*Sort|BulkUpdate"`
- 既知台帳: none
- Size: M (4/8 files)
- Status: 完了 ｜ 担当レーン: LANE-3 ｜ 完了 commit: e88b28369263e3d57c19b1518f3ad143c7d0038d

##### SOLO-34

- 含む所見 ID: POC-09, POC-10
- 所有パス (3):
  - `backend/internal/clinic/clinic_holiday_repository.go`
  - `backend/internal/clinic/clinic_settings_repository_test.go`
  - `backend/internal/clinic/clinic_holiday_repository_test.go`
- 依存 / 決裁: 決裁=なし。先行 unit=BE-X06-LSTEP-SETTINGS-01, BE-X09-CLOSING-01。
- 再実測: `git grep -n -F -e '.claude/refs/backend-application-invariants.md:11' -e 'Scopes' -e 'persistence.ClinicScope' -e 'db.Where' -e 'Save' -e 'ClinicScope' HEAD -- backend/internal/clinic/clinic_holiday_repository.go backend/internal/clinic/clinic_settings_repository_test.go backend/internal/clinic/clinic_holiday_repository_test.go`
- 検証: `docker compose exec backend go test ./internal/clinic/ -run "ClinicSettings|HolidayRepository|ClosingSettingsService"`
- 既知台帳: none
- Size: M (3/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

##### SOLO-36

- 含む所見 ID: TRM-09
- 所有パス (7):
  - `backend/internal/trimming/trimming_course_repository.go`
  - `backend/internal/trimming/trimming_option_repository.go`
  - `backend/internal/trimming/trimming_course_type_repository.go`
  - `backend/internal/trimming/trimming_course_repository_test.go`
  - `backend/internal/trimming/trimming_option_repository_test.go`
  - `backend/internal/trimming/trimming_course_type_repository_test.go`
  - `backend/internal/lintscan/dbortx_inventory_lint_test.go`
- 依存 / 決裁: 決裁=DEC-39。先行 unit=なし。
- 再実測: `git grep -n -F -e 'db.WithContext' -e 'CODING_RULES' -e 'DBOrTx' -e 'UpdateScopedByID' -e 'WithContext' -e 'ReorderByClinicID' HEAD -- backend/internal/trimming/trimming_course_repository.go backend/internal/trimming/trimming_option_repository.go backend/internal/trimming/trimming_course_type_repository.go backend/internal/trimming/trimming_course_repository_test.go backend/internal/trimming/trimming_option_repository_test.go backend/internal/trimming/trimming_course_type_repository_test.go backend/internal/lintscan/dbortx_inventory_lint_test.go`
- 検証: `docker compose exec backend go test ./internal/trimming/ && docker compose exec backend go test ./internal/lintscan/ -run DBOrTx`
- 既知台帳: none
- Size: L (7/8 files)
- Status: 完了 ｜ 担当レーン: LANE-2 ｜ 完了 commit: f51f25b7435d85901d33cc9d7acccdbff891364b

##### U-X04-MR-SUBRECORDS

- 含む所見 ID: MRC-04
- 所有パス (6):
  - `backend/internal/medicalrecord/medical_record_auto_create.go`
  - `backend/internal/medicalrecord/medical_record_handler.go`
  - `backend/internal/medicalrecord/medical_record_service.go`
  - `backend/internal/medicalrecord/medical_record_auto_create_test.go`
  - `backend/internal/medicalrecord/medical_record_handler_test.go`
  - `backend/internal/medicalrecord/medical_record_service_test.go`
- 依存 / 決裁: 決裁=DEC-32, DEC-35。先行 unit=BE-X09-MEDICAL-DIAGNOSIS-01。
- 再実測: `git grep -n -F -e '.claude/refs/backend-application-invariants.md:35' -e 'auditReservationDraftCleanupFailure' -e 'CreateSubRecords' -e 'Warn' -e 'Created' -e 'POST' HEAD -- backend/internal/medicalrecord/medical_record_auto_create.go backend/internal/medicalrecord/medical_record_handler.go backend/internal/medicalrecord/medical_record_service.go backend/internal/medicalrecord/medical_record_auto_create_test.go backend/internal/medicalrecord/medical_record_handler_test.go backend/internal/medicalrecord/medical_record_service_test.go`
- 検証: `docker compose exec backend go test ./internal/medicalrecord/... -run 'SubRecord|MedicalRecord.*Create|AutoCreate'`
- 既知台帳: none
- Size: M (6/8 files)
- Status: 未着手 ｜ 担当レーン: — ｜ 完了 commit: —

### Ownership TSV

本節の所有パス台帳は統合時に機械検証済み。

```tsv
unit_id	path
U-X04-COVERAGE-RATCHET	.github/workflows/ci.yml
SOLO-04	backend/cmd/api/base_routes.go
SOLO-04	backend/cmd/api/batch_scheduler_test.go
SOLO-04	backend/cmd/api/batch_scheduler.go
BE-X06-BIL-CAMPAIGN-01	backend/cmd/api/composition_billing_services.go
BE-X07-BODY-01	backend/cmd/api/composition_core_test.go
BE-X06-MEDICAL-ATOMIC-01	backend/cmd/api/composition_medicalrecord_services.go
BE-X07-BODY-01	backend/cmd/api/composition_runtime.go
SOLO-05	backend/cmd/api/lstep_adapters.go
SOLO-04	backend/cmd/api/main.go
SOLO-04	backend/cmd/api/route_composition_smoke_test.go
U-X04-COVERAGE-RATCHET	backend/cmd/coverage-ratchet/main_test.go
U-X04-COVERAGE-RATCHET	backend/cmd/coverage-ratchet/main.go
SOLO-06	backend/cmd/csv-import-failure-rehearsal/main_test.go
SOLO-06	backend/cmd/csv-import-failure-rehearsal/main.go
U-X04-LSTEP-MIGRATE	backend/cmd/lstep-migrate/main.go
U-X04-LSTEP-MIGRATE	backend/cmd/lstep-migrate/migrator.go
U-X03-CSVIMPORT-GUARD	backend/cmd/seed-export/main.go
SOLO-33	backend/docs/api.yaml
SOLO-22	backend/internal/apperrors/errors_test.go
SOLO-22	backend/internal/apperrors/errors.go
U-X04-AUDIT-MARSHAL	backend/internal/audit/repository_test.go
U-X04-AUDIT-MARSHAL	backend/internal/audit/repository.go
U-X04-AUDIT-MARSHAL	backend/internal/audit/service_test.go
U-X04-AUDIT-MARSHAL	backend/internal/audit/service.go
BE-X10-AUTH-RESPONSE-01	backend/internal/auth/http_permission_test.go
BE-X10-AUTH-RESPONSE-01	backend/internal/auth/http_permission.go
BE-X10-AUTH-RESPONSE-01	backend/internal/auth/http_response.go
BE-X10-AUTH-RESPONSE-01	backend/internal/auth/http_session_handlers_test.go
SOLO-15	backend/internal/billing/accounting_repository_ltv.go
SOLO-20	backend/internal/billing/accounting_repository_reports_close_test.go
SOLO-20	backend/internal/billing/accounting_repository_reports_close.go
SOLO-24	backend/internal/billing/accounting_repository_tx_atomicity_test.go
SOLO-24	backend/internal/billing/accounting_repository.go
SOLO-02	backend/internal/billing/accounting_service_core.go
SOLO-03	backend/internal/billing/accounting_service_correction_test.go
SOLO-03	backend/internal/billing/accounting_service_correction.go
SOLO-03	backend/internal/billing/accounting_service_reports.go
SOLO-03	backend/internal/billing/accounting_service.go
SOLO-02	backend/internal/billing/billing_item_handler_test.go
SOLO-02	backend/internal/billing/billing_item_repository_tx_atomicity_test.go
SOLO-02	backend/internal/billing/billing_item_repository.go
SOLO-02	backend/internal/billing/billing_item_service_test.go
SOLO-02	backend/internal/billing/billing_item_service.go
BE-X06-BIL-CAMPAIGN-01	backend/internal/billing/campaign_cross_tenant_master_fk_write_test.go
BE-X06-BIL-CAMPAIGN-01	backend/internal/billing/campaign_repository.go
BE-X06-BIL-CAMPAIGN-01	backend/internal/billing/campaign_service_test.go
BE-X06-BIL-CAMPAIGN-01	backend/internal/billing/campaign_service.go
BE-X09-CLOSING-01	backend/internal/billing/cash_register_service_test.go
BE-X09-CLOSING-01	backend/internal/billing/cash_register_service.go
BE-X09-ESTIMATE-TAX-01	backend/internal/billing/estimate_request_test.go
BE-X09-ESTIMATE-TAX-01	backend/internal/billing/estimate_request.go
BE-X09-ESTIMATE-TAX-01	backend/internal/billing/estimate_response_test.go
BE-X09-ESTIMATE-TAX-01	backend/internal/billing/estimate_response.go
BE-X09-ESTIMATE-TAX-01	backend/internal/billing/estimate_service_test.go
BE-X09-ESTIMATE-TAX-01	backend/internal/billing/estimate_service.go
SOLO-02	backend/internal/billing/routes.go
SOLO-32	backend/internal/clinic/clinic_holiday_handler_test.go
SOLO-32	backend/internal/clinic/clinic_holiday_handler.go
SOLO-34	backend/internal/clinic/clinic_holiday_repository_test.go
SOLO-34	backend/internal/clinic/clinic_holiday_repository.go
U-X02-CLINIC-CONTACT	backend/internal/clinic/clinic_request_test.go
U-X02-CLINIC-CONTACT	backend/internal/clinic/clinic_request.go
U-X01X05-CLINIC	backend/internal/clinic/clinic_service_test.go
SOLO-32	backend/internal/clinic/clinic_service.go
SOLO-34	backend/internal/clinic/clinic_settings_repository_test.go
BE-X06-LSTEP-SETTINGS-01	backend/internal/clinic/clinic_settings_repository.go
SOLO-32	backend/internal/clinic/closing_settings_handler_test.go
SOLO-32	backend/internal/clinic/closing_settings_handler.go
BE-X09-CLOSING-01	backend/internal/clinic/closing_settings_service_test.go
BE-X09-CLOSING-01	backend/internal/clinic/closing_settings_service.go
U-X01X05-CLINIC	backend/internal/clinic/closing_special_period_repository_test.go
U-X01X05-CLINIC	backend/internal/clinic/closing_special_period_repository.go
U-X02-CLINIC-CONTACT	backend/internal/clinic/company_request_test.go
U-X02-CLINIC-CONTACT	backend/internal/clinic/company_request.go
U-X01X05-CLINIC	backend/internal/clinic/company_service.go
SOLO-04	backend/internal/config/config.go
U-X03-CSVIMPORT-GUARD	backend/internal/csvimport/cutover_import_test.go
U-X03-CSVIMPORT-GUARD	backend/internal/csvimport/cutover_import.go
U-X03-CSVIMPORT-GUARD	backend/internal/csvimport/failure_rehearsal_test.go
U-X03-CSVIMPORT-GUARD	backend/internal/csvimport/failure_rehearsal.go
U-X03-CSVIMPORT-GUARD	backend/internal/csvimport/import_test.go
U-X03-CSVIMPORT-GUARD	backend/internal/csvimport/import.go
BE-X10-AUTH-RESPONSE-01	backend/internal/httpapi/context_test.go
BE-X10-AUTH-RESPONSE-01	backend/internal/httpapi/context.go
SOLO-08	backend/internal/infra/lstep/client.go
SOLO-08	backend/internal/infra/lstep/tag.go
SOLO-08	backend/internal/infra/lstep/user.go
U-X01X02-INVENTORY	backend/internal/inventory/inventory_request_test.go
U-X01X02-INVENTORY	backend/internal/inventory/inventory_request.go
U-X01X02-INVENTORY	backend/internal/inventory/medicine_inventory_tx_atomicity_test.go
U-X01X02-INVENTORY	backend/internal/inventory/merchandise_item_repository_test.go
U-X01X02-INVENTORY	backend/internal/inventory/merchandise_item_repository.go
U-X01X02-INVENTORY	backend/internal/inventory/repository.go
SOLO-36	backend/internal/lintscan/dbortx_inventory_lint_test.go
SOLO-26	backend/internal/lintscan/grandchild_parent_clinic_correlation_lint_test.go
SOLO-23	backend/internal/lintscan/migration_cascade_lint_test.go
SOLO-26	backend/internal/lintscan/preload_clinic_scope_lint_test.go
U-X02-LSTEP-AGGREGATION	backend/internal/lstep/aggregation_request_test.go
U-X02-LSTEP-AGGREGATION	backend/internal/lstep/aggregation_request.go
BE-X06-LSTEP-SETTINGS-01	backend/internal/lstep/composition_services.go
SOLO-16	backend/internal/lstep/line_customer_handler_test.go
SOLO-16	backend/internal/lstep/line_customer_handler.go
SOLO-16	backend/internal/lstep/line_customer_repository_test.go
SOLO-16	backend/internal/lstep/line_customer_repository.go
U-X01-LSTEP-LINE-CUSTOMER	backend/internal/lstep/line_customer_service_test.go
U-X01-LSTEP-LINE-CUSTOMER	backend/internal/lstep/line_customer_service.go
SOLO-10	backend/internal/lstep/line_messaging_service_test.go
SOLO-10	backend/internal/lstep/line_messaging_service.go
BE-X08-LSTEP-SEND-01	backend/internal/lstep/line_send_handler_test.go
BE-X08-LSTEP-SEND-01	backend/internal/lstep/line_send_handler.go
BE-X08-LSTEP-SEND-01	backend/internal/lstep/line_send_response.go
BE-X08-LSTEP-SEND-01	backend/internal/lstep/line_send_service_test.go
BE-X08-LSTEP-SEND-01	backend/internal/lstep/line_send_service.go
SOLO-11	backend/internal/lstep/lstep_batch_delivery.go
SOLO-19	backend/internal/lstep/lstep_batch_noshow_test.go
SOLO-19	backend/internal/lstep/lstep_batch_noshow.go
U-X04-LSTEP-BATCH	backend/internal/lstep/lstep_batch_segmentation_test.go
U-X04-LSTEP-BATCH	backend/internal/lstep/lstep_batch_segmentation.go
U-X04-LSTEP-BATCH	backend/internal/lstep/lstep_batch_service_test.go
U-X04-LSTEP-BATCH	backend/internal/lstep/lstep_batch_service.go
U-X04X05-LSTEP-DELIVERY	backend/internal/lstep/lstep_delivery_monitor_service_test.go
U-X04X05-LSTEP-DELIVERY	backend/internal/lstep/lstep_delivery_monitor_service.go
U-LSTEP-OPTOUT	backend/internal/lstep/lstep_delivery_trigger_batch_test.go
U-LSTEP-OPTOUT	backend/internal/lstep/lstep_delivery_trigger_batch.go
U-X04X05-LSTEP-DELIVERY	backend/internal/lstep/lstep_delivery_trigger_client_test.go
U-X04X05-LSTEP-DELIVERY	backend/internal/lstep/lstep_delivery_trigger_client.go
SOLO-11	backend/internal/lstep/lstep_delivery_trigger_service_test.go
U-LSTEP-OPTOUT	backend/internal/lstep/lstep_delivery_trigger_state_test.go
U-LSTEP-OPTOUT	backend/internal/lstep/lstep_delivery_trigger_state.go
SOLO-11	backend/internal/lstep/lstep_delivery_trigger_suppression_test.go
SOLO-11	backend/internal/lstep/lstep_delivery_trigger_suppression.go
U-X04-LSTEP-HEALTH-REMOVE	backend/internal/lstep/lstep_health_tag_sync_batch_test.go
U-X04-LSTEP-HEALTH-REMOVE	backend/internal/lstep/lstep_health_tag_sync_batch.go
SOLO-14	backend/internal/lstep/lstep_health_tag_sync_checkup_test.go
SOLO-14	backend/internal/lstep/lstep_health_tag_sync_checkup.go
U-X04-LSTEP-HEALTH-REMOVE	backend/internal/lstep/lstep_health_tag_sync_food.go
U-X04-LSTEP-HEALTH-REMOVE	backend/internal/lstep/lstep_health_tag_sync_prevention_test.go
U-X04-LSTEP-HEALTH-REMOVE	backend/internal/lstep/lstep_health_tag_sync_prevention.go
SOLO-14	backend/internal/lstep/lstep_health_tag_sync_vaccine_test.go
SOLO-14	backend/internal/lstep/lstep_health_tag_sync_vaccine.go
SOLO-05	backend/internal/lstep/lstep_lifecycle_deps.go
U-X01X03X04-LSTEP-LIFECYCLE	backend/internal/lstep/lstep_lifecycle_handler.go
U-X01X03X04-LSTEP-LIFECYCLE	backend/internal/lstep/lstep_lifecycle_service_test.go
U-X01X03X04-LSTEP-LIFECYCLE	backend/internal/lstep/lstep_lifecycle_service.go
BE-X08-LSTEP-CONNECTION-01	backend/internal/lstep/lstep_settings_connection_test.go
BE-X08-LSTEP-CONNECTION-01	backend/internal/lstep/lstep_settings_connection.go
U-X01X03X04-LSTEP-LIFECYCLE	backend/internal/lstep/lstep_settings_credentials_test.go
BE-X08-LSTEP-CONNECTION-01	backend/internal/lstep/lstep_settings_credentials.go
BE-X08-LSTEP-CONNECTION-01	backend/internal/lstep/lstep_settings_handler_test.go
BE-X06-LSTEP-SETTINGS-01	backend/internal/lstep/lstep_settings_repository.go
BE-X08-LSTEP-CONNECTION-01	backend/internal/lstep/lstep_settings_request.go
BE-X08-LSTEP-CONNECTION-01	backend/internal/lstep/lstep_settings_response.go
BE-X06-LSTEP-SETTINGS-01	backend/internal/lstep/lstep_settings_service_test.go
BE-X06-LSTEP-SETTINGS-01	backend/internal/lstep/lstep_settings_service.go
BE-X06-LSTEP-SETTINGS-01	backend/internal/lstep/lstep_settings_update_test.go
BE-X06-LSTEP-SETTINGS-01	backend/internal/lstep/lstep_settings_update.go
BE-X06-LSTEP-SETTINGS-01	backend/internal/lstep/lstep_sync_settings_repository.go
SOLO-12	backend/internal/lstep/lstep_tag_cache_repository.go
U-X02-LSTEP-TAG-MAPPING	backend/internal/lstep/lstep_tag_code_mapping_request.go
U-X02-LSTEP-TAG-MAPPING	backend/internal/lstep/lstep_tag_code_mapping_service_test.go
U-X02-LSTEP-TAG-MAPPING	backend/internal/lstep/lstep_tag_code_mapping_service.go
U-X02-LSTEP-TAG-CONFIG	backend/internal/lstep/lstep_tag_config_handler_test.go
U-X02-LSTEP-TAG-CONFIG	backend/internal/lstep/lstep_tag_config_handler.go
SOLO-09	backend/internal/lstep/lstep_tag_config_repository_test.go
SOLO-09	backend/internal/lstep/lstep_tag_config_repository.go
U-X02-LSTEP-TAG-CONFIG	backend/internal/lstep/lstep_tag_config_request_test.go
U-X02-LSTEP-TAG-CONFIG	backend/internal/lstep/lstep_tag_config_request.go
U-X02-LSTEP-TAG-CONFIG	backend/internal/lstep/lstep_tag_config_service_test.go
U-X02-LSTEP-TAG-CONFIG	backend/internal/lstep/lstep_tag_config_service.go
U-X04-LSTEP-OWNER-TAGS	backend/internal/lstep/lstep_tag_handler_test.go
U-X04-LSTEP-OWNER-TAGS	backend/internal/lstep/lstep_tag_handler.go
U-X04-LSTEP-OWNER-TAGS	backend/internal/lstep/lstep_tag_service_test.go
U-X04-LSTEP-OWNER-TAGS	backend/internal/lstep/lstep_tag_service.go
SOLO-13	backend/internal/lstep/lstep_tag_summary_handler_test.go
SOLO-13	backend/internal/lstep/lstep_tag_summary_handler.go
SOLO-13	backend/internal/lstep/lstep_tag_summary_service_test.go
SOLO-13	backend/internal/lstep/lstep_tag_summary_service.go
U-X04-LSTEP-STALE-TAGS	backend/internal/lstep/lstep_tag_sync_api_test.go
U-X04-LSTEP-STALE-TAGS	backend/internal/lstep/lstep_tag_sync_api.go
U-LSTEP-OPTOUT	backend/internal/lstep/lstep_tag_sync_pet_exclusion_test.go
U-LSTEP-OPTOUT	backend/internal/lstep/lstep_tag_sync_pet_exclusion.go
SOLO-15	backend/internal/lstep/lstep_tag_sync_visit_ltv_test.go
SOLO-15	backend/internal/lstep/lstep_tag_sync_visit_ltv.go
U-X02-LSTEP-TRIGGER-PRIORITY	backend/internal/lstep/lstep_trigger_priority_handler_test.go
U-X02-LSTEP-TRIGGER-PRIORITY	backend/internal/lstep/lstep_trigger_priority_handler.go
U-X02-LSTEP-TRIGGER-PRIORITY	backend/internal/lstep/lstep_trigger_priority_request.go
U-X02-LSTEP-TRIGGER-PRIORITY	backend/internal/lstep/lstep_trigger_priority_service_test.go
U-X02-LSTEP-TRIGGER-PRIORITY	backend/internal/lstep/lstep_trigger_priority_service.go
SOLO-09	backend/internal/lstep/routes.go
U-X02-LSTEP-SHARED-FILE	backend/internal/lstep/shared_file_handler_test.go
U-X02-LSTEP-SHARED-FILE	backend/internal/lstep/shared_file_handler.go
SOLO-17	backend/internal/lstep/shared_file_repository_test.go
SOLO-17	backend/internal/lstep/shared_file_repository.go
U-X02-LSTEP-SHARED-FILE	backend/internal/lstep/shared_file_request_test.go
U-X02-LSTEP-SHARED-FILE	backend/internal/lstep/shared_file_request.go
SOLO-17	backend/internal/lstep/shared_file_service_test.go
SOLO-17	backend/internal/lstep/shared_file_service.go
U-X01X03-MANUALARTICLE	backend/internal/manualarticle/handler_test.go
U-X01X03-MANUALARTICLE	backend/internal/manualarticle/handler.go
U-X01X03-MANUALARTICLE	backend/internal/manualarticle/repository_test.go
U-X01X03-MANUALARTICLE	backend/internal/manualarticle/repository.go
BE-X07-BODY-01	backend/internal/manualarticle/request_test.go
BE-X07-BODY-01	backend/internal/manualarticle/request.go
U-X01X03-MANUALARTICLE	backend/internal/manualarticle/service_test.go
U-X01X03-MANUALARTICLE	backend/internal/manualarticle/service.go
SOLO-25	backend/internal/medicalrecord/cage_handler.go
SOLO-25	backend/internal/medicalrecord/cage_repository.go
SOLO-25	backend/internal/medicalrecord/cage_request.go
SOLO-25	backend/internal/medicalrecord/cage_service.go
U-X01X03-MR-CARE	backend/internal/medicalrecord/care_plan_item_repository_test.go
U-X01X03-MR-CARE	backend/internal/medicalrecord/care_plan_item_repository.go
U-X01X03-MR-CARE	backend/internal/medicalrecord/care_plan_item_service_test.go
U-X01X03-MR-CARE	backend/internal/medicalrecord/care_plan_item_service.go
BE-X09-MEDICAL-DIAGNOSIS-01	backend/internal/medicalrecord/clinical_plan_service_test.go
BE-X09-MEDICAL-DIAGNOSIS-01	backend/internal/medicalrecord/clinical_plan_service.go
SOLO-25	backend/internal/medicalrecord/consultation_handler.go
SOLO-21	backend/internal/medicalrecord/consultation_repository_test.go
SOLO-21	backend/internal/medicalrecord/consultation_repository.go
U-X02-MR-CONSULTATION	backend/internal/medicalrecord/consultation_request_test.go
U-X02-MR-CONSULTATION	backend/internal/medicalrecord/consultation_request.go
SOLO-25	backend/internal/medicalrecord/consultation_service.go
BE-X06-MEDICAL-ATOMIC-01	backend/internal/medicalrecord/cross_tenant_master_fk_write_test.go
U-X05-MR-EXAMTYPE	backend/internal/medicalrecord/exam_type_field.go
U-X05-MR-EXAMTYPE	backend/internal/medicalrecord/exam_type_repository_test.go
U-X05-MR-EXAMTYPE	backend/internal/medicalrecord/exam_type_repository.go
U-X05-MR-EXAMTYPE	backend/internal/medicalrecord/exam_type_service_test.go
U-X05-MR-EXAMTYPE	backend/internal/medicalrecord/exam_type_service.go
U-X02X03X05-MR-HOSPITALIZATION	backend/internal/medicalrecord/hospitalization_repository_test.go
U-X02X03X05-MR-HOSPITALIZATION	backend/internal/medicalrecord/hospitalization_repository.go
U-X02X03X05-MR-HOSPITALIZATION	backend/internal/medicalrecord/hospitalization_request_test.go
U-X02X03X05-MR-HOSPITALIZATION	backend/internal/medicalrecord/hospitalization_request.go
U-X02X03X05-MR-HOSPITALIZATION	backend/internal/medicalrecord/hospitalization_service_test.go
U-X02X03X05-MR-HOSPITALIZATION	backend/internal/medicalrecord/hospitalization_service.go
BE-X06-MEDICAL-ATOMIC-01	backend/internal/medicalrecord/inquiry_repository_test.go
BE-X06-MEDICAL-ATOMIC-01	backend/internal/medicalrecord/inquiry_repository.go
BE-X06-MEDICAL-ATOMIC-01	backend/internal/medicalrecord/inquiry_service_test.go
BE-X06-MEDICAL-ATOMIC-01	backend/internal/medicalrecord/inquiry_service.go
BE-X06-MEDICAL-ATOMIC-01	backend/internal/medicalrecord/lab_import_examination_service_test.go
BE-X06-MEDICAL-ATOMIC-01	backend/internal/medicalrecord/lab_import_examination_service.go
U-X02-MR-LAB-IMPORT	backend/internal/medicalrecord/lab_import_request.go
U-X04-MR-SUBRECORDS	backend/internal/medicalrecord/medical_record_auto_create_test.go
U-X04-MR-SUBRECORDS	backend/internal/medicalrecord/medical_record_auto_create.go
U-X04-MR-SUBRECORDS	backend/internal/medicalrecord/medical_record_handler_test.go
U-X04-MR-SUBRECORDS	backend/internal/medicalrecord/medical_record_handler.go
SOLO-29	backend/internal/medicalrecord/medical_record_image_handler_test.go
SOLO-29	backend/internal/medicalrecord/medical_record_image_handler.go
SOLO-29	backend/internal/medicalrecord/medical_record_image_request_test.go
SOLO-29	backend/internal/medicalrecord/medical_record_image_request.go
SOLO-29	backend/internal/medicalrecord/medical_record_image_service_test.go
SOLO-29	backend/internal/medicalrecord/medical_record_image_service.go
SOLO-15	backend/internal/medicalrecord/medical_record_owner_visit_repository.go
U-X04-MR-SUBRECORDS	backend/internal/medicalrecord/medical_record_service_test.go
U-X04-MR-SUBRECORDS	backend/internal/medicalrecord/medical_record_service.go
BE-X09-MEDICAL-DIAGNOSIS-01	backend/internal/medicalrecord/medical_record_subrecords_test.go
BE-X09-MEDICAL-DIAGNOSIS-01	backend/internal/medicalrecord/medical_record_subrecords.go
U-X05-MR-MEDICINE-MASTERS	backend/internal/medicalrecord/medicine_dose_param_service_test.go
U-X05-MR-MEDICINE-MASTERS	backend/internal/medicalrecord/medicine_dose_param_service.go
U-X05-MR-MEDICINE-MASTERS	backend/internal/medicalrecord/medicine_service_test.go
U-X05-MR-MEDICINE-MASTERS	backend/internal/medicalrecord/medicine_service.go
U-X01-MR-PRESCRIPTION	backend/internal/medicalrecord/prescription_repository_test.go
U-X01-MR-PRESCRIPTION	backend/internal/medicalrecord/prescription_repository.go
U-X01-MR-PRESCRIPTION	backend/internal/medicalrecord/prescription_service_test.go
U-X01-MR-PRESCRIPTION	backend/internal/medicalrecord/prescription_service.go
U-X05-MR-MEDICINE-MASTERS	backend/internal/medicalrecord/procedure_repository_test.go
U-X05-MR-MEDICINE-MASTERS	backend/internal/medicalrecord/procedure_repository.go
U-X05-MR-MEDICINE-MASTERS	backend/internal/medicalrecord/procedure_service_test.go
U-X05-MR-MEDICINE-MASTERS	backend/internal/medicalrecord/procedure_service.go
U-MR-TREATMENT-PLAN	backend/internal/medicalrecord/treatment_plan_handler_test.go
U-MR-TREATMENT-PLAN	backend/internal/medicalrecord/treatment_plan_handler.go
U-MR-TREATMENT-PLAN	backend/internal/medicalrecord/treatment_plan_repository_test.go
U-MR-TREATMENT-PLAN	backend/internal/medicalrecord/treatment_plan_repository.go
U-MR-TREATMENT-PLAN	backend/internal/medicalrecord/treatment_plan_request_test.go
U-MR-TREATMENT-PLAN	backend/internal/medicalrecord/treatment_plan_request.go
U-MR-TREATMENT-PLAN	backend/internal/medicalrecord/treatment_plan_service_test.go
U-MR-TREATMENT-PLAN	backend/internal/medicalrecord/treatment_plan_service.go
SOLO-30	backend/internal/medicalrecord/treatment_repository_test.go
SOLO-30	backend/internal/medicalrecord/treatment_repository.go
SOLO-30	backend/internal/medicalrecord/treatment_service_test.go
SOLO-30	backend/internal/medicalrecord/treatment_service.go
BE-X07-BODY-01	backend/internal/middleware/sanitize_null_bytes_test.go
BE-X07-BODY-01	backend/internal/middleware/sanitize_null_bytes.go
SOLO-24	backend/internal/model/accounting.go
U-X01X03-MR-CARE	backend/internal/model/audit_log_test.go
U-X01X03-MR-CARE	backend/internal/model/audit_log.go
BE-X09-ESTIMATE-TAX-01	backend/internal/model/estimate_test.go
BE-X09-ESTIMATE-TAX-01	backend/internal/model/estimate.go
U-X02-LSTEP-SHARED-FILE	backend/internal/model/shared_file.go
SOLO-33	backend/internal/owner/http_owner.go
U-X02-PET-OWNER-FREETEXT	backend/internal/owner/http_request_test.go
U-X02-PET-OWNER-FREETEXT	backend/internal/owner/http_request.go
SOLO-33	backend/internal/owner/http_routes_test.go
SOLO-33	backend/internal/owner/http_routes.go
U-X05-OWNER-PHONE	backend/internal/owner/repository_test.go
U-X05-OWNER-PHONE	backend/internal/owner/repository.go
U-X05-OWNER-PHONE	backend/internal/owner/service_core_test.go
U-X05-OWNER-PHONE	backend/internal/owner/service_core.go
SOLO-08	backend/internal/owner/service_line.go
U-X02-CLINIC-CONTACT	backend/internal/owner/validators_contact.go
BE-X09-PET-ENUMS-01	backend/internal/owner/validators_test.go
BE-X09-PET-ENUMS-01	backend/internal/owner/validators.go
U-X03-PET-SPECIES-AUDIT	backend/internal/pet/animal_species_handler_test.go
U-X03-PET-SPECIES-AUDIT	backend/internal/pet/animal_species_handler.go
U-X03-PET-SPECIES-AUDIT	backend/internal/pet/animal_species_repository_test.go
U-X03-PET-SPECIES-AUDIT	backend/internal/pet/animal_species_repository.go
U-X03-PET-SPECIES-AUDIT	backend/internal/pet/animal_species_service_test.go
U-X03-PET-SPECIES-AUDIT	backend/internal/pet/animal_species_service.go
BE-X09-PET-PATCH-01	backend/internal/pet/chronic_condition_service_test.go
BE-X09-PET-PATCH-01	backend/internal/pet/chronic_condition_service.go
U-X05-PET-UPDATE	backend/internal/pet/owner_registration_test.go
U-X05-PET-UPDATE	backend/internal/pet/owner_registration.go
U-X02-PET-OWNER-FREETEXT	backend/internal/pet/pet_request_test.go
U-X02-PET-OWNER-FREETEXT	backend/internal/pet/pet_request.go
U-X03-PET-SPECIES-AUDIT	backend/internal/pet/ports.go
U-X05-PET-UPDATE	backend/internal/pet/repository_test.go
U-X05-PET-UPDATE	backend/internal/pet/repository.go
U-X05-PET-UPDATE	backend/internal/pet/service_test.go
U-X05-PET-UPDATE	backend/internal/pet/service.go
BE-X09-PET-ENUMS-01	backend/internal/pet/validators_test.go
BE-X09-PET-ENUMS-01	backend/internal/pet/validators.go
SOLO-18	backend/internal/reservation/appointment_admin_repository_test.go
SOLO-18	backend/internal/reservation/appointment_admin_repository.go
U-X01X05-RESERVATION	backend/internal/reservation/appointment_admin_service_test.go
U-X01X05-RESERVATION	backend/internal/reservation/appointment_admin_service.go
BE-X06-RSV-CANCEL-01	backend/internal/reservation/appointment_service_test.go
U-X02-RESERVATION-SETTINGS	backend/internal/reservation/available_dates_test.go
U-X02-RESERVATION-SETTINGS	backend/internal/reservation/available_dates.go
U-X02-RESERVATION-SETTINGS	backend/internal/reservation/liff_service_availability.go
U-X04-RESERVATION-AUTODELEGATE	backend/internal/reservation/liff_service_reservations_test.go
U-X04-RESERVATION-AUTODELEGATE	backend/internal/reservation/liff_service_reservations.go
U-X02-RESERVATION-SETTINGS	backend/internal/reservation/line_reservation_setting_request_test.go
U-X02-RESERVATION-SETTINGS	backend/internal/reservation/line_reservation_setting_request.go
U-X01X05-RESERVATION	backend/internal/reservation/line_reservation_setting_service.go
SOLO-19	backend/internal/reservation/reservation_repository_test.go
U-X01X05-RESERVATION	backend/internal/reservation/reservation_repository.go
SOLO-26	backend/internal/reservation/reservation_schedule_repository_test.go
SOLO-26	backend/internal/reservation/reservation_schedule_repository.go
BE-X06-RSV-CANCEL-01	backend/internal/reservation/reservation_service.go
U-X01X05-RESERVATION	backend/internal/reservation/reservation_type_liff_service_test.go
U-X01X05-RESERVATION	backend/internal/reservation/reservation_type_liff_service.go
U-X01X05-RESERVATION	backend/internal/reservation/reservation_type_repository.go
SOLO-04	backend/internal/scheduler/handler.go
BE-X09-PET-ENUMS-01	backend/internal/sharedkernel/enum_validators.go
BE-X09-PET-ENUMS-01	backend/internal/sharedkernel/sharedkernel_test.go
BE-X09-CLOSING-01	backend/internal/sharedkernel/shift_times.go
U-X03-STAFF-ASSIGNMENT-AUDIT	backend/internal/staff/handler.go
SOLO-01	backend/internal/staff/shift_entry_repository.go
SOLO-01	backend/internal/staff/shift_entry_service.go
SOLO-01	backend/internal/staff/shift_response_test.go
SOLO-01	backend/internal/staff/shift_response.go
SOLO-01	backend/internal/staff/shift_template_repository_integration_test.go
SOLO-01	backend/internal/staff/shift_template_repository.go
SOLO-01	backend/internal/staff/shift_template_service_test.go
SOLO-01	backend/internal/staff/shift_template_service.go
U-X03-STAFF-ASSIGNMENT-AUDIT	backend/internal/staff/staff_clinic_assignment_service_test.go
U-X03-STAFF-ASSIGNMENT-AUDIT	backend/internal/staff/staff_clinic_assignment_service.go
U-X03-STAFF-ASSIGNMENT-AUDIT	backend/internal/staff/staff_handler.go
U-X02-STAFF-TYPE	backend/internal/staff/staff_request_test.go
U-X02-STAFF-TYPE	backend/internal/staff/staff_request.go
U-X02-STAFF-TYPE	backend/internal/staff/staff_service_account_test.go
U-X02-STAFF-TYPE	backend/internal/staff/staff_service_account.go
U-X02-STAFF-TYPE	backend/internal/staff/staff_service_builders_test.go
U-X02-STAFF-TYPE	backend/internal/staff/staff_service_builders.go
U-X02-STAFF-TYPE	backend/internal/staff/staff_service_core_test.go
U-X02-STAFF-TYPE	backend/internal/staff/staff_service_core.go
U-X03-STAFF-ASSIGNMENT-AUDIT	backend/internal/staff/staff_service_permissions_test.go
U-X03-STAFF-ASSIGNMENT-AUDIT	backend/internal/staff/staff_service_permissions.go
SOLO-36	backend/internal/trimming/trimming_course_repository_test.go
SOLO-36	backend/internal/trimming/trimming_course_repository.go
SOLO-36	backend/internal/trimming/trimming_course_type_repository_test.go
SOLO-36	backend/internal/trimming/trimming_course_type_repository.go
U-TRIMMING-SERVICE	backend/internal/trimming/trimming_handler_test.go
U-TRIMMING-SERVICE	backend/internal/trimming/trimming_handler.go
SOLO-36	backend/internal/trimming/trimming_option_repository_test.go
SOLO-36	backend/internal/trimming/trimming_option_repository.go
SOLO-26	backend/internal/trimming/trimming_repository_test.go
SOLO-26	backend/internal/trimming/trimming_repository.go
BE-X07-BODY-01	backend/internal/trimming/trimming_request_test.go
BE-X07-BODY-01	backend/internal/trimming/trimming_request.go
U-TRIMMING-SERVICE	backend/internal/trimming/trimming_service_test.go
U-TRIMMING-SERVICE	backend/internal/trimming/trimming_service.go
U-SCHEMA-BARRIER	backend/migrations/001_init.sql
```

### Finding TSV

本節の所見割当台帳は統合時に機械検証済み。

```tsv
finding_id	unit_id
AUS-01	U-X03-STAFF-ASSIGNMENT-AUDIT
AUS-03	U-X02-STAFF-TYPE
AUS-04	SOLO-01
AUS-05	SOLO-01
AUS-09	BE-X10-AUTH-RESPONSE-01
BIL-01	SOLO-02
BIL-02	SOLO-03
BIL-03	BE-X06-BIL-CAMPAIGN-01
CMD-02	SOLO-04
CMD-03	U-X04-COVERAGE-RATCHET
CMD-04	SOLO-05
CMD-05	SOLO-04
CMD-06	SOLO-06
CMD-07	U-X04-LSTEP-MIGRATE
G2A-01	U-X01-LSTEP-LINE-CUSTOMER
G2A-05	U-X02-LSTEP-AGGREGATION
G2B-01	U-X04-LSTEP-HEALTH-REMOVE
G2B-02	BE-X06-LSTEP-SETTINGS-01
G2B-03	SOLO-11
G2C-01	SOLO-12
G2C-02	U-X02-LSTEP-TAG-MAPPING
G2C-04	SOLO-13
G2F-01	SOLO-12
G2F-02	SOLO-14
G2F-03	SOLO-15
G2F-04	SOLO-15
G2F-05	SOLO-16
G2F-06	SOLO-17
G2F-07	SOLO-18
G2F-08	SOLO-19
G2F-09	SOLO-20
G2F-10	SOLO-01
G2F-11	SOLO-21
G2P-02	U-X01X02-INVENTORY
G2P-03	U-X01X02-INVENTORY
INF-02	BE-X07-BODY-01
INF-03	SOLO-08
INF-04	SOLO-22
INF-06	U-X04-AUDIT-MARSHAL
LSA-01	BE-X08-LSTEP-CONNECTION-01
LSA-02	U-LSTEP-OPTOUT
LSA-03	U-X04-LSTEP-BATCH
LSA-04	SOLO-09
LSA-05	U-X01X03X04-LSTEP-LIFECYCLE
LSA-06	BE-X06-LSTEP-SETTINGS-01
LSA-07	SOLO-10
LSA-08	BE-X08-LSTEP-CONNECTION-01
LSA-09	BE-X08-LSTEP-SEND-01
LSA-10	U-X04-LSTEP-OWNER-TAGS
LSA-11	U-X04-LSTEP-STALE-TAGS
LSA-12	U-X04X05-LSTEP-DELIVERY
LSA-13	U-X02-LSTEP-TRIGGER-PRIORITY
LSA-14	U-X02-LSTEP-TAG-CONFIG
LSA-15	U-X04X05-LSTEP-DELIVERY
LSB-01	U-LSTEP-OPTOUT
LSB-02	U-X01X03X04-LSTEP-LIFECYCLE
LSB-03	U-X01X03X04-LSTEP-LIFECYCLE
LSB-04	U-X01X03X04-LSTEP-LIFECYCLE
LSB-06	U-X02-LSTEP-SHARED-FILE
MDL-01	BE-X09-ESTIMATE-TAX-01
MDL-05	SOLO-23
MDL-06	SOLO-24
MRA-01	U-X01X03-MR-CARE
MRA-02	U-X01X03-MR-CARE
MRA-03	U-X02-MR-CONSULTATION
MRA-04	SOLO-25
MRB-02	U-X02X03X05-MR-HOSPITALIZATION
MRB-03	U-X02X03X05-MR-HOSPITALIZATION
MRB-05	U-X02X03X05-MR-HOSPITALIZATION
MRB-06	U-X02X03X05-MR-HOSPITALIZATION
MRB-07	U-X05-MR-EXAMTYPE
MRB-08	U-X05-MR-EXAMTYPE
MRC-01	U-X01-MR-PRESCRIPTION
MRC-02	U-X05-MR-MEDICINE-MASTERS
MRC-03	U-X05-MR-MEDICINE-MASTERS
MRC-04	U-X04-MR-SUBRECORDS
MRC-05	BE-X06-MEDICAL-ATOMIC-01
MRC-07	U-X05-MR-MEDICINE-MASTERS
MRC-08	U-X02-MR-LAB-IMPORT
MRC-09	SOLO-29
MRC-12	BE-X06-MEDICAL-ATOMIC-01
MRC-14	BE-X09-MEDICAL-DIAGNOSIS-01
MRD-01	SOLO-30
MRD-02	U-MR-TREATMENT-PLAN
MRD-03	U-MR-TREATMENT-PLAN
MRD-04	U-MR-TREATMENT-PLAN
POC-01	SOLO-32
POC-02	U-X01X05-CLINIC
POC-03	U-X05-PET-UPDATE
POC-05	U-X01X05-CLINIC
POC-06	U-X05-OWNER-PHONE
POC-07	U-X03-PET-SPECIES-AUDIT
POC-08	SOLO-33
POC-09	SOLO-34
POC-10	SOLO-34
POC-11	BE-X09-PET-ENUMS-01
POC-12	BE-X07-BODY-01
POC-13	U-X02-PET-OWNER-FREETEXT
POC-14	BE-X09-PET-PATCH-01
POC-15	BE-X09-CLOSING-01
POC-16	BE-X09-CLOSING-01
POC-17	U-X02-CLINIC-CONTACT
RSV-02	U-X01X05-RESERVATION
RSV-03	U-X01X05-RESERVATION
RSV-04	U-X02-RESERVATION-SETTINGS
RSV-06	BE-X06-RSV-CANCEL-01
RSV-07	U-X01X05-RESERVATION
RSV-08	SOLO-26
RSV-09	U-X04-RESERVATION-AUTODELEGATE
TRM-01	U-X01X03-MANUALARTICLE
TRM-02	U-X01X03-MANUALARTICLE
TRM-03	BE-X07-BODY-01
TRM-04	U-TRIMMING-SERVICE
TRM-05	BE-X07-BODY-01
TRM-06	U-TRIMMING-SERVICE
TRM-07	U-X03-CSVIMPORT-GUARD
TRM-08	SOLO-26
TRM-09	SOLO-36
```

### Withdrawn exclusion ledger

Exactly 23 withdrawn IDs are excluded: AUS-06, CMD-01, G2A-02, G2A-03, G2A-04, G2A-06, G2A-07, G2B-04, G2B-05, G2C-03, G2C-05, G2C-06, G2C-07, G2C-08, G2P-01, G2T-02, G2T-03, INF-01, LSA-16, LSB-05, MDL-02, MDL-04, TRM-10.

### Machine checks

Run from repository root:

```bash
tail -n +2 /tmp/integrated-findings.tsv | cut -f1 | sort > /tmp/integrated-found-ids.txt
diff -u /tmp/agent-fast-be-live-ids.txt /tmp/integrated-found-ids.txt
tail -n +2 /tmp/integrated-findings.tsv | cut -f1 | sort | uniq -d
tail -n +2 /tmp/integrated-ownership.tsv | cut -f2 | sort | uniq -d
tail -n +2 /tmp/integrated-ownership.tsv | cut -f2 | while IFS= read -r owned_path; do git cat-file -e "HEAD:$owned_path" || printf '%s\n' "$owned_path"; done
awk -F '\t' 'NR>1 { count[$1]++ } END { for (unit in count) if (count[unit] > 8) print unit, count[unit] }' /tmp/integrated-ownership.tsv
awk '/^### U-X05-MR-EXAMTYPE$/{inside=1; next} /^### /{inside=0} inside' /tmp/integrated-plan.md | grep -F 'docker compose exec backend go test ./internal/lintscan/ -run DBOrTx'
```

Expected output for the first six structural checks is empty; the DBOrTx assertion must print the MRB-08 verification line.

##### Executed exact outputs

```text
CHECK_LIVE_DIFF
<empty>
CHECK_DUP_FINDINGS
<empty>
CHECK_DUP_PATHS
<empty>
CHECK_MISSING_PATHS
<empty>
CHECK_OVER_8
<empty>
CHECK_WITHDRAWN_INTERSECTION
<empty>
CHECK_COUNTS
live=118 assigned=118 unique_assigned=118 withdrawn=23 ownership_rows=382 unique_paths=382 max_unit_paths=8
CHECK_MRB08_DBORTX
- Scoped Docker verification: `docker compose exec backend go test ./internal/medicalrecord/... -run 'ExamType' && docker compose exec backend go test ./internal/lintscan/ -run DBOrTx`
CHECK_REPO_STATUS (BE-refactor.md, q&a.html, backend)
<empty>
```

### Boundary reconciliation notes

- `backend/cmd/api/composition_runtime.go` and `composition_core_test.go` are owned only by `BE-X07-BODY-01`; `SOLO-05` depends on that composition owner.
- The delivery-trigger batch/state implementation and tests are owned only by `U-LSTEP-OPTOUT`; delivery error/performance units depend on it.
- The examination-type field/service/repository paths are owned only by `U-X05-MR-EXAMTYPE`; MRB-07 is merged there and G2F-11 remains a sequential continuation.
- The trimming service implementation and test are owned only by `U-TRIMMING-SERVICE`; body/request work is isolated in `BE-X07-BODY-01`.
- `backend/migrations/001_init.sql` is owned only by `U-SCHEMA-BARRIER`; no implementation unit invents an unallocated incremental migration filename.
- Every new repository participant named in an atomic unit carries `docker compose exec backend go test ./internal/lintscan/ -run DBOrTx`.
