# U05 — C10 multipart アップロード検証 判定レコード

Campaign: emr-332-security-defense / Unit: U05-FILES-SSRF-SUPPLY / Attempt: att-u05-1
Format: `docs/work/emr-332-security/u01-case-verdict-template.md` (12 fields)

## 判定レコード (12 fields)

| # | Field | 記録内容 |
|---|-------|----------|
| 1 | ケース ID | `C10` — multipart upload 検証 (size / 拡張子 / MIME / content sniff / サーバ生成名 / 保存先 / 拒否副作用) |
| 2 | 対象リビジョン | worktree `feat/emr-332-security-defense-w05` HEAD `fc6acc6b8d267eb51050a4848069e1a1b569e7df` (静的レビュー対象) |
| 3 | 実施方式 | `static` + `simulated` — 静的レビューに加え、メモリ専用テスト `backend/internal/imageupload/emr332_u05/sniff_contract_test.go` を isolation Docker (`--network none`, backend 読み取り専用 mount) で実行。fixture は不活性マジックバイトのみ (PNG/JPEG 先頭バイト + HTML テキスト)。実ファイル・実 payload・oversized/executable/bomb/polyglot は一切不使用 |
| 4 | 正常条件 | 期待: 許可型 (JPEG/PNG + family 別 GIF/WebP/PDF) で size 上限内の upload が受諾・サーバ生成名で保存。観測: simulated で `VerifySniffedContent` が PNG fixture を受諾、`ValidateFileHeader` が適合 header を受諾、`NewStoredName` が `<unixnano>_<hex16><ext>` を生成 (ext は検証済み MIME 由来、traversal 文字なし) |
| 5 | 拒否条件 | 期待: 上限超過・非許可拡張子・ext↔MIME 不一致・sniff 不一致を拒否し保存呼出 0 回。観測 (simulated): HTML fixture を `image/png` 宣言で `VerifySniffedContent` に通すと拒否; oversize (11 MiB)・`.html`/`.exe` 拡張子・ext↔MIME mismatch を `ValidateFileHeader` で拒否 (全 5 テスト PASS)。観測 (static): shared-file 系のみ ext allowlist に依存し content sniff 不在 — scoped FAIL |
| 6 | 防御判定 | `FAIL` (scoped) — shared-file upload (family 4) が拡張子 allowlist のみで content sniffing を欠く (shared_file_request.go:28-41, handler :50-83)。残り 5 family は PASS。per-family 詳細は Section A 行列 |
| 7 | 副作用呼出回数 | 拒否分岐での `uploader.Upload` / `fileStorage.Upload` / DB insert 呼出 = `0` (静的経路確認: 全 reject が upload 呼出前に return)。例外: LSTEP CSV は構文 preflight 失敗時に `failed` ステータスの import 監査行を best-effort で 1 行作成 (lstep_csv_import_service.go:98) — ファイル/オブジェクト永続化は 0。これは機能的な監査記録でありペイロードを保存しない |
| 8 | 検出判定 | `UNKNOWN` — 拒否時に専用 audit row/marker を出す経路は存在しない (CSV の `failed` import 行のみ検出性あり)。成功時の監査 (audit log) は別経路。staging 観測は BLOCKED |
| 9 | 安全な証跡参照 | 本ファイル Section A 行列 + `backend/internal/imageupload/emr332_u05/sniff_contract_test.go` (5 テスト PASS, exit 0) |
| 10 | 制約・未検証範囲 | 実 multipart HTTP リクエスト・実 S3/local 保存・実 DB は未実行 (禁止条件)。fixture はメモリ内マジックバイトのみで持続化なし。`FileHeader.Size` は Go multipart parser が計上する実バイト値 (client 宣言値ではない)。LIFF family は未実装のため runtime 検証不能 |
| 11 | 修復案 | RC-1 (Section B): shared-file へ `VerifySniffedContent` 相当を追加する候補 — 本 unit では修復しない (audit_review / 非 owned path) |
| 12 | 再検証結果 | 未実施 — 修復後に同ケース ID (C10) で scoped 検証が必要 |

## Section A — 6 family 検証行列

