# EMR-284 検索欄棚卸し

Attempt `emr284-r2a1`. Plan C（Plane comment `a7afd297-51f3-49ae-9ab8-1ec6d10e630c`）は、欄の順序を揃えるのではなく、飼主名・ペット名・かな・電話を横断する単一キーワードにする。このリビジョンが編集してよい残り 2 欄は飼主名の次にペット名とする。新規予約と入院ホテルは棚卸しのみ。

このリビジョンは画面を編集しない。旧い並び（ペット名・飼主名・読みの別欄、または飼主名・ペット名の別欄）は、名前の付いた画面には残っていない。患者選択は既に Plan C の 1 欄である。一覧の 1 欄は一致先が画面ごとに違い、電話やかなの専用列まで揃えるのはバックエンド検索の変更になる。それは allowlist 外であり、`backend/internal/owner` と `backend/internal/medicalrecord` は同じ worktree の EMR-273 が触っている。2026-10-01 の横断検索障害の型を UI の文言だけで広げない。Phase 2 とする。

## 画面

| Screen | Current fields | Placeholder | Plan C delta | This revision edits |
| --- | --- | --- | --- | --- |
| new reservation | 新規予約の患者選択は 3 欄。`search`（ペット名、ペットかな、飼主名、飼主かな、電話）、`ownerId`（飼主No）、`species`（種別）。実装は `frontend/src/components/shared/ReservationFormModal/patient-selection-table-model.ts` の `TEXT_FIELDS` と `PatientSearchFilters.tsx`。予約管理カレンダーの担当医絞り込みは患者検索ではない。 | `search`: `例: もも、山田、090`。ラベル `検索（ペット名・飼主名・よみ・電話）`。飼主No: `例: 30042`。種別の未選択表示は `すべて`。 | キーワードは既に Plan C。飼主No と種別は追加の絞り込みであり、旧 5 欄ではない。 | No. Inventory only. `ReservationFormModal` と `frontend/src/features/reservations` は編集しない。 |
| periodic checkup | 一覧 `定期健診` は 1 欄。一致先は表示中ページの `petName`、`ownerName`、`checkupTypeName`、`result`（`filterCheckupsBySearch`）。かなは表示文字列の畳み込みだけで、電話列はない。ペット選択 `定期健診登録 - ペット選択` は共有 `PetSelectionSearchForm`（下記）。`CheckupForm.tsx` の `健診種別・所見で検索...` は過去記録の検索であり、飼主・ペット欄ではない。 | 一覧: `ペット名・飼主名・種別で検索...`（`CheckupsListPanels.tsx`）。ペット選択: `例: もも、山田、090`。 | ペット選択は Plan C。一覧は 1 欄だが電話と専用かな列を見ない。ページ内フィルタのまま電話を足すことはできない（`CheckupRecord` は `petName` と `ownerName` のみ）。 | No. 一覧の述語拡張は Phase 2。`CheckupForm.tsx` と `CheckupForm.test.tsx` は deferred。 |
| inpatient hotel | 一覧は 1 欄。一致先は表示中ページの `ownerName`、`petName`、`hospitalizationNo`（`applyHospitalizationClientFilters`）。かなは表示名の畳み込みのみ。電話はない。ペット選択 `入院・ホテル登録 - ペット選択` は共有 `PetSelectionSearchForm`。 | 一覧: `飼主名、ペット名、入院No...`（`HospitalizationListPanels.tsx`）。ペット選択: `例: もも、山田、090`。 | ペット選択は Plan C。一覧は 1 欄だが電話と専用かな列を見ない。 | No. Inventory only. 入院ホテルの検索行は是正対象外。 |
| owner and pet list | 飼主・ペット一覧は 1 欄。`OwnersList.tsx` が確定値を URL `search` にし、`loaders.ts` が owners API へ送る。サーバ `owner.FindAll` の一致先は飼主 `name`、`name_kana`、`phone`、`email` と、未コミットの EMR-273 差分が足した飼主 `name` の空白除去腕（`repository.go` の compact 腕）。ペット名・ペットかな・飼主No・ペット番号は述語に無い。この backend 差分は EMR-284 が書いていない。`OwnerPetsSection.tsx` に検索 input は無い（ペット表の見出しと操作のみ）。 | `飼主名、ペット名、電話番号、飼主No、ペット番号...`（`OwnersListTable.tsx`）。 | 1 欄ではある。placeholder はペット名・飼主No・ペット番号を名乗るが、現行述語は飼主名・かな・電話・email である。ペット名を含む Plan C には未達。 | No. placeholder のファイルは deferred。述語は `backend/internal/owner`（EMR-273、allowlist 外）。`OwnersList.tsx` と 2 本の route test は欄を描画しない。 |
| chart | カルテ一覧は 1 欄の横断検索に加え、列フィルタ `owner_name`、`pet_name`、`chief_complaint`（この順、EMR-245）。横断検索の一致先はカルテNo、飼主名、飼主かな、ペット名、ペットかな、主訴（`applyMedicalRecordSearch`）。電話は含まない。新規カルテのペット選択は共有 `PetSelectionSearchForm`。`MedicalRecordStickyHeader.tsx` に患者検索欄は無い。 | 横断検索: `診療内容を横断検索（カルテNo・飼主名・ペット名・主訴）...`（`MedicalRecordsListPanels.tsx`）。列フィルタに placeholder は無く、ラベルは `飼主名` の次が `ペット名`。ペット選択: `例: もも、山田、090`。 | ペット選択は Plan C。横断検索は電話が無い。残る列フィルタは既に飼主名の次がペット名であり、このリビジョンが並べ替える 2 欄ではない。列フィルタの削除は EMR-245 の取り消しであり、電話腕の追加は `backend/internal/medicalrecord`（EMR-273）になる。 | No. Phase 2。`MedicalRecordStickyHeader.test.tsx` は deferred。 |
| lab | 検査管理（`ExaminationsList` の title）は 1 欄。一致先は表示中ページの `ownerName`、`petName`、`testType`（`use-examination-records.ts`）。複数ページではページ内だけである。検査のペット選択は共有 `PetSelectionSearchForm`。`ExaminationFormHeader.tsx` に検索 input は無い。`frontend/src/features/lab-device` は機器ボードであり、飼主・ペット・かな・電話の検索欄は無い。 | 一覧: `飼主名、ペット名、検査種別...`（`ExaminationsListPanels.tsx`）。ペット選択: `例: もも、山田、090`。 | ペット選択は Plan C。一覧は 1 欄だが電話と専用かな列を見ず、取得済みページだけに効く。lab-device に Plan C の差分は無い。 | No. Phase 2 は検査一覧の述語だけ。`ExaminationFormHeader.tsx` は deferred。lab-device は編集しない。 |
| trimming | トリミング一覧は 1 欄。一致先は取得範囲内の `ownerName`、`petName`、`breed`（`useFilterTrimmingRecords`）。電話は無い。ペット選択は共有 `PetSelectionSearchForm`。`TrimmingFormColumns.tsx` は選択済み患者の表示であり、検索欄は無い。 | 一覧: `飼主名、ペット名、犬種...`（`TrimmingListTable.tsx`）。ペット選択: `例: もも、山田、090`。 | ペット選択は Plan C。一覧は 1 欄だが電話と専用かな列を見ない。 | No. Phase 2。deferred: `TrimmingFormColumns.tsx`、`TrimmingFormColumns.test.tsx`、`use-trimming-form-chrome.ts`、`use-trimming-form-chrome.test.ts`。 |
| vaccination | 予防接種一覧は 1 欄でサーバ検索。一致先はペット `name`、飼主 `name`、ワクチン `name` の畳み込みと、未コミットの EMR-273 差分が足した飼主 `name` の空白除去腕（`vaccination_repository.go`）。ペットかな、飼主かな、電話は腕に無い。この backend 差分は EMR-284 が書いていない。ペット選択は共有 `PetSelectionSearchForm`。`VaccinationFormPagePanels.tsx` に検索 input は無い。 | 一覧: `飼主名、ペット名、予防接種名...`（`VaccinationListPanels.tsx`）。ペット選択: `例: もも、山田、090`。 | ペット選択は Plan C。一覧は 1 欄だがかなの専用列と電話を見ない。述語変更は `backend/internal/medicalrecord` のワクチン検索であり allowlist 外かつ EMR-273 の領域。 | No. Phase 2。`VaccinationFormPagePanels.tsx` は deferred。 |

