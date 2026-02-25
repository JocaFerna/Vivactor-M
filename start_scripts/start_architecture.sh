docker compose down

docker network rm shared_network || true

docker network create shared_network

docker compose build

docker compose up