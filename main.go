package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"image"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	cacheDir        = ".cache"
	imageDir        = "images"
	transformations = map[string]func(image.Image, string) (image.Image, error){
		"blur":       imageEffect(imaging.Blur),
		"sharpen":    imageEffect(imaging.Sharpen),
		"gamma":      imageEffect(imaging.AdjustGamma),
		"contrast":   imageEffect(imaging.AdjustContrast),
		"brightness": imageEffect(imaging.AdjustBrightness),
		"saturation": imageEffect(imaging.AdjustSaturation),
		"hue":        imageEffect(imaging.AdjustHue),
		"resize":     imageResize,
		"fit":        imageFit,
		"fill":       imageFill,
		"crop":       imageCrop,
		"grayscale":  imageGrayscale,
		"invert":     imageInvert,
	}
)

func init() {
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create cache directory: %v", err)
	}
	log.Println("Cache directory:", cacheDir)
}

func main() {
	serve()
}

func serve() {
	r := gin.Default()

	r.GET("/images/:operations/*filename", func(c *gin.Context) {
		operations := c.Param("operations")
		filename := c.Param("filename")[1:]

		cacheKey := generateCacheKey(filename, operations)
		imageCache := filepath.Join(cacheDir, cacheKey+".jpg")
		imagePath := filepath.Join(imageDir, filename)

		if _, err := os.Stat(imageCache); err == nil {
			c.File(imagePath)
			return
		}

		src, err := imaging.Open(imagePath)
		if err != nil {
			c.String(http.StatusNotFound, "Image not found")
			return
		}

		img, err := applyTransformations(src, operations)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		if err := imaging.Save(img, imageCache); err != nil {
			c.String(http.StatusInternalServerError, "Failed to save cached image")
			return
		}

		c.File(imageCache)
	})

	log.Fatal(r.Run(":80"))
}

func applyTransformations(img image.Image, operations string) (image.Image, error) {
	for _, op := range strings.Split(operations, ",") {
		parts := strings.SplitN(op, "=", 2)
		opName := parts[0]
		opParam := ""
		if len(parts) == 2 {
			opParam = parts[1]
		}
		if transformFunc, exists := transformations[opName]; exists {
			var err error
			img, err = transformFunc(img, opParam)
			if err != nil {
				return nil, fmt.Errorf("error applying %s: %v", opName, err)
			}
		}
	}
	return img, nil
}

func generateCacheKey(filename, operations string) string {
	hash := md5.Sum([]byte(filename + operations))
	return hex.EncodeToString(hash[:])
}

func imageEffect(effectFunc func(image.Image, float64) *image.NRGBA) func(image.Image, string) (image.Image, error) {
	return func(img image.Image, param string) (image.Image, error) {
		value, err := strconv.ParseFloat(param, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid parameter value")
		}
		return effectFunc(img, value), nil
	}
}

func imageCrop(img image.Image, param string) (image.Image, error) {
	parts := strings.Split(param, "@")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid crop parameters")
	}
	width, height, err := parseDimensions(parts[0])
	if err != nil {
		return nil, err
	}
	anchorPoint, err := parseAnchor(parts[1])
	if err != nil {
		return nil, err
	}
	return imaging.CropAnchor(img, width, height, anchorPoint), nil
}

func imageFill(img image.Image, param string) (image.Image, error) {
	parts := strings.Split(param, "@")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid crop parameters")
	}
	width, height, err := parseDimensions(parts[0])
	if err != nil {
		return nil, err
	}
	anchor, err := parseAnchor(parts[1])
	if err != nil {
		return nil, err
	}
	return imaging.Fill(img, width, height, anchor, imaging.Lanczos), nil
}

func imageFit(img image.Image, param string) (image.Image, error) {
	width, height, err := parseDimensions(param)
	if err != nil {
		return nil, err
	}
	return imaging.Fit(img, width, height, imaging.Lanczos), nil
}

func imageGrayscale(img image.Image, _ string) (image.Image, error) {
	return imaging.Grayscale(img), nil
}

func imageInvert(img image.Image, _ string) (image.Image, error) {
	return imaging.Invert(img), nil
}

func imageResize(img image.Image, param string) (image.Image, error) {
	width, height, err := parseDimensions(param)
	if err != nil {
		return nil, err
	}
	return imaging.Resize(img, width, height, imaging.Lanczos), nil
}

func parseAnchor(anchor string) (imaging.Anchor, error) {
	switch anchor {
	case "top-left":
		return imaging.TopLeft, nil
	case "top":
		return imaging.Top, nil
	case "top-right":
		return imaging.TopRight, nil
	case "left":
		return imaging.Left, nil
	case "center":
		return imaging.Center, nil
	case "right":
		return imaging.Right, nil
	case "bottom-left":
		return imaging.BottomLeft, nil
	case "bottom":
		return imaging.Bottom, nil
	case "bottom-right":
		return imaging.BottomRight, nil
	default:
		return 0, fmt.Errorf("invalid anchor point")
	}
}

func parseDimensions(dims string) (int, int, error) {
	parts := strings.Split(dims, "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid dimensions format")
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid width")
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid height")
	}

	return width, height, nil
}
