package types

/*
Sample structs
*/

// Sample - A sample struct.
type Sample struct {
	SampleString string
	SampleInt    int
}

// String - string function
func (s Sample) String() string {
	return s.SampleString
}
