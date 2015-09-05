package edn

import (
	"bytes"
)

func newline(dst *bytes.Buffer, prefix, indent string, depth int) {
	dst.WriteByte('\n')
	dst.WriteString(prefix)
	for i := 0; i < depth; i++ {
		dst.WriteString(indent)
	}
}

// Indent appends to dst an indented form of the EDN-encoded src. Each EDN
// collection begins on a new, indented line beginning with prefix followed by
// one or more copies of indent according to the indentation nesting. The data
// appended to dst does not begin with the prefix nor any indentation, and has
// no trailing newline, to make it easier to embed inside other formatted EDN
// data.
//
// Indent filters away whitespace, including comments and discards.
func Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error {
	origLen := dst.Len()
	var lex lexer
	lex.reset()
	tokStack := newTokenStack()
	curType := tokenError
	curSize := 0
	d := NewDecoder(bytes.NewBuffer(src))
	depth := 0
	for {
		bs, tt, err := d.nextToken()
		if err != nil {
			dst.Truncate(origLen)
			return err
		}
		err = tokStack.push(tt)
		if err != nil {
			dst.Truncate(origLen)
			return err
		}
		prevType := curType
		prevSize := curSize
		if len(tokStack.toks) > 0 {
			curType = tokStack.peek()
			curSize = tokStack.peekCount()
		}
		switch tt {
		case tokenMapStart, tokenVectorStart, tokenListStart, tokenSetStart:
			if depth > 0 {
				newline(dst, prefix, indent, depth)
			}
			dst.Write(bs)
			depth++
		case tokenVectorEnd, tokenListEnd, tokenMapEnd: // tokenSetEnd == tokenMapEnd
			depth--
			if prevSize > 0 { // suppress indent for empty collections
				newline(dst, prefix, indent, depth)
			}
			// all of these are of length 1 in bytes, so utilise this for perf
			dst.WriteByte(bs[0])
		case tokenTag:
			// need to know what the previous type was.
			switch prevType {
			case tokenMapStart:
				if prevSize%2 == 0 { // If previous size modulo 2 is equal to 0, we're a key
					if prevSize > 0 {
						dst.WriteByte(',')
					}
					newline(dst, prefix, indent, depth)
				} else { // We're a value, add a space after the key
					dst.WriteByte(' ')
				}
				dst.Write(bs)
				dst.WriteByte(' ')
			case tokenSetStart, tokenVectorStart, tokenListStart:
				newline(dst, prefix, indent, depth)
				dst.Write(bs)
				dst.WriteByte(' ')
			default: // tokenError or nested tag
				dst.Write(bs)
				dst.WriteByte(' ')
			}
		default:
			switch prevType {
			case tokenMapStart:
				if prevSize%2 == 0 { // If previous size modulo 2 is equal to 0, we're a key
					if prevSize > 0 {
						dst.WriteByte(',')
					}
					newline(dst, prefix, indent, depth)
				} else { // We're a value, add a space after the key
					dst.WriteByte(' ')
				}
				dst.Write(bs)
			case tokenSetStart, tokenVectorStart, tokenListStart:
				newline(dst, prefix, indent, depth)
				dst.Write(bs)
			default: // toplevel or nested tag. This should collapse the whole tag tower
				dst.Write(bs)
			}
		}
		if tokStack.done() {
			break
		}
	}
	return nil
}
