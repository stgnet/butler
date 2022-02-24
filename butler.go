package main

// butler cache and processing

type Hand struct {
	Trays []Tray
}

/*

	tray, tErr := Hand.Tray("path")
	glass, gErr := tray.Glass("file") // returns CACHE copy?
	glass.Pour() - get filestream for file
*/
