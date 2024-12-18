

// This file was autogenerated by some hot garbage in the `uniffi` crate.
// Trust me, you don't want to mess with it!



#include <stdbool.h>
#include <stdint.h>

// The following structs are used to implement the lowest level
// of the FFI, and thus useful to multiple uniffied crates.
// We ensure they are declared exactly once, with a header guard, UNIFFI_SHARED_H.
#ifdef UNIFFI_SHARED_H
	// We also try to prevent mixing versions of shared uniffi header structs.
	// If you add anything to the #else block, you must increment the version suffix in UNIFFI_SHARED_HEADER_V6
	#ifndef UNIFFI_SHARED_HEADER_V6
		#error Combining helper code from multiple versions of uniffi is not supported
	#endif // ndef UNIFFI_SHARED_HEADER_V6
#else
#define UNIFFI_SHARED_H
#define UNIFFI_SHARED_HEADER_V6
// ⚠️ Attention: If you change this #else block (ending in `#endif // def UNIFFI_SHARED_H`) you *must* ⚠️
// ⚠️ increment the version suffix in all instances of UNIFFI_SHARED_HEADER_V6 in this file.           ⚠️

typedef struct RustBuffer {
	int32_t capacity;
	int32_t len;
	uint8_t *data;
} RustBuffer;

typedef int32_t (*ForeignCallback)(uint64_t, int32_t, uint8_t *, int32_t, RustBuffer *);

// Task defined in Rust that Go executes
typedef void (*RustTaskCallback)(const void *, int8_t);

// Callback to execute Rust tasks using a Go routine
//
// Args:
//   executor: ForeignExecutor lowered into a uint64_t value
//   delay: Delay in MS
//   task: RustTaskCallback to call
//   task_data: data to pass the task callback
typedef int8_t (*ForeignExecutorCallback)(uint64_t, uint32_t, RustTaskCallback, void *);

typedef struct ForeignBytes {
	int32_t len;
	const uint8_t *data;
} ForeignBytes;

// Error definitions
typedef struct RustCallStatus {
	int8_t code;
	RustBuffer errorBuf;
} RustCallStatus;

// Continuation callback for UniFFI Futures
typedef void (*RustFutureContinuation)(void * , int8_t);

// ⚠️ Attention: If you change this #else block (ending in `#endif // def UNIFFI_SHARED_H`) you *must* ⚠️
// ⚠️ increment the version suffix in all instances of UNIFFI_SHARED_HEADER_V6 in this file.           ⚠️
#endif // def UNIFFI_SHARED_H

// Needed because we can't execute the callback directly from go.
void cgo_rust_task_callback_bridge_maya_zcash(RustTaskCallback, const void *, int8_t);

int8_t uniffiForeignExecutorCallbackmaya_zcash(uint64_t, uint32_t, RustTaskCallback, void*);

void uniffiFutureContinuationCallbackmaya_zcash(void*, int8_t);

RustBuffer uniffi_maya_zcash_fn_func_apply_signatures(
	RustBuffer vault,
	RustBuffer ptx,
	RustBuffer signatures,
	RustCallStatus* out_status
);

RustBuffer uniffi_maya_zcash_fn_func_best_recipient_of_ua(
	RustBuffer address,
	RustCallStatus* out_status
);

RustBuffer uniffi_maya_zcash_fn_func_broadcast_raw_tx(
	RustBuffer tx,
	RustCallStatus* out_status
);

RustBuffer uniffi_maya_zcash_fn_func_combine_vault(
	uint32_t height,
	RustBuffer vault,
	RustCallStatus* out_status
);

RustBuffer uniffi_maya_zcash_fn_func_combine_vault_utxos(
	uint32_t height,
	RustBuffer vault,
	RustBuffer destination_vaults,
	RustBuffer utxos,
	RustCallStatus* out_status
);

uint64_t uniffi_maya_zcash_fn_func_get_balance(
	RustBuffer address,
	RustCallStatus* out_status
);

RustBuffer uniffi_maya_zcash_fn_func_get_latest_height(
	RustCallStatus* out_status
);

RustBuffer uniffi_maya_zcash_fn_func_get_ovk(
	RustBuffer pubkey,
	RustCallStatus* out_status
);

RustBuffer uniffi_maya_zcash_fn_func_get_vault_address(
	RustBuffer pubkey,
	RustCallStatus* out_status
);

void uniffi_maya_zcash_fn_func_init_logger(
	RustCallStatus* out_status
);

RustBuffer uniffi_maya_zcash_fn_func_list_utxos(
	RustBuffer address,
	RustCallStatus* out_status
);

RustBuffer uniffi_maya_zcash_fn_func_make_ua(
	RustBuffer transparent,
	RustBuffer sapling,
	RustBuffer orchard,
	RustCallStatus* out_status
);

int8_t uniffi_maya_zcash_fn_func_match_with_blockchain_receiver(
	RustBuffer address,
	RustBuffer receiver,
	RustCallStatus* out_status
);

