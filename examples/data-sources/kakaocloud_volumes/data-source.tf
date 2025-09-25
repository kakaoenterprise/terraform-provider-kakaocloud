# List all volumes or filtered volumes.
data "kakaocloud_volumes" "example_list" {
  filter = [
    {
      name  = "name"
      value = "example"
    }
  ]
}