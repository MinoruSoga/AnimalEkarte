# EMR-502：既存共通記事 edit の採択契約

更新日：2026-10-21。公開先：`docs/work/emr/emr-502-manual-edit-contract-20261021.md`。

## 状態・適用範囲

- 本文は Plane EMR-502 の 2026-10-20 21:15 UTC 採択と、20:59 UTC adapter 最終補正を文書化する。設計採択であり、実装済み・稼働済み・試験合格を意味しない。
- 対象は既存記事の `manual.article.edit`、EMR-513 の既存 FE 接続、EMR-489 の限定 claim hook/anchor。新内容版・current pointer・操作履歴・必須監査・成功 receipt を原子的に保存する。
- publish / restore / new article / hide / remove_override の移行、全 C06 owner 接続、EMR-494 全体、暗号化 draft は対象外。EMR-502 全体完了とは区別する。
- 記事は全院共通のまま。receipt の clinic は認証・認可 scope であり、記事の医院別所有・複製・医院別 CAS を導入しない。

## 1. 原対象を保持する claim anchor

- 新規表名：`manual_article_operation_targets`。列は `clinic_id` / `binding_id` / `article_id` のみ。各列は正の signed bigint、NOT NULL、CHECK > 0。
- PK は `(clinic_id, binding_id)`。即時・非遅延 FK `(binding_id, clinic_id)` → `operation_bindings(id, clinic_id)`、ON DELETE RESTRICT。
- article FK は設けない。claim 中に業務行の KEY SHARE lock を先取りしないため。別 clinic FK と article_id の UNIQUE も不要。生成 ID・default・version・本文・状態・時刻・actor を重複保存しない。
- trusted scope の `owner_kind="manualarticle"` / `intent="manual.article.edit"` に限り、canonical `Content.article_id` から挿入する。解決時は binding を結合し owner / intent / current scope を検査する。
- repository は INSERT / SELECT のみ。表ローカルの BEFORE UPDATE OR DELETE 拒否 trigger を追加する。ordinary invoker とし SECURITY DEFINER、新権限・role は追加しない。
- anchor の上書き・補修・推定 backfill・cleanup endpoint は設けない。旧 binding に anchor がなければ fail closed。成功履歴を claim の代用品にせず、架空の成功/no-op 履歴も作らない。
- 既存 `app_private.apply_rls_policy` と `app_private.has_clinic_access(clinic_id)` の FOR ALL USING / WITH CHECK 規約を使う。既存 ENABLE のままで FORCE / bypass 方針を新設しない。
- repository も trusted selected clinic + binding で絞る。RLS は actor / intent / 現在の編集権限の検査を代替しない。

## 2. claim と業務保存の順序

- optional capability `ClaimMetadataWriter` の `CreateClaimMetadata(ctx, binding, request) error` を Runner が認識する。502 participant は必須実装、無関係な既存 participant の挙動は維持する。
- callback は渡された ambient claim transaction と canonical req.Content を使う。別 payload 捕捉（未検証・Content commitment に非束縛の payload）、Execute 再入、nested transaction、記事/履歴/receipt 書込、外部作用は禁止。
- 短い claim TX：既存の正確な unique conflict target に INSERT ... DO NOTHING RETURNING → 新規挿入 winner のみ callback → 保存 binding 読取 → commit。
- winner 判定は RETURNING / RowsAffected。候補 ID の有無で推定しない。callback 失敗では binding と anchor が共に rollback。競合 loser は winner を待って原 binding / anchor を読み、callback を再実行しない。
- claim callback 内では記事行の読取・lock をしない。認証済み非 locking preflight を行っても Apply / lookup で再検査する。
- claim commit 結果不明は既存 internal-error / unavailable 応答で停止。同じ key を保持して同 scope の lookup へ進む。新 key による盲目的再実行、読取エラーから不存在の推定、旧 binding の補修は禁止。
- 固定順序：短い claim TX 終了 → owner Gate（global manual article は no-op）→ binding FOR UPDATE NOWAIT → 現在認可付き replay / unknown 確認 → savepoint → Apply 現在権限 → global article FOR UPDATE → CAS / current version → 内容版・current pointer・manual_article_operations・必須監査 → operation_results → commit。
- no-op Gate は予約 graph lock が不要という意味のみ。認可省略ではない。予約 graph / lifecycle / editor epoch lock は追加しない。
- succeeded / replayed / unknown / busy / key-conflict / settledState の開示前に、原 anchor に対する ResolveLookupTarget → AuthorizeLookup を行う。Apply を通らない経路と rollback / error 経路も同じ。入口 middleware だけでは不足する。

