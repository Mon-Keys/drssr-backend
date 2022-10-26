// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package models

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

func easyjson31a459deDecodeDrssrInternalModels(in *jlexer.Lexer, out *Clothes) {
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
		case "id":
			out.ID = uint64(in.Uint64())
		case "type":
			out.Type = string(in.String())
		case "color":
			out.Color = string(in.String())
		case "img":
			out.Img = string(in.String())
		case "mask":
			out.Mask = string(in.String())
		case "brand":
			out.Brand = string(in.String())
		case "sex":
			out.Sex = string(in.String())
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
func easyjson31a459deEncodeDrssrInternalModels(out *jwriter.Writer, in Clothes) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.Uint64(uint64(in.ID))
	}
	{
		const prefix string = ",\"type\":"
		out.RawString(prefix)
		out.String(string(in.Type))
	}
	{
		const prefix string = ",\"color\":"
		out.RawString(prefix)
		out.String(string(in.Color))
	}
	{
		const prefix string = ",\"img\":"
		out.RawString(prefix)
		out.String(string(in.Img))
	}
	{
		const prefix string = ",\"mask\":"
		out.RawString(prefix)
		out.String(string(in.Mask))
	}
	{
		const prefix string = ",\"brand\":"
		out.RawString(prefix)
		out.String(string(in.Brand))
	}
	{
		const prefix string = ",\"sex\":"
		out.RawString(prefix)
		out.String(string(in.Sex))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Clothes) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson31a459deEncodeDrssrInternalModels(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Clothes) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson31a459deEncodeDrssrInternalModels(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Clothes) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson31a459deDecodeDrssrInternalModels(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Clothes) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson31a459deDecodeDrssrInternalModels(l, v)
}
func easyjson31a459deDecodeDrssrInternalModels1(in *jlexer.Lexer, out *ArrayClothes) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		in.Skip()
		*out = nil
	} else {
		in.Delim('[')
		if *out == nil {
			if !in.IsDelim(']') {
				*out = make(ArrayClothes, 0, 0)
			} else {
				*out = ArrayClothes{}
			}
		} else {
			*out = (*out)[:0]
		}
		for !in.IsDelim(']') {
			var v1 Clothes
			(v1).UnmarshalEasyJSON(in)
			*out = append(*out, v1)
			in.WantComma()
		}
		in.Delim(']')
	}
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson31a459deEncodeDrssrInternalModels1(out *jwriter.Writer, in ArrayClothes) {
	if in == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
		out.RawString("null")
	} else {
		out.RawByte('[')
		for v2, v3 := range in {
			if v2 > 0 {
				out.RawByte(',')
			}
			(v3).MarshalEasyJSON(out)
		}
		out.RawByte(']')
	}
}

// MarshalJSON supports json.Marshaler interface
func (v ArrayClothes) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson31a459deEncodeDrssrInternalModels1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ArrayClothes) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson31a459deEncodeDrssrInternalModels1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ArrayClothes) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson31a459deDecodeDrssrInternalModels1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ArrayClothes) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson31a459deDecodeDrssrInternalModels1(l, v)
}
