package ndfile

import (
	"bytes"
	"encoding/binary"
)

type DataType int32


type NDFileHeader struct {
	Type                  int32
	La1                   float64
	La2                   float64
	Lo1                   float64
	Lo2                   float64
	Nx                    int32
	Ny                    int32
	Dx                    float64
	Dy                    float64
	DistinctLatitudes     []float64
	DistinctLongitudes    []float64
	StartTS               int64
	TimeIntervalInMinutes int32
	ForecastSteps         int32
}

func (h *NDFileHeader) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.LittleEndian, h.Type); err != nil {
		return nil, err
	}
	fields := []interface{}{h.La1, h.La2, h.Lo1, h.Lo2, h.Nx, h.Ny, h.Dx, h.Dy, h.StartTS, h.TimeIntervalInMinutes, h.ForecastSteps}
	for _, field := range fields {
		if err := binary.Write(buf, binary.LittleEndian, field); err != nil {
			return nil, err
		}
	}

	if err := writeFloat64Slice(buf, h.DistinctLatitudes); err != nil {
		return nil, err
	}
	if err := writeFloat64Slice(buf, h.DistinctLongitudes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (h *NDFileHeader) Deserialize(data []byte) error {
	buf := bytes.NewReader(data)

	if err := binary.Read(buf, binary.LittleEndian, &h.Type); err != nil {
		return err
	}
	fields := []interface{}{&h.La1, &h.La2, &h.Lo1, &h.Lo2, &h.Nx, &h.Ny, &h.Dx, &h.Dy, &h.StartTS, &h.TimeIntervalInMinutes, &h.ForecastSteps}
	for _, field := range fields {
		if err := binary.Read(buf, binary.LittleEndian, field); err != nil {
			return err
		}
	}

	if err := readFloat64Slice(buf, &h.DistinctLatitudes); err != nil {
		return err
	}
	if err := readFloat64Slice(buf, &h.DistinctLongitudes); err != nil {
		return err
	}

	return nil
}

func writeFloat64Slice(buf *bytes.Buffer, slice []float64) error {
	if err := binary.Write(buf, binary.LittleEndian, int32(len(slice))); err != nil {
		return err
	}
	for _, value := range slice {
		if err := binary.Write(buf, binary.LittleEndian, value); err != nil {
			return err
		}
	}
	return nil
}

func readFloat64Slice(buf *bytes.Reader, slice *[]float64) error {
	var length int32
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return err
	}
	*slice = make([]float64, length)
	for i := 0; i < int(length); i++ {
		if err := binary.Read(buf, binary.LittleEndian, &(*slice)[i]); err != nil {
			return err
		}
	}
	return nil
}
