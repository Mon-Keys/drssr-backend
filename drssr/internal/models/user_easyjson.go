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

func easyjson9e1087fdDecodeDrssrInternalModels(in *jlexer.Lexer, out *User) {
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
		case "nickname":
			out.Nickname = string(in.String())
		case "email":
			out.Email = string(in.String())
		case "avatar":
			out.Avatar = string(in.String())
		case "name":
			out.Name = string(in.String())
		case "stylist":
			out.Stylist = bool(in.Bool())
		case "age":
			out.Age = int(in.Int())
		case "description":
			out.Desc = string(in.String())
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
func easyjson9e1087fdEncodeDrssrInternalModels(out *jwriter.Writer, in User) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.Uint64(uint64(in.ID))
	}
	{
		const prefix string = ",\"nickname\":"
		out.RawString(prefix)
		out.String(string(in.Nickname))
	}
	{
		const prefix string = ",\"email\":"
		out.RawString(prefix)
		out.String(string(in.Email))
	}
	{
		const prefix string = ",\"avatar\":"
		out.RawString(prefix)
		out.String(string(in.Avatar))
	}
	if in.Name != "" {
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"stylist\":"
		out.RawString(prefix)
		out.Bool(bool(in.Stylist))
	}
	{
		const prefix string = ",\"age\":"
		out.RawString(prefix)
		out.Int(int(in.Age))
	}
	if in.Desc != "" {
		const prefix string = ",\"description\":"
		out.RawString(prefix)
		out.String(string(in.Desc))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v User) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson9e1087fdEncodeDrssrInternalModels(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v User) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9e1087fdEncodeDrssrInternalModels(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *User) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9e1087fdDecodeDrssrInternalModels(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *User) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9e1087fdDecodeDrssrInternalModels(l, v)
}
func easyjson9e1087fdDecodeDrssrInternalModels1(in *jlexer.Lexer, out *UpdateUserReq) {
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
		case "nickname":
			out.Nickname = string(in.String())
		case "email":
			out.Email = string(in.String())
		case "avatar":
			out.Avatar = string(in.String())
		case "name":
			out.Name = string(in.String())
		case "birth_date":
			out.BirthDate = string(in.String())
		case "description":
			out.Desc = string(in.String())
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
func easyjson9e1087fdEncodeDrssrInternalModels1(out *jwriter.Writer, in UpdateUserReq) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"nickname\":"
		out.RawString(prefix[1:])
		out.String(string(in.Nickname))
	}
	{
		const prefix string = ",\"email\":"
		out.RawString(prefix)
		out.String(string(in.Email))
	}
	{
		const prefix string = ",\"avatar\":"
		out.RawString(prefix)
		out.String(string(in.Avatar))
	}
	if in.Name != "" {
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"birth_date\":"
		out.RawString(prefix)
		out.String(string(in.BirthDate))
	}
	if in.Desc != "" {
		const prefix string = ",\"description\":"
		out.RawString(prefix)
		out.String(string(in.Desc))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v UpdateUserReq) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson9e1087fdEncodeDrssrInternalModels1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v UpdateUserReq) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9e1087fdEncodeDrssrInternalModels1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *UpdateUserReq) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9e1087fdDecodeDrssrInternalModels1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *UpdateUserReq) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9e1087fdDecodeDrssrInternalModels1(l, v)
}
func easyjson9e1087fdDecodeDrssrInternalModels2(in *jlexer.Lexer, out *StatusCheckStruct) {
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
		case "user_total":
			out.UserTotal = int(in.Int())
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
func easyjson9e1087fdEncodeDrssrInternalModels2(out *jwriter.Writer, in StatusCheckStruct) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"user_total\":"
		out.RawString(prefix[1:])
		out.Int(int(in.UserTotal))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v StatusCheckStruct) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson9e1087fdEncodeDrssrInternalModels2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v StatusCheckStruct) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9e1087fdEncodeDrssrInternalModels2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *StatusCheckStruct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9e1087fdDecodeDrssrInternalModels2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *StatusCheckStruct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9e1087fdDecodeDrssrInternalModels2(l, v)
}
func easyjson9e1087fdDecodeDrssrInternalModels3(in *jlexer.Lexer, out *SignupCredentials) {
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
		case "nickname":
			out.Nickname = string(in.String())
		case "email":
			out.Email = string(in.String())
		case "password":
			out.Password = string(in.String())
		case "name":
			out.Name = string(in.String())
		case "avatar":
			out.Avatar = string(in.String())
		case "birth_date":
			out.BirthDate = string(in.String())
		case "description":
			out.Desc = string(in.String())
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
func easyjson9e1087fdEncodeDrssrInternalModels3(out *jwriter.Writer, in SignupCredentials) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"nickname\":"
		out.RawString(prefix[1:])
		out.String(string(in.Nickname))
	}
	{
		const prefix string = ",\"email\":"
		out.RawString(prefix)
		out.String(string(in.Email))
	}
	{
		const prefix string = ",\"password\":"
		out.RawString(prefix)
		out.String(string(in.Password))
	}
	if in.Name != "" {
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"avatar\":"
		out.RawString(prefix)
		out.String(string(in.Avatar))
	}
	{
		const prefix string = ",\"birth_date\":"
		out.RawString(prefix)
		out.String(string(in.BirthDate))
	}
	if in.Desc != "" {
		const prefix string = ",\"description\":"
		out.RawString(prefix)
		out.String(string(in.Desc))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v SignupCredentials) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson9e1087fdEncodeDrssrInternalModels3(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v SignupCredentials) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9e1087fdEncodeDrssrInternalModels3(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *SignupCredentials) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9e1087fdDecodeDrssrInternalModels3(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *SignupCredentials) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9e1087fdDecodeDrssrInternalModels3(l, v)
}
func easyjson9e1087fdDecodeDrssrInternalModels4(in *jlexer.Lexer, out *LoginCredentials) {
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
		case "login":
			out.Login = string(in.String())
		case "password":
			out.Password = string(in.String())
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
func easyjson9e1087fdEncodeDrssrInternalModels4(out *jwriter.Writer, in LoginCredentials) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"login\":"
		out.RawString(prefix[1:])
		out.String(string(in.Login))
	}
	{
		const prefix string = ",\"password\":"
		out.RawString(prefix)
		out.String(string(in.Password))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v LoginCredentials) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson9e1087fdEncodeDrssrInternalModels4(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v LoginCredentials) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9e1087fdEncodeDrssrInternalModels4(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *LoginCredentials) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9e1087fdDecodeDrssrInternalModels4(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *LoginCredentials) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9e1087fdDecodeDrssrInternalModels4(l, v)
}
