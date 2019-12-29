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

type pfPixel struct {
	p image.Point
	c color.Color
}

func (px pfPixel) String() string {
	r, g, b, a := px.c.RGBA()
	x := px.p.X
	y := px.p.Y
	var as string
	if byte(a) != 0xff {
		as = fmt.Sprintf("%02x", byte(a))
	}
	return fmt.Sprintf("PX %d %d %02x%02x%02x%s\n",
		x, y, byte(r), byte(g), byte(b), as)
}

var (
	fImage = flag.String("image", "image.png", "image file name")
	fHost  = flag.String("host", "localhost:1234", "host and port to connect to")
	fX     = flag.Int("x", 0, "start of the image (x)")
	fY     = flag.Int("y", 0, "start of the image (y)")
	fOnce  = flag.Bool("once", false, "only run once")
)

func main() {
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	reader, err := os.Open(*fImage)
	if err != nil {
		panic(err)
	}

	img, _, err := image.Decode(reader)
	if err != nil {
		panic(err)
	}

	reader.Close()

	conn, err := net.Dial("tcp", *fHost)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	min := img.Bounds().Min
	max := img.Bounds().Max
	pxs := make([]pfPixel, 0, img.Bounds().Dx()*img.Bounds().Dy())
	for x := min.X; x < max.X; x++ {
		for y := min.Y; y < max.Y; y++ {
			pxs = append(pxs, pfPixel{
				image.Pt(x+*fX, y+*fY), img.At(x, y)})
		}
	}

	var b strings.Builder
	for _, i := range rand.Perm(len(pxs)) {
		b.WriteString(pxs[i].String())
	}

	for ok := true; ok; ok = !*fOnce {
		conn.Write([]byte(b.String()))
	}
}
