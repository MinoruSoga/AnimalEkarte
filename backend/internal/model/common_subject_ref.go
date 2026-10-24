package model

import "github.com/google/uuid"

// CommonSubjectRef は owner / pet の共通 subject（common_owner_subjects / common_pet_subjects）の読み取り時点の写し。
// Owner.BasicSubject / Pet.BasicSubject に commonbasic が設定し、更新側が CAS の期待値として使う。
// DB 行のモデルではないため gorm タグを持たない（TableName を持たない構造体に gorm タグを付けると lint が落ちる）。
type CommonSubjectRef struct {
	CommonID          uuid.UUID
	Version           int64
	CurrentRevisionID uint64
}
