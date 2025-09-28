# get a volume snapshot.
data "kakaocloud_volume_snapshot" "example" {
  id = "your-volume-snapshot-id-here"  # Replace with your volume snapshot ID
}

# Output the volume snapshot information
output "volume_snapshot_example" {
  description = "Information about the example volume snapshot"
  value = kakaocloud_volume_snapshot.example
}