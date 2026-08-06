# ブラウザ検証 手順書（単一入口）

| 項目 | 値 |
|------|-----|
| **作成** | 2026-08-06 |
| **対象** | `STATUS.md` §3 の **IMPLEMENTED_UNVERIFIED 32 件**（原文シナリオ再検証） |
| **結果表（記入先）** | [`BROWSER_VERIFICATION_BACKLOG.md`](./BROWSER_VERIFICATION_BACKLOG.md) |
| **実装状態の正本** | [`../STATUS.md`](../STATUS.md) §3 |
| **詳細シナリオ原文** | [`../docs/ops/testing/scenarios/`](../docs/ops/testing/scenarios/) |
| **環境前提（local 2026-08-06）** | `make reset` 済み · `exam_reference_ranges`=20 · stack healthy |
| **residual 方針** | **2026-08-06 以降 residual closeout 対象外**（USER）。保管用手順書。 |

> **この文書の役割**  
> ブラウザで何を・どの順で・どう判定するかを **1 ファイルに集約**する。  
> **現状 residual の必須作業ではない**（除外）。任意実施時のみ使う。  
> 結果の UNREPORTED/PASS/FAIL は **バックログ表** に書く。`VERIFIED_FIXED` は人間判断のみ（エージェントは付けない）。
---

## 0. ルール（必読）

| ルール | 内容 |
|--------|------|
| 実装 vs ブラウザ | 実装完了は IU のまま正。ブラウザ未実施でもコード完了を取り消さない |
| 結果の記入 | 本手順の実施後、必ず [`BROWSER_VERIFICATION_BACKLOG.md`](./BROWSER_VERIFICATION_BACKLOG.md) の対応行を更新 |
| VERIFIED_FIXED | ブラウザ PASS かつ回帰なしのとき、**人が** `STATUS.md` §3 個票を更新してよい |
| FAIL 時 | バックログに FAIL + メモ → 個票を OPEN に戻すか新規 BUG 起票 |
| 秘密 | パスワード・トークン・cookie を文書・スクショ説明に書かない |
| BUG-009 | **WAIVED**（2026-08-05 USER）。本 runbook では実施不要（表に WAIVED とあることを確認するだけ） |

### 結果コード

| コード | 意味 |
|--------|------|
| UNREPORTED | 未実施 |
| PASS | 下表の **合格条件** を満たした |
| FAIL | 再現 / 回帰あり |
| BLOCKED | 環境・データ・権限不足（理由をメモ） |
| WAIVED | 意図的見送り（BUG-009 のみ確定） |

---

## 1. 実施前チェック（5 分）

作業ディレクトリ: AnimalEkarte リポジトリ root。

```bash
# 1) スタック
docker compose --env-file .env.local ps
# 期待: db / backend / frontend が healthy

curl -sS -o /dev/null -w "backend=%{http_code}\n" http://127.0.0.1:8080/health   # 200
curl -sS -o /dev/null -w "frontend=%{http_code}\n" http://127.0.0.1:3003/          # 200

# 2) BUG-003 用 seed（H/L）
# .env.local の DB_USER / DB_NAME を使って:
docker compose --env-file .env.local exec -T db \
  psql -U "$DB_USER" -d "$DB_NAME" \
  -c 'SELECT COUNT(*) FROM exam_reference_ranges;'
# 期待: 20 以上。0 なら docs/ops/deploy/LOCAL_DB_RESET.md の make reset（USER）

# 3) LIFF mock（BUG-008 / 014）
# docker-compose.yml で LIFF_MOCK / VITE_LIFF_MOCK が true であること（main 固定済み）
# 変更直後なら: docker compose --env-file .env.local up -d --wait backend frontend
```

| 項目 | 推奨値（local） |
|------|-----------------|
| UI | http://localhost:3003 |
| API | http://localhost:8080 |
| ブラウザ | Chrome（DevTools Network を開く） |
| ログイン | デモ管理者（例: `admin@example.com`）。**パスワードは `.env.local` の DEV_ADMIN_* 等。文書に書かない** |
| seed | `003_demo`（reset 済み想定） |
| LIFF | mock 有効（実 LINE は STG 別セッション） |

### 実施メタ（バックログ先頭表に転記）

| 項目 | 記入例 |
|------|--------|
| 日付 | YYYY-MM-DD |
| 実施者 | 名前 |
| URL | http://localhost:3003 |
| seed | 003_demo（reset 2026-08-06） |
| ブラウザ | Chrome |
| ログイン | admin@…（秘密は書かない） |

---

## 2. 実施順（推奨）

