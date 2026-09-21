# SLACK-DECEASED: 死亡後の連絡記録は臨床 write と分け、死亡ガードは維持する

状態: **連絡メモ vs 臨床 write の比較 READY／PO 範囲 UNKNOWN／製品実装なし**。出典は [todo-issue.md](../../../todo-issue.md) 見出し `### SLACK-DECEASED`（L208–212、索引 L428）。保持する現場条件:

- 死亡後も飼主との連絡を記録したい（出典 982–986。現行 `todo-issue.md` は要約のみ。原文行は本票では再掲しない）
- **死亡 write ガードはバグではない。** `status=deceased OR deceased_at != nil` の fail-closed を「連絡できないから外す」対象にしない
- 死亡日の補完・フラグ解除とは別用途。 [PO-PET-DECEASED-DATA-BACKFILL](../todo-campaign-20260918/PO-PET-DECEASED-DATA-BACKFILL.md) を再開/統合しない
- 死亡日時は捏造しない。本票に実在個体の死亡日・件数を書かない

本ファイルは製品コードから import されない。キャンペーン unit `SLACK-DECEASED` の owned path および人間が読む比較票である。ledger `owned_paths` が本パス単体のため、他シートや dirty `todo-issue.md` への追記では unit 完了にならない。照合日 2026-09-20。worktree HEAD `873685b0b`。対象 `todo-issue.md` は読取のみ。

呼び出し行: **無い。** 既存 `docs/work/todo-campaign-20260918/` に本 unit の票は無く、`todo-issue.md` L208–212 は出典要約であり本票の代替ではない。

## 医院事実（コード外・UNKNOWN）

数値・院内ルール・誰が何を記録するかをコードから捏造しない。未裁定なら該当セルは **PO UNKNOWN**。

| 項目 | コードで分かること | 医院事実 |
| --- | --- | --- |
| 報告医院・担当者の個人名 | todo-issue に責任者個人名なし | **UNKNOWN**。名前のない要件で新欄を作らない |
| 記録対象（誰宛・どのペット・何の連絡） | 要望は「死亡後の飼主連絡」要約のみ | **UNKNOWN** |
| 閲覧/編集者 | owners `edit` が飼主/ペット備考 PATCH の RBAC | **UNKNOWN**。受付/看護/獣医師の誰が書いてよいかは PO |
| 時刻・記録者・監査 | 既存備考は上書き TEXT。行単位の actor/時刻列なし | **UNKNOWN**。追記ログが必要かは PO |
| 過去カルテとの関係 | 新規カルテ Create は死亡ペットで拒否 | **UNKNOWN**。既存カルテへの追記を許すかは PO。本票は許すと決めない |
| 連絡の保存先 | 飼主 `owners.remarks` / ペット `pets.remarks` は別列 | **UNKNOWN**。どちらが業務目的に足りるかは PO |
| 死亡日の根拠・対象件数 | ガードは日時を作らない。backfill 票は机上 | **UNKNOWN**。本票で日付を埋めない |

## 実践ゲート（product-philosophy 5 ステップ）

実装しない。本票の判断順だけ固定する。[docs/product-philosophy.md](../../product-philosophy.md) の逆行禁止。臨床の安全は効率に優先する。

1. **要件を疑う:** 「死亡後もカルテや予約が欲しい」は要件ではない。業務目的は「死亡後の**飼主連絡**を後から読める形で残す」こと。責任者の個人名は todo-issue に無い → **UNKNOWN**。名前のない要件でガード解除や新テーブルを作らない。
2. **削除:** 死亡フラグ解除、死亡日捏造、新規診療/処置/請求、第二の連絡ストア、BACKFILL との統合は削除対象。既存備考で足りるなら欄を足さない。
3. **簡素化:** 連絡は非診療メモ、臨床 write はガード。既存 `owners.remarks` / `pets.remarks` の上書きで足りるかだけを PO が決める。足りない場合の新ログは採否資料まで停止。
4. **サイクル短縮:** 確認ダイアログで死亡ガードを代替しない。新規カルテ・予約・会計・入院・検査・カルテ画像はサーバーで fail-closed のまま。
5. **自動化:** 死亡後連絡の自動転記、LSTEP 自動送信、日付自動補完はしない。手動で同じ use case が完結してから。停止手段・失敗通知・監査なしの自動化は禁止。

追加だけで削除（工程・画面・入力・二重管理）がゼロなら再検討。死亡ガードを外して臨床画面へ書く案は ①② 違反である。

## 混ぜてはいけないケース

「死亡後も記録したい」は次のどれでも同じ言葉になる。連絡メモと臨床 write を混ぜない。

