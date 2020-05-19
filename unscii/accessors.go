package unscii

var data []uint32
var charMap map[int32]uint16

func Data() []uint32 {
	return data
}

func CharMap() map[int32]uint16 {
	return charMap
}
