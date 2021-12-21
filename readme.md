Spec: Fast Alloc-Free Binary Format
===================================

# 1. Specification

## 1.0 Types
Scalars:
- `nil`
- `bool`
- ints: `int8`, `int16`, `int32`, `int64`
- uints: `uint8`/`byte`, `uint16`, `uint32`, `uint64`
- floats: `float32`, `float64`

Size-based:
- `bytes`
- `string`

Messages:
- `message`

Collections:
- lists: `list<element>`

More:
- enums
- uids: `u128`, `u256`, `uuid`
- timestamp: `timestamp`, `timestamptz`
- maps: `map<key, value>`

## 1.1 Constants
Constants allow to specify constant values, only support bool/int/string.

```spec
const (
    TypeUndefined = ""
    TypeUser = "user"
    TypeMachine = "machine"
)
```

## 1.2. Enums
Enums are generated as constants. Enums can be specified only with primitive values.

```spec
enum Type string {
    Undefined = ""
    User = "user"
    Machine = "machine"
}

enum Type uint32 {
    Undefined = 0
    User = 1
    Machine = 2
}
```

## 1.3. Structs
Structs hold a set of fields. Each field is specified by a unique tag (number). 
Fields can be added/removed/renamed, but their numbers must stay the same.
Tags can be auto-generated by the formatter and can be grouped into sections.

```spec
type User message {
    id          u128    1
    name        string  2
    age         int32   3
    email       Email   4

    // next section, tags start at 10
    created_at  int64   10
    updated_at  int64   11
    deleted_at  int64   12
}

type Email message {
    id      u128    1
    user_id u128    2
    email   string  3
}
```

## 1.5 Modules
Modules group multiple files in a directory into a single namespace.
- Directory is a module.
- Imports always use full path.
- Import may specify an alias as `alias "import/path"`.
- Compiler accepts multiple import paths.
- Compiler import paths specify mappings between directories and (search) paths: `-I dir:path`


```spec
module mymodule

import (
    "github.com/hello/world"
    uni "github.com/hello/universe"
)

```

## 1.6 More: Values
Values are immutable structs which cannot be changed later. 
Their fields automatically get their numbers.

```
type Point value {
    x int64
    y int64
}
```

# 2. Format
Objects are written in reverse order with the last byte specifying their type. 

- Children (nested objects) are written before their parents.
- Floats are encoded in IEEE-754 format.
- Everything is big-endian.

Example `"hello, world"`:
```
+--------------------------+---------------+---------------+
|  data (including last 0) | size (uint32) |  type (uint8) |
+--------------------------+---------------+---------------+
|      hello, world\0      |       13      |      0x50     |
+--------------------------+---------------+---------------+
```


Format:
```
type byte

const {
    typeNil   = 0x00
    typeTrue  = 0x01
    typeFalse = 0x02
    
    typeInt8  = 0x10
    typeInt16 = 0x11
    typeInt32 = 0x12
    typeInt64 = 0x13
    
    typeUInt8  = 0x20  // byte
    typeUInt16 = 0x21
    typeUInt32 = 0x22
    typeUInt64 = 0x23
    
    typeFloat32 = 0x30
    typeFloat64 = 0x31
    
    typeBytes  = 0x40
    typeString = 0x41

    typeList    = 0x50
	typeListBig = 0x51

    typeMessage    = 0x60
	typeMessageBig = 0x61
}

// value holds any value, it is a union of a value body and a type.
// value is written and read in reverse order.
value struct {
    union {
        int8  int8
        int16 varint
        int32 varint
        int64 varint
        
        uint8  uint8
        uint16 varint
        uint32 varint
        uint64 varint
        
        float32 float32
        float64 float64
        
        bytes {
            data  []byte
            size  uint32
        }
        
        string {
            data  []byte
            zero  byte
            size  varint // size without zero byte
        }

        list {
            body      []byte    // element values by offsets
            table     listTable // element offsets by indexes
            bodySize  varint
            tableSize varint
        }
        
        message {
            body       []byte       // field values by offsets
            table      messageTable // field offsets ordered by tags
            bodySize   varint
            tableSize  varint
        }
    }

    type byte
}

// listTable holds list element offsets ordered by index.
// table can be small/big. big table holds elements with offsets > uint16.
listTable union {
    small []uint16
    big   []uint32
}

// messageTable holds message field offsets sorted by tags.
// table can be small/big. big table holds fields with tags > uint8 or offsets > uint16.
messageTable union {
    small []field { 
        tag    uint8  // field tag
        offset uint16 // field value end offset relative to body start
    }

    big []field {
        tag    uint16 // field tag
        offset uint32 // field value end offset relative to body start
    }
}
```


# 3. Implementation
## 3.0 Readers
Compiler generates type readers which wrap buffers. Readers are passed by value as slices in Go.
Readers are lazy and do not read anything until required.

```spec
type Block struct {
    id      BlockID     1
    head    BlockHead   2
    body    BlockBody   3
}

type BlockID value {
    index   int64
    hash    u256
}

type BlockHead struct {
    type        BlockType   1
    index       int64       2
    parent      BlockID     3

    base        BlockID     10
    parent      *BlockID    11
}
```

```go
func readBlock(buffer []byte) {
    // get block reader
    block := blockchain.NewBlockReader(buffer)
    id := block.ID()
    fmt.Println("Processing block: ", id)

    // read head
    head := block.Head()
    base := head.Base()
    parent := head.Parent()

    fmt.Println("Block base: ", base)
    fmt.Println("Block parent: ", parent)

    // read body
    body := block.Body()
    readBody(body)
}
```

## 3.1 Writers
Compiler generates type writers which wrap buffers. Writers are passed by value as slices in Go.
Internally, write buffers hold a table buffer and a data buffer.

```go
func writeBlock(prev blockchain.Block, buffer spec.WriteBuffer) ([]byte, error) {
    // get block writer
    block := blockchain.NewBlockWriter(buffer)
    block.ID(id)

    // write head
    head := block.Head()
    head.Index(index)
    head.Type(type)
    head.Base(base)
    head.Parent(parent)
    head.End()

    // write body
    body := block.Body()
    body.Test(bytes)
    body.End()

    // or write body bytes
    block.BodyBytes(bodyBytes)

    // done
    block.End()
    return buffer.Bytes()
}
```

---

© 2021 Ivan Korobkov
