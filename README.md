# ndfile

`ndfile` is a binary file format designed for the efficient storage and querying of numerical weather prediction (NWP) data across multiple forecast steps. It enables rapid access to point-specific data without the need to parse the entire file each time, allowing for data queries in under 5 microseconds on NVMe storage systems.

## Features

- **Efficient Data Access**: Read the header once and perform subsequent data queries rapidly.
- **Optimized for Performance**: Tailored for high-speed access on NVMe disks, facilitating quick updates and retrievals.
- **Focused on NWP Data**: Ideal for scenarios requiring quick access to multiple forecast steps of weather prediction data.

Each ndfile aggregates a day's worth of numerical weather prediction (NWP) data. The frequency of data points, or timestep, can be customized via settings when initializing the data manager. Additionally, the manager intelligently identifies the appropriate timestep and file for each GRIB file you wish to insert, ensuring accurate and seamless integration into the dataset.

## Why is it so fast?

`ndfile` optimizes data retrieval by minimizing the operations needed to access specific data points. Each query involves just one disk seek and one read operation, bypassing the need to load extensive amounts of data into system memory. By reading data directly from the disk, `ndfile` allows for immediate access without parsing the entire file. This streamlined approach reduces processing time significantly, making it ideal for applications where speed is critical.

See: [File Format](docs/file_format.md) for more information on the file format.


## Installation

```bash
go get github.com/hstin-de/ndfile
```


## Usage

### Inserting Data
```go
package main

import (
	"log"
	"os"

	"github.com/hstin-de/ndfile"
)

func main() {

	ndFileManager := ndfile.NewNDFileManager("", 60)

	gribfile, err := os.ReadFile("path/to/gribfile.grib")
	if err != nil {
		log.Fatal(err)
	}

	ndFileManager.AddGrib(ndfile.ProcessGRIB(gribfile))

}
```

This will create a corresponding `.nd` file in the same directory as the GRIB file.

To read data from the `.nd` file, use the following code:

### Reading Data

```go
package main

import (
	"fmt"
	"math"

	"github.com/hstin-de/ndfile"
)

func main() {

	ndFile := ndfile.PreFetch("path/to/ndfile.nd")

	lat, lng := 52.52, 13.405
	latIndex, lngIndex := ndFile.GetIndex(lat, lng) // This can be cached for future queries to the same ndFile

	rawValues, err := ndFile.GetData(latIndex, lngIndex)
	if err != nil {
		fmt.Println("Error getting data:", err)
		return
	}

	var data []float64 = make([]float64, len(rawValues))

	for i, value := range rawValues {
		if value == 32767 {
			data[i] = math.NaN()
			continue
		}
		data[i] = float64(value) / 100.0
	}

	fmt.Println(data)

}
```


## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.