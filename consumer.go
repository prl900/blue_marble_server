// Sample storage-quickstart creates a Google Cloud Storage bucket.
package main

import (
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"os"

	// Imports the Google Cloud Storage client package.
	"cloud.google.com/go/storage"
	"github.com/golang/snappy"
	"golang.org/x/net/context"
)

const (
	rootPath  = "/Users/pablo/Downloads/"
	tileSize  = 1200
	imageSize = 21600
)

type TileRef struct {
	Path   string
	Offset []int
	Size   []int
}

type Tile struct {
	Data   []uint8
	Offset []int
	Size   []int
}

type Mosaic struct {
	Tiles      []TileRef
	TileStride int
	Shape      []int
}

func GetMosaic(latMin, latMax, lonMin, lonMax float64) Mosaic {
	minX := ((lonMin + 180) / 360) * (21600 * 4)
	maxX := ((lonMax + 180) / 360) * (21600 * 4)

	maxY := ((180 - (latMin + 90)) / 180) * (21600 * 2)
	minY := ((180 - (latMax + 90)) / 180) * (21600 * 2)

	x0 := int(minX) / tileSize
	x1 := int(maxX) / tileSize
	y0 := int(minY) / tileSize
	y1 := int(maxY) / tileSize

	fmt.Println(x0, x1, y0, y1)

	out := Mosaic{TileStride: x1 - x0 + 1, Shape: []int{int(maxX - minX), int(maxY - minY)}}

	var xOff, xSize int
	var yOff, ySize int

	for y := y0; y <= y1; y++ {
		yOff = 0
		ySize = tileSize
		if y == y0 {
			yOff = int(minY) % tileSize
		}
		if y == y1 {
			ySize = int(maxY) % tileSize
		}
		for x := x0; x <= x1; x++ {
			xOff = 0
			xSize = tileSize
			if x == x0 {
				xOff = int(minX) % tileSize
			}
			if x == x1 {
				xSize = int(maxX) % tileSize
			}
			tile := TileRef{Path: fmt.Sprintf("BM_R_12_%02d_%02d", x, y), Offset: []int{xOff, yOff}, Size: []int{xSize, ySize}}
			out.Tiles = append(out.Tiles, tile)
		}
	}
	return out
}

func ReadObject(bktName, objName string) ([]byte, error) {
	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	//projectID := "YOUR_PROJECT_ID"
	//projectID := "nci-gce"

	// Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		return []byte{}, fmt.Errorf("Failed to create client: %v", err)
	}

	// Creates a Bucket instance.
	bucket := client.Bucket(bktName)
	rc, err := bucket.Object(objName).NewReader(ctx)
	if err != nil {
		return []byte{}, fmt.Errorf("Failed creating reader: %v", err)
	}

	compData, err := ioutil.ReadAll(rc)
	rc.Close()
	if err != nil {
		return []byte{}, fmt.Errorf("Failed reading object: %v", err)
	}

	imgData, err := snappy.Decode(nil, compData)
	if err != nil {
		return []byte{}, fmt.Errorf("Failed decompressing data: %v", err)
	}

	// Close the client when finished.
	if err := client.Close(); err != nil {
		return []byte{}, fmt.Errorf("Failed to close client: %v", err)
	}

	return imgData, nil
}

func StitchMosaic(m Mosaic) *image.Gray {
	fmt.Println(m)

	out := &image.Gray{Pix: []byte{}, Stride: m.Shape[0], Rect: image.Rect(0, 0, m.Shape[0], m.Shape[1])}

	rows := len(m.Tiles) / m.TileStride

	for i := 0; i < rows; i++ {
		tiles := []Tile{}
		for _, tileRef := range m.Tiles[i*m.TileStride : (i+1)*m.TileStride] {
			data, err := ReadObject("blue_marble", tileRef.Path)
			if err != nil {
				panic(err)
			}
			tiles = append(tiles, Tile{Data: data, Offset: tileRef.Offset, Size: tileRef.Size})
			fmt.Println("Done:", tileRef)
		}
		for line := m.Tiles[i*m.TileStride].Offset[1]; line < m.Tiles[i*m.TileStride].Size[1]; line++ {
			for _, tile := range tiles {
				out.Pix = append(out.Pix, tile.Data[line*tileSize+tile.Offset[0]:line*tileSize+tile.Size[0]]...)
			}
		}
	}

	return out
}

func main() {
	//objName := GetTile(-34, -30, 150, 154)
	//m := GetMosaic(-34, -20, 142.5, 154)
	m := GetMosaic(42.6125, 43.0125, -1.8458, -1.4458)
	//m := GetMosaic(40, 43, 0, 3)
	//m := GetMosaic(-34.5, -30, 150, 155)
	fmt.Println(m)
	img := StitchMosaic(m)
	/*imgData, err := ReadObject("blue_marble", objName)
	if err != nil {
		fmt.Println(err)
		return
	}*/
	out, err := os.Create("./output.png")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = png.Encode(out, img)
	if err != nil {
		fmt.Println(err)
	}
}
