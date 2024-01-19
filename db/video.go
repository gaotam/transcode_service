package db

import "context"

func GetSrcVideoById(id string) (src string, err error) {
	err = Connect.QueryRow(context.Background(), "SELECT src FROM videos WHERE id = $1", id).Scan(&src)
	if err != nil {
		return "", err
	}
	return src, nil
}
