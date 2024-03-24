package db

import (
	"context"
	"encoding/json"
	"fmt"
)

func GetSrcVideoById(id string) (src string, err error) {
	err = Connect.QueryRow(context.Background(), "SELECT src FROM videos WHERE id = $1", id).Scan(&src)
	if err != nil {
		return "", err
	}
	return src, nil
}

func UpdateVideoById(id string, fileName string, resolution int) (err error) {
	srcTranscode := fmt.Sprintf("/transcodes/%s/%s_master.m3u8", fileName, fileName)
	var resolutions []int

	if resolution >= 360 {
		resolutions = append(resolutions, 360)
	}

	if resolution >= 480 {
		resolutions = append(resolutions, 480)
	}

	if resolution >= 720 {
		resolutions = append(resolutions, 720)
	}

	if resolution >= 1080 {
		resolutions = append(resolutions, 1080)
	}

	metadata := map[string]interface{}{
		"resolutions": resolutions,
	}

	jsonMetadata, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	_, err = Connect.Exec(context.Background(), "UPDATE videos SET \"srcTranscode\" = $1, \"metadata\" = $2 WHERE id = $3", srcTranscode, string(jsonMetadata), id)
	if err != nil {
		return err
	}
	return nil
}
