# EMR-569 締め設定の分入力と保存済み秒値の保全

2026-10-24、基準main `4b7ea65a3b692073bb41c4be52bd2c1a9be99d74`。採択済み締め設定の原値保持を具体化する技術選定であり実装済みではない。canonical DB再構築・old_db mappingも確定仕様へ適合する前提で、legacy救済・非00秒専用conversion UI・backfill/runtime修復は追加しない。実装受領先は親452の[EMR-569](https://app.plane.so/baritechllc/browse/EMR-569/)、採用docsリンク待ちの一時Backlog。509は境界保護、448は共有writer順、189は履歴参照。

## 標準入力のstate

各時刻はrawOriginal（GET原文字列）/draftMinute（HH:MM又は空）/dirtyを分離。正当HH:MM/HH:MM:SSの表示をHH:MMへ投影しただけでdirtyにしない。通常分入力で実際の分値が変わったfieldだけdirty。同じ分の再選択/再入力no-opや元の分へ戻す操作はdirty=false、非00秒を00へclearしない。

rawは実値として画面・秒previewへ保持し予想外だが正当な非00秒も未編集なら保持。明示変更は既存保存要約で旧raw→新HH:MM:00を表示できるようにするがfield別checkbox/切替button/modal/承認workflowなし。取消はraw基準へ戻す。標準canEdit=falseは入力/保存不可、label/aria-describedby/errorを維持。不正現在値は取得不整合として保存停止、default補完/切捨て/推測修復なし。実データ変換・old_db編集・DB再作成は本scope外。

## dirty-only PATCHとpreview

既存updateClosingSettings optional requestを使いdirty=false時刻keyは省略。保持代用のnull/空/toTimeInputValue(rawOriginal)を送らず、分が実際に変化したfieldだけHH:MM送信。closed_weekdaysも変更時だけ、取消/無変更保存はnetwork writeなし。既存UpdateStandardのclinic row lock→read→pointer有りだけ変更→validate→save→audit TXを維持し、hidden原値の強制再送で並行変更を上書きしない。

previewは未編集raw、編集中の正当draft HH:MM、空/不正draftは未確定。computeClosingTimeRangesの秒計算を維持しtoTimeInputValueをraw正本としない。固定AM開始は編集不可、特別期間は標準AM開始継承。SpecialPeriodSectionの既存canCreate/canDeleteを独立維持し標準canEditへ統一しない。特別期間の新編集UI/PATCHなし。

成功は返却/GETからraw基準を再設定しdirty clear、失敗はraw/draft保持・成功toastなし。編集中refetchで無言上書きせず利用者が再読込/取消で確認できる。既存validateStandardClosingTimesのParseHHMM分比較、billing resolvePeriodRangeの分解釈を維持する。global parser/締め集計/過去snapshotを全面秒対応へ変更せず、同分内非00秒境界で既存validatorが拒否するなら拒否を保持し無断丸めで通さない。

## 一つの実装PR

StandardClosingTimeSection.tsxの初期値/formAction/controls/preview、closing-time-ranges.ts表示helper、update-closing-settings.ts optional payload、SpecialPeriodSection.tsxの分入力/保存後raw previewを同PR照合。必要pure helperはfrontend/src/features/closing-settings/lib/closing-time-edit-state.ts。BEは既存pointer省略保持fixtureを追加し不必要なAPI/DDL変更なし。共有hunk/個人claimは着手時確認、方式未選定待ちではない。採用docs固定SHAを記録してReady判定可、実装供給/完成とは別。

fixtureは非00秒＋曜日だけ更新→時刻省略/DB原値保持、1fieldだけ分変更→他2省略、同分no-op/元分へ戻す→clearなし、省略、取消zero-write、:00無変更zero-write、失敗保持、編集中refetch、不正取得fail closed、標準readonlyと特別canCreate/canDelete独立、HH:MM入力とHH:MM:SS preview、EMG越日/境界1秒前の既存表示・集計不変。BE既存並行partial PATCH/監査rollback fixtureを保持。原値保持と集計境界は別検証、全実行NOT_RUN。製品/API YAML/DDL/生成/DB/Docker/STGは変更・実行しない。
