package capans

import (
	"bytes"
	"io"
)

var msgs = NewB0rker()

type MsgB0rker struct {
	topics map[string]*Topica
}

func NewB0rker() *MsgB0rker {
	return &MsgB0rker{
		topics: make(map[string]*Topica),
	}
}

func (m *MsgB0rker) Topic(name string) *Topica {
	t, ok := m.topics[name]
	if !ok {
		m.topics[name] = &Topica{}
		t = m.topics[name]
	}
	return t
}

type Topica struct {
	messages []*bytes.Buffer
	free     []int
}

func (m *Topica) Stream() *Stream {
	if len(m.free) == 0 {
		id := len(m.messages)
		s := &Stream{
			id: id,
			m:  m,
		}
		m.messages = append(m.messages, &bytes.Buffer{})
		s.buf = m.messages[id]
		return s
	}
	id := m.free[0]
	m.free = m.free[1:]
	s := &Stream{
		id: id,
		m:  m,
	}
	m.messages[id] = &bytes.Buffer{}
	s.buf = m.messages[id]
	return s
}

func (m *Topica) WriteFrom(id int, p []byte) (int, error) {
	for i := range m.messages {
		if i == id || m.messages[i] == nil {
			continue
		}
		_, err := m.messages[i].Write(p)
		if err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func (m *Topica) Close(id int) {
	m.messages[id] = nil
	m.free = append(m.free, id)
}

type Stream struct {
	id  int
	m   *Topica
	buf io.Reader
}

func (s *Stream) Id() int {
	return s.id
}

func (s *Stream) Read(p []byte) (int, error)  { return s.buf.Read(p) }
func (s *Stream) Write(p []byte) (int, error) { return s.m.WriteFrom(s.id, p) }
func (s *Stream) Close()                      { s.m.Close(s.id) }
