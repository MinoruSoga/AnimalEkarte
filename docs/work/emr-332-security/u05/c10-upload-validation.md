# U05 — C10 multipart アップロード検証 判定レコード

> 配置追記（2026-10-09）: 現行テストは `backend/tests/security/`。旧package名を含む実行コマンド・出力は当時の履歴として保持。移動対応と現行検証は [配置整理記録](../test-layout-20261009.md) を参照。

Campaign: emr-332-security-defense / Unit: U05-FILES-SSRF-SUPPLY / Attempt: att-u05-1
Format: `docs/work/emr-332-security/u01-case-verdict-template.md` (12 fields)

## 判定レコード (12 fields)

| # | Field | 記録内容 |
|---|-------|----------|
| 1 | ケース ID | `C10` — multipart upload 検証 (size / 拡張子 / MIME / content sniff / サーバ生成名 / 保存先 / 拒否副作用) |
| 2 | 対象リビジョン | 初回静的対象はw05記録revision。Codex follow-upはPR head `61bda0f7db8f21470d37121d6e032b1ad6f6b8a4`からのowned diff (コードhash・コマンドは `repair-20261009.md`)。staging稼働revisionは未確定。 |
| 3 | 実施方式 | `static` + `simulated` — 既存helper試験に加え、実UploadSharedFile handlerをhttptestとメモリservice mockで実行。PNG/JPEGは1pixel画像、PDFは不活性signature、拒否は英数字marker/空/形式違い。DB・storage・LINE・実HTTP通信なし。 |
| 4 | 正常条件 | 共有ファイルでPNG/JPG/JPEG/PDF signature・大文字PNG拡張子が201、Upload1回。service mockが読む全bytesがfixtureと一致し、sniff後のrewindを確認。完全なファイル構文・無害性を証明する試験ではない。 |
| 5 | 拒否条件 | 修復前RED: 内容不一致7ケースが201・Upload1回。修復後GREEN: 同ケースは400・Upload0回・Locationなし。正常と既存auth/bind/extension/service-error分岐も維持。 |
| 6 | 防御判定 | **PASS (scoped static/simulated)** — 共有ファイル保存前のMIME signature照合を修復。既存4familyのstatic PASSと共有ファイルのscoped PASS、LIFF stub N/Aを区別する。全ファイル安全性・実保存・配送の証明ではない。 |
| 7 | 副作用呼出回数 | 修復対象の拒否7ケースでmock `sharedFile.Upload`呼出0、正常5ケースで1。実storage/DB未接続。CSV失敗時監査行の記録は旧static所見であり実測ではない。 |
| 8 | 検出判定 | `UNKNOWN` — 拒否時に専用 audit row/marker を出す経路は存在しない (CSV の `failed` import 行のみ検出性あり)。成功時の監査 (audit log) は別経路。staging 観測は BLOCKED |
| 9 | 安全な証跡参照 | `backend/internal/lstep/shared_file_content_test.go`; `shared_file_handler_test.go`; `repair-20261009.md`。既存helper契約試験はSection C。 |
| 10 | 制約・未検証範囲 | multipartはhttptestのメモリ内処理。実S3/local保存・DB・LINE・staging未実行。sniffは先頭512Bの形式照合であり完全な画像/PDF検証、マルウェア・polyglot排除を保証しない。 |
| 11 | 修復案 | RC-1を適用: `imageupload.VerifySniffedContent(file, meta.contentType)`をフォーム・用途検証後、`sharedFile.Upload`前に配線。既存size/auth/permission/MIME由来契約を維持。 |
| 12 | 再検証結果 | Codex follow-upでRED→GREENを確認。コマンド・結果・保証範囲は `repair-20261009.md`。旧campaign G2/G3はこの後続修復のレビュー証跡ではない。 |

## Section A — 6 family 検証行列

