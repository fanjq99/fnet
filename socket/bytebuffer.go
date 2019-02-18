package socket

const (
	preSize  = 0
	initSize = 10
)

type ByteBuffer struct {
	buffers     []byte
	prependSize int
	readerIndex int
	writerIndex int
}

func NewByteBuffer() *ByteBuffer {
	return &ByteBuffer{
		buffers:     make([]byte, preSize+initSize),
		prependSize: preSize,
		readerIndex: preSize,
		writerIndex: preSize,
	}
}

func (b *ByteBuffer) Append(buff ...byte) {
	size := len(buff)
	if size == 0 {
		return
	}
	b.WrGrow(size)
	copy(b.buffers[b.writerIndex:], buff)
	b.WrFlip(size)
}

func (b *ByteBuffer) WrBuf() []byte {
	if b.writerIndex >= len(b.buffers) {
		return nil
	}
	return b.buffers[b.writerIndex:]
}

func (b *ByteBuffer) WrSize() int {
	return len(b.buffers) - b.writerIndex
}

func (b *ByteBuffer) WrFlip(size int) {
	b.writerIndex += size
}

func (b *ByteBuffer) WrGrow(size int) {
	if size > b.WrSize() {
		b.wrReserve(size)
	}
}

func (b *ByteBuffer) RdBuf() []byte {
	if b.readerIndex >= len(b.buffers) {
		return nil
	}
	return b.buffers[b.readerIndex:]
}

func (b *ByteBuffer) RdReady() bool {
	return b.writerIndex > b.readerIndex
}

func (b *ByteBuffer) RdSize() int {
	return b.writerIndex - b.readerIndex
}

func (b *ByteBuffer) RdFlip(size int) {
	if size < b.RdSize() {
		b.readerIndex += size
	} else {
		b.Reset()
	}
}

func (b *ByteBuffer) Reset() {
	b.readerIndex = b.prependSize
	b.writerIndex = b.prependSize
}

func (b *ByteBuffer) MaxSize() int {
	return len(b.buffers)
}

func (b *ByteBuffer) wrReserve(size int) {
	if b.WrSize()+b.readerIndex < size+b.prependSize {
		tmpBuff := make([]byte, b.writerIndex+size)
		copy(tmpBuff, b.buffers)
		b.buffers = tmpBuff
	} else {
		readable := b.RdSize()
		copy(b.buffers[b.prependSize:], b.buffers[b.readerIndex:b.writerIndex])
		b.readerIndex = b.prependSize
		b.writerIndex = b.readerIndex + readable
	}
}

func (b *ByteBuffer) Prepend(buff []byte) bool {
	size := len(buff)
	if b.readerIndex < size {
		return false
	}
	b.readerIndex -= size
	copy(b.buffers[b.readerIndex:], buff)
	return true
}
