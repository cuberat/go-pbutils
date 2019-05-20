

# pbutils
`import "github.com/cuberat/go-pbutils/pbutils"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>
The pbutils package provides various utilities for working with protobuffers
in Go.

Installation


	go get github.com/cuberat/go-pbutils/pbutils




## <a name="pkg-index">Index</a>
* [Constants](#pkg-constants)
* [type PBKRCodec](#PBKRCodec)
  * [func NewPBKRCodec(marshal_type interface{}) *PBKRCodec](#NewPBKRCodec)
  * [func (pbkr *PBKRCodec) CodecSame() bool](#PBKRCodec.CodecSame)
  * [func (pbkr *PBKRCodec) JoinKV(key, val []byte) ([]byte, error)](#PBKRCodec.JoinKV)
  * [func (pbkr *PBKRCodec) MarshalVal(data interface{}) ([]byte, error)](#PBKRCodec.MarshalVal)
  * [func (pbkr *PBKRCodec) SplitKV(wire_data []byte) ([]byte, []byte, error)](#PBKRCodec.SplitKV)
  * [func (pbkr *PBKRCodec) UnmarshalVal(val_bytes []byte) (interface{}, error)](#PBKRCodec.UnmarshalVal)
* [type PBKRScanner](#PBKRScanner)
  * [func NewPBKRScanner(r io.Reader, marshal_type interface{}) *PBKRScanner](#NewPBKRScanner)
  * [func NewPBKRScannerWithDecoder(r io.Reader, decoder libutils.KeyedRecordDecoder) *PBKRScanner](#NewPBKRScannerWithDecoder)
  * [func (krs *PBKRScanner) Err() error](#PBKRScanner.Err)
  * [func (krs *PBKRScanner) Record() *libutils.KeyedRecord](#PBKRScanner.Record)
  * [func (krs *PBKRScanner) Scan() bool](#PBKRScanner.Scan)
* [type PBKRWriter](#PBKRWriter)
  * [func NewPBKRWriter(w io.Writer, marshal_type interface{}) *PBKRWriter](#NewPBKRWriter)
  * [func NewPBKRWriterWithEncoder(w io.Writer, encoder libutils.KeyedRecordEncoder) *PBKRWriter](#NewPBKRWriterWithEncoder)
  * [func (krw *PBKRWriter) Write(rec *libutils.KeyedRecord) (int, error)](#PBKRWriter.Write)


#### <a name="pkg-files">Package files</a>
[pbutils.go](/src/github.com/cuberat/go-pbutils/pbutils/pbutils.go) 


## <a name="pkg-constants">Constants</a>
``` go
const (
    Version = "0.01"
)
```




## <a name="PBKRCodec">type</a> [PBKRCodec](/src/target/pbutils.go?s=1998:2053#L42)
``` go
type PBKRCodec struct {
    // contains filtered or unexported fields
}
```
Implements the KeyedRecordEncoder and KeyedRecordDecoder interfaces specified
by `github.com/cuberat/go-libutils/libutils`.

This codec encodes and decodes delimited keyed records where the value is a
protobuffer.







### <a name="NewPBKRCodec">func</a> [NewPBKRCodec](/src/target/pbutils.go?s=2323:2379#L50)
``` go
func NewPBKRCodec(marshal_type interface{}) *PBKRCodec
```
Returns a codec for keyed records where key is a varint length-prefixed
string, with the value immediately following. The value is serialized as a
protobuffer structure. `marshal_type` is a (reference to) a struct generated
by the `protoc-gen-go` utility.





### <a name="PBKRCodec.CodecSame">func</a> (\*PBKRCodec) [CodecSame](/src/target/pbutils.go?s=4931:4970#L135)
``` go
func (pbkr *PBKRCodec) CodecSame() bool
```
Returns true so that if this codec is used for both encoder and decoder,
unnecessary re-serialization can be avoided.

This allows for lazy encoding. That is, if the raw record bytes that were
read in do not need to change, they can be written back out as-is, instead of
actually re-encoding.




### <a name="PBKRCodec.JoinKV">func</a> (\*PBKRCodec) [JoinKV](/src/target/pbutils.go?s=4016:4078#L109)
``` go
func (pbkr *PBKRCodec) JoinKV(key, val []byte) ([]byte, error)
```
Joins the key and value bytes, returning the serialized record. The length of
the key is encoded as a varint, and the returned data is organized like so:


	<key_len><key><val>




### <a name="PBKRCodec.MarshalVal">func</a> (\*PBKRCodec) [MarshalVal](/src/target/pbutils.go?s=4387:4454#L121)
``` go
func (pbkr *PBKRCodec) MarshalVal(data interface{}) ([]byte, error)
```
Serializes the value data structure.




### <a name="PBKRCodec.SplitKV">func</a> (\*PBKRCodec) [SplitKV](/src/target/pbutils.go?s=3295:3371#L84)
``` go
func (pbkr *PBKRCodec) SplitKV(wire_data []byte) ([]byte, []byte,
    error)
