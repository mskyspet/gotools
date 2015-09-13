package main

import "os"
import "fmt"
import "crypto/md5"
import "encoding/hex"
import "flag"
import "path/filepath"

func md5_file(file_name string) {
	h := md5.New()

	fin, err := os.Open(file_name)
	defer fin.Close()
	buffer := make([]byte, 1024)
	if err != nil {
		fmt.Println(file_name, err)
		return
	}

	for {
		n, _ := fin.Read(buffer)
		if 0 == n {
			break
		}
		h.Write(buffer[:n])
	}

	fmt.Println(file_name, "\t", hex.EncodeToString(h.Sum(nil)))
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("please give the file name.")
		return
	}

	if flag.NArg() > 1 {
		fmt.Println("too many files.")
		return
	}

	file_pattern := flag.Arg(0)

	match_files, err := filepath.Glob(file_pattern)

	if err != nil {
		fmt.Println(err)
		return
	}

	if match_files == nil {
		fmt.Println("can't find match files.")
		return
	}

	for i := 0; i < len(match_files); i++ {
		check_fin, _ := os.Stat(match_files[i])
		if check_fin.IsDir() {
			continue
		}
		md5_file(match_files[i])
	}
}
