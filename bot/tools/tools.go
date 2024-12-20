package tools

import (
	"errors"
	"strconv"
	"strings"
)

// GetChannelID For some reason all ID's of channels and supergroups have "-100" as prefix, which need to be removed
func GetChannelID(id int64) (int64, error) {
	idStr := strconv.FormatInt(id, 10)
	if len(idStr) <= 4 {
		return 0, errors.New("incorrect channel id")
	}
	idStr, _ = strings.CutPrefix(idStr, "-100")
	newID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return newID, nil
}

func GetChannelUsername(username string) string {
	if username == "" {
		return ""
	}
	if username[0] == '@' {
		return username[1:]
	} else if strings.HasPrefix(username, "t.me/") {
		return username[5:]
	} else if strings.HasPrefix(username, "https://t.me/") {
		return username[13:]
	}
	return username
}

func GetIDs(start int, cnt int) *[]int {
	ids := make([]int, cnt)
	for i := 0; i < cnt; i++ {
		ids[i] = start + i
	}
	return &ids
}