RustBuffer uniffi_maya_zcash_fn_func_pay_from_vault(
	uint32_t height,
	RustBuffer vault,
	RustBuffer to,
	uint64_t amount,
	RustBuffer memo,
	RustCallStatus* out_status
);

RustBuffer uniffi_maya_zcash_fn_func_scan_blocks(
	RustBuffer pubkey,
	RustBuffer prev_hashes,
	RustCallStatus* out_status
);

RustBuffer uniffi_maya_zcash_fn_func_scan_mempool(
	RustBuffer pubkey,
	RustCallStatus* out_status
);

RustBuffer uniffi_maya_zcash_fn_func_send_to_vault(
	uint32_t expiry_height,
	RustBuffer sk,
	RustBuffer from,
	RustBuffer vault,
	uint64_t amount,
	RustBuffer memo,
	RustCallStatus* out_status
);

RustBuffer uniffi_maya_zcash_fn_func_sign_sighash(
	RustBuffer sk,
	RustBuffer sighash,
	RustCallStatus* out_status
);

RustBuffer uniffi_maya_zcash_fn_func_sk_to_pub(
	RustBuffer wif,
	RustCallStatus* out_status
);

int8_t uniffi_maya_zcash_fn_func_validate_address(
	RustBuffer address,
	RustCallStatus* out_status
);

RustBuffer ffi_maya_zcash_rustbuffer_alloc(
	int32_t size,
	RustCallStatus* out_status
);

RustBuffer ffi_maya_zcash_rustbuffer_from_bytes(
	ForeignBytes bytes,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rustbuffer_free(
	RustBuffer buf,
	RustCallStatus* out_status
);

RustBuffer ffi_maya_zcash_rustbuffer_reserve(
	RustBuffer buf,
	int32_t additional,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_continuation_callback_set(
	RustFutureContinuation callback,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_poll_u8(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_cancel_u8(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_free_u8(
	void* handle,
	RustCallStatus* out_status
);

uint8_t ffi_maya_zcash_rust_future_complete_u8(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_poll_i8(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_cancel_i8(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_free_i8(
	void* handle,
	RustCallStatus* out_status
);

int8_t ffi_maya_zcash_rust_future_complete_i8(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_poll_u16(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_cancel_u16(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_free_u16(
	void* handle,
	RustCallStatus* out_status
);

uint16_t ffi_maya_zcash_rust_future_complete_u16(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_poll_i16(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_cancel_i16(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_free_i16(
	void* handle,
	RustCallStatus* out_status
);

int16_t ffi_maya_zcash_rust_future_complete_i16(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_poll_u32(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_cancel_u32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_free_u32(
	void* handle,
	RustCallStatus* out_status
);

uint32_t ffi_maya_zcash_rust_future_complete_u32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_poll_i32(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_cancel_i32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_free_i32(
	void* handle,
	RustCallStatus* out_status
);

int32_t ffi_maya_zcash_rust_future_complete_i32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_poll_u64(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_cancel_u64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_free_u64(
	void* handle,
	RustCallStatus* out_status
);

uint64_t ffi_maya_zcash_rust_future_complete_u64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_poll_i64(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_cancel_i64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_free_i64(
	void* handle,
	RustCallStatus* out_status
);

int64_t ffi_maya_zcash_rust_future_complete_i64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_poll_f32(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_cancel_f32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_free_f32(
	void* handle,
	RustCallStatus* out_status
);

float ffi_maya_zcash_rust_future_complete_f32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_poll_f64(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_cancel_f64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_free_f64(
	void* handle,
	RustCallStatus* out_status
);

double ffi_maya_zcash_rust_future_complete_f64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_poll_pointer(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_cancel_pointer(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_free_pointer(
	void* handle,
	RustCallStatus* out_status
);

void* ffi_maya_zcash_rust_future_complete_pointer(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_poll_rust_buffer(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_cancel_rust_buffer(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_free_rust_buffer(
	void* handle,
	RustCallStatus* out_status
);

RustBuffer ffi_maya_zcash_rust_future_complete_rust_buffer(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_poll_void(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_cancel_void(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_free_void(
	void* handle,
	RustCallStatus* out_status
);

void ffi_maya_zcash_rust_future_complete_void(
	void* handle,
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_apply_signatures(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_best_recipient_of_ua(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_broadcast_raw_tx(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_combine_vault(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_combine_vault_utxos(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_get_balance(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_get_latest_height(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_get_ovk(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_get_vault_address(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_init_logger(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_list_utxos(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_make_ua(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_match_with_blockchain_receiver(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_pay_from_vault(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_scan_blocks(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_scan_mempool(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_send_to_vault(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_sign_sighash(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_sk_to_pub(
	RustCallStatus* out_status
);

uint16_t uniffi_maya_zcash_checksum_func_validate_address(
	RustCallStatus* out_status
);

uint32_t ffi_maya_zcash_uniffi_contract_version(
	RustCallStatus* out_status
);



