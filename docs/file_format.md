# NDFile Format Documentation

## Overview

The `ndfile` format is a binary file structure developed specifically for storing and efficiently accessing Numerical Weather Prediction (NWP) data across multiple forecast steps. It is designed to provide rapid, point-specific data access, significantly reducing the time required to retrieve information without the need to parse the entire file each time.

## File Structure

```
+------------------------------------------------------------------------------------------------+
|                                          NDFile                                                |
+------------------------------------------------------------------------------------------------+
|                                          File Header                                            |
+----------------------+-------------------+-----------------------------------------------------+
| Field                | Binary Type       | Description                                         |
+----------------------+-------------------+-----------------------------------------------------+
| Type                 | DataType (enum)   | Type of the GRIB data (e.g., Temperature, Pressure) |
| La1                  | float64           | Latitude of the first grid point                    |
| La2                  | float64           | Latitude of the last grid point                     |
| Lo1                  | float64           | Longitude of the first grid point                   |
| Lo2                  | float64           | Longitude of the last grid point                    |
| Nx                   | int32             | Number of points along the X-axis                   |
| Ny                   | int32             | Number of points along the Y-axis                   |
| Dx                   | float64           | Longitude grid length                               |
| Dy                   | float64           | Latitude grid length                                |
| DistinctLatitudes    | []float64         | List of distinct latitudes                          |
| DistinctLongitudes   | []float64         | List of distinct longitudes                         |
| StartTS              | int64             | Start timestamp of the day (Unix Time)              |
| TimeIntervalInMinutes| int32             | Time interval in minutes between forecast steps     |
| ForecastSteps        | int32             | Total number of forecast steps in the file          |
+----------------------+-------------------+-----------------------------------------------------+
|                                       Data Section                                             |
+------------------------------------------------------------------------------------------------+
| Index                | Data Structure    | Description                                         |
|                      |                   |                                                     |
|                      | - Each row corresponds to a grid point                                  |
|                      | - Each column within a row corresponds to a forecast step               |
|                      | - Values are typically int16, with 32767 representing missing data      |
+------------------------------------------------------------------------------------------------+
```

### Hint
The slices within the header are encoded as binary data, with the length of the slice encoded as an int32 followed by the slice elements.
