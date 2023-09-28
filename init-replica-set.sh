#!/bin/bash
init_rs()
{
    sleep 5
    mongosh --eval 'rs.initiate("{\"_id\":\"rs0\",\"members\":[{\"_id\":0,\"host\":\"localhost:27017\"}]}")'
}

init_rs & mongod --noauth --port 27017 --bind_ip_all --replSet rs0