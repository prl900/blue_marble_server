// Sample storage-quickstart creates a Google Cloud Storage bucket.
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	// Imports the Google Cloud Storage client package.
	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"github.com/golang/snappy"
)

const (
	rootPath  = "/Users/pablo/Downloads/"
	tileSize  = 1200
	imageSize = 21600
)

var letters map[string]int = map[string]int{"A": 0, "B": 1, "C": 2, "D": 3}
var channels map[int]string = map[int]string{0: "R", 1: "G", 2: "B"}

type Tile struct {
	Data  []byte
	Shape []int
}

func (t *Tile) Subset(x0, y0, x1, y1 int) *Tile {
	subset := []byte{}
	for i := x0 + y0*t.Shape[0]; i < x0+y1*(t.Shape[0]); i += t.Shape[0] {
		subset = append(subset, t.Data[i:i+(x1-x0)]...)
	}

	return &Tile{Data: subset, Shape: []int{x1 - x0, y1 - y0}}
}

func GetFileName(month int, letter string, number int) string {
	return fmt.Sprintf("world.topo.bathy.2004%.2d.3x21600x21600.%s%d.png", month, letter, number)
}

func GetTileOffsets(letter string, number int) []int {
	tileOffset := imageSize / tileSize
	return []int{letters[letter] * tileOffset, (number - 1) * tileOffset}
}

func ReadPNGImage(path string) (*image.RGBA, error) {
	src, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	img, err := png.Decode(src)
	if err != nil {
		return nil, err
	}

	return img.(*image.RGBA), nil
}

func SeparateChannels(rgba *image.RGBA) []*Tile {
	red := make([]byte, len(rgba.Pix)/4)
	green := make([]byte, len(rgba.Pix)/4)
	blue := make([]byte, len(rgba.Pix)/4)

	for i := range red {
		red[i] = rgba.Pix[i*4]
		green[i] = rgba.Pix[1+i*4]
		blue[i] = rgba.Pix[2+i*4]
	}

	t1 := &Tile{Data: red, Shape: []int{rgba.Rect.Dx(), rgba.Rect.Dy()}}
	t2 := &Tile{Data: green, Shape: []int{rgba.Rect.Dx(), rgba.Rect.Dy()}}
	t3 := &Tile{Data: blue, Shape: []int{rgba.Rect.Dx(), rgba.Rect.Dy()}}

	return []*Tile{t1, t2, t3}
}

func WriteObject(bktName, objName string, contents []byte) error {
	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	//projectID := "YOUR_PROJECT_ID"
	//projectID := "nci-gce"

	// Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("Failed to create client: %v", err)
	}

	// Creates a Bucket instance.
	bucket := client.Bucket(bktName)
	w := bucket.Object(objName).NewWriter(ctx)
	w.ContentType = "application/octet-stream"

	if _, err := w.Write([]byte(contents)); err != nil {
		return fmt.Errorf("Failed to write object to bucket: %v", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("Failed to close writer to bucket: %v", err)
	}
	// Close the client when finished.
	if err := client.Close(); err != nil {
		return fmt.Errorf("Failed to close client: %v", err)
	}

	return nil
}

func TileImage(month int, letter string, number int, bktName string) {
	fName := GetFileName(month, letter, number)
	img, err := ReadPNGImage(rootPath + fName)
	if err != nil {
		panic(err)
	}

	chans := SeparateChannels(img)
	offSets := GetTileOffsets(letter, number)

	for c, ch := range chans {
		for j := 0; j < imageSize; j += tileSize {
			for i := 0; i < imageSize; i += tileSize {
				subset := ch.Subset(i, j, i+tileSize, j+tileSize)
				oName := fmt.Sprintf("BM_%s_%.2d_%02d_%02d", channels[c], month, offSets[0]+i/tileSize, offSets[1]+j/tileSize)
				err := WriteObject(bktName, oName, snappy.Encode(nil, subset.Data))
				if err != nil {
					panic(err)
				}
				fmt.Println("Done", oName)
			}
		}
	}

}

func main() {
	//TileImage(12, "A", 1, "blue_marble")
	//TileImage(12, "A", 2, "blue_marble")
	TileImage(12, "B", 1, "blue_marble")
	TileImage(12, "B", 2, "blue_marble")
	TileImage(12, "C", 1, "blue_marble")
	TileImage(12, "C", 2, "blue_marble")
	TileImage(12, "D", 1, "blue_marble")
	TileImage(12, "D", 2, "blue_marble")
}
