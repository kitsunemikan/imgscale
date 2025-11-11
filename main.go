package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/disintegration/gift"

	_ "image/gif"

	_ "image/jpeg"

	_ "image/png"
)

var (
	inFilename    = flag.String("i", "", "input file")
	outFilename   = flag.String("o", "", "output file")
	scale         = flag.Float64("s", 1.0, "uniform scale")
	scalex        = flag.Float64("sx", 0.0, "horizontal scale, overrides -s")
	scaley        = flag.Float64("sy", 0.0, "vertical scale, overrides -s")
	ow            = flag.Int("ow", 0, "override output width, overrides -s")
	oh            = flag.Int("oh", 0, "override output height, overrides -s")
	maxside       = flag.Int("maxside", 0, "autocalculate uniform scale factor so that the largest dimension matches this value, 0 to disable, overrides -s")
	force         = flag.Bool("f", false, "overwrite if output file exists")
	quality       = flag.Int("q", 80, "output JPEG quality")
	resamplingStr = flag.String("r", "lanczos", fmt.Sprintf("resampling function, one of %s", formatKeys(resamplingMap)))
)

var resamplingMap = map[string]gift.Resampling{
	"nearest": gift.NearestNeighborResampling,
	"linear":  gift.LinearResampling,
	"cubic":   gift.CubicResampling,
	"box":     gift.BoxResampling,
	"lanczos": gift.LanczosResampling,
}

const (
	FormatUnknown Format = iota
	FormatPNG
	FormatJPEG
)

const (
	ErrNoExtension StringError = "no filename extension"
)

type Format uint8

type StringError string

func (e StringError) Error() string {
	return string(e)
}

func formatKeys[K comparable, V any](m map[K]V) string {
	if len(m) == 0 {
		return "[] (no values avaliable)"
	}

	var str strings.Builder
	str.WriteByte('[')

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, fmt.Sprint(k))
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	str.WriteString(keys[0])

	for _, k := range keys[1:] {
		str.WriteString(", ")
		str.WriteString(k)
	}

	str.WriteByte(']')

	return str.String()
}

func fatalf(format string, args ...any) {
	format = fmt.Sprintf("error: %s\n", format)
	fmt.Printf(format, args...)
	os.Exit(1)
}

func detectEncodeFormat(filename string) (Format, error) {
	ext := filepath.Ext(filename)
	if ext == "" {
		return FormatUnknown, ErrNoExtension
	}

	extUpper := strings.ToUpper(ext)

	switch extUpper {
	case ".PNG":
		return FormatPNG, nil
	case ".JPG", ".JPEG":
		return FormatJPEG, nil
	default:
		return FormatUnknown, fmt.Errorf("unknown or unsupported image extension '%s'", ext)
	}
}

func main() {
	flag.Parse()

	if *inFilename == "" {
		fatalf("missing input filename")
	}

	if *outFilename == "" {
		fatalf("missing output filename")
	}

	resampling := resamplingMap[*resamplingStr]
	if resampling == nil {
		fatalf("unknown resampling function '%s'", *resamplingStr)
	}

	encodeFmt, err := detectEncodeFormat(*outFilename)
	if err != nil {
		fatalf("coudn't detect output format: %v", err)
	}

	modeCount := 0
	if *scalex != 0 || *scaley != 0 {
		modeCount++
	}

	if *maxside != 0 {
		modeCount++
	}

	if *ow != 0 || *oh != 0 {
		modeCount++
	}

	if modeCount > 1 {
		fatalf("conflicting rescale options")
	}

	inFile, err := os.Open(*inFilename)
	if err != nil {
		fatalf("couldn't open input file: %v", err)
	}

	defer inFile.Close()

	outFile, err := os.Open(*outFilename)
	if err == nil && !*force {
		fatalf("output file already exists")
	}

	outFile.Close()

	outFile, err = os.Create(*outFilename)
	if err != nil {
		fatalf("couldn't open output file for write: %v", err)
	}

	defer outFile.Close()

	img, _, err := image.Decode(inFile)
	if err != nil {
		fatalf("couldn't decode input image '%s': %v", *inFilename, err)
	}

	inW, inH := img.Bounds().Dx(), img.Bounds().Dy()

	sx, sy := *scale, *scale
	if *scalex != 0 {
		sx = *scalex
	}

	if *scaley != 0 {
		sy = *scaley
	}

	if *maxside != 0 {
		maxD := max(inW, inH)
		sx = float64(*maxside) / float64(maxD)
		sy = sx
	}

	if *ow != 0 {
		sx = float64(*ow) / float64(inW)
	}

	if *oh != 0 {
		sy = float64(*oh) / float64(inH)
	}

	outW, outH := int(float64(inW)*sx), int(float64(inH)*sy)

	fmt.Printf("output image dimensions: %dx%d\n", outW, outH)

	filter := gift.New(
		gift.Resize(outW, outH, resampling),
	)

	outImg := image.NewRGBA(filter.Bounds(img.Bounds()))
	filter.Draw(outImg, img)

	switch encodeFmt {
	case FormatPNG:
		err = png.Encode(outFile, outImg)
	case FormatJPEG:
		err = jpeg.Encode(outFile, outImg, &jpeg.Options{
			Quality: *quality,
		})
	}

	if err != nil {
		fatalf("couldn't encode output image '%s': %v", *outFilename, err)
	}
}
