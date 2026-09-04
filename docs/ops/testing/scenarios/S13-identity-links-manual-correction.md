# S13: 同一飼主・ペット連携 — 手動訂正（link → history → unlink → relink）

> **目的**: 2 医院にまたがる同一飼主・ペットを、権限のあるスタッフが手動で link し、連携治療履歴を確認し、誤リンクを unlink してから正しい組み合わせで relink できることを納品前に証明する。
> **所要目安**: 20分 / **深度**: 中
> **仕様正本**: [screens/40-identity-links.md](../../../spec/screens/40-identity-links.md)。実装参照: `backend/internal/identitylink/`・`frontend/src/features/identity-links/`。

## 前提条件

- ローカル、または USER 管理の STG synthetic lane 専用。本番データは禁止する。
- clinic A/B、両 clinic に所属して `identity-links:view/edit` を持つ attached actor、各 clinic の合成 owner/pet/診療履歴を明示的に作成する。
- actor は anchor clinic と全 active owner member clinic に所属する。固定 ID や汎用 seed を仮定しない。
- cleanup では pet group と owner group の両方を unlink/remove し、作成した owner/pet/history を削除する。
- 依存シナリオ: なし。Phase 2（自動 link / merge）は対象外。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | `/identity-links` を開く | workbench が表示される。edit 権限ありなら link/unlink ボタンが出る（view のみなら閲覧バナー） |
| 2 | 飼主検索で clinic A・clinic B の対象飼主を選び「飼主をリンク」 | owner group が作成される。画面を閉じて再度開き owner を再選択すると reverse lookup で既存 group が解決され、member ごとの unlink が有効になる。session 内 state だけに依存しない。audit に owner link create（ID のみ・PHI なし） |
| 3 | 親 owner group を前提に、両医院の対象ペットを選び「ペットをリンク」 | pet group 作成成功。actor が親 owner の全医院をカバーしていない場合は Forbidden でゼロ書き込み |
| 4 | 選択ペットの「連携履歴」を押す | `include_linked` で相関ペットの治療履歴（または「（履歴なし）」）が表示される。他 clinic の非リンクペットは出ない |
| 5 | 意図的に誤ったペット組み合わせで link した場合、メンバー unlink | soft-delete。最終メンバー unlink 時は group soft-delete。audit に unlink（IDs のみ） |
| 6 | 正しいペット組み合わせで relink | Phase 1 UI は `POST /pet-groups` のみ。最終メンバー unlink 後は group が soft-delete されるので relink は **新規 group**。既存へ add するボタンは無い |
| 7 | view のみ権限の別アカウントで同 URL を開く | 閲覧可・link/unlink ボタン非表示。mutation API 直叩きは 403 |
| 8 | 親 owner group anchor 医院に所属しないアカウントで CreatePetGroup を試す | 403 Forbidden。DB に pet group / members / audit が増えないこと |

## 確認観点

- **全医院セット認可**: mutation は parent owner anchor（`CreatedClinicID`）+ 全 active owner member clinics + 対象 pet clinics を要求（`assertActorCoversOwnerGroupClinics`）。any-member フォールバックなし。親 owner member 1 院欠けでも CreatePetGroup は Forbidden・ゼロ書き込み（回帰: `TestCreatePetGroup_RejectsMissingParentOwnerMemberClinic_NoPartialWrite`）。
- **view / edit 分離**: GET/search/history は actor clinic でフィルタ。link/unlink は `identity-links:edit` 必須。`identity-links` は staff へ自動付与されない（運用で明示付与 — clinic_service）。
- **原子性**: mixed / hidden / cross-clinic ID は全体 reject・部分書き込みなし。audit 失敗は business write と同一 tx で rollback。
- **非 PHI audit**: audit payload は group_id / clinic_id / owner_id / pet_id のみ。氏名・電話を含めない。
- **API 表面**: `POST/DELETE /api/v1/identity-links/owner-groups…` / `pet-groups…`、treatment-history は `GET …/pets/:clinicId/:petId/treatment-history?include_linked=`（`frontend/src/features/identity-links/api/identity-links-api.ts`）。FE ルート `/identity-links`。
- DB ロールでの実行証明は本シナリオ外。実行結果は ignored run report にだけ記録する。

## 実行結果の記録

実施日・環境・実施者・手順結果・不具合・承認者は `reports/uat-YYYY-MM-DD/` に記録する。scenario source に sign-off outcome を書かない。

## 実装突合
- 変更:
  - parent owner 全医院セット認可（anchor + members）を現行 identitylink サービス／回帰テストと突合
  - API パス・FE `/identity-links`・`include_linked` history・resource 非自動付与を明記
  - 手順 1–8 の期待は実装と一致（大きな手順変更なし）
