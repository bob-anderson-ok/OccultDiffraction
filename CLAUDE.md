# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OccultDiffraction is a Go application that provides full diffraction analysis for the Occult astronomy software. It simulates and visualizes Fresnel diffraction effects during stellar occultation events by asteroids. Developed by Bob Anderson and Dave Herald.

## Build and Run Commands

```bash
# Build the application
go build -o OccultDiffractionApp_X_X_X.exe

# Run with parameter file
OccultDiffractionApp_X_X_X.exe parameters
```

The version number is defined in `main.go` (line 20) and should match the executable name pattern.

## Architecture

### Processing Pipeline

1. **Parameter Loading** (`jsonProcessing.go`) - Parse JSON5 parameter files, validate inputs
2. **Fundamental Plane Construction** (`main.go`, `ellipseFuncs.go`) - Load external image or create from ellipse definitions
3. **Fresnel Diffraction** (`sincDiffraction.go`, `fresnelIntegrals.go`) - Matrix-based sinc method with BLAS acceleration
4. **Star Convolution** (`convolution.go`) - 2D FFT convolution with limb-darkened star PSF
5. **Output Generation** (`imageFuncs.go`, `plotFuncs.go`) - Generate images and light curve plots
6. **GUI Display** (`main.go`) - Fyne-based visualization

### Key Data Structure

`OccultationEvent` struct in `main.go` (lines 24-67) holds all parameters for an occultation event including geometry, wavelengths, star properties, and camera response.

### Module Responsibilities

| File | Purpose |
|------|---------|
| `main.go` | Entry point, GUI, orchestration |
| `imageFuncs.go` | Image I/O, Gray8/Gray16 formats, matrix conversions |
| `convolution.go` | 2D FFT convolution, star PSF with limb darkening |
| `jsonProcessing.go` | JSON5 parsing, parameter validation |
| `sincDiffraction.go` | Fresnel diffraction via sinc basis functions |
| `fresnelIntegrals.go` | Cephes polynomial approximations |
| `plotFuncs.go` | Light curve and camera response plotting |
| `pathFuncs.go` | Path calculations, edge detection |
| `ellipseFuncs.go` | Ellipse mathematics for body/satellite shapes |
| `openBLASmatrixMultiply.go` | BLAS wrapper for matrix operations |

## Dependencies

- **fyne.io/fyne/v2** - GUI framework
- **github.com/KevinWang15/go-json5** - JSON5 parsing
- **gonum.org/v1/plot** - Scientific plotting
- **gonum.org/v1/gonum** - Numerical computing
- **libopenblas.a** - BLAS acceleration (linked statically)

## Configuration

Parameter files use JSON5 format (see `parameters` file for example). Key parameters:

- `fundamental_plane_width_km` / `fundamental_plane_width_num_points` - Grid dimensions
- `observation_wavelength_nm` - Single wavelength (or use `path_to_qe_table_file` for multi-wavelength)
- `distance_au` or `parallax_arcsec` - Distance to asteroid
- `main_body` - Ellipse definition with `a`, `b`, `rotation_deg`
- `star_diam_on_plane_mas` - Star angular diameter
- `star_class` or `limb_darkening_coeff` - Limb darkening model

## Output Files

- `geometricShadow.png` - 8-bit geometric shadow
- `diffractionImage8bit.png` - 8-bit diffraction pattern (percentile stretched)
- `occultImage16bit.png` - 16-bit scientific data
- `camera_response.png` - QE table plot (if provided)

## Physics Notes

- Fresnel diffraction uses sinc basis function method (Cabillos approach)
- Limb darkening coefficients by spectral class: O=0.05, B=0.2, A=0.5, F/G/K/M=0.7
- Multi-wavelength processing weights each wavelength by camera QE response
- Uses Babinet's principle: intensity = |E_field - incident_wave|^2