**所要の目安**: 優先バッチ A 約 45–90 分 · 全体 B まで 半日〜1 日。

| 波 | 対象 | 理由 |
|----|------|------|
| **A（優先）** | BUG-003, 006, 007, 008, 012, 014, 024, 026, 029, 031, 032 | 本 campaign / ブロッカー / seed・mock 依存 |
| **B（一括）** | 残り IU（009 は WAIVED 確認のみ） | 先行実装の原文シナリオ |
| **C（任意）** | 原文シナリオ全文 S01–S12 / V01–V05 | 納品 UAT（#254）と重複。必要なら scenarios/ へ |

同じ画面をまたぐものは連続実施（例: S01 の 001→002→021→022、S02 の 005→003→004→017）。

---

## 3. 共通操作

1. シークレットでない通常ウィンドウで http://localhost:3003 を開く  
2. ログイン（失敗したら backend ログと `/health` を確認）  
3. 各 BUG で **操作 → 画面上の期待 → Network で API 成否** をセットで見る  
4. 破壊的操作（死亡登録・会計確定・マスタ重複）は **デモデータ上** で行い、必要なら後で `make reset`  
5. 結果をバックログの「結果」「メモ」列に即記入  

---

## 4. 波 A — 優先バッチ（手順）

### A-1. BUG-003 — 検査 H/L ハイライト【重大】

| | |
|--|--|
| シナリオ | [S02](../docs/ops/testing/scenarios/S02-exam-abnormal-highlight-lock.md) 手順 2–4 |
| 画面 | `/examinations` → 検査詳細（CBC 等で ref がある項目） |
| 前提 | `exam_reference_ranges` ≥ 20 |

**手順**

1. 検査一覧から編集可能な検査を開く（または新規で CBC 系を入力）。  
2. 基準より高い値（例: WBC 高値）と低い値（例: RBC 低値）、境界ちょうどを入力。  
3. 保存後、一覧/詳細で H（高）・L（低）・正常のハイライトを確認。

**合格**: 高=赤系 H、低=青系 L、境界ちょうど=正常。常に「未判定」だけにならない。  
**原文**: S02。

---

### A-2. BUG-006 — 予防接種ヘッダーの患者属性【重大】

| | |
|--|--|
| シナリオ | [S03](../docs/ops/testing/scenarios/S03-vaccination-next-due-autocalc.md) 手順 1 |
| 画面 | `/vaccinations/new?petId=<id>` |

**手順**

1. 飼主画面でペット A の年齢・性別・去勢/避妊をメモ（正本）。  
2. `/vaccinations/new?petId=<A>` を開き、ヘッダーの年齢・性別・去勢が A と一致するか。  
3. 別ペット B で同様にし、A の固定値が残っていないか。

**合格**: ペットごとに正本と一致。全ペットで同じ固定値（例: 常に同じ年齢）にならない。

---

### A-3. BUG-007 — 予防接種登録後の履歴表示

| | |
|--|--|
| シナリオ | S03 手順 6–7 |
| 画面 | `/vaccinations/new?petId=...` · `/vaccinations` |

**手順**

1. 対象ペットで接種を 1 件登録（成功トースト）。  
2. 一覧 `/vaccinations` に当該行が見えるか（ページ・日付に注意）。  
3. 同じペットで新規画面を再度開き、「過去の接種履歴」に直近 1 件が出るか。  
4. 任意: Network で `GET .../vaccinations?pet_id=...` が 200 で件数 > 0。

**合格**: DB 保存だけでなく **UI 履歴・一覧** に即時反映。

---

### A-4. BUG-012 — 顧客集計ダッシュボード【重大】

| | |
|--|--|
| シナリオ | [S10](../docs/ops/testing/scenarios/S10-customer-aggregation-consistency.md) 手順 1 |
| 画面 | `/aggregation` |

**手順**

1. `/aggregation` を開く。  
2. 30 秒以上「読み込み中...」のまま固着しないこと。  
3. 売上ランキング / CPM 等の主要ブロックが表示されること。  
4. Network で集計 API が完了（200 または明確なエラー）。ハングしない。

**合格**: 画面が描画完了し API が応答する（タイムアウト無限待ちなし）。

---

### A-5. BUG-024 — 権限マトリクス保存

| | |
|--|--|
| シナリオ | [V03](../docs/ops/testing/scenarios/V03-owner-pet-staff-forms.md) §6 チェック 2 |
| 画面 | 設定 → 権限グループ（permission group side panel） |

**手順**

