package godimensionalflagfields

import (
	"fmt"
	"math/rand"
	"testing"
)

const (
	LikesCats = iota
	LikesDogs
	LikesRats
	LikesRed
	LikesBlue
	LikesGreen
	LikesBurgers
	LikesFries
	LikesHotdogs
)

const (
	TestFieldLength = LikesHotdogs + 1
)

var (
	AllTestFields = [TestFieldLength]byte{
		LikesCats,
		LikesDogs,
		LikesRats,
		LikesRed,
		LikesBlue,
		LikesGreen,
		LikesBurgers,
		LikesFries,
		LikesHotdogs,
	}
)

func TestOneDimensionalFlagFields(t *testing.T) {

	t.Run("InitialState", func(t *testing.T) {
		length := uint(rand.Intn(10))
		subject := MakeOneDFlagField(TestFieldLength, length)

		if reportedLength := subject.Len(); reportedLength != length {
			t.Fatalf("Expected subject have length of %d but was %d", length, reportedLength)
		}

		for i := uint(0); i < length; i++ {
			anySet, err := subject.AnySet(i, AllTestFields[0], AllTestFields[1:]...)

			if err != nil {
				t.Errorf("Unexpected error checking for 0 state %s", err)
			}

			if anySet {
				t.Errorf("Found set fields on entry %d. This should never happen", i)
			}
		}

	})

	t.Run("Simply setting and testing fields", func(t *testing.T) {
		subject := MakeOneDFlagField(TestFieldLength, 5)
		if err := subject.SetField(2, LikesBlue, LikesFries, LikesRats); err != nil {
			t.Fatalf("Unexpected error while setting fields %s", err)
		}
		if err := subject.SetField(0, LikesCats, LikesHotdogs); err != nil {
			t.Fatalf("Unexpected error while setting fields %s", err)
		}
		if err := subject.SetField(4, LikesCats, LikesGreen, LikesHotdogs); err != nil {
			t.Fatalf("Unexpected error while setting fields %s", err)
		}
		if err := subject.UnsetField(4, LikesGreen); err != nil {
			t.Fatalf("Unexpected error while setting fields %s", err)
		}

		anySet, err := subject.AnySet(1, AllTestFields[0], AllTestFields[1:]...)

		if err != nil {
			t.Fatalf("Unexpected error while setting fields %s", err)
		}

		if anySet {
			t.Errorf("Found set fields on entry %d. This should never happen", 5)
		}

		anySet, err = subject.AnySet(0, LikesCats)

		if err != nil {
			t.Fatalf("Unexpected error while setting fields %s", err)
		}

		if !anySet {
			t.Errorf("Didn't find expected fields set on %d.", 1)
		}

		allSet, err := subject.AllSet(4, LikesCats, LikesGreen, LikesHotdogs)

		if err != nil {
			t.Fatalf("Unexpected error while setting fields %s", err)
		}

		if allSet {
			t.Errorf("Found set fields on entry %d. This should never happen", 4)
		}

		allSet, err = subject.AllSet(2, LikesBlue, LikesFries, LikesRats)

		if err != nil {
			t.Fatalf("Unexpected error while setting fields %s", err)
		}

		if !allSet {
			t.Errorf("Found Unset fields on entry %d. This should never happen", 2)
		}

	})
}

type SingleSetOperations interface {
	Set(fieldIndex byte)
	Unset(fieldIndex byte)
	IsSet(fieldindex byte) bool
}

type SliceApproach []bool

func (saa SliceApproach) Set(fieldIndex byte) {
	if fieldIndex > TestFieldLength {
		panic("TOo biig failing")
	}
	saa[fieldIndex] = true
}

func (saa SliceApproach) Unset(fieldIndex byte) {
	saa[fieldIndex] = false
}

func (saa SliceApproach) IsSet(fieldIndex byte) bool {
	return saa[fieldIndex]
}

type StaticArrayApproach [TestFieldLength]bool

func (saa *StaticArrayApproach) Set(fieldIndex byte) {
	saa[fieldIndex] = true
}

func (saa *StaticArrayApproach) Unset(fieldIndex byte) {
	saa[fieldIndex] = false
}

func (saa *StaticArrayApproach) IsSet(fieldIndex byte) bool {
	return saa[fieldIndex]
}

type MapApproach map[byte]bool

func (ma MapApproach) Set(fieldIndex byte) {
	ma[fieldIndex] = true
}