| # | Family | 入口 | size cap | ext allowlist | MIME check | content sniff | server-generated name | destination control | 拒否時 save/call 数 | verdict |
|---|--------|------|----------|---------------|------------|---------------|----------------------|---------------------|---------------------|---------|
| 1 | カルテ画像 | `medicalrecord/medical_record_image_handler.go:146-244` | `ContentLength` 上限 11 MiB (:163) + `MaxBytesReader` (:186) + header.Size ≤10 MiB (constraints) + quota Acquire BEFORE parse (:168-178) | YES `.jpg/.jpeg/.png/.gif/.pdf` | YES 宣言 MIME ∈ allowlist + ext↔MIME 相互照合 (`ValidateFileHeader` imageupload.go:64) | YES `uploadMeta.verifySniffedContent` → `imageupload.VerifySniffedContent` (:205) | YES `newStoredName` → `NewStoredName` (:209) `<unixnano>_<rand16><ext>` ext=MIME 由来 | YES `uploadKey` → `medical-records/<recordID>/<name>` (request.go:111-113) → FileUploader | **0** — 全 reject が `h.uploader.Upload` (:217) 前に return; 後段失敗時は `cleanupUpload` で Delete (:221-225) | PASS |
| 2 | トリミング画像 | `trimming/trimming_image_handler.go:86-183` | `ContentLength` 上限 (:96) + `MaxBytesReader` (:136) + quota Acquire (:111-117) | YES `.jpg/.jpeg/.png/.gif` | YES `ValidateFileHeader` (:149) | YES `imageupload.VerifySniffedContent` (:154) | YES `NewStoredName` (:158) | YES `StoredKey` → `trimming-images/<clinicID>/<name>` (:165 + response.go:12) | **0** — reject は `imageUploader.Upload` 前に return; post-commit cleanup で superseded 削除 (:176, cleanup.go) | PASS |
| 3 | サポートスクショ | `support/handler.go:95-161` + `request.go:86-116` | `ContentLength` 8 MiB (:95) + `MaxBytesReader` (:99) + `fileHeader.Size` ≤8 MiB (:272) + `ParseMultipartForm` メモリ上限 (request.go:21) | n/a — 拡張子は sniff 結果から導出 (filename 非参照) | 宣言 MIME は検証対象外 (sniff のみが判定) | YES `validateScreenshot` = `http.DetectContentType` 512B → allowlist {png,jpeg,webp} (request.go:89-103) | YES `uploadKey` → `support-bug-reports/clinic-<id>/<ts>-<uuid><ext>` ext=sniff 由来 (request.go:108-116) | YES FileUploader | **0** — reject は `h.uploader.Upload` (:286) 前に return | PASS |
| 4 | 共有ファイル | `lstep/shared_file_handler.go` + `shared_file_request.go:28-41` | MaxBytesReader 10MiB (multipart全体) | YES {.pdf,.jpg,.jpeg,.png} | ext由来MIMEを使用、宣言headerを信頼しない | YES `imageupload.VerifySniffedContent`をUpload前に配線 | `shared/<clinicID>/<rand16><ext>`、extはallowlist + sniff照合 | 設定されたFileStorage | **0 (実handler・メモリmock)** | **PASS (scoped)** — 形式照合のみ。配送は別操作で署名URLをLINEへ渡す条件付き経路 (`line_send_service.go:139-158`, `shared_file_service.go:132-145`, `infra/line/push.go:36-67,94-120`) |
| 5 | LSTEP CSV | `lstep/lstep_csv_import_handler.go:66-97` + `lstep_csv_import_service.go` | `ContentLength` + `MaxBytesReader` 51 MiB (:71,:75, service.go:20-21) + row/col/cell 上限 (100k 行 / 64 列 / 64 KiB セル service.go:22-24) | NO ext 検査 (filename は記録のみ) | NO | NO sniff — ただし `csv.Reader` で構文 preflight + 行/列/セル上限。ファイルとして永続化しない | n/a — ファイルを保存しない (DB 行 import のみ) | temp 不在 — ストリーム parse → DB 行のみ | header preflight 失敗時 `failed` 監査行 1 (best-effort, service.go:95-98)。ファイル/オブジェクト永続化 0 | PASS (構造的に sniff 不適合 — CSV は grammar 検証が相当制御) |
| 6 | LIFF 予約画像 stub | `reservation/reservation_type_liff_handler.go:144-147` | n/a | n/a | n/a | n/a — `WrapNotImplemented` で無条件 501。`FormFile` 非呼出 | n/a | n/a | **0** — body 処理前に return | N/A — 未実装 stub (v2 scope) |

## Section B — remediation candidate

| ID | 対象 | 候補 (minimal) |
|----|------|----------------|
| RC-1 | `lstep/shared_file_handler.go` | **適用済み** — 拡張子由来MIMEと先頭512Bを既存helperで照合し、rewind後にUploadへ渡す。PDF signature照合は完全なPDF解析ではない。配送経路は別操作・資格情報/owner/opt-out条件付きであり、アップロードだけで外部送信されない。 |

## Section C — メモリ専用検証と拒否副作用

- `backend/tests/security/uploads/sniff_contract_test.go` — 5 テスト: PNG 受諾 / HTML-as-PNG 拒否 / rewind 確認 / oversize+非許可 ext+ext↔MIME mismatch+`.exe` 拒否 / `NewStoredName` の MIME 由来 ext と traversal 文字不在。全 PASS (exit 0)。
- 実行環境: `docker run --rm --pull never --network none -v <w05>/backend:/app:ro -v ekarte-go-mod-cache:/go/pkg/mod -v ekarte-go-build-cache:/root/.cache/go-build -w /app golang:1.26-alpine go test -count=1 -v ./internal/imageupload/emr332_u05/`。
- fixture は不活性マジックバイト (PNG/JPEG 先頭 16B) とプレーンテキスト HTML のみ — 実行可能 payload・oversized・decompression bomb・polyglot・持続化は不使用。
- 拒否副作用: 共有ファイルの内容不一致は実handlerのメモリmockでUpload0回。その他familyの旧static所見は実storage/DB副作用の実測ではない。CSV監査行も未実行。
- 保存先: `STORAGE_TYPE` 選択 — `s3` で `S3FileStorage`/`S3Uploader` (env 指定 endpoint/bucket、path-style override)、非 `s3` で local `/app/uploads` (cmd/api/main.go:121-183)。release は `STORAGE_TYPE=s3` 必須 (config.go:286)。

## Codex follow-up — 2026-10-09

旧「残り5family PASS」は4family PASS・LIFF N/Aへ訂正。旧w05 helper試験の記録を残し、追加した実handler試験と区別する。最新結果は [Codex follow-up](../repair-20261009.md)。
