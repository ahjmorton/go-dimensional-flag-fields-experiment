package godimensionalflagfields

const (
	slabLength        = 32
	log2ForSlabLength = 5
)

type (
	OneDFlagField struct {
		fieldWidth uint8
		length     uint
		data       []uint32
	}
)

func (ff *OneDFlagField) Len() uint {
	return ff.length
}

func (ff *OneDFlagField) dataIndexAndOffsetFor(index uint) (uint, byte, error) {
	if index >= ff.Len() {
		return 0, 0, ErrOutOfBounds{
			boundType: "index",
			bound:     ff.Len(),
			input:     index,
		}

	}
	bitsToStartPoint := (index * uint(ff.fieldWidth))
	firstIndex := bitsToStartPoint >> log2ForSlabLength
	fieldStartPoint := byte(bitsToStartPoint & (slabLength - 1))
	return firstIndex, fieldStartPoint, nil
}

func dataFieldOffsets(fieldStartOffset byte, fieldIndex byte) (uint, uint) {
	fromStartOffset := fieldStartOffset + fieldIndex
	fromStartIndex := (uint(fromStartOffset) >> log2ForSlabLength)
	landedByteOffset := uint(fromStartOffset) - (fromStartIndex << log2ForSlabLength)
	return landedByteOffset, fromStartIndex
}

func (ff *OneDFlagField) AllSet(index uint, firstFieldIndex byte, rest ...byte) (bool, error) {
	dataIndex, offset, err := ff.dataIndexAndOffsetFor(index)
	if err != nil {
		return false, err
	}

	landedByteOffset, fromStartIndex := dataFieldOffsets(offset, firstFieldIndex)

	/*
	 * We could early exit here if this turns out not to be set but we won't.
	 * The reasoning is that early exiting here will make the branch predictors work harder.
	 *
	 * Assuming the indices of fields looked up is consistent, but the values of them are not,
	 * having the same number of items iterated each time will be easier to predict than
	 * having early exits screwing with predictions.
	 */
	result := ff.data[dataIndex+uint(fromStartIndex)]&(1<<landedByteOffset) == (1 << landedByteOffset)

	for _, nextFieldIndex := range rest {
		landedByteOffset, fromStartIndex = dataFieldOffsets(offset, nextFieldIndex)

		// Go, y u no have &&=?
		result = result && ff.data[dataIndex+uint(fromStartIndex)]&(1<<landedByteOffset) == (1<<landedByteOffset)
	}

	return result, nil
}

func (ff *OneDFlagField) IsSet(index uint, firstFieldIndex byte) (bool, error) {
	dataIndex, offset, err := ff.dataIndexAndOffsetFor(index)
	if err != nil {
		return false, err
	}

	landedByteOffset, fromStartIndex := dataFieldOffsets(offset, firstFieldIndex)

	return ff.data[dataIndex+uint(fromStartIndex)]&(1<<landedByteOffset) == (1 << landedByteOffset), nil
}

func (ff *OneDFlagField) Set(index uint, firstFieldIndex byte) error {
	dataIndex, offset, err := ff.dataIndexAndOffsetFor(index)
	if err != nil {
		return err
	}

	landedByteOffset, fromStartIndex := dataFieldOffsets(offset, firstFieldIndex)

	ff.data[dataIndex+uint(fromStartIndex)] |= (1 << landedByteOffset)
	return nil
}

func (ff *OneDFlagField) Unset(index uint, firstFieldIndex byte) error {
	dataIndex, offset, err := ff.dataIndexAndOffsetFor(index)
	if err != nil {
		return err
	}

	landedByteOffset, fromStartIndex := dataFieldOffsets(offset, firstFieldIndex)

	ff.data[dataIndex+uint(fromStartIndex)] &= ^(1 << landedByteOffset)
	return nil
}

func (ff *OneDFlagField) AnySet(index uint, firstFieldIndex byte, rest ...byte) (bool, error) {
	dataIndex, offset, err := ff.dataIndexAndOffsetFor(index)
	if err != nil {
		return false, err
	}

	landedByteOffset, fromStartIndex := dataFieldOffsets(offset, firstFieldIndex)

	/*
	 * We could early exit here if this turns out not to be set but we won't.
	 * The reasoning is that early exiting here will make the branch predictors work harder.
	 *
	 * Assuming the indices of fields looked up is consistent, but the values of them are not,
	 * having the same number of items iterated each time will be easier to predict than
	 * having early exits screwing with predictions.
	 */
	result := ff.data[dataIndex+uint(fromStartIndex)]&(1<<landedByteOffset) == (1 << landedByteOffset)

	for _, nextFieldIndex := range rest {
		landedByteOffset, fromStartIndex = dataFieldOffsets(offset, nextFieldIndex)

		// Go, y u no have ||=?
		result = result || ff.data[dataIndex+uint(fromStartIndex)]&(1<<landedByteOffset) == (1<<landedByteOffset)
	}

	return result, nil
}

func (ff *OneDFlagField) SetField(index uint, firstFieldIndex byte, rest ...byte) error {
	dataIndex, offset, err := ff.dataIndexAndOffsetFor(index)
	if err != nil {
		return err
	}

	/*
	* Rather than pre-processing the fields indices in some way e.g. grouping
	* fields who happen to land into the same uint32 in data so we can do a one
	* off OR, just repeatedly OR.
	*
	* The assumption here is that any pre-processing done here is going to
	* add up to more instructions and / or memory space (both of which increase
	* pressure on the cache.
	 */
	landedByteOffset, fromStartIndex := dataFieldOffsets(offset, firstFieldIndex)

	ff.data[dataIndex+uint(fromStartIndex)] |= 1 << landedByteOffset

	for _, nextFieldIndex := range rest {
		landedByteOffset, fromStartIndex = dataFieldOffsets(offset, nextFieldIndex)

		ff.data[dataIndex+uint(fromStartIndex)] |= 1 << landedByteOffset
	}

	return nil
}

func (ff *OneDFlagField) UnsetField(index uint, firstFieldIndex byte, rest ...byte) error {
	index, offset, err := ff.dataIndexAndOffsetFor(index)
	if err != nil {
		return err
	}

	/*
	* Rather than pre-processing the fields indices in some way e.g. grouping
	* fields who happen to land into the same uint32 in data so we can do a one
	* off OR, just repeatedly OR.
	*
	* The assumption here is that any pre-processing done here is going to
	* add up to more instructions and / or memory space (both of which increase
	* pressure on the cache.
	 */
	landedByteOffset, fromStartIndex := dataFieldOffsets(offset, firstFieldIndex)

	ff.data[index+uint(fromStartIndex)] &= ^(1 << landedByteOffset)

	for _, nextFieldIndex := range rest {
		landedByteOffset, fromStartIndex = dataFieldOffsets(offset, nextFieldIndex)

		ff.data[index+uint(fromStartIndex)] &= ^(1 << landedByteOffset)
	}

	return nil
}

func MakeOneDFlagField(fieldWidth uint8, length uint) *OneDFlagField {
	requiredBits := (length * uint(fieldWidth))
	dataWidth := requiredBits / slabLength

	if requiredBits%slabLength != 0 {
		dataWidth++
	}
	return &OneDFlagField{
		fieldWidth: fieldWidth,
		length:     length,
		data:       make([]uint32, dataWidth),
	}
}
