package main

import (
	"bytes"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var a fyne.App
var w fyne.Window
var ripEntry *widget.Entry
var ripButton *widget.Button
var progressBar *widget.ProgressBarInfinite
var infoLabel *widget.Label

func init() {
	a = app.NewWithID("fun.fk.corps")
	w = a.NewWindow("Skillcapped Ripper")
	//Rip Entry
	ripEntry = widget.NewMultiLineEntry()
	ripEntry.PlaceHolder = "Put your urls here"

	//Rip Button
	ripButton = widget.NewButton("RIP!", RipQueue)
	//Progress bar
	progressBar = widget.NewProgressBarInfinite()
	progressBar.Stop()
	//Label init
	infoLabel = widget.NewLabel("Here will appear info about the rippin'")
	w.Resize(fyne.Size{800, 600})

	border := container.NewBorder(
		infoLabel,
		container.NewVBox(progressBar, ripButton),
		nil,
		nil,
		ripEntry,
		//widget.NewLabel("Border Container")
	)

	os.Mkdir("tmp", os.FileMode(0755))
	w.SetContent(border)
}

var (
	RipSize         int
	CurrentlyRippin int
	RipVideoID      string
	CurrentPartID   string
)

func UpdateRipInfo() {
	infoLabel.SetText(fmt.Sprintf("Ripping video %v/%v with ID: %v\nPart %v/???", RipSize, CurrentlyRippin, RipVideoID, CurrentPartID))
}

func exit() {
	os.Exit(0)
}

func RipQueue() {
	if !isFFMPEGInstalled() {
		b := dialog.NewError(errors.New("ffmpeg is not detected\nIf you are on windows it might not be added to PATH"), w)
		b.SetOnClosed(exit)
		b.Show()
		return
	}

	progressBar.Start()
	urls := strings.Split(ripEntry.Text, "\n")
	RipSize = len(urls)
	CurrentlyRippin = 0
	for _, url := range urls {
		RipVideoID = GetVideoIDfromURL(url)
		CurrentlyRippin++
		Rip(RipVideoID)
		UpdateRipInfo()
	}
	infoLabel.SetText("RIPPIN' DONE!")
	progressBar.Stop()
	ripEntry.SetText("DONE!")
}

const (
	strLength = 5
)

func Rip(vidID string) {
	//Generate all ids from 1 to math.Pow10(strLength)-1
	//They look like 00122, 01321 00003 etc
	var builder strings.Builder
	for i := 1; i < int(math.Pow10(strLength)); i++ {
		UpdateRipInfo()
		//Add the missing 0 to the ids
		var tempBuilder strings.Builder
		for j := 0; j < strLength-intLength(i); j++ {
			tempBuilder.WriteByte('0')
		}
		tempBuilder.WriteString(strconv.Itoa(i))

		CurrentPartID = tempBuilder.String()

		partPath := path.Join("tmp", fmt.Sprintf("%v.ts", CurrentPartID))

		status, err := DownloadFile(partPath, fmt.Sprintf("https://d13z5uuzt1wkbz.cloudfront.net/%v/HIDDEN4500-%v.ts", vidID, CurrentPartID))
		if err != nil {
			log.Fatalln("Error on downloading video, error", err.Error())
			os.Exit(0)
		}
		if status != 200 {
			log.Printf("Error on downloading part %v from video id %v, status %v", CurrentPartID, vidID, status)
			os.Remove(partPath)
			break
		}

		s, _ := filepath.Abs(partPath)
		builder.WriteString(fmt.Sprintf("file '%v'\n", s))
	}

	tsPath, _ := filepath.Abs(path.Join("tmp", "ts"))

	os.WriteFile(tsPath, []byte(builder.String()), os.FileMode(0755))

	infoLabel.SetText("Converting video to mp4")
	convertTs(tsPath, fmt.Sprintf("video_%v.mp4", vidID))
}

func isFFMPEGInstalled() bool {
	proc := exec.Command("ffmpeg", "-version")
	var outb bytes.Buffer
	proc.Stdout = &outb
	err := proc.Run()
	if err != nil {
		return false
	}
	if strings.Contains(outb.String(), "ffmpeg version") {
		return true
	}
	return false
}

func convertTs(tsPath, outputPath string) {
	proc := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", tsPath, "-c", "copy", outputPath)
	var outb, errb bytes.Buffer
	proc.Stdout = &outb
	proc.Stderr = &errb
	err := proc.Run()

	if err != nil {
		fmt.Println(outb.String(), errb.String())
		log.Fatalln("Error with ffmpeg", err)
	}

	os.RemoveAll("tmp")
	os.MkdirAll("tmp", os.FileMode(0755))
}

func DownloadFile(filepath string, url string) (int, error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return resp.StatusCode, nil
	}
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return -1, err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return resp.StatusCode, err
}

func intLength(nr int) int {
	c := 0
	for nr != 0 {
		nr = nr / 10
		c++
	}
	return c
}

func GetVideoIDfromURL(url string) string {
	ret := strings.Split(url, "/")
	return ret[len(ret)-1]
}

func main() {
	w.ShowAndRun()
}
