package trimming

import (
	"github.com/animal-ekarte/backend/internal/imageupload"
)

const (
	trimmingImageMaxUploadSize  = 10 * 1024 * 1024 // 10MB
	trimmingImageMaxRequestSize = trimmingImageMaxUploadSize + 1024*1024
)

// allowedTrimmingImageMIMETypes はトリミング画像（仕上がり・スタイル写真）として
// 受理する画像型。PDF は画像フィールドではないため含めない（FE は accept="image/*"）。
var allowedTrimmingImageMIMETypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
}

var trimmingImageConstraints = imageupload.FileConstraints{
	MaxSize:          trimmingImageMaxUploadSize,
	AllowedMIMETypes: allowedTrimmingImageMIMETypes,
	MIMEByExt: map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
	},
	ExtByMIME: map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
	},
	AllowedLabel: "jpeg, png, gif",
}
