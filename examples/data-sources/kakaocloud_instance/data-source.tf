# get a instance.
data "kakaocloud_instance" "example" {
  id = "your-instance-id-here"  # Replace with your instance ID
  
}

# Output the instance information
output "instance_example" {
  description = "Information about the example instance"
  value = kakaocloud_instance.example
}