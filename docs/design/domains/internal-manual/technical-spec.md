# D22 技術接続候補

業務要件の正本：[院内マニュアル](../../../spec/domains/internal-manual/README.md)。基準SHA `41fcf3eb0a8befb7c0b5f613d1df749a60e30be8`。実装変更なし、実機未検証。

## 対応と所有者

- [docs/spec/screens/35-internal-manual.md](../../../spec/screens/35-internal-manual.md)：現行の仕様／実装／認可根拠。API・schemaはここに複製しない。
- [frontend/src/features/manual/routes/ManualPage.tsx](../../../../frontend/src/features/manual/routes/ManualPage.tsx)：現行の仕様／実装／認可根拠。API・schemaはここに複製しない。
- [backend/internal/manualarticle/service.go](../../../../backend/internal/manualarticle/service.go)：現行の仕様／実装／認可根拠。API・schemaはここに複製しない。

## 受入trace

| 条件 | 技術確認の観点 |
|:---|:---|
| AC-internal-manual-01 | 画面別・業務別検索から記事を開き印刷時、選択した記事の内容を閲覧・印刷できることを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-internal-manual-02 | 直接記事URLを開く時、未発見と表示することを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-internal-manual-03 | 公開済み改訂記事を読む／編集を試す時、最新版閲覧は可、編集は拒否ことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-internal-manual-04 | 保存する時、公開成功を表示せず入力を保持ことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-internal-manual-05 | 過去版を明示復元時、新しい版として保存し旧履歴を保持ことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-internal-manual-06 | 一般スタッフが記事を検索・印刷時、未確定の編集案は公開済み本文として表示しないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-internal-manual-07 | 古い表示版から保存時、競合を検出して確認を要求し新しい版を無警告に上書きしない（技術提案）ことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-internal-manual-08 | 削除する時、静的版への復帰か追加記事の非表示かという現行の結果を区別し、公開・編集・復元履歴を無断で抹消しない。具体UIは設計ことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |

## 設計・検証残件

公開版全職員閲覧・編集／公開／過去版復元の履歴は今回合意済み。独立した下書き承認workflow、医院別記事、既読必須を今回新設しない。旧実装の権限付きoverride閲覧との差分、公開版と検索／印刷の一貫性、競合時の原子的保存は末尾で選定し、履歴UI実装は検証の残件であり業務回答待ちではない。

- 保存前後で対象医院・対象人物・実効権限を拘束する。権限や状態が競合したら最新状態を確認して再操作し、別医院への自動再送を行わない。
- 必須の履歴／監査があるwriteは対応する既存ownerの原子的保存契約へ接続する。全CUDに同一audit要件があるとは仮定しない。
- UI表示／API拒否／DB保存／外部送信の成功は別々に確認し、ソース静的照合だけで実機PASSにしない。
- 公開READと編集WRITEは分離する（FR-internal-manual-publication）。認証済みスタッフ全員が公開版を読み、GET list/GET {category}/{slug} は権限 middleware なし・編集系のみ ResourceManualEdit。Frontend は strict fetch（pending/error は query state で gate、bundled fallback なし）で検索・印刷にも同じ公開版を用いる。記事は現行serviceで医院共通。旧版復元UIの実装存在は未確認。

<a id="manual-selected-save-contract"></a>
## 選定済みの公開・履歴・結果照合契約

業務判断の追加ではなく、公開版閲覧と履歴保持の既決を成立させる技術契約。ownerはmanualarticle。記事の安定ID（category/slugの現行識別との対応）、現在版、公開版、内容digest、操作者、操作ID、操作種別（編集保存／公開／過去版復元／非表示）を区別する。新規articleも操作IDから生成済み対象を照合できるようにする。物理名や新endpointは本書では定義しない。

- 保存要求は対象ID・期待現在版・編集内容／digest・公開対象版または復元元版・操作IDを拘束する。現在の編集権限をowner境界で再検査する。医院共通記事を任意の医院専用データと推測しない。非公開編集案は公開READへ返さない。
- 編集保存は新しい内容版と編集履歴、公開は公開対象内容版と公開pointer／公開履歴、復元は保持済み元版から作る新しい内容版と復元元参照を保存する。公開確定を伴う操作では公開pointerも同じcommitに参加する。各操作の必須監査（操作者・対象・操作・前後版・時刻）と結果receiptは、その業務変更と同一owner transactionで成立する。いずれか失敗なら全変更をrollbackし、旧公開版・旧履歴は不変とする。
- 同時変更は期待版一致をcommit条件にする。版不一致・権限失効は拒否し、旧公開版を変更しない。画面は編集案と元版を保持し最新内容を別に取得、比較・再確認後に新しい操作を作る。最新読取で入力を上書きしない。
- 同じ操作ID・同じ内容／対象は既存receiptを返し新しい版・履歴・公開を重複作成しない。同じIDで異なる内容は拒否。応答不明は同じ操作IDでreceipt／現在版を照合し、新IDの再保存・復元で救済しない。照合不能なら未確定として入力を保持し停止する。UI成功はreceiptの対象版と一致した時だけ示す。
- 公開READ・検索・印刷は同じ公開版IDに束ねる。公開処理中や失敗時は直前の確定公開版を保持し、未公開内容を混ぜない。記事不存在と取得障害を分け、障害時に静的版へ黙って戻して最新版として表示しない。
- 旧静的版／DB overrideの出所と版を識別するadapterを置く。公開済みか識別できない旧overrideを全員へ自動公開しない。変更前内容／復元元版を再構成できない操作は停止して内容確認へ戻す。旧実装の権限付きoverride READや履歴APIの存在を、この目標の成立証拠としない。
- 非表示・override解除・復元も履歴を保持する。表示から外す操作と物理廃棄を分離し、C05の保持境界へ従う。履歴原本を削って静的版へ復帰する方式は採用しない。

