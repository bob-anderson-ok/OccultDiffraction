package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	json "github.com/KevinWang15/go-json5"
)

// !!!!! This MUST match the app name given in the run configuration !!!!!
const version = "1_1_0"

// !!!!! This MUST match the app name given in the run configuration !!!!!

type OccultationEvent struct {
	FplaneImage                     *image.Gray // A square array of uint8 values
	ShowInput                       bool
	RotateGroundShadowTo90pa        bool
	WindowSizePixels                int
	PathForGroundShadowOutputFolder string
	PathToExternalImage             string
	PathToQEtable                   string
	QEtableStride                   int
	Title                           string
	FundamentalPlaneWidthKm         float64
	FundamentalPlaneWidthPoints     int
	CameraExposureSecs              float64
	ObservationWavelengthNm         float64
	DxKmPerSec                      float64
	DyKmPerSec                      float64
	StarName                        string
	StarDiamMas                     float64
	StarDiamKm                      float64
	LimbDarkeningCoeff              float64
	StarClass                       string
	ParallaxArcsec                  float64
	DistanceAu                      float64
	MainBodyGiven                   bool
	MainBodyXCenterKm               float64
	MainBodyYCenterKm               float64
	MainbodyMajorAxisKm             float64
	MainbodyMinorAxisKm             float64
	MainbodyMajorAxisPaDegrees      float64
	SatelliteGiven                  bool
	SatelliteXCenterKm              float64
	SatelliteYCenterKm              float64
	SatelliteMajorAxisKm            float64
	SatelliteMinorAxisKm            float64
	SatelliteMajorAxisPaDegrees     float64
}