```
Splits the record, returning the key and the serialized value data
structure. The record is expected to be organized like so:


	<key_len><key><val>

where `key_len` is encoded as a varint.




### <a name="PBKRCodec.UnmarshalVal">func</a> (\*PBKRCodec) [UnmarshalVal](/src/target/pbutils.go?s=2736:2814#L65)
``` go
func (pbkr *PBKRCodec) UnmarshalVal(val_bytes []byte) (interface{},
    error)
```
Deserializes the value.




## <a name="PBKRScanner">type</a> [PBKRScanner](/src/target/pbutils.go?s=5181:5275#L142)
``` go
type PBKRScanner struct {
    // contains filtered or unexported fields
}
```
Implements the KeyedRecordScanner interface specified by
`github.com/cuberat/go-libutils/libutils`. This scanner assumes each record
is prefixed with a length encoded as a varint.







### <a name="NewPBKRScanner">func</a> [NewPBKRScanner](/src/target/pbutils.go?s=5657:5730#L155)
``` go
func NewPBKRScanner(r io.Reader, marshal_type interface{}) *PBKRScanner
```
Returns a KeyedRecordScanner that assumes each record is prefixed with a
length encoded as a varint. Each record itself is expected contain a key that
is length-prefixed with a varint, with the serialized (as a protobuffer)
value immediately following. That is:


	<record_len><key_len><key><value>

where `record_len` and `key_len` are encoded as varints.


### <a name="NewPBKRScannerWithDecoder">func</a> [NewPBKRScannerWithDecoder](/src/target/pbutils.go?s=5827:5926#L160)
``` go
func NewPBKRScannerWithDecoder(r io.Reader,
    decoder libutils.KeyedRecordDecoder) *PBKRScanner
```




### <a name="PBKRScanner.Err">func</a> (\*PBKRScanner) [Err](/src/target/pbutils.go?s=6673:6708#L186)
``` go
func (krs *PBKRScanner) Err() error
```
Returns the first non-EOF error that was encountered by the Scanner.




### <a name="PBKRScanner.Record">func</a> (\*PBKRScanner) [Record](/src/target/pbutils.go?s=6341:6397#L177)
``` go
func (krs *PBKRScanner) Record() *libutils.KeyedRecord
```
Returns the most recent serialized record generated by a call to Scan().




### <a name="PBKRScanner.Scan">func</a> (\*PBKRScanner) [Scan](/src/target/pbutils.go?s=6194:6229#L172)
``` go
func (krs *PBKRScanner) Scan() bool
```
Advances the scanner to the next record. It returns false when the scan
stops, either by reaching the end of the input or an error.




## <a name="PBKRWriter">type</a> [PBKRWriter](/src/target/pbutils.go?s=6851:6967#L192)
``` go
type PBKRWriter struct {
    // contains filtered or unexported fields
}
```
Implements the `libutils.KeyedRecordWriter` interface from
`github.com/cuberat/go-libutils/libutils`.







### <a name="NewPBKRWriter">func</a> [NewPBKRWriter](/src/target/pbutils.go?s=7092:7163#L200)
``` go
func NewPBKRWriter(w io.Writer, marshal_type interface{}) *PBKRWriter
```
Returns a `libutils.KeyedRecordWriter` that outputs length-prefixed records
where the length is encoded as a varint.


### <a name="NewPBKRWriterWithEncoder">func</a> [NewPBKRWriterWithEncoder](/src/target/pbutils.go?s=7333:7430#L206)
``` go
func NewPBKRWriterWithEncoder(w io.Writer,
    encoder libutils.KeyedRecordEncoder) *PBKRWriter
```
Returns a `libutils.KeyedRecordWriter` that uses the provided encoder.





### <a name="PBKRWriter.Write">func</a> (\*PBKRWriter) [Write](/src/target/pbutils.go?s=7603:7671#L217)
``` go
func (krw *PBKRWriter) Write(rec *libutils.KeyedRecord) (int, error)
```
Serializes and outputs the KeyedRecord








- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
