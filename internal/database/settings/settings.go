package settings

import (
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"

	"git.omicron.one/omicron/linkshare/internal/database"
)

var (
	ErrSettingsKeyMissing       = errors.New("Requested setting key is missing")
	ErrSettingsIncompatibleType = errors.New("Stored setting type incompatible with requested type")
)

func Get[T any](db *database.DB, key string) (T, error) {
	var zero T
	var value, kind string

	err := db.Transaction(func(tx *sql.Tx) error {
		row := tx.QueryRow("SELECT value, kind FROM settings WHERE key = ?", key)
		return row.Scan(&value, &kind)
	})

	if err == sql.ErrNoRows {
		return zero, ErrSettingsKeyMissing
	}
	if err != nil {
		return zero, err
	}

	if !isCompatibleType(reflect.TypeOf(zero), kind) {
		return zero, ErrSettingsIncompatibleType
	}

	return parseValue[T](value)
}

func GetOr[T any](db *database.DB, key string, defaultValue T) (T, error) {
	value, err := Get[T](db, key)
	if err == ErrSettingsKeyMissing {
		return defaultValue, nil
	}
	return value, err
}

func isCompatibleType(t reflect.Type, kind string) bool {
	switch t.Kind() {
	case reflect.String:
		return kind == "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return kind == "int"
	case reflect.Bool:
		return kind == "bool"
	default:
		return kind == "json"
	}
}

func parseValue[T any](value string) (output T, err error) {
	switch any(output).(type) {
	case string:
		return any(value).(T), nil
	case int:
		result, err := strconv.ParseInt(value, 10, strconv.IntSize)
		return any(int(result)).(T), err
	case int8:
		result, err := strconv.ParseInt(value, 10, 8)
		return any(int8(result)).(T), err
	case int16:
		result, err := strconv.ParseInt(value, 10, 16)
		return any(int16(result)).(T), err
	case int32:
		result, err := strconv.ParseInt(value, 10, 32)
		return any(int32(result)).(T), err
	case int64:
		result, err := strconv.ParseInt(value, 10, 64)
		return any(result).(T), err
	case uint:
		result, err := strconv.ParseUint(value, 10, strconv.IntSize)
		return any(uint(result)).(T), err
	case uint8:
		result, err := strconv.ParseUint(value, 10, 8)
		return any(uint8(result)).(T), err
	case uint16:
		result, err := strconv.ParseUint(value, 10, 16)
		return any(uint16(result)).(T), err
	case uint32:
		result, err := strconv.ParseUint(value, 10, 32)
		return any(uint32(result)).(T), err
	case uint64:
		result, err := strconv.ParseUint(value, 10, 64)
		return any(result).(T), err
	case bool:
		result, err := strconv.ParseBool(value)
		return any(result).(T), err
	default:
		err = json.Unmarshal([]byte(value), &output)
		return output, err
	}
}
