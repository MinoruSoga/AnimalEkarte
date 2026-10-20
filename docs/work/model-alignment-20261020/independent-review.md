# 独立した静的レビュー

code-reviewスキルのStandards軸とSpec軸を別担当・別worktreeで実施。基準は7f164f8a71baf5bfb23207ddce5b3049939e81a3と未commit全差分（新規ファイルを含む）。生成表の数だけで判定せず、実際の001のALTER/DROP・モデル・既存consumer/driverを読取確認した。修正後は同じ担当が対象箇所を再確認した。DB・アプリ・試験を実行したレビューではない。

## Standards

確定指摘1件（P2）。既存mutableモデルのUpdatedAtへautoUpdateTime:falseを付けると更新時刻が進まなくなる。Clinicを例に、clinic_repository.goのUpdates(fields)はupdated_atを明示せず、最終DDLにも更新用triggerがない。DEFAULT now()はUPDATEを代替しない。単なるSQL型整合から既存のORM動作変更は導けない。

修正：元から存在したcreated_at/updated_at/recorded_atの212フィールドは元の明示・暗黙auto設定を保持。新規モデル／新規列の扱いと分け、生成元も訂正した。担当はClinic/Account/Staff/Owner/MedicalRecord/ExaminationRevision/AuditLogを再読し、回帰解消・残る指摘なしと回答した。

別観点で、最終DROP TABLE/COLUMN/TYPE、共通基本情報revisionへの移動、複合PK/FK、配列/JSON/UUID/NULL、既存pgxのScanner/Valuer実装を読取確認。pgxのdatabase/sql numeric decodeはstringであり、委譲先は精度を保持する。実DB往復の合格とはしない。この範囲で追加の確定的問題は検出しなかった。会計差分には同等性を検証済みへ引き上げた記述を見つけなかった。

## Spec

確定指摘2件。

- P2：reservations/technical-spec.mdの既存「区分＋コース＋オプション」が、新しい「コースまたは区分」と矛盾。既存文をOR基準へ直し、コース60＋10＋0＝70分の例で二重加算しないことを明示した。
- P3：同書の「BE modelも未実装」が今回の保存モデル追加と矛盾。モデルの保存形状対応とservice/wire/FE等の未実装・未検証を分け、EMR-480の引継ぎも明記した。

担当は両修正を再読し、解消・残る指摘なしと回答した。別観点でcleanupの直接DROP COLUMN112件の残存mapping、廃止9表、廃止enum型タグ、末尾billing完了キー/treatment subtotal削除、deposit訂正列追加、数量numeric化、検査日NULL化を読取照合し、この範囲で追加の不整合を検出しなかった。既存文書17ファイルの差分に、計算完全一致の検証済み扱いや私的金融資料の新規転載は検出しなかった。scope creepの指摘なし。

Standards 1件（最悪P2）、Spec 2件（最悪P2）。3件とも修正・再確認済み。これは上記範囲の静的レビューで、全writer/API/UI/実機・DB・リリース適合を保証しない。

## 旧調査案再掲の注意を受けた追加の局所レビュー

初回の3指摘とは別に、本人回答の撤回ではない旧案再掲を確認した。カルテ設計とリンク先財務設計に残存したbilling_confirmations利用案・金額変更時再確認要求を1件の古い設計契約として訂正。Spec担当は同じ隔離review worktreeで2文書の修正を再読し、最終DDL接続と4つの本人決定、金額変更/臨床変更の区別、API/DDL未実装の境界を確認し、範囲内の指摘なしと回答した。元の17文書レビューとこの局所追加確認を区別する。

## PR #729 指摘への追加修正の再確認

比較点636f79f6bからのモデル・ERD差分を、同じ隔離worktreeの別担当で再確認。Standards 0件、Spec 0件。既存Go named-string10項目とSQLタグ保持、unique6組のSQL制約名/列順/非衝突、識別子3箇所、241/246の区別、廃止列を復活させず他層へ広げない境界を確認した。報告資料の同期前のモデル・ERD差分を対象にした読取レビューであり、同期後の対応表は親の静的照合で確認した。旧テスト・機能consumer・DB適合の合格ではない。
