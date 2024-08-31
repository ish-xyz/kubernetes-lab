locals {
  mock_instances = [
    {
      tags = {
        Name = "node-1"
      }
    },
    {
      tags = {
        Name = "node-2"
      }
    },
    {
      tags = {
        Name = "node-3"
      }
    }
  ]

  nodes = [
      for i in range(var.nodes_count) : {
          name = "node-${i}-${var.cluster_name}"
      }
  ]
  node_names_string = join(" ", [for node in local.nodes: node.name])


  controllers = [
    for i in range(var.nodes_count) : {
        name = "controller-${i}-${var.cluster_name}"
    }
  ]
}