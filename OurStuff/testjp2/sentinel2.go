package main

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"syscall"
	"unsafe"
)

type Pixel struct {
	R int
	G int
	B int
	A int
}

type ImgWithRGB struct {
	imagePath string
	rgb       RGB
}

type RGB struct {
	R int `json:"red"`
	G int `json:"green"`
	B int `json:"blue"`
}

func main() {

	var mod = syscall.NewLazyDLL("user32.dll")
	var proc = mod.NewProc("MessageBoxW")
	var MB_YESNOCANCEL = 0x00000003

	ret, _, _ := proc.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("This test is Done."))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Done Title"))),
		uintptr(MB_YESNOCANCEL))
	fmt.Printf("Return: %d\n", ret)

	//appengine.Main()
	//tryOpenJP2FileAndPrintPixels()

	// Connection string: "staging.johaa-178408.appspot.com"
	//resp, err := http.Get("staging.johaa-178408.appspot.com")

	/*err := http.ListenAndServe("localhost:5080", nil)
	if err != nil {
		log.Fatal(err)
	}*/
}

func tryOpenJP2FileAndPrintPixels() {
	// You can register another format here
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig) // Shouldn't really be needed - just an example I think

	file, err := os.Open("C:/Users/Archigo/Documents/GitHub/scalable_web_systems_fork/OurStuff/testjp2/qwe.png") // TODO: PATH TO JP2 FILE HERE - JARL :)

	if err != nil {
		fmt.Println("Error: File could not be opened")
		fmt.Println("%s", err)
		os.Exit(1)
	}

	defer file.Close()

	pixels, err := getPixels(file)

	if err != nil {
		fmt.Println("Error: Image could not be decoded")
		fmt.Println("%s", err)
		os.Exit(1)
	}

	fmt.Println(pixels)
}

func getPixels(file io.Reader) ([][]Pixel, error) {
	img, _, err := image.Decode(file)

	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	var pixels [][]Pixel
	for y := 0; y < height; y++ {
		var row []Pixel
		for x := 0; x < width; x++ {
			row = append(row, rgbaToPixel(img.At(x, y).RGBA()))
		}
		pixels = append(pixels, row)
	}

	return pixels, nil
}

func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	return Pixel{int(r / 257), int(g / 257), int(b / 257), int(a / 257)}
}
