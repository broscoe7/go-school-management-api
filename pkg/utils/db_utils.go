package utils

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

func GenerateInsertQuery(tableName string, model any) string {
	modelType := reflect.TypeOf(model)
	var columns, placeholders string
	for field := range modelType.Fields() {
		dbTag, _, _ := strings.Cut(field.Tag.Get("db"), ",")
		if dbTag == "" || dbTag == "id" {
			continue
		}
		if columns != "" {
			columns += ", "
			placeholders += ", "
		}
		columns += dbTag
		placeholders += "?"
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, columns, placeholders)
}

func GetStructValues(model any) []any {
	modelValue := reflect.ValueOf(model)
	modelType := modelValue.Type()
	values := []any{}
	for i := 0; i < modelType.NumField(); i++ {
		dbTag, _, _ := strings.Cut(modelType.Field(i).Tag.Get("db"), ",")
		if dbTag != "" && dbTag != "id" {
			values = append(values, modelValue.Field(i).Interface())
		}
	}
	return values
}

func AddSorting(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortby"]
	if len(sortParams) > 0 {
		query += " ORDER BY"
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if !isValidSortOrder(order) || !isValidSortField(field) {
				continue
			}
			if i > 0 {
				query += ","
			}
			query += " " + field + " " + order
		}
	}
	return query
}

func AddFilters(r *http.Request, query string, args []any) (string, []any) {
	params := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"class":      "class",
		"subject":    "subject",
	}
	for param, db_field := range params {
		value := r.URL.Query().Get(param)
		if value != "" {
			query += " AND " + db_field + " = ?"
			args = append(args, value)
		}
	}
	return query, args
}

func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func isValidSortField(field string) bool {
	// use a map instead of a slice/array because it is faster
	validFields := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"class":      true,
		"subject":    true,
	}
	return validFields[field]
}

func GetPaginationParams(r *http.Request) (int, int) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		// set default page value
		page = 1
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		// set default limit
		limit = 10
	}
	return page, limit
}
