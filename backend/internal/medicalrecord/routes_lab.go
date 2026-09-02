package medicalrecord

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

func (h *Handler) registerLabRoutes(rg *gin.RouterGroup, perm medicalRoutePerm) {
	// BRT-96: device item master (setup only; not on the daily send path).
	// source_type is varchar allowlist, not lab_import_source_type enum (F9).
	labDeviceMasters := rg.Group("/lab-device-item-masters")
	labDeviceMasters.GET("", perm(model.ResourceLabImport, "view"), h.labImport.ListLabDeviceItemMasters)
	labDeviceMasters.POST("/ensure", perm(model.ResourceLabImport, "edit"), h.labImport.EnsureLabDeviceItemMasters)
	labDeviceMasters.PATCH("/:id", perm(model.ResourceLabImport, "edit"), h.labImport.UpdateLabDeviceItemMaster)

	labDevices := rg.Group("/lab-devices")
	labDevices.GET("", perm(model.ResourceLabImport, "view"), h.labImport.ListLabDevices)
	labDevices.POST("", perm(model.ResourceLabImport, "create"), h.labImport.CreateLabDevice)
	labDevices.PATCH("/:id", perm(model.ResourceLabImport, "edit"), h.labImport.UpdateLabDevice)
	labDevices.PUT("/:id/configuration", perm(model.ResourceLabImport, "edit"), h.labImport.SaveLabDeviceConfiguration)

	// BRT-97: receive board / wait / frames. exam persist is BRT-98.
	labDevice := rg.Group("/lab-device")
	labDevice.POST("/frames", perm(model.ResourceLabImport, "create"), h.labImport.ReceiveLabDeviceFrames)
	labDevice.PUT("/wait", perm(model.ResourceLabImport, "create"), h.labImport.PutLabDeviceWait)
	labDevice.DELETE("/wait", perm(model.ResourceLabImport, "create"), h.labImport.DeleteLabDeviceWait)
	labDevice.GET("/board", perm(model.ResourceLabImport, "create"), h.labImport.GetLabDeviceBoard)
	labDevice.GET("/unlinked", perm(model.ResourceLabImport, "view"), h.labImport.GetLabDeviceUnlinked)
	labDevice.GET("/station", perm(model.ResourceLabImport, "view"), h.labImport.GetLabDeviceStation)
	labDevice.PUT("/station", perm(model.ResourceLabImport, "edit"), h.labImport.PutLabDeviceStation)

	// Lab import saga. All routes guard ResourceLabImport (preview/commit=create,
	// job/events reads=view) — P5 parity.
	labImports := rg.Group("/lab-imports")
	labImports.POST("/preview", perm(model.ResourceLabImport, "create"), h.labImport.PreviewLabImport)
	labImports.POST("", perm(model.ResourceLabImport, "create"), h.labImport.CommitLabImport)
	labImports.GET("/:job_id", perm(model.ResourceLabImport, "view"), h.labImport.GetLabImportJob)
	labImports.GET("/:job_id/events", perm(model.ResourceLabImport, "view"), h.labImport.ListLabImportEvents)
	// TASK-032: compensating revert is a dedicated endpoint under lab-import:edit (not examination unconfirm).
	labImports.POST("/:job_id/revert", perm(model.ResourceLabImport, "edit"), h.labImport.RevertLabImport)
	labImports.POST("/:job_id/attach", perm(model.ResourceLabImport, "edit"), h.labImport.AttachLabDeviceJob)
	labImports.POST("/:job_id/detach", perm(model.ResourceLabImport, "edit"), h.labImport.DetachLabDeviceJob)

	// Lab report reads use ResourceLabImport "view", same as lab-import reads.
	labReports := rg.Group("/lab-reports")
	labReports.GET("/jobs/:job_id/summaries", perm(model.ResourceLabImport, "view"), h.labReport.GetLabJobReportSummaries)
	labReports.GET("/exams/:exam_id", perm(model.ResourceLabImport, "view"), h.labReport.GetLabExamReport)
}
