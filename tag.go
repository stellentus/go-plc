package plc

/*
#include <stdint.h>
*/
import "C"
import (
	"strconv"
	"strings"
)

const SystemTagBit = 0x1000
const TagDimensionMask = 0x6000

type Tag struct {
	name        string
	tagType     C.uint16_t
	elementSize C.uint16_t
	dimensions  []int
}

func (tag *Tag) Name() string {
	return tag.name
}

func (tag *Tag) addDimension(dim int) {
	if dim <= 0 {
		return
	}
	tag.dimensions = append(tag.dimensions, dim)
}

func (tag Tag) String() string {
	if len(tag.dimensions) == 0 {
		return tag.name
	}

	strs := make([]string, len(tag.dimensions))
	for i, v := range tag.dimensions {
		strs[i] = strconv.Itoa(v)
	}
	return tag.name + "[" + strings.Join(strs, ",") + "]"
}

func (tag Tag) ElemCount() int {
	count := 1
	for _, dim := range tag.dimensions {
		if dim != 0 {
			count *= dim
		}
	}
	return count
}
