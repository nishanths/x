// Command mp copies stdin to stdout until EOF is reached or an error
// occurs. When it receives SIGINFO, it prints to stderr in a human
// readable format the amount of data written so far. The command name
// "mp" stands for "measure and pipe".
package main

import (
	"errors"
	"io"
	"log"
	"math"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
)

type countingWriter struct {
	w       io.Writer
	written int64 // must be accessed atomically
}

var (
	_ io.Writer = (*countingWriter)(nil)
)

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	atomic.AddInt64(&c.written, int64(n))
	return n, err
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("mp: ")

	in := os.Stdin
	out := countingWriter{os.Stdout, 0}

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, sigInfo)
		for range c {
			printInfo(atomic.LoadInt64(&out.written))
		}
	}()

	if _, err := io.Copy(&out, in); err != nil {
		log.Fatal(err)
	}
}

var infoPrefix = []byte("written: ")

func printInfo(i int64) {
	p, err := human(i)
	if err != nil {
		p = inBytes(i)
	}
	os.Stderr.Write(infoPrefix)
	os.Stderr.Write(append(p, '\n'))
}

func inBytes(n int64) []byte {
	return append(strconv.AppendInt(nil, n, 10), 'B')
}

type Scale struct {
	unit   byte
	factor int64
}

var scales = []Scale{
	{'B', 1},
	{'K', 1024},
	{'M', 1024 * 1024},
	{'G', 1024 * 1024 * 1024},
	{'T', 1024 * 1024 * 1024 * 1024},
	{'P', 1024 * 1024 * 1024 * 1024 * 1024},
	{'E', 1024 * 1024 * 1024 * 1024 * 1024 * 1024},
}

// human returns a human readable representation of n.
func human(n int64) ([]byte, error) {
	/*
	 * Adapted from openbsd source: lib/libutil/fmt_scaled.c
	 * The original license is below.
	 *
	 * Copyright (c) 2001, 2002, 2003 Ian F. Darwin.  All rights reserved.
	 *
	 * Redistribution and use in source and binary forms, with or without
	 * modification, are permitted provided that the following conditions
	 * are met:
	 * 1. Redistributions of source code must retain the above copyright
	 *    notice, this list of conditions and the following disclaimer.
	 * 2. Redistributions in binary form must reproduce the above copyright
	 *    notice, this list of conditions and the following disclaimer in the
	 *    documentation and/or other materials provided with the distribution.
	 * 3. The name of the author may not be used to endorse or promote products
	 *    derived from this software without specific prior written permission.
	 *
	 * THIS SOFTWARE IS PROVIDED BY THE AUTHOR ``AS IS'' AND ANY EXPRESS OR
	 * IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
	 * OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
	 * IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY DIRECT, INDIRECT,
	 * INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
	 * NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
	 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
	 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
	 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
	 * THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
	 */
	if n == math.MinInt64 {
		return nil, errors.New("input too small")
	}

	var absval int64
	if n < 0 {
		absval = -n
	} else {
		absval = n
	}

	if absval/1024 >= scales[len(scales)-1].factor {
		return nil, errors.New("input too large")
	}

	var chosenUnit byte = 'B'
	var fract int64 = 0

	for i := range scales {
		if absval/1024 >= scales[i].factor {
			continue
		}
		chosenUnit = scales[i].unit
		fract = absval % scales[i].factor
		n /= scales[i].factor
		if i != 0 {
			fract /= scales[i-1].factor
		}
		break
	}

	fract = (10*fract + 512) / 1024
	if fract >= 10 {
		if n >= 0 {
			n++
		} else {
			n--
		}
		fract = 0
	}

	if n == 0 {
		return []byte("0B"), nil
	}

	if chosenUnit == 'B' || n >= 100 || n <= -100 {
		if fract >= 5 {
			if n >= 0 {
				n++
			} else {
				n--
			}
		}
		var result []byte
		result = strconv.AppendInt(result, n, 10)
		result = append(result, chosenUnit)
		return result, nil
	}

	var result []byte
	result = strconv.AppendInt(result, n, 10)
	result = append(result, '.')
	result = strconv.AppendInt(result, fract, 10)
	result = append(result, chosenUnit)
	return result, nil
}
