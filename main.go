package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif" // unused imports to register image decoders
	_ "image/jpeg"
	_ "image/png"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

func pfColorString(c color.Color) string {
	r, g, b, a := c.RGBA()
	if byte(a) != 0xff {
		return fmt.Sprintf("%02x%02x%02x%02x",
			byte(r), byte(g), byte(b), byte(a))
	} else {
		return fmt.Sprintf("%02x%02x%02x", byte(r), byte(g), byte(b))
	}
}

func pfPixelString(x, y int, c color.Color) string {
	return fmt.Sprintf("PX %d %d %s\n", x, y, pfColorString(c))
}

var (
	flag_image = flag.String("image", "image.png", "image file name")
	flag_host  = flag.String("host", "localhost:1234", "host and port to connect to")
	flag_x     = flag.Int("x", 0, "start of the image (x)")
	flag_y     = flag.Int("y", 0, "start of the image (y)")
	flag_once  = flag.Bool("once", false, "only run once")
)

func main() {
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	reader, err := os.Open(*flag_image)
	if err != nil {
		panic(err)
	}

	img, _, err := image.Decode(reader)
	if err != nil {
		panic(err)
	}

	reader.Close()

	conn, err := net.Dial("tcp", *flag_host)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	min := img.Bounds().Min
	max := img.Bounds().Max
	pxs := make([]image.Point, 0, img.Bounds().Dx()*img.Bounds().Dy())
	for x := min.X; x < max.X; x++ {
		for y := min.Y; y < max.Y; y++ {
			pxs = append(pxs, image.Pt(x, y))
		}
	}

	var b strings.Builder
	for _, i := range rand.Perm(len(pxs)) {
		x, y := pxs[i].X, pxs[i].Y
		b.WriteString(pfPixelString(x+*flag_x, y+*flag_y, img.At(x, y)))
	}

	for ok := true; ok; ok = !*flag_once {
		conn.Write([]byte(b.String()))
	}
}
