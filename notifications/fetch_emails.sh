#!/bin/bash

set -xe

if [ $# -ne 1 ]; then
    echo "Error: Please provide a single .sql file as an argument."
    exit 1
fi

file="$1"

if [ ! -f "$file" ]; then
    echo "Error: The provided argument is not a file."
    exit 1
fi

if [[ "$file" != *.sql ]]; then
    echo "Error: The provided file is not a .sql file."
    exit 1
fi

container_name=postgres

docker rm -f $container_name 2>/dev/null
docker run -d --name $container_name -e POSTGRES_PASSWORD=password postgres
#  docker run -d --name $container_name -e POSTGRES_PASSWORD=password -p 0.0.0.0:5432:5432 postgres

while true; do
    container_status=$(docker ps -q --filter "name=$container_name" --filter "status=running")
    if [ -n "$container_status" ]; then
        echo "Container $container_name is now running."
        break
    fi

    # Sleep for a few seconds before checking again
    sleep 5
done
sleep 10

docker cp $file $container_name:/tmp/dump.sql
docker exec -it $container_name psql -U postgres -a -f /tmp/dump.sql
#  docker exec -it $container_name psql -U postgres
#  exit 0

rm to.txt 2>/dev/null || true
docker exec -it $container_name psql -U postgres -P pager=off -tA -c "select distinct u.email from members m join entries e on m.id=e.member_id join users u on u.id=m.user_id where e.confirmed='t' and e.deleted_at is null;">./to.txt
wc -l to.txt

docker rm -f $container_name 2>/dev/null
