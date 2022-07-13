#!/bin/bash

echo "Put PG config to etcd"
etcdctl put 'cluster_1' <<EOF
{
  "servers": {
    "server1": "postgres://app:pwd@localhost:5444/shipment",
    "server2": "postgres://app:pwd@localhost:5445/shipment"
  },
  "buckets": {
    "0": "server1",
    "1": "server2",
    "2": "server1",
    "3": "server2",
    "4": "server1",
    "5": "server2",
    "6": "server1",
    "7": "server2"
  }
}
EOF
