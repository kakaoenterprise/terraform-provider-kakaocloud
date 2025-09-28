# Manage example volume.
resource "kakaocloud_volume" "example" {
  name              = "example"
  description       = "terraform test"
  availability_zone = "kr-central-2-a"
  flavor_id         = "xxxxxx-yyyy-zzzz-81e8-836ebe461ba6"
  image_id          = "xxxxxx-yyyy-zzzz-81e8-836ebe461ba6"
  key_name          = "example-test"
  volumes = [
    { size = 30 },
  ]
  subnets = [
    {
      id = "xxxxxx-yyyy-zzzz-81e8-836ebe461ba6"
    }
  ]
  status = "active"
}