terraform {
  required_providers {
    kakaocloud = {
      source = "registry.terraform.io/hashicorp/kakaocloud"
    }
  }
}

# Configure the KakaoCloud Provider
# Make sure your X_AUTH_TOKEN is set as an environment variable
provider "kakaocloud" {
  # provider config
}

# Use the data source to fetch the load balancer
data "kakaocloud_load_balancer" "my_lb" {
  id = "lb-xxxxxxxx" # <-- TODO: Replace with a REAL load balancer ID
}

# Output the fetched data
output "load_balancer_name" {
  value = data.kakaocloud_load_balancer.my_lb.name
}

output "load_balancer_private_vip" {
  value = data.kakaocloud_load_balancer.my_lb.private_vip
}

output "load_balancer_operating_status" {
  value = data.kakaocloud_load_balancer.my_lb.operating_status
}