| # | Family | 入口 | size cap | ext allowlist | MIME check | content sniff | server-generated name | destination control | 拒否時 save/call 数 | verdict |
|---|--------|------|----------|---------------|------------|---------------|----------------------|---------------------|---------------------|---------|
| 1 | カルテ画像 | `medicalrecord/medical_record_image_handler.go:146-244` | `ContentLength` 上限 11 MiB (:163) + `MaxBytesReader` (:186) + header.Size ≤10 MiB (constraints) + quota Acquire BEFORE parse (:168-178) | YES `.jpg/.jpeg/.png/.gif/.pdf` | YES 宣言 MIME ∈ allowlist + ext↔MIME 相互照合 (`ValidateFileHeader` imageupload.go:64) | YES `uploadMeta.verifySniffedContent` → `imageupload.VerifySniffedContent` (:205) | YES `newStoredName` → `NewStoredName` (:209) `<unixnano>_<rand16><ext>` ext=MIME 由来 | YES `uploadKey` → `medical-records/<recordID>/<name>` (request.go:111-113) → FileUploader | **0** — 全 reject が `h.uploader.Upload` (:217) 前に return; 後段失敗時は `cleanupUpload` で Delete (:221-225) | PASS |
| 2 | トリミング画像 | `trimming/trimming_image_handler.go:86-183` | `ContentLength` 上限 (:96) + `MaxBytesReader` (:136) + quota Acquire (:111-117) | YES `.jpg/.jpeg/.png/.gif` | YES `ValidateFileHeader` (:149) | YES `imageupload.VerifySniffedContent` (:154) | YES `NewStoredName` (:158) | YES `StoredKey` → `trimming-images/<clinicID>/<name>` (:165 + response.go:12) | **0** — reject は `imageUploader.Upload` 前に return; post-commit cleanup で superseded 削除 (:176, cleanup.go) | PASS |
| 3 | サポートスクショ | `support/handler.go:95-161` + `request.go:86-116` | `ContentLength` 8 MiB (:95) + `MaxBytesReader` (:99) + `fileHeader.Size` ≤8 MiB (:272) + `ParseMultipartForm` メモリ上限 (request.go:21) | n/a — 拡張子は sniff 結果から導出 (filename 非参照) | 宣言 MIME は検証対象外 (sniff のみが判定) | YES `validateScreenshot` = `http.DetectContentType` 512B → allowlist {png,jpeg,webp} (request.go:89-103) | YES `uploadKey` → `support-bug-reports/clinic-<id>/<ts>-<uuid><ext>` ext=sniff 由来 (request.go:108-116) | YES FileUploader | **0** — reject は `h.uploader.Upload` (:286) 前に return | PASS |
| 4 | 共有ファイル | `lstep/shared_file_handler.go:39-83` + `shared_file_request.go:28-41` | `MaxBytesReader` 10 MiB (:50, `MaxSharedFileSizeBytes`) | YES `model.AllowedFileExtensions` {.pdf,.jpg,.jpeg,.png} (shared_file.go:45-50) | NO — 宣言 Content-Type を参照せず ext→MIME map のみ (:30) | **NO — content sniff 不在** | 一部 — `generateSharedFileKey` = `shared/<clinicID>/<rand16><ext>` (service:237-243)。rand16 はサーバ生成だが ext 末尾は client filename 由来 (allowlist 内) | YES FileStorage (S3 shared bucket / local) — key はサーバ構成 | **0** — reject は `sharedFile.Upload` (:76) 前に return | **FAIL (scoped)** — sniff 不在。LINE 配信経路で外部送信に至る (shared_file.go:42-44) |
| 5 | LSTEP CSV | `lstep/lstep_csv_import_handler.go:66-97` + `lstep_csv_import_service.go` | `ContentLength` + `MaxBytesReader` 51 MiB (:71,:75, service.go:20-21) + row/col/cell 上限 (100k 行 / 64 列 / 64 KiB セル service.go:22-24) | NO ext 検査 (filename は記録のみ) | NO | NO sniff — ただし `csv.Reader` で構文 preflight + 行/列/セル上限。ファイルとして永続化しない | n/a — ファイルを保存しない (DB 行 import のみ) | temp 不在 — ストリーム parse → DB 行のみ | header preflight 失敗時 `failed` 監査行 1 (best-effort, service.go:95-98)。ファイル/オブジェクト永続化 0 | PASS (構造的に sniff 不適合 — CSV は grammar 検証が相当制御) |
| 6 | LIFF 予約画像 stub | `reservation/reservation_type_liff_handler.go:144-147` | n/a | n/a | n/a | n/a — `WrapNotImplemented` で無条件 501。`FormFile` 非呼出 | n/a | n/a | **0** — body 処理前に return | N/A — 未実装 stub (v2 scope) |

## Section B — remediation candidate

| ID | 対象 | 候補 (minimal) |
|----|------|----------------|
| RC-1 | `lstep/shared_file_request.go:28-41` + handler :52-83 | `newSharedFileUploadMeta` の後に `imageupload.VerifySniffedContent` 相当を追加し、`ext` から導いた MIME (`AllowedFileExtensions[ext]`) と先頭 512B sniff を照合する候補。`http.DetectContentType` は `%PDF` を `application/pdf` として検出するため PDF 含む全 allowlist 項目で照合可能。`lstep/` は本 unit の owned_paths 外のため修復は別 unit。非 sniff で通過できるのは「拡張子だけ一致・内容が別物」のファイルであり、LINE Messaging 経由の外部送信経路である点 (shared_file.go:42-44) が影響度 |

## Section C — メモリ専用検証と拒否副作用

- `backend/internal/imageupload/emr332_u05/sniff_contract_test.go` — 5 テスト: PNG 受諾 / HTML-as-PNG 拒否 / rewind 確認 / oversize+非許可 ext+ext↔MIME mismatch+`.exe` 拒否 / `NewStoredName` の MIME 由来 ext と traversal 文字不在。全 PASS (exit 0)。
- 実行環境: `docker run --rm --pull never --network none -v <w05>/backend:/app:ro -v ekarte-go-mod-cache:/go/pkg/mod -v ekarte-go-build-cache:/root/.cache/go-build -w /app golang:1.26-alpine go test -count=1 -v ./internal/imageupload/emr332_u05/`。
- fixture は不活性マジックバイト (PNG/JPEG 先頭 16B) とプレーンテキスト HTML のみ — 実行可能 payload・oversized・decompression bomb・polyglot・持続化は不使用。
- 拒否副作用: 全 family の reject 分岐は upload/save 呼出前に return → 副作用呼出 0。CSV は failed 監査行のみ (payload 非保存)。
- 保存先: `STORAGE_TYPE` 選択 — `s3` で `S3FileStorage`/`S3Uploader` (env 指定 endpoint/bucket、path-style override)、非 `s3` で local `/app/uploads` (cmd/api/main.go:121-183)。release は `STORAGE_TYPE=s3` 必須 (config.go:286)。
