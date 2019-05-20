// BSD 2-Clause License
//
// Copyright (c) 2019 Don Owens <don@regexguy.com>.  All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// The pbutils package provides various utilities for working with protobuffers
// in Go.
//
// Installation
//
//   go get github.com/cuberat/go-pbutils/pbutils
package pbutils

import (
    "bufio"
    "fmt"
    "io"
    libutils "github.com/cuberat/go-libutils/libutils"
    proto "github.com/golang/protobuf/proto"
    "reflect"
)

// Implements the KeyedRecordEncoder and KeyedRecordDecoder interfaces specified
// by `github.com/cuberat/go-libutils/libutils`.
//
// This codec encodes and decodes delimited keyed records where the value is a
// protobuffer.
type PBKRCodec struct {
    marshal_type reflect.Type
}

// Returns a codec for keyed records where key is a varint length-prefixed
// string, with the value immediately following. The value is serialized as a
// protobuffer structure. `marshal_type` is a (reference to) a struct generated
// by the `protoc-gen-go` utility.
func NewPBKRCodec(marshal_type interface{}) (*PBKRCodec) {
    pbkr := new(PBKRCodec)
    value := reflect.ValueOf(marshal_type)
    kind := value.Kind()
    for kind == reflect.Ptr || kind == reflect.Interface {
        value = value.Elem()
        kind = value.Kind()
    }
    marshal_type = value.Interface()
    pbkr.marshal_type = reflect.TypeOf(marshal_type)

    return pbkr
}

// Deserializes the value.
func (pbkr *PBKRCodec) UnmarshalVal(val_bytes []byte) (interface{},
    error) {

    data_int := reflect.New(pbkr.marshal_type).Interface()
    data, ok := data_int.(proto.Message)
    if !ok {
        return nil, fmt.Errorf("type %T not a proto.Message", pbkr.marshal_type)
    }
    err := proto.Unmarshal(val_bytes, data)

    return data, err
}

// Splits the record, returning the key and the serialized value data
// structure. The record is expected to be organized like so:
//
//     <key_len><key><val>
//
// where `key_len` is encoded as a varint.
func (pbkr *PBKRCodec) SplitKV(wire_data []byte) ([]byte, []byte,
    error) {
    key_len, vi_len, err := libutils.DecodeVarint(wire_data)
    if err != nil {
        return nil, nil, fmt.Errorf("couldn't decode varint in SplitKV(): %s",
            err)
    }

    if uint64(len(wire_data)) < uint64(vi_len) + key_len {
        return []byte{}, []byte{},
        fmt.Errorf("wire_data too short in SplitKV()")
    }

    data := wire_data[vi_len:]

    key := data[:key_len]
    val := data[key_len:]

    return key, val, nil
}

// Joins the key and value bytes, returning the serialized record. The length of
// the key is encoded as a varint, and the returned data is organized like so:
//
//    <key_len><key><val>
func (pbkr *PBKRCodec) JoinKV(key, val []byte) ([]byte, error) {
    key_len_prefix := libutils.EncodeVarint(uint64(len(key)))

    data := make([]byte, 0, len(key_len_prefix) + len(key) + len(val))
    data = append(data, key_len_prefix...)
    data = append(data, key...)
    data = append(data, val...)

    return data, nil
}

// Serializes the value data structure.
func (pbkr *PBKRCodec) MarshalVal(data interface{}) ([]byte, error) {
    pb_data, ok := data.(proto.Message)
    if !ok {
        return nil, fmt.Errorf("type %T not a proto.Message", data)
    }
    return proto.Marshal(pb_data)
}

// Returns true so that if this codec is used for both encoder and decoder,
// unnecessary re-serialization can be avoided.
//
// This allows for lazy encoding. That is, if the raw record bytes that were
// read in do not need to change, they can be written back out as-is, instead of
// actually re-encoding.
func (pbkr *PBKRCodec) CodecSame() bool {
    return true
}

// Implements the KeyedRecordScanner interface specified by
// `github.com/cuberat/go-libutils/libutils`. This scanner assumes each record
// is prefixed with a length encoded as a varint.
type PBKRScanner struct {
    scanner *bufio.Scanner
    decoder libutils.KeyedRecordDecoder
}

// Returns a KeyedRecordScanner that assumes each record is prefixed with a
// length encoded as a varint. Each record itself is expected contain a key that
// is length-prefixed with a varint, with the serialized (as a protobuffer)
// value immediately following. That is:
//
//    <record_len><key_len><key><value>
//
// where `record_len` and `key_len` are encoded as varints.
func NewPBKRScanner(r io.Reader, marshal_type interface{}) (*PBKRScanner) {
    decoder := NewPBKRCodec(marshal_type)
    return NewPBKRScannerWithDecoder(r, decoder)
}

func NewPBKRScannerWithDecoder(r io.Reader,
    decoder libutils.KeyedRecordDecoder) (*PBKRScanner) {

    pbkrs := new(PBKRScanner)
    pbkrs.scanner = libutils.NewVILPScanner(r)
    pbkrs.decoder = decoder

    return pbkrs
}

// Advances the scanner to the next record. It returns false when the scan
// stops, either by reaching the end of the input or an error.
func (krs *PBKRScanner) Scan() bool {
    return krs.scanner.Scan()
}

// Returns the most recent serialized record generated by a call to Scan().
func (krs *PBKRScanner) Record() (*libutils.KeyedRecord) {
    wire_data := krs.scanner.Bytes()
    wire_data_copy := make([]byte, len(wire_data))
    copy(wire_data_copy, wire_data)

    return libutils.NewKeyedRecordFromBytes(wire_data_copy, krs.decoder)
}

// Returns the first non-EOF error that was encountered by the Scanner.
func (krs *PBKRScanner) Err() error {
    return krs.scanner.Err()
}

// Implements the `libutils.KeyedRecordWriter` interface from
// `github.com/cuberat/go-libutils/libutils`.
type PBKRWriter struct {
    marshal_type interface{}
    encoder libutils.KeyedRecordEncoder
    writer io.Writer
}

// Returns a `libutils.KeyedRecordWriter` that outputs length-prefixed records
// where the length is encoded as a varint.
func NewPBKRWriter(w io.Writer, marshal_type interface{}) (*PBKRWriter) {
    encoder := NewPBKRCodec(marshal_type)
    return NewPBKRWriterWithEncoder(w, encoder)
}

// Returns a `libutils.KeyedRecordWriter` that uses the provided encoder.
func NewPBKRWriterWithEncoder(w io.Writer,
    encoder libutils.KeyedRecordEncoder) (*PBKRWriter) {

    writer := new(PBKRWriter)
    writer.writer = libutils.NewVILPWriter(w)
    writer.encoder = encoder

    return writer
}

// Serializes and outputs the KeyedRecord
func (krw *PBKRWriter) Write(rec *libutils.KeyedRecord) (int, error) {
    rec_out_bytes, err := rec.RecordBytesOut(krw.encoder)
    if err != nil {
        return 0, err
    }

    return krw.writer.Write(rec_out_bytes)
}
