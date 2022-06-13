package env

import "reflect"

func getTarget(value interface{}) (target reflect.Value, ok bool) {
	target = reflect.ValueOf(value)
	// Continue to iterate until we get to a valid target value
	for {
		switch target.Kind() {
		case reflect.Ptr:
			// Target is a pointer, set new target as underlying element
			target = target.Elem()
		case reflect.Interface:
			// Target is an interface, set new target as underlying element
			target = target.Elem()

		default:
			// We've reached our element, target is valid if it can be set
			ok = target.CanSet()
			return
		}
	}
}