1. 任意グループを開き、マトリクスで **表示/作成/編集/削除** の 1 セルを OFF。  
2. 保存 → 成功表示。  
3. パネルを閉じ再オープン、または再 GET 後に OFF が保持されること。  
4. 自己剥奪ガード対象なら、拒否メッセージで保存されないこと（該当時）。

**合格**: 変更が永続化。成功トーストだけ出て中身が戻る、が無い。

---

### A-6. BUG-026 — 保険マスタ 補償率境界

| | |
|--|--|
| シナリオ | [V04](../docs/ops/testing/scenarios/V04-settings-master-forms.md) §1 保険 C1-3 |
| 画面 | `/settings/insurance` |

**手順**

1. 新規または編集で補償率 **101**（または -1）を入力して保存。  
2. Network: 不正値で POST/PATCH が飛ばない、または 4xx。  
3. UI: エラー表示。**「登録しました」成功トーストが出ない**。

**合格**: 範囲外は保存されず、偽成功がない。

---

### A-7. BUG-029 — 支払方法 名称重複

| | |
|--|--|
| シナリオ | V04 §1 支払方法 C3-2 |
| 画面 | `/settings/payment-methods` |

**手順**

1. 既存の支払方法名と同一名で新規保存を試す。  
2. 成功トーストが出ないこと。エラーが表示されること。  
3. 一覧件数が増えていないこと。

**合格**: 一意制約違反がユーザー向けエラーになり、偽成功なし。

---

### A-8. BUG-031 — ログイン済み `/login` リダイレクト

| | |
|--|--|
| シナリオ | [V05](../docs/ops/testing/scenarios/V05-auth-line-forms.md) V05-1 #3 |
| 画面 | `/login` |

**手順**

1. 通常ログイン済みの状態にする。  
2. アドレスバーで `/login` に直接遷移。  
3. ログインフォームが残らず、アプリ内（ダッシュボード等）へリダイレクトされること。

**合格**: 認証済みで login 画面に留まらない。

---

### A-9. BUG-032 — 健診対象者プレビューがハングしない

| | |
|--|--|
| シナリオ | V05-18 |
| 画面 | `/lstep/checkup-sync` |

**手順**

1. 健診対象者抽出 / プレビューを実行。  
2. 妥当な時間内（目安 30s 以内）に一覧・空・エラーのいずれかが返る。  
3. スピナーが永続しない。Network でタイムアウトまたは完了を確認。

**合格**: 無限ハングしない（外部 L ステップ障害時も明確な失敗）。

---

### A-10. BUG-008 — LIFF 予約 コース選択【S04 ブロッカー】

| | |
|--|--|
| シナリオ | [S04](../docs/ops/testing/scenarios/S04-liff-reservation-journey.md) 手順 1–2 |
| 画面 | `/line-reserve/1/`（line-reserve アプリ） |
| 前提 | `LIFF_MOCK=true` / `VITE_LIFF_MOCK=true`（compose 済み） |

**手順**

1. ブラウザで `/line-reserve/1/` を開く。  
2. コース選択まで進む。  
3. 「ログイン情報の有効期限が切れました」で止まらないこと。  
4. Network: `GET /api/liff/.../courses` が **200**（またはコース一覧 UI 表示）。

**合格**: mock 下でコース一覧が表示され、後続ステップに進める。  
**BLOCKED 例**: mock 無効・別オリジン・STG で real LINE 未用意。

---

### A-11. BUG-014 — LIFF ペットヘルス

| | |
|--|--|
| シナリオ | [S12](../docs/ops/testing/scenarios/S12-liff-pet-health.md) |
| 画面 | `/liff/1?clinic_id=1` 等（liff アプリ · S04 とは別） |

**手順**

1. mock 有効でペットヘルス URL を開く。  
2. 401 / 期限切れメッセージで白紙にならないこと。  
3. 飼主名・ペットカード（名前・種/品種・最終来院・ワクチン表）が表示されること。

**合格**: カードが描画される。profile/health-card API が成功。

---

## 5. 波 B — 残り IU（手順）

各項目: **画面 → 操作の要点 → 合格条件**。詳細は scenarios 原文。

### 飼主・ペット（S01 / V03）

| BUG | 画面 | 操作の要点 | 合格条件 |
|-----|------|------------|----------|
| **001** | `/owners` | 検索に `姓 名`（スペース区切りフルネーム） | 該当飼主がヒット |
| **002** | `/owners/:id` | ペット編集 → 死亡記録 → モーダル閉じる | **リロード前**に一覧の生死が「死亡」 |
| **021** | 死亡ダイアログ | 日付空 / 未来日で確定 | 必須・未来日のエラーメッセージ表示 |
| **022** | 同上のち F5 | 死亡記録 → フルリロード → 再オープン | 死亡ステータス・日・理由が保持 |

