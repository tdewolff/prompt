package prompt

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type ProgressStyle func([]byte, float64)

func DefaultProgressStyle(b []byte, f float64) {
	if len(b) < 3 {
		return
	}
	b[0] = '['
	if math.IsNaN(f) {
		for i := 1; i < len(b); i++ {
			b[i] = ' '
		}
	} else {
		f = math.Max(0.0, math.Min(1.0, f))
		pos := 1 + int(f*float64(len(b)-2)+0.5)
		for i := 1; i < pos; i++ {
			b[i] = '#'
		}
		for i := pos; i < len(b); i++ {
			b[i] = '-'
		}
	}
	b[len(b)-1] = ']'
}

type Progress struct {
	prefix, suffix []byte
	style          ProgressStyle
	buf            []byte

	active atomic.Bool
	c      chan os.Signal
	wg     sync.WaitGroup
}

func NewProgress(prefix, suffix string, style ProgressStyle) *Progress {
	return &Progress{
		prefix: []byte(prefix),
		suffix: []byte(suffix),
		style:  style,
	}
}

func (p *Progress) Start() {
	if !p.active.CompareAndSwap(false, true) {
		return
	}

	p.c = make(chan os.Signal, 1)
	signal.Notify(p.c, os.Interrupt)
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		interrupt := false
		for _ = range p.c {
			interrupt = true
			break
		}
		if interrupt {
			p.stop()
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
	}()

	fmt.Println()
}

func (p *Progress) stop() bool {
	if !p.active.CompareAndSwap(true, false) {
		return false
	}
	signal.Stop(p.c)
	return true
}

func (p *Progress) Stop() {
	if p.stop() {
		close(p.c)
		p.wg.Wait()
	}
}

func (p *Progress) Print(f float64) {
	if !p.active.Load() {
		return
	}

	_, w, _ := TerminalSize()
	if w != len(p.buf) {
		p.buf = make([]byte, w)
	}

	copy(p.buf, p.prefix)
	if len(p.prefix)+len(p.suffix) < w {
		copy(p.buf[w-len(p.suffix):], p.suffix)
	}
	if len(p.prefix)+len(p.suffix) < len(p.buf) {
		p.style(p.buf[len(p.prefix):w-len(p.suffix)], f)
	}

	fmt.Printf(escMoveStart + escMoveUp)
	os.Stdout.Write(p.buf)
	fmt.Printf("\n")
}

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

type PercentProgress[T Number] struct {
	Progress
	value, maximum T
}

func NewPercentProgress[T Number](prefix string, maximum T, style ProgressStyle) *PercentProgress[T] {
	suffix := make([]byte, 5)
	suffix[0] = ' '
	suffix[4] = '%'
	return &PercentProgress[T]{
		Progress: Progress{
			prefix: []byte(prefix),
			suffix: suffix,
			style:  style,
		},
		maximum: maximum,
	}
}

func (p *PercentProgress[T]) update() {
	f := float64(p.value) / float64(p.maximum)
	fmt.Appendf(p.suffix[:1:5], "%3.0f", f*100.0)
	p.Print(f)
}

func (p *PercentProgress[T]) Add(value T) {
	p.value += value
	p.update()
}

func (p *PercentProgress[T]) Set(value T) {
	p.value = value
	p.update()
}

type DownloadProgress struct {
	Progress
	value int64
	resp  *http.Response
	t     time.Time
}

func NewDownloadProgress(prefix string, resp *http.Response, style ProgressStyle) *DownloadProgress {
	p := &DownloadProgress{
		Progress: Progress{
			prefix: []byte(prefix),
			style:  style,
		},
		resp: resp,
		t:    time.Now(),
	}
	p.Start()
	p.update()
	return p
}

func (p *DownloadProgress) update() {
	var f float64
	dt := time.Since(p.t)

	size, sizeUnit := formatBytes(p.value)
	sizeStr := fmt.Sprintf("%3.1f %s", size, sizeUnit)
	rate, rateUnit := formatBytes(int64(float64(p.value)/dt.Seconds() + 0.5))
	rateStr := fmt.Sprintf("%3.1f %s/s", rate, rateUnit)

	if p.resp.ContentLength <= 0 {
		f = math.NaN()
		p.suffix = fmt.Appendf(p.suffix[:0], " %8s, %10s,   ?%%", sizeStr, rateStr)
	} else {
		f = float64(p.value) / float64(p.resp.ContentLength)
		p.suffix = fmt.Appendf(p.suffix[:0], " %8s, %10s, %3.0f%%", sizeStr, rateStr, f*100.0)
	}
	p.Print(f)
	p.t = time.Now()
}

func (p *DownloadProgress) Add(value int64) {
	p.value += value
	p.update()
}

func (p *DownloadProgress) Set(value int64) {
	p.value = value
	p.update()
}

func (p *DownloadProgress) read(n int, err error) {
	p.Add(int64(n))
	if err != nil || 0 < p.resp.ContentLength && p.resp.ContentLength <= p.value {
		p.Stop()
	}
}

func (p *DownloadProgress) Read(b []byte) (int, error) {
	n, err := p.resp.Body.Read(b)
	p.read(n, err)
	return n, err
}

func (p *DownloadProgress) Close() error {
	err := p.resp.Body.Close()
	p.Stop()
	return err
}

func formatBytes(n int64) (float64, string) {
	units := []string{"GB", "MB", "kB", "B"}
	factors := []int64{1000000000, 1000000, 1000, 1}
	for i, factor := range factors {
		f := float64(n) / float64(factor)
		if v, _ := math.Modf(f); 0 < v {
			return f, units[i]
		}
	}
	return 0.0, "0"
}

type MultiDownloadProgress struct {
	items []*MultiDownloadProgressItem
	style ProgressStyle
	mu    sync.Mutex
}

type MultiDownloadProgressItem struct {
	download *DownloadProgress
	parent   *MultiDownloadProgress
	idx      int
}

func (p *MultiDownloadProgressItem) Read(b []byte) (int, error) {
	n, err := p.download.resp.Body.Read(b)

	p.parent.mu.Lock()
	pos := len(p.parent.items) - p.idx - 1
	if 0 < pos {
		fmt.Printf(escMoveUpN, pos)
	}
	p.download.read(n, err)
	if 0 < pos {
		fmt.Printf(escMoveDownN, pos)
	}
	p.parent.mu.Unlock()
	return n, err
}

func (p *MultiDownloadProgressItem) Close() error {
	p.parent.mu.Lock()
	err := p.download.Close()
	p.parent.mu.Unlock()
	return err
}

func NewMultiDownloadProgress(style ProgressStyle) *MultiDownloadProgress {
	return &MultiDownloadProgress{
		style: style,
	}
}

func (p *MultiDownloadProgress) Add(prefix string, resp *http.Response) io.ReadCloser {
	p.mu.Lock()

	idx := len(p.items)
	item := &MultiDownloadProgressItem{
		download: NewDownloadProgress(prefix, resp, p.style),
		parent:   p,
		idx:      idx,
	}
	p.items = append(p.items, item)

	p.mu.Unlock()
	return item
}

func (p *MultiDownloadProgress) Stop() {
	p.mu.Lock()
	for _, item := range p.items {
		item.download.Stop()
	}
	p.mu.Unlock()
}
