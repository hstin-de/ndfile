# ndfile

`ndfile` is a binary file format designed for the efficient storage and querying of numerical weather prediction (NWP) data across multiple forecast steps. It enables rapid access to point-specific data without the need to parse the entire file each time, allowing for data queries in under 5 microseconds on NVMe storage systems.

## Features

- **Efficient Data Access**: Read the header once and perform subsequent data queries rapidly.
- **Optimized for Performance**: Tailored for high-speed access on NVMe disks, facilitating quick updates and retrievals.
- **Focused on NWP Data**: Ideal for scenarios requiring quick access to multiple forecast steps of weather prediction data.

## Installation

```bash
go get github.com/hstin-de/ndfile
```
