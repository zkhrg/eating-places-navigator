# Pull images
docker pull elasticsearch:8.4.2
docker pull golang:1.22

# Create network and volume
docker network create somenetwork
docker volume create elasticsearch_data

# Run Elasticsearch container
CONTAINER_ID=$(docker run -d --name elasticsearch --net somenetwork -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" -v elasticsearch_data:/usr/share/elasticsearch/data elasticsearch:8.4.2)

# Проверка на успешный запуск контейнера
if [ -z "$CONTAINER_ID" ]; then
  echo "Failed to start Elasticsearch container"
  exit 1
fi

# Даем Elasticsearch время для запуска
sleep 20

# Reset the password and retrieve the cert
ELASTIC_PASSWORD=$(docker exec -it $CONTAINER_ID bash -c 'echo "y" | bin/elasticsearch-reset-password -u elastic')

ES_CERT=$(docker exec $CONTAINER_ID cat /usr/share/elasticsearch/config/certs/http_ca.crt)
echo "$ES_CERT"

# Extract the new password
ELASTIC_PASSWORD=$(echo $ELASTIC_PASSWORD | sed -n 's/^.*New value: \([^ ]*\).*/\1/p')
echo "$ELASTIC_PASSWORD"

# Даем немного времени перед сборкой
sleep 5

# Build the Go application
echo "Building Go application..."
docker build -t go_day03_server .
echo "Build completed."

# Запуск контейнера Go приложения
echo "Running Go application container..."
docker run --net somenetwork \
  -e ES_USERNAME=elastic \
  -e ES_PASSWORD=$ELASTIC_PASSWORD \
  -e ES_ADDRESS=https://elasticsearch:9200 \
  -e SECRET_KEY="asdfgdsrstere123123" \
  -e ENV="local" \
  -e APP_VERSION="v1.0.0" \
  -e APP_NAME="go_day03" \
  -e ES_CERT_CONTENT="$ES_CERT" \
  -p 8888:8888 go_day03_server

# Проверка на успешный запуск Go контейнера
if [ $? -ne 0 ]; then
  echo "Failed to run Go application container"
  exit 1
fi
