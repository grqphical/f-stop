package thumbnail

import (
	"os"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/grqphical/f-stop/internal/database"
	"github.com/grqphical/f-stop/internal/models"
	"github.com/grqphical/f-stop/internal/storage"
)

const (
	thumbnailWidth  int = 800
	thumbnailheight int = 600

	thumbnailQuality int = 85
)

func GenerateThumbnail(payload models.JobPayload, db database.DBInterface, storageInterface storage.StorageInterface) error {
	image, err := vips.NewThumbnailFromFile(payload.Filepath, thumbnailWidth, thumbnailheight, vips.InterestingAttention)
	if err != nil {
		return err
	}

	buf, _, err := image.ExportJpeg(&vips.JpegExportParams{Quality: thumbnailQuality})
	if err != nil {
		return err
	}

	outputPath := storageInterface.StoreThumbnail(payload.PhotoID)

	err = os.WriteFile(outputPath, buf, 0644)
	if err != nil {
		return err
	}

	return db.SetPhotoThumbnailPath(payload.PhotoID, outputPath)
}
