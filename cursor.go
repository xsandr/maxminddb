package maxminddb

type Cursor struct {
	buffer []byte
	cursor int
}

func (c *Cursor) moveCaret(n int) {
	c.cursor += n
}

func (c *Cursor) currentByte() byte {
	c.cursor += 1
	return c.buffer[c.cursor-1]
}

func (c *Cursor) nextBytes(n int) []byte {
	c.cursor += n
	return c.buffer[c.cursor-n : c.cursor]
}
