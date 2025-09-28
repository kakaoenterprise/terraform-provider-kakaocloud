# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all load balancer listeners
data "kakaocloud_load_balancer_listeners" "all" {
  # No filters - get all listeners
}

# List listeners for a specific load balancer using filter
data "kakaocloud_load_balancer_listeners" "by_load_balancer" {
  filter = [
    {
      name  = "id"
      value = "your-listener-id-here"  # Replace with your listener ID
    },
    {
      name  = "load_balancer_id"
      value = "your-load-balancer-id-here"  # Replace with your load balancer ID
    },
    {
      name  = "protocol"
      value = "HTTP"
    },
    {
      name  = "protocol_port"
      value = "80"
    },
    {
      name  = "provisioning_status"
      value = "ACTIVE"
    },
    {
      name  = "operating_status"
      value = "ONLINE"
    },
    {
      name  = "secret_name"
      value = "your-secret-name"  # Replace with your secret name
    },
    {
      name  = "secret_id"
      value = "your-secret-id"  # Replace with your secret ID
    },
    {
      name  = "tls_certificate_id"
      value = "your-tls-certificate-id"  # Replace with your TLS certificate ID
    },
    {
      name  = "created_at"
      value = "2021-01-01T00:00:00Z"  # Replace with your created at
    },
    {
      name  = "updated_at"
      value = "2021-01-01T00:00:00Z"  # Replace with your updated at
    }
  ]
}

# List listeners with filters
data "kakaocloud_load_balancer_listeners" "filtered" {
  filter = [
    {
      name  = "protocol"
      value = "HTTP"
    },
    {
      name  = "protocol_port"
      value = "80"
    },
    {
      name  = "provisioning_status"
      value = "ACTIVE"
    }
  ]
}

# List HTTPS listeners with TLS certificate filters
data "kakaocloud_load_balancer_listeners" "https_listeners" {
  filter = [
    {
      name  = "protocol"
      value = "TERMINATED_HTTPS"
    },
    {
      name  = "operating_status"
      value = "ONLINE"
    },
    {
      name  = "secret_name"
      value = "your-secret-name"  # Replace with your secret name
    }
  ]
}

# List listeners by load balancer and TLS certificate
data "kakaocloud_load_balancer_listeners" "by_lb_and_tls" {
  filter = [
    {
      name  = "load_balancer_id"
      value = "your-load-balancer-id-here"  # Replace with your load balancer ID
    },
    {
      name  = "tls_certificate_id"
      value = "your-tls-certificate-id"  # Replace with your TLS certificate ID
    },
    {
      name  = "secret_id"
      value = "your-secret-id"  # Replace with your secret ID
    }
  ]
}

# Output all listeners
output "all_listeners" {
  description = "List of all load balancer listeners"
  value = {
    count = length(data.kakaocloud_load_balancer_listeners.all.listeners)
    ids   = data.kakaocloud_load_balancer_listeners.all.listeners[*].id
    protocols = data.kakaocloud_load_balancer_listeners.all.listeners[*].protocol
  }
}

# Output listeners by load balancer
output "load_balancer_listeners" {
  description = "Listeners for specific load balancer"
  value = {
    count = length(data.kakaocloud_load_balancer_listeners.by_load_balancer.listeners)
    ids   = data.kakaocloud_load_balancer_listeners.by_load_balancer.listeners[*].id
    protocols = data.kakaocloud_load_balancer_listeners.by_load_balancer.listeners[*].protocol
  }
}

# Output filtered listeners
output "filtered_listeners" {
  description = "Filtered load balancer listeners"
  value = {
    count = length(data.kakaocloud_load_balancer_listeners.filtered.listeners)
    ids   = data.kakaocloud_load_balancer_listeners.filtered.listeners[*].id
    protocols = data.kakaocloud_load_balancer_listeners.filtered.listeners[*].protocol
  }
}

# Output HTTPS listeners
output "https_listeners" {
  description = "HTTPS load balancer listeners"
  value = {
    count = length(data.kakaocloud_load_balancer_listeners.https_listeners.listeners)
    ids   = data.kakaocloud_load_balancer_listeners.https_listeners.listeners[*].id
    protocols = data.kakaocloud_load_balancer_listeners.https_listeners.listeners[*].protocol
  }
}