### 検査（S02 / V01）

| BUG | 画面 | 操作の要点 | 合格条件 |
|-----|------|------------|----------|
| **005** | `/examinations/new?petId=...` | 担当医セレクタを開く | 医師のみ（スタッフ以外が混ざらない） |
| **004** | `/examinations/:id` | 未確定検査を初めて「確定」保存 | 成功し confirmed。誤った「確定済みで編集不可」で初回拒否されない |
| **017** | `/examinations/new?...` | 種別・担当医空で保存 | 保存されず **必須エラーメッセージ** が出る |

### 入院（S05）— BUG-009

| BUG | 扱い |
|-----|------|
| **009** | **WAIVED**。実施不要。バックログ結果列が WAIVED であることだけ確認 |

（任意で見るなら `/hospitalization` のタブ「予約 / 退院済 / すべて」がサーバ件数と一致するか。）

### カルテ（S06 / V01）

| BUG | 画面 | 操作の要点 | 合格条件 |
|-----|------|------------|----------|
| **010** | `/medical-records/:id` 診察/治療プラン | 所見・診断・方針を入力して保存 → 再読込 | 入力内容が保持（空欄化・固定文置換なし） |
| **015** | 同・バイタルモーダル | 体重を kg↔g 切替して保存 | 単位換算され 1000 倍破損しない |

### 見積・会計（S07 / S11 / V02）

| BUG | 画面 | 操作の要点 | 合格条件 |
|-----|------|------------|----------|
| **011** | `/estimates/new` | 同じペットで見積を 2 件目以降作成 | 409 `estimate '' already exists` にならない |
| **019** | `/estimates/999999999` 等 | 存在しない ID | 空フォームではなくエラー or 一覧へ |
| **013** | `/accounting/new?petId=...` | 未請求の診察+トリミングがあるペット | 500 にならず未請求明細が統合表示 |
| **018** | 会計確定後の明細追加 | レジ締め済み期間のデータ | 一貫してブロック or 締め後理由を要求（壊れた会計を残さない） |

### フォーム基盤（V01）

| BUG | 画面 | 操作の要点 | 合格条件 |
|-----|------|------------|----------|
| **016** | `/vaccinations/999999999` · `/examinations/999999999` · 入院の存在しない ID | 直 URL | 空の編集フォームではなく「見つかりません」系 |

### 予約モーダル（V02）

| BUG | 画面 | 操作の要点 | 合格条件 |
|-----|------|------------|----------|
| **020** | 予約登録モーダル・新規飼主・電話 | 不正番号 → エラー → 正しい番号に修正 | 修正後にエラー表示が消える |

### 権限・マスタ（V03 / V04）

| BUG | 画面 | 操作の要点 | 合格条件 |
|-----|------|------------|----------|
| **023** | 権限グループ | 既存名と重複で保存 | 生の backend 文ではなく日本語。名前が空表示にならない |
| **025** | `/settings/interview/chief-complaint` 等 | 新規作成（有効のまま） | `is_active=true` で保存（無効固定にならない） |
| **027** | `/settings/animal-species` | 既存名で重複作成 | エラーに **入力した名称** が出る |
| **028** | `/settings/treatment-items?tab=procedure` | 処置タブで必須 UI のみ入力して新規 | 登録成功。内部フィールド名がエラーに出ない |

### LINE 予約設定（V05）

| BUG | 画面 | 操作の要点 | 合格条件 |
|-----|------|------------|----------|
| **030** | `/line-reservation/settings` | 最短予約受付日数を **0** にして保存 | 200 後も値が 0 のまま（再読込で戻らない） |

---

## 6. シナリオ原文インデックス

