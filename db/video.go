package db

import (
	"context"
	"fmt"
)

func GetSrcVideoById(id string) (src string, err error) {
	err = Connect.QueryRow(context.Background(), "SELECT src FROM videos WHERE id = $1", id).Scan(&src)
	if err != nil {
		return "", err
	}
	return src, nil
}

func UpdateSrcTranscode(id string, fileName string) (err error) {
	srcTranscode := fmt.Sprintf("/transcodes/%s/%s_master.m3u8", fileName, fileName)
	_, err = Connect.Exec(context.Background(), "UPDATE videos SET \"srcTranscode\" = $1 WHERE id = $2", srcTranscode, id)
	if err != nil {
		return err
	}
	return nil
}
