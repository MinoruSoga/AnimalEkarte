package trimming

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"log/slog"
	"os"
	"sort"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// トリミング画像オブジェクトキーのライフサイクルガード（durable 版）。
//
// 背景: post-commit のオブジェクト削除は「参照チェック → オブジェクト削除」の
// 間に writer が同じキーを再格納/再参照する check-to-delete ウィンドウを持つ。
// 行ロックでは writer が別 appointment へ書く経路を塞げないため、キー単位の
// 直列化が必要になる。さらに backend は最大 3 コンテナインスタンス
// （backend/wrangler*.jsonc containers[].max_instances）で同一 Postgres を
// 共有するため、ガード状態はプロセス内ではなく PostgreSQL/オブジェクト
// ストレージ上に durable に存在しなければならない。
//
// 構成は 2 層（durable tombstone = 「オブジェクト自体の不在」）:
//   - pg_advisory_xact_lock(hashtextextended("trimming-image:"+key)) — writer 側:
//     repository Create/Update の claim が ambient tx 内で取得し、
//     ロック下で uploader.Exists を確認してから格納する。
//   - pg_advisory_lock(hashtextextended("trimming-image:"+key)) — cleanup 側:
//     pinned connection 上のセッションスコープロック。「参照再確認 →
//     （tx commit）→ オブジェクト削除 → unlock」をひとつのクリティカル
//     セクションとして保持し、ロック解放後に進んだ writer は
//     Exists=false を見て conflict で拒否される。
//
// プロセス内のガード状態は一切持たない — 削除済みオブジェクトの不在が唯一の
// tombstone であり、別インスタンス・再起動後の writer も同じ判定に到達する。
//
// 参照確認・ロック・存在確認・削除のいずれも失敗した場合は常に保持側
// fail-closed（オブジェクトを消さない）。
//
// ロールアウトゲート（two-phase）: 旧バージョンのインスタンスは advisory
// プロトコルに参加しないため、新旧混在期に削除を実行すると dangling
// reference を生む。オブジェクト削除は
// TRIMMING_IMAGE_OBJECT_DELETE_ENABLED == "1" の場合のみ実行する。
// 手順: (1) 新コードを flag 無効で全インスタンスへデプロイ → (2) 旧
// バージョンのインスタンスを全て排出 → (3) flag を "1" に設定。
// flag の配線（backend/worker/index.ts envVars / backend/wrangler*.jsonc
// vars）は本変更のスコープ外で、有効化まで PR は Draft を維持する。

// trimmingImageKeyLockDomain は advisory lock のハッシュ入力に付ける domain
// prefix で、booking lock ("appointments:<clinic>") 等の他 domain との衝突を
// 避ける。
const trimmingImageKeyLockDomain = "trimming-image:"

// trimmingImageObjectDeleteEnabledEnv は上記ロールアウトゲートの環境変数名。
// "1" 以外（未設定を含む）はオブジェクト削除を行わず warn のみで保持する。
const trimmingImageObjectDeleteEnabledEnv = "TRIMMING_IMAGE_OBJECT_DELETE_ENABLED"

func trimmingImageObjectDeleteEnabled() bool {
	return os.Getenv(trimmingImageObjectDeleteEnabledEnv) == "1"
}

// ImageObjectDeleteFunc は durable オブジェクトストレージから key を削除する
// 関数（FileUploader.Delete と同型）。cleanup 側がセッションロック下で呼ぶ。
type ImageObjectDeleteFunc func(ctx context.Context, key string) error

// imageObjectStoreBinder は writer 側 claim が必要とする durable オブジェクト
// ストアを注入する内部契約。公開 interface（Service /
// AppointmentTrimmingDetailRepository）には載せず、Handler.WithImageSupport
// から構造的アサーションで配線する — composition 層（本変更の対象外パス）の
// シグネチャ変更を不要にするため。
type imageObjectStoreBinder interface {
	bindImageObjectStore(uploader FileUploader)
}

