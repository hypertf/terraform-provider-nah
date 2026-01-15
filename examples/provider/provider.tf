terraform {
  required_providers {
    nah = {
      source = "hypertf/nah"
    }
  }
}

# Configure the NahCloud provider
provider "nah" {
  # The API endpoint. Defaults to http://localhost:8080
  # Can also be set via NAH_ENDPOINT environment variable
  endpoint = "http://localhost:8080"

  # Optional authentication token
  # Can also be set via NAH_TOKEN environment variable
  # token = "your-token"
}
