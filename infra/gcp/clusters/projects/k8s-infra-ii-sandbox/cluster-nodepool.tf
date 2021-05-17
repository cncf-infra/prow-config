resource "google_container_node_pool" "pods-go-here" {
  name_prefix = "pods-go-here-"
  location    = google_container_cluster.cluster.location
  cluster     = google_container_cluster.cluster.name

  provider = google-beta
  project  = data.google_project.project.project_id

  // Start with a single node
  initial_node_count = 1

  // Auto repair, and auto upgrade nodes to match the master version
  management {
    auto_repair  = true
    auto_upgrade = true
  }

  // Autoscale the cluster as needed. Note that these values will be multiplied
  // by 3, as the cluster will exist in three zones
  autoscaling {
    min_node_count = 1
    max_node_count = 1
  }

  // Set machine type, and enable all oauth scopes tied to the service account
  node_config {
    // IMPORTANT: Should be n1-standard-1 on test clusters
    image_type   = "COS_CONTAINERD"
    machine_type = "n1-standard-2"
    disk_size_gb = 100
    disk_type    = "pd-standard"

    service_account = google_service_account.cluster_node_sa.email
    oauth_scopes    = ["https://www.googleapis.com/auth/cloud-platform"]

    // Needed for workload identity
    workload_metadata_config {
      node_metadata = "GKE_METADATA_SERVER"
    }
    metadata = {
      disable-legacy-endpoints = "true"
    }
  }

  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }

  // If we need to destroy the node pool, create the new one before destroying
  // the old one
  lifecycle {
    create_before_destroy = true
  }
}
