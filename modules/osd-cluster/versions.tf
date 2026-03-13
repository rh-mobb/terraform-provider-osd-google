terraform {
  required_version = ">= 1.0"

  required_providers {
    osdgoogle = {
      source  = "terraform.local/local/osd-google"
      version = ">= 0.0.1"
    }
    google = {
      source  = "hashicorp/google"
      version = ">= 5.0"
    }
  }
}
