package main

import (
	"fmt"
	"github.com/jonas747/cleverbot"
)

func main() {
	session := cleverbot.New()

	resp, err := session.Ask("How are you?")
	if err != nil {
		panic(err)
	}

	fmt.Println("ME: How are you?")
	fmt.Println(resp)
	resp, _ = session.Ask("Really, thats terrible")
	fmt.Println("Me: Really, thats terrrible")
	fmt.Println(resp)
}
