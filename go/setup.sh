# source ./setup.sh

flatc -g ../flatbuffer/data.fbs
uniffi-bindgen-go ../src/interface.udl
rm -rf maya_zcash
mv ../src/maya_zcash .
cp ../config.yaml .
export CGO_LDFLAGS=" -lmaya_zcash -L../target/release"