func (ma MapApproach) Unset(fieldIndex byte) {
	ma[fieldIndex] = false
}

func (ma MapApproach) IsSet(fieldIndex byte) bool {
	return ma[fieldIndex]
}

type BitewiseApproach uint16

func (ba *BitewiseApproach) Set(fieldIndex byte) {
	if fieldIndex > TestFieldLength {
		panic("TOo biig failing")
	}
	*ba |= (1 << fieldIndex)
}

func (ba *BitewiseApproach) Unset(fieldIndex byte) {
	if fieldIndex > TestFieldLength {
		panic("TOo biig failing")
	}
	*ba &= ^(1 << fieldIndex)
}

func (ba BitewiseApproach) IsSet(fieldIndex byte) bool {
	if fieldIndex > TestFieldLength {
		panic("TOo biig failing")
	}
	return ba&(1<<fieldIndex) == (1 << fieldIndex)
}

type Bitewise32Approach uint32

func (ba *Bitewise32Approach) Set(fieldIndex byte) {
	if fieldIndex > TestFieldLength {
		panic("TOo biig failing")
	}
	*ba |= (1 << fieldIndex)
}

func (ba *Bitewise32Approach) Unset(fieldIndex byte) {
	if fieldIndex > TestFieldLength {
		panic("TOo biig failing")
	}
	*ba &= ^(1 << fieldIndex)
}

func (ba Bitewise32Approach) IsSet(fieldIndex byte) bool {
	if fieldIndex > TestFieldLength {
		panic("TOo biig failing")
	}
	return ba&(1<<fieldIndex) == (1 << fieldIndex)
}

type StructApproach struct {
	LikesCatsVal    bool
	LikesDogsVal    bool
	LikesRatsVal    bool
	LikesRedVal     bool
	LikesBlueVal    bool
	LikesGreenVal   bool
	LikesBurgersVal bool
	LikesFriesVal   bool
	LikesHotdogsVal bool
}

func (sa *StructApproach) Set(fieldIndex byte) {
	switch fieldIndex {
	case LikesCats:
		sa.LikesCatsVal = true
	case LikesDogs:
		sa.LikesDogsVal = true
	case LikesRats:
		sa.LikesRatsVal = true
	case LikesRed:
		sa.LikesRedVal = true
	case LikesBlue:
		sa.LikesBlueVal = true
	case LikesGreen:
		sa.LikesGreenVal = true
	case LikesBurgers:
		sa.LikesBurgersVal = true
	case LikesFries:
		sa.LikesFriesVal = true
	case LikesHotdogs:
		sa.LikesHotdogsVal = true
	}
}

func (sa *StructApproach) Unset(fieldIndex byte) {
	switch fieldIndex {
	case LikesCats:
		sa.LikesCatsVal = false
	case LikesDogs:
		sa.LikesDogsVal = false
	case LikesRats:
		sa.LikesRatsVal = false
	case LikesRed:
		sa.LikesRedVal = false
	case LikesBlue:
		sa.LikesBlueVal = false
	case LikesGreen:
		sa.LikesGreenVal = false
	case LikesBurgers:
		sa.LikesBurgersVal = false
	case LikesFries:
		sa.LikesFriesVal = false
	case LikesHotdogs:
		sa.LikesHotdogsVal = false
	}
}

func (sa *StructApproach) IsSet(fieldIndex byte) bool {
	switch fieldIndex {
	case LikesCats:
		return sa.LikesCatsVal
	case LikesDogs:
		return sa.LikesDogsVal
	case LikesRats:
		return sa.LikesRatsVal
	case LikesRed:
		return sa.LikesRedVal
	case LikesBlue:
		return sa.LikesBlueVal
	case LikesGreen:
		return sa.LikesGreenVal
	case LikesBurgers:
		return sa.LikesBurgersVal
	case LikesFries:
		return sa.LikesFriesVal
	case LikesHotdogs:
		return sa.LikesHotdogsVal
	}

	panic("This should never happen")
}

