package env

import "reflect"

func getTarget(value interface{}) (target reflect.Value, ok bool) {
	target = reflect.ValueOf(value)
	for {
		switch target.Kind() {
		case reflect.Ptr:
			target = target.Elem()
		case reflect.Interface:
			target = target.Elem()

		default:
			ok = target.CanSet()
			return
		}
	}
}