func (s *trimmingService) bindImageObjectStore(uploader FileUploader) {
	if b, ok := s.trimmingDetail.(imageObjectStoreBinder); ok {
		b.bindImageObjectStore(uploader)
	}
}

// storedTrimmingImageObjectKeys は永続化値（style_image / completed_image）のうち
// オブジェクトキーとして解決できるものを重複除去して返す。署名 URL・外部
// HTTPS・`/uploads/` 以外の値はオブジェクト参照にならないため対象外。
// デッドロック順序を固定するためソート済みで返す。
func storedTrimmingImageObjectKeys(storedValues ...string) []string {
	seen := make(map[string]struct{}, len(storedValues))
	keys := make([]string, 0, len(storedValues))
	for _, value := range storedValues {
		key, isObject, err := resolveTrimmingImageStorageKey(value)
		if err != nil || !isObject {
			continue
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// acquireTrimmingImageKeyLock は key スコープの advisory xact lock を取得する。
// ambient transaction 必須 — xact lock はトランザクション終了まで保持され、
// ロック→確認→commit を不可分にする。tx 不在は直列化不能のため fail-closed。
// cleanup 側のセッションスコープ pg_advisory_lock と同じキー空間で衝突する。
func acquireTrimmingImageKeyLock(ctx context.Context, db *gorm.DB, key string) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("trimming image key lock requires an ambient transaction")
	}
	if err := db.WithContext(ctx).Exec(
		"SELECT pg_advisory_xact_lock(hashtextextended(?, 0))",
		trimmingImageKeyLockDomain+key,
	).Error; err != nil {
		return apperrors.Wrap(err, "failed to acquire trimming image key lock")
	}
	return nil
}

// acquireTrimmingImageKeySessionLock は pinned conn 上でセッションスコープの
// advisory lock を取得する。セッションロックは明示的な unlock（または conn
// close＝セッション終了）まで保持されるため、トランザクションの commit を
// またいで「参照再確認→オブジェクト削除→unlock」を不可分にできる。
func acquireTrimmingImageKeySessionLock(ctx context.Context, conn *sql.Conn, key string) error {
	if _, err := conn.ExecContext(ctx,
		"SELECT pg_advisory_lock(hashtextextended($1, 0))",
		trimmingImageKeyLockDomain+key); err != nil {
		discardTrimmingImagePinnedConn(conn)
		return apperrors.Wrap(err, "failed to acquire trimming image key session lock")
	}
	return nil
}

// discardTrimmingImagePinnedConn は pinned connection を pool へ戻さず
// 破棄する。driver.ErrBadConn は database/sql にこの接続を即座に物理
// discard させる（既に破棄済みの conn に対する後続の Close は
// sql.ErrConnDone を返す — セッション終了＝advisory lock 解放の証拠）。
func discardTrimmingImagePinnedConn(conn *sql.Conn) {
	_ = conn.Raw(func(any) error { return driver.ErrBadConn }) //nolint:errcheck // bad マーク自体が目的
}

// releaseTrimmingImageKeySessionLock はセッションロックを解放する。
// unlock が false（この session はロックを保持していなかった）を返した
// 場合や unlock 問い合わせ自体が失敗した場合、セッションのロック状態は
// 不確実 — pool へ返却すると残留ロックが別 cleanup に持ち越され得るため
// 物理接続を破棄する。成功時のみ pool へ返す。
// 呼び出し ctx がキャンセル済みでも unlock を試行できるよう detach するが、
// 無期限には伸ばさず supersededImageCleanupTimeout で bound する。
func releaseTrimmingImageKeySessionLock(ctx context.Context, conn *sql.Conn, key string) {
	releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), supersededImageCleanupTimeout)
	defer cancel()
	var unlocked bool
	err := conn.QueryRowContext(releaseCtx,
		"SELECT pg_advisory_unlock(hashtextextended($1, 0))",
		trimmingImageKeyLockDomain+key).Scan(&unlocked)
	switch {
	case err != nil:
		slog.WarnContext(releaseCtx,
			"failed to release trimming image key session lock; discarding pinned connection",
			slog.String("error", err.Error()),
			slog.String("key", key))
		discardTrimmingImagePinnedConn(conn)
	case !unlocked:
		slog.WarnContext(releaseCtx,
			"trimming image key session lock was not held at release; discarding pinned connection",
			slog.String("key", key))
		discardTrimmingImagePinnedConn(conn)
	}
}

