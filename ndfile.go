package ndfile

import (
	"encoding/binary"
	"io"
	"log"
	"os"
)

type NDFile struct {
	*NDFileHeader
	File         *os.File
	HeaderLength int64
}

func (f NDFile) Close() {
	f.File.Close()
}

func (f NDFile) GetData(latIndex, lngIndex int) ([]int16, error) {
	dataArrayLength := int((24 * 60) / f.TimeIntervalInMinutes)

	// Seek to the offset within the file:
	// - The position within the data matrix is determined by the latitude and longitude indices.
	// - Nx is the number of data points per latitude line.
	// - The product of (latIndex * Nx + lngIndex) gives the data point's linear index in a flat array.
	// - Each data point consists of `2 * forecastSteps` bytes (int16).
	// - The total offset in bytes from the beginning of the data section is then calculated by
	//   multiplying the linear index by the number of bytes per data point.
	if _, err := f.File.Seek(f.HeaderLength+int64((latIndex*int(f.NDFileHeader.Nx)+lngIndex)*(2*dataArrayLength)), io.SeekStart); err != nil {
		return nil, err
	}

	buffer := make([]int16, dataArrayLength)
	if err := binary.Read(f.File, binary.LittleEndian, buffer); err != nil {
		return nil, err
	}

	return buffer, nil
}


func (f NDFile) GetIndex(lat, lng float64) (int, int) {
	tolerance := f.Dx / 2

	lat1 := lat - tolerance
	lat2 := lat + tolerance
	lon1 := lng - tolerance
	lon2 := lng + tolerance

	var latIndex, lngIndex int

	for i, lat := range f.DistinctLatitudes {
		if lat >= lat1 && lat <= lat2 {
			for j, lon := range f.DistinctLongitudes {
				if lon >= lon1 && lon <= lon2 {
					latIndex = i
					lngIndex = j
					break
				}
			}
		}
	}

	return latIndex, lngIndex
}



func PreFetch(filename string) NDFile {

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("Error opening file: ", err)
	}

	var headerLength int64
	err = binary.Read(file, binary.LittleEndian, &headerLength)
	if err != nil {
		log.Fatal("Error reading header length: ", err)
	}

	headerBytes := make([]byte, headerLength)

	_, err = file.Seek(8, 0)
	if err != nil {
		log.Fatal("Error seeking to header: ", err)
	}

	_, err = file.Read(headerBytes)
	if err != nil {
		log.Fatal("Error reading header: ", err)
	}

	fh := &NDFileHeader{}
	err = fh.Deserialize(headerBytes)
	if err != nil {
		log.Fatal("Unmarshaling error: ", err)
	}

	return NDFile{
		NDFileHeader: fh,
		File:         file,
		HeaderLength: headerLength + 8,
	}
}
