package main

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/dustin/go-humanize"
)

var client = &http.Client{Transport: &myTransport{}, Timeout: time.Second * 30}

func (t *myTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(
		"User-Agent", "https://github.com/Sorrow446/Catbox-Uploader",
	)
	return http.DefaultTransport.RoundTrip(req)
}

const uploadUrl = "https://catbox.moe/user/api.php"

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Uploaded += uint64(n)
	percentage := float64(wc.Uploaded) / float64(wc.Total) * float64(100)
	wc.Percentage = int(percentage)
	fmt.Printf("\r%d%%, %s/%s ", wc.Percentage, humanize.Bytes(wc.Uploaded), wc.TotalStr)
	return n, nil
}

func parseArgs() *Args {
	var args Args
	arg.MustParse(&args)
	return &args
}

func fileExists(path string) (bool, error) {
	f, err := os.Stat(path)
	if err == nil {
		return !f.IsDir(), nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func contains(lines []string, value string) bool {
	for _, line := range lines {
		if strings.EqualFold(line, value) {
			return true
		}
	}
	return false
}

func filterPaths(paths []string) ([]string, error) {
	var filteredPaths []string
	for _, path := range paths {
		exists, err := fileExists(path)
		if err != nil {
			return nil, err
		}
		if exists && !contains(filteredPaths, path) {
			filteredPaths = append(filteredPaths, path)
		}
	}
	return filteredPaths, nil
}

func checkSize(path string) (int64, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return -1, err
	}
	size := stat.Size()
	if size == 0 {
		return -1, errors.New("File is 0 bytes.")
	} else if size > 209715200 {
		return -1, errors.New("File exceeds 200MB limit.")
	}
	return size, nil
}

func outSetup(path string, wipe bool) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return nil, err
	}
	if wipe {
		err = os.Truncate(path, 0)
		if err != nil {
			f.Close()
			return nil, err
		}
	}
	return f, nil
}

func upload(path string, fname string, size int64) (string, error) {
	r, w := io.Pipe()
	m := multipart.NewWriter(w)
	f, err := os.Open(path)
	if err != nil {
		w.Close()
		m.Close()
		return "", err
	}
	defer f.Close()
	totalBytes := uint64(size)
	counter := &WriteCounter{Total: totalBytes, TotalStr: humanize.Bytes(totalBytes)}
	// Implement and get err channel working. Seems to hang.
	go func() {
		defer w.Close()
		defer m.Close()
		userHash, err := m.CreateFormField("userhash")
		if err != nil {
			return
		}
		userHash.Write([]byte(""))
		reqType, err := m.CreateFormField("reqtype")
		if err != nil {
			return
		}
		reqType.Write([]byte("fileupload"))
		part, err := m.CreateFormFile("fileToUpload", fname)
		if err != nil {
			return
		}
		_, err = io.Copy(part, f)
		if err != nil {
			return
		}
	}()
	req, err := http.NewRequest(http.MethodPost, uploadUrl, io.TeeReader(r, counter))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", m.FormDataContentType())
	do, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return "", errors.New(do.Status)
	}
	bodyBytes, err := io.ReadAll(do.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)
	fmt.Println("")
	return bodyString, nil
}

func writeUrl(f *os.File, fileUrl string) error {
	_, err := f.Write([]byte(fileUrl + "\n"))
	return err
}

func main() {
	var f *os.File
	args := parseArgs()
	paths, err := filterPaths(args.Paths)
	if err != nil {
		errString := fmt.Sprintf("Failed to filter paths.\n%s", err)
		panic(errString)
	}
	outPath := args.OutPath
	if outPath != "" {
		f, err = outSetup(outPath, args.Wipe)
		if err != nil {
			panic(err)
		}
		defer f.Close()
	}
	pathTotal := len(paths)
	for num, path := range paths {
		fmt.Printf("File %d of %d:\n", num+1, pathTotal)
		fname := filepath.Base(path)
		fmt.Println(fname)
		size, err := checkSize(path)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fileUrl, err := upload(path, fname, size)
		if err != nil {
			fmt.Println("Upload failed.\n", err)
			continue
		}
		fmt.Println(fileUrl)
		if outPath != "" {
			err = writeUrl(f, fileUrl)
			if err != nil {
				fmt.Println("Failed to write URL to output text file.\n", err)
			}
		}
	}
}