## 共有の患者選択

定期健診、入院ホテル、カルテ、検査、トリミング、予防接種のペット選択は `frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx` の `FIELD_DEFS` を使う。欄は `search`（ラベル `検索（ペット名・飼主名・よみ・電話）`、placeholder `例: もも、山田、090`）、`ownerId`（`例: 30042`）、種別。コメントのとおり `search` は `pets.name` / `pets.name_kana` / `owners.name` / `owners.name_kana` / `owners.phone` を横断する。この共有ファイルは allowlist 外なので編集しない。新規予約は同じ形を `ReservationFormModal` 内に持っており、こちらも編集しない。

`frontend/src/features/owners/components/OwnerPetsSection.tsx` に、飼主名・ペット名・かな・電話の検索欄は無い。差分がこのファイルにだけある、という状態ではない。

## Phase 2

残る差分は欄の本数ではない。一覧キーワードの一致先を、飼主名・ペット名・かな・電話の四つへ実際に広げる仕事である。対象は次のとおり。

- 飼主・ペット一覧: ペット名（とかな）が述語に無く、placeholder が過大。`backend/internal/owner`。
- カルテ一覧: 横断検索に電話が無い。列フィルタは EMR-245 のまま残す。`backend/internal/medicalrecord`。
- 予防接種一覧: かなの専用列と電話が無い。同じ medicalrecord パッケージのワクチン検索。
- 定期健診・検査・トリミング・入院一覧: 電話列を持たないページ内（または取得範囲内）フィルタ。サーバ述語へ変える変更であり、UI の placeholder 差し替えでは閉じない。入院は棚卸しのみ。

