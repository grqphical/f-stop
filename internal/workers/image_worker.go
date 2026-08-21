package workers

import (
	"os"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/grqphical/f-stop/internal/database"
	"github.com/grqphical/f-stop/internal/models"
	"github.com/grqphical/f-stop/internal/storage"
	"github.com/rwcarlsen/goexif/exif"
)

const (
	thumbnailWidth  int = 800
	thumbnailheight int = 600

	thumbnailQuality int = 85
)

func ImageProcessingWorker(payload models.JobPayload, db database.DBInterface, storageInterface storage.StorageInterface) error {
	f, err := os.Open(payload.Filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	var latitude, longitude *float64 = nil, nil
	var takenAtTimestamp *time.Time = nil
	var cameraModel *string = nil

	x, err := exif.Decode(f)
	if err == nil {
		lat, lng, err := x.LatLong()
		if err == nil {
			latitude = &lat
			longitude = &lng
		}

		takenAt, err := x.DateTime()
		if err == nil {
			takenAtTimestamp = &takenAt
		}

		cam, err := x.Get(exif.Model)
		if err == nil {
			camStr, _ := cam.StringVal()
			cameraModel = &camStr
		}
	}

	err = db.SetPhotoEXIFData(payload.PhotoID, latitude, longitude, takenAtTimestamp, cameraModel)
	if err != nil {
		return err
	}

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