| ID | 何か | いまのコード | 混ぜてはいけない読み替え |
| --- | --- | --- | --- |
| G — 死亡 write ガード | [pet_not_deceased.go](../../../backend/internal/sharedkernel/pet_not_deceased.go) L16–30。`status=deceased OR deceased_at != nil` を InvalidInput。日時捏造・backfill はしない（L17） | ガードが連絡を妨げる = バグ、とはしない。FE 非表示だけでは足りない（L18） |
| C — 新規カルテ | [medical_record_appointment_context.go](../../../backend/internal/medicalrecord/medical_record_appointment_context.go) L105–132。メッセージ「死亡したペットは新規カルテを作成できません」。回帰 [medical_record_deceased_pet_test.go](../../../backend/internal/medicalrecord/medical_record_deceased_pet_test.go) L21–64 は Create が repository に届かない | 死亡後連絡の保存先を新規カルテにしない |
| B — 会計 | [accounting_service.go](../../../backend/internal/billing/accounting_service.go) L207–226。「死亡したペットは会計を作成できません」 | 請求・会計 Create/Complete を連絡記録に使わない |
| R — 予約 | [reservation_service.go](../../../backend/internal/reservation/reservation_service.go) L222–244。「死亡したペットは予約できません」 | 死亡後フォローを新規予約で立てない |
| H — 入院 | [hospitalization_service.go](../../../backend/internal/medicalrecord/hospitalization_service.go) L53–56。同一 helper | 入院メモを連絡ログの代替にしない |
| E — 検査 | [examination_pet_safety.go](../../../backend/internal/medicalrecord/examination_pet_safety.go) L11, L47–53。「死亡したペットには検査記録を登録できません」 | 検査記録を連絡に使わない |
| I — カルテ画像 | [medical_record_image_service.go](../../../backend/internal/medicalrecord/medical_record_image_service.go) L174–192。SEC-CS-F14 fail-closed | 画像 upload を「何か残せる経路」としない |
| P — ペット備考 | `pets.remarks` TEXT。PATCH `/pets/:id`（[routes.go](../../../backend/internal/pet/routes.go) L23、[pet_request.go](../../../backend/internal/pet/pet_request.go) L176、[service.go](../../../backend/internal/pet/service.go) L165–167, L333）。`ValidatePetNotDeceased` は pet package に無い | 備考上書き = 監査付き連絡ログ、とはしない。臨床メモと飼主連絡を同一欄に混ぜると二重用途 |
| O — 飼主備考 | `owners.remarks`。PATCH `/owners/:id`（[http_routes.go](../../../backend/internal/owner/http_routes.go) L15、[http_request.go](../../../backend/internal/owner/http_request.go) L235、[update_command.go](../../../backend/internal/owner/update_command.go) L124–125、[service_core.go](../../../backend/internal/owner/service_core.go) L99–139）。ペット死亡判定なし | 飼主1人に複数ペットがいるとき、個体別連絡先にしない。会員種別 `deceased`（退亡者）とペット死亡を同一視しない |
| D — 死亡記録 API | PATCH/DELETE `/pets/:id/death`（[lstep/routes.go](../../../backend/internal/lstep/routes.go) L198–202）。generic PATCH は status を持たない（BUG-415、pet_request.go L157–160） | 連絡のために死亡解除・日付変更しない |
| F — 死亡日 backfill | [PO-PET-DECEASED-DATA-BACKFILL](../todo-campaign-20260918/PO-PET-DECEASED-DATA-BACKFILL.md) は日付根拠のある限定訂正。本票と別 ID | `deceased_at` NULL を埋めて連絡可能にする、は本票の範囲外 |
| N — 新テーブル | 連絡専用 table / タイムライン API は repo に無い | PO 未裁定でスキーマ追加しない |

## 死亡ガード（維持する。バグではない）

正本は [ValidatePetNotDeceased](../../../backend/internal/sharedkernel/pet_not_deceased.go):

```16:30:backend/internal/sharedkernel/pet_not_deceased.go
// ValidatePetNotDeceased は死亡ペットへの業務 write を fail-closed で拒否する。
// 死亡契約: status=deceased OR deceased_at != nil（日時の捏造・backfill はしない）。
// FE の選択 UI ブロックだけでは API 直叩きを防げないため BE 側でも検証する。
func ValidatePetNotDeceased(...) error {
	// ...
	if pet.Status == model.PetStatusDeceased || pet.DeceasedAt != nil {
		return apperrors.WrapInvalidInput(message)
	}
```

kernel 回帰 [sharedkernel_test.go](../../../backend/internal/sharedkernel/sharedkernel_test.go) L25–57 は次を固定する。日付は fixture のみ（`2026-07-01`）。実在個体の死亡日ではない。

| 合成状態 | 結果 |
| --- | --- |
| alive / `deceased_at` nil | 通過 |
| deceased / nil | InvalidInput（日時を作らない） |
| deceased / dated | InvalidInput |
| alive / dated | InvalidInput |

カルテ Create 回帰 [medical_record_deceased_pet_test.go](../../../backend/internal/medicalrecord/medical_record_deceased_pet_test.go) L21–64: `status=deceased` かつ dated fixture で Create が repository に届かない。L66–99 は生存ペットのみ Create 許可。本票はテストを「緩める」対象にしない。

