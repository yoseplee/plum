version=$1
image_name='plum'
echo "build version: $version"

cd ../../src/peer/rmi/
docker build --tag $image_name:$version -f ../../../dockerfile/plum/Dockerfile .
