docker compose down

docker rm $(docker ps -a -q) || true

docker network rm shared_network || true

docker network create shared_network

docker compose build

docker compose up