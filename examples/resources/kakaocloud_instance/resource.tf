# Manage example volume.
resource "kakaocloud_instance" "example" {
  name              = "example"
  description       = "terraform test"
  size              = 1
  availability_zone = "kr-central-2-a"
}