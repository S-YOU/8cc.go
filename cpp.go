package main

var macros = make(map[string]TokenList)
var buffer = make(TokenList, 0)
var altbuffer TokenList = nil
var bol = true

func list_append(a TokenList, b TokenList) TokenList{
	r := a
	for _, tok := range b {
		r = append(r, tok)
	}
	return r
}

func read_ident2() *Token {
	r := read_cpp_token()
	if !r.is_ident_type() {
		errorf("identifier expected, but got %s", r)
	}
	return r
}

func expand(hideset *Dict, tok *Token) *Token {
	if !tok.is_ident_type() {
		return tok
	}
	if hideset.Get(tok.sval) != nil {
		return tok
	}
	body, ok := macros[tok.sval]
	if !ok {
		return tok
	}
	hideset.Put(tok.sval, &DictValue{})
	buffer = list_append(buffer, body)
	return read_token_int2(hideset, false)
}

func expect_newine() {
	tok := read_cpp_token()
	if tok == nil || !tok.is_newline() {
		errorf("Newline expected, but got %s", tok)
	}
}

func read_undef() {
	name := read_ident2()
	expect_newine()
	delete(macros, name.sval)
}

func read_define() {
	name := read_ident2()
	body := make(TokenList, 0)
	for {
		tok := read_cpp_token()
		if tok == nil || tok.is_newline() {
			break
		}
		body = append(body, tok)
	}
	macros[name.sval] = body
}

func read_line() TokenList {
	var r TokenList
	for {
		tok := read_token_int2(NewDict(), true)
		if tok == nil {
			return r
		}
		r = append(r, tok)
	}
}

func list_reverse(a TokenList) TokenList {
	for i := len(a)/2-1; i >= 0; i-- {
		opp := len(a)-1-i
		a[i], a[opp] = a[opp], a[i]
	}
	return a
}

func read_constexpr() bool {
	altbuffer = list_reverse(read_line())
	expr := read_expr()
	altbuffer = nil
	return eval_intexpr(expr) != 0
}

func read_if() {
	cond := read_constexpr()
	if !cond {
		skip_cond_incl()
	}
}

func read_else() {
	expect_newine()
}

func read_elif() {
	expect_newine()
}

func read_endif() {
	expect_newine()
}

func read_directive() {
	tok := read_cpp_token()
	if tok.is_ident("define") {
		read_define()
	} else if tok.is_ident("undef") {
		read_undef()
	} else if tok.is_ident("if") {
		read_if()
	} else if tok.is_ident("else") {
		read_else()
	} else if tok.is_ident("elif") {
		read_elif()
	} else if tok.is_ident("endif") {
		read_endif()
	} else {
		errorf("unsupported preprocessor directive: %s", tok)
	}
}

func unget_token(tok *Token) {
	if altbuffer != nil {
		altbuffer = append(altbuffer, tok)
	} else {
		buffer = append(buffer, tok)
	}
}

func peek_token() *Token {
	tok := read_token()
	unget_token(tok)
	return tok
}

func get_token() *Token {
	if altbuffer != nil {
		if (len(altbuffer) == 0) {
			return nil
		}
		tok := altbuffer[len(altbuffer)-1]
		altbuffer = altbuffer[:len(altbuffer)-1]
		return tok
	}

	var tok *Token
	if len(buffer) > 0 {
		// list_pop
		tok = buffer[len(buffer)-1]
		buffer = buffer[:len(buffer)-1]
	} else {
		tok = read_cpp_token()
	}

	return tok
}

func read_token_int2(hideset *Dict, return_at_eol bool) *Token {
	for {
		tok := get_token()
		if tok == nil {
			return nil
		}
		if tok != nil && tok.is_newline() {
			bol = true
			if return_at_eol {
				return nil
			}
			continue
		}
		if bol && tok.is_punct('#') {
			read_directive()
			bol = true
			continue
		}
		bol = false
		return expand(hideset, tok)
	}
}

func read_token() *Token {
	return read_token_int2(NewDict(), false)
}
