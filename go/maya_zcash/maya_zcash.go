package maya_zcash

// #include <maya_zcash.h>
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"unsafe"
)

type RustBuffer = C.RustBuffer

type RustBufferI interface {
	AsReader() *bytes.Reader
	Free()
	ToGoBytes() []byte
	Data() unsafe.Pointer
	Len() int
	Capacity() int
}

func RustBufferFromExternal(b RustBufferI) RustBuffer {
	return RustBuffer{
		capacity: C.int(b.Capacity()),
		len:      C.int(b.Len()),
		data:     (*C.uchar)(b.Data()),
	}
}

func (cb RustBuffer) Capacity() int {
	return int(cb.capacity)
}

func (cb RustBuffer) Len() int {
	return int(cb.len)
}

func (cb RustBuffer) Data() unsafe.Pointer {
	return unsafe.Pointer(cb.data)
}

func (cb RustBuffer) AsReader() *bytes.Reader {
	b := unsafe.Slice((*byte)(cb.data), C.int(cb.len))
	return bytes.NewReader(b)
}

func (cb RustBuffer) Free() {
	rustCall(func(status *C.RustCallStatus) bool {
		C.ffi_maya_zcash_rustbuffer_free(cb, status)
		return false
	})
}

func (cb RustBuffer) ToGoBytes() []byte {
	return C.GoBytes(unsafe.Pointer(cb.data), C.int(cb.len))
}

func stringToRustBuffer(str string) RustBuffer {
	return bytesToRustBuffer([]byte(str))
}

func bytesToRustBuffer(b []byte) RustBuffer {
	if len(b) == 0 {
		return RustBuffer{}
	}
	// We can pass the pointer along here, as it is pinned
	// for the duration of this call
	foreign := C.ForeignBytes{
		len:  C.int(len(b)),
		data: (*C.uchar)(unsafe.Pointer(&b[0])),
	}

	return rustCall(func(status *C.RustCallStatus) RustBuffer {
		return C.ffi_maya_zcash_rustbuffer_from_bytes(foreign, status)
	})
}

type BufLifter[GoType any] interface {
	Lift(value RustBufferI) GoType
}

type BufLowerer[GoType any] interface {
	Lower(value GoType) RustBuffer
}

type FfiConverter[GoType any, FfiType any] interface {
	Lift(value FfiType) GoType
	Lower(value GoType) FfiType
}

type BufReader[GoType any] interface {
	Read(reader io.Reader) GoType
}

type BufWriter[GoType any] interface {
	Write(writer io.Writer, value GoType)
}

type FfiRustBufConverter[GoType any, FfiType any] interface {
	FfiConverter[GoType, FfiType]
	BufReader[GoType]
}

func LowerIntoRustBuffer[GoType any](bufWriter BufWriter[GoType], value GoType) RustBuffer {
	// This might be not the most efficient way but it does not require knowing allocation size
	// beforehand
	var buffer bytes.Buffer
	bufWriter.Write(&buffer, value)

	bytes, err := io.ReadAll(&buffer)
	if err != nil {
		panic(fmt.Errorf("reading written data: %w", err))
	}
	return bytesToRustBuffer(bytes)
}

func LiftFromRustBuffer[GoType any](bufReader BufReader[GoType], rbuf RustBufferI) GoType {
	defer rbuf.Free()
	reader := rbuf.AsReader()
	item := bufReader.Read(reader)
	if reader.Len() > 0 {
		// TODO: Remove this
		leftover, _ := io.ReadAll(reader)
		panic(fmt.Errorf("Junk remaining in buffer after lifting: %s", string(leftover)))
	}
	return item
}

func rustCallWithError[U any](converter BufLifter[error], callback func(*C.RustCallStatus) U) (U, error) {
	var status C.RustCallStatus
	returnValue := callback(&status)
	err := checkCallStatus(converter, status)

	return returnValue, err
}

func checkCallStatus(converter BufLifter[error], status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		return converter.Lift(status.errorBuf)
	case 2:
		// when the rust code sees a panic, it tries to construct a rustbuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(status.errorBuf)))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func checkCallStatusUnknown(status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		panic(fmt.Errorf("function not returning an error returned an error"))
	case 2:
		// when the rust code sees a panic, it tries to construct a rustbuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(status.errorBuf)))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func rustCall[U any](callback func(*C.RustCallStatus) U) U {
	returnValue, err := rustCallWithError(nil, callback)
	if err != nil {
		panic(err)
	}
	return returnValue
}