FE 臨床契約は BE と同義: [isPetDeceasedForClinicalWrite](../../../frontend/src/lib/transforms/pet.ts) L26–35。`status === "死亡" || deceasedAt != null`。日時は捏造しない（L28）。

`status` の通常 PATCH は閉じている（BUG-415）。死亡の唯一の書込は Create と `/pets/:id/death`（監査 + `deceased_at` 同一 tx）。連絡記録のためにこの一本化を解かない。

## 既存備考 write（臨床ガードの外）

| 経路 | 保存先 | 権限 | 死亡ガード | 監査・時刻・記録者 | 用途の実態 |
| --- | --- | --- | --- | --- | --- |
| PATCH `/api/v1/owners/:id` | `owners.remarks` max 2000（bind）。FE [OwnerBasicFields.tsx](../../../frontend/src/features/owners/components/OwnerBasicFields.tsx) L129–139 は `maxLength={1000}` | `owners` `edit` | **無い。** Update は飼主行ロックと discount 再検証（SEC-CS-F15）のみ | 通常 owner Update に連絡専用 audit は無い。LINE 操作の audit は別経路 | 飼主全体の特記。個体・連絡イベント単位ではない。上書き |
| PATCH `/api/v1/pets/:id` `{remarks}` | `pets.remarks` max 2000。[PetCareSection.tsx](../../../frontend/src/features/owners/components/PetCareSection.tsx) L184–194 ラベル「備考・特記事項」 | `owners` `edit` | **pet Update は ValidatePetNotDeceased を呼ばない**（`backend/internal/pet` に 0 件） | 備考 PATCH 専用の actor/時刻列なし。死亡記録 API の監査とは別 | 個体のケア特記（食べ物・環境と同セクション）。上書き。連絡タイムラインではない |
| 新規カルテ / 予約 / 会計 / 入院 / 検査 / 画像 | 各 domain table | 各 resource | **掛かる** | 業務 write の既存監査 | 臨床・会計。連絡メモの保存先にしない |

飼主備考とペット備考は別列である。どちらか一方へ「死亡後連絡」を上書きすると、既存特記を消す。追記形式・誰がいつ書いたかは現行 schema に無い。PO が「既存備考で足りる」と裁定するまで、連絡専用欄は作らない。

慢性疾患 `notes`（[chronic_condition_service.go](../../../backend/internal/pet/chronic_condition_service.go)）は病名メモであり、死亡後の飼主連絡先ではない。

## 連絡メモ案（実装しない。比較用）

PO が記録対象・閲覧/編集者・時刻/記録者/監査・過去カルテ関係を確定するまで、いずれも製品変更しない。

| 案 | 削除できる工程 | 残る欠落 | 停止条件 |
| --- | --- | --- | --- |
| A. 既存ペット備考へ運用で追記 | 新画面・新 table ゼロ | 上書き衝突、actor/時刻なし、臨床特記と混在、複数回連絡の履歴なし | 既存特記を消してよいか PO 未決なら使わない |
| B. 既存飼主備考へ運用で追記 | 新画面ゼロ | 個体非紐付け、飼主死亡（会員 `deceased`）とペット死亡の混同、上書き | 複数ペットの飼主で個体が特定できないなら使わない |
| C. 限定連絡ログ（新） | 備考誤用は減る | 二重管理・新入力。削減工程がゼロなら再検討 | 責任者名と受入条件なしでは作らない |
| D. 新規カルテ/予約/会計 | なし。工程が増え臨床安全を壊す | ガードと矛盾 | **採用しない** |
| E. 死亡解除・死亡日捏造 | なし | 臨床 write が再開し、BACKFILL と衝突 | **採用しない** |

推奨（本票の机上結論、PO 裁定ではない）: **臨床 write 経路は閉じたまま。** 運用で足りるかは A/B を PO が既存備考の読者・上書き可否と照合して決める。足りなければ C の採否資料。D/E は出さない。

## PO へ渡す限定質問（再質問を増やさない）

[todo-issue.md](../../../todo-issue.md) L212 の完了条件を分解する。本票で答えを書かない。

1. 記録対象は飼主か、死亡した個体か、医院全体の連絡か。
2. 書いてよい職種と、読んでよい職種。
3. 1行上書きで足りるか。時刻・記録者・監査が必須か。必須なら既存 `audit_logs` と同じ tx か。
4. 過去カルテは参照のみか。死亡後のカルテ追記/新規は禁止のままか。
5. 既存 `pets.remarks` / `owners.remarks` で足りるか。足りないときだけ新ログを検討する。

用途未確定なら **ガード変更を停止**。許可した連絡だけが保存でき、新規臨床 write と請求は禁止のまま、が受入条件である。

## 検証（本票）

- ガードをバグ扱いしていないこと、死亡日を捏造していないこと、BACKFILL を統合していないこと。
- 製品コード・テスト・migration を変えていないこと。owned path のみ。
- Docker アプリ試験は docs-only のため対象外。
