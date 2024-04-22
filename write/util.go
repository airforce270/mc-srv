package write

// discardingWriter implements io.Writer, discarding the data.
// Writes will always succeed, and will never return an error.
type discardingWriter int

func (r *discardingWriter) Write(data []byte) (n int, err error) {
	n = len(data)
	*r += discardingWriter(n)
	return n, nil
}

// Len returns the number of bytes written.
func (w discardingWriter) Len() int { return int(w) }