func BenchmarkOneDimenstionalVsAlternateApproachesForSingleFields(b *testing.B) {
	elementCounts := []int{100, 1000, 10000, 100000, 1000000, 10000000, 10000000, 100000000}
	const iterationCount = 10000
	for _, elementCount := range elementCounts {

		subject := MakeOneDFlagField(TestFieldLength, uint(elementCount))
		structs := make([]StructApproach, elementCount)
		maps := make([]MapApproach, elementCount)
		for i := range maps {
			maps[i] = make(MapApproach, TestFieldLength)
		}
		slices := make([]SliceApproach, elementCount)
		for i := range maps {
			slices[i] = make(SliceApproach, TestFieldLength)
		}
		arrays := make([]StaticArrayApproach, elementCount)
		bitwise16 := make([]BitewiseApproach, elementCount)
		bitwise32 := make([]Bitewise32Approach, elementCount)

		b.Run(fmt.Sprintf("%s|%d|Random access", "Library", elementCount), func(b *testing.B) {
			for i := 0; i < iterationCount; i++ {
				subject.SetField(uint(rand.Intn(elementCount)), byte(rand.Intn(TestFieldLength)))
				subject.UnsetField(uint(rand.Intn(elementCount)), byte(rand.Intn(TestFieldLength)))
				subject.AnySet(uint(rand.Intn(elementCount)), byte(rand.Intn(TestFieldLength)))
			}
		})

		b.Run(fmt.Sprintf("%s|%d|Random access", "Library Single access", elementCount), func(b *testing.B) {
			for i := 0; i < iterationCount; i++ {
				subject.Set(uint(rand.Intn(elementCount)), byte(rand.Intn(TestFieldLength)))
				subject.Unset(uint(rand.Intn(elementCount)), byte(rand.Intn(TestFieldLength)))
				subject.IsSet(uint(rand.Intn(elementCount)), byte(rand.Intn(TestFieldLength)))
			}
		})

		b.Run(fmt.Sprintf("%s|%d|Random access", "Arrays", elementCount), func(b *testing.B) {
			for i := 0; i < iterationCount; i++ {
				arrays[rand.Intn(elementCount)].Set(byte(rand.Intn(TestFieldLength)))
				arrays[rand.Intn(elementCount)].Unset(byte(rand.Intn(TestFieldLength)))
				arrays[rand.Intn(elementCount)].IsSet(byte(rand.Intn(TestFieldLength)))
			}
		})

		b.Run(fmt.Sprintf("%s|%d|Random access", "Slices", elementCount), func(b *testing.B) {
			for i := 0; i < iterationCount; i++ {
				slices[rand.Intn(elementCount)].Set(byte(rand.Intn(TestFieldLength)))
				slices[rand.Intn(elementCount)].Unset(byte(rand.Intn(TestFieldLength)))
				slices[rand.Intn(elementCount)].IsSet(byte(rand.Intn(TestFieldLength)))
			}
		})
		b.Run(fmt.Sprintf("%s|%d|Random access", "Bitwise16Bit", elementCount), func(b *testing.B) {
			for i := 0; i < iterationCount; i++ {
				bitwise16[rand.Intn(elementCount)].Set(byte(rand.Intn(TestFieldLength)))
				bitwise16[rand.Intn(elementCount)].Unset(byte(rand.Intn(TestFieldLength)))
				bitwise16[rand.Intn(elementCount)].IsSet(byte(rand.Intn(TestFieldLength)))
			}
		})
		b.Run(fmt.Sprintf("%s|%d|Random access", "Bitwise32Bit", elementCount), func(b *testing.B) {
			for i := 0; i < iterationCount; i++ {
				bitwise32[rand.Intn(elementCount)].Set(byte(rand.Intn(TestFieldLength)))
				bitwise32[rand.Intn(elementCount)].Unset(byte(rand.Intn(TestFieldLength)))
				bitwise32[rand.Intn(elementCount)].IsSet(byte(rand.Intn(TestFieldLength)))
			}
		})
		b.Run(fmt.Sprintf("%s|%d|Random access", "Structs", elementCount), func(b *testing.B) {
			for i := 0; i < iterationCount; i++ {
				structs[rand.Intn(elementCount)].Set(byte(rand.Intn(TestFieldLength)))
				structs[rand.Intn(elementCount)].Unset(byte(rand.Intn(TestFieldLength)))
				structs[rand.Intn(elementCount)].IsSet(byte(rand.Intn(TestFieldLength)))
			}
		})
		b.Run(fmt.Sprintf("%s|%d|Random access", "Maps", elementCount), func(b *testing.B) {
			for i := 0; i < iterationCount; i++ {
				maps[rand.Intn(elementCount)].Set(byte(rand.Intn(TestFieldLength)))
				maps[rand.Intn(elementCount)].Unset(byte(rand.Intn(TestFieldLength)))
				maps[rand.Intn(elementCount)].IsSet(byte(rand.Intn(TestFieldLength)))
			}
		})
	}
}
