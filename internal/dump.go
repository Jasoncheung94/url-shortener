package common

import (
	"encoding/json"
	"fmt"
)

// DumpStruct dumps data in a readable format.
func DumpStruct(data any) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(jsonData))
}
