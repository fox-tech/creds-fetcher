package env

import (
	"fmt"
	"os"
	"reflect"
)

func Unmarshal(value interface{}) (err error) {
	rval, ok := getTarget(value)
	if !ok {
		return fmt.Errorf("value of %T is not settable", value)
	}

	rtype := rval.Type()
	end := rval.NumField()
	// Iterate through fields
	for i := 0; i < end; i++ {
		// Get field at current index
		field := rval.Field(i)

		// If this was a performance minded usage, we would create a lookup table for this.
		// Because this is a one time use function for configuration initialization, there
		// is no need for a lookup table.
		fieldTag, ok := rtype.Field(i).Tag.Lookup("env")
		if !ok {
			continue
		}

		// We currently only support string, as string is the only field type within the
		// Configuration struct. That being said, if we need to expand this later in the
		// future, we can.
		if field.Kind() != reflect.String {
			err = fmt.Errorf("invalid field type supported, <%s> was provided and only <%s> is supported", field.Kind(), reflect.String)
			return
		}

		// Get value for provided field tag
		envVal := os.Getenv(fieldTag)

		// Set the environment value for the given field
		field.SetString(envVal)
	}

	return
}
