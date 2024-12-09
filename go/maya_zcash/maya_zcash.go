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
			return C.uniffi_maya_zcash_checksum_func_apply_signatures(uniffiStatus)
		})
		if checksum != 24461 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_apply_signatures: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_maya_zcash_checksum_func_broadcast_raw_tx(uniffiStatus)
		})
		if checksum != 14042 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_broadcast_raw_tx: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_maya_zcash_checksum_func_build_vault_unauthorized_tx(uniffiStatus)
		})
		if checksum != 928 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_build_vault_unauthorized_tx: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_maya_zcash_checksum_func_combine_vault(uniffiStatus)
		})
		if checksum != 25110 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_combine_vault: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_maya_zcash_checksum_func_combine_vault_utxos(uniffiStatus)
		})
		if checksum != 39573 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_combine_vault_utxos: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_maya_zcash_checksum_func_get_balance(uniffiStatus)
		})
		if checksum != 16973 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_get_balance: UniFFI API checksum mismatch")
		}
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
			return C.uniffi_maya_zcash_checksum_func_get_ovk(uniffiStatus)
		})
		if checksum != 17238 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_get_ovk: UniFFI API checksum mismatch")
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
			return C.uniffi_maya_zcash_checksum_func_list_utxos(uniffiStatus)
		})
		if checksum != 39673 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_list_utxos: UniFFI API checksum mismatch")
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
			return C.uniffi_maya_zcash_checksum_func_pay_from_vault(uniffiStatus)
		})
		if checksum != 56589 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_pay_from_vault: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_maya_zcash_checksum_func_scan_blocks(uniffiStatus)
		})
		if checksum != 29804 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_scan_blocks: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_maya_zcash_checksum_func_scan_mempool(uniffiStatus)
		})
		if checksum != 2161 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_scan_mempool: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_maya_zcash_checksum_func_send_to_vault(uniffiStatus)
		})
		if checksum != 12684 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_send_to_vault: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_maya_zcash_checksum_func_sign_sighash(uniffiStatus)
		})
		if checksum != 29344 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_sign_sighash: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_maya_zcash_checksum_func_sk_to_pub(uniffiStatus)
		})
		if checksum != 14751 {
			// If this happens try cleaning and rebuilding your project
			panic("maya_zcash: uniffi_maya_zcash_checksum_func_sk_to_pub: UniFFI API checksum mismatch")
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

type FfiConverterUint64 struct{}

var FfiConverterUint64INSTANCE = FfiConverterUint64{}

func (FfiConverterUint64) Lower(value uint64) C.uint64_t {
	return C.uint64_t(value)
}

func (FfiConverterUint64) Write(writer io.Writer, value uint64) {
	writeUint64(writer, value)
}

func (FfiConverterUint64) Lift(value C.uint64_t) uint64 {
	return uint64(value)
}

func (FfiConverterUint64) Read(reader io.Reader) uint64 {
	return readUint64(reader)
}

type FfiDestroyerUint64 struct{}

func (FfiDestroyerUint64) Destroy(_ uint64) {}

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

type BlockTxs struct {
	StartHash   string
	EndHash     string
	StartHeight uint32
	EndHeight   uint32
	Txs         []VaultTx
}

func (r *BlockTxs) Destroy() {
	FfiDestroyerString{}.Destroy(r.StartHash)
	FfiDestroyerString{}.Destroy(r.EndHash)
	FfiDestroyerUint32{}.Destroy(r.StartHeight)
	FfiDestroyerUint32{}.Destroy(r.EndHeight)
	FfiDestroyerSequenceTypeVaultTx{}.Destroy(r.Txs)
}

type FfiConverterTypeBlockTxs struct{}

var FfiConverterTypeBlockTxsINSTANCE = FfiConverterTypeBlockTxs{}

func (c FfiConverterTypeBlockTxs) Lift(rb RustBufferI) BlockTxs {
	return LiftFromRustBuffer[BlockTxs](c, rb)
}

func (c FfiConverterTypeBlockTxs) Read(reader io.Reader) BlockTxs {
	return BlockTxs{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterSequenceTypeVaultTxINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeBlockTxs) Lower(value BlockTxs) RustBuffer {
	return LowerIntoRustBuffer[BlockTxs](c, value)
}

func (c FfiConverterTypeBlockTxs) Write(writer io.Writer, value BlockTxs) {
	FfiConverterStringINSTANCE.Write(writer, value.StartHash)
	FfiConverterStringINSTANCE.Write(writer, value.EndHash)
	FfiConverterUint32INSTANCE.Write(writer, value.StartHeight)
	FfiConverterUint32INSTANCE.Write(writer, value.EndHeight)
	FfiConverterSequenceTypeVaultTxINSTANCE.Write(writer, value.Txs)
}

type FfiDestroyerTypeBlockTxs struct{}

func (_ FfiDestroyerTypeBlockTxs) Destroy(value BlockTxs) {
	value.Destroy()
}

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

type Output struct {
	Address string
	Amount  uint64
	Memo    string
}

func (r *Output) Destroy() {
	FfiDestroyerString{}.Destroy(r.Address)
	FfiDestroyerUint64{}.Destroy(r.Amount)
	FfiDestroyerString{}.Destroy(r.Memo)
}

type FfiConverterTypeOutput struct{}

var FfiConverterTypeOutputINSTANCE = FfiConverterTypeOutput{}

func (c FfiConverterTypeOutput) Lift(rb RustBufferI) Output {
	return LiftFromRustBuffer[Output](c, rb)
}

func (c FfiConverterTypeOutput) Read(reader io.Reader) Output {
	return Output{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeOutput) Lower(value Output) RustBuffer {
	return LowerIntoRustBuffer[Output](c, value)
}

func (c FfiConverterTypeOutput) Write(writer io.Writer, value Output) {
	FfiConverterStringINSTANCE.Write(writer, value.Address)
	FfiConverterUint64INSTANCE.Write(writer, value.Amount)
	FfiConverterStringINSTANCE.Write(writer, value.Memo)
}

type FfiDestroyerTypeOutput struct{}

func (_ FfiDestroyerTypeOutput) Destroy(value Output) {
	value.Destroy()
}

type PartialTx struct {
	Height  uint32
	Inputs  []Utxo
	Outputs []Output
	Fee     uint64
	TxSeed  []byte
}

func (r *PartialTx) Destroy() {
	FfiDestroyerUint32{}.Destroy(r.Height)
	FfiDestroyerSequenceTypeUtxo{}.Destroy(r.Inputs)
	FfiDestroyerSequenceTypeOutput{}.Destroy(r.Outputs)
	FfiDestroyerUint64{}.Destroy(r.Fee)
	FfiDestroyerBytes{}.Destroy(r.TxSeed)
}

type FfiConverterTypePartialTx struct{}

var FfiConverterTypePartialTxINSTANCE = FfiConverterTypePartialTx{}

func (c FfiConverterTypePartialTx) Lift(rb RustBufferI) PartialTx {
	return LiftFromRustBuffer[PartialTx](c, rb)
}

func (c FfiConverterTypePartialTx) Read(reader io.Reader) PartialTx {
	return PartialTx{
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterSequenceTypeUTXOINSTANCE.Read(reader),
		FfiConverterSequenceTypeOutputINSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypePartialTx) Lower(value PartialTx) RustBuffer {
	return LowerIntoRustBuffer[PartialTx](c, value)
}

func (c FfiConverterTypePartialTx) Write(writer io.Writer, value PartialTx) {
	FfiConverterUint32INSTANCE.Write(writer, value.Height)
	FfiConverterSequenceTypeUTXOINSTANCE.Write(writer, value.Inputs)
	FfiConverterSequenceTypeOutputINSTANCE.Write(writer, value.Outputs)
	FfiConverterUint64INSTANCE.Write(writer, value.Fee)
	FfiConverterBytesINSTANCE.Write(writer, value.TxSeed)
}

type FfiDestroyerTypePartialTx struct{}

func (_ FfiDestroyerTypePartialTx) Destroy(value PartialTx) {
	value.Destroy()
}

type Sighashes struct {
	Hashes [][]byte
}

func (r *Sighashes) Destroy() {
	FfiDestroyerSequenceBytes{}.Destroy(r.Hashes)
}

type FfiConverterTypeSighashes struct{}

var FfiConverterTypeSighashesINSTANCE = FfiConverterTypeSighashes{}

func (c FfiConverterTypeSighashes) Lift(rb RustBufferI) Sighashes {
	return LiftFromRustBuffer[Sighashes](c, rb)
}

func (c FfiConverterTypeSighashes) Read(reader io.Reader) Sighashes {
	return Sighashes{
		FfiConverterSequenceBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeSighashes) Lower(value Sighashes) RustBuffer {
	return LowerIntoRustBuffer[Sighashes](c, value)
}

func (c FfiConverterTypeSighashes) Write(writer io.Writer, value Sighashes) {
	FfiConverterSequenceBytesINSTANCE.Write(writer, value.Hashes)
}

type FfiDestroyerTypeSighashes struct{}

func (_ FfiDestroyerTypeSighashes) Destroy(value Sighashes) {
	value.Destroy()
}

type TransparentKey struct {
	Sk   []byte
	Pk   []byte
	Addr string
}

func (r *TransparentKey) Destroy() {
	FfiDestroyerBytes{}.Destroy(r.Sk)
	FfiDestroyerBytes{}.Destroy(r.Pk)
	FfiDestroyerString{}.Destroy(r.Addr)
}

type FfiConverterTypeTransparentKey struct{}

var FfiConverterTypeTransparentKeyINSTANCE = FfiConverterTypeTransparentKey{}

func (c FfiConverterTypeTransparentKey) Lift(rb RustBufferI) TransparentKey {
	return LiftFromRustBuffer[TransparentKey](c, rb)
}

func (c FfiConverterTypeTransparentKey) Read(reader io.Reader) TransparentKey {
	return TransparentKey{
		FfiConverterBytesINSTANCE.Read(reader),
		FfiConverterBytesINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTransparentKey) Lower(value TransparentKey) RustBuffer {
	return LowerIntoRustBuffer[TransparentKey](c, value)
}

func (c FfiConverterTypeTransparentKey) Write(writer io.Writer, value TransparentKey) {
	FfiConverterBytesINSTANCE.Write(writer, value.Sk)
	FfiConverterBytesINSTANCE.Write(writer, value.Pk)
	FfiConverterStringINSTANCE.Write(writer, value.Addr)
}

type FfiDestroyerTypeTransparentKey struct{}

func (_ FfiDestroyerTypeTransparentKey) Destroy(value TransparentKey) {
	value.Destroy()
}

type TxBytes struct {
	Txid string
	Data []byte
}

func (r *TxBytes) Destroy() {
	FfiDestroyerString{}.Destroy(r.Txid)
	FfiDestroyerBytes{}.Destroy(r.Data)
}

type FfiConverterTypeTxBytes struct{}

var FfiConverterTypeTxBytesINSTANCE = FfiConverterTypeTxBytes{}

func (c FfiConverterTypeTxBytes) Lift(rb RustBufferI) TxBytes {
	return LiftFromRustBuffer[TxBytes](c, rb)
}

func (c FfiConverterTypeTxBytes) Read(reader io.Reader) TxBytes {
	return TxBytes{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterBytesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTxBytes) Lower(value TxBytes) RustBuffer {
	return LowerIntoRustBuffer[TxBytes](c, value)
}

func (c FfiConverterTypeTxBytes) Write(writer io.Writer, value TxBytes) {
	FfiConverterStringINSTANCE.Write(writer, value.Txid)
	FfiConverterBytesINSTANCE.Write(writer, value.Data)
}

type FfiDestroyerTypeTxBytes struct{}

func (_ FfiDestroyerTypeTxBytes) Destroy(value TxBytes) {
	value.Destroy()
}

type Utxo struct {
	Txid   string
	Height uint32
	Vout   uint32
	Script string
	Value  uint64
}

func (r *Utxo) Destroy() {
	FfiDestroyerString{}.Destroy(r.Txid)
	FfiDestroyerUint32{}.Destroy(r.Height)
	FfiDestroyerUint32{}.Destroy(r.Vout)
	FfiDestroyerString{}.Destroy(r.Script)
	FfiDestroyerUint64{}.Destroy(r.Value)
}

type FfiConverterTypeUTXO struct{}

var FfiConverterTypeUTXOINSTANCE = FfiConverterTypeUTXO{}

func (c FfiConverterTypeUTXO) Lift(rb RustBufferI) Utxo {
	return LiftFromRustBuffer[Utxo](c, rb)
}

func (c FfiConverterTypeUTXO) Read(reader io.Reader) Utxo {
	return Utxo{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeUTXO) Lower(value Utxo) RustBuffer {
	return LowerIntoRustBuffer[Utxo](c, value)
}

func (c FfiConverterTypeUTXO) Write(writer io.Writer, value Utxo) {
	FfiConverterStringINSTANCE.Write(writer, value.Txid)
	FfiConverterUint32INSTANCE.Write(writer, value.Height)
	FfiConverterUint32INSTANCE.Write(writer, value.Vout)
	FfiConverterStringINSTANCE.Write(writer, value.Script)
	FfiConverterUint64INSTANCE.Write(writer, value.Value)
}

type FfiDestroyerTypeUtxo struct{}

func (_ FfiDestroyerTypeUtxo) Destroy(value Utxo) {
	value.Destroy()
}

type VaultTx struct {
	Txid         string
	Height       uint32
	Counterparty Output
	Direction    Direction
}

func (r *VaultTx) Destroy() {
	FfiDestroyerString{}.Destroy(r.Txid)
	FfiDestroyerUint32{}.Destroy(r.Height)
	FfiDestroyerTypeOutput{}.Destroy(r.Counterparty)
	FfiDestroyerTypeDirection{}.Destroy(r.Direction)
}

type FfiConverterTypeVaultTx struct{}

var FfiConverterTypeVaultTxINSTANCE = FfiConverterTypeVaultTx{}

func (c FfiConverterTypeVaultTx) Lift(rb RustBufferI) VaultTx {
	return LiftFromRustBuffer[VaultTx](c, rb)
}

func (c FfiConverterTypeVaultTx) Read(reader io.Reader) VaultTx {
	return VaultTx{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterTypeOutputINSTANCE.Read(reader),
		FfiConverterTypeDirectionINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeVaultTx) Lower(value VaultTx) RustBuffer {
	return LowerIntoRustBuffer[VaultTx](c, value)
}

func (c FfiConverterTypeVaultTx) Write(writer io.Writer, value VaultTx) {
	FfiConverterStringINSTANCE.Write(writer, value.Txid)
	FfiConverterUint32INSTANCE.Write(writer, value.Height)
	FfiConverterTypeOutputINSTANCE.Write(writer, value.Counterparty)
	FfiConverterTypeDirectionINSTANCE.Write(writer, value.Direction)
}

type FfiDestroyerTypeVaultTx struct{}

func (_ FfiDestroyerTypeVaultTx) Destroy(value VaultTx) {
	value.Destroy()
}

type Direction uint

const (
	DirectionIncoming Direction = 1
	DirectionOutgoing Direction = 2
)

type FfiConverterTypeDirection struct{}

var FfiConverterTypeDirectionINSTANCE = FfiConverterTypeDirection{}

func (c FfiConverterTypeDirection) Lift(rb RustBufferI) Direction {
	return LiftFromRustBuffer[Direction](c, rb)
}

func (c FfiConverterTypeDirection) Lower(value Direction) RustBuffer {
	return LowerIntoRustBuffer[Direction](c, value)
}
func (FfiConverterTypeDirection) Read(reader io.Reader) Direction {
	id := readInt32(reader)
	return Direction(id)
}

func (FfiConverterTypeDirection) Write(writer io.Writer, value Direction) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeDirection struct{}

func (_ FfiDestroyerTypeDirection) Destroy(value Direction) {
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
var ErrZcashErrorInvalidVaultPubkey = fmt.Errorf("ZcashErrorInvalidVaultPubkey")
var ErrZcashErrorInvalidAddress = fmt.Errorf("ZcashErrorInvalidAddress")
var ErrZcashErrorNoOrchardReceiver = fmt.Errorf("ZcashErrorNoOrchardReceiver")
var ErrZcashErrorNotEnoughFunds = fmt.Errorf("ZcashErrorNotEnoughFunds")
var ErrZcashErrorTxRejected = fmt.Errorf("ZcashErrorTxRejected")
var ErrZcashErrorReorg = fmt.Errorf("ZcashErrorReorg")
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

type ZcashErrorInvalidVaultPubkey struct {
	message string
}

func NewZcashErrorInvalidVaultPubkey() *ZcashError {
	return &ZcashError{
		err: &ZcashErrorInvalidVaultPubkey{},
	}
}

func (err ZcashErrorInvalidVaultPubkey) Error() string {
	return fmt.Sprintf("InvalidVaultPubkey: %s", err.message)
}

func (self ZcashErrorInvalidVaultPubkey) Is(target error) bool {
	return target == ErrZcashErrorInvalidVaultPubkey
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

type ZcashErrorNotEnoughFunds struct {
	message string
}

func NewZcashErrorNotEnoughFunds() *ZcashError {
	return &ZcashError{
		err: &ZcashErrorNotEnoughFunds{},
	}
}

func (err ZcashErrorNotEnoughFunds) Error() string {
	return fmt.Sprintf("NotEnoughFunds: %s", err.message)
}

func (self ZcashErrorNotEnoughFunds) Is(target error) bool {
	return target == ErrZcashErrorNotEnoughFunds
}

type ZcashErrorTxRejected struct {
	message string
}

func NewZcashErrorTxRejected() *ZcashError {
	return &ZcashError{
		err: &ZcashErrorTxRejected{},
	}
}

func (err ZcashErrorTxRejected) Error() string {
	return fmt.Sprintf("TxRejected: %s", err.message)
}

func (self ZcashErrorTxRejected) Is(target error) bool {
	return target == ErrZcashErrorTxRejected
}

type ZcashErrorReorg struct {
	message string
}

func NewZcashErrorReorg() *ZcashError {
	return &ZcashError{
		err: &ZcashErrorReorg{},
	}
}

func (err ZcashErrorReorg) Error() string {
	return fmt.Sprintf("Reorg: %s", err.message)
}

func (self ZcashErrorReorg) Is(target error) bool {
	return target == ErrZcashErrorReorg
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
		return &ZcashError{&ZcashErrorInvalidVaultPubkey{message}}
	case 3:
		return &ZcashError{&ZcashErrorInvalidAddress{message}}
	case 4:
		return &ZcashError{&ZcashErrorNoOrchardReceiver{message}}
	case 5:
		return &ZcashError{&ZcashErrorNotEnoughFunds{message}}
	case 6:
		return &ZcashError{&ZcashErrorTxRejected{message}}
	case 7:
		return &ZcashError{&ZcashErrorReorg{message}}
	case 8:
		return &ZcashError{&ZcashErrorAssertError{message}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeZcashError.Read()", errorID))
	}

}

func (c FfiConverterTypeZcashError) Write(writer io.Writer, value *ZcashError) {
	switch variantValue := value.err.(type) {
	case *ZcashErrorRpc:
		writeInt32(writer, 1)
	case *ZcashErrorInvalidVaultPubkey:
		writeInt32(writer, 2)
	case *ZcashErrorInvalidAddress:
		writeInt32(writer, 3)
	case *ZcashErrorNoOrchardReceiver:
		writeInt32(writer, 4)
	case *ZcashErrorNotEnoughFunds:
		writeInt32(writer, 5)
	case *ZcashErrorTxRejected:
		writeInt32(writer, 6)
	case *ZcashErrorReorg:
		writeInt32(writer, 7)
	case *ZcashErrorAssertError:
		writeInt32(writer, 8)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeZcashError.Write", value))
	}
}

type FfiConverterOptionalTypeBlockTxs struct{}

var FfiConverterOptionalTypeBlockTxsINSTANCE = FfiConverterOptionalTypeBlockTxs{}

func (c FfiConverterOptionalTypeBlockTxs) Lift(rb RustBufferI) *BlockTxs {
	return LiftFromRustBuffer[*BlockTxs](c, rb)
}

func (_ FfiConverterOptionalTypeBlockTxs) Read(reader io.Reader) *BlockTxs {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeBlockTxsINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeBlockTxs) Lower(value *BlockTxs) RustBuffer {
	return LowerIntoRustBuffer[*BlockTxs](c, value)
}

func (_ FfiConverterOptionalTypeBlockTxs) Write(writer io.Writer, value *BlockTxs) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeBlockTxsINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeBlockTxs struct{}

func (_ FfiDestroyerOptionalTypeBlockTxs) Destroy(value *BlockTxs) {
	if value != nil {
		FfiDestroyerTypeBlockTxs{}.Destroy(*value)
	}
}

type FfiConverterSequenceString struct{}

var FfiConverterSequenceStringINSTANCE = FfiConverterSequenceString{}

func (c FfiConverterSequenceString) Lift(rb RustBufferI) []string {
	return LiftFromRustBuffer[[]string](c, rb)
}

func (c FfiConverterSequenceString) Read(reader io.Reader) []string {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]string, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterStringINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceString) Lower(value []string) RustBuffer {
	return LowerIntoRustBuffer[[]string](c, value)
}

func (c FfiConverterSequenceString) Write(writer io.Writer, value []string) {
	if len(value) > math.MaxInt32 {
		panic("[]string is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterStringINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceString struct{}

func (FfiDestroyerSequenceString) Destroy(sequence []string) {
	for _, value := range sequence {
		FfiDestroyerString{}.Destroy(value)
	}
}

type FfiConverterSequenceBytes struct{}

var FfiConverterSequenceBytesINSTANCE = FfiConverterSequenceBytes{}

func (c FfiConverterSequenceBytes) Lift(rb RustBufferI) [][]byte {
	return LiftFromRustBuffer[[][]byte](c, rb)
}

func (c FfiConverterSequenceBytes) Read(reader io.Reader) [][]byte {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([][]byte, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterBytesINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceBytes) Lower(value [][]byte) RustBuffer {
	return LowerIntoRustBuffer[[][]byte](c, value)
}

func (c FfiConverterSequenceBytes) Write(writer io.Writer, value [][]byte) {
	if len(value) > math.MaxInt32 {
		panic("[][]byte is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterBytesINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceBytes struct{}

func (FfiDestroyerSequenceBytes) Destroy(sequence [][]byte) {
	for _, value := range sequence {
		FfiDestroyerBytes{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeOutput struct{}

var FfiConverterSequenceTypeOutputINSTANCE = FfiConverterSequenceTypeOutput{}

func (c FfiConverterSequenceTypeOutput) Lift(rb RustBufferI) []Output {
	return LiftFromRustBuffer[[]Output](c, rb)
}

func (c FfiConverterSequenceTypeOutput) Read(reader io.Reader) []Output {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]Output, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeOutputINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeOutput) Lower(value []Output) RustBuffer {
	return LowerIntoRustBuffer[[]Output](c, value)
}

func (c FfiConverterSequenceTypeOutput) Write(writer io.Writer, value []Output) {
	if len(value) > math.MaxInt32 {
		panic("[]Output is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeOutputINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeOutput struct{}

func (FfiDestroyerSequenceTypeOutput) Destroy(sequence []Output) {
	for _, value := range sequence {
		FfiDestroyerTypeOutput{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeUTXO struct{}

var FfiConverterSequenceTypeUTXOINSTANCE = FfiConverterSequenceTypeUTXO{}

func (c FfiConverterSequenceTypeUTXO) Lift(rb RustBufferI) []Utxo {
	return LiftFromRustBuffer[[]Utxo](c, rb)
}

func (c FfiConverterSequenceTypeUTXO) Read(reader io.Reader) []Utxo {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]Utxo, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeUTXOINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeUTXO) Lower(value []Utxo) RustBuffer {
	return LowerIntoRustBuffer[[]Utxo](c, value)
}

func (c FfiConverterSequenceTypeUTXO) Write(writer io.Writer, value []Utxo) {
	if len(value) > math.MaxInt32 {
		panic("[]Utxo is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeUTXOINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeUtxo struct{}

func (FfiDestroyerSequenceTypeUtxo) Destroy(sequence []Utxo) {
	for _, value := range sequence {
		FfiDestroyerTypeUtxo{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeVaultTx struct{}

var FfiConverterSequenceTypeVaultTxINSTANCE = FfiConverterSequenceTypeVaultTx{}

func (c FfiConverterSequenceTypeVaultTx) Lift(rb RustBufferI) []VaultTx {
	return LiftFromRustBuffer[[]VaultTx](c, rb)
}

func (c FfiConverterSequenceTypeVaultTx) Read(reader io.Reader) []VaultTx {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]VaultTx, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeVaultTxINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeVaultTx) Lower(value []VaultTx) RustBuffer {
	return LowerIntoRustBuffer[[]VaultTx](c, value)
}

func (c FfiConverterSequenceTypeVaultTx) Write(writer io.Writer, value []VaultTx) {
	if len(value) > math.MaxInt32 {
		panic("[]VaultTx is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeVaultTxINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeVaultTx struct{}

func (FfiDestroyerSequenceTypeVaultTx) Destroy(sequence []VaultTx) {
	for _, value := range sequence {
		FfiDestroyerTypeVaultTx{}.Destroy(value)
	}
}

func ApplySignatures(vault []byte, ptx PartialTx, signatures [][]byte) ([]byte, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_maya_zcash_fn_func_apply_signatures(FfiConverterBytesINSTANCE.Lower(vault), FfiConverterTypePartialTxINSTANCE.Lower(ptx), FfiConverterSequenceBytesINSTANCE.Lower(signatures), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue []byte
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterBytesINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func BroadcastRawTx(tx []byte) (string, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_maya_zcash_fn_func_broadcast_raw_tx(FfiConverterBytesINSTANCE.Lower(tx), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue string
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterStringINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func BuildVaultUnauthorizedTx(vault []byte, ptx PartialTx) (Sighashes, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_maya_zcash_fn_func_build_vault_unauthorized_tx(FfiConverterBytesINSTANCE.Lower(vault), FfiConverterTypePartialTxINSTANCE.Lower(ptx), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Sighashes
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeSighashesINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func CombineVault(height uint32, vault []byte) (PartialTx, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_maya_zcash_fn_func_combine_vault(FfiConverterUint32INSTANCE.Lower(height), FfiConverterBytesINSTANCE.Lower(vault), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue PartialTx
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypePartialTxINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func CombineVaultUtxos(height uint32, vault []byte, utxos []Utxo) (PartialTx, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_maya_zcash_fn_func_combine_vault_utxos(FfiConverterUint32INSTANCE.Lower(height), FfiConverterBytesINSTANCE.Lower(vault), FfiConverterSequenceTypeUTXOINSTANCE.Lower(utxos), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue PartialTx
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypePartialTxINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func GetBalance(address string) (uint64, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) C.uint64_t {
		return C.uniffi_maya_zcash_fn_func_get_balance(FfiConverterStringINSTANCE.Lower(address), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue uint64
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterUint64INSTANCE.Lift(_uniffiRV), _uniffiErr
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

func GetOvk(pubkey []byte) ([]byte, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_maya_zcash_fn_func_get_ovk(FfiConverterBytesINSTANCE.Lower(pubkey), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue []byte
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterBytesINSTANCE.Lift(_uniffiRV), _uniffiErr
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

func ListUtxos(address string) ([]Utxo, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_maya_zcash_fn_func_list_utxos(FfiConverterStringINSTANCE.Lower(address), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue []Utxo
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterSequenceTypeUTXOINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
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

func PayFromVault(height uint32, vault []byte, to string, amount uint64, memo string) (PartialTx, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_maya_zcash_fn_func_pay_from_vault(FfiConverterUint32INSTANCE.Lower(height), FfiConverterBytesINSTANCE.Lower(vault), FfiConverterStringINSTANCE.Lower(to), FfiConverterUint64INSTANCE.Lower(amount), FfiConverterStringINSTANCE.Lower(memo), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue PartialTx
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypePartialTxINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func ScanBlocks(pubkey []byte, prevHashes []string) (*BlockTxs, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_maya_zcash_fn_func_scan_blocks(FfiConverterBytesINSTANCE.Lower(pubkey), FfiConverterSequenceStringINSTANCE.Lower(prevHashes), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *BlockTxs
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterOptionalTypeBlockTxsINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func ScanMempool(pubkey []byte) ([]VaultTx, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_maya_zcash_fn_func_scan_mempool(FfiConverterBytesINSTANCE.Lower(pubkey), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue []VaultTx
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterSequenceTypeVaultTxINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func SendToVault(expiryHeight uint32, sk []byte, from string, vault []byte, amount uint64, memo string) (TxBytes, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_maya_zcash_fn_func_send_to_vault(FfiConverterUint32INSTANCE.Lower(expiryHeight), FfiConverterBytesINSTANCE.Lower(sk), FfiConverterStringINSTANCE.Lower(from), FfiConverterBytesINSTANCE.Lower(vault), FfiConverterUint64INSTANCE.Lower(amount), FfiConverterStringINSTANCE.Lower(memo), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue TxBytes
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeTxBytesINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func SignSighash(sk []byte, sighash []byte) ([]byte, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_maya_zcash_fn_func_sign_sighash(FfiConverterBytesINSTANCE.Lower(sk), FfiConverterBytesINSTANCE.Lower(sighash), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue []byte
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterBytesINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func SkToPub(wif string) (TransparentKey, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeZcashError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_maya_zcash_fn_func_sk_to_pub(FfiConverterStringINSTANCE.Lower(wif), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue TransparentKey
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeTransparentKeyINSTANCE.Lift(_uniffiRV), _uniffiErr
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
