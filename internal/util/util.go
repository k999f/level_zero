package util

import (
	"encoding/json"
	"fmt"
	"log"
)

func PrintOrder(data interface{}) {
    var o []byte
    o, err := json.MarshalIndent(data, "", "\t")
    if err != nil {
        log.Print("Printing order error: ", err)
        return
    }
    fmt.Printf("%s \n", o)
}