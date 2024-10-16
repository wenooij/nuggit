package api

type Args struct {
	Bool    bool    `json:"bool,omitempty"`
	Int64   int64   `json:"int64,omitempty"`
	Uint64  uint64  `json:"uint64,omitempty"`
	Float64 float64 `json:"float64,omitempty"`
	Bytes   []byte  `json:"bytes,omitempty"`
	String  string  `json:"string,omitempty"`
}

func (a *Args) GetBool() bool {
	if a == nil {
		return false
	}
	return a.Bool
}
func (a *Args) GetInt64() int64 {
	if a == nil {
		return 0
	}
	return a.Int64
}
func (a *Args) GetUint64() uint64 {
	if a == nil {
		return 0
	}
	return a.Uint64
}
func (a *Args) GetFloat64() float64 {
	if a == nil {
		return 0
	}
	return a.Float64
}
func (a *Args) GetBytes() []byte {
	if a == nil {
		return nil
	}
	return a.Bytes
}
func (a *Args) GetString() string {
	if a == nil {
		return ""
	}
	return a.String
}

type BatchArgs struct {
	*Args      `json:",omitempty"`
	Bools      []bool    `json:"bools,omitempty"`
	Int64s     []int64   `json:"int64s,omitempty"`
	Uint64s    []uint64  `json:"uint64s,omitempty"`
	Float64s   []float64 `json:"float64s,omitempty"`
	BytesBatch [][]byte  `json:"bytes_batch,omitempty"`
	Strings    []string  `json:"strings,omitempty"`
}

func (a *BatchArgs) GetBools() []bool {
	if a == nil {
		return nil
	}
	return a.Bools
}
func (a *BatchArgs) GetInt64s() []int64 {
	if a == nil {
		return nil
	}
	return a.Int64s
}
func (a *BatchArgs) GetUint64s() []uint64 {
	if a == nil {
		return nil
	}
	return a.Uint64s
}
func (a *BatchArgs) GetFloat64s() []float64 {
	if a == nil {
		return nil
	}
	return a.Float64s
}
func (a *BatchArgs) GetBytesBatch() [][]byte {
	if a == nil {
		return nil
	}
	return a.BytesBatch
}
func (a *BatchArgs) GetStrings() []string {
	if a == nil {
		return nil
	}
	return a.Strings
}

func extract(args *BatchArgs, t Type) any {
	switch t {
	case scalar(TypeUndefined), scalar(TypeBytes):
		return args.GetBytes()
	case scalar(TypeString):
		return args.GetString()
	case scalar(TypeBool):
		return args.GetBool()
	case scalar(TypeInt64):
		return args.GetInt64()
	case scalar(TypeUint64):
		return args.GetUint64()
	case scalar(TypeFloat64):
		return args.GetFloat64()

	case repeated(TypeUndefined), repeated(TypeBytes):
		return args.GetBytesBatch()
	case repeated(TypeString):
		return args.GetStrings()
	case repeated(TypeBool):
		return args.GetBools()
	case repeated(TypeInt64):
		return args.GetInt64s()
	case repeated(TypeUint64):
		return args.GetUint64s()
	case repeated(TypeFloat64):
		return args.GetFloat64s()

	default:
		return nil
	}
}
