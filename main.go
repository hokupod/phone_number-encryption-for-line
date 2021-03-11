package main

import (
	"crypto/sha256"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/moxtsuan/go-nkf"
	"golang.org/x/sync/errgroup"
)

var isNumberOnly = func() func(string) bool {
	re := regexp.MustCompile("^[0-9]+$")

	return func(str string) bool {
		re := re.Copy()
		return re.MatchString(str)
	}
}()

func main() {
	flag.Parse()
	path := flag.Arg(0)
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("readFileError: %s\n", err.Error())
		os.Exit(1)
	}
	defer file.Close()
	reader, err := createCsvReader(file)
	if err != nil {
		print(err.Error())
		os.Exit(1)
	}
	var eg errgroup.Group
	m := new(sync.Mutex)
	done := false
	for {
		eg.Go(func() error {
			m.Lock()
			record, err := reader.Read()
			m.Unlock()
			if err == io.EOF {
				done = true
				return nil
			}
			if err != nil {
				done = true
				return fmt.Errorf("readLineError: %s\n", err.Error())
			}
			phone := formatPhoneNumber(record[0])
			if isMobileNumber(phone) {
				println(toHash(phone))
			}
			return nil
		})
		if done {
			break
		}
	}
	if err := eg.Wait(); err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
func toHash(str string) string {
	b := []byte(str)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}
func formatPhoneNumber(str string) string {
	str = strings.Replace(str, "-", "", -1)
	if strings.HasPrefix(str, "+81") {
		str = str[3:]
	}
	if !strings.HasPrefix(str, "0") {
		str = "0" + str
	}
	return str
}
func isMobileNumber(str string) bool {
	isMobileNumberLength := utf8.RuneCountInString(str) == 11
	return isNumberOnly(str) && isMobileNumberLength &&
		(strings.HasPrefix(str, "050") ||
			strings.HasPrefix(str, "070") ||
			strings.HasPrefix(str, "080") ||
			strings.HasPrefix(str, "090"))
}
func createCsvReader(file *os.File) (*csv.Reader, error) {
	csvString, err := nkf.Convert(file, "", "UTF8", "")
	if err != nil {
		return nil, fmt.Errorf("convertError: %s\n", err.Error())
	}
	return csv.NewReader(strings.NewReader(csvString)), nil
}
