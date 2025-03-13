package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

var transport = http.Transport{}

// const EXAMPLEURL = "http://google.com"
const EXAMPLEURL = "https://en.wikipedia.com"

//const EXAMPLEURL = "https://www.thetimenow.com/"

func t1() {
	u, _ := url.Parse(EXAMPLEURL)
	req := http.Request{
		URL: u,
	}
	client := http.Client{Transport: &transport}
	resp, err := client.Do(&req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		return
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}

func t2() {
	resp, err := http.Get(EXAMPLEURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		return
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}

func main() {
	fmt.Println(EXAMPLEURL)
	t1()
	fmt.Println()
	fmt.Println(EXAMPLEURL)
	t2()
}
