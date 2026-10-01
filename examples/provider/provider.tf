terraform {
  required_providers {
    nah = {
      source = "hypertf/nah"
    }
  }
}

# Configure the NahCloud provider
provider "nah" {
  # The API endpoint. Defaults to https://nahcloud.com
  # Can also be set via NAH_ENDPOINT environment variable
  endpoint = "https://nahcloud.com"

  # Export NAH_TOKEN with an existing organization API token.
  # Do not commit credentials to configuration.
}
