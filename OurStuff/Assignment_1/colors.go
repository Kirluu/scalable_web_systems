package main

import (
	"math"
	"sort"
	"fmt"
	"io"
	"image"
	"image/png"
	"os"
	"log"
	"google.golang.org/appengine/urlfetch"
	"google.golang.org/appengine"
	"net/http"
	"io/ioutil"
)

type ImgWithRGB struct {
	imagePath string
	rgb RGB
}

type RGB struct {
	R int `json:"red"`
	G int `json:"green"`
	B int `json:"blue"`
}



// ---------------------------------- IMAGE LOADING LOGIC (Based on 'image' library) ----------------------------------


// todo-NOTE: Make sure to manually delete file when done reading it, by using: "defer os.Remove(tmpfile.Name())" for each file you get from this function
func downloadFileAsTemp(imageUrl string, r *http.Request) (*os.File, error) {
	content := []byte("temporary file's content")
	tmpfile, err := ioutil.TempFile("", "jp2_IMG_")
	if err != nil {
		log.Fatal(err)
	}

	//defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write(content); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	return downloadFileIntoFilePath(tmpfile.Name(), imageUrl, r)
}

// Given a save-file-path and a URL with a resource of the save-file-format, the file on the URL is saved at the filepath.
// Returns a reference to the os.File saved.
func downloadFileIntoFilePath(filepath string, imageUrl string, r *http.Request) (*os.File, error) {
	// Create the file
	out, err := os.Create(filepath) // string with relative or full pathing, name and file-extension
	if err != nil  {
		return nil, err
	}
	defer out.Close()
	// first create a new context
	c := appengine.NewContext(r)
	// and use that context to create a new http client
	client := urlfetch.Client(c)
	// now we can use that http client as before
	resp, err := client.Get(imageUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Write the body to the created file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return nil, err
	}
	return out, nil
}


func getAverageRGBValue(redFile io.Reader, greenFile io.Reader, blueFile io.Reader) RGB {
	return RGB {
		R: getAverage_R_G_or_B_Value(redFile),
		G: getAverage_R_G_or_B_Value(greenFile),
		B: getAverage_R_G_or_B_Value(blueFile),
	}
}

// Takes a file-reader and
func getAverage_R_G_or_B_Value(file io.Reader) int {
	pixels, err := getPixels(file) // Assumption: This method works
	if (err != nil) {
		log.Fatal("%s", err)
	}

	xDimLength := 0
	sum := 0
	for y := 0; y < len(pixels); y++ {
		xDimLength = len(pixels[y])
		for x := 0; x < len(pixels[y]); x++ {
			sum += pixels[y][x]
		}
	}

	// Return average (as integer division to get concrete, viable value (floored)
	return sum / len(pixels) * xDimLength
}

// CALL THIS FOR TESTING DECODING OF JP2 FILES
func tryOpenJP2FileAndPrintPixels() {
	// You can register another format here
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig) // Shouldn't really be needed - just an example I think

	file, err := os.Open("./image.png") // TODO: PATH TO JP2 FILE HERE - JARL :)

	if err != nil {
		fmt.Println("Error: File could not be opened")
		os.Exit(1)
	}

	defer file.Close()

	pixels, err := getPixels(file)

	if err != nil {
		fmt.Println("Error: Image could not be decoded")
		os.Exit(1)
	}

	fmt.Println(pixels)
}

// Inspiration from: https://stackoverflow.com/questions/33186783/get-a-pixel-array-from-from-golang-image-image
func getPixels(file io.Reader) ([][]int, error) {
	img, _, err := image.Decode(file)

	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	var pixels [][]int
	for y := 0; y < height; y++ {
		var row []int
		for x := 0; x < width; x++ {
			row = append(row, rgbaToPixel(img.At(x, y).RGBA()))
		}
		pixels = append(pixels, row)
	}

	return pixels, nil
}
// img.At(x, y).RGBA() returns four uint32 values; we want a Pixel
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) int {
	return int(r / 257) // todo-NOTE: ASSUMPTION: Red, Green and Blue are the same for the given image's RGB values, so we just read Red.
}




// ---------------------------------- COLOR RANKING LOGIC ----------------------------------

// Performs Euclidean distance between two colors
func L2 (c1 RGB, c2 RGB) float64 {
	return math.Sqrt(math.Pow(float64(c2.R - c1.R), 2) + math.Pow(float64(c2.G - c1.G), 2) + math.Pow(float64(c2.B - c1.B), 2))
}

// Actually modifies the given list to be sorted, but also returns it again.
func orderByColorBand(r bool, g bool, b bool, lst []ImgWithRGB) []ImgWithRGB {
	cStr := ""
	if (r) {
		cStr = "red"
		sort.Slice(lst, func(i, j int) bool {
			return lst[i].rgb.R < lst[j].rgb.R
		})
	} else if (g) {
		cStr = "green"
		sort.Slice(lst, func(i, j int) bool {
			return lst[i].rgb.G < lst[j].rgb.G
		})
	} else if (b) {
		cStr = "blue"
		sort.Slice(lst, func(i, j int) bool {
			return lst[i].rgb.B < lst[j].rgb.B
		})
	}

	// Just printing for testing purposes:
	fmt.Printf("Printing sorted by the %s color band:\n", cStr)
	for _, v := range lst {
		if (r) {
			fmt.Printf("%+v | ", v.rgb.R)
		} else if (g) {
			fmt.Printf("%+v | ", v.rgb.G)
		} else if (b) {
			fmt.Printf("%+v | ", v.rgb.B)
		}
	}
	fmt.Println()

	return lst
}


// Actually modifies the given list to be sorted, but also returns it again.
func orderByColor(c RGB, lst []ImgWithRGB) []ImgWithRGB {
	sort.Slice(lst, func(i, j int) bool {
		return L2(lst[i].rgb, c) < L2(lst[j].rgb, c)
	})

	// Just printing for testing purposes:
	fmt.Printf("Printing sorted RGB values relative to %+v\n", c)
	for _, v := range lst {
		fmt.Printf("%+v | ", v.rgb)
	}
	fmt.Println()

	return lst
}

func orderBy2Colors(c1 RGB, c2 RGB, lst []ImgWithRGB) []ImgWithRGB {

	sort.Slice(lst, func(i, j int) bool {
		return L2(lst[i].rgb, c1)/L2(lst[i].rgb, c2) < L2(lst[j].rgb, c1)/L2(lst[j].rgb, c2)
	})

	// Just printing for testing purposes:
	fmt.Printf("Printing sorted RGB values relative to %+v and %+v (distance from the line between the two)\n", c1, c2)
	for _, v := range lst {
		fmt.Printf("%+v | ", v.rgb)
	}
	fmt.Println()

	return lst
}