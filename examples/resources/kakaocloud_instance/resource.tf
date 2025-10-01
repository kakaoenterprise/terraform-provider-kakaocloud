# Manage example instance
resource "kakaocloud_instance" "example" {
  name      = "example"
  flavor_id = "a5b8e9d2-4f0c-4bfc-a58b-b6c5e5f88e27"
  image_id  = "8cd12d35-fcd5-42af-b5ef-973b926a13e1"

  subnets = [
    {
      id = "93fd1234-6c18-4975-bcb7-818e3bde62c9"
    },
  ]

  volumes = [
    {
      size = 10
    },
  ]
}