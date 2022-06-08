package env

import "reflect"

func getTarget(value interface{}) (target reflect.Value, ok bool) {
	target = reflect.ValueOf(value)
	for {
		switch {
		case target.Kind() == reflect.Ptr:
			target = reflect.Indirect(target)
		case target.CanSet():
			ok = true
			return
		case target.IsValid() && target.Kind() != reflect.Struct:
			target = target.Elem()

		default:
			return
		}
	}
}
