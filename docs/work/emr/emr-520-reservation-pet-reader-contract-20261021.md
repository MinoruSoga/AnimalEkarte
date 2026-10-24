# EMR-520：予約の認可済み pet identity / name reader 契約

<a id="compatibility-scope-20261024"></a>
## 2026-10-24 現在の適用範囲（519 / 520）

旧number-only client解消→producer文字列化→旧producer不在後のnumber互換除去というclient-first工程、および旧pet_name-only/両省略shapeの救済を必須条件から外す。確定canonical十進string IDとpaired pet_id/pet_nameへBEと全必要consumerを同じ完結単位で揃える。unsafe ID拒否・結果不明の盲目再送禁止・現在認可・問診activation gateを維持し、同一認可snapshotの対、未特定/失権の両nullと名前からidentityを推測しない規則を維持する。以下の段階互換fixture/切替順は旧移行案の記録であり、新しい必須受入条件ではない。 詳細根拠は[全票監査の採用是正方針](backward-compatibility-audit-20261024.md#accepted-corrections)。以下の旧日付の記録はこの限定追補と区別する。製品置換・実行受入済みではない。

更新日：2026-10-21。公開先：`docs/work/emr/emr-520-reservation-pet-reader-contract-20261021.md`。

## 状態・適用範囲

- 本文は EMR-520 の 2026-10-20 21:15 UTC 採択と、2026-10-21 の PLAN-POLICY-20261021-520 追補を文書化する。採択済み設計であり、実装完了・API 同期済み・runtime 受入済みの宣言ではない。
- 現行束 4 は EMR-519 の予約 ID 文字列移行、EMR-520 の実 pet identity / name projection、EMR-491 の pet reader 限定を同一実装 PR で扱う。
- 旧本文の「設計待ち / Blocked 相当」「ID 移行 A と分離」「具体 reader 未成立なので実装 ready と記さない」は採択前の履歴。採択設計を未決扱いして停止する根拠にしない。
- 共通基本情報 reader 全体、owner / pet writer、新規登録、移管、全検索、複数頭 engine は非対象。EMR-491 全体完成を一律の開始条件にしない。
- 新しい業務上の閲覧権限、他院・副飼主へのアクセス拡張、過去問診の帰属書換を導入しない。

## 1. 一貫した認可・binding・名前の取得

以下を分離した事前確認と後続名前検索にせず、1 SQL snapshot で結び付ける。

1. 認証済み clinic / customer / sub を起点とする。入力 URL / body や表示値を認証 principal の代用にしない。
2. 現在の予約の `appointments.line_customer_id` が認証 customer と一致する。
3. 予約が保持する実 appointment.pet_id の pet と owner が非削除であり、pet.owner_id = appointment.owner_id を満たす。
4. 同 clinic / owner の owner_line_binding_states が linked で、その customer が認証 customer と一致する。
5. common_pet_clinic_bindings の `(clinic_id, local_pet_id)` から現在の common pet binding を解決する。
6. common_pet_subjects.current_basic_revision_id から、同じ common_pet_id の basic revision.name を読む。

- 認可・予約 pet 参照・binding・current revision・name を同じ snapshot で対応付ける。古い認可と新しい名前を別々に組み合わせない。
- customer.owner_id、古い cache、customer_fields の先頭 pet 名、名前一致、選択中 pet、別 pet の現在値への fallback は禁止。
- model.Pet に Name がないことを pets の旧 name 列復活や r.Pet.Name 参照で埋めない。保存正本の current common basic revision を読む。
- 内部 common UUID / revision ID は公開しない。公開 identity は予約の自院 local pet ID であり、common ID と置換しない。
- 同じ SQL reader を LIFF 側で再実装せず、EMR-491 限定 reader の責任に集約する。

## 2. 対で返す wire 契約

対象一覧：`GET /api/liff/{clinicId}/my-reservations`。

- pet_id と pet_name は一対。新 server は両 key を常に返し、認可済み実 pet が解決した場合のみ両 non-null、それ以外は両 null。
- pet_id は正の signed-int64 の canonical decimal string。上限は 9223372036854775807。ゼロ・負値・非正規表現・範囲外を成功値にしない。新 pet_id は string のみで、予約 id の旧 number 互換を流用しない。
- pet_name はその identity に対応する同 snapshot の認可済み current revision の名前。未知を空文字へ補完せず、名前から ID を推測しない。
- 旧移行案の記録（現在の必須受入条件ではない）：後方互換として旧 producer の両 key 省略を許容していた。対として扱う形は「両省略」「両 null」「両 non-null」とする。
- 旧移行案の記録（現在のadapter必須条件ではない）：旧 producer は `pet_name,omitempty` のみを持ち、pet_id なしで customer_fields 由来の名前を返す。旧 pet_name-only（pet_id 省略、pet_name が既存の string 型）の応答は、旧 producer 向け読取 adapter の限定互換として、認可済み identity 不明の両 null / unavailable へ正規化する。その名前を表示・ID 推定に利用しない。
- 旧移行案の限定正規化について：この正規化は旧 producer の生 wire が新契約に適合するという意味ではない。新 server の両 key・両 null / 両 non-null 保証は維持する。pet_id-only、片側 null、不正な型など、上記限定互換以外の partial pair は拒否する。この読取 adapter の契約を、現行 client が旧 name-only を既に拒否しているという実装済み判定にしない。
- 作成 201 は EMR-519 の id 文字列化と既存 notes を維持する。この subset では pet_id / pet_name ペアを作成応答へ追加しない。
- EMR-519 の予約 ID の最終契約は canonical positive signed-int64 decimal string。以下は旧移行案の第1段階の記録で、現在のnumber受付必須条件ではない。新 client の読取 adapter がこの string と、`Number.isSafeInteger(id) && id > 0` を満たす旧 number だけを受け、旧 number は直ちに string 化する。一覧と作成201の双方に適用し、後続 state / 比較まで string に統一する。unsafe number の丸め復元・FE reviver 等による救済は禁止する。
- 旧移行案の第2・第3段階の記録（現在の必須切替順ではない）：第2段階は、旧 number-only client が利用されないこと（既存 open client の更新 / 再読込を含む）を確認してから producer を string 化する。第3段階は旧 producer 不在を確認した後に numeric 互換を外す。新 pet_id の string-only と旧 pet_name-only の限定正規化は、この予約 id 移行中も維持する。
- backend/docs/api.yaml と専有 DTO / client / schema は同じ実装 PR で同期するが、同じ PR であることは atomic deploy の証拠ではない。上記のclient-first切替順は旧移行案の記録で、現在の必須受入条件ではない。本書の採択・公開を機械 API 反映済み・実デプロイ済みと読み替えない。今回、実デプロイは行わない。

## 3. null・失権・取得エラーの区分

- binding / current revision / 認可済み一致が欠落、未特定、または失権している場合は pet_id / pet_name を両 null にする。名前不足だけで予約一覧全体を破壊しない。
- DB 取得失敗は取得エラーとして扱い、成功応答の両 null に握りつぶさない。
- 非表示・null は問診の「未回答」と別の事実。回答状態や LINE 回答資格を推測しない。
- 問診 FE は一覧の pet identity と `current_target.ref.pet_id` が一致する場合だけ名前を表示する。不一致、片側欠落、未認可では名前を出さず、他 pet から補完しない。
- pet 付替え、binding / basic revision 更新、unlink / 移管 / 失権は現在の reader 条件で評価する。過去回答の原文・帰属を書き換える許可ではない。

## 4. 変更境界・共有ファイル

- 候補は reservation/liff_response.go、liff_service_reservations.go、既存予約 reader の認可済み projection、backend/docs/api.yaml、LIFF models / schema / summary wiring。
- session 4 が 519 / 520 / 491 pet reader 限定を担当。001 / ERD は session 2 → 3 の単一 writer 順、共通 operation は session 3 専有という現行調整を守る。
- api.yaml は自 domain の必要 hunk のみ。main.go / registration は自 domain の最小 DI / route patch。全体上書きは禁止。
- 専有 client は当該 session、共通 frontend models.ts の正規生成・最終統合は session 3。必要差分・採用 commit を渡し、各 PR が必要生成差分を受領して自己完結する順序を先に確定する。生成物は手編集しない。
- 開始・公開前に最新 main / claim / worktree / branch / PR と所有者を確認する。未担当表示、remote branch 不在、docs 差分なしは claim 不在・文書所有権不在の証拠ではない。
- 文書の公開と、実装 session が行う機械 API / reader / FE 接続の実装を分けて記録する。採択済み設計の再選定や別 reader の並行実装はしない。

## 5. fixture と未実行の記録

以下は検査内容の記録。旧移行案と明記した互換fixtureは現在の必須受入条件ではなく、その他の選定済み安全条件・必要consumer整合を維持する。実行結果ではない。

- customer_fields 先頭が A でも実予約 pet B の ID / name を返す。同名別 pet を混同しない。
- pet 付替え / binding 世代変更 / current basic revision 更新で、identity と name の同 snapshot 対応を保つ。
- pet 未特定、binding / revision 欠落、unlink / 移管 / 失権、他院 local ID では、非開示または認可条件に従う両 null を維持する。
- DB reader 失敗が取得エラーになり、null 成功に変わらない。
- 旧移行案の互換fixture（現在は必須でない）：両省略の旧 producer、両 null、両 non-null を扱う。旧 pet_name-only（pet_id 省略、pet_name は string）は読取 adapter で両 null / unavailable へ正規化し、名前非表示・ID 推定なしを確認する。pet_id-only、片側 null、不正な型など、それ以外の partial pair は拒否する。
- positive signed-int64 の境界・範囲外・非正規 ID、current_target.ref.pet_id 不一致での名前非表示を固定する。
- 作成 201 の id / notes を維持し、未採用 pet pair を追加しない。
- 旧移行案の互換fixture（number受付は現在必須でない）：予約 id は一覧・作成201の両方で canonical signed-int64 string と正の safe integer だけを受け、旧 number の即時 string 化と後続 state / 比較の string 統一を確認する。unsafe / 非整数 / 非正数の旧 number、不正・範囲外 string は拒否し、新 pet_id に number 互換を適用しない。
- 旧移行案の記録：旧 number-only client（既存 open client を含む）の解消確認 → producer の string 化 → 旧 producer 不在確認 → numeric 互換除去、の順序。これは現在の必須受入条件ではなく、切替・実デプロイ実施結果でもない。
- 試験実行は NOT_RUN。隔離環境は構築・使用しない。実鍵・資格設定、DB 適用、deploy は対象外。
- 実 DB / 結合 / 実画面 / UAT は概ね機能票消化後、本人の指示を受けて main ローカルで実施する。未実行を PASS と記さず、必要受入条件は削除しない。

## 根拠

- [EMR-520 現行着手契約・履歴優先関係](https://app.plane.so/baritechllc/browse/EMR-520/)（本文取得：2026-10-21、updated_at 2026-10-21T00:49:03.098808Z）。
- [EMR-519 ID 文字列移行](https://app.plane.so/baritechllc/browse/EMR-519/) / [EMR-491 共通情報](https://app.plane.so/baritechllc/browse/EMR-491/) / [EMR-472 verified LINE principal](https://app.plane.so/baritechllc/browse/EMR-472/)。
- [予約 DTO の参照基点](https://github.com/MinoruSoga/AnimalEkarte/blob/13942085357b16a1c1e167ce2233c1181702f222/backend/docs/api.yaml#L27973-L28077)。
- [既存 LIFF response の gap](https://github.com/MinoruSoga/AnimalEkarte/blob/a28f81bfe99010c440ff42a51b8f31811ae1904f/backend/internal/reservation/liff_response.go#L126-L138)。
- [LINE 画面仕様](https://github.com/MinoruSoga/AnimalEkarte/blob/13942085357b16a1c1e167ce2233c1181702f222/docs/spec/screens/37-line-reserve-owner-flow.md#decision-alignment-20261017)。これらの固定 SHA は根拠であり、最新 runtime 完了の証明ではない。
