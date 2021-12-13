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
	w.BeginStruct()

	// id
	w.BeginField(1)
	b.ID.Write(w)

	// head
	w.BeginField(2)
	b.Head.Write(w)

	// body
	w.BeginField(3)
	b.Body.Write(w)

	// checksum
	w.BeginField(10)
	w.WriteUInt32(b.Checksum)

	return w.EndStruct()
}

func (id BlockID) Write(w Writer) error {
	w.BeginStruct()

	// index
	w.BeginField(1)
	w.WriteInt64(id.Index)

	// hash
	w.BeginField(2)
	w.WriteUInt64(id.Hash)

	return w.EndStruct()
}

func (h BlockHead) Write(w Writer) error {
	w.BeginStruct()

	// type
	w.BeginField(1)
	w.WriteUInt8(h.Type)

	// index
	w.BeginField(2)
	w.WriteInt64(h.Index)

	// base
	w.BeginField(3)
	h.Base.Write(w)

	// parent
	w.BeginField(4)
	h.Parent.Write(w)

	return w.EndStruct()
}

func (b BlockBody) Write(w Writer) error {
	w.BeginStruct()

	// command
	w.BeginField(1)
	w.WriteString(b.Command)

	return w.EndStruct()
}