func writeInt8(writer io.Writer, value int8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint8(writer io.Writer, value uint8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt16(writer io.Writer, value int16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint16(writer io.Writer, value uint16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt32(writer io.Writer, value int32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint32(writer io.Writer, value uint32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt64(writer io.Writer, value int64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint64(writer io.Writer, value uint64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat32(writer io.Writer, value float32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat64(writer io.Writer, value float64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func readInt8(reader io.Reader) int8 {
	var result int8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint8(reader io.Reader) uint8 {
	var result uint8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt16(reader io.Reader) int16 {
	var result int16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint16(reader io.Reader) uint16 {
	var result uint16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt32(reader io.Reader) int32 {
	var result int32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint32(reader io.Reader) uint32 {
	var result uint32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt64(reader io.Reader) int64 {
	var result int64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint64(reader io.Reader) uint64 {
	var result uint64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat32(reader io.Reader) float32 {
	var result float32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat64(reader io.Reader) float64 {
	var result float64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func init() {

	uniffiCheckChecksums()
}

func uniffiCheckChecksums() {
	// Get the bindings contract version from our ComponentInterface
	bindingsContractVersion := 24
	// Get the scaffolding contract version by calling the into the dylib
	scaffoldingContractVersion := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint32_t {
		return C.ffi_maya_zcash_uniffi_contract_version(uniffiStatus)
	})
	if bindingsContractVersion != int(scaffoldingContractVersion) {
		// If this happens try cleaning and rebuilding your project
		panic("maya_zcash: UniFFI contract version mismatch")
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_maya_zcash_checksum_func_get_latest_height(uniffiStatus)
		})
		if checksum != 41262 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_get_latest_height: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_maya_zcash_checksum_func_get_vault_address(uniffiStatus)
		})
		if checksum != 10814 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_get_vault_address: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_maya_zcash_checksum_func_init_logger(uniffiStatus)
		})
		if checksum != 363 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_init_logger: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_maya_zcash_checksum_func_match_with_blockchain_receiver(uniffiStatus)
		})
		if checksum != 55511 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_match_with_blockchain_receiver: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_maya_zcash_checksum_func_validate_address(uniffiStatus)
		})
		if checksum != 64411 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_validate_address: UniFFI API checksum mismatch")
		}
	}
}

type FfiConverterUint32 struct{}

var FfiConverterUint32INSTANCE = FfiConverterUint32{}

func (FfiConverterUint32) Lower(value uint32) C.uint32_t {
	return C.uint32_t(value)
}

func (FfiConverterUint32) Write(writer io.Writer, value uint32) {
	writeUint32(writer, value)
}

func (FfiConverterUint32) Lift(value C.uint32_t) uint32 {
	return uint32(value)
}

func (FfiConverterUint32) Read(reader io.Reader) uint32 {
	return readUint32(reader)
}

type FfiDestroyerUint32 struct{}

func (FfiDestroyerUint32) Destroy(_ uint32) {}

type FfiConverterBool struct{}

var FfiConverterBoolINSTANCE = FfiConverterBool{}

func (FfiConverterBool) Lower(value bool) C.int8_t {
	if value {
		return C.int8_t(1)
	}
	return C.int8_t(0)
}

func (FfiConverterBool) Write(writer io.Writer, value bool) {
	if value {
		writeInt8(writer, 1)
	} else {
		writeInt8(writer, 0)
	}
}

func (FfiConverterBool) Lift(value C.int8_t) bool {
	return value != 0
}

func (FfiConverterBool) Read(reader io.Reader) bool {
	return readInt8(reader) != 0
}

type FfiDestroyerBool struct{}

func (FfiDestroyerBool) Destroy(_ bool) {}

type FfiConverterString struct{}

var FfiConverterStringINSTANCE = FfiConverterString{}

func (FfiConverterString) Lift(rb RustBufferI) string {
	defer rb.Free()
	reader := rb.AsReader()
	b, err := io.ReadAll(reader)
	if err != nil {
		panic(fmt.Errorf("reading reader: %w", err))
	}
	return string(b)
}

func (FfiConverterString) Read(reader io.Reader) string {
	length := readInt32(reader)
	buffer := make([]byte, length)
	read_length, err := reader.Read(buffer)
	if err != nil {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading string, expected %d, read %d", length, read_length))
	}
	return string(buffer)
}

func (FfiConverterString) Lower(value string) RustBuffer {
	return stringToRustBuffer(value)
}

func (FfiConverterString) Write(writer io.Writer, value string) {
	if len(value) > math.MaxInt32 {
		panic("String is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	write_length, err := io.WriteString(writer, value)
	if err != nil {
		panic(err)
	}
	if write_length != len(value) {
		panic(fmt.Errorf("bad write length when writing string, expected %d, written %d", len(value), write_length))
	}
}

type FfiDestroyerString struct{}

func (FfiDestroyerString) Destroy(_ string) {}

type FfiConverterBytes struct{}

var FfiConverterBytesINSTANCE = FfiConverterBytes{}

func (c FfiConverterBytes) Lower(value []byte) RustBuffer {
	return LowerIntoRustBuffer[[]byte](c, value)
}

func (c FfiConverterBytes) Write(writer io.Writer, value []byte) {
	if len(value) > math.MaxInt32 {
		panic("[]byte is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	write_length, err := writer.Write(value)
	if err != nil {
		panic(err)
	}
	if write_length != len(value) {
		panic(fmt.Errorf("bad write length when writing []byte, expected %d, written %d", len(value), write_length))
	}
}

func (c FfiConverterBytes) Lift(rb RustBufferI) []byte {
	return LiftFromRustBuffer[[]byte](c, rb)
}

func (c FfiConverterBytes) Read(reader io.Reader) []byte {
	length := readInt32(reader)
	buffer := make([]byte, length)
	read_length, err := reader.Read(buffer)
	if err != nil {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading []byte, expected %d, read %d", length, read_length))
	}
	return buffer
}

type FfiDestroyerBytes struct{}

func (FfiDestroyerBytes) Destroy(_ []byte) {}

type Height struct {
	Number uint32
	Hash   []byte
}

func (r *Height) Destroy() {
	FfiDestroyerUint32{}.Destroy(r.Number)
	FfiDestroyerBytes{}.Destroy(r.Hash)
}

type FfiConverterTypeHeight struct{}

var FfiConverterTypeHeightINSTANCE = FfiConverterTypeHeight{}

func (c FfiConverterTypeHeight) Lift(rb RustBufferI) Height {
	return LiftFromRustBuffer[Height](c, rb)
}

func (c FfiConverterTypeHeight) Read(reader io.Reader) Height {
	return Height{
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeHeight) Lower(value Height) RustBuffer {
	return LowerIntoRustBuffer[Height](c, value)
}

func (c FfiConverterTypeHeight) Write(writer io.Writer, value Height) {
	FfiConverterUint32INSTANCE.Write(writer, value.Number)
	FfiConverterBytesINSTANCE.Write(writer, value.Hash)
}

type FfiDestroyerTypeHeight struct{}

func (_ FfiDestroyerTypeHeight) Destroy(value Height) {
	value.Destroy()
}

type ZcashError struct {
	err error
}

func (err ZcashError) Error() string {
	return fmt.Sprintf("ZcashError: %s", err.err.Error())
}

func (err ZcashError) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrZcashErrorRpc = fmt.Errorf("ZcashErrorRpc")
var ErrZcashErrorInvalidPubkeyLength = fmt.Errorf("ZcashErrorInvalidPubkeyLength")
var ErrZcashErrorInvalidAddress = fmt.Errorf("ZcashErrorInvalidAddress")
var ErrZcashErrorNoOrchardReceiver = fmt.Errorf("ZcashErrorNoOrchardReceiver")
var ErrZcashErrorAssertError = fmt.Errorf("ZcashErrorAssertError")

// Variant structs
type ZcashErrorRpc struct {
	message string
}

func NewZcashErrorRpc() *ZcashError {
	return &ZcashError{
		err: &ZcashErrorRpc{},
	}
}

func (err ZcashErrorRpc) Error() string {
	return fmt.Sprintf("Rpc: %s", err.message)
}

func (self ZcashErrorRpc) Is(target error) bool {
	return target == ErrZcashErrorRpc
}

type ZcashErrorInvalidPubkeyLength struct {
	message string
}

func NewZcashErrorInvalidPubkeyLength() *ZcashError {
	return &ZcashError{
		err: &ZcashErrorInvalidPubkeyLength{},
	}
}

func (err ZcashErrorInvalidPubkeyLength) Error() string {
	return fmt.Sprintf("InvalidPubkeyLength: %s", err.message)
}

func (self ZcashErrorInvalidPubkeyLength) Is(target error) bool {
	return target == ErrZcashErrorInvalidPubkeyLength
}

type ZcashErrorInvalidAddress struct {
	message string
}

func NewZcashErrorInvalidAddress() *ZcashError {
	return &ZcashError{
		err: &ZcashErrorInvalidAddress{},
	}
}

func (err ZcashErrorInvalidAddress) Error() string {
	return fmt.Sprintf("InvalidAddress: %s", err.message)
}

func (self ZcashErrorInvalidAddress) Is(target error) bool {
	return target == ErrZcashErrorInvalidAddress
}

type ZcashErrorNoOrchardReceiver struct {
	message string
}

func NewZcashErrorNoOrchardReceiver() *ZcashError {
	return &ZcashError{
		err: &ZcashErrorNoOrchardReceiver{},
	}
}

func (err ZcashErrorNoOrchardReceiver) Error() string {
	return fmt.Sprintf("NoOrchardReceiver: %s", err.message)
}

func (self ZcashErrorNoOrchardReceiver) Is(target error) bool {
	return target == ErrZcashErrorNoOrchardReceiver
}

type ZcashErrorAssertError struct {
	message string
}

func NewZcashErrorAssertError() *ZcashError {
	return &ZcashError{
		err: &ZcashErrorAssertError{},
	}
}

func (err ZcashErrorAssertError) Error() string {
	return fmt.Sprintf("AssertError: %s", err.message)
}

func (self ZcashErrorAssertError) Is(target error) bool {
	return target == ErrZcashErrorAssertError
}

type FfiConverterTypeZcashError struct{}

var FfiConverterTypeZcashErrorINSTANCE = FfiConverterTypeZcashError{}

func (c FfiConverterTypeZcashError) Lift(eb RustBufferI) error {
	return LiftFromRustBuffer[*ZcashError](c, eb)
}

func (c FfiConverterTypeZcashError) Lower(value *ZcashError) RustBuffer {
	return LowerIntoRustBuffer[*ZcashError](c, value)
}

func (c FfiConverterTypeZcashError) Read(reader io.Reader) *ZcashError {
	errorID := readUint32(reader)

	message := FfiConverterStringINSTANCE.Read(reader)
	switch errorID {
	case 1:
		return &ZcashError{&ZcashErrorRpc{message}}
	case 2:
		return &ZcashError{&ZcashErrorInvalidPubkeyLength{message}}
	case 3:
		return &ZcashError{&ZcashErrorInvalidAddress{message}}
	case 4:
		return &ZcashError{&ZcashErrorNoOrchardReceiver{message}}
	case 5:
		return &ZcashError{&ZcashErrorAssertError{message}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeZcashError.Read()", errorID))
	}

}

func (c FfiConverterTypeZcashError) Write(writer io.Writer, value *ZcashError) {
	switch variantValue := value.err.(type) {
	case *ZcashErrorRpc:
		writeInt32(writer, 1)
	case *ZcashErrorInvalidPubkeyLength:
		writeInt32(writer, 2)
	case *ZcashErrorInvalidAddress:
		writeInt32(writer, 3)
	case *ZcashErrorNoOrchardReceiver:
		writeInt32(writer, 4)
	case *ZcashErrorAssertError:
		writeInt32(writer, 5)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeZcashError.Write", value))
	}
}

func GetLatestHeight() (Height, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_maya_zcash_fn_func_get_latest_height(_uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Height
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeHeightINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func GetVaultAddress(pubkey []byte) (string, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_maya_zcash_fn_func_get_vault_address(FfiConverterBytesINSTANCE.Lower(pubkey), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue string
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterStringINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func InitLogger() {
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_maya_zcash_fn_func_init_logger(_uniffiStatus)
		return false
	})
}

func MatchWithBlockchainReceiver(address string, receiver string) (bool, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_maya_zcash_fn_func_match_with_blockchain_receiver(FfiConverterStringINSTANCE.Lower(address), FfiConverterStringINSTANCE.Lower(receiver), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue bool
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterBoolINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func ValidateAddress(address string) (bool, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_maya_zcash_fn_func_validate_address(FfiConverterStringINSTANCE.Lower(address), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue bool
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterBoolINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}
