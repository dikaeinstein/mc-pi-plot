package main

import (
	"flag"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

const page = `
<html>
	<head>
		<title>Monte Carlo</title>
		<script type="text/javascript">
		var sock = null;
		var plot = "";

		function update() {
			var p = document.getElementById("plot");
			p.src = "data:image/png;base64,"+plot;
		};

		window.onload = function() {
			sock = new WebSocket("ws://"+location.host+"/data");

			sock.onmessage = function(event) {
				var data = JSON.parse(event.data);
				plot = data.plot;
				update();
			};
		};

		</script>
	</head>

	<body>
		<div id="content">
			<p style="text-align:center;">
				<img id="plot" src="" alt="Not Available"></img>
			</p>
		</div>
	</body>
</html>
`

func main() {
	// n := flag.Int("n", 1e7, "MC sample size")
	// flag.Parse()
	// fmt.Printf("pi(%d) = %1.16f\n", *n, pi(*n))

	// const seed = 1234
	// src := rand.NewSource(seed)
	// rnd := rand.New(src)
	// const N = 10000
	// huni := hbook.NewH1D(100, 0, 1.0)

	// for i := 0; i < N; i++ {
	// 	r := rnd.Float64() // r is in [0.0, 1.0)
	// 	// fmt.Printf("%v\n", r)
	// 	huni.Fill(r, 1)
	// }
	// plot(huni, "uniform.png")
	log.SetPrefix("monte-carlo: ")
	log.SetFlags(0)

	n := flag.Int("n", 1e7, "number of samples")
	flag.Parse()

	srv := newServer()
	go srv.pi(*n)

	http.HandleFunc("/", plotHandle)
	http.Handle("/data", websocket.Handler(srv.dataHandler))

	log.Printf("listening on :8000...")

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// func plot(hist *hbook.H1D, filename string) {
// 	p := hplot.New()
// 	// Create a histogram of our values drawn
// 	// from the standard normal.
// 	h := hplot.NewH1D(hist)
// 	h.Color = color.NRGBA{0, 0, 255, 255}
// 	p.Add(h, hplot.NewGrid())

// 	const (
// 		width  = 10 * vg.Centimeter
// 		height = -1 // choose height automatically
// 	)

// 	if err := p.Save(width, height, filename); err != nil {
// 		log.Fatal(err)
// 	}
// }
