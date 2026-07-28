package sharedkernel

import "github.com/animal-ekarte/backend/internal/model"

// AuditActorTypeFor は actorID の有無から監査 actor 種別（staff/system）を導出する。
func AuditActorTypeFor(actorID *uint64) string {
	if actorID != nil {
		return model.AuditActorTypeStaff
	}
	return model.AuditActorTypeSystem
}