上記の意味ID・版・原子的保存・競合・receipt・旧値停止は選定済み。物理schema／実装方法・runtime検証だけを後続工程へ残す。

### 追加AC trace

| 正本条件 | 選定契約への対応 |
|:---|:---|
| AC-internal-manual-09 | [業務正本の追加受入条件](../../../spec/domains/internal-manual/README.md)と本書の操作ID／原子保存・競合・結果照合・保持条件を照合。実機未実施 |
| AC-internal-manual-10 | [業務正本の追加受入条件](../../../spec/domains/internal-manual/README.md)と本書の操作ID／原子保存・競合・結果照合・保持条件を照合。実機未実施 |
| AC-internal-manual-11 | [業務正本の追加受入条件](../../../spec/domains/internal-manual/README.md)と本書の操作ID／原子保存・競合・結果照合・保持条件を照合。実機未実施 |
| AC-internal-manual-12 | [業務正本の追加受入条件](../../../spec/domains/internal-manual/README.md)と本書の操作ID／原子保存・競合・結果照合・保持条件を照合。実機未実施 |
| AC-internal-manual-multipage-print | 同じ公開版の選択記事を複数ページで表示しブラウザ印刷／PDF保存する際、本文・説明用スクリーンショットが最終ページまで出力され先頭1ページだけに欠落しないこと、別記事・未公開編集案を混ぜないことを確認。実機未実施 |

<a id="manual-edit-adopted-physical-contract"></a>
## 採用済みの物理契約（EMR-502 既存記事edit）

本PRの実装差分の実行検証は **NOT_RUN**（go/test/build/lint/codegen 未実行）。
正本は `docs/work/emr/emr-502-manual-edit-contract-20261021.md`（§4 immutable content/hash projection）。
`ManualArticleEditContent` の wire 契約どおり title/section は非空のみ・body_markdown は空を許容し、
旧 binding の `max=255`/`max=100000` 文字数制限・handler 独自 body 上限は継承しない（raw 上限は
global sanitizer と同じ 1 MiB・manual edit 専用 raw body guard を sanitizer より前に登録、
raw 0x00 / 不正 JSON は claim 前 400、超過は 413、検査済み bytes は書換えず後段へ戻す）。
decoded U+0000 は title/section/body_markdown の保存対象文字列で claim 前 400（6文字の
literal `\u0000` は区別して許容）。single factory が一度の検証で canonical Content /
Preconditions / immutable command の participant を生成し、order_value は OrderText が
唯一の数値源（digest と保存で同一 parsed value）。

### anchor 表

- `manual_article_operation_targets`（001_init.sql 内 `consolidated from 005_manual_article_operation_targets.sql` 節・2026-10-23 統合）:
  `(clinic_id, binding_id)` PK → `article_id`。binding 複合FK RESTRICT、RLS clinic scope、
  immutable UPDATE/DELETE trigger。article への FK・unique(article_id)・created_at/actor 列は持たない
  （lock順の都合で業務行 KEY SHARE を claim 内に入れない・同じglobal記事へ複数操作ありうる・
  actor等は operation_bindings が正本）。
- claim の winner が store の claim tx 内で `ClaimMetadataWriter.CreateClaimMetadata` を1回だけ
  呼び anchor を永続化する。callback 失敗は binding と anchor を同時 rollback、loser は呼ばれない。
  anchor 欠落・binding 不一致の legacy binding は fail-closed（別記事・request 値への fallback なし）。

### JCS 投影（`backend/internal/manualarticle/jcs.go`）

- `content_digest` = SHA-256 lower-hex of UTF-8 RFC8785 JCS over
  `{"schema":"manual-article-content-v1","title":<exact>,"order_value":<num>,"section":<exact>,"body_markdown":<exact>}`
  （member 昇順: body_markdown < order_value < schema < section < title）。
- `command_digest` = SHA-256 lower-hex of UTF-8 JCS over
  `{"schema":"manual-article-edit-v1","article_id":<canonical正数十進文字列>,"category":<enum>,"slug":<exact>,"content_digest":<64hex>}`
  （member 昇順: article_id < category < content_digest < schema < slug）。
