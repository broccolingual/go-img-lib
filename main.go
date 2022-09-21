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

type Img struct {
	Src  image.Image
	H, W int
}

type TimeData struct {
	Elapsed int64
	Name    string
}

func LoadImage(path string) image.Image {
	f, _ := os.Open(path)
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		log.Println(err)
	}

	return src
}

func SaveImage(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	png.Encode(f, img)
}

func Gray(src image.Image) *image.Gray {
	bounds := src.Bounds()
	dst := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := color.GrayModel.Convert(src.At(x, y)).(color.Gray)
			dst.Set(x, y, gray)
		}
	}
	return dst
}

func Binarization(src *image.Gray, threshold uint8) *image.Gray {
	bounds := src.Bounds()
	dst := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := src.GrayAt(x, y)
			if gray.Y > threshold {
				gray.Y = 255
			} else {
				gray.Y = 0
			}
			dst.Set(x, y, gray)
		}
	}
	return dst
}

// TODO:
func ParallelFilter(src *image.Gray, mask [][]int16, divider int16, maskSize int) *image.Gray {
	bounds := src.Bounds()
	dst := image.NewGray(bounds)
	margin := maskSize / 2
	for y := bounds.Min.Y + margin; y < bounds.Max.Y-margin; y++ {
		for x := bounds.Min.X + margin; x < bounds.Max.X-margin; x++ {
			gray := color.Gray{}
			pixelValue := int16(0)
			for i := -margin; i <= margin; i++ {
				for j := -margin; j <= margin; j++ {
					pixelValue += int16(src.GrayAt(x+i, y+j).Y) * mask[1+j][1+i]
				}
			}
			gray.Y = uint8(pixelValue / divider)
			dst.Set(x, y, gray)
		}
	}
	return dst
}

func FilterGray(src *image.Gray, mask [][]int16, divider int16, maskSize int) *image.Gray {
	bounds := src.Bounds()
	dst := image.NewGray(bounds)
	margin := maskSize / 2
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := color.Gray{}
			pixelValue := int16(0)
			for i := -margin; i <= margin; i++ {
				for j := -margin; j <= margin; j++ {
					pixelValue += int16(src.GrayAt(x+i, y+j).Y) * mask[1+j][1+i]
				}
			}
			gray.Y = uint8(pixelValue / divider)
			dst.Set(x, y, gray)
		}
	}
	return dst
}

func filterRGBA(src image.Image, mask [][]int16, divider int16, maskSize int) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	margin := maskSize / 2
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba := color.RGBA{}
			pixelR, pixelG, pixelB, pixelA := int16(0), int16(0), int16(0), int16(0)
			for i := -margin; i <= margin; i++ {
				for j := -margin; j <= margin; j++ {
					c := color.RGBAModel.Convert(src.At(x+i, y+j)).(color.RGBA)
					pixelR += int16(c.R) * mask[1+j][1+i]
					pixelG += int16(c.G) * mask[1+j][1+i]
					pixelB += int16(c.B) * mask[1+j][1+i]
					pixelA += int16(c.A) * mask[1+j][1+i]
				}
			}
			rgba.R = uint8(pixelR / divider)
			rgba.G = uint8(pixelG / divider)
			rgba.B = uint8(pixelB / divider)
			rgba.A = uint8(pixelA / divider)
			dst.Set(x, y, rgba)
		}
	}
	return dst
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

func main() {
	path := "img/lenna.png"
	img := LoadImage(path)
	timeData := []TimeData{}
	iteration := 10
	fmt.Printf("%s (%dx%d)\nAverage Elapsed Time(Iteration: %d)\n", path, img.Bounds().Dx(), img.Bounds().Dy(), iteration)

	// gray
	now := time.Now()
	for i := 0; i < iteration; i++ {
		Gray(img)
	}
	timeData = append(timeData, TimeData{Name: "Gray", Elapsed: int64(time.Since(now).Milliseconds() / int64(iteration))})
	imgGray := Gray(img)
	SaveImage("img/gray.png", imgGray)

	// binarization
	now = time.Now()
	for i := 0; i < iteration; i++ {
		Binarization(imgGray, 128)
	}
	timeData = append(timeData, TimeData{Name: "Binalization", Elapsed: int64(time.Since(now).Milliseconds() / int64(iteration))})
	imgBinarization := Binarization(imgGray, 128)
	SaveImage("img/binarization.png", imgBinarization)

	// simple
	simpleMask := [][]int16{
		{1, 1, 1},
		{1, 1, 1},
		{1, 1, 1},
	}
	now = time.Now()
	for i := 0; i < iteration; i++ {
		FilterGray(imgGray, simpleMask, 9, 3)
	}
	timeData = append(timeData, TimeData{Name: "Simple(Gray)", Elapsed: int64(time.Since(now).Milliseconds() / int64(iteration))})
	imgSimple := FilterGray(imgGray, simpleMask, 9, 3)
	SaveImage("img/simple.png", imgSimple)

	// rgba filter
	now = time.Now()
	for i := 0; i < iteration; i++ {
		filterRGBA(img, simpleMask, 9, 3)
	}
	timeData = append(timeData, TimeData{Name: "Simple(RGBA)", Elapsed: int64(time.Since(now).Milliseconds() / int64(iteration))})
	SaveImage("img/simpleRgba.png", filterRGBA(img, simpleMask, 9, 3))

	// gaussian
	gaussianMask := [][]int16{
		{1, 2, 1},
		{2, 4, 2},
		{1, 2, 1},
	}
	now = time.Now()
	for i := 0; i < iteration; i++ {
		FilterGray(imgGray, gaussianMask, 16, 3)
	}
	timeData = append(timeData, TimeData{Name: "Gaussian(Gray)", Elapsed: int64(time.Since(now).Milliseconds() / int64(iteration))})
	imgGaussian := FilterGray(imgGray, gaussianMask, 16, 3)
	SaveImage("img/gaussian.png", imgGaussian)

	// sharpening
	sharpeningMask := [][]int16{
		{0, -1, 0},
		{-1, 4, -1},
		{0, -1, 0},
	}
	now = time.Now()
	for i := 0; i < iteration; i++ {
		FilterGray(imgBinarization, sharpeningMask, 9, 3)
	}
	timeData = append(timeData, TimeData{Name: "Sharpening(Gray)", Elapsed: int64(time.Since(now).Milliseconds() / int64(iteration))})
	SaveImage("img/sharpening.png", FilterGray(imgBinarization, sharpeningMask, 9, 3))

	// gray - gaussian + 128
	SaveImage("img/subGaussian.png", SubPixel(imgGray, imgGaussian))

	// gray - simple + 128
	SaveImage("img/subSimple.png", SubPixel(imgGray, imgSimple))

	for _, td := range timeData {
		fmt.Printf("%-16s: %vms\n", td.Name, td.Elapsed)
	}
}
