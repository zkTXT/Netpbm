package Netpbm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// PBM struct represents a PBM image.
type PBM struct {
	data          [][]bool
	width, height int
	magicNumber   string
}

// ReadPBM reads the PBM image from a file and returns the image information in a struct.
func ReadPBM(filename string) (*PBM, error) {
	pbm := PBM{}

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	confirmMagicNumber := false
	confirmDimensions := false
	line := 0

	// Ignore all comments
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}

		if strings.HasPrefix(scanner.Text(), "#") {
			continue
		} else if !confirmMagicNumber {
			//Magic number
			pbm.magicNumber = scanner.Text()
			confirmMagicNumber = true
		} else if !confirmDimensions {
			//Size
			separated := strings.Fields(scanner.Text())
			if len(separated) > 0 {
				pbm.width, _ = strconv.Atoi(separated[0])
				pbm.height, _ = strconv.Atoi(separated[1])
			}
			confirmDimensions = true
			pbm.data = make([][]bool, pbm.height)
			for i := range pbm.data {
				pbm.data[i] = make([]bool, pbm.width)
			}

		} else {
			if pbm.magicNumber == "P1" {
				//P1 format
				parsedValues := strings.Fields(scanner.Text())
				for i := 0; i < pbm.width; i++ {
					pbm.data[line][i] = parsedValues[i] == "1"
				}
				line++
			} else if pbm.magicNumber == "P4" {
				//P4 format
				err := processP4Format(file, &pbm)
				if err != nil {
					return nil, fmt.Errorf("error processing P4 format: %v", err)
				}
			}
		}
	}
	return &pbm, nil
}

func processP4Format(file *os.File, pbm *PBM) error {
	expectedBytesPerRow := (pbm.width + 7) / 8
	totalExpectedBytes := expectedBytesPerRow * pbm.height
	fmt.Printf("Expected total bytes for pixel data: %d\n", totalExpectedBytes)
	allPixelData := make([]byte, totalExpectedBytes)
	fileContent, err := os.ReadFile(file.Name())
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}
	copy(allPixelData, fileContent[len(fileContent)-totalExpectedBytes:])
	byteIndex := 0
	for y := 0; y < pbm.height; y++ {
		for x := 0; x < pbm.width; x++ {
			if x%8 == 0 && x != 0 {
				byteIndex++
			}
			pbm.data[y][x] = (allPixelData[byteIndex]>>(7-(x%8)))&1 != 0
		}
		byteIndex++
	}
	return nil
}

// Size returns the height and width of the image.
func (pbm *PBM) Size() (int, int) {
	return pbm.height, pbm.width
}

// At returns the value of each pixel at (x, y).
func (pbm *PBM) At(x, y int) bool {
	return pbm.data[x][y]
}

// Set sets the value of each pixel at (x, y).
func (pbm *PBM) Set(x, y int, value bool) {
	pbm.data[x][y] = value
}

// Save saves the PBM image to a file and returns an error if there was a problem.
func (pbm *PBM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()
	_, err = fmt.Fprintf(file, "%s\n%d %d\n", pbm.magicNumber, pbm.width, pbm.height)
	if err != nil {
		return fmt.Errorf("error writing magic number and dimensions: %v", err)
	}
	if pbm.magicNumber == "P1" {
		err := writeP1Format(file, pbm)
		if err != nil {
			return fmt.Errorf("error writing P1 format data: %v", err)
		}
	} else if pbm.magicNumber == "P4" {
		err := writeP4Format(file, pbm)
		if err != nil {
			return fmt.Errorf("error writing P4 format data: %v", err)
		}
	}

	return nil
}

func writeP1Format(file *os.File, pbm *PBM) error {
	for _, row := range pbm.data {
		for _, pixel := range row {
			if pixel {
				_, err := file.WriteString("1 ")
				if err != nil {
					return fmt.Errorf("error writing pixel data: %v", err)
				}
			} else {
				_, err := file.WriteString("0 ")
				if err != nil {
					return fmt.Errorf("error writing pixel data: %v", err)
				}
			}
		}
		_, err := file.WriteString("\n")
		if err != nil {
			return fmt.Errorf("error writing pixel data: %v", err)
		}
	}
	return nil
}

func writeP4Format(file *os.File, pbm *PBM) error {
	for _, row := range pbm.data {
		for x := 0; x < pbm.width; x += 8 {
			var byteValue byte
			for i := 0; i < 8 && x+i < pbm.width; i++ {
				bitIndex := 7 - i
				if row[x+i] {
					byteValue |= 1 << bitIndex
				}
			}
			_, err := file.Write([]byte{byteValue})
			if err != nil {
				return fmt.Errorf("error writing pixel data: %v", err)
			}
		}
	}
	return nil
}

// Invert inverts the colors of the PBM image.
func (pbm *PBM) Invert() {
	for y := 0; y < pbm.height; y++ {
		for x := 0; x < pbm.width; x++ {
			pbm.data[y][x] = !pbm.data[y][x]
		}
	}
}

// Flip flips the PBM image horizontally.
func (pbm *PBM) Flip() {
	for y := 0; y < pbm.height; y++ {
		start := make([]bool, pbm.width)
		end := make([]bool, pbm.width)
		copy(start, pbm.data[y])
		for i := 0; i < len(start); i++ {
			end[i] = start[len(start)-1-i]
		}
		copy(pbm.data[y], end[:])
	}
}

// Flop flops the PBM image vertically.
func (pbm *PBM) Flop() {
	cursor := pbm.height - 1
	for y := range pbm.data {
		temp := pbm.data[y]
		pbm.data[y] = pbm.data[cursor]
		pbm.data[cursor] = temp
		cursor--
		if cursor < y || cursor == y {
			break
		}
	}
}

// SetMagicNumber sets the magic number of the PBM image.
func (pbm *PBM) SetMagicNumber(magicNumber string) {
	pbm.magicNumber = magicNumber
}
