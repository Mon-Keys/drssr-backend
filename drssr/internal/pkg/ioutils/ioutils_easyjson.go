// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package ioutils

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson77835261DecodeDrssrInternalPkgIoutils(in *jlexer.Lexer, out *ModelError) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "error":
			out.Error = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson77835261EncodeDrssrInternalPkgIoutils(out *jwriter.Writer, in ModelError) {
	out.RawByte('{')
	first := true
	_ = first
	if in.Error != "" {
		const prefix string = ",\"error\":"
		first = false
		out.RawString(prefix[1:])
		out.String(string(in.Error))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ModelError) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson77835261EncodeDrssrInternalPkgIoutils(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ModelError) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson77835261EncodeDrssrInternalPkgIoutils(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ModelError) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson77835261DecodeDrssrInternalPkgIoutils(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ModelError) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson77835261DecodeDrssrInternalPkgIoutils(l, v)
}