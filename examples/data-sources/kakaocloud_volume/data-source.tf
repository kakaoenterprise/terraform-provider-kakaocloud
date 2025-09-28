# get a volume.
data "kakaocloud_volume" "example" {
  id = "your-volume-id-here"  # Replace with your volume ID
}

# Output the volume information
output "volume_example" {
  description = "Information about the example volume"
  value = kakaocloud_volume.example
}