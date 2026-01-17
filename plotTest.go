package main

import (
	"image"
	"strconv"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

func makePlotImage(wPx, hPx, row int, curve []float64) (image.Image, error) {
	// Build plot
	p := plot.New()
	p.Title.Text = "Light curve at row " + strconv.Itoa(row)
	p.X.Label.Text = "x"
	p.Y.Label.Text = "y"
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
	if err != nil {
		return nil, err
	}
	p.Add(line)
	p.Legend.Add("signal", line)

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
