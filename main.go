package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	var file string
	flag.StringVar(&file, "file", "", "file to be read")
	flag.Parse()
	if file == "" {
		log.Fatal("please input file")
	}

	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ret := strings.Split(scanner.Text(), ":")
		fmt.Println(ret[0], ret[1])
	}

}
