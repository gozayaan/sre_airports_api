## Sets up 

## Make sure docker daemon is up and running

Write-Host 'Setting up environment 🔨'

$AIRPORTS_API_IMAGE_TAG='v1.7'
$GCS_STORAGE_MOUNT_PATH="D:\Your\Storage\Path\"

docker network create airport-net

# setup fake-gcs-server; 4443 https, 8000 http;
docker run -d --name fake-gcs-server --network airport-net -p 4443:4443 -p 8000:8000 -v ${GCS_STORAGE_MOUNT_PATH}:/data fsouza/fake-gcs-server -scheme both -public-host localhost

Write-Host 'GCS mock server up at localhost:4443 & 8000 🚀 !'

# setup bd-airports server; 8080 port
# use container name of gcs server for connectivity
docker run -d --name bd-airport-api --network airport-net -p 8080:8080 -e STORAGE_EMULATOR_HOST='http://fake-gcs-server:4443' -e GCS_BUCKET_NAME='dev_bd_airport' -e GCS_LOCALHOST_URL='http://fake-gcs-server:8000/storage/v1/' -e GCS_BUCKET_DOMAIN='storage.googleapis.com' bijoy26/bd-airports:$AIRPORTS_API_IMAGE_TAG

Write-Host 'BD Airports API up at localhost:4443 🚀 !'
