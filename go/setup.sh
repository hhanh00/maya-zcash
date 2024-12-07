# source ./setup.sh

uniffi-bindgen-go ../src/interface.udl
cp ../src/maya_zcash/* maya_zcash/
cp ../config.yaml .
export CGO_LDFLAGS=" -lmaya_zcash -L../target/release"
