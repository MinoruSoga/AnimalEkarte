# FE-refactor.md — 残バックログ（第 4 期以降）

- **更新日**: 2026-07-12（#201 薬量自動計算の FE 実装が完了したため残件 1 をクローズ）
- **完了済み（本文に残さない・git が正本）**: FE4-1〜18 / FE5-1〜4 / FE6-1〜9 / FE7-1〜3 / FE8-1〜4 / LinkOwner cross-clinic（`44e35b3b`）/ avatar_url 削除（`48c5b084`+`deb9d1bc`）/ meClinicInfoSchema `.default()` 撤去（`fa3a95d0`）/ liff pet_id FormatUint（`aef2d6ff`）/ query-keys registry 全面採用 + ESLint（`git log --grep='FE-R1'`）/ #201 薬量自動計算 FE 実装（5 Phase・型追随・`lib/medicine-dose.ts`・薬マスタ dose-params UI・カルテ TreatmentsTab hard gate）
- **クローズ・PO 決定の記録**: 本書には残さない。正本はメモリ `fe_backlog_decision_pack_20260711.md`
- **本書の規約**: 行動可能な未対応タスクのみを記載する。各項目は「判断が必要な点」を明示し、判断が下れば追加調査なしで着手できる粒度にする
- **#201 フォローアップ（新規発見・別チケット化推奨）**: `backend/internal/handler/treatment_response.go` の `toTreatmentResponse` が `dose_weight_kg`/`dose_amount_mg`/`dose_param_snapshot` 等を JSON に含めていない（DB には保存されるが API 応答に出ない）。そのため保存済み投与量根拠の「丸め前/丸め後 mg」表示はカルテ側の都度再計算プレビューのみで、再読込後に保存済みスナップショットを表示する監査 UI は BE レスポンス追加まで実装できない。BE 変更は本タスクの Out of Scope。

---

## 残件 2: `NextRecommendedVisitDate` 常時 null — PO 判断待ち（選択肢・実コスト確定済み）

（残件 1 は #201 として本更新でクローズ済み。番号は既存メモリ参照との対応を保つためそのまま維持する）

### 現状（2026-07-12 実測）

- `liff_service_health_card.go:23` で宣言のみ・未代入。契約上は必須フィールド（`api.yaml:7219` の required に含まれ nullable）。
- LIFF 健康手帳（`PetHealthPage.tsx`）のペットカードは上から「最終来院日 → **次回来院推奨日（常に「なし」）** → ワクチン記録表」の順で、**ワクチン表には既に列「次回予定日」（`next_due_at`）が表示されている**。つまり現状の画面は「次回来院推奨日: なし」のすぐ下に個別ワクチンの次回日付が並ぶ、という不整合な見え方をしている。
- 集約元の `GetHealthCard` は `vaccinationRepo.FindByOwner` で `NextDueAt` を既にメモリ上に持っている — 選択肢 (B) の材料は追加クエリなしで揃っている。

### 選択肢と実コスト

**(A) フィールド削除** — 「ワクチン表の次回予定日で足りる」という判断の場合。
- タッチ箇所は 6 点で確定: `liff_service_health_card.go:23`（struct フィールド）/ `liff_response.go:240,267`（response 変換）/ `api.yaml:7207` + required 行 `:7219` / `liff-api.ts:22`（zod）/ `PetHealthPage.tsx` の表示 6 行。
- 注意: api.yaml と Go response の同時変更になるため **OpenAPI Drift Gate**（FE8 で `last_visit_date` の既知 drift を allowlist 記録した経緯 `70f4c298` あり）を通ることを確認してから push。
- 工数: 最小（1 コミット）。PRODUCT_PHILOSOPHY ②削除に整合。

**(B) ワクチン由来で算出** — 「ペット単位のサマリー日付」に価値がある場合。
- 実装: `GetHealthCard` のペットループ内で `min(NextDueAt where NextDueAt > now)` を代入。追加クエリ・API 変更・FE 変更**すべて不要**（契約は nullable のまま、FE は値が来れば表示する）。実装 ~10 行 + service unit test。
- 制約を仕様として明記する必要がある: 「ワクチンのみ由来。フィラリア・ノミダニ・健診は含まない」。過去日しか無いペットは null のまま（期限切れを『推奨日』として出すかは別判断 — 出すなら「超過」表示が必要で工数が跳ねる）。
- リスク: 欄名「次回来院推奨日」が実態「次回ワクチン予定日」より広い約束に見える。**名前を「次回ワクチン予定」に変えるなら (A) との差がほぼ消える**ことに注意。

**(C) 予防スケジュール統合版** — 健康手帳を予防管理のハブにする場合。
- 既存資産: lstep 側にワクチン期限計算（`hasVaccineDeadlineSoon`・`TriggerVaccineDeadline60/30`・フィラリア/ノミダニ tag sync = H1 トラック）が存在する。ただし全て **owner 単位の「期限が近いか」判定**であり、**pet 単位の「次回日付の算出」ロジックは存在しない** — 転用ではなく新規設計になる。
- 健診周期はデータ源自体が未定義。工数: 大（プロダクト仕様策定から）。

### 推奨と PO への質問

- **質問は 1 つ**: 「この欄は飼い主に何を約束する欄か」。回答 → (ワクチン表で十分)=A / (ペット単位の直近予定サマリー)=B / (予防全体の来院管理)=C。
- **推奨は (A)**。現画面はワクチン次回日を既に表示しており、二重表示欄を null のまま置く価値がない。(B) を選ぶ場合も欄名の再検討（上記リスク）まで含めて決定すること。
- 回答が出るまで現状（null 表示）維持。捏造値の表示だけは絶対にしない。