- C06 Content = `{article_id, command_digest, content_digest}` の3 member のみ。
- `expected_version`・`expected_content_version_id` は Preconditions（fingerprint 対象外）。
- actor・clinic・timestamp・operation_key・生成版IDは投影に含めない。
- order_value は有限・±99999999.99・小数2桁以下を検査し、最短10進表現（-0→0）を
  JCS と numeric(10,2) 保存の共通源とする。文字列は trim・NFC・改行正規化しない。

### participant（`backend/internal/manualarticle/participant.go`）

- `EditParticipant` は request-scope（goroutine 共有しない）。`Effects`=`EffectsDatabaseOnly`、
  `Gate` は no-op。`Apply` は fingerprint 済み `req.Content.Fields()` のみから対象を解決し:
  現在の manual-edit:edit grant 再検査（失効→`edit-right-lost`=403）→ anchor 解決（欠落はfail-closed）
  → 対象行 FOR UPDATE lock → locked.id==anchor==Content.article_id かつ locked.category/slug==route
  selector（不一致→`target_mismatch`、効果なし）→ 元出所解決（current NULL・legacy_unknown・
  digest/revision NULL→`source_unresolved`）→ CAS（version と current_version_id==expected、
  不一致→`version_conflict`）→ digest 再計算と Content field の照合（不一致は効果なしで失敗）
  → 新内容版（revision=max+1、source_kind=`edited`、digest付き）→ CAS update で current のみ差替
  （version は storage trigger が +1、published は不変）→ `manual_article_operations`（intent=`edit`、
  before/after_state）→ `LogEntryTx` 必須監査（失敗は全体 rollback）→ receipt。
- `ResolveLookupTarget` は anchor→対象のみ、`AuthorizeLookup` は記事存在＋現在grantを再検査し、
  認可後にだけ selector を記録する（handler が route selector と比較し不一致は `409 target_mismatch`）。

### HTTP 写像

- `PUT /manual/articles/{category}/{slug}`: staff なし 403 → edit deps 未設定 500 → slug 空/parse 400 →
  経路記事不存在 404 → `article_id`≠経路記事 409 `target_mismatch` → `Runner.Execute` →
  `succeeded`/`replayed`=200 `ManualEditReceipt`（永続化 receipt のみ）、`operation-key-conflict`=409
  `operation_key_reused`、`version-conflict`=409、`validation`=`target_mismatch`/`source_unresolved`、
  `edit-right-lost`/`not-authorized`=403、その他=500 generic `ErrorResponse`。
- `GET .../operations/{operation_key}`: runner 未設定 500 → `Runner.Lookup` →
  `not-authorized`=403、anchor 認可後に route selector 比較（不一致=409 `target_mismatch`）、
  `succeeded`/`replayed`=`{state:"succeeded",receipt}`、
  `result-unavailable`（記録済みunknownのみ）=`{state:"unknown",receipt:null}`、
  それ以外（`in-progress`/`not-succeeded`/`scope-not-found`/結果store障害を含む）=
  500 generic `ErrorResponse`。公開 wire は `ManualEditReceipt`・`ManualEditConflict`・
  `ManualEditOperationResult` のみ（生 C06 envelope・binding・fingerprint・actor_ref・before/after JSON は非公開）。
- `GET .../edit`: manual-edit:edit。`ManualArticleEditView`（id 正数十進文字列、CAS版 version、
  current=内容版 DTO または null、source_kind NULL→`legacy_unknown`、digest NULL→null を補完しない）。
- 公開 READ（GET list / GET {category}/{slug}）は既存 `ManualArticleResponse` 形式を維持し、
  内容は `published_version_id` の版のみ（hidden・未公開は除外/404、取得障害は 500）。
- `GET versions` / `DELETE` は既存 route・応答形式を維持（versions は非公開内容を返すため
  edit 権限）。prune 削除は行わない。

### HMAC 設定

- `OPERATION_FINGERPRINT_CURRENT_KEY_VERSION`＋`OPERATION_FINGERPRINT_KEYS_JSON`
  （`config.OperationFingerprintKeys`）。両方必須または両方未設定。重複 member・非object・
  非文字列 value・空 map・空白/trim 変化 version・不正 base64・32バイト未満・current 不在は
  `Validate` が拒否。key 値はログ・エラーへ出さない。両方有効時に composition root で
  `Fingerprinter`/`Runner` を1つ作り注入し、未設定は edit 系 endpoint が fail-closed 500。

### 残件

- publish・restore・新規記事作成・hide/remove_override・all-owner C06・EMR-494 暗号化draftは対象外。
- ERD 生成（`docs/architecture/erd-detail/manual_article_operation_targets.md`）は `scripts/tools/generate-erd-detail.py` 公式 generator で生成済み（generated file・手編集なし。`--check` PASS・246 tables）。
- 実機検証（migration 適用・実行・並行・RLS・deferred trigger）は stack 起動後に別途実施。
