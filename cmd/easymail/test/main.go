package main

import (
	"bytes"
	"easymail/vender/dkim"
	"fmt"
	"github.com/jhillyerd/enmime"
	"log"
	"os"
)

func main() {
	data, err := os.ReadFile("//home/bobxiao/tmp/dkim1.eml")
	if err != nil {
		panic(err)
	}

	mailObj, err := enmime.ReadEnvelope(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}

	fmt.Println(mailObj.GetHeader("Subject"))

	verifications, err := dkim.Verify(bytes.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range verifications {
		if v.Err == nil {
			log.Printf("Valid signature for %v", v.Domain)
		} else {
			log.Printf("Invalid signature for %v: %v", v.Domain, v.Err)
		}
	}
}
