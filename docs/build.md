# Quick Start

- Build: `cargo b -r`
- In the `go` directory,
    - `export CGO_LDFLAGS=" -lmaya_zcash -L../target/release"`
    - `cp ../config.yaml .`
    - Ensure your `zcashd` server is running and
    **update the config file** with the IP, username and password
    - Build the example code: `go build main.go`
    - Run: `./main`

- To run on linux, make sure the library `libmaya_zcash.so`
is in the `LD_LIBRARY_PATH`

# Misc
## Flatbuffers
- Need `flatc` from the flatbuffers project

## Go Binding Generation
[uniffi-bindgen-go](https://github.com/NordSecurity/uniffi-bindgen-go)

```
$ cargo install uniffi-bindgen-go --git https://github.com/NordSecurity/uniffi-bindgen-go --tag v0.2.2+v0.25.0
$ uniffi-bindgen-go src/interface.udl
```

## Regenerate code
- To generate the go bindings, use `setup.sh`
- To generate the flatbuffer rust serializers, run
    `flatc -r --gen-object-api ../flatbuffer/data.fbs`