// DeleteUnreferencedImageObject は key が live な trimming 行から参照されて
// いないことを確認し、参照不在ならオブジェクトを削除する。
// pinned connection 上のセッションスコープ advisory lock を
// 「参照再確認（tx）→ オブジェクト削除 → unlock」の全体にわたり保持する —
// writer 側 claim（xact lock + Exists）はこのロックで直列化され、
// チェック〜削除の間に待機した writer は解放後に Exists=false を見て
// conflict で拒否される。これが cross-process な durable tombstone である。
//
// 戻り値: true = オブジェクトを削除した、false = 参照ありまたは削除が
// ロールアウトゲートで無効。参照確認・存在確認・削除の失敗や依存未配線は
// error で返し、呼び出し元はオブジェクトを保持する（fail-closed）。
//
// TRIMMING_IMAGE_OBJECT_DELETE_ENABLED != "1" の間は削除を実行せず
// warn + false を返す — 旧バージョンインスタンスが advisory プロトコルに
// 参加しない混在期の誤削除を防ぐロールアウトゲート。
func (s *trimmingService) DeleteUnreferencedImageObject(
	ctx context.Context,
	clinicID uint64,
	key string,
	deleteObject ImageObjectDeleteFunc,
) (bool, error) {
	if !trimmingImageObjectDeleteEnabled() {
		slog.WarnContext(ctx,
			"skipped trimming image object delete; TRIMMING_IMAGE_OBJECT_DELETE_ENABLED is not \"1\"",
			slog.String("key", key))
		return false, nil
	}
	repo, ok := s.trimmingDetail.(*appointmentTrimmingDetailRepository)
	if !ok || repo.db == nil || deleteObject == nil {
		return false, apperrors.WrapInternalServerError("trimming image lifecycle guard is not configured")
	}
	sqlDB, err := repo.db.DB()
	if err != nil {
		return false, apperrors.Wrap(err, "failed to access trimming image lifecycle database")
	}
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return false, apperrors.Wrap(err, "failed to pin trimming image lifecycle connection")
	}
	defer conn.Close() //nolint:errcheck // 返却失敗はセッション終了＝lock 解放扱いのため無視
	if err := acquireTrimmingImageKeySessionLock(ctx, conn, key); err != nil {
		return false, err
	}
	defer releaseTrimmingImageKeySessionLock(ctx, conn, key)

	// セッションロック保持 conn 上で参照を再確認する。gorm を同一 conn に
	// バインドして ImageKeyInUse の参照解決（正規化・live-parent soft-delete
	// 参照契約・xact lock）をそのまま再利用する。xact lock は同一セッション
	// が保持するセッションロックと同一キー空間で再入可能なため衝突しない。
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: conn}), &gorm.Config{})
	if err != nil {
		return false, apperrors.Wrap(err, "failed to open trimming image lifecycle session")
	}
	unreferenced := false
	if err := gdb.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		inUse, err := repo.ImageKeyInUse(persistence.WithTxValue(ctx, tx), clinicID, key)
		if err != nil {
			return err
		}
		unreferenced = !inUse
		return nil
	}); err != nil {
		return false, apperrors.Wrap(err, "failed to recheck trimming image reference")
	}
	if !unreferenced {
		return false, nil
	}
	if err := deleteObject(ctx, key); err != nil {
		return false, apperrors.Wrap(err, "failed to delete trimming image object")
	}
	return true, nil
}
