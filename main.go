package main

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/tarm/serial"
	"golang.org/x/image/bmp"
	"image"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

func getLastestFile() string {
	files, err := ioutil.ReadDir("./out")
	if err != nil {
		log.Fatal(err)
	}
	return files[len(files)-1].Name()
}

// check if file is appear in folder but not sure that file is complete writing .
func checkFileExist(filename string) bool {
	fmt.Println(filename)
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// read bit map image file and process it into [4][4]bool
// return true if image match with some pattern and  2 dimension position array that  has true value if color is black
func readimage(s string) (bool, [4][4]bool, string) {
	reader, err := os.Open(s)
	if err != nil {
		log.Fatalf("err:", err)
	}
	defer reader.Close()
	found := false
	var im image.Image
	for im, err = bmp.Decode(reader); err != nil; {
	}
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
	// black value threshold
	black := []uint8{41, 41, 41}
	//white := []uint8{255, 255, 255}
	var pos [4][4]bool
	var result string
	// map pixel which color value which is higher than black value threshold
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
	// process image
	if !pos[0][1] && !pos[0][2] &&
		!pos[1][3] && !pos[2][3] &&
		pos[3][1] && pos[3][2] &&
		pos[1][0] && pos[2][0] {
		//fmt.Println("topright")
		found = true
		result = "upper"
	}
	if pos[0][1] && pos[0][2] &&
		pos[1][3] && pos[2][3] &&
		!pos[3][1] && !pos[3][2] &&
		!pos[1][0] && !pos[2][0] {
		//fmt.Println("bottomleft")
		found = true
		result = "lower"
	}
	if !pos[0][1] && pos[0][2] &&
		pos[1][3] && pos[2][3] &&
		!pos[3][1] && pos[3][2] &&
		!pos[1][0] && !pos[2][0] {
		//fmt.Println("left")
		result = "left"
		found = true
	}
	if pos[0][1] && !pos[0][2] &&
		!pos[1][3] && !pos[2][3] &&
		pos[3][1] && !pos[3][2] &&
		pos[1][0] && pos[2][0] {
		//fmt.Println("right")
		result = "right"
		found = true
	}
	if pos[0][1] && pos[0][2] &&
		pos[1][3] && !pos[2][3] &&
		!pos[3][1] && !pos[3][2] &&
		pos[1][0] && !pos[2][0] {
		//fmt.Println("bottom")
		result = "bottom"
		found = true
	}
	if !pos[0][1] && !pos[0][2] &&
		!pos[1][3] && pos[2][3] &&
		pos[3][1] && pos[3][2] &&
		pos[1][0] && !pos[2][0] {
		//fmt.Println("top")
		result = "top"
		found = true
	}
	fmt.Println("success")
	return found, pos, result
}

func main() {
	state := 0
	store := make([]string, 3)
	//setup config
	serialCofig := &serial.Config{
		Name:     "COM9",
		Baud:     9600,
		Parity:   0,
		StopBits: 1,
		Size:     8,
	}
	s, err := serial.OpenPort(serialCofig)
	defer s.Close()
	if err != nil {
		log.Fatalf("Error occur at open serial port")
	}
	time.Sleep(2 * time.Second)
	//main program loop
	for {
		cmd := make([]byte, 1)
		if state != 1 {
			m, _ := s.Read(cmd)
			fmt.Printf("%v", cmd[:m])
			if cmd[0] == byte('R') {
				state = 1
				break
			}
		} else {
			filename := strings.Split(getLastestFile(), "\\")[2]
			dirpath := "out/" + filename
			if checkFileExist(dirpath) {
				fmt.Printf("Found image  %v \n", filename)
				fmt.Println("waiting complete image")
				//loop until get valid file
				for testfile, err := os.Open(dirpath); err != nil; {
					log.Printf("chk file %v ,%v", testfile, err)
					_ = testfile.Close()
				}
				fmt.Println("image complete saved")
				if found, pos, result := readimage(dirpath); found == true {
					_, exist := Find(store, result)
					if !exist {
						// send data response to arduino via serial
						for row := range pos {
							buffer := make([]byte, 4)
							for col := range pos[row] {
								if pos[row][col] {
									buffer[col] = 1
								} else {
									buffer[col] = 0
								}
							}
							// write 1 image row buffer if not already has
							for n := range buffer {
								b, _ := s.Write([]byte{buffer[n]})
								temp := make([]byte, b)
								b, _ = s.Read(temp)
								fmt.Print(temp[:b])
							}
							fmt.Printf("||%v\n", buffer)
						}
						store = append(store, result)
						state = 0
					} else {
						fmt.Println("Already get this image process next file . . .")
					}
				}
				// increase number of file
				fmt.Println("Process success waiting for next file . . .")
				fmt.Println("===========================================")
			}
		}
	}
}
