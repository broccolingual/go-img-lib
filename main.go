package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"os"
	"time"
)

type arrRGBAImg [][]color.RGBA
type arrGrayImg [][]color.Gray

type TimeData struct {
	Elapsed int64
	Name    string
}

func LoadRGBAImage(path string) *image.RGBA {
	f, _ := os.Open(path)
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		log.Println(err)
	}

	return src.(*image.RGBA)
}

func SaveImage(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	png.Encode(f, img)
}

func ConvertArray(src *image.RGBA) (array arrRGBAImg) {
	size := src.Bounds().Size()
	for i := 0; i < size.X; i++ {
		var y []color.RGBA
		for j := 0; j < size.Y; j++ {
			y = append(y, src.At(i, j).(color.RGBA))
		}
		array = append(array, y)
	}
	return
}

func ConvertGrayArray(src *image.Gray) (array arrGrayImg) {
	size := src.Bounds().Size()
	for i := 0; i < size.X; i++ {
		var y []color.Gray
		for j := 0; j < size.Y; j++ {
			y = append(y, src.At(i, j).(color.Gray))
		}
		array = append(array, y)
	}
	return
}

func ConvertGrayImage(array arrGrayImg) *image.Gray {
	xlen, ylen := array.GetSize()
	rect := image.Rect(0, 0, xlen, ylen)
	dst := image.NewGray(rect)
	for x := 0; x < xlen; x++ {
		for y := 0; y < ylen; y++ {
			dst.Set(x, y, array[x][y])
		}
	}
	return dst
}

func ConvertRGBAImage(array arrRGBAImg) *image.RGBA {
	xlen, ylen := array.GetSize()
	rect := image.Rect(0, 0, xlen, ylen)
	dst := image.NewRGBA(rect)
	for x := 0; x < xlen; x++ {
		for y := 0; y < ylen; y++ {
			dst.Set(x, y, array[x][y])
		}
	}
	return dst
}

func AllocGrayArray(x, y int) (dst arrGrayImg) {
	dst = make(arrGrayImg, x)
	for i := 0; i < len(dst); i++ {
		dst[i] = make([]color.Gray, y)
	}
	return
}

func AllocRGBAArray(x, y int) (dst arrRGBAImg) {
	dst = make(arrRGBAImg, x)
	for i := 0; i < len(dst); i++ {
		dst[i] = make([]color.RGBA, y)
	}
	return
}

func (src arrGrayImg) GetSize() (x int, y int) {
	x, y = len(src), len(src[0])
	return
}

func (src arrRGBAImg) GetSize() (x int, y int) {
	x, y = len(src), len(src[0])
	return
}

func (src arrGrayImg) FlipHorizontal() (dst arrGrayImg) {
	xlen, ylen := src.GetSize()
	dst = AllocGrayArray(xlen, ylen)
	for x := 0; x < xlen; x++ {
		for y := 0; y < ylen; y++ {
			dst[x][y] = color.Gray{uint8(src[xlen-x-1][y].Y)}
		}
	}
	return
}

func (src arrGrayImg) FlipVertical() (dst arrGrayImg) {
	xlen, ylen := src.GetSize()
	dst = AllocGrayArray(xlen, ylen)
	for x := 0; x < xlen; x++ {
		for y := 0; y < ylen; y++ {
			dst[x][y] = color.Gray{uint8(src[x][ylen-y-1].Y)}
		}
	}
	return
}

func (src arrRGBAImg) ToGrayscale() (dst arrGrayImg) {
	xlen, ylen := src.GetSize()
	dst = AllocGrayArray(xlen, ylen)
	for x := 0; x < xlen; x++ {
		for y := 0; y < ylen; y++ {
			pix := src[x][y]
			gray := uint8((float64(pix.R) + float64(pix.G) + float64(pix.B)) / 3.0)
			dst[x][y] = color.Gray{gray}
		}
	}
	return
}