## 3. lookup・公開 adapter・selector

- binding は trusted clinic + actor + owner + intent + key で検索する。原 article は anchor だけで解決し、存在と現在の manual-edit 権限を検査する。
- 欠落・曖昧 anchor / 認可不能では非開示、result_refs を返さない。incoming URL / body / 別記事 / 現在 source からの fallback は禁止。
- Apply は原 anchor を解決し現在認可を再確認して global article を lock。書込前に locked article.id = anchor.article_id = canonical request.article_id、さらに category / slug と commitment の route selector の一致を検査する。
- 認可済み不一致は既存 HTTP 409 ManualEditConflict `target_mismatch`、業務作用ゼロ。incoming selector で別記事を引き直さない。
- Lookup も原 anchor の現在認可を先に行い、その後 route category / slug を照合する。認可済み不一致のみ同じ 409。未認可に mismatch 詳細を漏らさない。
- internal binding intent は `manual.article.edit`。成功 domain history `manual_article_operations.intent` と公開 `ManualEditReceipt.intent` は `edit`。adapter で明示変換し dotted family を history enum / wire に流さない。
- 公開 DTO は既存 ManualEditReceipt / ManualEditConflict / ManualEditOperationResult。raw C06 ResultResponse に差し替えず、既存 513 client schema を保持する。
- 公開 lookup status は既存 succeeded / unknown のみ。未選定の internal outcome を新公開 state にせず、既存 generic ErrorResponse に写像する。
- same-key 成功 replay は保存済み履歴/result の原 receipt・原 version を返し、最新 pointer から再構築しない。現在版の変更は原 receipt を失効させない。
- unknown は sticky な停止結果。後続失敗で消さず新 key で救済しない。clinic 切替・別 actor へ receipt を横流ししない。既存 rollback / unknown 機構を維持する。

## 4. immutable content と正確な hash projection

- 大きな本文と C06 Content の各値 256-byte 制限を両立するため、本文を Content に入れず共有上限も変更しない。
- nested content は title:string / order_value:number(double) / section:string / body_markdown:string の全置換。category / slug は immutable route selector。
- content の 4 項目は required / non-null。JSON 重複 member、lone surrogate / invalid UTF-8、非有限・範囲外 JSON number を拒否する。
- 公開 API の `ManualArticleEditContent` に従い、title / section だけ空文字を拒否し、body_markdown は空文字を許容する。新たな文字数制限・trim・空白のみ拒否を導入せず、旧 handler の `max=255` / `required,max=100000` は継承しない。
- order_value は公開 API どおり numeric(10,2) に対応する number、multipleOf 0.01、最小 -99999999.99・最大 99999999.99 とする。
- 上の 4 項目限定は nested content の意味。outer operation_key / article_id / expected_version / expected_content_version_id は公開契約どおり許容し、whole request の新たな一括 unknown-field 拒否を導入しない。
- decoded string を保存/hash 共通に用い、trim / NFC / 改行変換をしない。RFC 8785 JCS の key sort・escape・有限 IEEE754 serialization を用いる。-0 は 0 に canonicalize し同じ parsed value を保存する。order_value を数値文字列契約に変えない。
- `content_digest = lowerhex SHA256(UTF8(JCS({schema:"manual-article-content-v1",title:<exact>,order_value:<parsed finite number>,section:<exact>,body_markdown:<exact>})))`
- `command_digest = lowerhex SHA256(UTF8(JCS({schema:"manual-article-edit-v1",article_id:<canonical positive signed64 decimal string>,category:<exact enum>,slug:<exact route-decoded string>,content_digest:<above64hex>})))`
- C06 Content は article_id / command_digest / content_digest の 3 strings のみ、各値 ≤ 256 bytes。operation_key は binding 選択。expected_version / expected_content_version_id は Preconditions で、両 digest に含めない。
- generated version ID / actor / clinic / timestamp / operation key も両 projection に含めない。
- domain 側の typed immutable input として、HTTP で一度検証した値から単一 factory が canonical Content と `ManualArticleEditParticipant` を同時生成する。factory の設計上の提案名は `NewManualEditExecution` であり、既存実装済みの名称ではない。
- participant の private command 値が唯一の write input。article_id / category / slug / title / order_value / section / body_markdown を所有し、setter や可変 map / pointer alias を公開せず必要な copy を行う。別 DTO の再 parse 差込み、クロージャによる別 payload 捕捉は禁止する。
- Runner は既存 `Participant.Apply` をその typed instance へ dispatch する。Apply は最初の domain write 前に、自身が保持する同一 immutable command の article / selector / content から両 projection を再計算し、`req.Content.Fields()` の 3 項目を完全照合する。article_id を含む全 field 一致・両 SHA256 一致後の verified command だけを writer / snapshot へ渡す。不一致では作用ゼロ。Content 生成時と異なる値は保存せず、snapshot と digest に同じ parsed 値を使う。
- 同 key の再 attempt も同 factory から作成した Content で既存 binding の HMAC 一致を通す。成功 Lookup は保存済み receipt / 原 anchor / 現在認可を使い、本文を現在値から再構成しない。
- この typed participant constructor 経路は明示注入と Content commitment で full command を束縛する方式であり、「別 payload 禁止」は未検証・非束縛 payload の禁止を意味する。generic Content の 3 項目と共有 C06 契約を維持し、raw 本文の claim anchor 保存や Content 上限変更は行わない。以上は実装経路の契約であり、runtime 実装済みの宣言ではない。

