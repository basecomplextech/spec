package protocol

type Block struct {
	ID       BlockID   `tag:"1"`
	Head     BlockHead `tag:"2"`
	Body     BlockBody `tag:"3"`
	Checksum uint32    `tag:"10"`
}

type BlockID struct {
	Index int64
	Hash  uint64
}

type BlockHead struct {
	Type   uint8
	Index  int64
	Base   BlockID
	Parent BlockID
}

type BlockBody struct {
	Command string
}
