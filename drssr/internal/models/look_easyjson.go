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

func easyjson9db9635DecodeDrssrInternalModels(in *jlexer.Lexer, out *Look) {
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
		case "name":
			out.Name = string(in.String())
		case "description":
			out.Desc = string(in.String())
		case "creator_id":
			out.CreatorID = uint64(in.Uint64())
		case "clothes":
			if in.IsNull() {
				in.Skip()
				out.Clothes = nil
			} else {
				in.Delim('[')
				if out.Clothes == nil {
					if !in.IsDelim(']') {
						out.Clothes = make([]ClothesStruct, 0, 0)
					} else {
						out.Clothes = []ClothesStruct{}
					}
				} else {
					out.Clothes = (out.Clothes)[:0]
				}
				for !in.IsDelim(']') {
					var v1 ClothesStruct
					(v1).UnmarshalEasyJSON(in)
					out.Clothes = append(out.Clothes, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "img_path":
			out.ImgPath = string(in.String())
		case "img":
			out.Img = string(in.String())
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
func easyjson9db9635EncodeDrssrInternalModels(out *jwriter.Writer, in Look) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.Uint64(uint64(in.ID))
	}
	{
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"description\":"
		out.RawString(prefix)
		out.String(string(in.Desc))
	}
	{
		const prefix string = ",\"creator_id\":"
		out.RawString(prefix)
		out.Uint64(uint64(in.CreatorID))
	}
	{
		const prefix string = ",\"clothes\":"
		out.RawString(prefix)
		if in.Clothes == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.Clothes {
				if v2 > 0 {
					out.RawByte(',')
				}
				(v3).MarshalEasyJSON(out)
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"img_path\":"
		out.RawString(prefix)
		out.String(string(in.ImgPath))
	}
	{
		const prefix string = ",\"img\":"
		out.RawString(prefix)
		out.String(string(in.Img))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Look) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson9db9635EncodeDrssrInternalModels(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Look) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9db9635EncodeDrssrInternalModels(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Look) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9db9635DecodeDrssrInternalModels(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Look) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9db9635DecodeDrssrInternalModels(l, v)
}
func easyjson9db9635DecodeDrssrInternalModels1(in *jlexer.Lexer, out *CoordsStruct) {
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
		case "x":
			out.X = int(in.Int())
		case "y":
			out.Y = int(in.Int())
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
func easyjson9db9635EncodeDrssrInternalModels1(out *jwriter.Writer, in CoordsStruct) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"x\":"
		out.RawString(prefix[1:])
		out.Int(int(in.X))
	}
	{
		const prefix string = ",\"y\":"
		out.RawString(prefix)
		out.Int(int(in.Y))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v CoordsStruct) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson9db9635EncodeDrssrInternalModels1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v CoordsStruct) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9db9635EncodeDrssrInternalModels1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *CoordsStruct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9db9635DecodeDrssrInternalModels1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *CoordsStruct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9db9635DecodeDrssrInternalModels1(l, v)
}
func easyjson9db9635DecodeDrssrInternalModels2(in *jlexer.Lexer, out *ClothesStruct) {
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
		case "name":
			out.Name = string(in.String())
		case "description":
			out.Desc = string(in.String())
		case "type":
			out.Type = string(in.String())
		case "brand":
			out.Brand = string(in.String())
		case "coords":
			(out.Coords).UnmarshalEasyJSON(in)
		case "img_path":
			out.ImgPath = string(in.String())
		case "mask_path":
			out.MaskPath = string(in.String())
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
func easyjson9db9635EncodeDrssrInternalModels2(out *jwriter.Writer, in ClothesStruct) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.Uint64(uint64(in.ID))
	}
	{
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"description\":"
		out.RawString(prefix)
		out.String(string(in.Desc))
	}
	{
		const prefix string = ",\"type\":"
		out.RawString(prefix)
		out.String(string(in.Type))
	}
	{
		const prefix string = ",\"brand\":"
		out.RawString(prefix)
		out.String(string(in.Brand))
	}
	{
		const prefix string = ",\"coords\":"
		out.RawString(prefix)
		(in.Coords).MarshalEasyJSON(out)
	}
	{
		const prefix string = ",\"img_path\":"
		out.RawString(prefix)
		out.String(string(in.ImgPath))
	}
	{
		const prefix string = ",\"mask_path\":"
		out.RawString(prefix)
		out.String(string(in.MaskPath))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ClothesStruct) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson9db9635EncodeDrssrInternalModels2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ClothesStruct) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9db9635EncodeDrssrInternalModels2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ClothesStruct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9db9635DecodeDrssrInternalModels2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ClothesStruct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9db9635DecodeDrssrInternalModels2(l, v)
}
func easyjson9db9635DecodeDrssrInternalModels3(in *jlexer.Lexer, out *ArrayLooks) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		in.Skip()
		*out = nil
	} else {
		in.Delim('[')
		if *out == nil {
			if !in.IsDelim(']') {
				*out = make(ArrayLooks, 0, 0)
			} else {
				*out = ArrayLooks{}
			}
		} else {
			*out = (*out)[:0]
		}
		for !in.IsDelim(']') {
			var v4 Look
			(v4).UnmarshalEasyJSON(in)
			*out = append(*out, v4)
			in.WantComma()
		}
		in.Delim(']')
	}
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson9db9635EncodeDrssrInternalModels3(out *jwriter.Writer, in ArrayLooks) {
	if in == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
		out.RawString("null")
	} else {
		out.RawByte('[')
		for v5, v6 := range in {
			if v5 > 0 {
				out.RawByte(',')
			}
			(v6).MarshalEasyJSON(out)
		}
		out.RawByte(']')
	}
}

// MarshalJSON supports json.Marshaler interface
func (v ArrayLooks) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson9db9635EncodeDrssrInternalModels3(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ArrayLooks) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9db9635EncodeDrssrInternalModels3(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ArrayLooks) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9db9635DecodeDrssrInternalModels3(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ArrayLooks) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9db9635DecodeDrssrInternalModels3(l, v)
}
