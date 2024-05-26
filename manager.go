package ndfile

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

type NDFileManager struct {
	RootPath              string
	TimeIntervalInMinutes int
}

func NewNDFileManager(rootPath string, timeInterval int) *NDFileManager {
	return &NDFileManager{
		RootPath:              rootPath,
		TimeIntervalInMinutes: timeInterval,
	}
}

func (nd *NDFileManager) AddGrib(grib GRIBFile) {

	fileName := nd.getFileName(grib)

	if !checkFileExists(fileName) {
		nd.CreateNDFile(fileName, grib)
	} else {
		nd.AddToNDFile(fileName, grib)
	}

}

func (nd *NDFileManager) CreateNDFile(fileName string, grib GRIBFile) {

	unixTsOfGrib := grib.ReferenceTime.Unix()
	unixStartOfDay := unixTsOfGrib - (unixTsOfGrib % 86400)

	indexOfInterval := (unixTsOfGrib - unixStartOfDay) / (int64(nd.TimeIntervalInMinutes) * 60)

	if (unixTsOfGrib-unixStartOfDay)%(int64(nd.TimeIntervalInMinutes)*60) != 0 {
		log.Fatal("Reference time of GRIB file is not a multiple of the time interval in minutes")
	}

	fh := &NDFileHeader{
		Type:                  grib.Type,
		La1:                   grib.La1,
		La2:                   grib.La2,
		Lo1:                   grib.Lo1,
		Lo2:                   grib.Lo2,
		Nx:                    int32(grib.Nx),
		Ny:                    int32(grib.Ny),
		Dx:                    grib.DX,
		Dy:                    grib.DY,
		DistinctLatitudes:     grib.DistinctLatitudes,
		DistinctLongitudes:    grib.DistinctLongitudes,
		StartTS:               unixStartOfDay,
		TimeIntervalInMinutes: int32(nd.TimeIntervalInMinutes),
		ForecastSteps:         int32(indexOfInterval + 1),
	}

	headerBytes, err := fh.Serialize()
	if err != nil {
		log.Fatal("Marshaling error: ", err)
	}

	var completeFile bytes.Buffer

	var headerLength int64 = int64(len(headerBytes))

	binary.Write(&completeFile, binary.LittleEndian, headerLength)
	completeFile.Write(headerBytes)

	dataArrayLength := (24 * 60) / nd.TimeIntervalInMinutes

	var dataBuffer bytes.Buffer

	scaleFactor := 100.0

	for i := int64(0); i < int64(fh.Nx*fh.Ny); i++ {

		value := grib.DataValues[i]

		var dataArr []int16 = make([]int16, int(dataArrayLength))
		for i := 0; i < int(dataArrayLength); i++ {
			dataArr[i] = 32767
		}

		dataArr[indexOfInterval] = int16(value * scaleFactor)

		binary.Write(&dataBuffer, binary.LittleEndian, dataArr)
	}

	completeFile.Write(dataBuffer.Bytes())

	err = os.WriteFile(fileName, completeFile.Bytes(), 0644)
	if err != nil {
		log.Fatal("Error writing data to file: ", err)
	}

}

func (nd *NDFileManager) AddToNDFile(fileName string, grib GRIBFile) {
	unixTsOfGrib := grib.ReferenceTime.Unix()
	unixStartOfDay := unixTsOfGrib - (unixTsOfGrib % 86400)

	indexOfInterval := (unixTsOfGrib - unixStartOfDay) / (int64(nd.TimeIntervalInMinutes) * 60)

	if (unixTsOfGrib-unixStartOfDay)%(int64(nd.TimeIntervalInMinutes)*60) != 0 {
		log.Fatal("Reference time of GRIB file is not a multiple of the time interval in minutes")
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal("Error reading file: ", err)
	}

	headerLength := int64(binary.LittleEndian.Uint64(data[:8]))
	headerBytes := data[8 : headerLength+8]

	fh := &NDFileHeader{}
	err = fh.Deserialize(headerBytes)
	if err != nil {
		log.Fatal("Unmarshaling error: ", err)
	}

	fh.ForecastSteps++

	headerBytes, err = fh.Serialize()
	if err != nil {
		log.Fatal("Marshaling error: ", err)
	}

	existingData := data[headerLength+8:]

	// Write updated header back to the file
	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		log.Fatal("Error opening file: ", err)
	}
	defer file.Close()

	_, err = file.Seek(0, 0)
	if err != nil {
		log.Fatal("Error seeking to beginning of file: ", err)
	}

	binary.Write(file, binary.LittleEndian, int64(len(headerBytes)))
	_, err = file.Write(headerBytes)
	if err != nil {
		log.Fatal("Error writing header: ", err)
	}

	dataArrayLength := int64((24 * 60) / nd.TimeIntervalInMinutes)

	for i := int64(0); i < int64(fh.Nx*fh.Ny); i++ {
		existingDataStart := i * dataArrayLength * 2

		buffer := make([]int16, dataArrayLength)
		binary.Read(bytes.NewReader(existingData[existingDataStart:existingDataStart+dataArrayLength*2]), binary.LittleEndian, &buffer)

		buffer[indexOfInterval] = int16(grib.DataValues[i] * 100)

		// Write only the modified byte back to the file
		modifiedBytePosition := headerLength + 8 + existingDataStart + indexOfInterval*2
		_, err = file.Seek(modifiedBytePosition, 0)
		if err != nil {
			log.Fatal("Error seeking to byte position: ", err)
		}

		err = binary.Write(file, binary.LittleEndian, buffer[indexOfInterval])
		if err != nil {
			log.Fatal("Error writing byte: ", err)
		}
	}
}

func checkFileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		return false
	}

	headerLength := int64(binary.LittleEndian.Uint64(data[:8]))
	headerBytes := data[8 : headerLength+8]

	fh := &NDFileHeader{}
	err = fh.Deserialize(headerBytes)

	return err == nil
}

func (nd *NDFileManager) getFileName(grib GRIBFile) string {

	utcTime := grib.ReferenceTime.UTC()
	daysSinceEpoch := int(utcTime.Sub(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)).Hours() / 24)

	return path.Join(nd.RootPath, fmt.Sprintf("%d_%d.nd", grib.Type, daysSinceEpoch))
}
