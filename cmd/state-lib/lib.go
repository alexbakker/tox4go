package main

//FYI: THIS MESS DOESN'T WORK YET

//go build -buildmode=c-shared -o libtoxstate.so lib.go

/*
#include <stdint.h>
#include <stdlib.h>

struct Tox_Node {

};

struct Tox_Friend {
	uint8_t status;
	uint8_t user_status;
	uint8_t public_key[32];
	uint32_t nospam;
	uint64_t last_seen;

	const char *name;
	const char *status_message;
	const char *request_message;
};

struct Tox_State {
	uint8_t public_key[32];
	uint8_t secret_key[32];
	uint32_t nospam;

	const char *name;
	const char *status_message;
	uint8_t status;

	struct Tox_Friend **friends;
	size_t friends_length;
	struct Tox_Node **nodes;
	struct Tox_Node **path_nodes;
	struct Tox_Node **tcp_relays;
};

*/
import "C"

import (
	"reflect"
	"unsafe"

	"github.com/alexbakker/tox4go/state"
)

//export state_unmarshal
func state_unmarshal(data *C.uint8_t, length C.size_t, rst **C.struct_Tox_State) C.uint32_t {
	if data == nil || rst == nil {
		return 1
	}

	st := (*C.struct_Tox_State)(C.malloc(C.sizeof_struct_Tox_State))
	bytes := *(*[]byte)(unsafe.Pointer(
		&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(data)),
			Len:  int(length),
			Cap:  int(length),
		}),
	)

	temp := new(state.State)
	err := temp.UnmarshalBinary(bytes)
	if err != nil {
		return 2
	}

	st.public_key = *(*[32]C.uint8_t)(unsafe.Pointer(temp.PublicKey))
	st.secret_key = *(*[32]C.uint8_t)(unsafe.Pointer(temp.SecretKey))
	st.nospam = C.uint32_t(temp.Nospam)
	st.status = C.uint8_t(temp.Status)
	st.name = C.CString(temp.Name)
	st.status_message = C.CString(temp.StatusMessage)
	st.friends_length = C.size_t(len(temp.Friends))

	friends := *(*[]*C.struct_Tox_Friend)(unsafe.Pointer(
		&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(C.malloc(C.size_t(unsafe.Sizeof(uintptr(0))) * st.friends_length))),
			Len:  int(C.size_t(unsafe.Sizeof(uintptr(0))) * st.friends_length),
			Cap:  int(C.size_t(unsafe.Sizeof(uintptr(0))) * st.friends_length),
		}),
	)

	for i, friend := range temp.Friends {
		friends[i] = (*C.struct_Tox_Friend)(C.malloc(C.sizeof_struct_Tox_Friend))
		friends[i].name = C.CString(friend.Name)
		friends[i].status = C.uint8_t(friend.Status)
		friends[i].nospam = C.uint32_t(friend.Nospam)
		friends[i].user_status = C.uint8_t(friend.UserStatus)
		friends[i].status_message = C.CString(friend.StatusMessage)
		friends[i].request_message = C.CString(friend.RequestMessage)
		friends[i].public_key = *(*[32]C.uint8_t)(unsafe.Pointer(friend.PublicKey))
	}
	st.friends = (**C.struct_Tox_Friend)(unsafe.Pointer(&friends[0]))

	*rst = st
	return 0
}

//export state_free
func state_free(st *C.struct_Tox_State) {
	if st == nil {
		return
	}

	friends := *(*[]*C.struct_Tox_Friend)(unsafe.Pointer(
		&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(st.friends)),
			Len:  int(C.size_t(unsafe.Sizeof(uintptr(0))) * st.friends_length),
			Cap:  int(C.size_t(unsafe.Sizeof(uintptr(0))) * st.friends_length),
		}),
	)

	for _, f := range friends {
		C.free(unsafe.Pointer(f.name))
		C.free(unsafe.Pointer(f.status_message))
		C.free(unsafe.Pointer(f.request_message))
		C.free(unsafe.Pointer(f))
	}

	C.free(unsafe.Pointer(st.friends))
	C.free(unsafe.Pointer(st.name))
	C.free(unsafe.Pointer(st.status_message))
	C.free(unsafe.Pointer(st))
}

//export state_marshal
func state_marshal() {

}

func main() {}