見積は棚卸しの 1 時間を超える。検索述語、医院分離、ページング、既存の畳み込みテストを各画面で直す必要があり、2 日を超える見込みである。このリビジョンでは始めない。

## Deferred

他 worktree で dirty なため未編集: `frontend/src/features/checkups/routes/CheckupForm.tsx`、`frontend/src/features/checkups/routes/CheckupForm.test.tsx`、`frontend/src/features/examinations/components/ExaminationFormHeader.tsx`、`frontend/src/features/medical-records/components/MedicalRecordStickyHeader.test.tsx`、`frontend/src/features/trimming/routes/TrimmingFormColumns.tsx`、`frontend/src/features/trimming/routes/TrimmingFormColumns.test.tsx`、`frontend/src/features/trimming/hooks/use-trimming-form-chrome.ts`、`frontend/src/features/trimming/hooks/use-trimming-form-chrome.test.ts`、`frontend/src/features/vaccinations/routes/VaccinationFormPagePanels.tsx`、`frontend/src/features/owners/components/OwnersListTable.tsx`。

未編集（禁止）: `frontend/src/components/shared/ReservationFormModal`、`frontend/src/features/reservations`、`frontend/src/features/owners/components/OwnerPetsSection.tsx`、`frontend/src/hooks/use-update-reservation.ts`、`backend/internal/reservation`、`frontend/src/features/shifts`。owners は `OwnersList.tsx`、`OwnersList.test.tsx`、`OwnersList.species-status.test.tsx` 以外を編集していない。これら 3 ファイルもこのリビジョンでは未変更。
