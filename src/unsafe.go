package main

import "unsafe"

func uGetA[T any](ptr *T, index uint) T {
	uPtr := unsafe.Pointer(ptr)
	size := unsafe.Sizeof(*ptr)
	offset := uintptr(index) * size
	return *(*T)(unsafe.Pointer(uintptr(uPtr) + offset))
}

func uGetAPtr[T any](ptr *T, index uint) *T {
	uPtr := unsafe.Pointer(ptr)
	size := unsafe.Sizeof(*ptr)
	offset := uintptr(index) * size
	return (*T)(unsafe.Pointer(uintptr(uPtr) + offset))
}

func uGetA2[T any](ptr *T, index0 uint, index1 uint, rowLength uint) T {
	uPtr := unsafe.Pointer(ptr)
	size := unsafe.Sizeof(*ptr)
	offset0 := uintptr(index0) * size * uintptr(rowLength)
	offset1 := uintptr(index1) * size
	return *(*T)(unsafe.Pointer(uintptr(uPtr) + offset0 + offset1))
}

func uSetA[T any](ptr *T, value T, index uint) {
	uPtr := unsafe.Pointer(ptr)
	size := unsafe.Sizeof(*ptr)
	offset := uintptr(index) * size
	*(*T)(unsafe.Pointer(uintptr(uPtr) + offset)) = value
}