| ID | ファイル | 主な BUG |
|----|----------|----------|
| S01 | [`S01-deceased-pet-guard.md`](../docs/ops/testing/scenarios/S01-deceased-pet-guard.md) | 001, 002 |
| S02 | [`S02-exam-abnormal-highlight-lock.md`](../docs/ops/testing/scenarios/S02-exam-abnormal-highlight-lock.md) | 003, 004, 005 |
| S03 | [`S03-vaccination-next-due-autocalc.md`](../docs/ops/testing/scenarios/S03-vaccination-next-due-autocalc.md) | 006, 007 |
| S04 | [`S04-liff-reservation-journey.md`](../docs/ops/testing/scenarios/S04-liff-reservation-journey.md) | 008 |
| S05 | [`S05-hospitalization-cycle.md`](../docs/ops/testing/scenarios/S05-hospitalization-cycle.md) | 009 (WAIVED) |
| S06 | [`S06-record-lock-audit-trail.md`](../docs/ops/testing/scenarios/S06-record-lock-audit-trail.md) | 010 |
| S07 | [`S07-estimate-status-control.md`](../docs/ops/testing/scenarios/S07-estimate-status-control.md) | 011 |
| S10 | [`S10-customer-aggregation-consistency.md`](../docs/ops/testing/scenarios/S10-customer-aggregation-consistency.md) | 012 |
| S11 | [`S11-trimming-combined-accounting.md`](../docs/ops/testing/scenarios/S11-trimming-combined-accounting.md) | 013 |
| S12 | [`S12-liff-pet-health.md`](../docs/ops/testing/scenarios/S12-liff-pet-health.md) | 014 |
| V01 | [`V01-clinical-forms.md`](../docs/ops/testing/scenarios/V01-clinical-forms.md) | 015–017 |
| V02 | [`V02-accounting-reservation-forms.md`](../docs/ops/testing/scenarios/V02-accounting-reservation-forms.md) | 018–020 |
| V03 | [`V03-owner-pet-staff-forms.md`](../docs/ops/testing/scenarios/V03-owner-pet-staff-forms.md) | 021–024 |
| V04 | [`V04-settings-master-forms.md`](../docs/ops/testing/scenarios/V04-settings-master-forms.md) | 025–029 |
| V05 | [`V05-auth-line-forms.md`](../docs/ops/testing/scenarios/V05-auth-line-forms.md) | 030–032 |
| 全体ガイド | [`SECTION_14_MANUAL_TEST_GUIDE.md`](../docs/ops/testing/SECTION_14_MANUAL_TEST_GUIDE.md) | ドメイン横断 |

---

## 7. 実施後の記録手順

1. [`BROWSER_VERIFICATION_BACKLOG.md`](./BROWSER_VERIFICATION_BACKLOG.md)  
   - セクション A/B の **結果** 列を PASS/FAIL/BLOCKED/WAIVED に更新  
   - **メモ** に日時・ブロッカー・スクショ置き場（repo 外 path 可）  
2. PASS の BUG のみ、任意で `STATUS.md` §3 個票の `原文シナリオ再検証` を PASS に更新  
3. ブラウザ PASS かつ問題なし → 人が `VERIFIED_FIXED` を検討  
4. FAIL → 個票を OPEN に戻すか新規 BUG。再現手順をメモに残す  
5. 件数サマリは個票更新後に機械抽出:

```bash
rg -n '^- \*\*対応状況' STATUS.md \
  | rg -o 'IMPLEMENTED_UNVERIFIED|OPEN|BLOCKED|VERIFIED_FIXED|WAIVED' \
  | sort | uniq -c
```

---

## 8. チェックリスト（コピー用）

### 波 A

- [ ] 事前チェック（health · ranges · LIFF mock）
- [ ] BUG-003 H/L
- [ ] BUG-006 予防接種ヘッダー
- [ ] BUG-007 接種履歴
- [ ] BUG-012 顧客集計
- [ ] BUG-024 権限マトリクス
- [ ] BUG-026 保険 101
- [ ] BUG-029 支払方法重複
- [ ] BUG-031 /login リダイレクト
- [ ] BUG-032 checkup-sync プレビュー
- [ ] BUG-008 LIFF 予約コース
- [ ] BUG-014 LIFF ヘルス
- [ ] バックログ A 表を更新

### 波 B

- [ ] 001 002 021 022（飼主）
- [ ] 005 004 017（検査）
- [ ] 009 WAIVED 確認のみ
- [ ] 010 015（カルテ）
- [ ] 011 019 013 018（見積・会計）
- [ ] 016（存在しない ID）
- [ ] 020（電話バリデーション）
- [ ] 023 025 027 028（権限・マスタ）
- [ ] 030（LINE 最短日数 0）
- [ ] バックログ B 表を更新
- [ ] （任意）STATUS 個票・VERIFIED_FIXED

---

## 9. 関連

| 文書 | 役割 |
|------|------|
| 本ファイル | **手順の単一入口** |
| [`BROWSER_VERIFICATION_BACKLOG.md`](./BROWSER_VERIFICATION_BACKLOG.md) | 結果表 |
| [`../STATUS.md`](../STATUS.md) §3 / §4 | 実装状態 · ブラウザ節 |
| [`2026-08-06-residual-user-ops-pack.md`](./2026-08-06-residual-user-ops-pack.md) | seed/migrate/E2E 等の ops |
| `docs/ops/testing/scenarios/*` | 原文シナリオ詳細 |
