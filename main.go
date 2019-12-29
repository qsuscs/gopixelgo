package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif" // unused imports to register image decoders
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
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
	fDeterm = flag.Bool(
		"deterministic", false, "initialise RNG deterministically")
	fHost = flag.String(
		"host", "localhost:1234", "host and port to connect to")
	fImage = flag.String("image", "image.png", "image file name")
	fN     = flag.Int("n", 1, "number of concurrent routines")
	fOnce  = flag.Bool("once", false, "only run once")
	fX     = flag.Int("x", 0, "start of the image (x)")
	fY     = flag.Int("y", 0, "start of the image (y)")
)

func main() {
	flag.Parse()

	if !*fDeterm {
		t := time.Now().UnixNano()
		log.Println("rand seed:", t)
		rand.Seed(t)
	}

	reader, err := os.Open(*fImage)
	if err != nil {
		log.Panic(err)
	}

	img, _, err := image.Decode(reader)
	if err != nil {
		log.Panic(err)
	}

	reader.Close()

	min := img.Bounds().Min
	max := img.Bounds().Max
	pxs := make([]pfPixel, 0, img.Bounds().Dx()*img.Bounds().Dy())
	offset := image.Pt(*fX, *fY)
	for x := min.X; x < max.X; x++ {
		for y := min.Y; y < max.Y; y++ {
			pxs = append(pxs, pfPixel{
				image.Pt(x, y).Add(offset), img.At(x, y)})
		}
	}

	log.Printf("image has %d pixels", len(pxs))

	var b strings.Builder
	for _, i := range rand.Perm(len(pxs)) {
		b.WriteString(pxs[i].String())
	}

	log.Println("length of pixel data:", len(b.String()))

	ctx, done := context.WithCancel(context.Background())
	log.Println("ctx:", ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		s := <-sig
		log.Println("got signal", s)
		done()
	}()

	work := make(chan []byte)
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println(
					"ctx closed, closing worker channel")
				close(work)
				return
			default:
				work <- []byte(b.String())
				if *fOnce {
					close(work)
					return
				}
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < *fN; i++ {
		go func(n int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", *fHost)
			if err != nil {
				log.Panic(n, err)
			}
			defer conn.Close()

			for w := range work {
				len, err := conn.Write(w)
				if err != nil {
					log.Panic(n, err)
				}
				log.Printf("%d written %d bytes", n, len)
			}
			log.Print(n, " done")
		}(i)
		wg.Add(1)
	}
	wg.Wait()
	log.Print("done")
}
