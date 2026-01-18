package main

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"strconv"

	"gonum.org/v1/plot"

	// Liberation fonts register automatically on import
	_ "gonum.org/v1/plot/font/liberation"

	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

func makePlotImage(wPx, hPx, row int, curve, edges []float64) (image.Image, error) {

	p := plot.New()

	p.Y.Min = -0.2
	p.Y.Max = 1.5

	// Modify the font fields directly on existing styles
	p.Title.TextStyle.Font.Typeface = "Liberation"
	p.Title.TextStyle.Font.Variant = "Sans"
	p.Title.TextStyle.Font.Size = vg.Points(12)

	p.X.Label.TextStyle.Font.Typeface = "Liberation"
	p.X.Label.TextStyle.Font.Variant = "Sans"
	p.X.Label.TextStyle.Font.Size = vg.Points(12)

	p.Y.Label.TextStyle.Font.Typeface = "Liberation"
	p.Y.Label.TextStyle.Font.Variant = "Sans"
	p.Y.Label.TextStyle.Font.Size = vg.Points(12)

	p.X.Tick.Label.Font.Typeface = "Liberation"
	p.X.Tick.Label.Font.Variant = "Sans"
	p.X.Tick.Label.Font.Size = vg.Points(10)

	p.Y.Tick.Label.Font.Typeface = "Liberation"
	p.Y.Tick.Label.Font.Variant = "Sans"
	p.Y.Tick.Label.Font.Size = vg.Points(10)

	p.Title.Text = "Light curve at row " + strconv.Itoa(row)
	p.X.Label.Text = "column (pixel units)"
	p.Y.Label.Text = "normalized intensity"
	p.X.Tick.Marker = StepTicks{Step: float64(len(curve) / 20), Format: "%.0f"}
	//p.X.Tick.Marker = NoLabelTicks{Ticker: plot.DefaultTicks{}}

	p.Y.Tick.Marker = StepTicks{Step: 0.2, Format: "%.2f"}
	p.Add(plotter.NewGrid()) // grid + ticks

	// Data
	n := len(curve)
	pts := make(plotter.XYs, n)
	for i := 0; i < n; i++ {
		x := float64(i)
		y := curve[i]
		pts[i].X = x
		pts[i].Y = y
	}

	line, err := plotter.NewLine(pts)
	line.Color = color.RGBA{R: 0, G: 0, B: 255, A: 255} // blue

	if err != nil {
		return nil, err
	}
	p.Add(line)

	if len(edges) > 0 {
		for _, edge := range edges {
			vpts := plotter.XYs{
				{X: edge, Y: -0.1},
				{X: edge, Y: 1.3},
			}

			vline, err := plotter.NewLine(vpts)
			if err != nil {
				panic(err)
			}

			p.Add(vline)
			//p.Legend.Add("signal", line)

			vline.Dashes = []vg.Length{
				vg.Points(6), // dash length
				vg.Points(4), // gap length
			}
			vline.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255} // red
		}
	}

	hpts := plotter.XYs{
		{X: 0.0, Y: 0.0},
		{X: float64(n), Y: 0.0},
	}

	hline, err := plotter.NewLine(hpts)
	if err != nil {
		panic(err)
	}

	p.Add(hline)

	hline.Dashes = []vg.Length{
		vg.Points(6), // dash length
		vg.Points(4), // gap length
	}
	hline.Color = color.RGBA{R: 0, G: 0, B: 0, A: 255} // black

	// Render into an in-memory image
	// Choose a "virtual" size in vg units and map to pixels via DPI.
	const dpi = 96
	width := vg.Length(float64(wPx)) * vg.Inch / dpi
	height := vg.Length(float64(hPx)) * vg.Inch / dpi

	c := vgimg.New(width, height)
	dc := draw.New(c)
	p.Draw(dc)

	return c.Image(), nil
}

type StepTicks struct {
	Step   float64
	Format string
}

func (t StepTicks) Ticks(min, max float64) []plot.Tick {
	var ticks []plot.Tick
	start := math.Ceil(min/t.Step) * t.Step
	for v := start; v <= max; v += t.Step {
		ticks = append(ticks, plot.Tick{
			Value: v,
			Label: fmt.Sprintf(t.Format, v),
		})
	}
	return ticks
}

func MakeCameraResponsePlot(data [][2]float64, filename string) {
	p := plot.New()

	// Modify the font fields directly on existing styles
	p.Title.TextStyle.Font.Typeface = "Liberation"
	p.Title.TextStyle.Font.Variant = "Sans"
	p.Title.TextStyle.Font.Size = vg.Points(12)

	p.X.Label.TextStyle.Font.Typeface = "Liberation"
	p.X.Label.TextStyle.Font.Variant = "Sans"
	p.X.Label.TextStyle.Font.Size = vg.Points(12)

	p.Y.Label.TextStyle.Font.Typeface = "Liberation"
	p.Y.Label.TextStyle.Font.Variant = "Sans"
	p.Y.Label.TextStyle.Font.Size = vg.Points(12)

	p.X.Tick.Label.Font.Typeface = "Liberation"
	p.X.Tick.Label.Font.Variant = "Sans"
	p.X.Tick.Label.Font.Size = vg.Points(10)

	p.Title.Text = "Camera response vs Wavelength from file: " + filename
	p.X.Label.Text = "Wavelength (nm)"
	p.Y.Label.Text = "Relative response"

	p.X.Tick.Marker = StepTicks{Step: 25.0, Format: "%.0f"}

	p.Y.Tick.Marker = StepTicks{Step: 0.1, Format: "%.2f"}
	p.Add(plotter.NewGrid()) // grid + ticks

	p.Y.Min = 0.0
	p.Y.Max = 1.1

	// Find the max weight - we will use that to calculate relative response
	var maxWeight = 0.0
	for _, pair := range data {
		if pair[1] > maxWeight {
			maxWeight = pair[1]
		}
	}
	// Data
	n := len(data)
	pts := make(plotter.XYs, n)
	for i := 0; i < n; i++ {
		x := data[i][0]
		y := data[i][1] / maxWeight
		pts[i].X = x
		pts[i].Y = y
	}

	linePoints, scatterPoints, err := plotter.NewLinePoints(pts)
	if err != nil {
		log.Fatal(err)
	}
	linePoints.Color = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	linePoints.Width = vg.Points(1)

	scatterPoints.Shape = draw.CircleGlyph{}
	scatterPoints.Radius = vg.Points(2)
	scatterPoints.Color = color.RGBA{R: 120, G: 120, B: 120, A: 255}

	p.Add(linePoints, scatterPoints)

	hpts := plotter.XYs{
		{X: data[0][0], Y: 0.0},
		{X: data[n-1][0], Y: 0.0},
	}

	hline, err := plotter.NewLine(hpts)
	if err != nil {
		panic(err)
	}

	p.Add(hline)

	hline.Dashes = []vg.Length{
		vg.Points(6), // dash length
		vg.Points(4), // gap length
	}
	hline.Color = color.RGBA{R: 0, G: 0, B: 0, A: 255} // black

	if err := p.Save(8*vg.Inch, 4*vg.Inch, "camera_response.png"); err != nil {
		log.Fatal(err)
	}
	return
}
