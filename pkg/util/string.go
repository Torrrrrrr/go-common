package util

import (
	"database/sql"
	"strconv"
	"strings"
)

func IsNotEmptyOrNull(str string) bool {
	return !(strings.TrimSpace(str) == "" || strings.EqualFold(str, "null"))
}

func IsEmptyOrNull(str string) bool {
	return strings.TrimSpace(str) == "" || strings.EqualFold(str, "null")
}

func NullStringToString(str *sql.NullString) string {
	if str != nil && str.Valid {
		return str.String
	}
	return ""
}

func StringToNullString(str string) *sql.NullString {
	return &sql.NullString{
		String: str,
		Valid:  true,
	}
}

func Itoa[number int | int64 | float64](num number) string {
	return strconv.FormatInt(int64(num), 10)
}

func Atoi(str string) int64 {
	if num, err := strconv.ParseInt(str, 10, 64); err != nil {
		return 0
	} else {
		return num
	}
}
