# Applications

- [ETL Pipeline](#etl-pipeline)
- [UDF](#udf)

## Requirements

Development requirements:
- [`go >~ 1.17`](https://go.dev/)
- [gnuMake](https://www.gnu.org/software/make/)

Run to read what exec commands are available for corresponding applications:

```commandline
make help
```

## ETL Pipeline

- [Makefile](Makefile):
- [Codebase](pipeline)

### Indexmodules

The app to fetch cached module path and version from the [Go module index](https://index.golang.org/).

_The tool_: [application codebase](pipeline/indexmodules)

## UDF

Applications to define [BigQuery UDF](https://cloud.google.com/bigquery/docs/reference/standard-sql/remote-functions).

### Moduleversions

The app to extract `min` and `max` module version given the array of its versions.

_The tool_: [applaication codebase](udf/moduleversions)
