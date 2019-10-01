package rango

import (
	"testing"
)

func TestRandStr(t *testing.T) {
	if randStr(10) == randStr(10) {
		t.Fatalf("%s randStr() is fatal, randStr(10) == randStr(10).", failFlag)
	}
}

func TestGetDebugStackArr(t *testing.T) {
	DebugOff()
	emptyArr := getDebugStackArr()

	if len(emptyArr) != 0 {
		t.Fatalf("%s getDebugStackArr() need return Enmpty, when DebugOff().", failFlag)
	}

	DebugOn()
	arr := getDebugStackArr()

	if len(arr) == 0 {
		t.Fatalf("%s getDebugStackArr() is return Enmpty.", failFlag)
	}

	if !includeString(arr[1]["func"], "getDebugStackArr") {
		t.Fatalf("%s getDebugStackArr() is fatal.", failFlag)
	}

	if !includeString(arr[2]["func"], "TestGetDebugStackArr") {
		t.Fatalf("%s getDebugStackArr() is fatal.", failFlag)
	}
}

func TestSliceHasPrefix(t *testing.T) {
	s1 := []string{"image/png", "image/gif"}
	s2 := []string{"*"}
	s3 := []string{"image/"}
	s4 := []string{}

	testCases := []struct {
		expected bool

		s     []string
		value string
	}{
		{false, s1, "image/jpeg"},
		{true, s1, "image/gif"},
		{true, s2, "image/jpeg"},
		{true, s2, "video/mp4"},
		{true, s2, "audio/mp3"},
		{true, s2, "1234567890qwertyuiop"},
		{true, s3, "image/jpeg"},
		{false, s3, "video/mp4"},
		{false, s3, "audio/mp3"},
		{false, s4, "image/jpeg"},
		{false, s4, "video/mp4"},
		{false, s4, "audio/mp3"},
	}

	for _, testCase := range testCases {
		actual := sliceHasPrefix(testCase.s, testCase.value)
		if actual != testCase.expected {
			t.Fatalf("%s sliceIndexPrefix(%v,%v) is fatal, need %v, but %v.", failFlag, testCase.s, testCase.value, testCase.expected, actual)
		}
	}
}
