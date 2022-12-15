package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/MeteorsLiu/StreamsbDownload/stream"
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
		ret := strings.SplitN(scanner.Text(), ":", 2)
		link := ret[1]
		path := "/home/nfs/py/" + strings.Split(ret[0], "_")[0] + "/" + ret[0]
		s, err := stream.Parse(link)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("Download to ", path)
		s.Download(path)
	}

}
