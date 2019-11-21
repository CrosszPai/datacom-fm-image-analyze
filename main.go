package main

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"golang.org/x/image/bmp"
	"log"
	"os"
	"strconv"
	"strings"
)

func checkFileExist(filename string) bool {
	if _, err := os.Stat(filename + ".bmp"); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func readimage(s string) (bool, [4][4]bool) {
	reader, err := os.Open(s + ".bmp")
	if err != nil {
		log.Fatalf("err:", err)
	}
	defer reader.Close()
	found := false
	im, err := bmp.Decode(reader)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	croptop := imaging.CropAnchor(im, 240, 200, imaging.Center)
	resize := imaging.Resize(croptop, 4, 4, imaging.Lanczos)
	newImage := imaging.AdjustContrast(resize, 60)
	lpp := new(bytes.Buffer)
	_ = bmp.Encode(lpp, newImage)
	//for i := range lpp.Bytes() {
	//	fmt.Printf("0x%02x,", lpp.Bytes()[i])
	//}

	//fmt.Println("")

	out, _ := os.Create(strings.Split(s, "/")[0]+"/" + "_4x4_" + strings.Split(s, "/")[1] + ".bmp")
	defer out.Close()
	_, _ = out.Write(lpp.Bytes())
	black := []uint8{41, 41, 41}
	//white := []uint8{255, 255, 255}
	var pos [4][4]bool
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			r, g, b, _ := newImage.At(x, y).RGBA()
			value := []uint8{uint8(r), uint8(g), uint8(b)}
			//bl := true
			pos[y][x] = true
			for i := range value {
				if value[i] > black[i] {
				//	bl = false
					pos[y][x] = false
				}
			}
			//if bl == true {
			//	fmt.Printf("%s", "b")
			//} else {
			//	fmt.Printf("%s", "w")
			//}
		}
		//fmt.Println("")
	}

	if !pos[0][1] && !pos[0][2] &&
		!pos[1][3] && !pos[2][3] &&
		pos[3][1] && pos[3][2] &&
		pos[1][0] && pos[2][0] {
		//fmt.Println("topright")
		found = true
	}
	if pos[0][1] && pos[0][2] &&
		pos[1][3] && pos[2][3] &&
		!pos[3][1] && !pos[3][2] &&
		!pos[1][0] && !pos[2][0] {
		//fmt.Println("bottomleft")
		found = true
	}
	if !pos[0][1] && pos[0][2] &&
		pos[1][3] && pos[2][3] &&
		!pos[3][1] && pos[3][2] &&
		!pos[1][0] && !pos[2][0] {
		//fmt.Println("left")
		found = true
	}
	if pos[0][1] && !pos[0][2] &&
		!pos[1][3] && !pos[2][3] &&
		pos[3][1] && !pos[3][2] &&
		pos[1][0] && pos[2][0] {
		//fmt.Println("right")
		found = true
	}
	if pos[0][1] && pos[0][2] &&
		pos[1][3] && !pos[2][3] &&
		!pos[3][1] && !pos[3][2] &&
		pos[1][0] && !pos[2][0] {
		//fmt.Println("bottom")
		found = true
	}
	if !pos[0][1] && !pos[0][2] &&
		!pos[1][3] && pos[2][3] &&
		pos[3][1] && pos[3][2] &&
		pos[1][0] && !pos[2][0] {
		//fmt.Println("top")
		found = true
	}
	return found, pos
}

func main() {
	start := 0
	//serialCofig := &serial.Config{
	//	Name: "COM7",
	//	Baud: 115209,
	//}
	//s, err := serial.OpenPort(serialCofig)
	//defer s.Close()
	//if err != nil {
	//	log.Fatalf("Error occur at open serial port")
	//}
	for {
		dirpath := "out/" + strconv.Itoa(start)
		if checkFileExist(dirpath) {
			if found, pos := readimage(dirpath); found == true {
				// send response to arduino via serial
				for row := range pos {
					buffer := make([]byte, 4)
					for col := range pos[row] {
						if pos[row][col] {
							buffer[col] = 1
						} else {
							buffer[col] = 0
						}
					}
					//for _,err:=s.Write(buffer);err != nil;{
					//}
					fmt.Printf("%v\n", buffer)
				}
			}
			start++
		}
	}
}