func (src arrGrayImg) ToBinarize(threshold uint8) (dst arrGrayImg) {
	xlen, ylen := src.GetSize()
	dst = AllocGrayArray(xlen, ylen)
	for x := 0; x < xlen; x++ {
		for y := 0; y < ylen; y++ {
			gray := uint8(src[x][y].Y)
			if gray > threshold {
				gray = 255
			} else {
				gray = 0
			}
			dst[x][y] = color.Gray{gray}
		}
	}
	return
}

func (src arrGrayImg) Filter(mask [][]float64) (dst arrGrayImg) {
	xlen, ylen := src.GetSize()
	dst = AllocGrayArray(xlen, ylen)
	mxlen, mylen := len(mask), len(mask[0])
	if mxlen != mylen {
		return nil
	}
	margin := mxlen / 2
	for x := 0; x < xlen; x++ {
		for y := 0; y < ylen; y++ {
			pValue := 0.0
			for i := -margin; i <= margin; i++ {
				for j := -margin; j <= margin; j++ {
					if x+i < 0 || x+i >= xlen || y+j < 0 || y+j >= ylen {
						pValue += float64(0) * mask[margin+i][margin+j]
					} else {
						pValue += float64(src[x+i][y+j].Y) * mask[margin+i][margin+j]
					}
				}
			}
			dst[x][y] = color.Gray{uint8(pValue)}
		}
	}
	return
}

func SubPixel(src1 *image.Gray, src2 *image.Gray) *image.Gray {
	bounds := src1.Bounds()
	dst := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := color.Gray{}
			gray.Y = uint8(src1.GrayAt(x, y).Y - src2.GrayAt(x, y).Y + 128)
			dst.Set(x, y, gray)
		}
	}
	return dst
}

func (src arrGrayImg) ImageProc(threshold int) (dst arrGrayImg) {
	xlen, ylen := src.GetSize()
	dst = AllocGrayArray(xlen, ylen)
	for x := 0; x < xlen-1; x++ {
		for y := 0; y < ylen; y++ {
			pValue := int(src[x][y].Y) - int(src[x+1][y].Y)
			gray := uint8(0)
			if pValue > threshold {
				gray = 255
			} else if pValue < -threshold {
				gray = 128
			}
			dst[x][y] = color.Gray{gray}
		}
	}
	return
}

