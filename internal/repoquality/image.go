package repoquality

import (
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"image/png"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

func assetDimensions(path string) (int, int, error) {
	extension := strings.ToLower(filepath.Ext(path))
	switch extension {
	case ".svg":
		data, err := readFileLimited(path, maxPolicyFileBytes)
		if err != nil {
			return 0, 0, err
		}
		var root struct {
			XMLName xml.Name `xml:"svg"`
			Width   string   `xml:"width,attr"`
			Height  string   `xml:"height,attr"`
		}
		if err := xml.Unmarshal(data, &root); err != nil {
			return 0, 0, fmt.Errorf("invalid SVG document %s: %w", path, err)
		}
		width, widthErr := strconv.Atoi(root.Width)
		height, heightErr := strconv.Atoi(root.Height)
		if widthErr != nil || heightErr != nil || width <= 0 || height <= 0 {
			return 0, 0, fmt.Errorf("SVG has no integer width and height: %s", path)
		}
		return width, height, nil
	case ".png":
		file, err := openRegularFile(path, maxMediaBytes)
		if err != nil {
			return 0, 0, err
		}
		defer file.Close()
		config, err := png.DecodeConfig(io.LimitReader(file, maxMediaBytes+1))
		if err != nil {
			return 0, 0, fmt.Errorf("invalid PNG header %s: %w", path, err)
		}
		return config.Width, config.Height, nil
	case ".webp":
		return webpDimensions(path)
	default:
		return 0, 0, fmt.Errorf("unsupported media extension %s in %s", extension, path)
	}
}

func webpDimensions(path string) (int, int, error) {
	file, err := openRegularFile(path, maxMediaBytes)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return 0, 0, err
	}
	data, err := io.ReadAll(io.LimitReader(file, 30))
	if err != nil {
		return 0, 0, err
	}
	if len(data) < 20 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
		return 0, 0, fmt.Errorf("invalid WebP header: %s", path)
	}
	fileSize := int64(binary.LittleEndian.Uint32(data[4:8])) + 8
	chunkSize := int64(binary.LittleEndian.Uint32(data[16:20]))
	if fileSize != info.Size() || fileSize%2 != 0 || chunkSize+(chunkSize%2) > fileSize-20 {
		return 0, 0, fmt.Errorf("invalid WebP container size: %s", path)
	}
	chunk := string(data[12:16])
	switch chunk {
	case "VP8X":
		if len(data) < 30 || chunkSize != 10 {
			return 0, 0, fmt.Errorf("invalid extended WebP header: %s", path)
		}
		width := uint24(data[24:27]) + 1
		height := uint24(data[27:30]) + 1
		return width, height, nil
	case "VP8 ":
		if len(data) < 30 || chunkSize < 10 || data[20]&1 != 0 || data[23] != 0x9d || data[24] != 0x01 || data[25] != 0x2a {
			return 0, 0, fmt.Errorf("invalid lossy WebP frame header: %s", path)
		}
		width := int(binary.LittleEndian.Uint16(data[26:28])) & 0x3fff
		height := int(binary.LittleEndian.Uint16(data[28:30])) & 0x3fff
		if width == 0 || height == 0 {
			return 0, 0, fmt.Errorf("invalid lossy WebP dimensions: %s", path)
		}
		return width, height, nil
	case "VP8L":
		if len(data) < 25 || chunkSize < 5 || data[20] != 0x2f || data[24]&0xe0 != 0 {
			return 0, 0, fmt.Errorf("invalid lossless WebP frame header: %s", path)
		}
		width := 1 + int(data[21]) + int(data[22]&0x3f)<<8
		height := 1 + int(data[22]&0xc0)>>6 + int(data[23])<<2 + int(data[24]&0x0f)<<10
		return width, height, nil
	default:
		return 0, 0, fmt.Errorf("unsupported WebP chunk %s in %s", chunk, path)
	}
}

func uint24(data []byte) int {
	return int(data[0]) | int(data[1])<<8 | int(data[2])<<16
}
