package main

type myTransport struct{}

type WriteCounter struct {
	Total      uint64
	Uploaded   uint64
	Percentage int
	TotalStr   string
}

type Args struct {
	Paths   []string `arg:"positional, required"`
	OutPath string   `arg:"-o"`
	Wipe    bool     `arg:"-w"`
}
