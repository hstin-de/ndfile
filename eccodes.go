package ndfile

/*
#cgo CFLAGS: -I/usr/include/x86_64-linux-gnu/
#cgo LDFLAGS: -leccodes
#include "eccodes.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"time"
	"unsafe"
)

// GRIBFile struct to hold GRIB data information
type GRIBFile struct {
	Type               int32     `json:"type"`
	Nx                 int       `json:"nx"`
	Ny                 int       `json:"ny"`
	La1                float64   `json:"la1"`
	La2                float64   `json:"la2"`
	Lo1                float64   `json:"lo1"`
	Lo2                float64   `json:"lo2"`
	DX                 float64   `json:"dx"`
	DY                 float64   `json:"dy"`
	DataValues         []float64 `json:"dataValues"`
	ScanMode           int       `json:"scanMode"`
	ReferenceTime      time.Time `json:"referenceTime"`
	ForecastTime       string    `json:"forecastTime"`
	DistinctLatitudes  []float64 `json:"distinctLatitudes"`
	DistinctLongitudes []float64 `json:"distinctLongitudes"`
}


func ProcessGRIB(gribData []byte) GRIBFile {
	dataPtr := unsafe.Pointer(&gribData[0])
	dataSize := C.size_t(len(gribData))

	var gid *C.codes_handle = C.codes_handle_new_from_message(C.codes_context_get_default(), dataPtr, dataSize)
	if gid == nil {
		fmt.Println("Failed to create handle from message")
		return GRIBFile{}
	}
	defer C.codes_handle_delete(gid)

	// Extract grid information
	var nx, ny C.long
	var la1, la2, lo1, lo2, dx, dy, basicAngle, subdivisions C.double
	var values *C.double
	var numValues C.size_t
	var year, month, day, hour, minute, second, timeUnit, forecastTime, scanMode C.long

	var discipline, parameterCategory, parameterNumber C.long

	C.codes_get_long(gid, C.CString("Ni"), &nx)
	C.codes_get_long(gid, C.CString("Nj"), &ny)
	C.codes_get_double(gid, C.CString("latitudeOfFirstGridPointInDegrees"), &la1)
	C.codes_get_double(gid, C.CString("latitudeOfLastGridPointInDegrees"), &la2)
	C.codes_get_double(gid, C.CString("longitudeOfFirstGridPointInDegrees"), &lo1)
	C.codes_get_double(gid, C.CString("longitudeOfLastGridPointInDegrees"), &lo2)
	C.codes_get_double(gid, C.CString("iDirectionIncrement"), &dx)
	C.codes_get_double(gid, C.CString("jDirectionIncrement"), &dy)
	C.codes_get_double(gid, C.CString("basicAngleOfTheInitialProductionDomain"), &basicAngle)
	C.codes_get_double(gid, C.CString("subdivisionsOfBasicAngle"), &subdivisions)

	scale := 1e6 // Default scale if no basicAngle is defined

	if basicAngle != 0 {
		scale = float64(basicAngle) / float64(subdivisions)
	}

	// Extract reference time
	C.codes_get_long(gid, C.CString("year"), &year)
	C.codes_get_long(gid, C.CString("month"), &month)
	C.codes_get_long(gid, C.CString("day"), &day)
	C.codes_get_long(gid, C.CString("hour"), &hour)
	C.codes_get_long(gid, C.CString("minute"), &minute)
	C.codes_get_long(gid, C.CString("second"), &second)
	C.codes_get_long(gid, C.CString("indicatorOfUnitOfTimeRange"), &timeUnit)
	C.codes_get_long(gid, C.CString("forecastTime"), &forecastTime)
	C.codes_get_long(gid, C.CString("scanMode"), &scanMode)
	C.codes_get_long(gid, C.CString("discipline"), &discipline)
	C.codes_get_long(gid, C.CString("parameterCategory"), &parameterCategory)
	C.codes_get_long(gid, C.CString("parameterNumber"), &parameterNumber)

	gribType := int32((discipline & 0xFF) | ((parameterCategory & 0xFF) << 8) | ((parameterNumber & 0xFF) << 16))

	latitudes := make([]float64, ny)
	C.codes_get_double_array(gid, C.CString("distinctLatitudes"), (*C.double)(unsafe.Pointer(&latitudes[0])), (*C.size_t)(unsafe.Pointer(&ny)))

	longitudes := make([]float64, nx)
	C.codes_get_double_array(gid, C.CString("distinctLongitudes"), (*C.double)(unsafe.Pointer(&longitudes[0])), (*C.size_t)(unsafe.Pointer(&nx)))

	// Reverse latitudes if they are inverted
	if ny > 1 && la2 < la1 && latitudes[len(latitudes)-1] > latitudes[0] {
		for i, j := 0, len(latitudes)-1; i < j; i, j = i+1, j-1 {
			latitudes[i], latitudes[j] = latitudes[j], latitudes[i]
		}
	}

	referenceTime := time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), int(second), 0, time.UTC)
	var forecastDuration time.Duration

	// Adjust reference time by forecast period
	// https://codes.ecmwf.int/grib/format/grib2/ctables/4/4/
	switch timeUnit {
	case 0: // Minute
		forecastDuration = time.Duration(forecastTime) * time.Minute
	case 1: // Hour
		forecastDuration = time.Duration(forecastTime) * time.Hour
	case 2: // Day
		forecastDuration = time.Duration(forecastTime) * 24 * time.Hour
	case 3: // Month
		forecastDuration = time.Duration(forecastTime) * 24 * time.Hour * 30
	case 4: // Year
		forecastDuration = time.Duration(forecastTime) * 24 * time.Hour * 365
	case 5: // Decade (10 years)
		forecastDuration = time.Duration(forecastTime) * 24 * time.Hour * 365 * 10
	case 6: // Normal (30 years)
		forecastDuration = time.Duration(forecastTime) * 24 * time.Hour * 365 * 30
	case 7: // Century (100 years)
		forecastDuration = time.Duration(forecastTime) * 24 * time.Hour * 365 * 100
	case 10: // 3 hours
		forecastDuration = time.Duration(forecastTime) * 3 * time.Hour
	case 11: // 6 hours
		forecastDuration = time.Duration(forecastTime) * 6 * time.Hour
	case 12: // 12 hours
		forecastDuration = time.Duration(forecastTime) * 12 * time.Hour
	case 13: // Second
		forecastDuration = time.Duration(forecastTime) * time.Second
	case 255: // Missing
		fmt.Println("Forecast time is missing.")
	default:
		fmt.Printf("Unsupported time unit: %d\n", timeUnit)
	}

	forecastReferenceTime := referenceTime.Add(forecastDuration)

	// Getting the values
	if C.codes_get_size(gid, C.CString("values"), &numValues) == C.CODES_SUCCESS {
		values = (*C.double)(C.malloc(numValues * C.sizeof_double))
		defer C.free(unsafe.Pointer(values))

		if C.codes_get_double_array(gid, C.CString("values"), values, &numValues) == C.CODES_SUCCESS {
			dataValues := make([]float64, numValues)
			for i := C.size_t(0); i < numValues; i++ {
				dataValues[i] = float64(*(*C.double)(unsafe.Pointer(uintptr(unsafe.Pointer(values)) + uintptr(i)*uintptr(C.sizeof_double))))
			}

			la1 := float64(la1)
			la2 := float64(la2)
			lo1 := float64(lo1)
			lo2 := float64(lo2)

			dx := float64(dx) / scale
			dy := float64(dy) / scale

			// // Correct DY if latitudes are inverted
			// if la1 > la2 {
			// 	dy = -dy
			// }
			// // Correct DX if longitudes are inverted
			// if lo1 > lo2 {
			// 	dx = -dx
			// }

			parsedGrib := GRIBFile{
				Type:               gribType,
				Nx:                 int(nx),
				Ny:                 int(ny),
				La1:                la1,
				La2:                la2,
				Lo1:                lo1,
				Lo2:                lo2,
				DX:                 dx,
				DY:                 dy,
				DataValues:         dataValues,
				ScanMode:           int(scanMode),
				ReferenceTime:      forecastReferenceTime,
				ForecastTime:       fmt.Sprintf("%d", forecastTime),
				DistinctLatitudes:  latitudes,
				DistinctLongitudes: longitudes,
			}

			// Corrected code to offset even rows in the x-direction.
			var correctedDataValues []float64

			offset := int(parsedGrib.DX / 2) // Convert the offset to an integer.

			for y := 0; y < parsedGrib.Ny; y++ {
				for x := 0; x < parsedGrib.Nx; x++ {
					// Calculate the index in the original DataValues slice.
					index := y*parsedGrib.Nx + x

					// Apply the offset for even rows if necessary.
					// Ensure that you stay within the bounds of your DataValues slice.
					if y%2 == 0 && (index+offset) < len(parsedGrib.DataValues) {
						correctedDataValues = append(correctedDataValues, parsedGrib.DataValues[index+offset])
					} else {
						correctedDataValues = append(correctedDataValues, parsedGrib.DataValues[index])
					}
				}
			}

			// Update the parsedGrib structure with the corrected data.
			parsedGrib.DataValues = correctedDataValues

			// Update the parsedGrib structure with the corrected data.
			parsedGrib.DataValues = correctedDataValues

			return parsedGrib
		}
	}

	return GRIBFile{}
}
