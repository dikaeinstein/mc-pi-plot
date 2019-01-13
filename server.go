package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"net/http"

	"go-hep.org/x/hep/hplot"
	"golang.org/x/net/websocket"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

type wplot struct {
	Plot string `json:"plot"`
}

// Server that opens a websocket connection
// to send points and plots
type server struct {
	in    plotter.XYs     // points inside the circle
	out   plotter.XYs     // points outside the circle
	n     int             // number of samples
	data  chan [2]float64 // channel of (x, y) points randomly drawn
	plots chan wplot      // channel of base64-encoded PNG plots
}

func (s *server) dataHandler(ws *websocket.Conn) {
	for data := range s.plots {
		err := websocket.JSON.Send(ws, data)
		if err != nil {
			log.Printf("error sending data: %v\n", err)
		}
	}
}

func plotHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, page)
}

func newServer() *server {
	srv := &server{
		in:    make(plotter.XYs, 0, 1024),
		out:   make(plotter.XYs, 0, 1024),
		data:  make(chan [2]float64),
		plots: make(chan wplot),
	}

	go srv.run()

	return srv
}

func (s *server) run() {
	for v := range s.data {
		s.n++
		x := v[0]
		y := v[1]
		d := x*x + y*y
		pt := plotter.XY{X: x, Y: y}
		switch {
		case d < 1:
			s.in = append(s.in, pt)
		default:
			s.out = append(s.out, pt)
		}
		s.plots <- plot(s.n, s.in, s.out)
	}
}

func plot(n int, in, out plotter.XYs) wplot {
	radius := vg.Points(0.1)

	p := hplot.New()

	p.X.Label.Text = "x"
	p.X.Min = 0
	p.X.Max = 1
	p.Y.Label.Text = "y"
	p.Y.Min = 0
	p.Y.Max = 1

	pi := 4 * float64(len(in)) / float64(n)
	p.Title.Text = fmt.Sprintf("n = %d\nπ = %v", n, pi)

	sin, err := hplot.NewScatter(in)
	if err != nil {
		log.Fatal(err)
	}
	sin.Color = color.RGBA{255, 0, 0, 255} // red
	sin.Radius = radius

	sout, err := hplot.NewScatter(out)
	if err != nil {
		log.Fatal(err)
	}
	sout.Color = color.RGBA{0, 0, 255, 255} // blue
	sout.Radius = radius

	p.Add(sin, sout, hplot.NewGrid())

	return wplot{Plot: renderImg(p)}
}

// Approximate PI using Monte Carlo method
// draw a square, inscribe a circle within it,
// uniformly scatter objects of uniform size over the square,
// count the number of objects inside the circle and the
// total number of objects, the ratio of the inside-count
// and the total-sample-count is an estimate of the ratio
// of the two areas, which is π/4. So PI = π / 4 * 4
// area(circle) = πr^2 / 4; area(square) = r^2
// ratio area(circle) / area(square) = π / 4
func (s *server) pi(n int) {
	for i := 0; i < n; i++ {
		x := rand.Float64()
		y := rand.Float64()
		s.data <- [2]float64{x, y}
	}
	// return 4 * float64(inside) / float64(n)
}

func renderImg(p *hplot.Plot) string {
	size := 20 * vg.Centimeter
	canvas := vgimg.PngCanvas{Canvas: vgimg.New(size, size)}
	p.Draw(draw.New(canvas))
	out := new(bytes.Buffer)
	_, err := canvas.WriteTo(out)
	if err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(out.Bytes())
}
