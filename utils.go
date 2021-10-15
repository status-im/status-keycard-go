package statuskeycardgo

import "C"

import "encoding/json"

func retValue(pairs ...interface{}) *C.char {
	obj := make(map[string]interface{})
	for i := 0; i < len(pairs)/2; i++ {
		key := pairs[i*2]
		value := pairs[(i*2)+1]
		obj[key.(string)] = value
	}

	b, err := json.Marshal(obj)
	if err != nil {
		return C.CString(err.Error())
	}

	return C.CString(string(b))
}
