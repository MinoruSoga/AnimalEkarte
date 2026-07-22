package service

import (
	"github.com/animal-ekarte/backend/internal/lstep"
)

// LstepTagSyncService remains only for the owner, pet, chronic-condition, and accounting
// production consumers that have not moved yet. Delete this alias with those consumers in
// BE9-2E; BE9-2F is the final compatibility-surface backstop.
type LstepTagSyncService = lstep.LstepTagSyncService
