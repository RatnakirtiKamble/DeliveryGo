#!/bin/bash

docker exec -it kafka kafka-topics \
  --bootstrap-server localhost:9092 \
  --create --if-not-exists \
  --topic batches.assigned --partitions 1 --replication-factor 1

docker exec -it kafka kafka-topics \
  --bootstrap-server localhost:9092 \
  --create --if-not-exists \
  --topic routes.refine.requested --partitions 1 --replication-factor 1

docker exec -it kafka kafka-topics \
  --bootstrap-server localhost:9092 \
  --create --if-not-exists \
  --topic routes.refined.requested --partitions 1 --replication-factor 1

docker exec -it kafka kafka-topics \
  --bootstrap-server localhost:9092 \
  --create --if-not-exists \
  --topic batch.reassign.requested --partitions 1 --replication-factor 1

docker exec -it kafka kafka-topics \
  --bootstrap-server localhost:9092 \
  --create --if-not-exists \
  --topic path.provision.requested --partitions 1 --replication-factor 1
