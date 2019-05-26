// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package commands

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/mattermost/mattermost-server/mlog"
)

// prettyPrintStruct will return a prettyPrint version of a given struct
func prettyPrintStruct(t interface{}) string {
	return prettyPrintMap(structToMap(t))
}

// structToMap converts a struct into a map
func structToMap(t interface{}) map[string]interface{} {
	defer func() {
		if r := recover(); r != nil {
			mlog.Error(fmt.Sprintf("Panicked in structToMap. This should never happen. %v", r))
		}
	}()

	val := reflect.ValueOf(t)

	if val.Kind() != reflect.Struct {
		return nil
	}

	out := map[string]interface{}{}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		var value interface{}

		switch field.Kind() {
		case reflect.Struct:
			value = structToMap(field.Interface())
		case reflect.Ptr:
			indirectType := field.Elem()

			if indirectType.Kind() == reflect.Struct {
				value = structToMap(indirectType.Interface())
			} else {
				value = indirectType.Interface()
			}
		default:
			value = field.Interface()
		}

		out[val.Type().Field(i).Name] = value
	}

	return out
}

// prettyPrintMap will return a prettyPrint version of a given map
func prettyPrintMap(configMap map[string]interface{}) string {
	value := reflect.ValueOf(configMap)
	return printMap(value, 0)
}

// printMap takes a reflect.Value and prints it out, recursively if it's a map with the given tab settings
func printMap(value reflect.Value, tabVal int) string {
	out := &bytes.Buffer{}

	for _, key := range value.MapKeys() {
		val := value.MapIndex(key)
		if newVal, ok := val.Interface().(map[string]interface{}); !ok {
			fmt.Fprintf(out, "%s", strings.Repeat("\t", tabVal))
			fmt.Fprintf(out, "%v: \"%v\"\n", key.Interface(), val.Interface())
		} else {
			fmt.Fprintf(out, "%s", strings.Repeat("\t", tabVal))
			fmt.Fprintf(out, "%v:\n", key.Interface())
			// going one level in, increase the tab
			tabVal++
			fmt.Fprintf(out, "%s", printMap(reflect.ValueOf(newVal), tabVal))
			// coming back one level, decrease the tab
			tabVal--
		}
	}

	return out.String()
}
