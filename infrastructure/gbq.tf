resource "google_bigquery_dataset" "raw" {
  dataset_id                      = "raw"
  project                         = "go-mod-analysis"
  friendly_name                   = "raw"
  description                     = "Dataset to store extracted data"
  location                        = local.region
  # one week
  default_partition_expiration_ms = 7 * 86400 * 1000

  delete_contents_on_destroy = true
}

resource "google_bigquery_table" "index" {
  dataset_id    = google_bigquery_dataset.raw.dataset_id
  project       = google_bigquery_dataset.raw.project
  table_id      = "index"
  friendly_name = "index"
  description   = "Data from https://index.golang.org/"

  time_partitioning {
    type          = "DAY"
    expiration_ms = 0
  }

  deletion_protection = false

  schema = <<EOF
[
  {
    "name": "path",
    "type": "STRING",
    "mode": "REQUIRED",
    "description": "The module path"
  },
  {
    "name": "version",
    "type": "STRING",
    "mode": "REQUIRED",
    "description": "The module version"
  },
  {
    "name": "timestamp",
    "type": "TIMESTAMP",
    "mode": "REQUIRED",
    "description": "Time the version was first cached by proxy.golang.org"
  }
]
EOF
}

resource "google_bigquery_table" "index_stat" {
  dataset_id    = google_bigquery_table.index.dataset_id
  project       = google_bigquery_table.index.project
  table_id      = "index_stat"
  friendly_name = "index_stat"
  description   = "Stats on the index table"

  time_partitioning {
    type = "DAY"
  }

  deletion_protection = false

  view {
    use_legacy_sql = false

    query = join(" ", [
      "SELECT DISTINCT",
      join(", ", [
        "path",
        "COUNT(DISTINCT version) OVER (PARTITION BY path) AS cnt_versions",
        "FIRST_VALUE(version) OVER (PARTITION BY path ORDER BY version DESC) AS latest",
        "FIRST_VALUE(version) OVER (PARTITION BY path ORDER BY version) AS earliest",
        "MIN(timestamp) OVER (PARTITION BY path) AS cache_earliest",
        "MAX(timestamp) OVER (PARTITION BY path) AS cache_latest",
      ]),
      "FROM `${google_bigquery_table.index.project}.${google_bigquery_table.index.dataset_id}.${google_bigquery_table.index.table_id}`;",
    ])
  }
}

resource "google_bigquery_table" "pkggodev" {
  dataset_id    = google_bigquery_dataset.raw.dataset_id
  project       = google_bigquery_dataset.raw.project
  table_id      = "pkggodev"
  friendly_name = "pkggodev"
  description   = "Data from https://pkg.go.dev/"

  time_partitioning {
    type          = "DAY"
    expiration_ms = 0
  }

  deletion_protection = false

  schema = <<EOF
[
  {
    "name": "path",
    "type": "STRING",
    "mode": "REQUIRED",
    "description": "The module path"
  },
  {
    "name": "version",
    "type": "STRING",
    "mode": "REQUIRED",
    "description": "The module version"
  },
  {
    "name": "imports",
    "type": "RECORD",
    "mode": "REQUIRED",
    "description": "The dependencies of the given module",
    "fields": [
      {
        "name": "std",
        "type": "STRING",
        "mode": "REPEATED",
        "description": "The std libraries imported as dependencies by the given module"
      },
      {
        "name": "nonstd",
        "type": "STRING",
        "mode": "REPEATED",
        "description": "The non-std libraries imported as dependencies by the given module"
      }
    ]
  },
  {
    "name": "importedby",
    "type": "STRING",
    "mode": "REPEATED",
    "description": "The modules which use the give on as dependency"
  },
  {
    "name": "timestamp",
    "type": "TIMESTAMP",
    "mode": "REQUIRED",
    "description": "Time the version was first cached by proxy.golang.org"
  }
]
EOF
}

resource "google_bigquery_table" "pkggodev_remain" {
  dataset_id    = google_bigquery_table.index.dataset_id
  project       = google_bigquery_table.index.project
  table_id      = "pkggodev_remain"
  friendly_name = "pkggodev_remain"
  description   = "Modules remaining to fetch from https://pkg.go.dev/"

  time_partitioning {
    type = "DAY"
  }

  deletion_protection = false

  view {
    use_legacy_sql = false

    query = join(" ", [
      "SELECT DISTINCT",
      join(", ", [
        "a.path",
      ]),
      "FROM `${google_bigquery_table.index.project}.${google_bigquery_table.index.dataset_id}.${google_bigquery_table.index.table_id}` AS a",
      "LEFT JOIN `${google_bigquery_table.pkggodev.project}.${google_bigquery_table.pkggodev.dataset_id}.${google_bigquery_table.pkggodev.table_id}` AS b",
      "USING (path, version)",
      "WHERE a.timestamp > COALESCE(b.timestamp, '1970-01-01');"
    ])
  }
}