func main() {

	programStart := time.Now()

	// We supply an ID (hopefully unique) because we may need to use the preferences API
	myApp := app.NewWithID("com.gmail.ok.anderson.intensityMatrix")
	w := myApp.NewWindow("OccultDiffractionApp - user friendly diffraction image (8 bit grayscale png)")
	w.Resize(fyne.Size{Height: 800, Width: 1200})
	args := os.Args

	if len(args) != 2 {
		fmt.Println("\n\tWrong number of arguments.\n\tUsage: OccultDiffractionApp <parameter-file>")
		os.Exit(1)
	}

	path := args[1]

	// Read the Json5 (or Json) parameter file
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(fmt.Errorf("\n\tAttempt to read input file %q failed: %w\n", path, err))
		os.Exit(2)
	}

	// Parse json(5) data into a generic container
	var jsonTable map[string]interface{}
	err = json.Unmarshal(data, &jsonTable)
	if err != nil {
		fmt.Println(fmt.Errorf("\n\tFormat error in file %q: %w\n", path, err))
		os.Exit(3)
	}

	var event OccultationEvent
	msg, ok := validateJsonFileAndFillEvent(jsonTable, &event)
	if !ok {
		fmt.Println(msg)
		os.Exit(4)
	}

	// Check for user wanting printout of complete jsonTable
	if event.ShowInput {
		fmt.Printf("%s", "\nPrintout of  complete jsonTable contents...\n")
		fmt.Println(string(data))
	}

	// Sanity check on number of points in a fundamental plane
	if event.FundamentalPlaneWidthPoints < 10 {
		fmt.Println(fmt.Errorf("\n\tThe fundamental plane width must be at least 10 points."))
		os.Exit(15)
	}

	Npts := event.FundamentalPlaneWidthPoints

	fmt.Printf("\nVersion %s\n\n", version)
	// Calculate resolution in fundamental plane
	resolution := event.FundamentalPlaneWidthKm / float64(Npts)
	fmt.Printf("Resolution in fundamental plane is %0.3f km/pixel\n", resolution)
	fresnelScale := FresnelScale(event.ObservationWavelengthNm, event.DistanceAu)
	fmt.Printf("Fresnel scale is %0.3f km\n", fresnelScale)
	samplesPerFresnelScale := int(fresnelScale / resolution)
	fmt.Printf("Samples per Fresnel scale is %d  (To see diffraction effects, this number should be at least 5)\n\n", samplesPerFresnelScale)

	start := time.Now() // Time generation of geometric shadow

	// Deal with external image supplied by the user.
	if event.PathToExternalImage != "" {
		f, err := os.Open(event.PathToExternalImage)
		if err != nil {
			fmt.Println(fmt.Errorf("\n\tAttempt to read external image %q failed: %w\n", event.PathToExternalImage, err))
			os.Exit(5)
		}
		defer f.Close()

		img, err := png.Decode(f)
		if err != nil {
			fmt.Println(fmt.Errorf("\n\tAttempt to decode external image %q failed: %w\n", event.PathToExternalImage, err))
			os.Exit(6)
		}

		if img.Bounds().Dx() != img.Bounds().Dy() {
			fmt.Println(fmt.Errorf("\n\tThe supplied external image %q is not square.", event.PathToExternalImage))
			os.Exit(7)
		}

		// We require that an external image is supplied in GRAY format (uint8) to match
		// our internal use when we build the fundamental plane image ourselves. We do this
		// so that we can add (overlay) any ellipses defined in the json file. We expect
		// that external image files are used only to define odd or polygon shapes.
		if img.ColorModel() != color.GrayModel {
			fmt.Println(fmt.Errorf("\n\tThe supplied external image %q is not type GRAY.", event.PathToExternalImage))
			os.Exit(8)
		}

		event.FplaneImage = img.(*image.Gray)

		// Override the value (possibly) supplied in the fundamental_plane_width_num_points parameter
		event.FundamentalPlaneWidthPoints = img.Bounds().Dx()
		fmt.Println(ColorModelString(img.ColorModel()))
	} else { // No image supplied by user, so we start our own.
		event.FplaneImage = image.NewGray(image.Rect(0, 0, event.FundamentalPlaneWidthPoints, event.FundamentalPlaneWidthPoints))
		FillFplane(event.FplaneImage, true)
	}

	AddEllipses(event, true)
	err = SaveGrayPNG("geometricShadow.png", event.FplaneImage)
	if err != nil {
		fmt.Println(fmt.Errorf("\n\tFailed to write %q.", "geometricShadow.png"))
		os.Exit(9)
	}

	sourcePlane := ConvertSourcePlaneImageToComplex(event.FplaneImage)

	elapsed := time.Since(start)
	fmt.Printf("Generation of the geometric shadow took %s\n", elapsed)

	// If a user gave us distance in arcseconds, it is given priority, and
	// we overwrite any value that may also have been given in AU.
	if event.ParallaxArcsec > 0.0 {
		event.DistanceAu = 8.79414 / event.ParallaxArcsec
	}

	auToKm := 1.495979e+8
	nmToKm := 1e-9 * 1e-3

	WavelengthKm := event.ObservationWavelengthNm * nmToKm
	Lkm := event.FundamentalPlaneWidthKm
	Zkm := event.DistanceAu * auToKm
	//Npts := event.FundamentalPlaneWidthPoints

	// Some elementary checks to make sure that the user has not supplied bad parameters
	if Lkm <= 0.0 {
		fmt.Println(fmt.Errorf("\n\tFundamental plane width must be positive."))
		os.Exit(10)
	}

	if Zkm <= 0.0 {
		fmt.Println(fmt.Errorf("\n\tDistance given is invalid."))
		os.Exit(10)
	}

	event.StarDiamKm = 1.496e8 * event.DistanceAu * event.StarDiamMas / (1000.0 * 206265)

	start = time.Now()
	eField := FullObservationPlaneSincSolution(Lkm, Zkm, WavelengthKm, sourcePlane)
	elapsed = time.Since(start)
	fmt.Printf("Calculation of the observation e-field took %s\n", elapsed)

	start = time.Now()
	// incidentWave is used to convert the aperture image to an occulter image using Babinet's formula
	incidentWave := complex(1.0, 0.0)

	intensity := make([]float64, len(eField))
	for i := 0; i < len(eField); i++ {
		intensity[i] = real(incidentWave-eField[i])*real(incidentWave-eField[i]) +
			imag(incidentWave-eField[i])*imag(incidentWave-eField[i])
	}

	intensityMatrix, err := Reshape1DTo2D(intensity, Npts, Npts)
	if err != nil {
		fmt.Println(fmt.Errorf("reshape of intensity vector failed: %w", err))
		os.Exit(10)
	}
	elapsed = time.Since(start)
	fmt.Printf("Calculation of the observation intensity took %s\n", elapsed)

	imgForDisplay, err := MatrixToGrayViewPercentile(intensityMatrix, 0.0, 100)
	if err != nil {
		fmt.Println(fmt.Errorf("creation of the display image failed: %w", err))
		os.Exit(11)
	}

	err = SaveGrayPNG("diffractionImage8bit.png", imgForDisplay)
	if err != nil {
		fmt.Println(fmt.Errorf("writing of %q failed: %w", "diffractionImage8bit.png", err))
		os.Exit(12)
	}

	occultImage, err := MatrixToGray16Data(intensityMatrix, 4000)
	if err != nil {
		fmt.Println(fmt.Errorf("creation of occultImage failed: %w", err))
		os.Exit(13)
	}
	err = SaveGray16PNG("occultImage16bit.png", occultImage)
	if err != nil {
		fmt.Println(fmt.Errorf("writing of %q failed: %w", "occultImage16bit.png", err))
		os.Exit(14)
	}

	if event.StarDiamKm > 0.0 {
		fmt.Printf("\nStar diameter projected at the plane of the asteroid is %0.3f km\n\n", event.StarDiamKm)
		starImage, sumOfWeights := BuildStarPsf(event.StarDiamKm, resolution, event.LimbDarkeningCoeff)
		//fmt.Printf("Sum of weights in star image is %0.6f\n", sumOfWeights)
		if false { // debug output for use during development
			imgForDisplay, err = MatrixToGrayViewPercentile(starImage, 0.0, 100)
			if err != nil {
				fmt.Println(fmt.Errorf("creation of the display image failed: %w", err))
				os.Exit(11)
			}
			err = SaveGrayPNG("diffractionImage8bit.png", imgForDisplay)
			if err != nil {
				fmt.Println(fmt.Errorf("writing of %q failed: %w", "diffractionImage8bit.png", err))
				os.Exit(12)
			}
		}
		start := time.Now()
		newImage, err := ConvolvePSFFFT(intensityMatrix, starImage, sumOfWeights, ConvSame, PadReplicate, false)
		elapsed := time.Since(start)
		fmt.Printf("Convolution of intensity matrix with star image took %s\n", elapsed)
		if err != nil {
			fmt.Println(fmt.Errorf("convolution of intensity matrix with star image failed: %w", err))
			os.Exit(13)
		}

		imgForDisplay, err := MatrixToGrayViewPercentile(newImage, 0.0, 100)
		if err != nil {
			fmt.Println(fmt.Errorf("creation of the display image failed: %w", err))
			os.Exit(11)
		}
		err = SaveGrayPNG("diffractionImage8bit.png", imgForDisplay)
		if err != nil {
			fmt.Println(fmt.Errorf("writing of %q failed: %w", "diffractionImage8bit.png", err))
			os.Exit(12)
		}
	}
	elapsed = time.Since(programStart)
	fmt.Printf("\nTotal program run time is %s\n", elapsed)

	if event.WindowSizePixels > 0 {
		size := event.WindowSizePixels
		w.SetTitle(event.Title)
		w.CenterOnScreen()
		img := canvas.NewImageFromFile("diffractionImage8bit.png")
		img.FillMode = canvas.ImageFillContain
		w.SetContent(container.NewStack(img))
		w.Resize(fyne.Size{Height: float32(size), Width: float32(size)})

		w.ShowAndRun()
	}
}

func FresnelScale(wavelengthNm, ZAu float64) float64 {
	// Unit	conversions.
	auToKm := 1.495979e+8 // Convert distance expressed in AU to km
	nmToKm := 1e-9 * 1e-3 // Convert nm to km
	wavelengthKm := wavelengthNm * nmToKm
	ZKm := ZAu * auToKm
	return math.Sqrt(wavelengthKm * ZKm / 2)
}
