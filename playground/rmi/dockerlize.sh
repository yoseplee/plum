version=$1
container_name='peer'
image_name='peer'
echo "version: $version"

docker rmi -f $image_name:$version
docker build --tag $image_name:$version -f ../../dockerfile/peer/Dockerfile .

docker stop $container_name 
docker rm $container_name 
docker run -t --name $container_name -d $image_name:$version
echo "dockerize done on version $version"
