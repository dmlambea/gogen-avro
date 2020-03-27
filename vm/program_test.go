package vm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	programByteCode = []byte{
		byte(OpMov), byte(TypeInt),
		byte(OpMovOpt), byte(TypeInt), 0,
		byte(OpCall), 4,
		byte(OpLoad),
		byte(OpJmpEq), 1, 0,
		byte(OpCall), 1,
		byte(OpHalt),
		byte(OpMov), byte(TypeString), // xx
		byte(OpLoad),
		byte(OpJmpEq), 1, 0,
		byte(OpCall), 1,
		byte(OpRet),
		byte(OpMov), byte(TypeInt), // yy
		byte(OpLoad),
		byte(OpJmpEq), 1, 0,
		byte(OpJmp), -4 & 0xff,
		byte(OpRet),
	}
)

func TestProgram(t *testing.T) {
	_, err := NewProgram(programByteCode)
	require.Nil(t, err)
}
