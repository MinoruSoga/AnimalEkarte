// Package commonbasic は owner / pet の共通基本情報（EMR-491）を読む read-only の reader。
//
// 旧 owners / pets の基本情報列（name, phone, status など）は 001_init.sql で廃止された。基本情報の正本は
// common_{owner,pet}_clinic_bindings → common_{owner,pet}_subjects → common_{owner,pet}_basic_revisions の
// 現行 revision（subject.current_basic_revision_id）である。本パッケージはその現行 revision を
// model.Owner.Basic / model.Pet.Basic（gorm:"-" の一時フィールド）へ載せる。
//
// 書き込みは行わない。owner revision の書き込みは internal/owner、pet revision の書き込みは internal/pet
// だけが持つ（ADR-006 の単一 write owner）。本パッケージはドメイン間 import の allowlist 対象外の
// 中立パッケージとして置き、identitylink / trimming / billing などが共通に参照できる。
//
// fail-closed 規則（R-1）: 要求された owner / pet のうち 1 件でも、binding・subject・現行 revision・
// 「revision が同じ subject に属すること」のいずれかを確認できなければ、呼び出し全体を
// apperrors.CodeCommonBasicUnresolved の 422 で失敗させる。空値の補完、廃止列へのフォールバック、
// 該当行の黙った除外はしない。失敗時は引数の owner / pet を一切更新しない。
//
// 解決は (clinic_id, ローカル ID) の組で行い、ローカル ID 単独では引かない。soft-delete 済みの owner / pet も
// binding が残っていれば解決する（履歴表示で名前を失わないため）。
//
// 変換関数の規則（R-NIL）: 呼び出し側を所有するパッケージの変換関数は OwnerOf / PetOf でエラーを返す。
// 戻り値にエラーを持てない変換関数だけが MustOwnerBasic / MustPetBasic を使う。これらは未 hydrate で
// panic し、Gin の recovery が 500 にする（空値を返して成功を装わない）。使う側の取得経路は必ず
// Hydrate* を通すこと。
//
// 読み取りの一貫性: subject の version と現行ポインタは 1 文で読み、revision は不変行を ID で読む。
// ロックは取らない。writer と直列化が必要な判定（生死など）は internal/pet の LifeStateReader を使う。
package commonbasic
