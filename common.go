package combustion

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v1"
)

func marshalToYAML(o interface{}) ([]byte, error) {
	y, err := yaml.Marshal(o)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}

	var i map[string]interface{}
	if err := yaml.Unmarshal(y, &i); err != nil {
		fmt.Println("err", err)
		return nil, err
	}

	const cleanIteration = 4
	for it := 0; it < cleanIteration; it++ {
		cleanZeroValues(&i)
	}

	return yaml.Marshal(i)
}

func cleanZeroValues(i interface{}) {
	v := reflect.ValueOf(i)

	switch v.Kind() {
	case reflect.Ptr:
		v = v.Elem()
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			cleanZeroValues(v.Index(i).Interface())
		}
		return
	case reflect.Map:
	default:
		return
	}

	for _, k := range v.MapKeys() {
		value := v.MapIndex(k).Elem()

		var isZero bool
		switch value.Kind() {
		case reflect.Invalid:
			isZero = true
		case reflect.Slice, reflect.Map:
			isZero = value.Len() == 0
		default:
			zero := reflect.Zero(value.Type())
			isZero = reflect.DeepEqual(value.Interface(), zero.Interface())
		}

		if isZero {
			v.SetMapIndex(k, reflect.ValueOf(nil))
		}

		if value.IsValid() {
			cleanZeroValues(value.Interface())
		}
	}
}
