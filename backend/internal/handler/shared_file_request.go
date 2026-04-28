package handler

// uploadSharedFileRequest はPOST /shared-files のフォームパラメータ
type uploadSharedFileRequest struct {
	Purpose string  `form:"purpose"`
	OwnerID *uint64 `form:"owner_id"`
}