func main() {
	path := "img/lenna.png"
	img := LoadRGBAImage(path)
	timeData := []TimeData{}
	iteration := 1
	fmt.Printf("%s (%dx%d)\nAverage Elapsed Time(Iteration: %d)\n", path, img.Bounds().Dx(), img.Bounds().Dy(), iteration)

	// grayscale
	now := time.Now()
	for i := 0; i < iteration; i++ {
		ConvertGrayImage(ConvertArray(img).ToGrayscale())
	}
	timeData = append(timeData, TimeData{Name: "Gray", Elapsed: int64(time.Since(now).Milliseconds() / int64(iteration))})
	grayArray := ConvertArray(img).ToGrayscale()
	imgGray := ConvertGrayImage(grayArray)
	SaveImage("img/gray.png", imgGray)
	SaveImage("img/grayFlipH.png", ConvertGrayImage(grayArray.FlipHorizontal()))
	SaveImage("img/grayFlipV.png", ConvertGrayImage(grayArray.FlipVertical()))

	// binarize
	now = time.Now()
	for i := 0; i < iteration; i++ {
		ConvertGrayImage(grayArray.ToBinarize(128))
	}
	timeData = append(timeData, TimeData{Name: "Binalize", Elapsed: int64(time.Since(now).Milliseconds() / int64(iteration))})
	binArray := grayArray.ToBinarize(128)
	imgBinarization := ConvertGrayImage(binArray)
	SaveImage("img/binarize.png", imgBinarization)

	// simple (3x3)
	simpleMask3 := [][]float64{
		{1.0 / 9, 1.0 / 9, 1.0 / 9},
		{1.0 / 9, 1.0 / 9, 1.0 / 9},
		{1.0 / 9, 1.0 / 9, 1.0 / 9},
	}
	now = time.Now()
	for i := 0; i < iteration; i++ {
		ConvertGrayImage(grayArray.Filter(simpleMask3))
	}
	timeData = append(timeData, TimeData{Name: "Simple3", Elapsed: int64(time.Since(now).Milliseconds() / int64(iteration))})
	simple3Array := grayArray.Filter(simpleMask3)
	imgSimple3 := ConvertGrayImage(simple3Array)
	SaveImage("img/simple3.png", imgSimple3)

	// simple (5x5)
	simpleMask5 := [][]float64{
		{1.0 / 25, 1.0 / 25, 1.0 / 25, 1.0 / 25, 1.0 / 25},
		{1.0 / 25, 1.0 / 25, 1.0 / 25, 1.0 / 25, 1.0 / 25},
		{1.0 / 25, 1.0 / 25, 1.0 / 25, 1.0 / 25, 1.0 / 25},
		{1.0 / 25, 1.0 / 25, 1.0 / 25, 1.0 / 25, 1.0 / 25},
		{1.0 / 25, 1.0 / 25, 1.0 / 25, 1.0 / 25, 1.0 / 25},
	}
	now = time.Now()
	for i := 0; i < iteration; i++ {
		ConvertGrayImage(grayArray.Filter(simpleMask5))
	}
	timeData = append(timeData, TimeData{Name: "Simple5", Elapsed: int64(time.Since(now).Milliseconds() / int64(iteration))})
	simple5Array := grayArray.Filter(simpleMask5)
	imgSimple5 := ConvertGrayImage(simple5Array)
	SaveImage("img/simple5.png", imgSimple5)

	// gaussian
	gaussianMask := [][]float64{
		{1.0 / 16, 2.0 / 16, 1.0 / 16},
		{2.0 / 16, 4.0 / 16, 2.0 / 16},
		{1.0 / 16, 2.0 / 16, 1.0 / 16},
	}
	now = time.Now()
	for i := 0; i < iteration; i++ {
		ConvertGrayImage(grayArray.Filter(gaussianMask))
	}
	timeData = append(timeData, TimeData{Name: "Gaussian", Elapsed: int64(time.Since(now).Milliseconds() / int64(iteration))})
	gaussianArray := grayArray.Filter(gaussianMask)
	imgGaussian := ConvertGrayImage(gaussianArray)
	SaveImage("img/gaussian.png", imgGaussian)

	// sharpening
	sharpeningMask := [][]float64{
		{0.0, 1.0, 0.0},
		{1.0, -4.0, 1.0},
		{0.0, 1.0, 0.0},
	}
	now = time.Now()
	for i := 0; i < iteration; i++ {
		ConvertGrayImage(binArray.Filter(sharpeningMask))
	}
	timeData = append(timeData, TimeData{Name: "Sharpening", Elapsed: int64(time.Since(now).Milliseconds() / int64(iteration))})
	sharpeningArray := binArray.Filter(sharpeningMask)
	imgSharpening := ConvertGrayImage(sharpeningArray)
	SaveImage("img/sharpening.png", imgSharpening)

	// gray - gaussian + 128
	SaveImage("img/subGaussian.png", SubPixel(imgGray, imgGaussian))

	// gray - simple3 + 128
	SaveImage("img/subSimple3.png", SubPixel(imgGray, imgSimple3))

	// gray - simple5 + 128
	SaveImage("img/subSimple5.png", SubPixel(imgGray, imgSimple5))

	// test
	imgOut := ConvertGrayArray(SubPixel(imgGray, imgGaussian))
	SaveImage("img/test.png", ConvertGrayImage(imgOut.ImageProc(24)))

	for _, td := range timeData {
		fmt.Printf("%-16s: %vms\n", td.Name, td.Elapsed)
	}
}