## 5. HMAC 設定の選択と非実行境界

- 既存 NewFingerprinter(current, keys)、domain separation、length-prefixed canonical fingerprint、binding ごとの key_version を維持。上記 digest は HMAC の置換ではない。
- 設定名は `OPERATION_FINGERPRINT_CURRENT_KEY_VERSION` と `OPERATION_FINGERPRINT_KEYS_JSON`。後者は version → standard padded base64 key bytes の JSON object。default は設けない。
- duplicate member / 非 object / 非 string 値 / 空 map / blank または trim で変わる version / 不正 base64 / decoded 32-byte 未満 / current version 不在を拒否。decoded map を copy して既存 constructor へ渡す。
- 両方 absent：当該 manual edit / lookup は登録を残し claim / write 前に fail closed、既存 generic 500 ErrorResponse。committed unknown と誤表現しない。他の既存 domain / manual reader は継続可能。
- 一方だけ present または値不正：値を含まない field / error 説明で startup 拒否。両方 valid：composition root の共有 Fingerprinter / Runner を interface 注入する。
- 未知の保存 key_version は Matches を fail closed。current key への代替は禁止。純粋な認可済み Lookup は valid Runner 下で保存 result を返せるが、write を許可する意味ではない。
- retained binding に必要な旧 key version は保持。JWT / integration secret 流用、runtime random default、fixture/dev key fallback、実鍵生成・配置・rotation・secret-manager grants は対象外。設定値/body/key/base64 を log しない。

## 6. 実装引継ぎ・確認項目

- 現行 session 3 が 502 / 513 / 489 限定を所有。共通 operation は同 session 専有。001 / ERD は session 2 → 3 の単一 writer 順で統合する。
- api.yaml は必要な domain hunk のみ。main.go / registration は最小 DI / route patch。共通 models.ts 正規生成・最終統合は session 3、生成物手編集なし。他 session の必要差分を受領する順序を明示する。
- 文書・operation 参加表も採択内容へ同期する。公開前に現在の claim / worktree / branch / PR と文書所有者を照合する。docs 差分がないことを所有権不在とみなさない。
- fixture：claim+anchor rollback / concurrent same-key 1 anchor / target B が原 A を置換しない / failed・unknown の原 A 解決 / missing anchor 非補修 / 失権・別 clinic・actor 非開示。
- fixture：global article cross-clinic CAS / required audit rollback / 大本文と digest mismatch / 原 version replay / absent・partial・malformed・duplicate config / 旧 key 不在 / selector と公開 intent 写像。
- 試験実行は NOT_RUN。隔離環境を構築・使用しない。実鍵設定・DB 適用・deploy は行わない。実 DB / 結合 / UAT は本人指示後に main ローカルでまとめ検証し、未実行を PASS と記さない。

## 根拠

- [EMR-502 採択全文・最終 adapter 補正](https://app.plane.so/baritechllc/browse/EMR-502/)（本文取得：2026-10-21、updated_at 2026-10-20T21:17:05.443217Z）。
- [EMR-489](https://app.plane.so/baritechllc/browse/EMR-489/) / [EMR-513](https://app.plane.so/baritechllc/browse/EMR-513/) / [EMR-494](https://app.plane.so/baritechllc/browse/EMR-494/)。
- [D22 selected-save-contract](https://github.com/MinoruSoga/AnimalEkarte/blob/0c498155f8c6ade9fea291079ad2b8e2aaaa5caf/docs/design/domains/internal-manual/technical-spec.md#manual-selected-save-contract)。
- [C06 参加表](https://github.com/MinoruSoga/AnimalEkarte/blob/0c498155f8c6ade9fea291079ad2b8e2aaaa5caf/docs/design/shared/operation-participation.md)。コード・DDL の実装済み判定は最新採用 SHA で別途行う。
