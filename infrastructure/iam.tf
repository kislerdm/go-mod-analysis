resource "google_service_account" "gbq_admin" {
  account_id   = "gbq-admin"
  display_name = "GBQ Admin"
  description  = "Role to interact with GBQ"
  project      = "go-mod-analysis"
}

resource "google_service_account_iam_member" "gbq_admin" {
  service_account_id = google_service_account.gbq_admin.name
  for_each           = toset([
    "roles/iam.serviceAccountUser",
  ])
  role   = each.value
  member = "serviceAccount:${google_service_account.gbq_admin.email}"
}

resource "google_project_iam_member" "project" {
  project  = google_service_account.gbq_admin.project
  for_each = toset(["roles/bigquery.dataOwner", "roles/bigquery.user"])
  role     = each.value
  member   = "serviceAccount:${google_service_account.gbq_admin.email}"
}
