package protocol

type Block struct {
	ID       BlockID   `tag:"1"`
	Head     BlockHead `tag:"2"`
	Body     BlockBody `tag:"3"`
	Checksum uint32    `tag:"10"`
}

type BlockID struct {
	Index int64  `tag:"1"`
	Hash  uint64 `tag:"2"`
}

type BlockHead struct {
	Type   uint8   `tag:"1"`
	Index  int64   `tag:"2"`
	Base   BlockID `tag:"3"`
	Parent BlockID `tag:"4"`
}

type BlockBody struct {
	Command string
}

// Write

func (b Block) Write(w Writer) error {
	w.BeginMessage()

	// id
	w.BeginField(1)
	b.ID.Write(w)
	w.EndField()

	// head
	w.BeginField(2)
	b.Head.Write(w)
	w.EndField()

	// body
	w.BeginField(3)
	b.Body.Write(w)
	w.EndField()

	// checksum
	w.BeginField(10)
	w.WriteUInt32(b.Checksum)
	w.EndField()

	return w.EndMessage()
}

func (id BlockID) Write(w Writer) error {
	w.BeginMessage()

	// index
	w.BeginField(1)
	w.WriteInt64(id.Index)
	w.EndField()

	// hash
	w.BeginField(2)
	w.WriteUInt64(id.Hash)
	w.EndField()

	return w.EndMessage()
}

func (h BlockHead) Write(w Writer) error {
	w.BeginMessage()

	// type
	w.BeginField(1)
	w.WriteUInt8(h.Type)
	w.EndField()

	// index
	w.BeginField(2)
	w.WriteInt64(h.Index)
	w.EndField()

	// base
	w.BeginField(3)
	h.Base.Write(w)
	w.EndField()

	// parent
	w.BeginField(4)
	h.Parent.Write(w)
	w.EndField()

	return w.EndMessage()
}

func (b BlockBody) Write(w Writer) error {
	w.BeginMessage()

	// command
	w.BeginField(1)
	w.WriteString(b.Command)
	w.EndField()

	return w.EndMessage()
}
