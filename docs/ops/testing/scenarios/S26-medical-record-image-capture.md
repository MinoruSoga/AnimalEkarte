# S26: カルテ画像 — 撮影・アップロード入力と同一ファイル再選択

> **目的**: カルテの画像タブで、端末カメラ撮影とファイルアップロードの 2 系統が正しく動き、許可形式・複数選択・同一ファイルの再選択・アップロード中の二重送信防止が仕様どおりであることを納品前に証明する。
> **所要目安**: 15分 / **深度**: 薄い
> **仕様正本**: [screens/06-medical-records-form.md §2.4](../../../spec/screens/06-medical-records-form.md)。検証キュー: `SLACK-CAMERA`（[todo-verification.md](../../../../todo-verification.md)）。実装参照: `frontend/src/features/medical-records/components/ImageGalleryFilter.tsx`。

## 前提条件

- ローカルの使い捨て clinic、または承認済みの対象 build。カメラ付き端末（スマホ/タブレット想定）があれば実機で、なければ対応ブラウザで確認範囲を分ける。
- 画像権限（medical-records 編集・画像アップロード可）の attached account と生存ペットの fixture。
- テスト用ファイル: jpeg・png・gif・pdf・複数ファイル・非対応形式（例: txt）。実患者画像は使わない。
- 試験後にアップロードした画像を削除する。
- 依存シナリオ: なし。ギャラリー表示・削除の網羅は V01 を参照する。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 画像タブで「撮影」ボタンを押す | カメラ起動または環境カメラ（`capture="environment"`）で画像を取得できる。対応端末では背面カメラが優先される |
| 2 | 撮影入力で受け付ける形式を確認する | 撮影は画像のみ（jpeg/png/gif）。pdf は撮影対象外で、画像だけが取れる |
| 3 | 「アップロード」で jpeg/png/gif/pdf を選択する | 4 形式ともアップロードできる。対応外形式（txt 等）は選択または保存で拒否される |
| 4 | 複数ファイルを一度に選択する | 通常アップロードは複数選択可で、複数件がギャラリーへ追加される |
| 5 | 同じファイルを続けて選択する | input の value がリセットされ、**同じファイルを連続で選んでも再度選択イベントが発火する**（同一ファイル再選択の対応） |
| 6 | アップロード中に再度アップロード/撮影を試みる | ボタンが `disabled={isUploading}` で、二重送信・多重アップロードにならない |
| 7 | 件数・サイズ上限（`MAX_UPLOAD_FILES`・`MAX_FILE_SIZE_MB`/ファイル・合計 50MB/バッチ）を超える選択をする | バッチ全体が fail-closed で拒否され（SEC-CS-F08）、エラー toast が出て部分アップロードされない |
| 8 | アップロード後にギャラリーへ反映される | 追加した画像/資料が一覧に出る。再読込で保持される |

## 確認観点

- `ImageGalleryFilter` は撮影用 input（`capture="environment"`・accept jpeg/png/gif）とアップロード用 input（accept jpeg/png/gif/pdf・`multiple`）を分ける。撮影で pdf が取れる設計ではない。
- `handleFileChange` は選択後に `e.target.value = ""` でリセットし、「同じファイルを選び直せない」既知のブラウザ挙動を回避している。同一ファイル再選択は明示的に確認する（手順 5）。
- アップロード中は `disabled={isUploading}` で両ボタンを無効化。`canUpload` はセクション自体の表示可否で、権限なしではボタン自体が出ない。二重送信は重複画像・部分作成の原因になるため FAIL。
- 件数・1 件サイズ・合計バイトの上限超過はバッチ全体を fail-closed で拒否し toast を出す（SEC-CS-F08）。部分アップロードなし。
- 対応外形式は選択段階または保存段階で拒否されること。サーバー側で拒否される場合はエラー表示を確認する。
- 実機カメラがない環境では撮影を BLOCKED として分け、対象端末の receipt を別途取る。「アップロードが動いた」だけで撮影済みとみなさない。

## 実装突合

- 変更サマリ:
  - `ImageGalleryFilter` の撮影 input（`capture="environment"`・画像のみ）とアップロード input（jpeg/png/gif/pdf・`multiple`）の分離、`e.target.value = ""` リセット、`disabled={isUploading}`、上限 fail-closed（SEC-CS-F08）を現行コードと突合
  - 同一ファイル再選択・アップロード中二重防止・上限拒否を手順 5–7 に対応づけ
