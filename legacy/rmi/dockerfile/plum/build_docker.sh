version=$1
image_name='yoseplee/plum'
echo "build version: $version"

cd ../../src/peer/rmi/
docker rmi -f $image_name:$version
docker build --tag $image_name:$version -f ../../../dockerfile/plum/Dockerfile .
