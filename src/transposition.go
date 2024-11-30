package main

import (
	"unsafe"
)

// first 14 bits contain depth that this node was searched to
// last 2 bits contain what type of node it is 0 = PV, 1 = Cut, 2 = All
type NodeInfo uint16
type NodeType uint16

type TableData struct {
	eval         int16
	bestMove     Move
	ply          uint16
	depthAndNode NodeInfo
}
type TableEntry struct {
	key  uint64
	data TableData
}

type TranspositionTable []TableEntry

const (
	EntrySize        = 16 // Size in bytes of a table entry
	megabytesToBytes = 1_048_576

	pVNode       NodeType = 0
	CutNode      NodeType = 1
	AllNode      NodeType = 2
	TerminalNode NodeType = 3

	bitMask14 uint16 = 0x3FFF
)

// Capacity in Bytes
var tableCapacity uint64

// Number of entries
var tableSize uint64
var transpositionTable TranspositionTable

//go:noescape
func Prefetch(dPrt unsafe.Pointer)

// capacity is in megabytes
func SetupTable(capacity uint64) {
	tableCapacity = capacity * megabytesToBytes
	tableSize = tableCapacity / EntrySize
	transpositionTable = make([]TableEntry, tableSize)
	SetupHashRandoms()
}

func (tt *TranspositionTable) AddState(s *State, eval int32, bestMove Move, depth uint16, nodeType NodeType) {
	hash := s.hashcode
	index := hash % tableSize
	(*tt)[index] = TableEntry{key: hash, data: TableData{eval: EvalHighToLow(eval), bestMove: bestMove, ply: s.ply, depthAndNode: NodeInfo(nodeType)<<14 | NodeInfo(depth)}}
}

func (tt *TranspositionTable) SearchState(s *State) (TableData, bool) {
	hash := s.hashcode
	index := hash % tableSize
	if (*tt)[index].key == hash {
		return (*tt)[index].data, true
	}
	return TableData{}, false
}

func (tt *TranspositionTable) Prefetch(s *State) {
	hash := s.hashcode
	addressOffset := (hash % tableSize) * EntrySize
	dataPointer := unsafe.Pointer(uintptr(unsafe.Pointer(unsafe.SliceData(*tt))) + uintptr(addressOffset))
	Prefetch(dataPointer)
}

// First return is depth, second is node type
func (nI *NodeInfo) parseDepthandNode() (uint16, NodeType) {
	return uint16(*nI) & bitMask14, NodeType(uint16(*nI>>14) & BitMask2)
}
