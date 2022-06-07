package env

import (
	"fmt"
	"os"
	"reflect"
)

func Unmarshal(value interface{}) (err error) {
	rval := reflect.ValueOf(value)
	rtype := rval.Type()
	end := rval.NumField()
	for i := 0; i < end; i++ {
		field := rval.Field(i)

		// If this was a performance minded usage, we would create a lookup table for this.
		// Because this is a one time use function for configuration initialization, there
		// is no need for a lookup table.
		fieldTag, ok := rtype.Field(i).Tag.Lookup("env")
		if !ok {
			continue
		}

		envVal := os.Getenv(fieldTag)
		if field.Kind() != reflect.String {
			err = fmt.Errorf("invalid field type supported, <%s> was provided and only <%s> is supported", field.Kind(), reflect.String)
			return
		}

		field.SetString(envVal)
	}

	return
}
