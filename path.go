// Copyright 2013 Julien Schmidt. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package httprouter

// CleanPath is the URL version of path.Clean, it returns a canonical URL path
// for p, eliminating . and .. elements.
//
// The following rules are applied iteratively until no further processing can
// be done:
//	1. Replace multiple slashes with a single slash.
//	2. Eliminate each . path name element (the current directory).
//	3. Eliminate each inner .. path name element (the parent directory)
//	   along with the non-.. element that precedes it.
//	4. Eliminate .. elements that begin a rooted path:
//	   that is, replace "/.." by "/" at the beginning of a path.
//
// If the result of this process is an empty string, "/" is returned
func CleanPath(p string) string {
	// Turn empty string into "/"
	if p == "" {
		return "/"
	}

	n := len(p)

	// Depending of the length of the input p, call either a helper function
	// providing an appropriately sized buffer on the stack, or allocate a
	// buffer dynamically on the heap for very large inputs.
	switch {
	case n < 64:
		return cleanPathStack64(p)
	case n < 256:
		return cleanPathStack256(p)
	case n < 1024:
		return cleanPathStack1024(p)
	default:
		return cleanPathDynamic(p)
	}
}

func cleanPathStack64(p string) string {
	buf := make([]byte, 0, 64)
	return cleanPath(p, &buf)
}

func cleanPathStack256(p string) string {
	buf := make([]byte, 0, 256)
	return cleanPath(p, &buf)
}

func cleanPathStack1024(p string) string {
	buf := make([]byte, 0, 1024)
	return cleanPath(p, &buf)
}

func cleanPathDynamic(p string) string {
	buf := make([]byte, 0, len(p)+1)
	return cleanPath(p, &buf)
}

func cleanPath(p string, buf *[]byte) string {
	n := len(p)

	// Invariants:
	//      reading from path; r is index of next byte to process.
	//      writing to buf; w is index of next byte to write.

	// path must start with '/'
	r := 1
	w := 1

	if p[0] != '/' {
		r = 0
		*buf = (*buf)[:n+1]
		(*buf)[0] = '/'
	}

	trailing := n > 1 && p[n-1] == '/'

	// A bit more clunky without a 'lazybuf' like the path package, but the loop
	// gets completely inlined (bufApp). So in contrast to the path package this
	// loop has no expensive function calls (except 1x make)

	for r < n {
		switch {
		case p[r] == '/':
			// empty path element, trailing slash is added after the end
			r++

		case p[r] == '.' && r+1 == n:
			trailing = true
			r++

		case p[r] == '.' && p[r+1] == '/':
			// . element
			r += 2

		case p[r] == '.' && p[r+1] == '.' && (r+2 == n || p[r+2] == '/'):
			// .. element: remove to last /
			r += 3

			if w > 1 {
				// can backtrack
				w--

				if len(*buf) == 0 {
					for w > 1 && p[w] != '/' {
						w--
					}
				} else {
					for w > 1 && (*buf)[w] != '/' {
						w--
					}
				}
			}

		default:
			// real path element.
			// add slash if needed
			if w > 1 {
				bufApp(buf, p, w, '/')
				w++
			}

			// copy element
			for r < n && p[r] != '/' {
				bufApp(buf, p, w, p[r])
				w++
				r++
			}
		}
	}

	// re-append trailing slash
	if trailing && w > 1 {
		bufApp(buf, p, w, '/')
		w++
	}

	if len(*buf) == 0 {
		return p[:w]
	}
	return string((*buf)[:w])
}

func bufApp(buf *[]byte, s string, w int, c byte) {
	b := *buf
	if len(b) == 0 {
		if s[w] == c {
			return
		}

		*buf = (*buf)[:len(s)]
		b = *buf

		copy(b, s[:w])
	}
	b[w] = c
}
