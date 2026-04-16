docker compose down -v

docker rm $(docker ps -a -q) || true

docker network rm shared_network || true

docker network create shared_network

docker compose build --no-cache

docker compose up